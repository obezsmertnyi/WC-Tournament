import type { Match, GroupStanding } from '../types'

/**
 * Fetch all tournament matches.
 * The backend owns the /api namespace and returns `{ "matches": [...] }`.
 */
export async function fetchMatches(signal?: AbortSignal): Promise<Match[]> {
  const res = await fetch('/api/matches', { signal })
  if (!res.ok) throw new Error(`Request failed with status ${res.status}`)
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
  const res = await fetch('/api/standings', { signal })
  if (!res.ok) throw new Error(`Request failed with status ${res.status}`)
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
  const res = await fetch('/api/healthz', { signal })
  if (!res.ok) throw new Error(`status ${res.status}`)
  const data = (await res.json().catch(() => ({}))) as HealthResponse
  return !data.status || ['ok', 'healthy', 'up'].includes(data.status.toLowerCase())
}
