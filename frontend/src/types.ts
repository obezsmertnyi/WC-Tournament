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

/**
 * One team in the cross-group "best third-placed teams" ranking. Carries the
 * usual standings stats plus the source group letter and whether the team
 * currently sits in a Round-of-32 qualifying slot (top 8).
 */
export interface ThirdPlaceRow extends StandingRow {
  group: string
  qualified: boolean
}

/** Full standings payload: per-group tables + the third-placed ranking. */
export interface Standings {
  groups: GroupStanding[]
  thirdPlace: ThirdPlaceRow[]
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
  /** Grand total = matchPoints + bonusPoints. */
  points: number
  /** Points from match predictions only. */
  matchPoints: number
  /** Points from awarded tournament bonuses (champion/finalist/top scorer). */
  bonusPoints: number
  exactCount: number
  played: number
}

// ── Personal results history (GET /api/me/history) ───────────────────────────

export interface HistoryTeam {
  code: string
  name: string
  flagUrl: string
}

export interface HistoryMatch {
  matchId: number
  stage: Stage
  group: string
  kickoffAt: string | null
  status: MatchStatus
  home: HistoryTeam
  away: HistoryTeam
  homeScore: number | null
  awayScore: number | null
  predHome: number
  predAway: number
  winnerPickTeamId: number | null
  points: number
  exact: boolean
  /** True once the match is scored (points materialized); else pending. */
  scored: boolean
}

export interface HistoryBonus {
  kind: string
  pickRef: string
  /** Resolved team for champion/finalist picks; null for the top-scorer (player). */
  team: HistoryTeam | null
  tierPoints: number | null
  awarded: boolean
}

export interface MyHistory {
  matches: HistoryMatch[]
  bonuses: HistoryBonus[]
  matchPoints: number
  bonusPoints: number
  total: number
}

/** One player on the top-scorers board (`GET /api/top-scorers`). */
export interface TopScorer {
  rank: number
  name: string
  teamCode: string
  goals: number
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

// ── Match detail (FIFA statistics) ───────────────────────────────────────────

/** A single player entry in a lineup. */
export interface DetailPlayer {
  name: string
  shirtNumber: number
  position: string
  captain: boolean
  pictureUrl?: string
}

/** One team's lineup (formation + players). */
export interface DetailLineup {
  teamName: string
  formation: string
  players: DetailPlayer[]
}

/** Which side a match event belongs to. */
export type DetailSide = 'home' | 'away'

export interface DetailGoal {
  team: DetailSide
  scorer: string
  assist?: string
  minute: string
  /** FIFA goal type — numeric code (0=goal, etc.) or a label. */
  type: string | number
}

export interface DetailCard {
  team: DetailSide
  player: string
  minute: string
  /** FIFA card — numeric code (1=yellow, 2/3=red) or a colour label. */
  card: string | number
}

export interface DetailSubstitution {
  team: DetailSide
  playerIn: string
  playerOut: string
  minute: string
}

export interface DetailOfficial {
  name: string
  type: string
}

/** Ball possession split (percentages summing to ~100). */
export interface DetailPossession {
  home: number
  away: number
}

/**
 * FIFA match statistics for a single match
 * (`GET /api/matches/{id}/detail`). When stats have not been published yet the
 * backend returns `{ available: false }`; otherwise `available: true` with the
 * full payload below. Use the `available` discriminant to branch.
 */
export type MatchDetail =
  | { available: false }
  | {
      available: true
      matchTime: string
      attendance: string
      stadium: string
      winnerTeamId: string
      possession: DetailPossession | null
      homeLineup: DetailLineup | null
      awayLineup: DetailLineup | null
      goals: DetailGoal[]
      cards: DetailCard[]
      substitutions: DetailSubstitution[]
      officials: DetailOfficial[]
      homePenaltyScore?: number
      awayPenaltyScore?: number
      aggregateHomeScore?: number
      aggregateAwayScore?: number
    }
