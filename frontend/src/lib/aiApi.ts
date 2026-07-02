// Client for the football AI assistant (ADR-0017). Streams chat via fetch +
// ReadableStream (not EventSource — we need the auth cookie + a POST body) and
// fetches structured cards. All calls send the session cookie (credentials).

export interface AiTurn {
  role: 'user' | 'model'
  text: string
}

export interface AiCard {
  name: string
  full_name?: string
  country: string
  club?: string
  position?: string
  achievements?: string[]
  summary: string
  confidence: 'high' | 'medium' | 'low'
  imageUrl?: string
  stats?: { label: string; value: string }[]
}

export type CardResult =
  | AiCard
  | { refused: true; message: string }
  | { clarify: true; message: string } // ambiguous name (e.g. "Роналду") — ask which entity

const creds: RequestInit = { credentials: 'include' }

/** Whether the AI backend is wired (false → the tab shows an "unavailable" note). */
export async function aiStatus(signal?: AbortSignal): Promise<boolean> {
  try {
    const res = await fetch('/api/ai/status', { ...creds, signal })
    if (!res.ok) return false
    const data = (await res.json()) as { available?: boolean }
    return !!data.available
  } catch {
    return false
  }
}

/**
 * Stream a chat reply. Calls onToken for each delta. Resolves when the stream
 * ends; throws on a pre-stream HTTP error (503/429/400) with a short code.
 */
export async function streamChat(
  message: string,
  history: AiTurn[],
  onToken: (t: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch('/api/ai/chat', {
    ...creds,
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message, history }),
    signal,
  })
  if (!res.ok || !res.body) {
    const data = (await res.json().catch(() => ({}))) as { error?: string }
    throw new Error(data.error || `http_${res.status}`)
  }
  const reader = res.body.getReader()
  const dec = new TextDecoder()
  let buf = ''
  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buf += dec.decode(value, { stream: true })
    const events = buf.split('\n\n')
    buf = events.pop() ?? ''
    for (const evt of events) {
      const line = evt.split('\n').find((l) => l.startsWith('data:'))
      if (!line) continue
      const payload = line.slice(5).trim()
      if (evt.includes('event: error')) throw new Error('ai_error')
      if (evt.includes('event: done')) return
      try {
        onToken(JSON.parse(payload) as string)
      } catch {
        /* skip non-token control frames */
      }
    }
  }
}

/** Fetch a structured club/player card (or a refusal). */
export async function fetchCard(query: string, signal?: AbortSignal): Promise<CardResult> {
  const res = await fetch('/api/ai/card', {
    ...creds,
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query }),
    signal,
  })
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string }
    throw new Error(data.error || `http_${res.status}`)
  }
  return (await res.json()) as CardResult
}
