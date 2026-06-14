package digest

import (
	"strings"
	"testing"
	"time"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

func ip(v int) *int { return &v }

func TestComposeDigest(t *testing.T) {
	d := New(nil, nil, time.UTC, nil)
	kick := time.Date(2026, 6, 14, 19, 0, 0, 0, time.UTC)
	results := []storage.Match{{
		Status: "finished", HomeScore: ip(2), AwayScore: ip(0),
		Home: &storage.Team{Name: "Canada"}, Away: &storage.Team{Name: "South Africa"},
		KickoffAt: &kick,
	}}
	board := []storage.LeaderboardRow{
		{Nickname: "Vova", Points: 14}, {Nickname: "Sanya", Points: 11},
	}
	up := []storage.Match{{
		Status: "scheduled", Home: &storage.Team{Name: "Mexico"}, Away: &storage.Team{Name: "Korea"},
		KickoffAt: &kick,
	}}
	msg := d.compose(results, board, up)

	for _, want := range []string{"Доброго ранку", "Canada 2–0 South Africa", "🥇 Vova", "Mexico — Korea", "19:00"} {
		if !strings.Contains(msg, want) {
			t.Errorf("digest missing %q in:\n%s", want, msg)
		}
	}
}

// Regression: more players than medals must not panic (prod had 5).
func TestComposeDigestManyPlayers(t *testing.T) {
	d := New(nil, nil, time.UTC, nil)
	board := []storage.LeaderboardRow{
		{Nickname: "A", Points: 8}, {Nickname: "B", Points: 4}, {Nickname: "C", Points: 3},
		{Nickname: "D", Points: 3}, {Nickname: "E", Points: 0},
	}
	msg := d.compose(nil, board, nil)
	if !strings.Contains(msg, "4. D") || !strings.Contains(msg, "5. E") {
		t.Errorf("expected numbered ranks beyond medals:\n%s", msg)
	}
}
