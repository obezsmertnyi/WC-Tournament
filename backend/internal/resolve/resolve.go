// Package resolve awards the tournament-wide bonus picks once the data exists:
// champion + finalist from the FINAL match result, and the top scorer by
// aggregating goals across all finished matches. It is idempotent and only does
// real work once the final has been played; callers gate the (heavier) top-scorer
// aggregation behind a "resolved" flag so it runs once.
package resolve

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"unicode"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Store is the storage capability the resolver needs.
type Store interface {
	ListMatches(ctx context.Context) ([]storage.Match, error)
	GetMatchByID(ctx context.Context, id int64) (storage.Match, error)
	TeamIDByFifaID(ctx context.Context, fifaID string) (int64, bool, error)
	AwardBonusByKind(ctx context.Context, kind, correctPickRef string) error
	ListTopScorers(ctx context.Context, limit int) ([]storage.ScorerRow, error)
	ListTournamentPicksByKind(ctx context.Context, kind string) ([]storage.TournamentPick, error)
	SetPickAwarded(ctx context.Context, id int64, awarded bool) error
}

// Provider fetches a match's live/finished detail (for the final's winner and
// for goal aggregation). *results.FIFAClient satisfies it.
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

// Run resolves and awards the bonuses. Returns true once the final has been
// played and bonuses were awarded (so the caller can mark it done and stop
// re-running the heavy aggregation). A no-op returning false before the final.
func (r *Resolver) Run(ctx context.Context) (bool, error) {
	matches, err := r.store.ListMatches(ctx)
	if err != nil {
		return false, err
	}
	var final *storage.Match
	for i := range matches {
		m := matches[i]
		if m.Stage == "final" && m.Status == "finished" && m.Home != nil && m.Away != nil {
			final = &matches[i]
			break
		}
	}
	if final == nil {
		return false, nil // final not played yet — nothing to resolve
	}

	// ── Champion + finalist from the final's result ──
	// Use the live detail's WinnerTeamID so penalty-shootout finals resolve
	// correctly (the scoreline alone can be a draw).
	championLocal, finalistLocal, ok := r.finalWinnerLoser(ctx, final)
	if ok {
		if err := r.store.AwardBonusByKind(ctx, "champion", strconv.FormatInt(championLocal, 10)); err != nil {
			r.log.Warn("resolve: award champion failed", slog.Any("error", err))
		}
		if err := r.store.AwardBonusByKind(ctx, "finalist", strconv.FormatInt(finalistLocal, 10)); err != nil {
			r.log.Warn("resolve: award finalist failed", slog.Any("error", err))
		}
		r.log.Info("resolve: champion/finalist awarded",
			slog.Int64("champion", championLocal), slog.Int64("finalist", finalistLocal))
	}

	// ── Top scorer by aggregating goals across all finished matches ──
	if name := r.topScorer(ctx); name != "" {
		r.awardTopScorer(ctx, name)
	}

	return true, nil
}

// finalWinnerLoser returns the local team ids of the final's winner (champion)
// and loser (finalist). The winner comes from the live detail's WinnerTeamID;
// the loser is the final's other team.
func (r *Resolver) finalWinnerLoser(ctx context.Context, final *storage.Match) (champion, finalist int64, ok bool) {
	ref, err := r.store.GetMatchByID(ctx, final.ID)
	if err != nil {
		r.log.Warn("resolve: load final refs failed", slog.Any("error", err))
		return 0, 0, false
	}
	detail, err := r.provider.LiveMatch(ctx, ref.FifaStageID, ref.FifaID)
	if err != nil || detail == nil || detail.WinnerTeamID == "" {
		r.log.Warn("resolve: final winner unavailable", slog.Any("error", err))
		return 0, 0, false
	}
	winnerLocal, found, err := r.store.TeamIDByFifaID(ctx, detail.WinnerTeamID)
	if err != nil || !found {
		r.log.Warn("resolve: winner team not found", slog.String("fifaWinner", detail.WinnerTeamID))
		return 0, 0, false
	}
	switch winnerLocal {
	case final.Home.ID:
		return final.Home.ID, final.Away.ID, true
	case final.Away.ID:
		return final.Away.ID, final.Home.ID, true
	default:
		r.log.Warn("resolve: winner is not one of the finalists", slog.Int64("winner", winnerLocal))
		return 0, 0, false
	}
}

// topScorer returns the name of the tournament's leading scorer from the
// (periodically aggregated) top-scorers board, or "" if none yet.
func (r *Resolver) topScorer(ctx context.Context) string {
	rows, err := r.store.ListTopScorers(ctx, 1)
	if err != nil {
		r.log.Warn("resolve: load top scorer failed", slog.Any("error", err))
		return ""
	}
	if len(rows) == 0 {
		return ""
	}
	r.log.Info("resolve: top scorer", slog.String("player", rows[0].Name), slog.Int("goals", rows[0].Goals))
	return rows[0].Name
}

// awardTopScorer marks each top_scorer pick awarded iff its (normalized) name
// matches the resolved top scorer — tolerant of diacritics, case and partial
// (first/last name) entries.
func (r *Resolver) awardTopScorer(ctx context.Context, actual string) {
	picks, err := r.store.ListTournamentPicksByKind(ctx, "top_scorer")
	if err != nil {
		r.log.Warn("resolve: list top_scorer picks failed", slog.Any("error", err))
		return
	}
	want := normalizeName(actual)
	for _, p := range picks {
		if err := r.store.SetPickAwarded(ctx, p.ID, namesMatch(normalizeName(p.PickRef), want)); err != nil {
			r.log.Warn("resolve: set top_scorer awarded failed", slog.Int64("pick", p.ID), slog.Any("error", err))
		}
	}
}

// namesMatch reports whether two normalized player names refer to the same
// player: exact, or one is a whole-word subset of the other (e.g. "mbappe" vs
// "kylian mbappe").
func namesMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	return wordSubset(a, b) || wordSubset(b, a)
}

// wordSubset reports whether every word in `sub` appears in `full`.
func wordSubset(sub, full string) bool {
	fw := make(map[string]bool)
	for _, w := range strings.Fields(full) {
		fw[w] = true
	}
	for _, w := range strings.Fields(sub) {
		if !fw[w] {
			return false
		}
	}
	return true
}

// normalizeName lowercases, strips diacritics and collapses whitespace.
func normalizeName(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		out = s
	}
	return strings.Join(strings.Fields(strings.ToLower(out)), " ")
}
