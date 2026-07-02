package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/genai"
)

// vertexGen is the production generator backed by Vertex AI, authenticated by
// keyless WIF via ADC (GOOGLE_APPLICATION_CREDENTIALS -> WIF cred-config -> tbot
// JWT; or `gcloud auth application-default login` locally). See docs/gemini-wif.md.
type vertexGen struct{ client *genai.Client }

// New builds the assistant, reusing one goroutine-safe *genai.Client process-wide.
// The feature is OPT-IN: it stays disabled unless AI_ENABLED=true is set in the
// env (logged either way). When enabled but the client can't be created (no
// ADC/WIF), it is Unavailable and handlers return 503 — the rest of the app is
// unaffected regardless.
func New(ctx context.Context) *Assistant {
	if os.Getenv("AI_ENABLED") != "true" {
		slog.Info("gemini: AI assistant DISABLED (set AI_ENABLED=true to enable)")
		return NewWithGenerator(nil)
	}
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: os.Getenv("GOOGLE_CLOUD_LOCATION"),
	})
	if err != nil {
		slog.Warn("gemini: AI_ENABLED=true but assistant unavailable (no ADC/WIF?)", slog.Any("error", err))
		return NewWithGenerator(nil)
	}
	slog.Info("gemini: AI assistant ENABLED", slog.String("project", os.Getenv("GOOGLE_CLOUD_PROJECT")))
	return NewWithGenerator(&vertexGen{client: c})
}

func (g *vertexGen) ok() bool { return g != nil && g.client != nil }

func toContents(turns []Turn) []*genai.Content {
	out := make([]*genai.Content, 0, len(turns))
	for _, t := range turns {
		var role genai.Role = genai.RoleUser
		if t.Role == "model" {
			role = genai.RoleModel
		}
		out = append(out, genai.NewContentFromText(t.Text, role))
	}
	return out
}

func genConfig(system string, schema *genai.Schema, temp float32) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{
		// System instruction is role-agnostic content; don't label the guardrail
		// persona as a `user` turn (that can weaken its priority vs real user turns).
		SystemInstruction: &genai.Content{Parts: []*genai.Part{genai.NewPartFromText(system)}},
		Temperature:       genai.Ptr(temp),
		MaxOutputTokens:   2048,
		SafetySettings:    safetySettings(),
	}
	if schema != nil {
		cfg.ResponseMIMEType = "application/json"
		cfg.ResponseSchema = schema
	}
	return cfg
}

func (g *vertexGen) text(ctx context.Context, model, system string, turns []Turn, schema *genai.Schema) (string, error) {
	temp := float32(0.2)
	switch {
	case model == modelClassifier:
		temp = 0
	case schema != nil:
		temp = 0.1
	}
	resp, err := g.client.Models.GenerateContent(ctx, model, toContents(turns), genConfig(system, schema, temp))
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}
	return resp.Text(), nil
}

func (g *vertexGen) streamText(ctx context.Context, model, system string, turns []Turn, onToken func(string) error) error {
	sent := 0
	for chunk, err := range g.client.Models.GenerateContentStream(ctx, model, toContents(turns), genConfig(system, nil, 0.2)) {
		if err != nil {
			return fmt.Errorf("gemini stream: %w", err)
		}
		t := chunk.Text()
		if t == "" {
			continue
		}
		sent += len(t)
		if sent > maxOutputChars { // L3 output cap
			break
		}
		if err := onToken(t); err != nil {
			return err
		}
	}
	return nil
}

// toolDeclarations describes the grounding tools to the model (ADR-0018).
func toolDeclarations() []*genai.FunctionDeclaration {
	obj := func(props map[string]*genai.Schema, required ...string) *genai.Schema {
		return &genai.Schema{Type: genai.TypeObject, Properties: props, Required: required}
	}
	str := func(d string) *genai.Schema { return &genai.Schema{Type: genai.TypeString, Description: d} }
	num := func(d string) *genai.Schema { return &genai.Schema{Type: genai.TypeInteger, Description: d} }
	return []*genai.FunctionDeclaration{
		{Name: toolOverview, Description: "Quick snapshot of the whole championship: current stage, matches played/total, recent results, upcoming fixtures, and the prediction-pool leader. Call this first for a general summary of the tournament."},
		{Name: toolRecentResults, Description: "Recently finished matches with scores.", Parameters: obj(map[string]*genai.Schema{"limit": num("how many, default 8")})},
		{Name: toolTeamMatches, Description: "A single team's matches (results and upcoming). Use the English team name or 3-letter code.", Parameters: obj(map[string]*genai.Schema{"team": str("team name or code in English, e.g. England or ENG")}, "team")},
		{Name: toolGroupStandings, Description: "The group table for one group.", Parameters: obj(map[string]*genai.Schema{"group": str("group letter A-L")}, "group")},
		{Name: toolLeaderboard, Description: "Top of the friends prediction-pool leaderboard.", Parameters: obj(map[string]*genai.Schema{"limit": num("how many, default 10")})},
		{Name: toolTopScorers, Description: "Tournament top goalscorers (golden-boot race): players with the most goals.", Parameters: obj(map[string]*genai.Schema{"limit": num("how many, default 10")})},
	}
}

