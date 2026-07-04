// Package results defines the pluggable results/calendar provider abstraction
// for the WC-Tournament backend and the official FIFA API implementation.
//
// The provider is the source of truth for fixtures, scores and status; a
// manual admin override (result_source='manual') always wins on persistence
// (see docs/architecture.md §3.2, ADR-0006).
package results

import (
	"context"
	"time"
)

// Stage enumerates the tournament stages used across the system. The string
// values match the matches.stage CHECK constraint.
type Stage string

const (
	StageGroup Stage = "group"
	StageR32   Stage = "r32"
	StageR16   Stage = "r16"
	StageQF    Stage = "qf"
	StageSF    Stage = "sf"
	StageFinal Stage = "final"
	StageThird Stage = "third"
)

// Status enumerates match lifecycle states (matches.status CHECK constraint).
type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusLive      Status = "live"
	StatusFinished  Status = "finished"
)

// FixtureTeam is one side of a fixture. All fields may be empty before the
// draw (knockout placeholder slots).
type FixtureTeam struct {
	FifaID  string // IdTeam
	Name    string // TeamName[en].Description
	Code    string // Abbreviation / IdCountry
	FlagURL string // PictureUrl
}

// Fixture is a normalized, source-agnostic match as returned by a
// ResultsProvider. Pointers are nil when the source omits the value.
type Fixture struct {
	FifaID          string // IdMatch
	FifaStageID     string // IdStage (raw FIFA stage id, needed for the live endpoint)
	Stage           Stage
	GroupLabel      string // GroupName[en].Description
	MatchNumber     *int
	Home            FixtureTeam
	Away            FixtureTeam
	KickoffAt       *time.Time // Date (UTC)
	Status          Status
	HomeScore       *int
	AwayScore       *int
	VenueStadium    string
	VenueCity       string
	VenueCountry    string
	PlaceholderHome string // PlaceHolderA
	PlaceholderAway string // PlaceHolderB
}

// --- Live match detail (statistics) ---

// LivePlayer is a player in a starting lineup / bench.
type LivePlayer struct {
	FifaID      string
	Name        string
	ShirtNumber *int
	Position    string
	Captain     bool
	PictureURL  string
}

// LiveLineup is a team's formation and players for the match.
type LiveLineup struct {
	TeamFifaID string
	TeamName   string
	Formation  string // Tactics, e.g. "4-1-2-3"
	Players    []LivePlayer
}

// LiveGoal is a goal event normalized with resolved scorer/assist names.
type LiveGoal struct {
	TeamFifaID string
	Side       string // "home" or "away" — which team scored
	ScorerName string
	AssistName string // "" when none
	Minute     string
	Type       *int // raw FIFA goal type (0 normal, etc.)
	// Period is FIFA's phase code for the goal: 3/5 = regulation halves,
	// 7/9 = extra-time halves. It is the only signal distinguishing a
	// regulation goal from an extra-time one (Minute is fuzzy under stoppage);
	// RegulationScore uses it to recover the 90-minute scoreline.
	Period *int
}

// LiveCard is a booking event normalized with the resolved player name.
type LiveCard struct {
	TeamFifaID string
	Side       string // "home" or "away"
	PlayerName string
	Minute     string
	Card       *int // raw FIFA card type (1 yellow, etc.)
}

// LiveSubstitution is a substitution event.
type LiveSubstitution struct {
	TeamFifaID string
	Side       string // "home" or "away"
	PlayerIn   string
	PlayerOut  string
	Minute     string
}

// LiveOfficial is a match official.
type LiveOfficial struct {
	Name string
	Type *int // OfficialType: 1=Referee
}

// LivePossession is overall ball possession when reported.
type LivePossession struct {
	Home *float64
	Away *float64
}

// LiveMatch is the normalized, source-agnostic live/finished match statistics
// payload returned by a provider's LiveMatch method.
type LiveMatch struct {
	FifaStageID   string
	MatchTime     string // e.g. "98'"
	Attendance    string
	WinnerTeamID  string // IdTeam of the winner, "" when none/draw
	Possession    *LivePossession
	HomeLineup    *LiveLineup
	AwayLineup    *LiveLineup
	Goals         []LiveGoal
	Cards         []LiveCard
	Substitutions []LiveSubstitution
	Officials     []LiveOfficial
	Stadium       string

	HomePenaltyScore   *int
	AwayPenaltyScore   *int
	AggregateHomeScore *int
	AggregateAwayScore *int
}

// RegulationScore recovers the 90-minute (regulation) scoreline of a knockout
// match from its goal events, counting only goals scored in regulation halves
// (Period 3 or 5) per side. It is the root fix for extra-time GOAL wins: the
// FIFA calendar feed reports only the final aet-inclusive score (e.g. a 2:2
// regulation draw won 3:2 in extra time), which the regulation-based scoring
// otherwise reads as decisive.
//
// finalHome/finalAway are the stored aet-inclusive scores. ok reports whether
// the derivation is trustworthy: every goal must carry a Period and belong to a
// known side, and the goals in periods 3/5/7/9 (regulation + extra time) must
// sum exactly to the final aet score. When ok is false the caller must keep the
// stored score rather than risk zeroing a match whose goal events are missing
// or incomplete (many finished matches carry a winner but no goal timeline).
func RegulationScore(goals []LiveGoal, finalHome, finalAway int) (regHome, regAway int, ok bool) {
	var totHome, totAway int
	for _, g := range goals {
		if g.Period == nil {
			return 0, 0, false // incomplete data — don't trust the derivation
		}
		p := *g.Period
		reg := p == 3 || p == 5 // regulation first/second half
		et := p == 7 || p == 9  // extra-time first/second half
		if !reg && !et {
			continue // shootout / other phases are not part of the aet scoreline
		}
		switch g.Side {
		case "home":
			totHome++
			if reg {
				regHome++
			}
		case "away":
			totAway++
			if reg {
				regAway++
			}
		default:
			return 0, 0, false // goal not attributable to a side
		}
	}
	ok = totHome == finalHome && totAway == finalAway
	return regHome, regAway, ok
}

// ResultsProvider is the decoupled source of calendar + results data.
type ResultsProvider interface {
	// Fixtures returns the full known calendar (schedule + any scores/status).
	Fixtures(ctx context.Context) ([]Fixture, error)
}
