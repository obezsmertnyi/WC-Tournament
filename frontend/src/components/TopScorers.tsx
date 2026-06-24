import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TopScorer } from '../types'
import { fetchTopScorers } from '../lib/api'
import { teamName } from '../lib/teamNames'
import Flag from './Flag'
import { ErrorState } from './states'

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; rows: TopScorer[] }

/**
 * Top goal scorers board (top 10), aggregated server-side from finished-match
 * data. Player name, national-team flag and goal count, most goals first.
 */
export default function TopScorers() {
  const { t, i18n } = useTranslation()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback((signal?: AbortSignal) => {
    setState({ phase: 'loading' })
    fetchTopScorers(10, signal)
      .then((rows) => {
        if (signal?.aborted) return
        setState({ phase: 'ready', rows })
      })
      .catch((err) => {
        if (signal?.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setState({ phase: 'error' })
      })
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    return () => controller.abort()
  }, [load])

  if (state.phase === 'loading') {
    return (
      <div className="space-y-2">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="h-12 animate-pulse rounded-xl bg-white/[0.05]" />
        ))}
      </div>
    )
  }

  if (state.phase === 'error') return <ErrorState onRetry={() => load()} />

  if (state.rows.length === 0) {
    return (
      <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
        {t('scorers.empty')}
      </p>
    )
  }

  const medals = ['🥇', '🥈', '🥉']
  return (
    <div className="overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] shadow-[0_8px_24px_-16px_rgba(0,0,0,0.8)] backdrop-blur-md">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-hairline text-[0.6rem] uppercase tracking-[0.12em] text-muted/70">
            <th className="py-2.5 pl-3 pr-1 text-left font-semibold sm:pl-4">#</th>
            <th className="py-2.5 pr-2 text-left font-semibold">{t('scorers.player')}</th>
            <th className="px-2 py-2.5 pr-3 text-center font-semibold text-accent/80 sm:pr-4">
              {t('scorers.goals')}
            </th>
          </tr>
        </thead>
        <tbody>
          {state.rows.map((r) => (
            <tr
              key={`${r.rank}-${r.name}`}
              className="border-b border-hairline/60 transition-colors last:border-0 hover:bg-white/[0.03]"
            >
              <td className="py-2.5 pl-3 pr-1 text-center sm:pl-4">
                <span className="text-sm tabular-nums text-muted">
                  {r.rank <= 3 ? medals[r.rank - 1] : r.rank}
                </span>
              </td>
              <td className="min-w-0 py-2.5 pr-2">
                <div className="flex items-center gap-2.5">
                  {r.teamCode && (
                    <Flag code={r.teamCode} flagUrl={undefined} label={r.teamCode} className="h-[1.05rem] w-6" />
                  )}
                  <span className="truncate font-medium text-text">{r.name}</span>
                  {r.teamCode && (
                    <span className="hidden text-xs text-muted/70 sm:inline">
                      {teamName(r.teamCode, r.teamCode, i18n.resolvedLanguage)}
                    </span>
                  )}
                </div>
              </td>
              <td className="px-2 py-2.5 pr-3 text-center text-base font-bold tabular-nums text-accent sm:pr-4">
                {r.goals}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className="border-t border-hairline px-3 py-2 text-[0.62rem] leading-relaxed text-muted/60 sm:px-4">
        {t('scorers.legend')}
      </p>
    </div>
  )
}
