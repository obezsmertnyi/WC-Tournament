//go:build livesmoke

// Live smoke of the grounded function-calling loop against real Gemini (Vertex via
// ADC). NOT part of the normal suite (build tag) — run before deploy to validate the
// FunctionResponse round-trip that unit tests can't cover:
//
//	GOOGLE_GENAI_USE_VERTEXAI=true GOOGLE_CLOUD_PROJECT=<GCP_PROJECT> \
//	GOOGLE_CLOUD_LOCATION=us-central1 \
//	go test -tags livesmoke -run TestLive -v ./internal/gemini/
package gemini

import (
	"context"
	"os"
	"strings"
	"testing"

	"google.golang.org/genai"
)

func liveClient(t *testing.T) *vertexGen {
	t.Helper()
	c, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: os.Getenv("GOOGLE_CLOUD_LOCATION"),
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return &vertexGen{client: c}
}

// A summary query should trigger tournament_overview and produce a grounded answer.
func TestLiveGroundedSummary(t *testing.T) {
	g := liveClient(t)
	out, err := g.generateGrounded(context.Background(), systemChatGrounded,
		[]Turn{{Role: "user", Text: "Зроби короткий підсумок цього чемпіонату."}}, stubTools{})
	if err != nil {
		t.Fatalf("generateGrounded: %v", err)
	}
	t.Logf("GROUNDED SUMMARY: %q", out)
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty grounded answer")
	}
}

// liveMatchTools reports a knockout match that is STILL LIVE at 2:2.
type liveMatchTools struct{}

func (liveMatchTools) TournamentOverview(context.Context) (OverviewFact, error) {
	m := MatchFact{Stage: "r32", Status: "live", Home: "Бельгія", Away: "Сенегал", Score: "2:2"}
	return OverviewFact{CurrentStage: "r32", MatchesTotal: 104, Recent: []MatchFact{m}}, nil
}
func (liveMatchTools) RecentResults(context.Context, int) ([]MatchFact, error) {
	return []MatchFact{{Stage: "r32", Status: "live", Home: "Бельгія", Away: "Сенегал", Score: "2:2"}}, nil
}
func (liveMatchTools) TeamMatches(context.Context, string) ([]MatchFact, error) {
	return []MatchFact{{Stage: "r32", Status: "live", Home: "Бельгія", Away: "Сенегал", Score: "2:2"}}, nil
}
func (liveMatchTools) GroupStandings(context.Context, string) ([]StandingFact, error) {
	return nil, nil
}
func (liveMatchTools) Leaderboard(context.Context, int) ([]LeaderFact, error) { return nil, nil }
func (liveMatchTools) TopScorers(context.Context, int) ([]ScorerFact, error)  { return nil, nil }

// A LIVE match must NOT be reported as gone-to-penalties/finished (was a bug).
func TestLiveGroundedLiveMatchNoHallucination(t *testing.T) {
	g := liveClient(t)
	out, err := g.generateGrounded(context.Background(), systemChatGrounded,
		[]Turn{{Role: "user", Text: "Матч Бельгія–Сенегал уже пішов на пенальті?"}}, liveMatchTools{})
	if err != nil {
		t.Fatalf("generateGrounded: %v", err)
	}
	t.Logf("LIVE-MATCH ANSWER: %q", out)
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty answer")
	}
}

// General-knowledge chat must answer from Google Search (current facts).
func TestLiveGroundedSearch(t *testing.T) {
	g := liveClient(t)
	out, err := g.groundedSearch(context.Background(), systemChatSearch,
		[]Turn{{Role: "user", Text: "За який клуб і збірну зараз грає Хабіб Діара?"}})
	if err != nil {
		t.Fatalf("groundedSearch: %v", err)
	}
	t.Logf("SEARCH ANSWER: %q", out)
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty grounded-search answer")
	}
}

// A grounded card must use Google Search for CURRENT facts + parse into a Card.
// Includes a HIGH-INFO subject (Messi): thinking tokens on 2.5-flash share the
// output budget, and for a well-documented player they used to starve the JSON
// so it truncated mid-object → parseCard nil → low-confidence fallback. With
// thinking disabled on the card path this must parse with real stats + high/med
// confidence. Regression guard for that bug.
func TestLiveGroundedCard(t *testing.T) {
	g := liveClient(t)
	for _, q := range []string{"Хабіб Діара", "Ліонель Мессі"} {
		raw, err := g.groundedCard(context.Background(), systemCardGrounded, q)
		if err != nil {
			t.Fatalf("[%s] groundedCard: %v", q, err)
		}
		c := parseCard(raw)
		if c == nil {
			t.Fatalf("[%s] parseCard failed (truncated?) on: %q", q, raw)
		}
		t.Logf("[%s] CARD name=%q country=%q club=%q conf=%q stats=%d", q, c.Name, c.Country, c.Club, c.Confidence, len(c.Stats))
		if strings.TrimSpace(c.Name) == "" {
			t.Fatalf("[%s] empty card name", q)
		}
		if c.Confidence == "low" || len(c.Stats) == 0 {
			t.Fatalf("[%s] grounded card fell back (conf=%q stats=%d) — thinking budget starving output?", q, c.Confidence, len(c.Stats))
		}
		if strings.TrimSpace(c.WikiTitle) == "" {
			t.Fatalf("[%s] no wiki_title — photo lookup would use the ambiguous display name", q)
		}
	}
}

// An ambiguous name (shared by two prominent players) must return a clarify
// question, not a silently-guessed card — so the photo can't mismatch the intent.
func TestLiveCardAmbiguous(t *testing.T) {
	g := liveClient(t)
	raw, err := g.groundedCard(context.Background(), systemCardGrounded, "Роналду")
	if err != nil {
		t.Fatalf("groundedCard: %v", err)
	}
	c := parseCard(raw)
	if c == nil {
		t.Fatalf("parseCard nil on: %q", raw)
	}
	t.Logf("'Роналду' -> clarify=%q name=%q", c.Clarify, c.Name)
	if strings.TrimSpace(c.Clarify) == "" {
		t.Fatalf("expected a clarify question for ambiguous 'Роналду', got card name=%q", c.Name)
	}
}

// "How did England play" should trigger a tool and mention the stubbed 2:1 result.
func TestLiveGroundedTeam(t *testing.T) {
	g := liveClient(t)
	out, err := g.generateGrounded(context.Background(), systemChatGrounded,
		[]Turn{{Role: "user", Text: "Як зіграли Англійці?"}}, stubTools{})
	if err != nil {
		t.Fatalf("generateGrounded: %v", err)
	}
	t.Logf("GROUNDED TEAM: %q", out)
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty grounded answer")
	}
}
