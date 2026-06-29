import type { Match } from '../types'
import { teamName } from './teamNames'

/** Index matches by their FIFA match number so "W75"/"RU101" can be resolved. */
export function buildMatchByNumber(matches: Match[]): Map<number, Match> {
  const m = new Map<number, Match>()
  for (const x of matches) if (x.matchNumber != null) m.set(x.matchNumber, x)
  return m
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
