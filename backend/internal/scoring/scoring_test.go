package scoring

import "testing"

func ip(v int) *int      { return &v }
func i64(v int64) *int64 { return &v }

func TestScore_GroupStage(t *testing.T) {
	rules := DefaultRules()
	tests := []struct {
		name       string
		pred       Prediction
		home, away int
		wantPts    int
		wantExact  bool
		wantOut    bool
	}{
		{"exact draw", Prediction{Home: 1, Away: 1}, 1, 1, 3, true, false},
		{"exact win", Prediction{Home: 2, Away: 0}, 2, 0, 3, true, false},
		{"correct outcome home win", Prediction{Home: 3, Away: 1}, 2, 0, 1, false, true},
		{"correct outcome draw", Prediction{Home: 0, Away: 0}, 2, 2, 1, false, true},
		{"correct outcome away win", Prediction{Home: 0, Away: 3}, 1, 2, 1, false, true},
		{"wrong outcome", Prediction{Home: 2, Away: 0}, 0, 1, 0, false, false},
		{"predicted draw actual win", Prediction{Home: 1, Away: 1}, 2, 0, 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Match{HomeScore: ip(tt.home), AwayScore: ip(tt.away)}
			pts, bd := Score(tt.pred, m, rules)
			if pts != tt.wantPts {
				t.Errorf("points = %d, want %d", pts, tt.wantPts)
			}
			if bd.Exact != tt.wantExact || bd.Outcome != tt.wantOut {
				t.Errorf("breakdown = %+v, want exact=%v outcome=%v", bd, tt.wantExact, tt.wantOut)
			}
			if bd.Total != pts {
				t.Errorf("breakdown total %d != pts %d", bd.Total, pts)
			}
		})
	}
}

func TestScore_KnockoutWinnerPick(t *testing.T) {
	rules := DefaultRules()
	france, brazil := int64(10), int64(20)

	// Predict 1:1 reg time, France advances; actual 1:1, France advanced => 3+1=4.
	m := Match{
		HomeScore: ip(1), AwayScore: ip(1),
		HomeTeamID: i64(france), AwayTeamID: i64(brazil),
		Knockout: true, AdvancerTeamID: i64(france),
	}
	pred := Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64(france)}
	pts, bd := Score(pred, m, rules)
	if pts != 4 || !bd.Exact || !bd.WinnerPick {
		t.Fatalf("exact+advancer: pts=%d bd=%+v, want 4 exact+winnerPick", pts, bd)
	}

	// Wrong advancer: 1:1 correct exact (3) but picked Brazil => 3.
	pred2 := Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64(brazil)}
	pts2, bd2 := Score(pred2, m, rules)
	if pts2 != 3 || bd2.WinnerPick {
		t.Fatalf("wrong advancer: pts=%d bd=%+v, want 3 no winnerPick", pts2, bd2)
	}

	// DECISIVE prediction (3:1): the advancer is implied by the score, so NO
	// separate +1 even though France advanced — just the outcome point. Actual
	// 2:0 (France win): outcome correct => 1, no winnerPick.
	m3 := Match{
		HomeScore: ip(2), AwayScore: ip(0),
		HomeTeamID: i64(france), AwayTeamID: i64(brazil),
		Knockout: true, AdvancerTeamID: i64(france),
	}
	pred3 := Prediction{Home: 3, Away: 1, WinnerPickTeamID: i64(france)}
	pts3, bd3 := Score(pred3, m3, rules)
	if pts3 != 1 || !bd3.Outcome || bd3.WinnerPick || bd3.Exact {
		t.Fatalf("decisive prediction must not get advancer +1: pts=%d bd=%+v, want 1", pts3, bd3)
	}

	// DRAW prediction 2:2, actual regulation draw 1:1 decided by ET/pens with
	// France advancing: outcome (draw) correct => 1, advancer correct => +1 = 2.
	m4 := Match{
		HomeScore: ip(1), AwayScore: ip(1),
		HomeTeamID: i64(france), AwayTeamID: i64(brazil),
		Knockout: true, AdvancerTeamID: i64(france),
	}
	pred4 := Prediction{Home: 2, Away: 2, WinnerPickTeamID: i64(france)}
	pts4, bd4 := Score(pred4, m4, rules)
	if pts4 != 2 || !bd4.Outcome || !bd4.WinnerPick || bd4.Exact {
		t.Fatalf("draw prediction + advancer: pts=%d bd=%+v, want 2", pts4, bd4)
	}

	// DRAW prediction but DECISIVE actual result (0:1): the match didn't go to
	// ET/pens, so NO advancer bonus even though you named the winner. You simply
	// got the outcome wrong (predicted draw, was a win) => 0.
	m5 := Match{
		HomeScore: ip(0), AwayScore: ip(1),
		HomeTeamID: i64(france), AwayTeamID: i64(brazil),
		Knockout: true, AdvancerTeamID: i64(brazil),
	}
	pred5 := Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64(brazil)}
	pts5, bd5 := Score(pred5, m5, rules)
	if pts5 != 0 || bd5.WinnerPick || bd5.Outcome || bd5.Exact {
		t.Fatalf("draw prediction, decisive actual: pts=%d bd=%+v, want 0 (no advancer)", pts5, bd5)
	}
}

func TestScore_NoResult(t *testing.T) {
	pts, bd := Score(Prediction{Home: 1, Away: 0}, Match{}, DefaultRules())
	if pts != 0 || bd.Total != 0 {
		t.Fatalf("no result should be 0, got pts=%d bd=%+v", pts, bd)
	}
}

func TestScore_Idempotent(t *testing.T) {
	rules := DefaultRules()
	m := Match{HomeScore: ip(2), AwayScore: ip(1)}
	pred := Prediction{Home: 2, Away: 1}
	p1, b1 := Score(pred, m, rules)
	p2, b2 := Score(pred, m, rules)
	if p1 != p2 || b1 != b2 {
		t.Fatalf("not idempotent: (%d,%+v) vs (%d,%+v)", p1, b1, p2, b2)
	}
}

func TestScore_NoWinnerPickWithoutAdvancer(t *testing.T) {
	rules := DefaultRules()
	// Knockout but advancer unknown yet: winner pick should not score.
	m := Match{HomeScore: ip(1), AwayScore: ip(1), Knockout: true}
	pred := Prediction{Home: 1, Away: 1, WinnerPickTeamID: i64(10)}
	pts, bd := Score(pred, m, rules)
	if pts != 3 || bd.WinnerPick {
		t.Fatalf("advancer unknown: pts=%d bd=%+v, want 3 no winnerPick", pts, bd)
	}
}
