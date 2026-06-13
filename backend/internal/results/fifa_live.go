package results

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// ErrLiveNotAvailable signals the FIFA live endpoint returned no usable
// statistics (match not started, or an empty/204 payload). Callers map this to
// an "available: false" response rather than an error.
var ErrLiveNotAvailable = errors.New("live match statistics not available")

// LiveMatch fetches live/finished statistics for a single match from the
// official FIFA live football endpoint and normalizes them to *LiveMatch.
// It reuses the shared *http.Client (cookie jar, timeout) and the same UA /
// Accept headers as the calendar fetch.
//
// Returns ErrLiveNotAvailable when the match has no statistics yet (e.g. not
// kicked off). Network / decode failures are returned as wrapped errors so the
// API layer can map them to 502.
func (c *FIFAClient) LiveMatch(ctx context.Context, idStage, idMatch string) (*LiveMatch, error) {
	if idStage == "" || idMatch == "" {
		return nil, ErrLiveNotAvailable
	}

	q := url.Values{}
	q.Set("language", "en")
	endpoint := fmt.Sprintf("%s/live/football/%s/%s/%s/%s?%s",
		c.baseURL, fifaCompetitionID, fifaSeasonID,
		url.PathEscape(idStage), url.PathEscape(idMatch), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build live request: %w", err)
	}
	req.Header.Set("User-Agent", fifaUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		// fall through to decode
	case http.StatusNoContent, http.StatusNotFound:
		return nil, ErrLiveNotAvailable
	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out fifaLiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode live response: %w", err)
	}

	lm := parseLiveMatch(out)
	if lm == nil {
		return nil, ErrLiveNotAvailable
	}
	return lm, nil
}

// parseLiveMatch maps the decoded FIFA live envelope to the normalized
// LiveMatch. Returns nil when the payload carries no usable statistics (no
// lineups and no events) so callers can surface "not available". Player names
// referenced by goals/cards are resolved from the lineup rosters.
func parseLiveMatch(resp fifaLiveResponse) *LiveMatch {
	lm := &LiveMatch{
		FifaStageID:        resp.IdStage,
		MatchTime:          resp.MatchTime,
		Attendance:         resp.Attendance,
		WinnerTeamID:       resp.Winner,
		HomePenaltyScore:   resp.HomeTeamPenaltyScore,
		AwayPenaltyScore:   resp.AwayTeamPenaltyScore,
		AggregateHomeScore: resp.AggregateHomeTeamScore,
		AggregateAwayScore: resp.AggregateAwayTeamScore,
	}

	if resp.Stadium != nil {
		lm.Stadium = localized(resp.Stadium.Name)
	}

	// Build a player-id -> display-name index from both rosters so goal/card
	// events (which only carry IdPlayer) can be resolved to names.
	playerName := make(map[string]string)
	indexPlayers(playerName, resp.HomeTeam)
	indexPlayers(playerName, resp.AwayTeam)

	lm.HomeLineup = mapLineup(resp.HomeTeam)
	lm.AwayLineup = mapLineup(resp.AwayTeam)

	for _, t := range []*fifaLiveTeam{resp.HomeTeam, resp.AwayTeam} {
		if t == nil {
			continue
		}
		for _, g := range t.Goals {
			lm.Goals = append(lm.Goals, LiveGoal{
				TeamFifaID: g.IdTeam,
				ScorerName: playerName[g.IdPlayer],
				AssistName: playerName[g.IdAssistPlayer],
				Minute:     g.Minute,
				Type:       g.Type,
			})
		}
		for _, b := range t.Bookings {
			lm.Cards = append(lm.Cards, LiveCard{
				TeamFifaID: b.IdTeam,
				PlayerName: playerName[b.IdPlayer],
				Minute:     b.Minute,
				Card:       b.Card,
			})
		}
		for _, s := range t.Substitutions {
			lm.Substitutions = append(lm.Substitutions, LiveSubstitution{
				TeamFifaID: s.IdTeam,
				PlayerIn:   localized(s.PlayerOnName),
				PlayerOut:  localized(s.PlayerOffName),
				Minute:     s.Minute,
			})
		}
	}

	if resp.BallPossession != nil {
		p := &LivePossession{
			Home: resp.BallPossession.OverallHome,
			Away: resp.BallPossession.OverallAway,
		}
		if p.Home != nil || p.Away != nil {
			lm.Possession = p
		}
	}

	for _, o := range resp.Officials {
		lm.Officials = append(lm.Officials, LiveOfficial{
			Name: localized(o.Name),
			Type: o.OfficialType,
		})
	}

	// Treat a payload with no lineups and no events as "not available".
	if lm.HomeLineup == nil && lm.AwayLineup == nil &&
		len(lm.Goals) == 0 && len(lm.Cards) == 0 && len(lm.Substitutions) == 0 {
		return nil
	}
	return lm
}

func indexPlayers(idx map[string]string, t *fifaLiveTeam) {
	if t == nil {
		return
	}
	for _, p := range t.Players {
		name := localized(p.PlayerName)
		if name == "" {
			name = localized(p.ShortName)
		}
		if p.IdPlayer != "" {
			idx[p.IdPlayer] = name
		}
	}
}

func mapLineup(t *fifaLiveTeam) *LiveLineup {
	if t == nil || len(t.Players) == 0 {
		return nil
	}
	l := &LiveLineup{
		TeamFifaID: t.IdTeam,
		TeamName:   localized(t.TeamName),
		Formation:  t.Tactics,
		Players:    make([]LivePlayer, 0, len(t.Players)),
	}
	for _, p := range t.Players {
		name := localized(p.PlayerName)
		if name == "" {
			name = localized(p.ShortName)
		}
		l.Players = append(l.Players, LivePlayer{
			FifaID:      p.IdPlayer,
			Name:        name,
			ShirtNumber: p.ShirtNumber,
			Position:    positionLabel(p.Position),
			Captain:     p.Captain,
		})
	}
	return l
}

// positionLabel maps FIFA's numeric position code to a readable label.
func positionLabel(code int) string {
	switch code {
	case 0:
		return "Goalkeeper"
	case 1:
		return "Defender"
	case 2:
		return "Midfielder"
	case 3:
		return "Forward"
	default:
		return ""
	}
}
