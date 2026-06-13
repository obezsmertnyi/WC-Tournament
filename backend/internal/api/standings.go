package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/standings"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// StandingsReader is the storage capability the standings endpoint depends on.
// Defined consumer-side so the handler stays decoupled and unit-testable.
type StandingsReader interface {
	ListTeams(ctx context.Context) ([]storage.Team, error)
	ListFinishedGroupMatches(ctx context.Context) ([]storage.Match, error)
}

// standingsRowDTO is the frontend contract for one row in a group table.
type standingsRowDTO struct {
	TeamID  int64  `json:"teamId"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	FlagURL string `json:"flagUrl"`
	Played  int    `json:"played"`
	Win     int    `json:"win"`
	Draw    int    `json:"draw"`
	Loss    int    `json:"loss"`
	GF      int    `json:"gf"`
	GA      int    `json:"ga"`
	GD      int    `json:"gd"`
	Points  int    `json:"points"`
	Rank    int    `json:"rank"`
}

// standingsGroupDTO is the frontend contract for one group's table.
type standingsGroupDTO struct {
	Group string            `json:"group"`
	Rows  []standingsRowDTO `json:"rows"`
}

// thirdPlaceRowDTO is one line in the cross-group "best third-placed teams"
// ranking. It carries the standard stats plus the source group and whether the
// team currently sits in a Round-of-32 qualifying slot.
type thirdPlaceRowDTO struct {
	standingsRowDTO
	Group     string `json:"group"`
	Qualified bool   `json:"qualified"`
}

// RegisterStandingsRoutes wires the standings endpoint onto the router.
func RegisterStandingsRoutes(r gin.IRouter, reader StandingsReader) {
	r.GET("/api/standings", standingsHandler(reader))
}

func standingsHandler(reader StandingsReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		teamRows, err := reader.ListTeams(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load teams"})
			return
		}
		matchRows, err := reader.ListFinishedGroupMatches(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load matches"})
			return
		}

		teams := make([]standings.Team, 0, len(teamRows))
		for _, t := range teamRows {
			teams = append(teams, standings.Team{
				ID:         t.ID,
				Name:       t.Name,
				Code:       t.Code,
				FlagURL:    t.FlagURL,
				GroupLabel: t.GroupLabel,
			})
		}

		matches := make([]standings.Match, 0, len(matchRows))
		for _, m := range matchRows {
			var homeID, awayID *int64
			if m.Home != nil {
				homeID = &m.Home.ID
			}
			if m.Away != nil {
				awayID = &m.Away.ID
			}
			matches = append(matches, standings.Match{
				HomeTeamID: homeID,
				AwayTeamID: awayID,
				HomeScore:  m.HomeScore,
				AwayScore:  m.AwayScore,
			})
		}

		computed := standings.ComputeStandings(teams, matches)

		groups := make([]standingsGroupDTO, 0, len(computed))
		for _, g := range computed {
			rows := make([]standingsRowDTO, 0, len(g.Rows))
			for _, r := range g.Rows {
				rows = append(rows, standingsRowDTO{
					TeamID:  r.TeamID,
					Name:    r.Name,
					Code:    r.Code,
					FlagURL: r.FlagURL,
					Played:  r.Played,
					Win:     r.Win,
					Draw:    r.Draw,
					Loss:    r.Loss,
					GF:      r.GF,
					GA:      r.GA,
					GD:      r.GD,
					Points:  r.Points,
					Rank:    r.Rank,
				})
			}
			groups = append(groups, standingsGroupDTO{Group: g.Group, Rows: rows})
		}

		// Cross-group ranking of the third-placed teams — the 8 best advance to
		// the Round of 32 under the 48-team format.
		thirds := standings.ThirdPlaceRanking(computed)
		thirdPlace := make([]thirdPlaceRowDTO, 0, len(thirds))
		for _, r := range thirds {
			thirdPlace = append(thirdPlace, thirdPlaceRowDTO{
				standingsRowDTO: standingsRowDTO{
					TeamID:  r.TeamID,
					Name:    r.Name,
					Code:    r.Code,
					FlagURL: r.FlagURL,
					Played:  r.Played,
					Win:     r.Win,
					Draw:    r.Draw,
					Loss:    r.Loss,
					GF:      r.GF,
					GA:      r.GA,
					GD:      r.GD,
					Points:  r.Points,
					Rank:    r.Rank,
				},
				Group:     r.Group,
				Qualified: r.Qualified,
			})
		}

		c.JSON(http.StatusOK, gin.H{"groups": groups, "thirdPlace": thirdPlace})
	}
}
