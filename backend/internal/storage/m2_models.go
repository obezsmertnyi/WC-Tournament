package storage

import "time"

// User is a persisted user row.
type User struct {
	ID               int64
	GoogleSub        *string
	Email            *string
	Nickname         string
	AvatarURL        *string
	FavoriteTeamCode *string
	TelegramChatID   *string
	Role             string
	Approved         bool
	AccessLevel      string
	CreatedAt        time.Time
}

// Prediction is a persisted prediction row.
type Prediction struct {
	ID               int64
	UserID           int64
	MatchID          int64
	HomePred         int
	AwayPred         int
	WinnerPickTeamID *int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PointRow is a materialized per-match points row.
type PointRow struct {
	UserID        int64
	MatchID       int64
	Points        int
	BreakdownJSON []byte
}

// BonusRule is a persisted bonus rule row.
type BonusRule struct {
	Kind    string
	Enabled bool
	Pts     int
}

// ChampionTier is a time-tiered champion bonus window.
type ChampionTier struct {
	WindowEnd *time.Time
	Pts       int
}

// TournamentPick is a persisted tournament bonus pick row.
type TournamentPick struct {
	ID         int64
	UserID     int64
	Kind       string
	PickRef    string
	LockedAt   *time.Time
	TierPoints *int
	Awarded    bool // true once the pick is resolved correct (points count)
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// UserHistoryRow is one match in a player's personal results history: their
// prediction, the actual result, and the points earned.
type UserHistoryRow struct {
	MatchID          int64
	Stage            string
	GroupLabel       string
	KickoffAt        *time.Time
	Status           string
	HomeCode         string
	HomeName         string
	HomeFlag         string
	AwayCode         string
	AwayName         string
	AwayFlag         string
	HomeScore        *int
	AwayScore        *int
	PredHome         int
	PredAway         int
	WinnerPickTeamID *int64
	Points           int
	Exact            bool
	Scored           bool // true once points are materialized (match scored)
}

// AuditEntry is a persisted, immutable audit row (actions only, never values).
type AuditEntry struct {
	ID           int64
	ActorUserID  *int64
	ActorRole    string
	Action       string
	MatchID      *int64
	TargetUserID *int64
	CreatedAt    time.Time
	// ActorNickname is joined on read for the public feed (empty for system).
	ActorNickname string
}

// LeaderboardRow is one aggregated leaderboard entry. Points is the grand total
// (match points + bonus points); MatchPoints and BonusPoints break it down so
// the UI can show a separate bonus column.
type LeaderboardRow struct {
	UserID      int64
	Nickname    string
	AvatarURL   *string
	Points      int
	MatchPoints int
	BonusPoints int
	ExactCount  int
	Played      int
}

// MatchPrediction is a revealed prediction for the per-match reveal endpoint.
type MatchPrediction struct {
	UserID           int64
	Nickname         string
	AvatarURL        *string
	HomePred         int
	AwayPred         int
	WinnerPickTeamID *int64
	Points           int
}
