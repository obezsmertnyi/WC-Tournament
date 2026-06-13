package storage

import "time"

// Team is a persisted national team row.
type Team struct {
	ID         int64
	FifaID     string
	Name       string
	Code       string
	FlagURL    string
	GroupLabel string
}

// Match is a persisted fixture row joined with optional team data when read.
type Match struct {
	ID              int64
	FifaID          string
	Stage           string
	GroupLabel      string
	MatchNumber     *int
	KickoffAt       *time.Time
	Status          string
	HomeScore       *int
	AwayScore       *int
	VenueStadium    string
	VenueCity       string
	VenueCountry    string
	PlaceholderHome string
	PlaceholderAway string
	ResultSource    string
	UpdatedAt       time.Time

	// Joined team data (nil before the draw / when not yet assigned).
	Home *Team
	Away *Team
}

// UpsertTeam carries the fields written during a FIFA sync.
type UpsertTeam struct {
	FifaID     string
	Name       string
	Code       string
	FlagURL    string
	GroupLabel string
}

// UpsertMatch carries the fields written during a FIFA sync. HomeFifaID /
// AwayFifaID reference teams by their FIFA id; they are resolved to local
// team ids inside the upsert. Empty team ids mean "not yet assigned"
// (pre-draw knockout slots).
type UpsertMatch struct {
	FifaID          string
	Stage           string
	GroupLabel      string
	MatchNumber     *int
	HomeFifaID      string
	AwayFifaID      string
	KickoffAt       *time.Time
	Status          string
	HomeScore       *int
	AwayScore       *int
	VenueStadium    string
	VenueCity       string
	VenueCountry    string
	PlaceholderHome string
	PlaceholderAway string
}
