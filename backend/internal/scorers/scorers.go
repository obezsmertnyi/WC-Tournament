// Package scorers aggregates goal tallies per player from finished-match detail
// data (the FIFA calendar carries only scores; scorer names are in match detail).
// The result is stored for the public top-scorers board and the top-scorer bonus.
package scorers

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// Store is the storage capability the aggregator needs.
type Store interface {
	ListMatches(ctx context.Context) ([]storage.Match, error)
	ListFinishedFifaRefs(ctx context.Context) ([]storage.FinishedFifaRef, error)
	ReplaceTopScorers(ctx context.Context, rows []storage.ScorerRow) error
}

// Provider fetches a match's detail (goals with scorer names).
type Provider interface {
	LiveMatch(ctx context.Context, idStage, idMatch string) (*results.LiveMatch, error)
}

// Aggregator ties storage + the FIFA provider together.
type Aggregator struct {
	store    Store
	provider Provider
	log      *slog.Logger
}

func New(store Store, provider Provider, log *slog.Logger) *Aggregator {
	return &Aggregator{store: store, provider: provider, log: log}
}

type tally struct {
	goals    int
	teamCode string
}

// Run rebuilds the top-scorers table from all finished matches' goal data.
func (a *Aggregator) Run(ctx context.Context) error {
	matches, err := a.store.ListMatches(ctx)
	if err != nil {
		return err
	}
	// Map match id → home/away team codes so a goal's side resolves to a team.
	type sides struct{ home, away string }
	codes := make(map[int64]sides, len(matches))
	for _, m := range matches {
		var s sides
		if m.Home != nil {
			s.home = m.Home.Code
		}
		if m.Away != nil {
			s.away = m.Away.Code
		}
		codes[m.ID] = s
	}

	refs, err := a.store.ListFinishedFifaRefs(ctx)
	if err != nil {
		return err
	}

	agg := make(map[string]*tally)
	for _, ref := range refs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		detail, err := a.provider.LiveMatch(ctx, ref.FifaStageID, ref.FifaID)
		if err != nil || detail == nil {
			continue
		}
		for _, g := range detail.Goals {
			name := strings.TrimSpace(g.ScorerName)
			if name == "" {
				continue
			}
			t := agg[name]
			if t == nil {
				t = &tally{}
				agg[name] = t
			}
			t.goals++
			if t.teamCode == "" {
				if g.Side == "away" {
					t.teamCode = codes[ref.ID].away
				} else {
					t.teamCode = codes[ref.ID].home
				}
			}
		}
		time.Sleep(250 * time.Millisecond) // gentle on the FIFA API
	}

	rows := make([]storage.ScorerRow, 0, len(agg))
	for name, t := range agg {
		rows = append(rows, storage.ScorerRow{Name: name, TeamCode: t.teamCode, Goals: t.goals})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Goals != rows[j].Goals {
			return rows[i].Goals > rows[j].Goals
		}
		return rows[i].Name < rows[j].Name
	})

	if err := a.store.ReplaceTopScorers(ctx, rows); err != nil {
		return err
	}
	a.log.Info("top scorers aggregated", slog.Int("players", len(rows)))
	return nil
}
