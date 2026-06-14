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

// HistoryStore is the storage capability the personal-history endpoint needs.
type HistoryStore interface {
	ListUserHistory(ctx context.Context, userID int64) ([]storage.UserHistoryRow, error)
	ListTournamentPicksByUser(ctx context.Context, userID int64) ([]storage.TournamentPick, error)
	ListTeams(ctx context.Context) ([]storage.Team, error)
}

type historyTeamDTO struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	FlagURL string `json:"flagUrl"`
}

type historyMatchDTO struct {
	MatchID          int64          `json:"matchId"`
	Stage            string         `json:"stage"`
	Group            string         `json:"group"`
	KickoffAt        *string        `json:"kickoffAt"`
	Status           string         `json:"status"`
	Home             historyTeamDTO `json:"home"`
	Away             historyTeamDTO `json:"away"`
	HomeScore        *int           `json:"homeScore"`
	AwayScore        *int           `json:"awayScore"`
	PredHome         int            `json:"predHome"`
	PredAway         int            `json:"predAway"`
	WinnerPickTeamID *int64         `json:"winnerPickTeamId"`
	Points           int            `json:"points"`
	Exact            bool           `json:"exact"`
	Scored           bool           `json:"scored"`
}

type historyBonusDTO struct {
	Kind       string          `json:"kind"`
	PickRef    string          `json:"pickRef"`
	Team       *historyTeamDTO `json:"team"` // resolved team for champion/finalist; null for top scorer
	TierPoints *int            `json:"tierPoints"`
	Awarded    bool            `json:"awarded"`
}

type historyResponse struct {
	Matches     []historyMatchDTO `json:"matches"`
	Bonuses     []historyBonusDTO `json:"bonuses"`
	MatchPoints int               `json:"matchPoints"`
	BonusPoints int               `json:"bonusPoints"`
	Total       int               `json:"total"`
}

// RegisterHistoryRoutes wires GET /api/me/history (RequireUser): the caller's own
// per-match prediction results + bonus picks, for the personal "my results" view.
func RegisterHistoryRoutes(r gin.IRouter, store HistoryStore) {
	r.GET("/api/me/history", auth.RequireUser(), historyHandler(store))
}

func historyHandler(store HistoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		ctx := c.Request.Context()

		rows, err := store.ListUserHistory(ctx, claims.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load history"})
			return
		}
		picks, err := store.ListTournamentPicksByUser(ctx, claims.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bonus picks"})
			return
		}

		// Resolve team-referenced picks (champion/finalist) to a team for display.
		teams, err := store.ListTeams(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load teams"})
			return
		}
		teamByID := make(map[string]historyTeamDTO, len(teams))
		for _, tm := range teams {
			teamByID[strconv.FormatInt(tm.ID, 10)] = historyTeamDTO{Code: tm.Code, Name: tm.Name, FlagURL: tm.FlagURL}
		}

		matches := make([]historyMatchDTO, 0, len(rows))
		matchPoints := 0
		for _, r := range rows {
			var kickoff *string
			if r.KickoffAt != nil {
				s := r.KickoffAt.UTC().Format(time.RFC3339)
				kickoff = &s
			}
			if r.Scored {
				matchPoints += r.Points
			}
			matches = append(matches, historyMatchDTO{
				MatchID:          r.MatchID,
				Stage:            r.Stage,
				Group:            r.GroupLabel,
				KickoffAt:        kickoff,
				Status:           r.Status,
				Home:             historyTeamDTO{Code: r.HomeCode, Name: r.HomeName, FlagURL: r.HomeFlag},
				Away:             historyTeamDTO{Code: r.AwayCode, Name: r.AwayName, FlagURL: r.AwayFlag},
				HomeScore:        r.HomeScore,
				AwayScore:        r.AwayScore,
				PredHome:         r.PredHome,
				PredAway:         r.PredAway,
				WinnerPickTeamID: r.WinnerPickTeamID,
				Points:           r.Points,
				Exact:            r.Exact,
				Scored:           r.Scored,
			})
		}

		bonuses := make([]historyBonusDTO, 0, len(picks))
		bonusPoints := 0
		for _, p := range picks {
			if p.Awarded && p.TierPoints != nil {
				bonusPoints += *p.TierPoints
			}
			var team *historyTeamDTO
			if p.Kind == "champion" || p.Kind == "finalist" {
				if td, ok := teamByID[p.PickRef]; ok {
					team = &td
				}
			}
			bonuses = append(bonuses, historyBonusDTO{
				Kind:       p.Kind,
				PickRef:    p.PickRef,
				Team:       team,
				TierPoints: p.TierPoints,
				Awarded:    p.Awarded,
			})
		}

		c.JSON(http.StatusOK, historyResponse{
			Matches:     matches,
			Bonuses:     bonuses,
			MatchPoints: matchPoints,
			BonusPoints: bonusPoints,
			Total:       matchPoints + bonusPoints,
		})
	}
}
