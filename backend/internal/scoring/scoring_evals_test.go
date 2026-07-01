//go:build evals

// Golden-fixture evals for the scoring model — the executable quality bar for
// CAP-02 (docs/features/scoring/spec.md). Unlike the unit tests, these are
// framed as evals: each case is a named scenario from the spec with a golden
// expected outcome (the "rubric" for a deterministic function is the exact
// expected points + breakdown), annotated with the requirement it proves via
// `@trace`. Run: `go test -tags=evals ./internal/scoring/`.
//
// Determinism (FR-015) is itself eval'd: every case is scored twice and must
// yield byte-identical results (the function is pure — no clock, no IO).
package scoring

import "testing"

func iptr(v int) *int       { return &v }
func i64ptr(v int64) *int64 { return &v }

const (
	teamA int64 = 10
	teamB int64 = 20
)

type scoringEval struct {
	name      string
	trace     string // requirement id this case proves
	pred      Prediction
	match     Match
	wantTotal int
	wantBD    Breakdown
}

func evalCases() []scoringEval {
	return []scoringEval{
		{
			name: "exact score", trace: "FR-010",
			pred:      Prediction{Home: 2, Away: 1},
			match:     Match{HomeScore: iptr(2), AwayScore: iptr(1)},
			wantTotal: 3, wantBD: Breakdown{Exact: true, Total: 3},
		},
		{
			name: "correct outcome wrong score", trace: "FR-011",
			pred:      Prediction{Home: 3, Away: 1},
			match:     Match{HomeScore: iptr(2), AwayScore: iptr(1)},
			wantTotal: 1, wantBD: Breakdown{Outcome: true, Total: 1},
		},
		{
			name: "correct outcome draw, wrong score", trace: "FR-011",
			pred:      Prediction{Home: 1, Away: 1},
			match:     Match{HomeScore: iptr(2), AwayScore: iptr(2)},
			wantTotal: 1, wantBD: Breakdown{Outcome: true, Total: 1},
		},
		{
			name: "wrong outcome", trace: "FR-012",
			pred:      Prediction{Home: 1, Away: 0},
			match:     Match{HomeScore: iptr(0), AwayScore: iptr(1)},
			wantTotal: 0, wantBD: Breakdown{Total: 0},
		},
		{
			name: "knockout advancer awarded (exact draw + correct advancer)", trace: "FR-013",
			pred:      Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64ptr(teamA)},
			match:     Match{HomeScore: iptr(1), AwayScore: iptr(1), Knockout: true, AdvancerTeamID: i64ptr(teamA)},
			wantTotal: 4, wantBD: Breakdown{Exact: true, WinnerPick: true, Total: 4},
		},
		{
			name: "knockout decisive scoreline: no separate +1", trace: "FR-013",
			pred:      Prediction{Home: 2, Away: 1},
			match:     Match{HomeScore: iptr(2), AwayScore: iptr(1), Knockout: true, AdvancerTeamID: i64ptr(teamA)},
			wantTotal: 3, wantBD: Breakdown{Exact: true, Total: 3},
		},
		{
			name: "knockout regulation not a draw: no +1", trace: "FR-013",
			pred:      Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64ptr(teamA)},
			match:     Match{HomeScore: iptr(0), AwayScore: iptr(1), Knockout: true, AdvancerTeamID: i64ptr(teamB)},
			wantTotal: 0, wantBD: Breakdown{Total: 0},
		},
		{
			name: "knockout draw but wrong advancer pick: exact only", trace: "FR-013",
			pred:      Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64ptr(teamA)},
			match:     Match{HomeScore: iptr(1), AwayScore: iptr(1), Knockout: true, AdvancerTeamID: i64ptr(teamB)},
			wantTotal: 3, wantBD: Breakdown{Exact: true, Total: 3},
		},
		{
			name: "no recorded result yet", trace: "FR-015",
			pred:      Prediction{Home: 1, Away: 0},
			match:     Match{HomeScore: nil, AwayScore: nil},
			wantTotal: 0, wantBD: Breakdown{Total: 0},
		},
	}
}

func TestScoringEvals(t *testing.T) {
	// @trace: FR-010, FR-011, FR-012, FR-013, FR-015
	rules := DefaultRules()
	for _, c := range evalCases() {
		t.Run(c.name, func(t *testing.T) {
			got, bd := Score(c.pred, c.match, rules)
			if got != c.wantTotal || bd != c.wantBD {
				t.Fatalf("[%s] got total=%d bd=%+v, want total=%d bd=%+v",
					c.trace, got, bd, c.wantTotal, c.wantBD)
			}
			// FR-015: purity — scoring twice must be byte-identical.
			got2, bd2 := Score(c.pred, c.match, rules)
			if got2 != got || bd2 != bd {
				t.Fatalf("[%s] non-deterministic: first (%d,%+v) != second (%d,%+v)",
					c.trace, got, bd, got2, bd2)
			}
		})
	}
}
