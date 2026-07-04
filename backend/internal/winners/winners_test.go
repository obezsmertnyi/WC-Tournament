package winners

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeStore records winner + score-correction calls and resolves team ids.
type fakeStore struct {
	refs         []storage.KnockoutRef
	teamByFifa   map[string]int64
	setWinner    map[int64]int64 // matchID -> winnerTeamID
	correctScore map[int64][2]int
}

func newFakeStore(refs []storage.KnockoutRef) *fakeStore {
	return &fakeStore{
		refs:         refs,
		teamByFifa:   map[string]int64{"43922": 100, "43850": 200},
		setWinner:    map[int64]int64{},
		correctScore: map[int64][2]int{},
	}
}

func (f *fakeStore) ListKnockoutsNeedingWinner(context.Context) ([]storage.KnockoutRef, error) {
	return f.refs, nil
}

func (f *fakeStore) TeamIDByFifaID(_ context.Context, fifaID string) (int64, bool, error) {
	id, ok := f.teamByFifa[fifaID]
	return id, ok, nil
}

func (f *fakeStore) SetMatchWinner(_ context.Context, matchID, winnerTeamID int64) error {
	f.setWinner[matchID] = winnerTeamID
	return nil
}

func (f *fakeStore) CorrectKnockoutRegulationScore(_ context.Context, matchID int64, home, away int) error {
	f.correctScore[matchID] = [2]int{home, away}
	return nil
}

// fakeProvider returns a preset LiveMatch keyed by fifa match id.
type fakeProvider struct{ byMatch map[string]*results.LiveMatch }

func (p *fakeProvider) LiveMatch(_ context.Context, _, idMatch string) (*results.LiveMatch, error) {
	return p.byMatch[idMatch], nil
}

func goal(side string, period int) results.LiveGoal {
	p := period
	return results.LiveGoal{Side: side, Period: &p}
}

func run(t *testing.T, store Store, prov Provider) {
	t.Helper()
	r := New(store, prov, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if _, err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_CorrectsExtraTimeScoreToRegulation(t *testing.T) {
	store := newFakeStore([]storage.KnockoutRef{
		{ID: 87, FifaID: "M87", FifaStageID: "S1", HomeScore: new(3), AwayScore: new(2), ResultSource: "fifa"},
	})
	prov := &fakeProvider{byMatch: map[string]*results.LiveMatch{
		"M87": {WinnerTeamID: "43922", Goals: []results.LiveGoal{
			goal("home", 3), goal("home", 7), goal("home", 9), // ARG: 1 reg + 2 ET
			goal("away", 5), goal("away", 7), // CPV: 1 reg + 1 ET
		}},
	}}
	run(t, store, prov)

	if store.setWinner[87] != 100 {
		t.Errorf("winner: got %d want 100", store.setWinner[87])
	}
	if got := store.correctScore[87]; got != [2]int{1, 1} {
		t.Errorf("regulation correction: got %v want [1 1]", got)
	}
}

func TestRun_SkipsManualOverride(t *testing.T) {
	store := newFakeStore([]storage.KnockoutRef{
		{ID: 81, FifaID: "M81", FifaStageID: "S1", HomeScore: new(2), AwayScore: new(2), ResultSource: "manual"},
	})
	prov := &fakeProvider{byMatch: map[string]*results.LiveMatch{
		"M81": {WinnerTeamID: "43922", Goals: []results.LiveGoal{
			goal("home", 3), goal("home", 7), goal("away", 5),
		}},
	}}
	run(t, store, prov)

	if store.setWinner[81] != 100 {
		t.Errorf("winner should still resolve for a manual row: got %d", store.setWinner[81])
	}
	if _, corrected := store.correctScore[81]; corrected {
		t.Errorf("manual override must not be re-scored (ADR-0006): %v", store.correctScore[81])
	}
}

func TestRun_NoCorrectionWhenDecidedInRegulation(t *testing.T) {
	store := newFakeStore([]storage.KnockoutRef{
		{ID: 5, FifaID: "M5", FifaStageID: "S1", HomeScore: new(2), AwayScore: new(1), ResultSource: "fifa"},
	})
	prov := &fakeProvider{byMatch: map[string]*results.LiveMatch{
		"M5": {WinnerTeamID: "43922", Goals: []results.LiveGoal{
			goal("home", 3), goal("home", 5), goal("away", 5),
		}},
	}}
	run(t, store, prov)

	if _, corrected := store.correctScore[5]; corrected {
		t.Errorf("regulation win must not be corrected: %v", store.correctScore[5])
	}
}

func TestRun_NoCorrectionWhenGoalDataIncomplete(t *testing.T) {
	store := newFakeStore([]storage.KnockoutRef{
		{ID: 9, FifaID: "M9", FifaStageID: "S1", HomeScore: new(3), AwayScore: new(2), ResultSource: "fifa"},
	})
	// Winner present but no goal timeline — the safeguard must keep the aet score.
	prov := &fakeProvider{byMatch: map[string]*results.LiveMatch{
		"M9": {WinnerTeamID: "43922"},
	}}
	run(t, store, prov)

	if store.setWinner[9] != 100 {
		t.Errorf("winner should resolve: got %d", store.setWinner[9])
	}
	if _, corrected := store.correctScore[9]; corrected {
		t.Errorf("incomplete goal data must not overwrite the score: %v", store.correctScore[9])
	}
}
