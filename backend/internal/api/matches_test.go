package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeReader implements MatchReader with in-memory data.
type fakeReader struct {
	matches []storage.Match
	teams   []storage.Team
	err     error
}

func (f *fakeReader) ListMatches(context.Context) ([]storage.Match, error) {
	return f.matches, f.err
}
func (f *fakeReader) ListTeams(context.Context) ([]storage.Team, error) {
	return f.teams, f.err
}

func intp(v int) *int { return &v }

func TestListMatches_Contract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	kick := time.Date(2026, 6, 11, 19, 0, 0, 0, time.UTC)
	reader := &fakeReader{
		matches: []storage.Match{
			{
				ID: 1, Stage: "group", GroupLabel: "A", MatchNumber: intp(1),
				KickoffAt: &kick, Status: "scheduled",
				Home:         &storage.Team{Name: "Mexico", Code: "MEX", FlagURL: "https://flag/MEX"},
				Away:         &storage.Team{Name: "South Africa", Code: "RSA", FlagURL: "https://flag/RSA"},
				VenueStadium: "Estadio Azteca", VenueCity: "Mexico City", VenueCountry: "Mexico",
			},
			{
				// Pre-draw knockout: no teams, placeholders present.
				ID: 73, Stage: "r32", Status: "scheduled",
				PlaceholderHome: "Winner Group A", PlaceholderAway: "Runner-up Group B",
				VenueStadium: "LA Stadium", VenueCity: "Los Angeles", VenueCountry: "USA",
			},
		},
	}

	r := gin.New()
	RegisterReadRoutes(r, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/matches", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}

	var env struct {
		Matches []json.RawMessage `json:"matches"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody=%s", err, rec.Body.String())
	}
	if len(env.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(env.Matches))
	}

	// Assert the first match matches the documented frontend contract exactly.
	var m1 map[string]any
	if err := json.Unmarshal(env.Matches[0], &m1); err != nil {
		t.Fatalf("decode m1: %v", err)
	}
	wantKeys := []string{
		"id", "stage", "group", "matchNumber", "kickoffAt", "status",
		"home", "away", "homeScore", "awayScore", "venue",
		"placeholderHome", "placeholderAway",
	}
	for _, k := range wantKeys {
		if _, ok := m1[k]; !ok {
			t.Errorf("missing key %q in match DTO", k)
		}
	}
	if m1["stage"] != "group" || m1["group"] != "A" {
		t.Errorf("m1 stage/group: %v / %v", m1["stage"], m1["group"])
	}
	if m1["kickoffAt"] != "2026-06-11T19:00:00Z" {
		t.Errorf("m1 kickoffAt: %v", m1["kickoffAt"])
	}
	if m1["homeScore"] != nil || m1["awayScore"] != nil {
		t.Errorf("m1 scores should be null: %v / %v", m1["homeScore"], m1["awayScore"])
	}
	home := m1["home"].(map[string]any)
	if home["name"] != "Mexico" || home["code"] != "MEX" || home["flagUrl"] != "https://flag/MEX" {
		t.Errorf("m1 home contract mismatch: %v", home)
	}
	venue := m1["venue"].(map[string]any)
	if venue["stadium"] != "Estadio Azteca" || venue["city"] != "Mexico City" || venue["country"] != "Mexico" {
		t.Errorf("m1 venue contract mismatch: %v", venue)
	}
	if m1["placeholderHome"] != nil {
		t.Errorf("m1 placeholderHome should be null, got %v", m1["placeholderHome"])
	}

	// Knockout pre-draw row: home/away null, placeholders set.
	var m2 map[string]any
	_ = json.Unmarshal(env.Matches[1], &m2)
	if m2["home"] != nil || m2["away"] != nil {
		t.Errorf("m2 home/away should be null pre-draw: %v / %v", m2["home"], m2["away"])
	}
	if m2["placeholderHome"] != "Winner Group A" || m2["placeholderAway"] != "Runner-up Group B" {
		t.Errorf("m2 placeholders: %v / %v", m2["placeholderHome"], m2["placeholderAway"])
	}
}

func TestListTeams_Contract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reader := &fakeReader{
		teams: []storage.Team{
			{ID: 1, Name: "Mexico", Code: "MEX", FlagURL: "https://flag/MEX", GroupLabel: "A"},
			{ID: 2, Name: "Unassigned", Code: "", FlagURL: "", GroupLabel: ""},
		},
	}
	r := gin.New()
	RegisterReadRoutes(r, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/teams", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var env struct {
		Teams []map[string]any `json:"teams"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(env.Teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(env.Teams))
	}
	if env.Teams[0]["group"] != "A" || env.Teams[0]["flagUrl"] != "https://flag/MEX" {
		t.Errorf("team0 contract mismatch: %v", env.Teams[0])
	}
	if env.Teams[1]["group"] != nil {
		t.Errorf("team1 group should be null, got %v", env.Teams[1]["group"])
	}
}
