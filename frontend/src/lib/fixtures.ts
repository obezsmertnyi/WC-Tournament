import type { Match, Stage } from '../types'
import i18n from '../i18n'

const KYIV_TZ = 'Europe/Kyiv'

/** Intl locale to use for weekday/month names, derived from the UI language. */
function intlLocale(): string {
  return i18n.resolvedLanguage === 'en' ? 'en-GB' : 'uk-UA'
}

/**
 * Date/time formatters are rebuilt whenever the language changes (timezone is
 * always Europe/Kyiv — the pool's fixed zone — only the locale switches).
 */
let dateFmt: Intl.DateTimeFormat
let timeFmt: Intl.DateTimeFormat
let weekdayFmt: Intl.DateTimeFormat
let dayMonthFmt: Intl.DateTimeFormat
let fullDateFmt: Intl.DateTimeFormat

function rebuildFormatters() {
  const locale = intlLocale()
  dateFmt = new Intl.DateTimeFormat(locale, {
    timeZone: KYIV_TZ,
    weekday: 'short',
    day: 'numeric',
    month: 'long',
  })
  timeFmt = new Intl.DateTimeFormat(locale, {
    timeZone: KYIV_TZ,
    hour: '2-digit',
    minute: '2-digit',
  })
  // Short weekday for the date-strip chips ("Mon" / "пн").
  weekdayFmt = new Intl.DateTimeFormat(locale, {
    timeZone: KYIV_TZ,
    weekday: 'short',
  })
  // Day + short month for the date-strip chips ("14 Jun" / "14 черв.").
  dayMonthFmt = new Intl.DateTimeFormat(locale, {
    timeZone: KYIV_TZ,
    day: 'numeric',
    month: 'short',
  })
  // Long, human heading for the selected day.
  fullDateFmt = new Intl.DateTimeFormat(locale, {
    timeZone: KYIV_TZ,
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  })
}

rebuildFormatters()
i18n.on('languageChanged', rebuildFormatters)

/** A short date key (YYYY-MM-DD in Kyiv) used to group fixtures by day. */
const dayKeyFmt = new Intl.DateTimeFormat('en-CA', {
  timeZone: KYIV_TZ,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
})

export function formatKyivDate(iso: string): string {
  return dateFmt.format(new Date(iso))
}

export function formatKyivTime(iso: string): string {
  return timeFmt.format(new Date(iso))
}

export function kyivDayKey(iso: string): string {
  return dayKeyFmt.format(new Date(iso))
}

/** Today's day key in the Kyiv timezone (YYYY-MM-DD). */
export function todayKyivKey(): string {
  return dayKeyFmt.format(new Date())
}

/** Short weekday for a date-strip chip, e.g. "Mon" / "пн". */
export function formatKyivWeekday(iso: string): string {
  return weekdayFmt.format(new Date(iso))
}

/** Day + short month for a date-strip chip, e.g. "14 Jun" / "14 черв.". */
export function formatKyivDayMonth(iso: string): string {
  return dayMonthFmt.format(new Date(iso))
}

/** Long localized heading for the selected day, e.g. "Sunday, 14 June". */
export function formatKyivFullDate(iso: string): string {
  return fullDateFmt.format(new Date(iso))
}

/** A single match-day bucket for the day-centric calendar. */
export interface MatchDay {
  /** YYYY-MM-DD day key in Kyiv. */
  key: string
  /** Representative ISO timestamp (first kickoff of the day) for formatting. */
  iso: string
  matches: Match[]
}

/**
 * Group matches into chronologically ordered match-days (Kyiv tz). Matches
 * within a day are sorted by kickoff. Matches without a kickoff time are
 * skipped (they have no day to belong to).
 */
export function buildMatchDays(matches: Match[]): MatchDay[] {
  const map = new Map<string, Match[]>()
  for (const m of matches) {
    if (!m.kickoffAt) continue
    const key = kyivDayKey(m.kickoffAt)
    const arr = map.get(key) ?? []
    arr.push(m)
    map.set(key, arr)
  }
  return [...map.keys()].sort().map((key) => {
    const dayMatches = (map.get(key) ?? []).slice().sort(byKickoff)
    return { key, iso: dayMatches[0].kickoffAt, matches: dayMatches }
  })
}

/**
 * Pick the default selected day key: today if it has matches, else the nearest
 * upcoming match-day, else the first match-day. Returns undefined when empty.
 */
export function defaultMatchDayKey(days: MatchDay[]): string | undefined {
  if (days.length === 0) return undefined
  const today = todayKyivKey()
  const exact = days.find((d) => d.key === today)
  if (exact) return exact.key
  const upcoming = days.find((d) => d.key >= today)
  return (upcoming ?? days[0]).key
}

/**
 * Compose a venue caption without duplicating the city when the stadium name
 * already leads with it (e.g. "Mexico City" + "Mexico City Stadium" →
 * "Mexico City Stadium", not "Mexico City · Mexico City Stadium").
 */
export function venueCaption(city: string, stadium: string): string {
  const c = city.trim()
  const s = stadium.trim()
  if (!c) return s
  if (!s) return c
  const norm = (v: string) => v.toLowerCase()
  if (norm(s) === norm(c) || norm(s).startsWith(norm(c))) return s
  return `${c} · ${s}`
}

