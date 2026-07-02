package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/genai"
)

// stubGen is a scripted generator — the guardrail evals run the full L0–L3 logic
// with NO live LLM (ADR-0017: the deterministic plumbing is unit-tested; the model
// calls in vertex.go are not). Fields can't share the method names, hence the *Fn.
type stubGen struct {
	okv        bool
	textFn     func(model, system string, turns []Turn, schema *genai.Schema) (string, error)
	streamFn   func(turns []Turn, onToken func(string) error) error
	groundedFn func(system string, turns []Turn, tools Tools) (string, error)
	cardFn     func(query string) (string, error)
	searchFn   func(system string, turns []Turn) (string, error)
}

func (s stubGen) ok() bool { return s.okv }
func (s stubGen) text(_ context.Context, model, system string, turns []Turn, schema *genai.Schema) (string, error) {
	return s.textFn(model, system, turns, schema)
}
func (s stubGen) streamText(_ context.Context, _, _ string, turns []Turn, onToken func(string) error) error {
	return s.streamFn(turns, onToken)
}
func (s stubGen) generateGrounded(_ context.Context, system string, turns []Turn, tools Tools) (string, error) {
	if s.groundedFn == nil {
		return "", nil
	}
	return s.groundedFn(system, turns, tools)
}
func (s stubGen) groundedCard(_ context.Context, _, query string) (string, error) {
	if s.cardFn == nil {
		return "", nil
	}
	return s.cardFn(query)
}
func (s stubGen) groundedSearch(_ context.Context, system string, turns []Turn) (string, error) {
	if s.searchFn == nil {
		return "", nil
	}
	return s.searchFn(system, turns)
}

// stubTools is a scripted Tools for grounding evals — no DB.
type stubTools struct{}

func (stubTools) TournamentOverview(context.Context) (OverviewFact, error) {
	return OverviewFact{CurrentStage: "r16", MatchesPlayed: 3, MatchesTotal: 104}, nil
}
func (stubTools) RecentResults(context.Context, int) ([]MatchFact, error) {
	return []MatchFact{{Home: "England", Away: "Congo DR", Score: "2:1", Status: "finished"}}, nil
}
func (stubTools) TeamMatches(context.Context, string) ([]MatchFact, error)       { return nil, nil }
func (stubTools) GroupStandings(context.Context, string) ([]StandingFact, error) { return nil, nil }
func (stubTools) Leaderboard(context.Context, int) ([]LeaderFact, error)         { return nil, nil }
func (stubTools) TopScorers(context.Context, int) ([]ScorerFact, error) {
	return []ScorerFact{{Name: "Kane", Team: "ENG", Goals: 5}}, nil
}

// classifierVerdict answers the L1 gate by model; non-classifier calls return body.
func classifierVerdict(onTopic bool, body string) func(string, string, []Turn, *genai.Schema) (string, error) {
	return func(model, _ string, _ []Turn, _ *genai.Schema) (string, error) {
		if model == modelClassifier {
			if onTopic {
				return `{"on_topic":true}`, nil
			}
			return `{"on_topic":false}`, nil
		}
		return body, nil
	}
}

// L0 input hygiene.
func TestSanitize(t *testing.T) { // @trace FR-091
	if _, err := Sanitize("   "); !errors.Is(err, ErrInputInvalid) {
		t.Errorf("empty/whitespace should be ErrInputInvalid, got %v", err)
	}
	if _, err := Sanitize(strings.Repeat("a", maxInputChars+1)); !errors.Is(err, ErrInputInvalid) {
		t.Errorf("over-long should be ErrInputInvalid, got %v", err)
	}
	got, err := Sanitize("Who wins\x07 the\n\x00World Cup?")
	if err != nil {
		t.Fatalf("valid input errored: %v", err)
	}
	if strings.ContainsRune(got, '\x07') || strings.ContainsRune(got, '\x00') {
		t.Errorf("control chars not stripped: %q", got)
	}
	if !strings.Contains(got, "\n") {
		t.Errorf("interior newline should be preserved: %q", got)
	}
}

// The canned refusal is localized by script (Cyrillic -> Ukrainian).
func TestRefusalFor(t *testing.T) { // @trace FR-091
	if RefusalFor("привіт") != RefusalUK {
		t.Error("Cyrillic message should get the Ukrainian refusal")
	}
	if RefusalFor("hello there") != Refusal {
		t.Error("Latin message should get the English refusal")
	}
}

