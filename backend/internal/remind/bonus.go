package remind

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// BonusLeadTime is how long before the hard lock (first Round-of-16 kickoff) the
// one-off bonus reminder fires — long enough that picking a champion is not a
// rushed decision.
const BonusLeadTime = 24 * time.Hour

// bonusRemindedStateKey guards the reminder so it fires at most once, surviving
// restarts (stored in app_state).
const bonusRemindedStateKey = "bonus_reminded"

// BonusStore is the storage capability the bonus reminder needs.
type BonusStore interface {
	// FirstRoundOf16Kickoff is the hard lock: bonuses can't be set/changed after it.
	FirstRoundOf16Kickoff(ctx context.Context) (*time.Time, error)
	ListUsersMissingBonus(ctx context.Context, demoOn bool) ([]string, error)
	IsDemoMode(ctx context.Context) (bool, error)
	GetAppState(ctx context.Context, key string) (string, bool, error)
	SetAppState(ctx context.Context, key, value string) error
}

// BonusReminder nudges players to set their tournament bonuses (champion /
// finalist / top scorer) before they lock at the start of the Round of 16 —
// the tournament-wide analog of the pre-match prediction reminder.
type BonusReminder struct {
	store  BonusStore
	sender Sender
	log    *slog.Logger
}

// NewBonus builds a BonusReminder (Sender is shared with the match reminder).
func NewBonus(store BonusStore, sender Sender, log *slog.Logger) *BonusReminder {
	if log == nil {
		log = slog.Default()
	}
	return &BonusReminder{store: store, sender: sender, log: log}
}

// Run sends the one-off bonus reminder when the lock is within BonusLeadTime.
// It is idempotent (fires at most once, guarded by app_state) and safe to call
// on the periodic tick. Returns true when a reminder was actually posted.
//
// `now` is injected for testability; callers pass time.Now().UTC(). No-ops
// (returning false) when: already fired, the R16 isn't scheduled yet, it's still
// too early, or the lock already passed (marked handled so it never fires late —
// a reminder after the deadline would be useless).
func (r *BonusReminder) Run(ctx context.Context, now time.Time) (bool, error) {
	if _, done, err := r.store.GetAppState(ctx, bonusRemindedStateKey); err != nil {
		return false, err
	} else if done {
		return false, nil
	}

	deadline, err := r.store.FirstRoundOf16Kickoff(ctx)
	if err != nil {
		return false, err
	}
	if deadline == nil {
		return false, nil // Round of 16 not scheduled yet — nothing to remind about
	}
	d := deadline.UTC()

	switch {
	case !now.Before(d):
		// Lock already passed — record that we handled it so we stop checking, and
		// never send a pointless post-deadline nudge.
		if err := r.store.SetAppState(ctx, bonusRemindedStateKey, "skipped:"+now.Format(time.RFC3339)); err != nil {
			return false, err
		}
		return false, nil
	case now.Before(d.Add(-BonusLeadTime)):
		return false, nil // too early — check again next tick
	}

	demoOn, err := r.store.IsDemoMode(ctx)
	if err != nil {
		return false, fmt.Errorf("bonus remind: read demo mode: %w", err)
	}
	missing, err := r.store.ListUsersMissingBonus(ctx, demoOn)
	if err != nil {
		return false, err
	}

	posted := false
	if len(missing) > 0 && r.sender.Enabled() {
		if err := r.sender.Send(ctx, composeBonusReminder(d, missing)); err != nil {
			// Leave the flag unset so the next tick retries within the window.
			return false, fmt.Errorf("bonus remind: telegram send: %w", err)
		}
		posted = true
	}

	// Mark fired once we've handled this window (even if nobody was missing or
	// Telegram is off) so it never repeats.
	if err := r.store.SetAppState(ctx, bonusRemindedStateKey, now.Format(time.RFC3339)); err != nil {
		return false, err
	}
	return posted, nil
}

// composeBonusReminder builds the friendly Ukrainian bonus-deadline nudge. The
// deadline is shown in Kyiv time (tzdata is embedded in the binary).
func composeBonusReminder(deadline time.Time, missing []string) string {
	when := deadline.UTC()
	if loc, err := time.LoadLocation("Europe/Kyiv"); err == nil {
		when = deadline.In(loc)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>Останній шанс проставити турнірні бонуси</b>\n")
	fmt.Fprintf(&b, "Бонуси (чемпіон / фіналіст / бомбардир) закриваються зі стартом 1/8 фіналу — <b>%s</b>. Після цього змінити не можна.\n",
		esc(when.Format("02.01 о 15:04")))
	b.WriteString("\nЩе не проставили всі бонуси:\n")
	for _, n := range missing {
		fmt.Fprintf(&b, "• %s\n", esc(n))
	}
	b.WriteString("\nВстигніть до старту плей-оф ↗\nhttps://wc2026.mtgrd-das.app")
	return b.String()
}
