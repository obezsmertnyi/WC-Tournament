package results

import "testing"

// goal is a tiny helper to build a LiveGoal with a given side and period.
func goal(side string, period int) LiveGoal {
	p := period
	return LiveGoal{Side: side, Period: &p}
}

// goalNilPeriod builds a goal whose Period is unknown (nil).
func goalNilPeriod(side string) LiveGoal { return LiveGoal{Side: side} }

func TestRegulationScore(t *testing.T) {
	tests := []struct {
		name                 string
		goals                []LiveGoal
		finalHome, finalAway int
		wantHome, wantAway   int
		wantOK               bool
	}{
		{
			// Real ARG 3:2 CPV: Messi (P3) + two ET goals (P7,P9) for ARG;
			// Duarte (P5) + one ET goal (P7) for CPV. Regulation = 1:1.
			name: "extra-time goal win recovers regulation draw",
			goals: []LiveGoal{
				goal("home", 3), goal("home", 7), goal("home", 9),
				goal("away", 5), goal("away", 7),
			},
			finalHome: 3, finalAway: 2,
			wantHome: 1, wantAway: 1, wantOK: true,
		},
		{
			name: "decided in regulation equals final",
			goals: []LiveGoal{
				goal("home", 3), goal("home", 5), goal("home", 5),
				goal("away", 5),
			},
			finalHome: 3, finalAway: 1,
			wantHome: 3, wantAway: 1, wantOK: true,
		},
		{
			name:      "no goal events but non-zero final is untrusted",
			goals:     nil,
			finalHome: 3, finalAway: 2,
			wantOK: false,
		},
		{
			name:      "empty goals with 0:0 final is trusted",
			goals:     nil,
			finalHome: 0, finalAway: 0,
			wantHome: 0, wantAway: 0, wantOK: true,
		},
		{
			name:      "missing goal events (sum != final) is untrusted",
			goals:     []LiveGoal{goal("home", 3)},
			finalHome: 3, finalAway: 2,
			wantOK: false,
		},
		{
			name:      "goal with unknown period is untrusted",
			goals:     []LiveGoal{goal("home", 3), goalNilPeriod("away")},
			finalHome: 1, finalAway: 1,
			wantOK: false,
		},
		{
			name:      "goal with unknown side is untrusted",
			goals:     []LiveGoal{goal("home", 3), goal("", 5)},
			finalHome: 1, finalAway: 1,
			wantOK: false,
		},
		{
			// Penalty shootout with no ET goals: regulation == final, no change.
			name:      "penalty shootout without ET goals keeps score",
			goals:     []LiveGoal{goal("home", 3), goal("away", 5)},
			finalHome: 1, finalAway: 1,
			wantHome: 1, wantAway: 1, wantOK: true,
		},
		{
			// Rare: ET goals for both sides, still drawn, decided on penalties.
			// aet 2:2 but regulation is 1:1.
			name:      "penalty shootout with ET goals recovers regulation",
			goals:     []LiveGoal{goal("home", 3), goal("home", 7), goal("away", 5), goal("away", 9)},
			finalHome: 2, finalAway: 2,
			wantHome: 1, wantAway: 1, wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, a, ok := RegulationScore(tt.goals, tt.finalHome, tt.finalAway)
			if ok != tt.wantOK {
				t.Fatalf("ok: got %v want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return // score is meaningless when untrusted
			}
			if h != tt.wantHome || a != tt.wantAway {
				t.Errorf("regulation score: got %d:%d want %d:%d", h, a, tt.wantHome, tt.wantAway)
			}
		})
	}
}