// L1 classifier fail-closed: an unparseable verdict is treated as off-topic, not an error.
func TestClassifyUnparseableFailsClosed(t *testing.T) { // @trace FR-091
	a := NewWithGenerator(stubGen{okv: true, textFn: func(string, string, []Turn, *genai.Schema) (string, error) {
		return "not-json", nil
	}})
	on, _, err := a.classify(context.Background(), "hi")
	if err != nil || on {
		t.Errorf("unparseable verdict must be (false,nil), got (%v,%v)", on, err)
	}
}

// Availability / access (FR-093): no generator, or a not-ok one, is Unavailable.
func TestAvailable(t *testing.T) { // @trace FR-093
	if NewWithGenerator(nil).Available() {
		t.Error("nil generator must be Unavailable")
	}
	if NewWithGenerator(stubGen{okv: false}).Available() {
		t.Error("not-ok generator must be Unavailable")
	}
	if !NewWithGenerator(stubGen{okv: true}).Available() {
		t.Error("ok generator must be Available")
	}
}

func TestStreamChatUnavailable(t *testing.T) { // @trace FR-093
	err := NewWithGenerator(nil).StreamChat(context.Background(), nil, "hi", func(string) error { return nil })
	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("want ErrUnavailable, got %v", err)
	}
}

func TestStreamChatInputCap(t *testing.T) { // @trace FR-091
	a := NewWithGenerator(stubGen{okv: true})
	err := a.StreamChat(context.Background(), nil, strings.Repeat("x", maxInputChars+1), func(string) error { return nil })
	if !errors.Is(err, ErrInputInvalid) {
		t.Errorf("want ErrInputInvalid, got %v", err)
	}
}

