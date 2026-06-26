// Package winners resolves the actual advancing team for finished knockout
// matches from FIFA's live detail (WinnerTeamID), which accounts for extra time
// and penalties. The stored winner_team_id then drives the knockout "+1
// advancer" scoring correctly even when the 90-minute scoreline is a draw.
package winners

import (
	"context"
	"log/slog"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// Store is the storage capability the resolver needs.
type Store interface {
	ListKnockoutsNeedingWinner(ctx context.Context) ([]storage.FinishedFifaRef, error)
	TeamIDByFifaID(ctx context.Context, fifaID string) (int64, bool, error)
	SetMatchWinner(ctx context.Context, matchID, winnerTeamID int64) error
}

// Provider fetches a match's live detail (WinnerTeamID).
type Provider interface {
	LiveMatch(ctx context.Context, idStage, idMatch string) (*results.LiveMatch, error)
}

// Resolver ties storage + the FIFA provider together.
type Resolver struct {
	store    Store
	provider Provider
	log      *slog.Logger
}

func New(store Store, provider Provider, log *slog.Logger) *Resolver {
	return &Resolver{store: store, provider: provider, log: log}
}

// Run resolves and stores the advancer for any finished knockout match that
// doesn't have one yet. Returns the number newly resolved (so the caller can
// rescore). Empty/no-op during the group stage.
func (r *Resolver) Run(ctx context.Context) (int, error) {
	refs, err := r.store.ListKnockoutsNeedingWinner(ctx)
	if err != nil {
		return 0, err
	}
	resolved := 0
	for _, ref := range refs {
		select {
		case <-ctx.Done():
			return resolved, ctx.Err()
		default:
		}
		detail, err := r.provider.LiveMatch(ctx, ref.FifaStageID, ref.FifaID)
		if err != nil || detail == nil || detail.WinnerTeamID == "" {
			continue // detail/winner not available yet — retry next tick
		}
		teamID, ok, err := r.store.TeamIDByFifaID(ctx, detail.WinnerTeamID)
		if err != nil || !ok {
			r.log.Warn("winners: winner team not found", slog.Int64("match", ref.ID), slog.String("fifaWinner", detail.WinnerTeamID))
			continue
		}
		if err := r.store.SetMatchWinner(ctx, ref.ID, teamID); err != nil {
			r.log.Warn("winners: set winner failed", slog.Int64("match", ref.ID), slog.Any("error", err))
			continue
		}
		resolved++
	}
	if resolved > 0 {
		r.log.Info("winners: knockout advancers resolved", slog.Int("count", resolved))
	}
	return resolved, nil
}
