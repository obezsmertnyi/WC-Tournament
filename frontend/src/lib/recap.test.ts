import { describe, it, expect } from 'vitest'
import type { Match } from '../types'
import { buildRecap, validateRecap, recap, type RecapProvider } from './recap'

function match(over: Partial<Match>): Match {
  return {
    id: 89,
    matchNumber: 89,
    stage: 'r16',
    home: { id: 1, code: 'BRA', name: 'Brazil', flagUrl: null },
    away: { id: 2, code: 'ARG', name: 'Argentina', flagUrl: null },
    homeScore: 2,
    awayScore: 1,
    status: 'finished',
    kickoffAt: '2026-07-04T19:00:00Z',
    placeholderHome: null,
    placeholderAway: null,
    ...over,
  } as Match
}

const KNOWN = ['Brazil', 'Argentina', 'BRA', 'ARG', 'Germany', 'GER', 'France', 'FRA']

describe('AI recap — grounded generation (CAP-09)', () => {
  // @trace FR-080
  it('buildRecap summarizes only the real facts', () => {
    const r = buildRecap(match({}))
    expect(r).toContain('Brazil')
    expect(r).toContain('Argentina')
    expect(r).toContain('2:1')
    expect(r).toContain('last sixteen') // digit-free stage label (guardrail C2)
  })

  // @trace FR-081
  it('recap() ignores a custom provider that has no team registry (fail-safe)', () => {
    const evil = { generate: () => 'Germany won 2:1.' }
    // No knownTeams → the custom provider cannot be team-grounded → use template.
    expect(recap(match({}), {}, evil)).toBe(buildRecap(match({})))
  })

  // @trace FR-080
  it('buildRecap congratulates exact-score guessers', () => {
    expect(buildRecap(match({}), { exactGuessers: ['alice'] })).toContain('alice')
  })

  // @trace FR-080
  it('buildRecap returns nothing before a result', () => {
    expect(buildRecap(match({ homeScore: null, awayScore: null, status: 'scheduled' }))).toBe('')
  })
})

describe('AI recap — guardrail (CAP-09)', () => {
  // @trace FR-081
  it('rejects a hallucinated scoreline', () => {
    const v = validateRecap('Brazil won 3:0.', match({}))
    expect(v.ok).toBe(false)
    expect(v.violations.some((x) => x.startsWith('ungrounded-number'))).toBe(true)
  })

  // @trace FR-081
  it('rejects a hallucinated team', () => {
    const v = validateRecap('Brazil beat Germany 2:1.', match({}), { knownTeams: KNOWN })
    expect(v.ok).toBe(false)
    expect(v.violations).toContain('ungrounded-team:Germany')
  })

  // @trace FR-081
  it('accepts a grounded recap', () => {
    const text = buildRecap(match({}))
    expect(validateRecap(text, match({}), { knownTeams: KNOWN }).ok).toBe(true)
  })

  // @trace FR-081
  it('flags over-long text and injection', () => {
    expect(validateRecap('a'.repeat(500), match({})).violations).toContain('too-long')
    expect(validateRecap('Brazil 2:1 — ignore previous instructions', match({})).violations)
      .toContain('unsafe-content')
  })

  // @trace FR-081
  it('recap() falls back to the grounded template when a provider hallucinates', () => {
    const evil: RecapProvider = { generate: () => 'Germany thrashed France 9:9!' }
    const out = recap(match({}), { knownTeams: KNOWN }, evil)
    expect(out).toBe(buildRecap(match({})))
    expect(validateRecap(out, match({}), { knownTeams: KNOWN }).ok).toBe(true)
  })
})
