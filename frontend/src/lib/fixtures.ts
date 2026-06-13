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

/** Knockout stages in the canonical bracket order. */
export const KNOCKOUT_ORDER: Stage[] = ['r32', 'r16', 'qf', 'sf', 'third', 'final']

/** Localized label for a tournament stage. */
export function stageLabel(stage: Stage): string {
  return i18n.t(`stage.${stage}`)
}

/** Localized label for a match status chip. */
export function statusLabel(status: Match['status']): string {
  return i18n.t(`status.${status}`)
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
