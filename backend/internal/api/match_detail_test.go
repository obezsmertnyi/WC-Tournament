package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

type fakeLookup struct {
	match storage.Match
	err   error
}

func (f *fakeLookup) GetMatchByID(context.Context, int64) (storage.Match, error) {
	return f.match, f.err
}

type fakeLiveProvider struct {
	live  *results.LiveMatch
	err   error
	calls int
}

func (f *fakeLiveProvider) LiveMatch(context.Context, string, string) (*results.LiveMatch, error) {
	f.calls++
	return f.live, f.err
}

func f64p(v float64) *float64 { return &v }

func setupDetail(lookup MatchLookup, provider LiveProvider) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterMatchDetailRoutes(r, lookup, provider)
	return r
}

func TestMatchDetail_Contract(t *testing.T) {
	lookup := &fakeLookup{match: storage.Match{ID: 1, FifaID: "400021443", FifaStageID: "289273"}}
	provider := &fakeLiveProvider{live: &results.LiveMatch{
		MatchTime:          "98'",
		Attendance:         "104103",
		Stadium:            "Estadio Azteca",
		WinnerTeamID:       "43922",
		Possession:         &results.LivePossession{Home: f64p(61.5), Away: f64p(38.5)},
		AggregateHomeScore: intp(2),
		AggregateAwayScore: intp(0),
		HomeLineup: &results.LiveLineup{
			TeamName: "Mexico", Formation: "4-1-2-3",
			Players: []results.LivePlayer{{Name: "Raul Jimenez", ShirtNumber: intp(9), Position: "Forward", Captain: true, PictureURL: "https://pic/p1.png"}},
		},
		Goals:         []results.LiveGoal{{TeamFifaID: "43922", ScorerName: "Raul Jimenez", AssistName: "Hirving Lozano", Minute: "23'", Type: intp(0)}},
		Cards:         []results.LiveCard{{TeamFifaID: "43922", PlayerName: "Hirving Lozano", Minute: "67'", Card: intp(1)}},
		Substitutions: []results.LiveSubstitution{{TeamFifaID: "43922", PlayerIn: "Santiago Gimenez", PlayerOut: "Raul Jimenez", Minute: "75'"}},
		Officials:     []results.LiveOfficial{{Name: "Cesar Ramos", Type: intp(1)}},
	}}

	r := setupDetail(lookup, provider)
	req := httptest.NewRequest(http.MethodGet, "/api/matches/1/detail", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rec.Code, rec.Body.String())
	}
	var dto map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if dto["available"] != true {
		t.Errorf("available should be true, got %v", dto["available"])
	}
	for _, k := range []string{
		"available", "matchTime", "attendance", "stadium", "winnerTeamId", "possession",
		"homeLineup", "awayLineup", "goals", "cards", "substitutions", "officials",
		"homePenaltyScore", "awayPenaltyScore", "aggregateHomeScore", "aggregateAwayScore",
	} {
		if _, ok := dto[k]; !ok {
			t.Errorf("missing key %q in detail DTO", k)
		}
	}

	goals := dto["goals"].([]any)
	if len(goals) != 1 {
		t.Fatalf("goals len: got %d", len(goals))
	}
	g := goals[0].(map[string]any)
	if g["scorer"] != "Raul Jimenez" || g["assist"] != "Hirving Lozano" || g["minute"] != "23'" {
		t.Errorf("goal contract mismatch: %v", g)
	}

	lineup := dto["homeLineup"].(map[string]any)
	if lineup["formation"] != "4-1-2-3" {
		t.Errorf("formation: %v", lineup["formation"])
	}
	if dto["awayLineup"] != nil {
		t.Errorf("awayLineup should be null, got %v", dto["awayLineup"])
	}
	poss := dto["possession"].(map[string]any)
	if poss["home"].(float64) != 61.5 {
		t.Errorf("possession home: %v", poss["home"])
	}
}

func TestMatchDetail_NotAvailable(t *testing.T) {
	lookup := &fakeLookup{match: storage.Match{ID: 2, FifaID: "x", FifaStageID: "y"}}
	provider := &fakeLiveProvider{err: results.ErrLiveNotAvailable}

	r := setupDetail(lookup, provider)
	req := httptest.NewRequest(http.MethodGet, "/api/matches/2/detail", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var dto map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &dto)
	if dto["available"] != false {
		t.Errorf("available should be false, got %v", dto["available"])
	}
}

func TestMatchDetail_NotFound(t *testing.T) {
	lookup := &fakeLookup{err: storage.ErrNotFound}
	provider := &fakeLiveProvider{}

	r := setupDetail(lookup, provider)
	req := httptest.NewRequest(http.MethodGet, "/api/matches/99/detail", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d", rec.Code)
	}
	if provider.calls != 0 {
		t.Errorf("provider should not be called when match not found")
	}
}

func TestMatchDetail_UpstreamError_502(t *testing.T) {
	lookup := &fakeLookup{match: storage.Match{ID: 3, FifaID: "x", FifaStageID: "y"}}
	provider := &fakeLiveProvider{err: errors.New("connection refused")}

	r := setupDetail(lookup, provider)
	req := httptest.NewRequest(http.MethodGet, "/api/matches/3/detail", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status: got %d want 502", rec.Code)
	}
}

func TestMatchDetail_Caches(t *testing.T) {
	lookup := &fakeLookup{match: storage.Match{ID: 4, FifaID: "x", FifaStageID: "y"}}
	provider := &fakeLiveProvider{live: &results.LiveMatch{
		MatchTime: "10'",
		Goals:     []results.LiveGoal{{ScorerName: "x", Minute: "5'"}},
	}}

	r := setupDetail(lookup, provider)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/matches/4/detail", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d", rec.Code)
		}
	}
	if provider.calls != 1 {
		t.Errorf("provider should be called once (cached), got %d", provider.calls)
	}
}

func TestMatchDetail_InvalidID(t *testing.T) {
	r := setupDetail(&fakeLookup{}, &fakeLiveProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/matches/abc/detail", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}
