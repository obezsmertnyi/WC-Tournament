// Package announce posts finished-match results to the Telegram group and
// congratulates anyone who nailed the exact score. It is idempotent: each match
// is announced at most once (guarded by matches.announced_at), so it is safe to
// call after every sync.
package announce

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// exactScoreThreshold is the points floor that means a player nailed the exact
// score: 3 for a group match, 4 for a knockout (exact 3 + advancer 1).
const exactScoreThreshold = 3

// Store is the storage capability the announcer needs.
type Store interface {
	ListUnannouncedFinishedMatches(ctx context.Context) ([]storage.Match, error)
	ListPredictionsByMatch(ctx context.Context, matchID int64) ([]storage.MatchPrediction, error)
	MarkMatchAnnounced(ctx context.Context, id int64) error
}

// Sender posts an HTML message. *notify.Telegram satisfies this.
type Sender interface {
	Enabled() bool
	Send(ctx context.Context, htmlMsg string) error
}

// Announcer ties a store to a sender.
type Announcer struct {
	store  Store
	sender Sender
	log    *slog.Logger
}

// New builds an Announcer.
func New(store Store, sender Sender, log *slog.Logger) *Announcer {
	return &Announcer{store: store, sender: sender, log: log}
}

// Run announces every finished-but-unannounced match. On a disabled sender it
// still marks matches announced so a later enable doesn't dump the whole backlog.
// Per-match failures are logged and skipped; the match is left unannounced to
// retry next run.
func (a *Announcer) Run(ctx context.Context) (int, error) {
	matches, err := a.store.ListUnannouncedFinishedMatches(ctx)
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, m := range matches {
		preds, err := a.store.ListPredictionsByMatch(ctx, m.ID)
		if err != nil {
			a.log.Warn("announce: load predictions failed", slog.Int64("match", m.ID), slog.Any("error", err))
			continue
		}

		if a.sender.Enabled() {
			if err := a.sender.Send(ctx, composeMessage(m, preds)); err != nil {
				a.log.Warn("announce: telegram send failed", slog.Int64("match", m.ID), slog.Any("error", err))
				continue // leave unannounced; retry next run
			}
			sent++
		}

		if err := a.store.MarkMatchAnnounced(ctx, m.ID); err != nil {
			a.log.Warn("announce: mark announced failed", slog.Int64("match", m.ID), slog.Any("error", err))
		}
	}
	return sent, nil
}

// composeMessage builds the friendly Ukrainian result post for one match.
func composeMessage(m storage.Match, preds []storage.MatchPrediction) string {
	home := teamLabel(m.Home, m.PlaceholderHome)
	away := teamLabel(m.Away, m.PlaceholderAway)
	hs, as := derefScore(m.HomeScore), derefScore(m.AwayScore)

	var b strings.Builder
	fmt.Fprintf(&b, "🏁 <b>%s %d–%d %s</b>\n", esc(home), hs, as, esc(away))
	b.WriteString("<i>" + esc(stageLabel(m)) + "</i>\n")

	// Collect everyone who nailed the exact score (3 or 4 points).
	var winners []storage.MatchPrediction
	for _, p := range preds {
		if p.Points >= exactScoreThreshold {
			winners = append(winners, p)
		}
	}

	b.WriteString("\n")
	if len(winners) == 0 {
		b.WriteString("🤷 Цього разу ніхто не вгадав точний рахунок. Наступного разу пощастить!")
	} else {
		b.WriteString("🎯 <b>Точно вгадали рахунок:</b>\n")
		for _, w := range winners {
			fmt.Fprintf(&b, "🏆 %s <b>+%d</b>\n", esc(w.Nickname), w.Points)
		}
		b.WriteString("\nРеспект! 👏")
	}
	return b.String()
}

func teamLabel(t *storage.Team, placeholder string) string {
	if t != nil && t.Name != "" {
		return t.Name
	}
	if placeholder != "" {
		return placeholder
	}
	return "TBD"
}

func derefScore(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// stageLabel renders a short Ukrainian stage caption.
func stageLabel(m storage.Match) string {
	switch m.Stage {
	case "group":
		g := strings.TrimSpace(m.GroupLabel)
		if g != "" {
			return g // already e.g. "Group A"; kept compact
		}
		return "Груповий етап"
	case "r32":
		return "1/16 фіналу"
	case "r16":
		return "1/8 фіналу"
	case "qf":
		return "Чвертьфінал"
	case "sf":
		return "Півфінал"
	case "third":
		return "Матч за 3-тє місце"
	case "final":
		return "Фінал"
	default:
		return "Матч"
	}
}

func esc(s string) string { return html.EscapeString(s) }
