import type { Match } from '../types'
import { teamName } from './teamNames'

/** Index matches by their FIFA match number so "W75"/"RU101" can be resolved. */
export function buildMatchByNumber(matches: Match[]): Map<number, Match> {
  const m = new Map<number, Match>()
  for (const x of matches) if (x.matchNumber != null) m.set(x.matchNumber, x)
  return m
}

// Main-line knockout stages, earliest → latest (third-place is a separate node).
const PREV_STAGE: Record<string, string> = { r16: 'r32', qf: 'r16', sf: 'qf', final: 'sf' }

/**
 * Compute a tree-position key per match number so each knockout round can be
 * ordered top-to-bottom for a correct bracket. FIFA's match numbering does NOT
 * follow the tree (e.g. R16 #91/#92 feed a lower QF than #93/#94), so ordering
 * by number/kickoff misaligns the connectors. We derive each tie's parent (the
 * next-round tie it feeds, via "W<n>" placeholders or, once resolved, via the
 * team's previous-round match) and assign a heap-style index: parent*2 + slot
 * (home=0, away=1). Sorting a round by this key yields the classic geometry.
 */
export function bracketOrderKey(matches: Match[]): Map<number, number> {
  const mainline = matches.filter((m) => m.matchNumber != null && (m.stage === 'r32' || m.stage in PREV_STAGE))

  // teamId → matchNumber, per stage (to resolve a now-decided slot's feeder).
  const teamMatchByStage = new Map<string, Map<number, number>>()
  for (const m of mainline) {
    let mm = teamMatchByStage.get(m.stage)
    if (!mm) {
      mm = new Map()
      teamMatchByStage.set(m.stage, mm)
    }
    if (m.home) mm.set(m.home.id, m.matchNumber!)
    if (m.away) mm.set(m.away.id, m.matchNumber!)
  }

  // feeder matchNumber → { parent matchNumber, slot } (slot: 0=home, 1=away).
  const parent = new Map<number, { p: number; s: number }>()
  const feederOf = (placeholder: string | null, team: Match['home'], prevStage: string): number | null => {
    const code = /^(?:W|RU)(\d+)$/i.exec((placeholder ?? '').trim())
    if (code) return Number(code[1])
    if (team) return teamMatchByStage.get(prevStage)?.get(team.id) ?? null
    return null
  }
  for (const m of mainline) {
    const prev = PREV_STAGE[m.stage]
    if (!prev) continue // r32 has no feeders
    const hf = feederOf(m.placeholderHome, m.home, prev)
    const af = feederOf(m.placeholderAway, m.away, prev)
    if (hf != null) parent.set(hf, { p: m.matchNumber!, s: 0 })
    if (af != null) parent.set(af, { p: m.matchNumber!, s: 1 })
  }

  const memo = new Map<number, number>()
  const idx = (n: number, depth = 0): number => {
    if (memo.has(n)) return memo.get(n)!
    const par = depth > 8 ? undefined : parent.get(n) // depth guard against cycles
    const v = par ? idx(par.p, depth + 1) * 2 + par.s : 0
    memo.set(n, v)
    return v
  }

  const out = new Map<number, number>()
  for (const m of mainline) out.set(m.matchNumber!, idx(m.matchNumber!))
  return out
}

/**
 * Turn a bracket slot placeholder into a human-readable label:
 *   "W75"   → the winner of match 75 once it's decided, else the two
 *             competitors ("Нідерланди / Марокко"), else "Winner 1/8".
 *   "RU101" → the loser (third-place feeders), with the same fallbacks.
 * Anything else (or unknown) is returned as-is.
 */
export function resolveSlotLabel(
  placeholder: string | null,
  byNumber: Map<number, Match>,
  lang: string | undefined,
  t: (k: string, o?: Record<string, unknown>) => string,
): string {
  if (!placeholder) return t('fixture.tbd')
  const mt = /^(W|RU)(\d+)$/i.exec(placeholder.trim())
  if (!mt) return placeholder
  const kind = mt[1].toUpperCase()
  const src = byNumber.get(Number(mt[2]))
  if (!src) return placeholder

  const hk = !!src.home
  const ak = !!src.away

  // Decided (non-draw) feeder → the advancing team (W) or eliminated team (RU).
  if (
    hk && ak &&
    src.status === 'finished' &&
    src.homeScore != null && src.awayScore != null &&
    src.homeScore !== src.awayScore
  ) {
    const homeWon = src.homeScore > src.awayScore
    const team = kind === 'W' ? (homeWon ? src.home! : src.away!) : homeWon ? src.away! : src.home!
    return teamName(team.code, team.name, lang)
  }

  // Both competitors known but not decided yet → "A / B".
  if (hk && ak) {
    const a = teamName(src.home!.code, src.home!.name, lang)
    const b = teamName(src.away!.code, src.away!.name, lang)
    return `${a} / ${b}`
  }

  // Feeder itself isn't resolved (deeper round) → clean stage label.
  if (kind === 'RU') return t('bracket.thirdPlace')
  return t('bracket.winnerOf', { stage: t(`stageShort.${src.stage}`) })
}
