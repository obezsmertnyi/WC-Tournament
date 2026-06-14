package results

import (
	"strings"
	"time"
)

// Knockout stage ids for WC2026 (docs/architecture.md §3.2). The group stage
// is "First Stage" (289273). Unknown ids fall back to inference from the
// localized stage name.
var fifaStageByID = map[string]Stage{
	"289273": StageGroup, // First Stage / group
	"289287": StageR32,
	"289288": StageR16,
	"289289": StageQF,
	"289290": StageSF,
	"289291": StageThird,
	"289292": StageFinal,
}

// localized picks the en / en-GB entry from a FIFA localized array, falling
// back to the first element, then "".
func localized(arr []fifaLocalized) string {
	if len(arr) == 0 {
		return ""
	}
	for _, l := range arr {
		lc := strings.ToLower(l.Locale)
		if lc == "en" || lc == "en-gb" {
			return l.Description
		}
	}
	for _, l := range arr {
		if strings.HasPrefix(strings.ToLower(l.Locale), "en") {
			return l.Description
		}
	}
	return arr[0].Description
}

// mapStage resolves the stage enum from the FIFA stage id, falling back to
// keyword inference from the localized stage name.
func mapStage(idStage string, stageName []fifaLocalized) Stage {
	if s, ok := fifaStageByID[idStage]; ok {
		return s
	}
	return inferStage(localized(stageName))
}

func inferStage(name string) Stage {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "round of 32"):
		return StageR32
	case strings.Contains(n, "round of 16"):
		return StageR16
	case strings.Contains(n, "quarter"):
		return StageQF
	case strings.Contains(n, "semi"):
		return StageSF
	case strings.Contains(n, "third") || strings.Contains(n, "3rd") || strings.Contains(n, "play-off for third") || strings.Contains(n, "bronze"):
		return StageThird
	case strings.Contains(n, "final"):
		return StageFinal
	case strings.Contains(n, "group") || strings.Contains(n, "first stage"):
		return StageGroup
	default:
		// Safest default: treat unknown as group so the row is still
		// persistable (passes the CHECK constraint).
		return StageGroup
	}
}

// FIFA MatchStatus codes as observed in this feed's /calendar/matches:
//
//	0 = played (full time, result is final)
//	1 = not started
//	3 = live / in progress (stays 3 through extra time & penalties)
//
// NB: 3 is LIVE, not "full time" — an earlier misreading announced matches
// mid-play (Germany was 3/live at 6–1, then finished 7–1). Trusting 0 for
// "finished" is both reliable and immediate, and 3 naturally covers ET/pens.
const (
	fifaMatchPlayed     = 0
	fifaMatchNotStarted = 1
	fifaMatchLive       = 3
)

const (
	// groupMaxDuration / knockoutMaxDuration bound how long after kickoff a match
	// could still be running — used as a FALLBACK only when MatchStatus is absent
	// or an unexpected value. Generous enough for halves + half-time + stoppage
	// (group) plus extra time & penalties (knockout).
	groupMaxDuration    = 140 * time.Minute
	knockoutMaxDuration = 175 * time.Minute
)

// mapStatus derives the lifecycle state. It trusts FIFA's MatchStatus code with
// the CORRECT semantics (0=played, 1=not started, 3=live) and falls back to a
// time window only when the code is missing/unexpected. A match is finished only
// when MatchStatus==0 (or, in the fallback, when its max duration has elapsed) —
// never from the running score, which the calendar updates during play.
func mapStatus(code *int, hasScore bool, kickoff *time.Time, now time.Time, maxDur time.Duration) Status {
	past := kickoff != nil && !now.Before(*kickoff)

	if code != nil {
		switch *code {
		case fifaMatchPlayed:
			if hasScore && past {
				return StatusFinished
			}
		case fifaMatchLive:
			return StatusLive // covers extra time & penalties (stays 3 until done)
		case fifaMatchNotStarted:
			return StatusScheduled
		}
	}

	// Fallback for a nil / unexpected code: pure time window.
	if kickoff == nil || now.Before(*kickoff) {
		return StatusScheduled
	}
	if now.Sub(*kickoff) < maxDur {
		return StatusLive
	}
	// Window elapsed: finished if a score is recorded, else a data anomaly
	// (postponed/abandoned) — leave unresolved rather than perpetually live.
	if hasScore {
		return StatusFinished
	}
	return StatusScheduled
}

// maxMatchDuration returns the stage-appropriate upper bound for mapStatus.
func maxMatchDuration(stage Stage) time.Duration {
	if stage == StageGroup {
		return groupMaxDuration
	}
	return knockoutMaxDuration
}

// parseCalendar maps a decoded FIFA calendar envelope to Fixtures. now is
// injected so status inference is deterministic and testable.
func parseCalendar(resp fifaCalendarResponse, now time.Time) []Fixture {
	out := make([]Fixture, 0, len(resp.Results))
	for _, m := range resp.Results {
		out = append(out, mapMatch(m, now))
	}
	return out
}

func mapMatch(m fifaMatch, now time.Time) Fixture {
	f := Fixture{
		FifaID:          m.IdMatch,
		FifaStageID:     m.IdStage,
		Stage:           mapStage(m.IdStage, m.StageName),
		GroupLabel:      localized(m.GroupName),
		MatchNumber:     m.MatchNumber,
		HomeScore:       m.HomeTeamScore,
		AwayScore:       m.AwayTeamScore,
		PlaceholderHome: m.PlaceHolderA,
		PlaceholderAway: m.PlaceHolderB,
	}

	if m.Home != nil {
		f.Home = mapTeam(m.Home)
	}
	if m.Away != nil {
		f.Away = mapTeam(m.Away)
	}

	if m.Date != "" {
		if t, err := time.Parse(time.RFC3339, m.Date); err == nil {
			utc := t.UTC()
			f.KickoffAt = &utc
		}
	}

	hasScore := m.HomeTeamScore != nil && m.AwayTeamScore != nil
	f.Status = mapStatus(m.MatchStatus, hasScore, f.KickoffAt, now, maxMatchDuration(f.Stage))

	if m.Stadium != nil {
		f.VenueStadium = localized(m.Stadium.Name)
		f.VenueCity = localized(m.Stadium.CityName)
		f.VenueCountry = m.Stadium.IdCountry
	}

	return f
}

func mapTeam(t *fifaTeam) FixtureTeam {
	code := t.Abbreviation
	if code == "" {
		code = t.IdCountry
	}
	return FixtureTeam{
		FifaID:  t.IdTeam,
		Name:    localized(t.TeamName),
		Code:    code,
		FlagURL: resolveFlagURL(t.PictureUrl),
	}
}

// flagURLReplacer substitutes the literal {format}/{size} tokens FIFA embeds in
// flag picture URLs (e.g. ".../flags-{format}-{size}/MEX") with concrete values
// so the stored URL is directly usable: {format}->sq (square), {size}->4.
var flagURLReplacer = strings.NewReplacer("{format}", "sq", "{size}", "4")

// resolveFlagURL renders a FIFA flag URL template into a usable URL. URLs
// without tokens (or empty strings) pass through unchanged.
func resolveFlagURL(raw string) string {
	if raw == "" {
		return ""
	}
	return flagURLReplacer.Replace(raw)
}
