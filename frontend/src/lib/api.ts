import type {
  Match,
  MatchStatus,
  MatchDetail,
  GroupStanding,
  Standings,
  ThirdPlaceRow,
  MyHistory,
  TopScorer,
  User,
  AdminPlayer,
  AccessLevel,
  MyPrediction,
  MatchReveal,
  LeaderboardEntry,
  AuditEntry,
  BonusPick,
  TeamOption,
} from '../types'

/**
 * Every call below carries the session cookie (`credentials: 'include'`) so the
 * backend can identify the signed-in user. The cookie is HttpOnly and set by
 * the auth endpoints; the frontend never reads or stores it directly.
 */
const withCreds: RequestInit = { credentials: 'include' }

/** Thrown so callers can branch on HTTP status (e.g. 401 anon, 409 locked). */
export class ApiError extends Error {
  status: number
  constructor(status: number, message?: string) {
    super(message ?? `Request failed with status ${status}`)
    this.name = 'ApiError'
    this.status = status
  }
}

/**
 * Run `fn`, and on a transient failure (network error / 5xx / bad shape) wait
 * briefly and retry once. Auth/validation errors (<500) and aborts are not
 * retried. This self-heals brief blips (e.g. a backend redeploy) so pages don't
 * fall into an error state that needs a manual refresh.
 */
async function withRetry<T>(fn: () => Promise<T>, signal?: AbortSignal): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    if (signal?.aborted) throw err
    if (err instanceof ApiError && err.status < 500) throw err
    await new Promise((r) => setTimeout(r, 700))
    return await fn()
  }
}

/**
 * Fetch all tournament matches.
 * The backend owns the /api namespace and returns `{ "matches": [...] }`.
 */
export async function fetchMatches(signal?: AbortSignal): Promise<Match[]> {
  return withRetry(async () => {
    const res = await fetch('/api/matches', { ...withCreds, signal })
    if (!res.ok) throw new ApiError(res.status)
    const data = (await res.json()) as { matches?: unknown }
    if (!data || !Array.isArray(data.matches)) {
      throw new Error('Unexpected response shape: expected { matches: [...] }')
    }
    return data.matches as Match[]
  }, signal)
}

/**
 * Fetch computed group-stage standings plus the cross-group ranking of the
 * third-placed teams. The backend returns
 * `{ "groups": [{ group, rows: [...] }], "thirdPlace": [...] }`.
 */
export async function fetchStandings(signal?: AbortSignal): Promise<Standings> {
  return withRetry(async () => {
    const res = await fetch('/api/standings', { ...withCreds, signal })
    if (!res.ok) throw new ApiError(res.status)
    const data = (await res.json()) as { groups?: unknown; thirdPlace?: unknown }
    if (!data || !Array.isArray(data.groups)) {
      throw new Error('Unexpected response shape: expected { groups: [...] }')
    }
    return {
      groups: data.groups as GroupStanding[],
      thirdPlace: Array.isArray(data.thirdPlace) ? (data.thirdPlace as ThirdPlaceRow[]) : [],
    }
  }, signal)
}

export type HealthState = 'loading' | 'online' | 'offline'

interface HealthResponse {
  status?: string
}

export async function fetchHealth(signal?: AbortSignal): Promise<boolean> {
  const res = await fetch('/api/healthz', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json().catch(() => ({}))) as HealthResponse
  return !data.status || ['ok', 'healthy', 'up'].includes(data.status.toLowerCase())
}

// ── Auth / session ──────────────────────────────────────────────────────────

/** Resolve the current session. Returns null on 401 (anonymous). */
export async function fetchMe(signal?: AbortSignal): Promise<User | null> {
  const res = await fetch('/api/me', { ...withCreds, signal })
  if (res.status === 401) return null
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as User
}

