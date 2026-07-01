import { describe, it, expect } from 'vitest'
import { evalClient } from './fixtures.js'
import { listFixtures, groupStandings, leaderboard, bracket, playerPredictions } from '../src/tools.js'

const client = evalClient()

describe('MCP tool contracts (golden fixtures)', () => {
  // @trace FR-070
  it('list_fixtures shapes rows and filters by stage', async () => {
    const all = (await listFixtures.run({}, client)) as unknown[]
    expect(all).toHaveLength(3)
    const ko = (await listFixtures.run({ stage: 'r16' }, client)) as unknown[]
    expect(ko).toHaveLength(2)
    expect(ko[0]).toMatchObject({ stage: 'r16' })
  })

  // @trace FR-070
  it('list_fixtures filters by date', async () => {
    const day = (await listFixtures.run({ date: '2026-07-04' }, client)) as unknown[]
    expect(day).toHaveLength(2)
  })

  // @trace FR-070
  it('group_standings filters to one group', async () => {
    const b = (await groupStandings.run({ group: 'B' }, client)) as unknown[]
    expect(b).toHaveLength(1)
    expect(b[0]).toMatchObject({ group: 'B' })
  })

  // @trace FR-070
  it('leaderboard respects the limit', async () => {
    const top1 = (await leaderboard.run({ limit: 1 }, client)) as unknown[]
    expect(top1).toHaveLength(1)
    expect(top1[0]).toMatchObject({ nickname: 'alice' })
  })

  // @trace FR-070
  it('bracket excludes the group stage', async () => {
    const rows = (await bracket.run({}, client)) as unknown[]
    expect(rows).toHaveLength(2)
    expect(rows.every((r) => (r as { stage: string }).stage !== 'group')).toBe(true)
  })

  // @trace FR-070, FR-072
  it('player_predictions returns only kicked-off matches for the player (reveal lock)', async () => {
    const res = (await playerPredictions.run({ nickname: 'alice' }, client)) as { revealed: unknown[] }
    // alice predicted the scheduled group match (#1) too, but it must NOT appear
    // — only kicked-off matches (#89, #90) are revealed.
    expect(res.revealed).toHaveLength(2)
    expect(res.revealed).toContainEqual({ matchNumber: 89, home: 2, away: 1, points: 3 })
  })
})

describe('MCP input validation (reject before any fetch)', () => {
  // @trace FR-071
  it('rejects unknown parameters (strict schema)', () => {
    expect(() => listFixtures.schema.parse({ foo: 1 })).toThrow()
  })

  // @trace FR-071
  it('rejects a malformed date', () => {
    expect(() => listFixtures.schema.parse({ date: '04-07-2026' })).toThrow()
  })

  // @trace FR-071
  it('rejects an over-long nickname', () => {
    expect(() => playerPredictions.schema.parse({ nickname: 'x'.repeat(41) })).toThrow()
  })

  // @trace FR-071
  it('rejects an out-of-range limit', () => {
    expect(() => leaderboard.schema.parse({ limit: 999 })).toThrow()
    expect(() => playerPredictions.schema.parse({ nickname: 'a', limit: 0 })).toThrow()
  })
})
