import type { Match, Stage } from '../types'

const KYIV_TZ = 'Europe/Kyiv'

const dateFmt = new Intl.DateTimeFormat('uk-UA', {
  timeZone: KYIV_TZ,
  weekday: 'short',
  day: 'numeric',
  month: 'long',
})

const timeFmt = new Intl.DateTimeFormat('uk-UA', {
  timeZone: KYIV_TZ,
  hour: '2-digit',
  minute: '2-digit',
})

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

export const STAGE_LABELS: Record<Stage, string> = {
  group: 'Груповий етап',
  r32: '1/16 фіналу',
  r16: '1/8 фіналу',
  qf: 'Чвертьфінали',
  sf: 'Півфінали',
  third: 'Матч за 3-тє місце',
  final: 'Фінал',
}

export const STATUS_LABELS: Record<Match['status'], string> = {
  scheduled: 'Заплановано',
  live: 'LIVE',
  finished: 'FT',
}

export interface GroupSection {
  key: string
  /** Display title, e.g. "Група A". */
  title: string
  matches: Match[]
}

export interface KnockoutSection {
  key: Stage
  title: string
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
 * section are sorted by kickoff time.
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
      title: key === '—' ? 'Група' : `Група ${key}`,
      matches: (groupMap.get(key) ?? []).slice().sort(byKickoff),
    }))

  const knockout: KnockoutSection[] = KNOCKOUT_ORDER.filter((stage) =>
    knockoutMap.has(stage),
  ).map((stage) => ({
    key: stage,
    title: STAGE_LABELS[stage],
    matches: (knockoutMap.get(stage) ?? []).slice().sort(byKickoff),
  }))

  return { groupStage, knockout }
}
