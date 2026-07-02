// Package remind sends a one-off Telegram nudge before a match kicks off,
// listing the players who still haven't entered a prediction. It is idempotent:
// each match is reminded at most once (guarded by matches.reminded_at), so it is
// safe to call on a short timer.
package remind

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// LeadTime is how long before kickoff the reminder fires (the user wants ~1h).
const LeadTime = time.Hour

// Store is the storage capability the reminder needs.
type Store interface {
	ListUpcomingUnremindedMatches(ctx context.Context, window time.Duration) ([]storage.Match, error)
	ListUsersMissingPrediction(ctx context.Context, matchID int64, demoOn bool) ([]string, error)
	IsDemoMode(ctx context.Context) (bool, error)
	MarkMatchReminded(ctx context.Context, id int64) error
}

// Sender posts an HTML message. *notify.Telegram satisfies this.
type Sender interface {
	Enabled() bool
	Send(ctx context.Context, htmlMsg string) error
}

// Reminder ties a store to a sender.
type Reminder struct {
	store  Store
	sender Sender
	log    *slog.Logger
}

// New builds a Reminder.
func New(store Store, sender Sender, log *slog.Logger) *Reminder {
	return &Reminder{store: store, sender: sender, log: log}
}

// Run sends a reminder for every match kicking off within LeadTime that hasn't
// been reminded yet. Matches where everyone already predicted are marked
// reminded without posting. Per-match send failures are logged and left
// unreminded to retry next tick.
func (r *Reminder) Run(ctx context.Context) (int, error) {
	matches, err := r.store.ListUpcomingUnremindedMatches(ctx, LeadTime)
	if err != nil {
		return 0, err
	}
	// In demo mode only the rw tier can submit predictions, so the reminder must
	// not nudge none/ro (preview-only) users who couldn't predict anyway.
	demoOn, err := r.store.IsDemoMode(ctx)
	if err != nil {
		return 0, fmt.Errorf("remind: read demo mode: %w", err)
	}

	sent := 0
	for _, m := range matches {
		missing, err := r.store.ListUsersMissingPrediction(ctx, m.ID, demoOn)
		if err != nil {
			r.log.Warn("remind: load missing predictors failed", slog.Int64("match", m.ID), slog.Any("error", err))
			continue
		}

		// Only post when someone is actually missing a pick and Telegram is on.
		if len(missing) > 0 && r.sender.Enabled() {
			if err := r.sender.Send(ctx, composeReminder(m, missing)); err != nil {
				r.log.Warn("remind: telegram send failed", slog.Int64("match", m.ID), slog.Any("error", err))
				continue // leave unreminded; retry next tick
			}
			sent++
		}

		if err := r.store.MarkMatchReminded(ctx, m.ID); err != nil {
			r.log.Warn("remind: mark reminded failed", slog.Int64("match", m.ID), slog.Any("error", err))
		}
	}
	return sent, nil
}

// composeReminder builds the friendly Ukrainian pre-match nudge.
func composeReminder(m storage.Match, missing []string) string {
	home := teamLabel(m.Home, m.PlaceholderHome)
	away := teamLabel(m.Away, m.PlaceholderAway)

	var b strings.Builder
	fmt.Fprintf(&b, "⏰ <b>%s — %s</b> вже за годину!\n", esc(home), esc(away))
	if g := strings.TrimSpace(m.GroupLabel); g != "" {
		b.WriteString("<i>" + esc(g) + "</i>\n")
	}
	b.WriteString("\nЩе не зробили прогноз:\n")
	for _, n := range missing {
		fmt.Fprintf(&b, "• %s\n", esc(n))
	}
	b.WriteString("\nВстигніть до стартового свистка ⚽️\n👉 https://wc2026.mtgrd-das.app")
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

func esc(s string) string { return html.EscapeString(s) }
