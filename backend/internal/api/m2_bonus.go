package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// BonusStore is the storage capability the bonus endpoints depend on.
type BonusStore interface {
	ListTournamentPicksByUser(ctx context.Context, userID int64) ([]storage.TournamentPick, error)
	UpsertTournamentPick(ctx context.Context, p storage.TournamentPick) (storage.TournamentPick, error)
	GetBonusRule(ctx context.Context, kind string) (storage.BonusRule, bool, error)
	ListChampionTiers(ctx context.Context) ([]storage.ChampionTier, error)
	GroupStageLastKickoff(ctx context.Context) (*time.Time, error)
	FirstKnockoutKickoff(ctx context.Context) (*time.Time, error)
	AppendAudit(ctx context.Context, e storage.AuditEntry) error
}

type bonusPickDTO struct {
	Kind       string  `json:"kind"`
	PickRef    string  `json:"pickRef"`
	TierPoints *int    `json:"tierPoints"`
	LockedAt   *string `json:"lockedAt"`
}

// RegisterBonusRoutes wires the bonus pick endpoints (RequireUser).
func RegisterBonusRoutes(r gin.IRouter, store BonusStore) {
	grp := r.Group("/api/bonus", auth.RequireUser())
	grp.GET("/me", bonusMeHandler(store))
	grp.PUT("/champion", bonusChampionHandler(store))
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
			var locked *string
			if p.LockedAt != nil {
				s := p.LockedAt.UTC().Format(time.RFC3339)
				locked = &s
			}
			dtos = append(dtos, bonusPickDTO{
				Kind:       p.Kind,
				PickRef:    p.PickRef,
				TierPoints: p.TierPoints,
				LockedAt:   locked,
			})
		}
		c.JSON(http.StatusOK, gin.H{"picks": dtos})
	}
}

// bonusChampionHandler upserts the champion pick. The pick is changeable until
// the R32 hard lock (first knockout kickoff). The awarded tier is resolved from
// the window in which the pick is set (group stage => higher; post-group => lower),
// per ADR-0008. locked_at is re-stamped on every change so a late edit drops the
// tier — there is no "lock early, switch late for full points".
func bonusChampionHandler(store BonusStore) gin.HandlerFunc {
	type req struct {
		TeamID int64 `json:"teamId"`
	}
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		ctx := c.Request.Context()

		var body req
		if err := c.ShouldBindJSON(&body); err != nil || body.TeamID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "teamId required"})
			return
		}

		now := time.Now().UTC()

		// Hard lock at R32 kickoff.
		r32, err := store.FirstKnockoutKickoff(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule"})
			return
		}
		if r32 != nil && !now.Before(r32.UTC()) {
			c.JSON(http.StatusConflict, gin.H{"error": "champion pick is locked"})
			return
		}

		tierPts, err := resolveChampionTier(ctx, store, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve tier"})
			return
		}

		pick, err := store.UpsertTournamentPick(ctx, storage.TournamentPick{
			UserID:     claims.Sub,
			Kind:       "champion",
			PickRef:    strconv.FormatInt(body.TeamID, 10),
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
			Action:      "bonus_champion_set",
		})

		var locked *string
		if pick.LockedAt != nil {
			s := pick.LockedAt.UTC().Format(time.RFC3339)
			locked = &s
		}
		c.JSON(http.StatusOK, bonusPickDTO{
			Kind:       pick.Kind,
			PickRef:    pick.PickRef,
			TierPoints: pick.TierPoints,
			LockedAt:   locked,
		})
	}
}

// resolveChampionTier returns the tier points for a champion pick set at `at`.
// Preference order:
//  1. champion_tiers rows (the first window whose window_end >= at, ordered asc),
//  2. else the bonus_rules "champion" pts when enabled,
//  3. else nil (bonus off by default — no points stored).
//
// The window boundary is the group stage's last kickoff: a pick set on/before
// it lands in the (higher) group-stage tier; after it, the (lower) post-group tier.
func resolveChampionTier(ctx context.Context, store BonusStore, at time.Time) (*int, error) {
	tiers, err := store.ListChampionTiers(ctx)
	if err != nil {
		return nil, err
	}
	if len(tiers) > 0 {
		// Tiers are ordered by window_end asc (nulls last). The first tier whose
		// window_end >= at is the active one (earlier window => higher points).
		for _, t := range tiers {
			if t.WindowEnd == nil || !at.After(t.WindowEnd.UTC()) {
				pts := t.Pts
				return &pts, nil
			}
		}
		// Past all configured windows: use the last (lowest) tier.
		pts := tiers[len(tiers)-1].Pts
		return &pts, nil
	}

	// Fallback to a single flat bonus_rules value when enabled.
	rule, ok, err := store.GetBonusRule(ctx, "champion")
	if err != nil {
		return nil, err
	}
	if ok && rule.Enabled {
		// Time-tier via the group-stage boundary even with a flat rule:
		// before/at boundary keeps full pts; after boundary is unchanged here
		// (a single flat value has no second tier).
		groupEnd, gerr := store.GroupStageLastKickoff(ctx)
		if gerr != nil {
			return nil, gerr
		}
		_ = groupEnd // boundary informs higher/lower split only when tiers exist
		pts := rule.Pts
		return &pts, nil
	}

	// Bonus off by default: store no points.
	return nil, nil
}
