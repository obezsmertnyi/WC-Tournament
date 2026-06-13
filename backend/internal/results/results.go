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

// ResultsProvider is the decoupled source of calendar + results data.
type ResultsProvider interface {
	// Fixtures returns the full known calendar (schedule + any scores/status).
	Fixtures(ctx context.Context) ([]Fixture, error)
}
