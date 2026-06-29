// Package scoring implements the deterministic, idempotent prediction scoring
// model from ADR-0008. The core Score function is pure (no DB, no clock) so it
// is fully unit-testable.
//
// Per-match scoring (regular time only):
//   - Exact regular-time score        => 3 points
//   - Else correct outcome (W/D/L)    => 1 point
//   - Else                            => 0 points
//
// Knockout matches add a winner pick: +1 point when the predicted advancer
// matches the team that actually advanced (resolved via extra time/penalties
// upstream — never part of the predicted score). This +1 applies ONLY when the
// prediction was a draw — with a decisive predicted score the advancer is
// already implied by the scoreline, so it is not a separate point.
package scoring

// Rules carries the configurable point values. Defaults match ADR-0008
// (exact=3, outcome=1, knockoutWinner=1).
type Rules struct {
	Exact          int
	Outcome        int
	KnockoutWinner int
}

// DefaultRules returns the ADR-0008 default scoring rules.
func DefaultRules() Rules {
	return Rules{Exact: 3, Outcome: 1, KnockoutWinner: 1}
}

// Prediction is a single player's prediction for one match.
//
// Home/Away are the predicted regular-time scoreline. WinnerPickTeamID is the
// predicted advancer for knockout matches (nil for group stage).
type Prediction struct {
	Home             int
	Away             int
	WinnerPickTeamID *int64
}

// Match is the minimal finished-match input Score needs.
//
// HomeScore/AwayScore are the actual regular-time scoreline. Knockout is true
// for any non-group stage. AdvancerTeamID is the team that actually advanced
// (knockout only) — resolved from the FIFA result upstream, including any
// extra time / penalties.
type Match struct {
	HomeScore      *int
	AwayScore      *int
	HomeTeamID     *int64
	AwayTeamID     *int64
	Knockout       bool
	AdvancerTeamID *int64
}

// Breakdown explains how the points for one match were derived. It is stored
// as breakdown_json on the points row for transparency.
type Breakdown struct {
	Exact      bool `json:"exact"`
	Outcome    bool `json:"outcome"`
	WinnerPick bool `json:"winnerPick"`
	Total      int  `json:"total"`
}

// sign returns -1, 0, or +1 for the comparison of home vs away (the outcome).
func sign(home, away int) int {
	switch {
	case home > away:
		return 1
	case home < away:
		return -1
	default:
		return 0
	}
}

// Score computes the points and breakdown for one prediction against a match
// result. It is pure and deterministic. When the match has no recorded result
// (either score nil) it returns 0 points and an empty breakdown.
func Score(pred Prediction, m Match, rules Rules) (int, Breakdown) {
	var bd Breakdown
	if m.HomeScore == nil || m.AwayScore == nil {
		return 0, bd
	}

	actualHome, actualAway := *m.HomeScore, *m.AwayScore

	switch {
	case pred.Home == actualHome && pred.Away == actualAway:
		bd.Exact = true
		bd.Total += rules.Exact
	case sign(pred.Home, pred.Away) == sign(actualHome, actualAway):
		bd.Outcome = true
		bd.Total += rules.Outcome
	}

	// Knockout winner pick: +1 ONLY when BOTH the prediction AND the actual
	// regulation result are draws — i.e. the match really went to extra time /
	// penalties, making "who advances" a separate question. With a decisive
	// actual result the winner is just the match winner (no advancer bonus, even
	// if you predicted a draw and happened to name them); and a decisive
	// prediction implies the advancer from the scoreline, so it gets no +1
	// either. Awarded when the predicted advancer matches who actually went through.
	if m.Knockout && pred.Home == pred.Away && actualHome == actualAway &&
		pred.WinnerPickTeamID != nil && m.AdvancerTeamID != nil &&
		*pred.WinnerPickTeamID == *m.AdvancerTeamID {
		bd.WinnerPick = true
		bd.Total += rules.KnockoutWinner
	}

	return bd.Total, bd
}
