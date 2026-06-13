import { useCallback, useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { AuditEntry } from '../types'
import { fetchAudit } from '../lib/api'
import { formatRelativeTime } from '../lib/fixtures'
import { ErrorState } from '../components/states'

const POLL_MS = 30_000

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; entries: AuditEntry[] }

/**
 * Map a raw audit action to a localized, human-readable sentence. Unknown
 * actions fall back to a generic line so the feed never shows raw keys, and we
 * NEVER surface predicted score values (the API doesn't return them).
 */
function describe(t: (k: string, o?: Record<string, unknown>) => string, e: AuditEntry): string {
  const actor = e.actor || t('audit.someone')
  const matchRef = e.matchId != null ? ` ${t('audit.matchRef', { id: e.matchId })}` : ''
  switch (e.action) {
    case 'prediction.submitted':
      return t('audit.predictionSubmitted', { actor }) + matchRef
    case 'prediction.updated':
      return t('audit.predictionUpdated', { actor }) + matchRef
    case 'result.updated':
      return t('audit.resultUpdated', { actor }) + matchRef
    case 'champion.picked':
      return t('audit.championPicked', { actor })
    default:
      return t('audit.generic', { actor, action: e.action }) + matchRef
  }
}

export default function Audit() {
  const { t } = useTranslation()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback((signal?: AbortSignal) => {
    fetchAudit(signal)
      .then((entries) => {
        if (signal?.aborted) return
        setState({ phase: 'ready', entries })
      })
      .catch((err) => {
        if (signal?.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setState((prev) => (prev.phase === 'ready' ? prev : { phase: 'error' }))
      })
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    const id = setInterval(() => load(), POLL_MS)
    return () => {
      controller.abort()
      clearInterval(id)
    }
  }, [load])

  return (
    <div className="mx-auto w-full max-w-2xl">
      <header className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
          {t('audit.title')}
        </h1>
        <p className="mt-1 text-sm text-muted">{t('audit.subtitle')}</p>
      </header>

      {state.phase === 'loading' && (
        <div className="space-y-2">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="h-12 animate-pulse rounded-xl bg-white/[0.05]" />
          ))}
        </div>
      )}

      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}

      {state.phase === 'ready' &&
        (state.entries.length === 0 ? (
          <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
            {t('audit.empty')}
          </p>
        ) : (
          <ol className="relative space-y-1 border-l border-hairline pl-4">
            {state.entries.map((e, i) => (
              <motion.li
                key={`${e.createdAt}-${i}`}
                initial={{ opacity: 0, x: -6 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.25, delay: Math.min(i, 12) * 0.02 }}
                className="relative rounded-r-xl py-2 pl-2 pr-1"
              >
                <span className="absolute -left-[1.32rem] top-3.5 h-1.5 w-1.5 rounded-full bg-accent/70 shadow-[0_0_6px_1px_rgba(201,162,75,0.5)]" />
                <p className="text-sm leading-snug text-text">{describe(t, e)}</p>
                <p className="mt-0.5 text-[0.65rem] uppercase tracking-[0.12em] text-muted/60">
                  {formatRelativeTime(e.createdAt)}
                </p>
              </motion.li>
            ))}
          </ol>
        ))}
    </div>
  )
}
