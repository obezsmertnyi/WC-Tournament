import { describe, it, expect } from 'vitest'
import type { Match } from '../types'
import { bracketOrderKey } from './bracket'

// Regression guard for the tree-ordering fix: rounds must order by bracket
// geometry (feeder links), not by FIFA match number. A feeder that fills the
// home slot of its parent must sort before the one filling the away slot.
function m(partial: Partial<Match>): Match {
  return {
    id: partial.matchNumber ?? 0,
    matchNumber: partial.matchNumber ?? null,
    stage: partial.stage ?? 'r32',
    home: null,
    away: null,
    homeScore: null,
    awayScore: null,
    status: 'scheduled',
    kickoffAt: '',
    placeholderHome: partial.placeholderHome ?? null,
    placeholderAway: partial.placeholderAway ?? null,
    ...partial,
  } as Match
}

describe('bracketOrderKey', () => {
  it('orders feeders by the slot they fill in the parent tie', () => {
    // R32 #1 and #2 feed R16 #10 (home=W1, away=W2).
    const matches = [
      m({ matchNumber: 1, stage: 'r32' }),
      m({ matchNumber: 2, stage: 'r32' }),
      m({ matchNumber: 10, stage: 'r16', placeholderHome: 'W1', placeholderAway: 'W2' }),
    ]
    const key = bracketOrderKey(matches)
    // Parent sits at the root of this mini-tree.
    expect(key.get(10)).toBe(0)
    // Home-slot feeder (#1) sorts before away-slot feeder (#2).
    expect(key.get(1)!).toBeLessThan(key.get(2)!)
  })
})
