package scoring

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// RulesProvider supplies the active scoring rules. Defaults to DefaultRules.
type RulesProvider interface {
	Rules(ctx context.Context) (Rules, error)
}

// Recomputer materializes points rows from predictions + results. It is
// idempotent: re-running produces the same points rows.
type Recomputer struct {
	store *storage.Store
	rules Rules
}

// NewRecomputer builds a Recomputer using the given (frozen) rules.
func NewRecomputer(store *storage.Store, rules Rules) *Recomputer {
	return &Recomputer{store: store, rules: rules}
}

// advancer derives the actual advancing team for a knockout match from the
// stored result. For a non-draw, the winner advances. For a regular-time draw
// the advancer is unknown from the score alone (extra time/penalties are not
// modeled in M2), so it returns nil and the +1 winner-pick simply doesn't score.
func advancer(m storage.MatchScoringRow) *int64 {
	if m.HomeScore == nil || m.AwayScore == nil {
		return nil
	}
	switch {
	case *m.HomeScore > *m.AwayScore:
		return m.HomeTeamID
	case *m.AwayScore > *m.HomeScore:
		return m.AwayTeamID
	default:
		return nil
	}
}

// RecomputeMatch recomputes and materializes points for every prediction on
// one match. Safe to call after each result change (sync hook).
func (rc *Recomputer) RecomputeMatch(ctx context.Context, matchID int64) error {
	m, err := rc.store.GetMatchForScoring(ctx, matchID)
	if err != nil {
		return err
	}
	// No result yet: nothing to score.
	if m.HomeScore == nil || m.AwayScore == nil {
		return nil
	}

	preds, err := rc.store.ListPredictionsForMatchRaw(ctx, matchID)
	if err != nil {
		return err
	}

	knockout := m.Stage != "group"
	adv := advancer(m)

	sm := Match{
		HomeScore:      m.HomeScore,
		AwayScore:      m.AwayScore,
		HomeTeamID:     m.HomeTeamID,
		AwayTeamID:     m.AwayTeamID,
		Knockout:       knockout,
		AdvancerTeamID: adv,
	}

	for _, p := range preds {
		pts, bd := Score(Prediction{
			Home:             p.HomePred,
			Away:             p.AwayPred,
			WinnerPickTeamID: p.WinnerPickTeamID,
		}, sm, rc.rules)

		bj, err := json.Marshal(bd)
		if err != nil {
			return fmt.Errorf("marshal breakdown: %w", err)
		}
		if err := rc.store.UpsertPoints(ctx, storage.PointRow{
			UserID:        p.UserID,
			MatchID:       matchID,
			Points:        pts,
			BreakdownJSON: bj,
		}); err != nil {
			return err
		}
	}
	return nil
}

// RecomputeAll recomputes points for every match that has a recorded result.
// Used by the recompute-scores CLI and after a frozen-rule change.
func (rc *Recomputer) RecomputeAll(ctx context.Context) error {
	ids, err := rc.store.ListMatchIDsWithResults(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := rc.RecomputeMatch(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
