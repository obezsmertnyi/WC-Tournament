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
