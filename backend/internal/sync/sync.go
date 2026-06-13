// Package sync orchestrates fetching fixtures from a ResultsProvider and
// upserting teams + matches into Postgres. It is idempotent on fifa_id and
// never overwrites rows whose result_source='manual' (enforced in storage).
package sync

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// Syncer ties a results provider to the storage layer.
type Syncer struct {
	provider results.ResultsProvider
	store    *storage.Store
	logger   *slog.Logger
}

// New constructs a Syncer.
func New(provider results.ResultsProvider, store *storage.Store, logger *slog.Logger) *Syncer {
	if logger == nil {
		logger = slog.Default()
	}
	return &Syncer{provider: provider, store: store, logger: logger}
}

// Result summarizes a sync run.
type Result struct {
	Fixtures int
	Teams    int
	Matches  int
}

// Run fetches the current calendar and upserts teams + matches in a single
// transaction. Teams are written first so match foreign keys resolve.
func (s *Syncer) Run(ctx context.Context) (Result, error) {
	fixtures, err := s.provider.Fixtures(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("fetch fixtures: %w", err)
	}

	res := Result{Fixtures: len(fixtures)}

	err = s.store.WithTx(ctx, func(tx pgx.Tx) error {
		// 1) Upsert every distinct team that has a FIFA id.
		seen := make(map[string]bool)
		for _, f := range fixtures {
			for _, t := range []results.FixtureTeam{f.Home, f.Away} {
				if t.FifaID == "" || seen[t.FifaID] {
					continue
				}
				seen[t.FifaID] = true
				if _, err := s.store.UpsertTeam(ctx, tx, storage.UpsertTeam{
					FifaID:     t.FifaID,
					Name:       t.Name,
					Code:       t.Code,
					FlagURL:    t.FlagURL,
					GroupLabel: f.GroupLabel,
				}); err != nil {
					return err
				}
				res.Teams++
			}
		}

		// 2) Upsert matches (team FKs resolved by fifa_id inside the upsert).
		for _, f := range fixtures {
			if err := s.store.UpsertMatch(ctx, tx, storage.UpsertMatch{
				FifaID:          f.FifaID,
				FifaStageID:     f.FifaStageID,
				Stage:           string(f.Stage),
				GroupLabel:      f.GroupLabel,
				MatchNumber:     f.MatchNumber,
				HomeFifaID:      f.Home.FifaID,
				AwayFifaID:      f.Away.FifaID,
				KickoffAt:       f.KickoffAt,
				Status:          string(f.Status),
				HomeScore:       f.HomeScore,
				AwayScore:       f.AwayScore,
				VenueStadium:    f.VenueStadium,
				VenueCity:       f.VenueCity,
				VenueCountry:    f.VenueCountry,
				PlaceholderHome: f.PlaceholderHome,
				PlaceholderAway: f.PlaceholderAway,
			}); err != nil {
				return err
			}
			res.Matches++
		}
		return nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("persist fixtures: %w", err)
	}

	s.logger.Info("fifa sync complete",
		slog.Int("fixtures", res.Fixtures),
		slog.Int("teams", res.Teams),
		slog.Int("matches", res.Matches),
	)
	return res, nil
}
