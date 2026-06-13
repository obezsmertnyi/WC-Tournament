import type {
  Match,
  GroupStanding,
  User,
  AdminPlayer,
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
 * Fetch all tournament matches.
 * The backend owns the /api namespace and returns `{ "matches": [...] }`.
 */
export async function fetchMatches(signal?: AbortSignal): Promise<Match[]> {
  const res = await fetch('/api/matches', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { matches?: unknown }
  if (!data || !Array.isArray(data.matches)) {
    throw new Error('Unexpected response shape: expected { matches: [...] }')
  }
  return data.matches as Match[]
}

/**
 * Fetch computed group-stage standings.
 * The backend returns `{ "groups": [{ group, rows: [...] }] }`.
 */
export async function fetchStandings(signal?: AbortSignal): Promise<GroupStanding[]> {
  const res = await fetch('/api/standings', { ...withCreds, signal })
  if (!res.ok) throw new ApiError(res.status)
  const data = (await res.json()) as { groups?: unknown }
  if (!data || !Array.isArray(data.groups)) {
    throw new Error('Unexpected response shape: expected { groups: [...] }')
  }
  return data.groups as GroupStanding[]
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
   */
  forUserId?: string
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
  const res = await fetch('/api/bonus/champion', {
    ...withCreds,
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ teamId }),
    signal,
  })
  if (!res.ok) throw new ApiError(res.status)
  return (await res.json()) as BonusPick
}
