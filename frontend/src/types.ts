export type Team = { name: string; code: string; flagUrl: string } | null

export type Stage = 'group' | 'r32' | 'r16' | 'qf' | 'sf' | 'final' | 'third'

export type MatchStatus = 'scheduled' | 'live' | 'finished'

export interface Venue {
  stadium: string
  city: string
  country: string
}

export interface Match {
  id: number
  stage: Stage
  group: string | null
  matchNumber: number
  kickoffAt: string // UTC ISO
  status: MatchStatus
  home: Team
  away: Team
  homeScore: number | null
  awayScore: number | null
  venue: Venue
  placeholderHome: string | null
  placeholderAway: string | null
}

export interface StandingRow {
  teamId: number
  name: string
  code: string
  flagUrl: string
  played: number
  win: number
  draw: number
  loss: number
  gf: number
  ga: number
  gd: number
  points: number
  rank: number
}

export interface GroupStanding {
  group: string
  rows: StandingRow[]
}

// ── M2: users, predictions, competition, audit ──────────────────────────────

export type Role = 'user' | 'admin'

export interface User {
  id: string
  nickname: string
  avatarUrl: string | null
  favoriteTeamCode: string | null
  role: Role
}

/** A player as returned by the admin roster endpoint (`GET /api/admin/users`). */
export interface AdminPlayer {
  id: string
  nickname: string
  avatarUrl: string | null
  favoriteTeamCode: string | null
  role: Role
}

/** A user's own prediction for one match. */
export interface MyPrediction {
  matchId: number
  home: number
  away: number
  winnerPickTeamId: number | null
}

/** A single revealed prediction (returned only after kickoff). */
export interface RevealedPrediction {
  userId: string
  nickname: string
  avatarUrl: string | null
  home: number
  away: number
  winnerPickTeamId: number | null
  points: number | null
}

/**
 * Per-match reveal response. Before kickoff the API hides everything
 * (`{ locked: true, predictions: [] }`); after kickoff it returns the array.
 */
export interface MatchRevealLocked {
  locked: true
  predictions: []
}
export type MatchReveal = RevealedPrediction[] | MatchRevealLocked

export interface LeaderboardEntry {
  userId: string
  nickname: string
  avatarUrl: string | null
  points: number
  exactCount: number
  played: number
}

export type AuditAction =
  | 'prediction.submitted'
  | 'prediction.updated'
  | 'result.updated'
  | 'champion.picked'
  | string

export interface AuditEntry {
  actor: string
  action: AuditAction
  matchId: number | null
  createdAt: string // UTC ISO
}

export interface ChampionBonus {
  teamId: number | null
  /** Optional metadata the backend may attach; tolerated, not required. */
  lockedAt?: string | null
  points?: number | null
}

/** A team option for the champion picker (`GET /api/teams`). */
export interface TeamOption {
  id: number
  name: string
  code: string
  flagUrl: string
  group: string | null
}

/**
 * A single tournament-wide bonus pick (`GET /api/bonus/me` → `{ picks: [...] }`).
 * `pickRef` is the picked entity reference — for `kind === 'champion'` it is the
 * teamId as a string. `tierPoints` is the locked-in points for the pick (null
 * when the bonus is disabled / no tier applies). `lockedAt` is the timestamp the
 * pick was last stamped (a late edit re-stamps it and may drop the tier).
 */
export interface BonusPick {
  kind: string
  pickRef: string
  tierPoints: number | null
  lockedAt: string | null
}
