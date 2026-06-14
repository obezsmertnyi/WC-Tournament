// Package digest builds and posts the once-a-day "morning" Telegram summary:
// overnight results + current standings + the upcoming day's fixtures. The World
// Cup 2026 is in the Americas, so matches fall overnight Kyiv time; a single
// morning digest (sent at a civilized Kyiv hour) looks back at finished matches
// and forward at the ones to predict, without pinging anyone at night.
package digest

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// lookbackWindow / lookaheadWindow bound the results and fixtures shown.
const (
	lookbackWindow  = 24 * time.Hour
	lookaheadWindow = 24 * time.Hour
)

// Store is the storage capability the digest needs (both already exist).
type Store interface {
	ListMatches(ctx context.Context) ([]storage.Match, error)
	Leaderboard(ctx context.Context) ([]storage.LeaderboardRow, error)
}

// Sender posts an HTML message. *notify.Telegram satisfies this.
type Sender interface {
	Enabled() bool
	Send(ctx context.Context, htmlMsg string) error
}

// Digest ties a store to a sender + the display timezone for kickoff times.
type Digest struct {
	store  Store
	sender Sender
	loc    *time.Location
	log    *slog.Logger
}

// New builds a Digest. loc is the timezone kickoff times are shown in (Kyiv).
func New(store Store, sender Sender, loc *time.Location, log *slog.Logger) *Digest {
	if loc == nil {
		loc = time.UTC
	}
	return &Digest{store: store, sender: sender, loc: loc, log: log}
}

// Run composes and posts the morning digest. Returns true if a message was sent.
// A no-op (returns false, nil) when there is nothing to report or the sender is
// disabled — callers should still record the digest as done for the day.
func (d *Digest) Run(ctx context.Context) (bool, error) {
	matches, err := d.store.ListMatches(ctx)
	if err != nil {
		return false, err
	}
	board, err := d.store.Leaderboard(ctx)
	if err != nil {
		return false, err
	}

	now := time.Now().UTC()
	var results, upcoming []storage.Match
	for _, m := range matches {
		if m.KickoffAt == nil {
			continue
		}
		k := m.KickoffAt.UTC()
		switch {
		case m.Status == "finished" && m.HomeScore != nil && m.AwayScore != nil &&
			!k.Before(now.Add(-lookbackWindow)) && !k.After(now):
			results = append(results, m)
		case m.Status == "scheduled" && k.After(now) && !k.After(now.Add(lookaheadWindow)):
			upcoming = append(upcoming, m)
		}
	}

	// Nothing worth saying (e.g. a quiet day with no recent or upcoming games).
	if len(results) == 0 && len(upcoming) == 0 {
		return false, nil
	}
	if !d.sender.Enabled() {
		return false, nil
	}
	if err := d.sender.Send(ctx, d.compose(results, board, upcoming)); err != nil {
		return false, err
	}
	return true, nil
}

func (d *Digest) compose(results []storage.Match, board []storage.LeaderboardRow, upcoming []storage.Match) string {
	var b strings.Builder
	b.WriteString("☀️ <b>Доброго ранку! WC2026</b>\n")

	b.WriteString("\n🏁 <b>Результати:</b>\n")
	if len(results) == 0 {
		b.WriteString("Вночі матчів не було.\n")
	} else {
		for _, m := range results {
			fmt.Fprintf(&b, "%s %d–%d %s\n",
				esc(teamLabel(m.Home, m.PlaceholderHome)),
				derefScore(m.HomeScore), derefScore(m.AwayScore),
				esc(teamLabel(m.Away, m.PlaceholderAway)))
		}
	}

	if len(board) > 0 {
		b.WriteString("\n📊 <b>Таблиця:</b>\n")
		medals := []string{"🥇", "🥈", "🥉"}
		shown := 0
		for i, r := range board {
			if shown >= 8 { // small pool, but cap defensively
				break
			}
			rank := fmt.Sprintf("%d.", i+1)
			if i < len(medals) {
				rank = medals[i]
			}
			fmt.Fprintf(&b, "%s %s — <b>%d</b>\n", rank, esc(r.Nickname), r.Points)
			shown++
		}
	}

	b.WriteString("\n⚽️ <b>Сьогодні:</b>\n")
	if len(upcoming) == 0 {
		b.WriteString("Сьогодні матчів немає.\n")
	} else {
		for _, m := range upcoming {
			t := m.KickoffAt.In(d.loc).Format("15:04")
			fmt.Fprintf(&b, "%s  %s — %s\n", t,
				esc(teamLabel(m.Home, m.PlaceholderHome)),
				esc(teamLabel(m.Away, m.PlaceholderAway)))
		}
		b.WriteString("\nНе забудьте прогнози 👉 https://wc2026.mtgrd-das.app")
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

func esc(s string) string { return html.EscapeString(s) }
