package standings

import (
	"reflect"
	"testing"
)

func i64(v int64) *int64 { return &v }
func ip(v int) *int      { return &v }

// m is a terse finished-match constructor for tests.
func m(home, away int64, hs, as int) Match {
	return Match{HomeTeamID: i64(home), AwayTeamID: i64(away), HomeScore: ip(hs), AwayScore: ip(as)}
}

func TestComputeStandings(t *testing.T) {
	// Group A: 4 teams. Two results played.
	//   Mexico(1) 2-0 South Africa(2)   -> MEX +3, RSA +0
	//   Canada(3) 1-1 Uruguay(4)        -> draw, both +1
	// Expected order by points desc, gd desc, gf desc, name asc:
	//   1. Mexico   (3 pts, gd +2)
	//   2. Canada   (1 pt,  gd 0, name "Canada" < "Uruguay")
	//   3. Uruguay  (1 pt,  gd 0)
	//   4. South Africa (0 pts, gd -2)
	teams := []Team{
		{ID: 1, Name: "Mexico", Code: "MEX", FlagURL: "f/MEX", GroupLabel: "Group A"},
		{ID: 2, Name: "South Africa", Code: "RSA", FlagURL: "f/RSA", GroupLabel: "Group A"},
		{ID: 3, Name: "Canada", Code: "CAN", FlagURL: "f/CAN", GroupLabel: "Group A"},
		{ID: 4, Name: "Uruguay", Code: "URU", FlagURL: "f/URU", GroupLabel: "Group A"},
		// A second group, single team, no matches — must still appear with zeros.
		{ID: 5, Name: "Brazil", Code: "BRA", FlagURL: "f/BRA", GroupLabel: "Group B"},
		// Unassigned team — excluded entirely.
		{ID: 99, Name: "TBD", Code: "", FlagURL: "", GroupLabel: ""},
	}
	finished := []Match{
		m(1, 2, 2, 0),
		m(3, 4, 1, 1),
	}

	got := ComputeStandings(teams, finished)

	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %d (%+v)", len(got), got)
	}

	// Groups ordered A then B.
	if got[0].Group != "A" || got[1].Group != "B" {
		t.Fatalf("group order: got %q,%q want A,B", got[0].Group, got[1].Group)
	}

	wantA := []Row{
		{TeamID: 1, Name: "Mexico", Code: "MEX", FlagURL: "f/MEX", Played: 1, Win: 1, Draw: 0, Loss: 0, GF: 2, GA: 0, GD: 2, Points: 3, Rank: 1},
		{TeamID: 3, Name: "Canada", Code: "CAN", FlagURL: "f/CAN", Played: 1, Win: 0, Draw: 1, Loss: 0, GF: 1, GA: 1, GD: 0, Points: 1, Rank: 2},
		{TeamID: 4, Name: "Uruguay", Code: "URU", FlagURL: "f/URU", Played: 1, Win: 0, Draw: 1, Loss: 0, GF: 1, GA: 1, GD: 0, Points: 1, Rank: 3},
		{TeamID: 2, Name: "South Africa", Code: "RSA", FlagURL: "f/RSA", Played: 1, Win: 0, Draw: 0, Loss: 1, GF: 0, GA: 2, GD: -2, Points: 0, Rank: 4},
	}
	if !reflect.DeepEqual(got[0].Rows, wantA) {
		t.Errorf("group A rows mismatch:\n got=%+v\nwant=%+v", got[0].Rows, wantA)
	}

	// Group B: single team, zero matches played, rank 1.
	wantB := []Row{
		{TeamID: 5, Name: "Brazil", Code: "BRA", FlagURL: "f/BRA", Rank: 1},
	}
	if !reflect.DeepEqual(got[1].Rows, wantB) {
		t.Errorf("group B rows mismatch:\n got=%+v\nwant=%+v", got[1].Rows, wantB)
	}
}

func TestComputeStandings_TieBreakGDoverGF(t *testing.T) {
	// Both teams 1 win each (3 pts), but differing GD then GF.
	//   A(1) beats B(2) 3-0  -> A: gf3 ga0 gd+3 ; B: gf0 ga3 gd-3
	//   C(3) beats D(4) 1-0  -> C: gf1 ga0 gd+1 ; D: gf0 ga1 gd-1
	//   A(1) beats C(3) 1-0  -> A: +gf1 ga0 ; C: +gf0 ga1
	// After: A pts6 gd+4 gf4 ; C pts3 gd0 gf1 ; D pts0 gd-1 ; B pts0 gd-3
	// Order: A, C, D, B (D above B on GD).
	teams := []Team{
		{ID: 1, Name: "Alpha", GroupLabel: "Group A"},
		{ID: 2, Name: "Bravo", GroupLabel: "Group A"},
		{ID: 3, Name: "Charlie", GroupLabel: "Group A"},
		{ID: 4, Name: "Delta", GroupLabel: "Group A"},
	}
	finished := []Match{
		m(1, 2, 3, 0),
		m(3, 4, 1, 0),
		m(1, 3, 1, 0),
	}
	got := ComputeStandings(teams, finished)
	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got))
	}
	gotOrder := make([]int64, len(got[0].Rows))
	for i, r := range got[0].Rows {
		gotOrder[i] = r.TeamID
	}
	wantOrder := []int64{1, 3, 4, 2}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Errorf("rank order by id: got %v want %v\nrows=%+v", gotOrder, wantOrder, got[0].Rows)
	}
	// Spot-check rank assignment is 1..n contiguous.
	for i, r := range got[0].Rows {
		if r.Rank != i+1 {
			t.Errorf("row %d rank: got %d want %d", i, r.Rank, i+1)
		}
	}
}

