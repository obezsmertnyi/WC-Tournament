package results

import (
	"encoding/json"
	"os"
	"testing"
)

func loadLiveSample(t *testing.T) fifaLiveResponse {
	t.Helper()
	body, err := os.ReadFile("testdata/fifa_live_sample.json")
	if err != nil {
		t.Fatalf("read live sample: %v", err)
	}
	var resp fifaLiveResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal live sample: %v", err)
	}
	return resp
}

func TestParseLiveMatch(t *testing.T) {
	lm := parseLiveMatch(loadLiveSample(t))
	if lm == nil {
		t.Fatal("parseLiveMatch returned nil for a populated payload")
	}

	if lm.FifaStageID != "289273" {
		t.Errorf("FifaStageID: got %q", lm.FifaStageID)
	}
	if lm.MatchTime != "98'" {
		t.Errorf("MatchTime: got %q", lm.MatchTime)
	}
	if lm.Attendance != "104103" {
		t.Errorf("Attendance: got %q", lm.Attendance)
	}
	if lm.WinnerTeamID != "43922" {
		t.Errorf("WinnerTeamID: got %q", lm.WinnerTeamID)
	}
	if lm.Stadium != "Estadio Azteca" {
		t.Errorf("Stadium: got %q", lm.Stadium)
	}

	// Aggregate scores.
	if lm.AggregateHomeScore == nil || *lm.AggregateHomeScore != 2 {
		t.Errorf("AggregateHomeScore: got %v", lm.AggregateHomeScore)
	}
	if lm.AggregateAwayScore == nil || *lm.AggregateAwayScore != 0 {
		t.Errorf("AggregateAwayScore: got %v", lm.AggregateAwayScore)
	}
	if lm.HomePenaltyScore != nil || lm.AwayPenaltyScore != nil {
		t.Errorf("penalty scores should be nil: %v / %v", lm.HomePenaltyScore, lm.AwayPenaltyScore)
	}

	// Possession.
	if lm.Possession == nil {
		t.Fatal("possession should be present")
	}
	if lm.Possession.Home == nil || *lm.Possession.Home != 61.5 {
		t.Errorf("possession home: got %v", lm.Possession.Home)
	}
	if lm.Possession.Away == nil || *lm.Possession.Away != 38.5 {
		t.Errorf("possession away: got %v", lm.Possession.Away)
	}

	// Lineups: home has 2 players + formation; away empty -> nil.
	if lm.HomeLineup == nil {
		t.Fatal("home lineup should be present")
	}
	if lm.HomeLineup.Formation != "4-1-2-3" {
		t.Errorf("home formation: got %q", lm.HomeLineup.Formation)
	}
	if lm.HomeLineup.TeamName != "Mexico" {
		t.Errorf("home team name: got %q", lm.HomeLineup.TeamName)
	}
	if len(lm.HomeLineup.Players) != 2 {
		t.Fatalf("home players: got %d want 2", len(lm.HomeLineup.Players))
	}
	p0 := lm.HomeLineup.Players[0]
	if p0.Name != "Raul Jimenez" || p0.ShirtNumber == nil || *p0.ShirtNumber != 9 ||
		p0.Position != "Forward" || !p0.Captain {
		t.Errorf("player0 mapping mismatch: %+v", p0)
	}
	if lm.AwayLineup != nil {
		t.Errorf("away lineup should be nil (no players), got %+v", lm.AwayLineup)
	}

	// Goal: scorer + assist names resolved from the roster.
	if len(lm.Goals) != 1 {
		t.Fatalf("goals: got %d want 1", len(lm.Goals))
	}
	g := lm.Goals[0]
	if g.ScorerName != "Raul Jimenez" || g.AssistName != "Hirving Lozano" ||
		g.Minute != "23'" || g.TeamFifaID != "43922" {
		t.Errorf("goal mapping mismatch: %+v", g)
	}

	// Card: player resolved.
	if len(lm.Cards) != 1 {
		t.Fatalf("cards: got %d want 1", len(lm.Cards))
	}
	if lm.Cards[0].PlayerName != "Hirving Lozano" || lm.Cards[0].Minute != "67'" ||
		lm.Cards[0].Card == nil || *lm.Cards[0].Card != 1 {
		t.Errorf("card mapping mismatch: %+v", lm.Cards[0])
	}

	// Substitution.
	if len(lm.Substitutions) != 1 {
		t.Fatalf("subs: got %d want 1", len(lm.Substitutions))
	}
	s := lm.Substitutions[0]
	if s.PlayerOut != "Raul Jimenez" || s.PlayerIn != "Santiago Gimenez" || s.Minute != "75'" {
		t.Errorf("sub mapping mismatch: %+v", s)
	}

	// Officials.
	if len(lm.Officials) != 1 {
		t.Fatalf("officials: got %d want 1", len(lm.Officials))
	}
	if lm.Officials[0].Name != "Cesar Ramos" || lm.Officials[0].Type == nil || *lm.Officials[0].Type != 1 {
		t.Errorf("official mapping mismatch: %+v", lm.Officials[0])
	}
}

func TestParseLiveMatch_NotAvailable(t *testing.T) {
	// Empty payload (no lineups, no events) -> nil so callers return available:false.
	if lm := parseLiveMatch(fifaLiveResponse{}); lm != nil {
		t.Errorf("expected nil for empty payload, got %+v", lm)
	}
}