// generateGrounded runs the function-calling loop: the model may call tools to read
// live data; we execute them and feed results back until it produces a final answer.
func (g *vertexGen) generateGrounded(ctx context.Context, system string, turns []Turn, tools Tools) (string, error) {
	cfg := genConfig(system, nil, 0.2)
	cfg.Tools = []*genai.Tool{{FunctionDeclarations: toolDeclarations()}}
	contents := toContents(turns)
	for round := 0; round < 4; round++ {
		resp, err := g.client.Models.GenerateContent(ctx, modelChat, contents, cfg)
		if err != nil {
			slog.Warn("gemini: grounded generate failed", slog.Int("round", round), slog.Any("error", err))
			return "", fmt.Errorf("gemini grounded: %w", err)
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			return resp.Text(), nil
		}
		cand := resp.Candidates[0]
		var calls []*genai.FunctionCall
		for _, p := range cand.Content.Parts {
			if p.FunctionCall != nil {
				calls = append(calls, p.FunctionCall)
			}
		}
		if len(calls) == 0 {
			return resp.Text(), nil // model answered directly
		}
		contents = append(contents, cand.Content) // the model's function-call turn
		for _, fc := range calls {
			result := dispatchTool(ctx, tools, fc.Name, fc.Args)
			contents = append(contents, genai.NewContentFromFunctionResponse(fc.Name, result, genai.RoleUser))
		}
	}
	// Tool budget exhausted — force a final answer without tools.
	resp, err := g.client.Models.GenerateContent(ctx, modelChat, contents, genConfig(system, nil, 0.2))
	if err != nil {
		return "", fmt.Errorf("gemini grounded final: %w", err)
	}
	return resp.Text(), nil
}

// groundedCard answers a card query with Google Search grounding (current web
// facts, e.g. a player's 2026 club/national team). Google Search is incompatible
// with a JSON responseSchema, so we ask for JSON in the prompt and parse the text.
func (g *vertexGen) groundedCard(ctx context.Context, system, query string) (string, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{genai.NewPartFromText(system)}},
		Temperature:       genai.Ptr[float32](0.1),
		MaxOutputTokens:   2048,
		Tools:             []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
		SafetySettings:    safetySettings(),
		// Disable "thinking": a factual card needs no chain-of-thought, and on
		// 2.5-flash thinking tokens (1300-1500 for a high-info subject like Messi)
		// share the MaxOutputTokens budget — variable thinking sometimes starved
		// the JSON output, truncating it mid-object so parseCard failed and the
		// card silently fell back to low-confidence model knowledge. Off => the
		// full budget goes to output: deterministic complete JSON, and faster.
		ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: genai.Ptr[int32](0)},
	}
	resp, err := g.client.Models.GenerateContent(ctx, modelChat, toContents([]Turn{{Role: "user", Text: query}}), cfg)
	if err != nil {
		return "", fmt.Errorf("gemini grounded card: %w", err)
	}
	return resp.Text(), nil
}

// groundedSearch answers a general-football question with Google Search grounding
// (current web facts). Text out; no responseSchema (incompatible with search).
func (g *vertexGen) groundedSearch(ctx context.Context, system string, turns []Turn) (string, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{genai.NewPartFromText(system)}},
		Temperature:       genai.Ptr[float32](0.2),
		MaxOutputTokens:   2048,
		Tools:             []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
		SafetySettings:    safetySettings(),
	}
	resp, err := g.client.Models.GenerateContent(ctx, modelChat, toContents(turns), cfg)
	if err != nil {
		return "", fmt.Errorf("gemini grounded search: %w", err)
	}
	return resp.Text(), nil
}

func safetySettings() []*genai.SafetySetting {
	block := genai.HarmBlockThresholdBlockMediumAndAbove
	return []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: block},
		{Category: genai.HarmCategoryHateSpeech, Threshold: block},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: block},
		{Category: genai.HarmCategoryDangerousContent, Threshold: block},
	}
}
