package api

import (
	"testing"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// TestMatchFactOutcome pins the explicit winner/advanced resolution that grounds
// the AI assistant — previously the model had to infer the result from the raw
// score and got it backwards (reported teams as losing matches they won).
func TestMatchFactOutcome(t *testing.T) {
	colombia := &storage.Team{ID: 48, Name: "Colombia"}
	congo := &storage.Team{ID: 42, Name: "Congo DR"}
	uzb := &storage.Team{ID: 47, Name: "Uzbekistan"}
	arg := &storage.Team{ID: 37, Name: "Argentina"}
	cpv := &storage.Team{ID: 40, Name: "Cabo Verde"}

	sc := func(h, a int) (*int, *int) { return &h, &a }

	tests := []struct {
		name         string
		m            storage.Match
		wantScore    string
		wantWinner   string
		wantAdvanced string
	}{
		{
			name:       "home win",
			m:          finishedMatch("group", colombia, congo, 1, 0, nil),
			wantScore:  "1:0",
			wantWinner: "Colombia",
		},
		{
			name:       "away win (score is home:away, winner is the away side)",
			m:          finishedMatch("group", uzb, colombia, 1, 3, nil),
			wantScore:  "1:3",
			wantWinner: "Colombia",
		},
		{
			name:       "group draw has no winner",
			m:          finishedMatch("group", colombia, arg, 0, 0, nil),
			wantScore:  "0:0",
			wantWinner: "draw",
		},
		{
			// ARG 1:1 CPV, Argentina advanced (ET/penalties): the scoreline is a
			// draw but the advancer is Argentina — the two must not be conflated.
			name:         "knockout draw with advancer",
			m:            finishedMatch("r32", arg, cpv, 1, 1, &arg.ID),
			wantScore:    "1:1",
			wantWinner:   "draw",
			wantAdvanced: "Argentina",
		},
		{
			name: "live match has a score but no winner yet",
			m: func() storage.Match {
				h, a := sc(1, 0)
				return storage.Match{Stage: "r16", Status: "live", Home: colombia, Away: congo, HomeScore: h, AwayScore: a}
			}(),
			wantScore:  "1:0",
			wantWinner: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := matchFact(tt.m)
			if f.Score != tt.wantScore {
				t.Errorf("Score: got %q want %q", f.Score, tt.wantScore)
			}
			if f.Winner != tt.wantWinner {
				t.Errorf("Winner: got %q want %q", f.Winner, tt.wantWinner)
			}
			if f.Advanced != tt.wantAdvanced {
				t.Errorf("Advanced: got %q want %q", f.Advanced, tt.wantAdvanced)
			}
		})
	}
}

// TestMatchFactResolution pins how the stored result_detail code becomes the
// grounding hint the AI uses to say exactly how a level knockout was decided
// (instead of guessing "probably ET or penalties").
func TestMatchFactResolution(t *testing.T) {
	arg := &storage.Team{ID: 37, Name: "Argentina"}
	cpv := &storage.Team{ID: 40, Name: "Cabo Verde"}
	mk := func(detail string) storage.Match {
		m := finishedMatch("r32", arg, cpv, 1, 1, &arg.ID)
		m.ResultDetail = detail
		return m
	}
	tests := []struct{ detail, wantRes, wantScore string }{
		{"et:3:2", "extra_time", "3:2"},
		{"pen:4:2", "penalties", "4:2"},
		{"", "", ""},
	}
	for _, tt := range tests {
		f := matchFact(mk(tt.detail))
		if f.Resolution != tt.wantRes || f.ResolutionScore != tt.wantScore {
			t.Errorf("detail %q -> %q/%q, want %q/%q", tt.detail, f.Resolution, f.ResolutionScore, tt.wantRes, tt.wantScore)
		}
	}
}

func finishedMatch(stage string, home, away *storage.Team, h, a int, winner *int64) storage.Match {
	return storage.Match{
		Stage:        stage,
		Status:       "finished",
		Home:         home,
		Away:         away,
		HomeScore:    &h,
		AwayScore:    &a,
		WinnerTeamID: winner,
	}
}
