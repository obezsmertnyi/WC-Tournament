// Package standings computes group-stage tables from finished group matches.
//
// This is a DISPLAY-ONLY computation: it ranks teams within their group for
// presentation but does NOT decide qualification or knockout seeding (see
// ADR-0007 — qualification is resolved elsewhere). The core ComputeStandings
// function is pure (no DB, no clock) so it is fully unit-testable.
package standings

import (
	"sort"
	"strings"
)

// Points awarded per result.
const (
	pointsWin  = 3
	pointsDraw = 1
	pointsLoss = 0
)

// Team is the minimal team input ComputeStandings needs. GroupLabel is the
// verbatim stored label (e.g. "Group A"); Group on the output is the bare
// letter.
type Team struct {
	ID         int64
	Name       string
	Code       string
	FlagURL    string
	GroupLabel string
}

// Match is a finished group match input. HomeTeamID / AwayTeamID reference
// Team.ID. Scores are required (caller passes only finished matches with both
// scores recorded); a match missing either team or score is skipped defensively.
type Match struct {
	HomeTeamID *int64
	AwayTeamID *int64
	HomeScore  *int
	AwayScore  *int
}

// Row is one team's computed line in a group table.
type Row struct {
	TeamID  int64
	Name    string
	Code    string
	FlagURL string
	Played  int
	Win     int
	Draw    int
	Loss    int
	GF      int // goals for
	GA      int // goals against
	GD      int // goal difference (GF - GA)
	Points  int
	Rank    int // 1..n within the group, assigned after sorting
}

// GroupStanding is a single group's ordered table.
type GroupStanding struct {
	Group string // bare letter, e.g. "A"
	Rows  []Row
}

// ComputeStandings builds display-only group tables. Every team with a group
// label is included (even with zero matches played) so the full table renders
// before any kickoff. Groups are ordered A->Z; within a group, rows are ordered
// by points desc, goal difference desc, goals-for desc, then name asc. Rank is
// assigned 1..n after sorting.
//
// It is pure: same inputs always produce the same output, with no I/O.
func ComputeStandings(teams []Team, finished []Match) []GroupStanding {
	// Per-group accumulator and team->group index for fast match attribution.
	type acc struct {
		group string
		rows  map[int64]*Row // keyed by team id
	}
	groups := make(map[string]*acc)
	teamGroup := make(map[int64]string)

	for _, t := range teams {
		g := bareGroup(t.GroupLabel)
		if g == "" {
			continue // team not assigned to a group — excluded from group tables
		}
		a := groups[g]
		if a == nil {
			a = &acc{group: g, rows: make(map[int64]*Row)}
			groups[g] = a
		}
		if _, ok := a.rows[t.ID]; !ok {
			a.rows[t.ID] = &Row{
				TeamID:  t.ID,
				Name:    t.Name,
				Code:    t.Code,
				FlagURL: t.FlagURL,
			}
			teamGroup[t.ID] = g
		}
	}

	for _, m := range finished {
		if m.HomeTeamID == nil || m.AwayTeamID == nil || m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		hg, hok := teamGroup[*m.HomeTeamID]
		ag, aok := teamGroup[*m.AwayTeamID]
		// Both teams must belong to the same known group for the match to count.
		if !hok || !aok || hg != ag {
			continue
		}
		rows := groups[hg].rows
		home := rows[*m.HomeTeamID]
		away := rows[*m.AwayTeamID]

		hs, as := *m.HomeScore, *m.AwayScore
		home.Played++
		away.Played++
		home.GF += hs
		home.GA += as
		away.GF += as
		away.GA += hs

		switch {
		case hs > as:
			home.Win++
			home.Points += pointsWin
			away.Loss++
			away.Points += pointsLoss
		case hs < as:
			away.Win++
			away.Points += pointsWin
			home.Loss++
			home.Points += pointsLoss
		default:
			home.Draw++
			away.Draw++
			home.Points += pointsDraw
			away.Points += pointsDraw
		}
	}

	out := make([]GroupStanding, 0, len(groups))
	for _, a := range groups {
		rows := make([]Row, 0, len(a.rows))
		for _, r := range a.rows {
			r.GD = r.GF - r.GA
			rows = append(rows, *r)
		}
		sortRows(rows)
		for i := range rows {
			rows[i].Rank = i + 1
		}
		out = append(out, GroupStanding{Group: a.group, Rows: rows})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Group < out[j].Group })
	return out
}

// ThirdPlaceQualifiers is how many of the third-placed teams advance to the
// Round of 32 under the 48-team (12-group) WC2026 format.
const ThirdPlaceQualifiers = 8

// ThirdPlaceRow is one third-placed team's line in the cross-group ranking.
// It embeds the team's full group Row and adds the group letter plus whether
// the team currently sits in a qualifying position.
type ThirdPlaceRow struct {
	Row
	Group     string `json:"group"`
	Qualified bool   `json:"qualified"`
}

// ThirdPlaceRanking takes the per-group standings and ranks the third-placed
// team of each group against the others, per the WC2026 criteria (points desc,
// goal difference desc, goals-for desc; conduct / FIFA ranking are not modelled
// so a stable name tie-break stands in last). The top ThirdPlaceQualifiers rows
// are flagged Qualified. Groups without a resolved third place (fewer than three
// teams) are skipped. The returned Rank is the 1..n cross-group position.
//
// This is DISPLAY-ONLY: it mirrors the official "ranking of third-placed teams"
// table and does not itself seed the knockout bracket (see ADR-0007).
func ThirdPlaceRanking(groups []GroupStanding) []ThirdPlaceRow {
	thirds := make([]ThirdPlaceRow, 0, len(groups))
	for _, g := range groups {
		for _, r := range g.Rows {
			if r.Rank == 3 {
				thirds = append(thirds, ThirdPlaceRow{Row: r, Group: g.Group})
				break
			}
		}
	}

	sort.SliceStable(thirds, func(i, j int) bool {
		a, b := thirds[i], thirds[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.GD != b.GD {
			return a.GD > b.GD
		}
		if a.GF != b.GF {
			return a.GF > b.GF
		}
		return a.Name < b.Name
	})

	for i := range thirds {
		thirds[i].Rank = i + 1
		thirds[i].Qualified = i < ThirdPlaceQualifiers
	}
	return thirds
}

// sortRows orders a group's rows by the display tie-break chain:
// points desc, goal difference desc, goals-for desc, then name asc.
func sortRows(rows []Row) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.GD != b.GD {
			return a.GD > b.GD
		}
		if a.GF != b.GF {
			return a.GF > b.GF
		}
		return a.Name < b.Name
	})
}

// bareGroup strips a leading case-insensitive "Group " prefix and trims
// whitespace, returning the bare group letter (e.g. "Group A" -> "A").
func bareGroup(label string) string {
	s := strings.TrimSpace(label)
	const prefix = "group "
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		s = strings.TrimSpace(s[len(prefix):])
	}
	return s
}
