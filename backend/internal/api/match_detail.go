package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// detailCacheTTL is how long a successful match-detail response is held in
// process to respect the FIFA edge cache and avoid hammering the upstream.
const detailCacheTTL = 30 * time.Second

// MatchLookup is the storage capability the match-detail endpoint depends on:
// resolve a local match id to its FIFA identifiers.
type MatchLookup interface {
	GetMatchByID(ctx context.Context, id int64) (storage.Match, error)
}

// LiveProvider fetches normalized live/finished statistics for a match.
type LiveProvider interface {
	LiveMatch(ctx context.Context, idStage, idMatch string) (*results.LiveMatch, error)
}

// --- JSON DTOs (camelCase) ---

type detailPlayerDTO struct {
	Name        string  `json:"name"`
	ShirtNumber *int    `json:"shirtNumber"`
	Position    string  `json:"position"`
	Captain     bool    `json:"captain"`
	PictureURL  *string `json:"pictureUrl"`
}

type detailLineupDTO struct {
	TeamName  string            `json:"teamName"`
	Formation string            `json:"formation"`
	Players   []detailPlayerDTO `json:"players"`
}

type detailGoalDTO struct {
	Team   string  `json:"team"` // team FIFA id
	Scorer string  `json:"scorer"`
	Assist *string `json:"assist"`
	Minute string  `json:"minute"`
	Type   *int    `json:"type"`
}

type detailCardDTO struct {
	Team   string `json:"team"`
	Player string `json:"player"`
	Minute string `json:"minute"`
	Card   *int   `json:"card"`
}

type detailSubDTO struct {
	Team      string `json:"team"`
	PlayerIn  string `json:"playerIn"`
	PlayerOut string `json:"playerOut"`
	Minute    string `json:"minute"`
}

type detailOfficialDTO struct {
	Name string `json:"name"`
	Type *int   `json:"type"`
}

type detailPossessionDTO struct {
	Home *float64 `json:"home"`
	Away *float64 `json:"away"`
}

// matchDetailDTO is the exact JSON contract consumed by the frontend. When the
// match has no statistics yet, only {"available": false} is returned.
type matchDetailDTO struct {
	Available bool `json:"available"`

	MatchTime    string               `json:"matchTime"`
	Attendance   *string              `json:"attendance"`
	Stadium      *string              `json:"stadium"`
	WinnerTeamID *string              `json:"winnerTeamId"`
	Possession   *detailPossessionDTO `json:"possession"`

	HomeLineup *detailLineupDTO `json:"homeLineup"`
	AwayLineup *detailLineupDTO `json:"awayLineup"`

	Goals         []detailGoalDTO     `json:"goals"`
	Cards         []detailCardDTO     `json:"cards"`
	Substitutions []detailSubDTO      `json:"substitutions"`
	Officials     []detailOfficialDTO `json:"officials"`

	HomePenaltyScore   *int `json:"homePenaltyScore"`
	AwayPenaltyScore   *int `json:"awayPenaltyScore"`
	AggregateHomeScore *int `json:"aggregateHomeScore"`
	AggregateAwayScore *int `json:"aggregateAwayScore"`
}

// detailCache is a tiny in-process TTL cache keyed by match id.
type detailCache struct {
	mu  sync.Mutex
	ttl time.Duration
	now func() time.Time
	m   map[int64]detailCacheEntry
}

type detailCacheEntry struct {
	dto     matchDetailDTO
	expires time.Time
}

func newDetailCache(ttl time.Duration) *detailCache {
	return &detailCache{ttl: ttl, now: time.Now, m: make(map[int64]detailCacheEntry)}
}

func (c *detailCache) get(id int64) (matchDetailDTO, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[id]
	if !ok || c.now().After(e.expires) {
		return matchDetailDTO{}, false
	}
	return e.dto, true
}

func (c *detailCache) set(id int64, dto matchDetailDTO) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[id] = detailCacheEntry{dto: dto, expires: c.now().Add(c.ttl)}
}

