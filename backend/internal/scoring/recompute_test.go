package scoring

import (
	"testing"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

func p64(v int64) *int64 { return &v }
func pi(v int) *int      { return &v }

func TestAdvancer(t *testing.T) {
	const home, away = int64(10), int64(20)

	// ET/penalties: 90-min scoreline is a draw, but the resolved winner stands.
	got := advancer(storage.MatchScoringRow{
		HomeScore: pi(1), AwayScore: pi(1),
		HomeTeamID: p64(home), AwayTeamID: p64(away),
		WinnerTeamID: p64(home),
	})
	if got == nil || *got != home {
		t.Errorf("draw + resolved winner: want %d, got %v", home, got)
	}

	// No resolved winner, non-draw scoreline → scoreline winner.
	got = advancer(storage.MatchScoringRow{
		HomeScore: pi(2), AwayScore: pi(0),
		HomeTeamID: p64(home), AwayTeamID: p64(away),
	})
	if got == nil || *got != home {
		t.Errorf("non-draw fallback: want %d, got %v", home, got)
	}

	// Draw, no resolved winner → unknown (nil), +1 doesn't score yet.
	got = advancer(storage.MatchScoringRow{
		HomeScore: pi(1), AwayScore: pi(1),
		HomeTeamID: p64(home), AwayTeamID: p64(away),
	})
	if got != nil {
		t.Errorf("draw + no winner: want nil, got %v", got)
	}

	// Resolved winner overrides even a non-draw scoreline (authoritative).
	got = advancer(storage.MatchScoringRow{
		HomeScore: pi(2), AwayScore: pi(0),
		HomeTeamID: p64(home), AwayTeamID: p64(away),
		WinnerTeamID: p64(away),
	})
	if got == nil || *got != away {
		t.Errorf("resolved winner authoritative: want %d, got %v", away, got)
	}
}
