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

const (
	// groupMaxDuration / knockoutMaxDuration are how long after kickoff a match is
	// certainly over — past this the result is final. We deliberately do NOT trust
	// FIFA's MatchStatus code: it has been observed flipping to "full time" (3)
	// while a match was still being played (e.g. Germany 6–1 was announced, then
	// became 7–1). So status is purely time-based. Generous enough to cover halves
	// + half-time + stoppage (group), plus extra time & penalties (knockout).
	groupMaxDuration    = 140 * time.Minute
	knockoutMaxDuration = 175 * time.Minute
)

// mapStatus derives the lifecycle state PURELY from kickoff time and elapsed
// duration — never from the score or FIFA's (unreliable) status code.
//
// CRITICAL: the FIFA calendar reports the running score during play, and its
// MatchStatus code can read "finished" mid-match, so neither can be trusted to
// mean the match is over. A match is finished only once enough time has passed
// that it must be over (maxDur). This trades a small delay (~20-40 min after the
// final whistle) for never announcing a non-final score.
//
//	finished : the stage's max duration has elapsed and a score is recorded
//	live     : kickoff is past but the match window hasn't fully elapsed
//	scheduled: future, pre-draw (no kickoff), or a stale anomaly with no score
func mapStatus(hasScore bool, kickoff *time.Time, now time.Time, maxDur time.Duration) Status {
	if kickoff == nil || now.Before(*kickoff) {
		return StatusScheduled
	}
	elapsed := now.Sub(*kickoff)

	// Still inside the match window → live (even with a running score).
	if elapsed < maxDur {
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
	f.Status = mapStatus(hasScore, f.KickoffAt, now, maxMatchDuration(f.Stage))

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
