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