/** Dev login: exchange a nickname for a session cookie. */
export async function devLogin(nickname: string, signal?: AbortSignal): Promise<User> {
  const res = await fetch('/api/auth/dev-login', {
    ...withCreds,
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nickname }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as User
}

/** Admin login with the shared password (from .env on the server). */
export async function adminLogin(password: string, signal?: AbortSignal): Promise<User> {
  const res = await fetch('/api/auth/admin-login', {
    ...withCreds,
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as User
}

/** URL the "Continue with Google" affordance navigates to. */
export const GOOGLE_LOGIN_URL = '/api/auth/google/login'

export async function logout(signal?: AbortSignal): Promise<void> {
  const res = await fetch('/api/auth/logout', { ...withCreds, method: 'POST', signal })
  if (!res.ok) throw new ApiError(res.status)
}

/** Patch the signed-in user's profile. */
export interface ProfilePatch {
  nickname?: string
  favoriteTeamCode?: string | null
  avatarUrl?: string | null
}
export async function updateMe(patch: ProfilePatch, signal?: AbortSignal): Promise<User> {
  const res = await fetch('/api/me', {
    ...withCreds,
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(patch),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as User
}

// ── Predictions ─────────────────────────────────────────────────────────────

/** Fetch the signed-in user's predictions. */
export async function fetchMyPredictions(signal?: AbortSignal): Promise<MyPrediction[]> {
  const res = await fetch('/api/predictions/me', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { predictions?: unknown }
  if (!data || !Array.isArray(data.predictions)) return []
  return data.predictions as MyPrediction[]
}

export interface PredictionInput {
  home: number
  away: number
  winnerPickTeamId?: number | null
  /**
   * Admin-only: write this prediction on behalf of the given player. When set,
   * the backend bypasses the kickoff lock. Ignored (and rejected) for non-admins.
   * Numeric — the backend binds it as an int64 user id.
   */
  forUserId?: number
}

/** Upsert a prediction for a match. Throws ApiError(409) when locked. */
export async function savePrediction(
  matchId: number,
  input: PredictionInput,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch(`/api/predictions/${matchId}`, {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
}

/**
 * Fetch the signed-in user's personal results history (per-match prediction,
 * actual result, points) plus bonus picks. `GET /api/me/history`.
 */
export async function fetchMyHistory(signal?: AbortSignal): Promise<MyHistory> {
  return withRetry(async () => {
    const res = await fetch('/api/me/history', { ...withCreds, signal })
    if (!res.ok) throw new ApiError(res.status)
    return (await res.json()) as MyHistory
  }, signal)
}

/** Fetch the top goal scorers board (`GET /api/top-scorers?limit=N`). */
export async function fetchTopScorers(limit = 10, signal?: AbortSignal): Promise<TopScorer[]> {
  return withRetry(async () => {
    const res = await fetch(`/api/top-scorers?limit=${limit}`, { ...withCreds, signal })
    if (!res.ok) throw new ApiError(res.status)
    const data = (await res.json()) as { scorers?: unknown }
    return Array.isArray(data.scorers) ? (data.scorers as TopScorer[]) : []
  }, signal)
}

// ── Admin: player roster ─────────────────────────────────────────────────────

/** List all players (admin only). Returns `[{ id, nickname, … }]`. */
export async function fetchAdminUsers(signal?: AbortSignal): Promise<AdminPlayer[]> {
  const res = await fetch('/api/admin/users', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as unknown
  return Array.isArray(data) ? (data as AdminPlayer[]) : []
}

/** Create a player by nickname (admin only). */
export async function createPlayer(
  nickname: string,
  signal?: AbortSignal,
): Promise<AdminPlayer> {
  const res = await fetch('/api/admin/users', {
    ...withCreds,
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nickname }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as AdminPlayer
}

/** Delete a player by id (admin only). */
export async function deletePlayer(id: string, signal?: AbortSignal): Promise<void> {
  const res = await fetch(`/api/admin/users/${encodeURIComponent(id)}`, {
    ...withCreds,
    method: 'DELETE',
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
}

/** Set a player's demo access level (admin only). */
export async function setUserAccess(
  id: string,
  level: AccessLevel,
  signal?: AbortSignal,
): Promise<AdminPlayer> {
  const res = await fetch(`/api/admin/users/${encodeURIComponent(id)}/access`, {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ level }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as AdminPlayer
}

/** Read whether demo mode is enabled (admin only). */
export async function fetchDemoMode(signal?: AbortSignal): Promise<boolean> {
  const res = await fetch('/api/admin/demo', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { enabled?: boolean }
  return !!data.enabled
}

/** Enable or disable demo mode (admin only). */
export async function setDemoMode(enabled: boolean, signal?: AbortSignal): Promise<boolean> {
  const res = await fetch('/api/admin/demo', {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { enabled?: boolean }
  return !!data.enabled
}

/**
 * Set (or correct) the actual result of a match (admin only). The backend
 * recomputes scores for everyone's predictions. Works for any match, including
 * already-finished ones. Returns the updated match when the backend provides
 * it, or `null` when it answers with a bare `{ ok: true }`.
 *
 * `PUT /api/admin/matches/{id}/result` body `{ homeScore, awayScore, status? }`.
 */
export async function setMatchResult(
  matchId: number,
  homeScore: number,
  awayScore: number,
  status?: MatchStatus,
  signal?: AbortSignal,
): Promise<Match | null> {
  const res = await fetch(`/api/admin/matches/${matchId}/result`, {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ homeScore, awayScore, ...(status ? { status } : {}) }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json().catch(() => null)) as unknown
  if (data && typeof data === 'object' && 'id' in data) return data as Match
  return null
}

/** Reveal everyone's predictions for a match (locked until kickoff). */
export async function fetchMatchReveal(
  matchId: number,
  signal?: AbortSignal,
): Promise<MatchReveal> {
  const res = await fetch(`/api/matches/${matchId}/predictions`, { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as unknown
  if (Array.isArray(data)) return data as MatchReveal
  // Locked shape: { locked: true, predictions: [] }
  return { locked: true, predictions: [] }
}

/**
 * Fetch FIFA match statistics for a single match
 * (`GET /api/matches/{id}/detail`, RequireUser). The backend answers either
 * `{ available: false }` (no stats published yet) or the full statistics
 * payload (`{ available: true, … }`). Callers branch on the `available`
 * discriminant. Retries once on a transient failure (see `withRetry`).
 */
export async function fetchMatchDetail(
  matchId: number,
  signal?: AbortSignal,
): Promise<MatchDetail> {
  return withRetry(async () => {
    const res = await fetch(`/api/matches/${matchId}/detail`, { ...withCreds, signal })
    if (!res.ok) throw new ApiError(res.status)
    const data = (await res.json()) as unknown
    if (!data || typeof data !== 'object' || !('available' in data)) {
      throw new Error('Unexpected response shape: expected { available: … }')
    }
    return data as MatchDetail
  }, signal)
}

// ── Competition / audit ─────────────────────────────────────────────────────

export async function fetchLeaderboard(signal?: AbortSignal): Promise<LeaderboardEntry[]> {
  const res = await fetch('/api/leaderboard', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as unknown
  return Array.isArray(data) ? (data as LeaderboardEntry[]) : []
}

export async function fetchAudit(signal?: AbortSignal): Promise<AuditEntry[]> {
  const res = await fetch('/api/audit', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as unknown
  return Array.isArray(data) ? (data as AuditEntry[]) : []
}

// ── Teams ─────────────────────────────────────────────────────────────────--

/**
 * Fetch the full team field (`GET /api/teams` → `{ teams: [...] }`). Used by the
 * champion-bonus picker so the user can choose from all 48 nations.
 */
export async function fetchTeams(signal?: AbortSignal): Promise<TeamOption[]> {
  const res = await fetch('/api/teams', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { teams?: unknown }
  if (!data || !Array.isArray(data.teams)) return []
  return data.teams as TeamOption[]
}

// ── Champion bonus ──────────────────────────────────────────────────────────

/**
 * Fetch the signed-in user's tournament-wide bonus picks
 * (`GET /api/bonus/me` → `{ picks: [...] }`).
 */
export async function fetchMyBonus(signal?: AbortSignal): Promise<BonusPick[]> {
  const res = await fetch('/api/bonus/me', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { picks?: unknown }
  if (!data || !Array.isArray(data.picks)) return []
  return data.picks as BonusPick[]
}

/**
 * Upsert the champion pick (`PUT /api/bonus/champion {teamId}`). The backend
 * resolves the time-tiered points and re-stamps `lockedAt` on every change.
 * Throws ApiError(409) once the knockout stage starts (hard lock).
 */
export async function saveChampionBonus(
  teamId: number,
  signal?: AbortSignal,
): Promise<BonusPick> {
  return saveTeamBonus('champion', teamId, signal)
}

/** Upsert the losing-finalist pick (`PUT /api/bonus/finalist {teamId}`). */
export async function saveFinalistBonus(
  teamId: number,
  signal?: AbortSignal,
): Promise<BonusPick> {
  return saveTeamBonus('finalist', teamId, signal)
}

/** Shared helper for team-referenced bonus picks (champion / finalist). */
async function saveTeamBonus(
  kind: 'champion' | 'finalist',
  teamId: number,
  signal?: AbortSignal,
): Promise<BonusPick> {
  const res = await fetch(`/api/bonus/${kind}`, {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ teamId }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as BonusPick
}

/**
 * Upsert the top-scorer pick (`PUT /api/bonus/top-scorer {player}`). Free-text
 * player name. Throws ApiError(409) once the Round of 16 starts (hard lock).
 */
export async function saveTopScorerBonus(
  player: string,
  signal?: AbortSignal,
): Promise<BonusPick> {
  const res = await fetch('/api/bonus/top-scorer', {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ player }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as BonusPick
}
