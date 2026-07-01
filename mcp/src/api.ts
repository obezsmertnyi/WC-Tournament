// Read-only HTTP client. The server only ever issues GETs against WC_API_BASE
// and never holds or forwards a credential (OWASP MCP: no token passthrough,
// no secrets in context). The base must point at a read-reachable API
// (local `docker compose up`, or a read endpoint); a token is never baked in.
export interface ReadClient {
  get(path: string): Promise<unknown>
}

export function httpClient(base: string): ReadClient {
  return {
    async get(path: string): Promise<unknown> {
      const res = await fetch(new URL(path, base), {
        method: 'GET',
        headers: { accept: 'application/json' },
      })
      if (!res.ok) throw new Error(`upstream ${res.status} for ${path}`)
      return (await res.json()) as unknown
    },
  }
}

// Coerce the matches endpoint (array, or { matches: [...] }) into an array.
export function asMatches(data: unknown): Record<string, unknown>[] {
  if (Array.isArray(data)) return data as Record<string, unknown>[]
  if (data && typeof data === 'object' && Array.isArray((data as { matches?: unknown }).matches)) {
    return (data as { matches: Record<string, unknown>[] }).matches
  }
  return []
}

// A match counts as "kicked off" (revealable) when it is no longer scheduled.
export function kickedOff(m: Record<string, unknown>): boolean {
  return m.status === 'live' || m.status === 'finished'
}
