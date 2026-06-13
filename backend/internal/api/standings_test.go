package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeStandingsReader implements StandingsReader with in-memory data.
type fakeStandingsReader struct {
	teams   []storage.Team
	matches []storage.Match
	err     error
}

func (f *fakeStandingsReader) ListTeams(context.Context) ([]storage.Team, error) {
	return f.teams, f.err
}
func (f *fakeStandingsReader) ListFinishedGroupMatches(context.Context) ([]storage.Match, error) {
	return f.matches, f.err
}

func teamID(v int64) *int64 { return &v }

func TestStandings_Contract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reader := &fakeStandingsReader{
		teams: []storage.Team{
			{ID: 1, Name: "Mexico", Code: "MEX", FlagURL: "https://x/flags-sq-4/MEX", GroupLabel: "Group A"},
			{ID: 2, Name: "South Africa", Code: "RSA", FlagURL: "https://x/flags-sq-4/RSA", GroupLabel: "Group A"},
		},
		matches: []storage.Match{
			{
				ID:        1,
				Home:      &storage.Team{ID: 1},
				Away:      &storage.Team{ID: 2},
				HomeScore: intp(2),
				AwayScore: intp(0),
			},
		},
	}

	r := gin.New()
	RegisterStandingsRoutes(r, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/standings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rec.Code, rec.Body.String())
	}

	var env struct {
		Groups []struct {
			Group string           `json:"group"`
			Rows  []map[string]any `json:"rows"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if len(env.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(env.Groups))
	}
	g := env.Groups[0]
	if g.Group != "A" {
		t.Errorf("group should be bare letter A, got %q", g.Group)
	}
	if len(g.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(g.Rows))
	}

	// Verify exact camelCase contract keys on the winner row.
	top := g.Rows[0]
	wantKeys := []string{"teamId", "name", "code", "flagUrl", "played", "win", "draw", "loss", "gf", "ga", "gd", "points", "rank"}
	for _, k := range wantKeys {
		if _, ok := top[k]; !ok {
			t.Errorf("missing key %q in standings row", k)
		}
	}
	// JSON numbers decode to float64.
	if top["name"] != "Mexico" || top["points"].(float64) != 3 || top["gd"].(float64) != 2 || top["rank"].(float64) != 1 {
		t.Errorf("winner row mismatch: %+v", top)
	}
	if top["flagUrl"] != "https://x/flags-sq-4/MEX" {
		t.Errorf("flagUrl: got %v", top["flagUrl"])
	}

	loser := g.Rows[1]
	if loser["name"] != "South Africa" || loser["points"].(float64) != 0 || loser["rank"].(float64) != 2 {
		t.Errorf("loser row mismatch: %+v", loser)
	}
}

func TestStripGroupPrefix(t *testing.T) {
	cases := map[string]string{
		"Group A":     "A",
		"group b":     "b",
		"GROUP C":     "C",
		"  Group D  ": "D",
		"A":           "A",
		"":            "",
		"Grouping":    "Grouping",
	}
	for in, want := range cases {
		if got := stripGroupPrefix(in); got != want {
			t.Errorf("stripGroupPrefix(%q): got %q want %q", in, got, want)
		}
	}
}