/** Knockout stages in the canonical bracket order. */
export const KNOCKOUT_ORDER: Stage[] = ['r32', 'r16', 'qf', 'sf', 'third', 'final']

/** Localized label for a tournament stage. */
export function stageLabel(stage: Stage): string {
  return i18n.t(`stage.${stage}`)
}

/** Short localized label for a knockout stage (round chips / bracket headers). */
export function stageShortLabel(stage: Stage): string {
  return i18n.t(`stageShort.${stage}`, { defaultValue: stageLabel(stage) })
}

/** Localized label for a match status chip. */
export function statusLabel(status: Match['status']): string {
  return i18n.t(`status.${status}`)
}

/**
 * Whether a match has kicked off (predictions lock at this point). True for any
 * non-`scheduled` status, and also once the kickoff timestamp is in the past —
 * so the UI locks correctly even if a status hasn't flipped server-side yet.
 */
export function hasKickedOff(match: Pick<Match, 'status' | 'kickoffAt'>): boolean {
  if (match.status !== 'scheduled') return true
  if (!match.kickoffAt) return false
  return new Date(match.kickoffAt).getTime() <= Date.now()
}

/**
 * A compact, localized "relative time" string for the audit feed, e.g.
 * "2m ago" / "2 хв тому". Uses Intl.RelativeTimeFormat so it localizes cleanly.
 */
export function formatRelativeTime(iso: string): string {
  const rtf = new Intl.RelativeTimeFormat(intlLocale(), { numeric: 'auto', style: 'short' })
  const diffSec = Math.round((new Date(iso).getTime() - Date.now()) / 1000)

  // Ordered (seconds-per-unit, unit) — largest unit whose span fits wins.
  const steps: [number, Intl.RelativeTimeFormatUnit][] = [
    [31557600, 'year'],
    [2629800, 'month'],
    [604800, 'week'],
    [86400, 'day'],
    [3600, 'hour'],
    [60, 'minute'],
    [1, 'second'],
  ]
  for (const [secs, unit] of steps) {
    if (Math.abs(diffSec) >= secs || unit === 'second') {
      return rtf.format(Math.round(diffSec / secs), unit)
    }
  }
  return rtf.format(0, 'second')
}

/**
 * Audit-feed timestamp: a relative string for recent events (< 24h, e.g.
 * "2 хв тому") and an absolute Kyiv date + time for older ones (e.g.
 * "13 черв. · 22:00"), so the feed stays readable as it ages.
 */
export function formatAuditTime(iso: string): string {
  const ageMs = Date.now() - new Date(iso).getTime()
  if (ageMs < 24 * 3600 * 1000) return formatRelativeTime(iso)
  return `${formatKyivDayMonth(iso)} · ${formatKyivTime(iso)}`
}

/** Full Kyiv date + time for a tooltip, e.g. "13 черв. · 22:00". */
export function formatKyivDateTime(iso: string): string {
  return `${formatKyivDayMonth(iso)} · ${formatKyivTime(iso)}`
}

export interface GroupSection {
  key: string
  /** Group letter ("A"), or "—" when missing. The display title is derived in the UI. */
  letter: string
  matches: Match[]
}

export interface KnockoutSection {
  key: Stage
  matches: Match[]
}

export interface GroupedFixtures {
  groupStage: GroupSection[]
  knockout: KnockoutSection[]
}

function byKickoff(a: Match, b: Match): number {
  const t = new Date(a.kickoffAt).getTime() - new Date(b.kickoffAt).getTime()
  if (t !== 0) return t
  return a.matchNumber - b.matchNumber
}

/**
 * Split matches into the group stage (sectioned by group letter, A→Z) and the
 * knockout phase (sectioned by stage in bracket order). Matches inside each
 * section are sorted by kickoff time. Display titles are localized in the UI.
 */
export function groupFixtures(matches: Match[]): GroupedFixtures {
  const groupMap = new Map<string, Match[]>()
  const knockoutMap = new Map<Stage, Match[]>()

  for (const m of matches) {
    if (m.stage === 'group') {
      // Fall back to "—" when the group letter is missing.
      const key = (m.group ?? '—').toUpperCase()
      const arr = groupMap.get(key) ?? []
      arr.push(m)
      groupMap.set(key, arr)
    } else {
      const arr = knockoutMap.get(m.stage) ?? []
      arr.push(m)
      knockoutMap.set(m.stage, arr)
    }
  }

  const groupStage: GroupSection[] = [...groupMap.keys()]
    .sort((a, b) => a.localeCompare(b, 'uk'))
    .map((key) => ({
      key,
      letter: key,
      matches: (groupMap.get(key) ?? []).slice().sort(byKickoff),
    }))

  const knockout: KnockoutSection[] = KNOCKOUT_ORDER.filter((stage) =>
    knockoutMap.has(stage),
  ).map((stage) => ({
    key: stage,
    matches: (knockoutMap.get(stage) ?? []).slice().sort(byKickoff),
  }))

  return { groupStage, knockout }
}