// RegisterMatchDetailRoutes wires GET /api/matches/:id/detail. The caller is
// responsible for applying the auth wall (RequireUser) at the router group
// level, consistent with the other read routes.
func RegisterMatchDetailRoutes(r gin.IRouter, lookup MatchLookup, provider LiveProvider) {
	cache := newDetailCache(detailCacheTTL)
	r.GET("/api/matches/:id/detail", matchDetailHandler(lookup, provider, cache))
}

func matchDetailHandler(lookup MatchLookup, provider LiveProvider, cache *detailCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
			return
		}

		if dto, ok := cache.get(id); ok {
			c.JSON(http.StatusOK, dto)
			return
		}

		match, err := lookup.GetMatchByID(c.Request.Context(), id)
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load match"})
			return
		}

		live, err := provider.LiveMatch(c.Request.Context(), match.FifaStageID, match.FifaID)
		if errors.Is(err, results.ErrLiveNotAvailable) {
			dto := matchDetailDTO{Available: false}
			cache.set(id, dto)
			c.JSON(http.StatusOK, dto)
			return
		}
		if err != nil {
			// Network / upstream failures: never crash, surface a clear 502.
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch live match statistics"})
			return
		}

		dto := toMatchDetailDTO(live)
		cache.set(id, dto)
		c.JSON(http.StatusOK, dto)
	}
}

func toMatchDetailDTO(l *results.LiveMatch) matchDetailDTO {
	dto := matchDetailDTO{
		Available:          true,
		MatchTime:          l.MatchTime,
		Attendance:         emptyToNil(l.Attendance),
		Stadium:            emptyToNil(l.Stadium),
		WinnerTeamID:       emptyToNil(l.WinnerTeamID),
		HomeLineup:         toLineupDTO(l.HomeLineup),
		AwayLineup:         toLineupDTO(l.AwayLineup),
		Goals:              make([]detailGoalDTO, 0, len(l.Goals)),
		Cards:              make([]detailCardDTO, 0, len(l.Cards)),
		Substitutions:      make([]detailSubDTO, 0, len(l.Substitutions)),
		Officials:          make([]detailOfficialDTO, 0, len(l.Officials)),
		HomePenaltyScore:   l.HomePenaltyScore,
		AwayPenaltyScore:   l.AwayPenaltyScore,
		AggregateHomeScore: l.AggregateHomeScore,
		AggregateAwayScore: l.AggregateAwayScore,
	}

	if l.Possession != nil {
		dto.Possession = &detailPossessionDTO{Home: l.Possession.Home, Away: l.Possession.Away}
	}

	for _, g := range l.Goals {
		dto.Goals = append(dto.Goals, detailGoalDTO{
			Team:   g.TeamFifaID,
			Scorer: g.ScorerName,
			Assist: emptyToNil(g.AssistName),
			Minute: g.Minute,
			Type:   g.Type,
		})
	}
	for _, card := range l.Cards {
		dto.Cards = append(dto.Cards, detailCardDTO{
			Team:   card.TeamFifaID,
			Player: card.PlayerName,
			Minute: card.Minute,
			Card:   card.Card,
		})
	}
	for _, s := range l.Substitutions {
		dto.Substitutions = append(dto.Substitutions, detailSubDTO{
			Team:      s.TeamFifaID,
			PlayerIn:  s.PlayerIn,
			PlayerOut: s.PlayerOut,
			Minute:    s.Minute,
		})
	}
	for _, o := range l.Officials {
		dto.Officials = append(dto.Officials, detailOfficialDTO{Name: o.Name, Type: o.Type})
	}
	return dto
}

func toLineupDTO(l *results.LiveLineup) *detailLineupDTO {
	if l == nil {
		return nil
	}
	out := &detailLineupDTO{
		TeamName:  l.TeamName,
		Formation: l.Formation,
		Players:   make([]detailPlayerDTO, 0, len(l.Players)),
	}
	for _, p := range l.Players {
		out.Players = append(out.Players, detailPlayerDTO{
			Name:        p.Name,
			ShirtNumber: p.ShirtNumber,
			Position:    p.Position,
			Captain:     p.Captain,
			PictureURL:  emptyToNil(p.PictureURL),
		})
	}
	return out
}