func TestComputeStandings_SkipsCrossGroupAndIncomplete(t *testing.T) {
	teams := []Team{
		{ID: 1, Name: "Alpha", GroupLabel: "Group A"},
		{ID: 2, Name: "Bravo", GroupLabel: "Group B"},
	}
	finished := []Match{
		m(1, 2, 5, 0), // cross-group: must be ignored
		{HomeTeamID: i64(1), AwayTeamID: i64(1), HomeScore: ip(1)}, // missing away score: ignored
	}
	got := ComputeStandings(teams, finished)
	for _, g := range got {
		for _, r := range g.Rows {
			if r.Played != 0 {
				t.Errorf("group %s team %d should have 0 played, got %d", g.Group, r.TeamID, r.Played)
			}
		}
	}
}

func TestBareGroup(t *testing.T) {
	cases := map[string]string{
		"Group A":     "A",
		"group b":     "b",
		"GROUP C":     "C",
		"  Group D  ": "D",
		"A":           "A",
		"":            "",
		"Grouping":    "Grouping", // not the prefix "group "
	}
	for in, want := range cases {
		if got := bareGroup(in); got != want {
			t.Errorf("bareGroup(%q): got %q want %q", in, got, want)
		}
	}
}

func TestThirdPlaceRanking(t *testing.T) {
	// Two groups of 3 so each resolves a clear 3rd place; group A's third has
	// more points than group B's third, so A's third must rank first.
	teams := []Team{
		{ID: 1, Name: "A1", GroupLabel: "Group A"},
		{ID: 2, Name: "A2", GroupLabel: "Group A"},
		{ID: 3, Name: "A3", GroupLabel: "Group A"},
		{ID: 4, Name: "B1", GroupLabel: "Group B"},
		{ID: 5, Name: "B2", GroupLabel: "Group B"},
		{ID: 6, Name: "B3", GroupLabel: "Group B"},
	}
	finished := []Match{
		// Group A: A1 beats A3, A2 beats A3, A1 draws A2 -> A3 is 3rd with 0 pts
		// but scores a goal so has GF=1 (loses 1-2 twice).
		{HomeTeamID: i64(1), AwayTeamID: i64(3), HomeScore: ip(2), AwayScore: ip(1)},
		{HomeTeamID: i64(2), AwayTeamID: i64(3), HomeScore: ip(2), AwayScore: ip(1)},
		{HomeTeamID: i64(1), AwayTeamID: i64(2), HomeScore: ip(0), AwayScore: ip(0)},
		// Group B: B3 also loses both but by larger margins (worse GD/GF).
		{HomeTeamID: i64(4), AwayTeamID: i64(6), HomeScore: ip(3), AwayScore: ip(0)},
		{HomeTeamID: i64(5), AwayTeamID: i64(6), HomeScore: ip(3), AwayScore: ip(0)},
		{HomeTeamID: i64(4), AwayTeamID: i64(5), HomeScore: ip(0), AwayScore: ip(0)},
	}
	groups := ComputeStandings(teams, finished)
	thirds := ThirdPlaceRanking(groups)

	if len(thirds) != 2 {
		t.Fatalf("expected 2 third-placed teams, got %d", len(thirds))
	}
	if thirds[0].TeamID != 3 {
		t.Errorf("expected A3 (id 3) ranked first by GD/GF, got id %d", thirds[0].TeamID)
	}
	if thirds[0].Rank != 1 || thirds[1].Rank != 2 {
		t.Errorf("ranks not assigned 1..n: %d, %d", thirds[0].Rank, thirds[1].Rank)
	}
	// With only 2 third-placed teams (< 8 slots) both qualify.
	if !thirds[0].Qualified || !thirds[1].Qualified {
		t.Errorf("expected both within %d slots to qualify", ThirdPlaceQualifiers)
	}
	if thirds[0].Group != "A" {
		t.Errorf("expected group A, got %q", thirds[0].Group)
	}
}
