package remind

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type fakeBonusStore struct {
	deadline *time.Time
	missing  []string
	demoOn   bool
	state    map[string]string
}

func (f *fakeBonusStore) FirstRoundOf16Kickoff(context.Context) (*time.Time, error) {
	return f.deadline, nil
}
func (f *fakeBonusStore) ListUsersMissingBonus(_ context.Context, _ bool) ([]string, error) {
	return f.missing, nil
}
func (f *fakeBonusStore) IsDemoMode(context.Context) (bool, error) { return f.demoOn, nil }
func (f *fakeBonusStore) GetAppState(_ context.Context, key string) (string, bool, error) {
	v, ok := f.state[key]
	return v, ok, nil
}
func (f *fakeBonusStore) SetAppState(_ context.Context, key, value string) error {
	f.state[key] = value
	return nil
}

type fakeSender struct {
	enabled bool
	err     error
	sent    []string
}

func (f *fakeSender) Enabled() bool { return f.enabled }
func (f *fakeSender) Send(_ context.Context, msg string) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, msg)
	return nil
}

func at(y int, mo time.Month, d, h int) time.Time { return time.Date(y, mo, d, h, 0, 0, 0, time.UTC) }

func newBonus(store BonusStore, s Sender) *BonusReminder {
	return NewBonus(store, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestBonusReminder(t *testing.T) {
	deadline := at(2026, 7, 4, 17) // first R16 kickoff (UTC)
	dl := deadline

	t.Run("fires within the lead window when someone is missing", func(t *testing.T) {
		store := &fakeBonusStore{deadline: &dl, missing: []string{"Mihulin", "Yevhen"}, state: map[string]string{}}
		snd := &fakeSender{enabled: true}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline.Add(-2*time.Hour))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !posted || len(snd.sent) != 1 {
			t.Fatalf("expected one post, got posted=%v sent=%d", posted, len(snd.sent))
		}
		if !strings.Contains(snd.sent[0], "Mihulin") {
			t.Errorf("message should list the missing player: %q", snd.sent[0])
		}
		if _, ok := store.state[bonusRemindedStateKey]; !ok {
			t.Error("reminder should be marked sent")
		}
	})

	t.Run("does not fire twice", func(t *testing.T) {
		store := &fakeBonusStore{deadline: &dl, missing: []string{"Mihulin"}, state: map[string]string{bonusRemindedStateKey: "2026-07-04T00:00:00Z"}}
		snd := &fakeSender{enabled: true}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline.Add(-2*time.Hour))
		if err != nil || posted || len(snd.sent) != 0 {
			t.Fatalf("already-sent must no-op: posted=%v sent=%d err=%v", posted, len(snd.sent), err)
		}
	})

	t.Run("too early: no-op, not marked", func(t *testing.T) {
		store := &fakeBonusStore{deadline: &dl, missing: []string{"Mihulin"}, state: map[string]string{}}
		snd := &fakeSender{enabled: true}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline.Add(-30*time.Hour))
		if err != nil || posted || len(snd.sent) != 0 {
			t.Fatalf("too-early must no-op: posted=%v sent=%d", posted, len(snd.sent))
		}
		if _, ok := store.state[bonusRemindedStateKey]; ok {
			t.Error("too-early must NOT mark sent (so it can fire later)")
		}
	})

	t.Run("deadline passed: skip, mark handled, never send late", func(t *testing.T) {
		store := &fakeBonusStore{deadline: &dl, missing: []string{"Mihulin"}, state: map[string]string{}}
		snd := &fakeSender{enabled: true}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline.Add(1*time.Hour))
		if err != nil || posted || len(snd.sent) != 0 {
			t.Fatalf("post-deadline must not send: posted=%v sent=%d", posted, len(snd.sent))
		}
		if v := store.state[bonusRemindedStateKey]; !strings.HasPrefix(v, "skipped:") {
			t.Errorf("post-deadline should mark skipped, got %q", v)
		}
	})

	t.Run("everyone has bonuses: mark sent, no post", func(t *testing.T) {
		store := &fakeBonusStore{deadline: &dl, missing: nil, state: map[string]string{}}
		snd := &fakeSender{enabled: true}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline.Add(-2*time.Hour))
		if err != nil || posted || len(snd.sent) != 0 {
			t.Fatalf("nobody missing => no post: posted=%v sent=%d", posted, len(snd.sent))
		}
		if _, ok := store.state[bonusRemindedStateKey]; !ok {
			t.Error("should still mark sent so it fires once")
		}
	})

	t.Run("R16 not scheduled: no-op", func(t *testing.T) {
		store := &fakeBonusStore{deadline: nil, missing: []string{"Mihulin"}, state: map[string]string{}}
		snd := &fakeSender{enabled: true}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline)
		if err != nil || posted {
			t.Fatalf("no deadline => no-op: posted=%v err=%v", posted, err)
		}
	})

	t.Run("telegram send fails: leave unmarked to retry", func(t *testing.T) {
		store := &fakeBonusStore{deadline: &dl, missing: []string{"Mihulin"}, state: map[string]string{}}
		snd := &fakeSender{enabled: true, err: errors.New("telegram down")}
		posted, err := newBonus(store, snd).Run(context.Background(), deadline.Add(-2*time.Hour))
		if err == nil || posted {
			t.Fatalf("send failure should surface an error and not post")
		}
		if _, ok := store.state[bonusRemindedStateKey]; ok {
			t.Error("must NOT mark sent on send failure (retry next tick)")
		}
	})
}

func TestComposeBonusReminder(t *testing.T) {
	msg := composeBonusReminder(at(2026, 7, 4, 17), []string{"Vova", "<b>x</b>"})
	if !strings.Contains(msg, "1/8") {
		t.Errorf("should mention the R16 lock: %q", msg)
	}
	if !strings.Contains(msg, "Vova") {
		t.Errorf("should list the missing player: %q", msg)
	}
	if strings.Contains(msg, "<b>x</b>") {
		t.Errorf("nickname must be escaped: %q", msg)
	}
}