// Off-topic must yield exactly the canned refusal and never reach the main model.
func TestStreamChatOffTopicRefuses(t *testing.T) { // @trace FR-091
	streamed := false
	a := NewWithGenerator(stubGen{
		okv:      true,
		textFn:   classifierVerdict(false, ""),
		streamFn: func([]Turn, func(string) error) error { streamed = true; return nil },
	})
	var out strings.Builder
	if err := a.StreamChat(context.Background(), nil, "write me some python", func(tok string) error {
		out.WriteString(tok)
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != Refusal {
		t.Errorf("want canned Refusal, got %q", out.String())
	}
	if streamed {
		t.Error("off-topic must NOT call the main model")
	}
}

// On-topic streams the model tokens through.
func TestStreamChatOnTopicStreams(t *testing.T) { // @trace FR-090
	a := NewWithGenerator(stubGen{
		okv:    true,
		textFn: classifierVerdict(true, ""),
		streamFn: func(_ []Turn, onToken func(string) error) error {
			_ = onToken("Brazil ")
			return onToken("are favourites.")
		},
	})
	var out strings.Builder
	if err := a.StreamChat(context.Background(), nil, "who wins the world cup?", func(tok string) error {
		out.WriteString(tok)
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "Brazil are favourites." {
		t.Errorf("stream not delivered: %q", out.String())
	}
}

// Client-supplied history is untrusted (L0 parity): control chars stripped,
// over-long turns capped, empties dropped, role normalized.
func TestSanitizeHistory(t *testing.T) { // @trace FR-091
	out := sanitizeHistory([]Turn{
		{Role: "user", Text: "  Who\x00 plays?  "},                   // control char + padding
		{Role: "assistant", Text: "\x07\x00"},                        // strips to empty → dropped
		{Role: "model", Text: strings.Repeat("z", maxInputChars+50)}, // over-long → rune-capped
		{Role: "user", Text: "   "},                                  // empty → dropped
	})
	if len(out) != 2 {
		t.Fatalf("want 2 clean turns, got %d: %+v", len(out), out)
	}
	if out[0].Text != "Who plays?" || out[0].Role != "user" {
		t.Errorf("turn0 not cleaned: %+v", out[0])
	}
	if out[1].Role != "model" || len([]rune(out[1].Text)) != maxInputChars {
		t.Errorf("over-long model turn not capped: role=%s len=%d", out[1].Role, len([]rune(out[1].Text)))
	}
}

// StreamChat must pass sanitized history to the model (no L0 bypass) with the
// live message appended last.
func TestStreamChatSanitizesHistory(t *testing.T) { // @trace FR-091
	var seen []Turn
	a := NewWithGenerator(stubGen{
		okv:    true,
		textFn: classifierVerdict(true, ""),
		streamFn: func(turns []Turn, onToken func(string) error) error {
			seen = turns
			return onToken("ok")
		},
	})
	hist := []Turn{{Role: "user", Text: "prev\x00question"}}
	if err := a.StreamChat(context.Background(), hist, "who wins?", func(string) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seen) != 2 {
		t.Fatalf("want history+live = 2 turns, got %d", len(seen))
	}
	if strings.ContainsRune(seen[0].Text, '\x00') {
		t.Errorf("history reached the model with a control char: %q", seen[0].Text)
	}
	if seen[1].Text != "who wins?" || seen[1].Role != "user" {
		t.Errorf("live turn must be last + sanitized: %+v", seen[1])
	}
}

// A classifier transport error propagates (handler maps it to 503).
func TestStreamChatClassifyErrorPropagates(t *testing.T) { // @trace FR-091
	boom := errors.New("rpc failed")
	a := NewWithGenerator(stubGen{okv: true, textFn: func(string, string, []Turn, *genai.Schema) (string, error) {
		return "", boom
	}})
	if err := a.StreamChat(context.Background(), nil, "who wins?", func(string) error { return nil }); !errors.Is(err, boom) {
		t.Errorf("want transport error propagated, got %v", err)
	}
}

// L3 card validation: valid JSON parses; confidence defaults to "low" when absent.
func TestCardValid(t *testing.T) { // @trace FR-092
	body := `{"name":"Lionel Messi","country":"Argentina","club":"Inter Miami","summary":"Prolific forward."}`
	a := NewWithGenerator(stubGen{okv: true, textFn: classifierVerdict(true, body)})
	c, ok, err := a.Card(context.Background(), "messi")
	if err != nil || !ok || c == nil {
		t.Fatalf("want a valid card, got (%v,%v,%v)", c, ok, err)
	}
	if c.Name != "Lionel Messi" || c.Country != "Argentina" {
		t.Errorf("card fields not parsed: %+v", c)
	}
	if c.Confidence != "low" {
		t.Errorf("missing confidence should default to low, got %q", c.Confidence)
	}
}

func TestCardOffTopicRefuses(t *testing.T) { // @trace FR-092
	a := NewWithGenerator(stubGen{okv: true, textFn: classifierVerdict(false, "")})
	c, ok, err := a.Card(context.Background(), "the president")
	if err != nil || ok || c != nil {
		t.Errorf("off-topic card must be (nil,false,nil), got (%v,%v,%v)", c, ok, err)
	}
}

// Fail-closed: on-topic but the model returns an empty/invalid card -> refuse.
func TestCardFailClosedOnEmpty(t *testing.T) { // @trace FR-092
	a := NewWithGenerator(stubGen{okv: true, textFn: classifierVerdict(true, `{"name":"","country":"x","summary":""}`)})
	c, ok, err := a.Card(context.Background(), "nobody")
	if err != nil || ok || c != nil {
		t.Errorf("empty card must fail closed to (nil,false,nil), got (%v,%v,%v)", c, ok, err)
	}
}

func TestCardUnavailable(t *testing.T) { // @trace FR-093
	_, _, err := NewWithGenerator(nil).Card(context.Background(), "messi")
	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("want ErrUnavailable, got %v", err)
	}
}

// A repeat card lookup is served from cache — the model is called once.
func TestCardCache(t *testing.T) { // @trace FR-092
	cardCalls := 0
	a := NewWithGenerator(stubGen{okv: true, textFn: func(model, _ string, _ []Turn, _ *genai.Schema) (string, error) {
		if model == modelClassifier {
			return `{"on_topic":true}`, nil
		}
		cardCalls++
		return `{"name":"Messi","country":"Argentina","summary":"GOAT."}`, nil
	}})
	for i := 0; i < 2; i++ {
		c, ok, err := a.Card(context.Background(), "Messi")
		if err != nil || !ok || c == nil {
			t.Fatalf("call %d: want a card, got (%v,%v,%v)", i, c, ok, err)
		}
	}
	if cardCalls != 1 {
		t.Errorf("card model should be called once (2nd served from cache), got %d", cardCalls)
	}
}

// With tools wired, cards use Google-Search grounding (current facts) over stale
// model knowledge, and parse JSON even inside markdown fences.
func TestCardUsesGrounding(t *testing.T) { // @trace FR-092
	a := NewWithGenerator(stubGen{
		okv:    true,
		textFn: classifierVerdict(true, `{"name":"Diarra","country":"France","club":"Strasbourg","summary":"stale."}`),
		cardFn: func(string) (string, error) {
			return "```json\n{\"name\":\"Habib Diarra\",\"country\":\"Senegal\",\"club\":\"Sunderland\",\"summary\":\"Current.\",\"confidence\":\"high\"}\n```", nil
		},
	})
	a.SetTools(stubTools{})
	c, ok, err := a.Card(context.Background(), "Habib Diarra")
	if err != nil || !ok || c == nil {
		t.Fatalf("want a grounded card, got (%v,%v,%v)", c, ok, err)
	}
	if c.Club != "Sunderland" || c.Country != "Senegal" {
		t.Errorf("card should use grounded current facts, not stale model knowledge: %+v", c)
	}
}

// Grounding: dispatchTool routes to the right Tools method + fails safe.
func TestDispatchTool(t *testing.T) { // @trace FR-100
	st := stubTools{}
	if dispatchTool(context.Background(), st, toolOverview, nil)["data"] == nil {
		t.Error("overview should return data")
	}
	if dispatchTool(context.Background(), st, toolRecentResults, map[string]any{"limit": float64(5)})["data"] == nil {
		t.Error("recent_results should return data")
	}
	if dispatchTool(context.Background(), st, "does_not_exist", nil)["error"] == nil {
		t.Error("unknown tool should return an error result")
	}
	if dispatchTool(context.Background(), nil, toolOverview, nil)["error"] == nil {
		t.Error("nil Tools should return an error result")
	}
}

// When tools are wired, on-topic chat goes through the grounded path (not raw streaming).
func TestStreamChatUsesGrounding(t *testing.T) { // @trace FR-100
	a := NewWithGenerator(stubGen{
		okv:        true,
		textFn:     classifierVerdict(true, ""),
		streamFn:   func([]Turn, func(string) error) error { t.Error("must not stream when grounding succeeds"); return nil },
		groundedFn: func(string, []Turn, Tools) (string, error) { return "England beat Congo DR.", nil },
	})
	a.SetTools(stubTools{})
	var out strings.Builder
	if err := a.StreamChat(context.Background(), nil, "summary?", func(tok string) error { out.WriteString(tok); return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "England beat Congo DR." {
		t.Errorf("want the grounded answer, got %q", out.String())
	}
}

// scope=general routes chat to Google Search grounding, NOT the DB tools.
func TestStreamChatGeneralUsesSearch(t *testing.T) { // @trace FR-100
	a := NewWithGenerator(stubGen{
		okv: true,
		textFn: func(model, _ string, _ []Turn, _ *genai.Schema) (string, error) {
			return `{"on_topic":true,"scope":"general"}`, nil
		},
		groundedFn: func(string, []Turn, Tools) (string, error) {
			t.Error("general scope must NOT use DB function-calling")
			return "", nil
		},
		searchFn: func(string, []Turn) (string, error) { return "Diarra plays for Sunderland and Senegal.", nil },
	})
	a.SetTools(stubTools{})
	var out strings.Builder
	if err := a.StreamChat(context.Background(), nil, "де грає Хабіб Діара?", func(tok string) error { out.WriteString(tok); return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "Diarra plays for Sunderland and Senegal." {
		t.Errorf("general scope should use Google Search grounding, got %q", out.String())
	}
}

// Grounding failure must fall back to ungrounded streaming (chat never hard-fails).
func TestStreamChatGroundedFallback(t *testing.T) { // @trace FR-100
	streamed := false
	a := NewWithGenerator(stubGen{
		okv:        true,
		textFn:     classifierVerdict(true, ""),
		streamFn:   func(_ []Turn, onToken func(string) error) error { streamed = true; return onToken("fallback") },
		groundedFn: func(string, []Turn, Tools) (string, error) { return "", errors.New("grounding boom") },
	})
	a.SetTools(stubTools{})
	var out strings.Builder
	if err := a.StreamChat(context.Background(), nil, "who won?", func(tok string) error { out.WriteString(tok); return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !streamed || out.String() != "fallback" {
		t.Errorf("want fallback streaming, got streamed=%v out=%q", streamed, out.String())
	}
}
