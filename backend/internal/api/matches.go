package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// MatchReader is the storage capability the read API depends on. Defined here
// (consumer side) so handlers stay decoupled and easily testable.
type MatchReader interface {
	ListMatches(ctx context.Context) ([]storage.Match, error)
	ListTeams(ctx context.Context) ([]storage.Team, error)
}

// teamDTO is the frontend contract for a team embedded in a match.
type teamDTO struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	FlagURL string `json:"flagUrl"`
}

// venueDTO is the frontend contract for a match venue.
type venueDTO struct {
	Stadium string `json:"stadium"`
	City    string `json:"city"`
	Country string `json:"country"`
}

// matchDTO is the exact JSON contract consumed by the frontend.
type matchDTO struct {
	ID              int64    `json:"id"`
	Stage           string   `json:"stage"`
	Group           *string  `json:"group"`
	MatchNumber     *int     `json:"matchNumber"`
	KickoffAt       *string  `json:"kickoffAt"`
	Status          string   `json:"status"`
	Home            *teamDTO `json:"home"`
	Away            *teamDTO `json:"away"`
	HomeScore       *int     `json:"homeScore"`
	AwayScore       *int     `json:"awayScore"`
	WinnerTeamID    *int64   `json:"winnerTeamId"` // knockout advancer (ET/pens); nil otherwise
	Venue           venueDTO `json:"venue"`
	PlaceholderHome *string  `json:"placeholderHome"`
	PlaceholderAway *string  `json:"placeholderAway"`
}

// teamListDTO is the frontend contract for GET /api/teams entries.
type teamListDTO struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Code    string  `json:"code"`
	FlagURL string  `json:"flagUrl"`
	Group   *string `json:"group"`
}

// RegisterReadRoutes wires the read API onto the router.
func RegisterReadRoutes(r gin.IRouter, reader MatchReader) {
	r.GET("/api/matches", listMatchesHandler(reader))
	r.GET("/api/teams", listTeamsHandler(reader))
}

func listMatchesHandler(reader MatchReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := reader.ListMatches(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load matches"})
			return
		}
		dtos := make([]matchDTO, 0, len(rows))
		for i := range rows {
			dtos = append(dtos, toMatchDTO(rows[i]))
		}
		c.JSON(http.StatusOK, gin.H{"matches": dtos})
	}
}

func listTeamsHandler(reader MatchReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := reader.ListTeams(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load teams"})
			return
		}
		dtos := make([]teamListDTO, 0, len(rows))
		for _, t := range rows {
			dtos = append(dtos, teamListDTO{
				ID:      t.ID,
				Name:    t.Name,
				Code:    t.Code,
				FlagURL: t.FlagURL,
				Group:   emptyToNil(t.GroupLabel),
			})
		}
		c.JSON(http.StatusOK, gin.H{"teams": dtos})
	}
}

func toMatchDTO(m storage.Match) matchDTO {
	dto := matchDTO{
		ID:              m.ID,
		Stage:           m.Stage,
		Group:           emptyToNil(stripGroupPrefix(m.GroupLabel)),
		MatchNumber:     m.MatchNumber,
		Status:          m.Status,
		HomeScore:       m.HomeScore,
		AwayScore:       m.AwayScore,
		WinnerTeamID:    m.WinnerTeamID,
		Venue:           venueDTO{Stadium: m.VenueStadium, City: m.VenueCity, Country: m.VenueCountry},
		PlaceholderHome: emptyToNil(m.PlaceholderHome),
		PlaceholderAway: emptyToNil(m.PlaceholderAway),
	}
	if m.KickoffAt != nil {
		s := m.KickoffAt.UTC().Format(time.RFC3339)
		dto.KickoffAt = &s
	}
	if m.Home != nil {
		dto.Home = &teamDTO{ID: m.Home.ID, Name: m.Home.Name, Code: m.Home.Code, FlagURL: m.Home.FlagURL}
	}
	if m.Away != nil {
		dto.Away = &teamDTO{ID: m.Away.ID, Name: m.Away.Name, Code: m.Away.Code, FlagURL: m.Away.FlagURL}
	}
	return dto
}

// emptyToNil maps "" to a nil *string so the JSON renders null, not "".
func emptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// stripGroupPrefix returns the bare group letter for a stored group_label.
// The DB stores the verbatim FIFA label (e.g. "Group A"); the API contract is
// the bare letter ("A") so the frontend can compose its own label without
// duplicating the word. A leading "Group " prefix is removed case-insensitively;
// any surrounding whitespace is trimmed. Empty input returns "".
func stripGroupPrefix(label string) string {
	s := strings.TrimSpace(label)
	const prefix = "group "
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		s = strings.TrimSpace(s[len(prefix):])
	}
	return s
}
