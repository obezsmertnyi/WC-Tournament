package results

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// fixedNow is the clock used for deterministic status inference. It sits after
// the first (finished) sample match and before the future scheduled ones.
var fixedNow = time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)

func loadSample(t *testing.T) fifaCalendarResponse {
	t.Helper()
	body, err := os.ReadFile("testdata/fifa_calendar_sample.json")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	var resp fifaCalendarResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal sample: %v", err)
	}
	return resp
}

func ptrInt(v int) *int { return &v }

func TestParseCalendar_Mapping(t *testing.T) {
	fixtures := parseCalendar(loadSample(t), fixedNow)

	if len(fixtures) != 3 {
		t.Fatalf("expected 3 fixtures, got %d", len(fixtures))
	}

	// --- Match 1: finished group match with scores ---
	m1 := fixtures[0]
	if m1.FifaID != "400021443" {
		t.Errorf("m1 fifa id: got %q", m1.FifaID)
	}
	if m1.Stage != StageGroup {
		t.Errorf("m1 stage: got %q want group", m1.Stage)
	}
	if m1.GroupLabel != "Group A" {
		t.Errorf("m1 group: got %q", m1.GroupLabel)
	}
	if m1.MatchNumber == nil || *m1.MatchNumber != 1 {
		t.Errorf("m1 match number: got %v", m1.MatchNumber)
	}
	if m1.Status != StatusFinished {
		t.Errorf("m1 status: got %q want finished", m1.Status)
	}
	if m1.HomeScore == nil || *m1.HomeScore != 2 || m1.AwayScore == nil || *m1.AwayScore != 0 {
		t.Errorf("m1 score: got %v-%v want 2-0", m1.HomeScore, m1.AwayScore)
	}
	if m1.Home.Name != "Mexico" || m1.Home.Code != "MEX" {
		t.Errorf("m1 home: got %+v", m1.Home)
	}
	if m1.Home.FlagURL != "https://api.fifa.com/api/v3/picture/flags-{format}-{size}/MEX" {
		t.Errorf("m1 home flag: got %q", m1.Home.FlagURL)
	}
	if m1.Away.Name != "South Africa" || m1.Away.Code != "RSA" {
		t.Errorf("m1 away: got %+v", m1.Away)
	}
	if m1.VenueStadium != "Mexico City Stadium" || m1.VenueCity != "Mexico City" || m1.VenueCountry != "MEX" {
		t.Errorf("m1 venue: got %q / %q / %q", m1.VenueStadium, m1.VenueCity, m1.VenueCountry)
	}
	wantKick := time.Date(2026, 6, 11, 19, 0, 0, 0, time.UTC)
	if m1.KickoffAt == nil || !m1.KickoffAt.Equal(wantKick) {
		t.Errorf("m1 kickoff: got %v want %v", m1.KickoffAt, wantKick)
	}

	// --- Match 2: scheduled group match, localized array prefers en-GB ---
	m2 := fixtures[1]
	if m2.Stage != StageGroup {
		t.Errorf("m2 stage: got %q", m2.Stage)
	}
	if m2.GroupLabel != "Group B" {
		t.Errorf("m2 group (should prefer en-GB over fr-FR): got %q", m2.GroupLabel)
	}
	if m2.Status != StatusScheduled {
		t.Errorf("m2 status: got %q want scheduled", m2.Status)
	}
	if m2.HomeScore != nil || m2.AwayScore != nil {
		t.Errorf("m2 scores should be nil, got %v-%v", m2.HomeScore, m2.AwayScore)
	}
	if m2.Home.Name != "Canada" {
		t.Errorf("m2 home name: got %q", m2.Home.Name)
	}

	// --- Match 3: knockout R32 placeholder, no teams (pre-draw) ---
	m3 := fixtures[2]
	if m3.Stage != StageR32 {
		t.Errorf("m3 stage: got %q want r32", m3.Stage)
	}
	if m3.Home.FifaID != "" || m3.Away.FifaID != "" {
		t.Errorf("m3 teams should be empty pre-draw, got home=%q away=%q", m3.Home.FifaID, m3.Away.FifaID)
	}
	if m3.PlaceholderHome != "Winner Group A" || m3.PlaceholderAway != "Runner-up Group B" {
		t.Errorf("m3 placeholders: got %q / %q", m3.PlaceholderHome, m3.PlaceholderAway)
	}
	if m3.Status != StatusScheduled {
		t.Errorf("m3 status: got %q want scheduled", m3.Status)
	}
}

