package announce

import (
	"strings"
	"testing"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

func ip(v int) *int { return &v }

func TestComposeMessageWithWinners(t *testing.T) {
	m := storage.Match{
		Stage:      "group",
		GroupLabel: "Group B",
		HomeScore:  ip(2),
		AwayScore:  ip(0),
		Home:       &storage.Team{Name: "Canada"},
		Away:       &storage.Team{Name: "South Africa"},
	}
	preds := []storage.MatchPrediction{
		{Nickname: "Dimon", Points: 4},
		{Nickname: "Sanya", Points: 3},
		{Nickname: "Nobody", Points: 1},
	}
	msg := composeMessage(m, preds)

	if !strings.Contains(msg, "Canada 2–0 South Africa") {
		t.Errorf("scoreline missing: %q", msg)
	}
	if !strings.Contains(msg, "Dimon") || !strings.Contains(msg, "+4") {
		t.Errorf("exact-score winner Dimon missing: %q", msg)
	}
	if !strings.Contains(msg, "Sanya") || !strings.Contains(msg, "+3") {
		t.Errorf("exact-score winner Sanya missing: %q", msg)
	}
	if strings.Contains(msg, "Nobody") {
		t.Errorf("non-exact predictor should not be congratulated: %q", msg)
	}
}

func TestComposeMessageNoWinners(t *testing.T) {
	m := storage.Match{
		Stage:     "final",
		HomeScore: ip(1),
		AwayScore: ip(1),
		Home:      &storage.Team{Name: "Brazil"},
		Away:      &storage.Team{Name: "France"},
	}
	preds := []storage.MatchPrediction{{Nickname: "Dimon", Points: 1}}
	msg := composeMessage(m, preds)

	if !strings.Contains(msg, "ніхто не вгадав") {
		t.Errorf("expected the no-winner line: %q", msg)
	}
	if !strings.Contains(msg, "Фінал") {
		t.Errorf("expected Ukrainian stage label: %q", msg)
	}
}

func TestComposeMessageEscapesHTML(t *testing.T) {
	m := storage.Match{
		Stage:     "group",
		HomeScore: ip(0),
		AwayScore: ip(0),
		Home:      &storage.Team{Name: "A"},
		Away:      &storage.Team{Name: "B"},
	}
	preds := []storage.MatchPrediction{{Nickname: "<b>x</b>", Points: 3}}
	msg := composeMessage(m, preds)
	if strings.Contains(msg, "<b>x</b>") {
		t.Errorf("nickname HTML must be escaped: %q", msg)
	}
}
