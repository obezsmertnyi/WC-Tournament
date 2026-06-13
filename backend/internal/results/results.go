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

// ResultsProvider is the decoupled source of calendar + results data.
type ResultsProvider interface {
	// Fixtures returns the full known calendar (schedule + any scores/status).
	Fixtures(ctx context.Context) ([]Fixture, error)
}
