import type { Match } from '../types'

/**
 * Fetch all tournament matches.
 * The dev proxy maps /api -> backend (stripping the /api prefix), so the
 * browser-facing path is /api/matches.
 */
export async function fetchMatches(signal?: AbortSignal): Promise<Match[]> {
  const res = await fetch('/api/matches', { signal })
  if (!res.ok) throw new Error(`Request failed with status ${res.status}`)
  const data = (await res.json()) as unknown
  if (!Array.isArray(data)) {
    throw new Error('Unexpected response shape: expected an array of matches')
  }
  return data as Match[]
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
