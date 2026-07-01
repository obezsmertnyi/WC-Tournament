import type { ReadClient } from '../src/api.js'

// Golden API responses, keyed by path. The eval client serves these so the
// suite is fully offline (no network, no secrets) — the quality bar for the
// tool contracts and safety invariants.
export const FIXTURES: Record<string, unknown> = {
  '/api/matches': [
    { matchNumber: 1, id: 1, stage: 'group', group: 'A', home: { code: 'MEX' }, away: { code: 'CAN' },
      kickoffAt: '2026-06-11T19:00:00Z', status: 'scheduled', homeScore: null, awayScore: null },
    { matchNumber: 89, id: 89, stage: 'r16', placeholderHome: 'W77', placeholderAway: 'W78',
      kickoffAt: '2026-07-04T19:00:00Z', status: 'finished', homeScore: 2, awayScore: 1 },
    { matchNumber: 90, id: 90, stage: 'r16', home: { code: 'BRA' }, away: { code: 'ARG' },
      kickoffAt: '2026-07-04T22:00:00Z', status: 'live', homeScore: 0, awayScore: 0 },
  ],
  '/api/standings': {
    groups: [
      { group: 'A', table: [{ team: 'MEX', pts: 6 }] },
      { group: 'B', table: [{ team: 'ESP', pts: 9 }] },
    ],
  },
  '/api/leaderboard': [
    { nickname: 'alice', points: 24, matchPoints: 18, bonusPoints: 6 },
    { nickname: 'bob', points: 17, matchPoints: 17, bonusPoints: 0 },
  ],
  '/api/matches/89/predictions': [
    { nickname: 'alice', home: 2, away: 1, points: 3 },
    { nickname: 'bob', home: 1, away: 1, points: 0 },
  ],
  '/api/matches/90/predictions': [{ nickname: 'alice', home: 1, away: 0, points: 0 }],
  // A scheduled match is not queried by player_predictions, but if it were the
  // reveal endpoint returns the locked shape.
  '/api/matches/1/predictions': { locked: true, predictions: [] },
}

export function evalClient(): ReadClient {
  return {
    async get(path: string): Promise<unknown> {
      if (!(path in FIXTURES)) throw new Error(`eval: unexpected upstream path ${path}`)
      return FIXTURES[path]
    },
  }
}
