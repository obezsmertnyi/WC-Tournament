import type { Match } from '../types'

// AI match recap — grounded generation behind a guardrail (ADR-0016,
// docs/features/recap/spec.md). The default provider is a deterministic grounded
// template (no API key, offline); whatever produces the prose, it must pass
// validateRecap before display, so the feature cannot hallucinate a score/team.

// Digit-free stage phrasing on purpose: the number guardrail grounds recap
// numbers against the scoreline ONLY, so the recap text must not carry stray
// digits (a stage like "round of 32" would otherwise let a fabricated "32" pass).
const STAGE_LABEL: Record<string, string> = {
  group: 'the group stage',
  r32: 'the first knockout round',
  r16: 'the last sixteen',
  qf: 'the quarter-final',
  sf: 'the semi-final',
  third: 'the third-place play-off',
  final: 'the final',
}

export interface RecapOptions {
  /** Nicknames of players who nailed the exact score (for a congrats line). */
  exactGuessers?: string[]
  /** Full set of tournament team tokens (codes/names) for the anti-hallucination check. */
  knownTeams?: string[]
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
  return m.homeScore != null && m.awayScore != null && m.home != null && m.away != null
}

/** Grounded template recap (the trusted default provider). Empty before a result. */
export function buildRecap(m: Match, opts: RecapOptions = {}): string {
  if (!hasResult(m)) return ''
  const home = m.home!.name
  const away = m.away!.name
  const hs = m.homeScore as number
  const as = m.awayScore as number
  const stage = STAGE_LABEL[m.stage] ?? 'the match'

  let line: string
  if (hs > as) line = `${home} beat ${away} ${hs}:${as} in ${stage}.`
  else if (as > hs) line = `${away} beat ${home} ${as}:${hs} in ${stage}.`
  else line = `${home} and ${away} drew ${hs}:${as} in ${stage}.`

  const guessers = (opts.exactGuessers ?? []).filter(Boolean)
  if (guessers.length) {
    line += ` Exact-score call by ${guessers.join(', ')} — nicely done.`
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
