package results

// Wire types for the official FIFA API v3 calendar endpoint.
// Only the fields we consume are modeled; the API returns far more.

// fifaCalendarResponse is the envelope of GET /calendar/matches.
type fifaCalendarResponse struct {
	ContinuationToken *string     `json:"ContinuationToken"`
	Results           []fifaMatch `json:"Results"`
}

// fifaLocalized is FIFA's localized string array element.
type fifaLocalized struct {
	Locale      string `json:"Locale"`
	Description string `json:"Description"`
}

type fifaTeam struct {
	IdTeam       string          `json:"IdTeam"`
	IdCountry    string          `json:"IdCountry"`
	Abbreviation string          `json:"Abbreviation"`
	PictureUrl   string          `json:"PictureUrl"`
	TeamName     []fifaLocalized `json:"TeamName"`
	Score        *int            `json:"Score"`
}

type fifaStadium struct {
	Name      []fifaLocalized `json:"Name"`
	CityName  []fifaLocalized `json:"CityName"`
	IdCountry string          `json:"IdCountry"`
}

// --- Live match endpoint wire types (GET /live/football/...) ---

// fifaLivePlayer is one player in a team lineup.
type fifaLivePlayer struct {
	IdPlayer      string          `json:"IdPlayer"`
	ShirtNumber   *int            `json:"ShirtNumber"`
	Captain       bool            `json:"Captain"`
	PlayerName    []fifaLocalized `json:"PlayerName"`
	ShortName     []fifaLocalized `json:"ShortName"`
	Position      int             `json:"Position"`
	PlayerPicture any             `json:"PlayerPicture"` // sometimes object, sometimes string — ignored
	FieldStatus   *int            `json:"FieldStatus"`
}

// fifaLiveBooking is a card event.
type fifaLiveBooking struct {
	Card     *int   `json:"Card"`
	Minute   string `json:"Minute"`
	IdPlayer string `json:"IdPlayer"`
	IdTeam   string `json:"IdTeam"`
}

// fifaLiveGoal is a goal event.
type fifaLiveGoal struct {
	Type           *int   `json:"Type"`
	IdPlayer       string `json:"IdPlayer"`
	Minute         string `json:"Minute"`
	IdAssistPlayer string `json:"IdAssistPlayer"`
	IdTeam         string `json:"IdTeam"`
}

// fifaLiveSubstitution is a substitution event.
type fifaLiveSubstitution struct {
	Minute        string          `json:"Minute"`
	PlayerOffName []fifaLocalized `json:"PlayerOffName"`
	PlayerOnName  []fifaLocalized `json:"PlayerOnName"`
	IdTeam        string          `json:"IdTeam"`
}

// fifaLiveCoach is a team coach.
type fifaLiveCoach struct {
	Name []fifaLocalized `json:"Name"`
	// Role is omitted: FIFA returns it as a number for some matches and a
	// string for others, and the parser doesn't use it.
}

// fifaLiveTeam is one side of the live match.
type fifaLiveTeam struct {
	IdTeam        string                 `json:"IdTeam"`
	TeamName      []fifaLocalized        `json:"TeamName"`
	Tactics       string                 `json:"Tactics"`
	Players       []fifaLivePlayer       `json:"Players"`
	Bookings      []fifaLiveBooking      `json:"Bookings"`
	Goals         []fifaLiveGoal         `json:"Goals"`
	Substitutions []fifaLiveSubstitution `json:"Substitutions"`
	Coaches       []fifaLiveCoach        `json:"Coaches"`
}

// fifaLiveOfficial is a match official.
type fifaLiveOfficial struct {
	Name         []fifaLocalized `json:"Name"`
	OfficialType *int            `json:"OfficialType"`
}

// fifaLiveResponse is the envelope of GET /live/football/...
type fifaLiveResponse struct {
	IdStage                string             `json:"IdStage"`
	MatchTime              string             `json:"MatchTime"`
	Attendance             string             `json:"Attendance"`
	Winner                 string             `json:"Winner"`
	BallPossession         *fifaPossession    `json:"BallPossession"`
	HomeTeam               *fifaLiveTeam      `json:"HomeTeam"`
	AwayTeam               *fifaLiveTeam      `json:"AwayTeam"`
	Officials              []fifaLiveOfficial `json:"Officials"`
	Stadium                *fifaStadium       `json:"Stadium"`
	HomeTeamPenaltyScore   *int               `json:"HomeTeamPenaltyScore"`
	AwayTeamPenaltyScore   *int               `json:"AwayTeamPenaltyScore"`
	AggregateHomeTeamScore *int               `json:"AggregateHomeTeamScore"`
	AggregateAwayTeamScore *int               `json:"AggregateAwayTeamScore"`
}

// fifaPossession is the BallPossession object (often null pre-match).
type fifaPossession struct {
	Intervals   []fifaPossessionPoint `json:"Intervals"`
	OverallHome *float64              `json:"OverallHome"`
	OverallAway *float64              `json:"OverallAway"`
	LastX       *float64              `json:"LastX"`
}

type fifaPossessionPoint struct {
	HomePercentage *float64 `json:"HomePercentage"`
	AwayPercentage *float64 `json:"AwayPercentage"`
}

type fifaMatch struct {
	IdMatch       string          `json:"IdMatch"`
	IdStage       string          `json:"IdStage"`
	MatchNumber   *int            `json:"MatchNumber"`
	StageName     []fifaLocalized `json:"StageName"`
	GroupName     []fifaLocalized `json:"GroupName"`
	Date          string          `json:"Date"`
	MatchStatus   *int            `json:"MatchStatus"`
	Home          *fifaTeam       `json:"Home"`
	Away          *fifaTeam       `json:"Away"`
	HomeTeamScore *int            `json:"HomeTeamScore"`
	AwayTeamScore *int            `json:"AwayTeamScore"`
	Stadium       *fifaStadium    `json:"Stadium"`
	PlaceHolderA  string          `json:"PlaceHolderA"`
	PlaceHolderB  string          `json:"PlaceHolderB"`
}
