package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// bonusTier is the points a tournament-wide bonus pick is worth, by when it was
// made: `group` while the group stage is still running, `post` once the knockout
// stage has begun (a later, cheaper pick). These are the pool's fixed rules.
type bonusTier struct {
	group int
	post  int
}

// bonusScheme maps each bonus kind to its time-tiered point values. Picks lock
// entirely at the Round of 16 (see bonus handler). Awarding is gated on the
// pick proving correct (tournament_picks.awarded) — these are only the
// *potential* points.
var bonusScheme = map[string]bonusTier{
	"champion":   {group: 10, post: 6},
	"finalist":   {group: 5, post: 3},
	"top_scorer": {group: 5, post: 2},
}

// BonusStore is the storage capability the bonus endpoints depend on.
type BonusStore interface {
	ListTournamentPicksByUser(ctx context.Context, userID int64) ([]storage.TournamentPick, error)
	UpsertTournamentPick(ctx context.Context, p storage.TournamentPick) (storage.TournamentPick, error)
	FirstKnockoutKickoff(ctx context.Context) (*time.Time, error)
	FirstRoundOf16Kickoff(ctx context.Context) (*time.Time, error)
	AppendAudit(ctx context.Context, e storage.AuditEntry) error
}

type bonusPickDTO struct {
	Kind       string  `json:"kind"`
	PickRef    string  `json:"pickRef"`
	TierPoints *int    `json:"tierPoints"`
	LockedAt   *string `json:"lockedAt"`
}

// RegisterBonusRoutes wires the bonus pick endpoints (RequireUser).
//
//	GET /api/bonus/me          list the caller's picks
//	PUT /api/bonus/champion    {teamId}  pick the tournament winner
//	PUT /api/bonus/finalist    {teamId}  pick the losing finalist
//	PUT /api/bonus/top-scorer  {player}  pick the golden-boot winner
func RegisterBonusRoutes(r gin.IRouter, store BonusStore) {
	grp := r.Group("/api/bonus", auth.RequireUser())
	grp.GET("/me", bonusMeHandler(store))
	grp.PUT("/champion", teamPickHandler(store, "champion"))
	grp.PUT("/finalist", teamPickHandler(store, "finalist"))
	grp.PUT("/top-scorer", playerPickHandler(store, "top_scorer"))
}

func bonusMeHandler(store BonusStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		picks, err := store.ListTournamentPicksByUser(c.Request.Context(), claims.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bonus picks"})
			return
		}
		dtos := make([]bonusPickDTO, 0, len(picks))
		for _, p := range picks {
			dtos = append(dtos, toBonusPickDTO(p))
		}
		c.JSON(http.StatusOK, gin.H{"picks": dtos})
	}
}

// teamPickHandler upserts a team-referenced bonus pick (champion / finalist).
func teamPickHandler(store BonusStore, kind string) gin.HandlerFunc {
	type req struct {
		TeamID int64 `json:"teamId"`
	}
	return func(c *gin.Context) {
		var body req
		if err := c.ShouldBindJSON(&body); err != nil || body.TeamID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "teamId required"})
			return
		}
		upsertBonus(c, store, kind, strconv.FormatInt(body.TeamID, 10))
	}
}

// playerPickHandler upserts a free-text player-referenced bonus pick (top scorer).
func playerPickHandler(store BonusStore, kind string) gin.HandlerFunc {
	type req struct {
		Player string `json:"player"`
	}
	return func(c *gin.Context) {
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		name := strings.TrimSpace(body.Player)
		if name == "" || len(name) > 80 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "player name required (max 80 chars)"})
			return
		}
		upsertBonus(c, store, kind, name)
	}
}

// upsertBonus is the shared core for all bonus kinds: it enforces the Round-of-16
// hard lock, resolves the time tier (group stage vs after), stores the pick
// (re-stamping locked_at so a late edit drops to the lower tier), and audits it.
func upsertBonus(c *gin.Context, store BonusStore, kind, pickRef string) {
	claims, _ := auth.Current(c)
	ctx := c.Request.Context()
	now := time.Now().UTC()

	// Hard lock once the Round of 16 begins — no pick may be set or changed.
	r16, err := store.FirstRoundOf16Kickoff(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule"})
		return
	}
	if r16 != nil && !now.Before(r16.UTC()) {
		c.JSON(http.StatusConflict, gin.H{"error": "bonus picks are locked"})
		return
	}

	tierPts, err := resolveBonusTier(ctx, store, kind, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve tier"})
		return
	}

	pick, err := store.UpsertTournamentPick(ctx, storage.TournamentPick{
		UserID:     claims.Sub,
		Kind:       kind,
		PickRef:    pickRef,
		LockedAt:   &now,
		TierPoints: tierPts,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save pick"})
		return
	}

	actor := claims.Sub
	_ = store.AppendAudit(ctx, storage.AuditEntry{
		ActorUserID: &actor,
		ActorRole:   claims.Role,
		Action:      "bonus_" + kind + "_set",
	})

	c.JSON(http.StatusOK, toBonusPickDTO(pick))
}

// resolveBonusTier returns the points a `kind` pick made at `at` is worth: the
// higher (group) value while the group stage is still running, the lower (post)
// value once the knockout stage has begun. Unknown kinds yield nil (no points).
func resolveBonusTier(ctx context.Context, store BonusStore, kind string, at time.Time) (*int, error) {
	tier, ok := bonusScheme[kind]
	if !ok {
		return nil, nil
	}
	firstKnockout, err := store.FirstKnockoutKickoff(ctx)
	if err != nil {
		return nil, err
	}
	pts := tier.group
	if firstKnockout != nil && !at.Before(firstKnockout.UTC()) {
		pts = tier.post // group stage is over → cheaper tier
	}
	return &pts, nil
}

func toBonusPickDTO(p storage.TournamentPick) bonusPickDTO {
	var locked *string
	if p.LockedAt != nil {
		s := p.LockedAt.UTC().Format(time.RFC3339)
		locked = &s
	}
	return bonusPickDTO{
		Kind:       p.Kind,
		PickRef:    p.PickRef,
		TierPoints: p.TierPoints,
		LockedAt:   locked,
	}
}
