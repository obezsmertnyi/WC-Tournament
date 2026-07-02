import type { Match } from '../types'
import { teamName } from './teamNames'

// AI match recap — grounded generation behind a guardrail (ADR-0016,
// docs/features/recap/spec.md). The default provider is a deterministic grounded
// template (no API key, offline); whatever produces the prose, it must pass
// validateRecap before display, so the feature cannot hallucinate a score/team.
//
// It deliberately does NOT restate the scoreline (that's shown right above the
// recap) — it leads with pool insight (who nailed the exact score). Bilingual
// (UA/EN). Digit-free on purpose: the number guardrail grounds numbers against the
// scoreline ONLY, so the text must carry no stray digits (a stage like "round of
// 32" or a "3 of 5" count would otherwise let a fabricated number pass).
const STAGE_LABEL: Record<string, { en: string; uk: string }> = {
  group: { en: 'the group stage', uk: 'груповому етапі' },
  r32: { en: 'the first knockout round', uk: 'першому раунді плейоф' },
  r16: { en: 'the last sixteen', uk: 'стадії останніх шістнадцяти' },
  qf: { en: 'the quarter-final', uk: 'чвертьфіналі' },
  sf: { en: 'the semi-final', uk: 'півфіналі' },
  third: { en: 'the third-place play-off', uk: 'матчі за третє місце' },
  final: { en: 'the final', uk: 'фіналі' },
}

export interface RecapOptions {
  /** Nicknames of players who nailed the exact score (for a congrats line). */
  exactGuessers?: string[]
  /** Full set of tournament team tokens (codes/names) for the anti-hallucination check. */
  knownTeams?: string[]
  /** UI language ('uk' | 'en'); controls prose + localized team names. Default 'en'. */
  lang?: string
  /** Localized name of the team that advanced (knockout draw decided on penalties). */
  advancer?: string
}

export interface RecapProvider {
  generate(match: Match, opts?: RecapOptions): string
}

const MAX_LEN = 400

function teamTokens(m: Match): string[] {
  const t: string[] = []
  if (m.home) t.push(m.home.code, m.home.name)
  if (m.away) t.push(m.away.code, m.away.name)
  return t.filter(Boolean)
}

/** The only numbers a recap may contain: the actual scoreline. */
function factNumbers(m: Match): Set<string> {
  const out = new Set<string>()
  if (m.homeScore != null) out.add(String(m.homeScore))
  if (m.awayScore != null) out.add(String(m.awayScore))
  return out
}

function hasResult(m: Match): boolean {
  // Only FINISHED matches get a result recap — a live match has scores too, but it
  // isn't full-time (our hard-won rule: never treat a live score as a final result).
  return (
    m.status === 'finished' &&
    m.homeScore != null &&
    m.awayScore != null &&
    m.home != null &&
    m.away != null
  )
}

/**
 * Grounded template recap (the trusted default provider). Empty before a result.
 * Leads with the outcome + pool insight, NOT the scoreline (shown above). Bilingual.
 */
export function buildRecap(m: Match, opts: RecapOptions = {}): string {
  if (!hasResult(m)) return ''
  const uk = opts.lang === 'uk'
  const home = teamName(m.home!.code, m.home!.name, opts.lang)
  const away = teamName(m.away!.code, m.away!.name, opts.lang)
  const hs = m.homeScore as number
  const as = m.awayScore as number
  const stageEntry = STAGE_LABEL[m.stage]
  const stage = stageEntry ? (uk ? stageEntry.uk : stageEntry.en) : uk ? 'матчі' : 'the match'

  // Label form (no conjugated verb taking the team name as object) so Ukrainian
  // stays grammatical without per-team accusative/gender declension.
  let line: string
  if (hs === as) {
    line = uk ? `${home} — ${away}, ${stage}: нічия.` : `${home} vs ${away}, ${stage}: a draw.`
    if (opts.advancer) {
      line += uk
        ? ` У серії пенальті далі — ${opts.advancer}.`
        : ` ${opts.advancer} advanced on penalties.`
    }
  } else {
    const winner = hs > as ? home : away
    line = uk
      ? `${home} — ${away}, ${stage}. Переможець — ${winner}.`
      : `${home} vs ${away}, ${stage}. Winner: ${winner}.`
  }

  const guessers = (opts.exactGuessers ?? []).filter(Boolean)
  if (guessers.length) {
    line += uk
      ? ` Точний рахунок вгадали: ${guessers.join(', ')} — чудово.`
      : ` Exact-score call by ${guessers.join(', ')} — nicely done.`
  } else {
    line += uk ? ` Точний рахунок не вгадав ніхто.` : ` Nobody nailed the exact score.`
  }
  return line
}

/**
 * Anti-hallucination guardrail. A candidate recap is grounded only if every
 * number in it appears in the match facts and it names no team outside this
 * match. Also bounds length and flags injection/markup. Returns violations.
 */
export function validateRecap(text: string, m: Match, opts: RecapOptions = {}): {
  ok: boolean
  violations: string[]
} {
  const violations: string[] = []

  const allowedNums = factNumbers(m)
  for (const n of text.match(/\d+/g) ?? []) {
    if (!allowedNums.has(n)) violations.push(`ungrounded-number:${n}`)
  }

  const own = new Set(teamTokens(m).map((s) => s.toLowerCase()))
  for (const tok of opts.knownTeams ?? []) {
    const t = tok.toLowerCase()
    if (own.has(t)) continue
    const re = new RegExp(`\\b${t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\b`, 'i')
    if (re.test(text)) violations.push(`ungrounded-team:${tok}`)
  }

  if (text.length > MAX_LEN) violations.push('too-long')
  // Control characters, markup, template braces, or a prompt-injection tell.
  const controlChars = /[\x00-\x1f]/
  const injection = /ignore (all |the )?previous|[`<>]|\{\{/i
  if (controlChars.test(text) || injection.test(text)) violations.push('unsafe-content')

  return { ok: violations.length === 0, violations }
}

const groundedProvider: RecapProvider = { generate: buildRecap }

/**
 * Produce a recap to display: run the provider, validate it, and fall back to
 * the grounded template if the candidate fails the guardrail. The returned text
 * is always grounded (FR-080/FR-081). Raw provider output is never displayed.
 */
export function recap(m: Match, opts: RecapOptions = {}, provider: RecapProvider = groundedProvider): string {
  // Fail-safe for a non-grounded provider (e.g. a future LLM, FR-082): the team
  // guardrail can only catch a hallucinated team when a team registry is given.
  // Without one, don't trust a custom provider — use the grounded template.
  if (provider !== groundedProvider && !(opts.knownTeams && opts.knownTeams.length)) {
    return buildRecap(m, opts)
  }
  const candidate = provider.generate(m, opts)
  if (validateRecap(candidate, m, opts).ok) return candidate
  return buildRecap(m, opts)
}
