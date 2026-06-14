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
	// minMatchDuration is the earliest a match can plausibly be over (90' +
	// half-time + stoppage). A match younger than this is never "finished",
	// even if a status code or score says otherwise — this guards against the
	// calendar's running score being mistaken for a final result.
	minMatchDuration = 105 * time.Minute
	// groupMaxDuration / knockoutMaxDuration are the hard upper bounds after
	// which a match is certainly over (used as the fallback when FIFA's status
	// code is missing). Knockouts allow for extra time + penalties.
	groupMaxDuration    = 140 * time.Minute
	knockoutMaxDuration = 170 * time.Minute
)

// mapStatus derives the lifecycle state from kickoff time and elapsed duration,
// using the FIFA MatchStatus code only as a *guarded* confirming signal.
//
// CRITICAL: the FIFA calendar reports the running score DURING play, so the
// presence of a score does NOT mean the match is over (this previously caused
// results to be announced at the first goal). A match is finished only once a
// full match could have elapsed.
//
// FIFA MatchStatus terminal codes.
const (
	fifaStatusFullTime = 3 // end of regulation (in knockouts, ET/pens may follow)
	fifaStatusComplete = 7 // match complete, after any extra time / penalties
)
//
//	finished : code 7 (always over), OR code 3 when no ET is possible (group),
//	           each only after ≥ minMatchDuration; OR the stage's max duration has
//	           elapsed and a score is on record (fallback when codes are absent)
//	live     : kickoff is past but the match window hasn't fully elapsed
//	scheduled: future, pre-draw (no kickoff), or a stale anomaly with no score
//
// code3Final must be true only for stages that cannot go to extra time (group
// stage): there, end-of-regulation (code 3) is the final whistle. In knockouts
// code 3 is just the end of normal time — the match continues into ET/penalties,
// so only code 7 (or the time-window fallback) marks it finished.
func mapStatus(code *int, hasScore bool, kickoff *time.Time, now time.Time, maxDur time.Duration, code3Final bool) Status {
	if kickoff == nil || now.Before(*kickoff) {
		return StatusScheduled
	}
	elapsed := now.Sub(*kickoff)

	// Trust an explicit terminal code, but only once a plausible full match has
	// elapsed — never finish a match minutes after kickoff.
	if code != nil && elapsed >= minMatchDuration {
		if *code == fifaStatusComplete || (*code == fifaStatusFullTime && code3Final) {
			return StatusFinished
		}
	}
	// Still inside the match window → live (even with a running score, and even at
	// end-of-regulation of a knockout that's heading into extra time).
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
	code3Final := f.Stage == StageGroup // groups can't go to extra time
	f.Status = mapStatus(m.MatchStatus, hasScore, f.KickoffAt, now, maxMatchDuration(f.Stage), code3Final)

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
