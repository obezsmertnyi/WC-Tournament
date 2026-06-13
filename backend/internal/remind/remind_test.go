package remind

import (
	"strings"
	"testing"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

func TestComposeReminder(t *testing.T) {
	m := storage.Match{
		GroupLabel: "Group B",
		Home:       &storage.Team{Name: "Canada"},
		Away:       &storage.Team{Name: "South Africa"},
	}
	msg := composeReminder(m, []string{"Vova", "Sanya"})

	if !strings.Contains(msg, "Canada — South Africa") {
		t.Errorf("teams missing: %q", msg)
	}
	if !strings.Contains(msg, "за годину") {
		t.Errorf("lead-time phrasing missing: %q", msg)
	}
	if !strings.Contains(msg, "Vova") || !strings.Contains(msg, "Sanya") {
		t.Errorf("missing predictors not listed: %q", msg)
	}
}

func TestComposeReminderEscapes(t *testing.T) {
	m := storage.Match{Home: &storage.Team{Name: "A"}, Away: &storage.Team{Name: "B"}}
	msg := composeReminder(m, []string{"<b>x</b>"})
	if strings.Contains(msg, "<b>x</b>") {
		t.Errorf("nickname must be escaped: %q", msg)
	}
}
