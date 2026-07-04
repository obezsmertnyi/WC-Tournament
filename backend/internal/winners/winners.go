// Package winners resolves the actual advancing team for finished knockout
// matches from FIFA's live detail (WinnerTeamID), which accounts for extra time
// and penalties. The stored winner_team_id then drives the knockout "+1
// advancer" scoring correctly even when the 90-minute scoreline is a draw.
//
// In the same pass it also corrects the stored scoreline to its regulation
// (90-minute) value when the match was won by an extra-time goal: the calendar
// feed stores only the final aet-inclusive score (e.g. a 2:2 regulation draw
// won 3:2 in extra time), which the regulation-based scoring would otherwise
// read as decisive. See results.RegulationScore (derives it from goal periods)
// and ADR-0020. This is the root fix that retires the per-match manual override
// previously needed for extra-time wins (ADR-0006).
package winners

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// Store is the storage capability the resolver needs.
type Store interface {
	ListKnockoutsNeedingWinner(ctx context.Context) ([]storage.KnockoutRef, error)
	TeamIDByFifaID(ctx context.Context, fifaID string) (int64, bool, error)
	SetMatchWinner(ctx context.Context, matchID, winnerTeamID int64) error
	CorrectKnockoutRegulationScore(ctx context.Context, matchID int64, home, away int) error
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
		r.correctRegulationScore(ctx, ref, detail)
		resolved++
	}
	if resolved > 0 {
		r.log.Info("winners: knockout advancers resolved", slog.Int("count", resolved))
	}
	return resolved, nil
}

// correctRegulationScore rewrites a just-resolved knockout's stored scoreline to
// its 90-minute value when the match was decided in extra time. It is a no-op
// for a manual override (ADR-0006 wins), when the stored score is unknown, when
// the goal data is incomplete (RegulationScore ok=false — keep the aet score),
// or when regulation already equals the stored score (decided in regulation, or
// a penalty shootout with no extra-time goals). Failures are logged, not fatal:
// the winner is already resolved and scoring falls back to the aet score.
func (r *Resolver) correctRegulationScore(ctx context.Context, ref storage.KnockoutRef, detail *results.LiveMatch) {
	if ref.ResultSource == "manual" || ref.HomeScore == nil || ref.AwayScore == nil {
		return
	}
	regHome, regAway, ok := results.RegulationScore(detail.Goals, *ref.HomeScore, *ref.AwayScore)
	if !ok || (regHome == *ref.HomeScore && regAway == *ref.AwayScore) {
		return
	}
	if err := r.store.CorrectKnockoutRegulationScore(ctx, ref.ID, regHome, regAway); err != nil {
		r.log.Warn("winners: correct regulation score failed", slog.Int64("match", ref.ID), slog.Any("error", err))
		return
	}
	r.log.Info("winners: knockout scoreline corrected to regulation",
		slog.Int64("match", ref.ID),
		slog.String("aet", fmt.Sprintf("%d:%d", *ref.HomeScore, *ref.AwayScore)),
		slog.String("regulation", fmt.Sprintf("%d:%d", regHome, regAway)))
}