func TestLocalized(t *testing.T) {
	cases := []struct {
		name string
		in   []fifaLocalized
		want string
	}{
		{"empty", nil, ""},
		{"en-GB", []fifaLocalized{{"en-GB", "Group A"}}, "Group A"},
		{"prefer en over fr", []fifaLocalized{{"fr-FR", "Groupe A"}, {"en-GB", "Group A"}}, "Group A"},
		{"fallback first", []fifaLocalized{{"es-ES", "Grupo A"}}, "Grupo A"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := localized(tc.in); got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestMapStage(t *testing.T) {
	cases := []struct {
		id   string
		name string
		want Stage
	}{
		{"289273", "First Stage", StageGroup},
		{"289287", "Round of 32", StageR32},
		{"289288", "Round of 16", StageR16},
		{"289289", "Quarter-final", StageQF},
		{"289290", "Semi-final", StageSF},
		{"289291", "Play-off for third place", StageThird},
		{"289292", "Final", StageFinal},
		{"999999", "Round of 16", StageR16}, // unknown id, infer from name
		{"999999", "Quarter-finals", StageQF},
	}
	for _, tc := range cases {
		got := mapStage(tc.id, []fifaLocalized{{"en-GB", tc.name}})
		if got != tc.want {
			t.Errorf("mapStage(%q,%q): got %q want %q", tc.id, tc.name, got, tc.want)
		}
	}
}

func TestMapStatus(t *testing.T) {
	// fixedNow = 2026-06-12T00:00Z
	kickPastSettled := time.Date(2026, 6, 11, 19, 0, 0, 0, time.UTC) // 5h before now
	kickJustStarted := time.Date(2026, 6, 11, 23, 0, 0, 0, time.UTC) // 1h before now (in live window)
	kickFuture := time.Date(2026, 7, 1, 19, 0, 0, 0, time.UTC)

	// Explicit terminal code wins regardless of score/time.
	if got := mapStatus(ptrInt(3), true, &kickPastSettled, fixedNow); got != StatusFinished {
		t.Errorf("code 3 should be finished, got %q", got)
	}
	// FIFA code 1 (upcoming) on a future fixture must be scheduled, not live.
	if got := mapStatus(ptrInt(1), false, &kickFuture, fixedNow); got != StatusScheduled {
		t.Errorf("code 1 on future fixture should be scheduled, got %q", got)
	}
	// Score recorded + kickoff past → finished even with ambiguous code 0.
	if got := mapStatus(ptrInt(0), true, &kickPastSettled, fixedNow); got != StatusFinished {
		t.Errorf("code 0 + score + past should be finished, got %q", got)
	}
	// Kickoff just passed, no final score yet, within live window → live.
	if got := mapStatus(ptrInt(1), false, &kickJustStarted, fixedNow); got != StatusLive {
		t.Errorf("recent kickoff, no score, in window should be live, got %q", got)
	}
	// Past kickoff but well outside the live window and no score → scheduled
	// (defensive: treat as not-yet-resolved rather than perpetually live).
	if got := mapStatus(ptrInt(1), false, &kickPastSettled, fixedNow); got != StatusScheduled {
		t.Errorf("stale past kickoff, no score should be scheduled, got %q", got)
	}
	// No kickoff (pre-draw knockout slot) → scheduled.
	if got := mapStatus(ptrInt(1), false, nil, fixedNow); got != StatusScheduled {
		t.Errorf("nil kickoff should be scheduled, got %q", got)
	}
	// Future fixture, no score → scheduled.
	if got := mapStatus(ptrInt(0), false, &kickFuture, fixedNow); got != StatusScheduled {
		t.Errorf("future no-score should be scheduled, got %q", got)
	}
}
