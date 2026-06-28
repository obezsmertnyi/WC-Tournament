import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { Match, RevealedPrediction } from '../types'
import { fetchMatchReveal } from '../lib/api'
import { hasKickedOff } from '../lib/fixtures'
import Avatar from './Avatar'
import Flag from './Flag'

type State =
  | { phase: 'loading' }
  | { phase: 'locked' }
  | { phase: 'ready'; predictions: RevealedPrediction[] }
  | { phase: 'error' }

interface MatchRevealPanelProps {
  match: Match
}

/**
 * Reveals everyone's predictions for a single match. Before kickoff the API
 * hides them and we render a tasteful lock state; after kickoff we show a
 * compact grid of picks with per-user points.
 */
export default function MatchRevealPanel({ match }: MatchRevealPanelProps) {
  const { t } = useTranslation()
  const [state, setState] = useState<State>({ phase: 'loading' })

  useEffect(() => {
    if (!hasKickedOff(match)) {
      setState({ phase: 'locked' })
      return
    }
    const controller = new AbortController()
    setState({ phase: 'loading' })
    fetchMatchReveal(match.id, controller.signal)
      .then((reveal) => {
        if (controller.signal.aborted) return
        if (Array.isArray(reveal)) setState({ phase: 'ready', predictions: reveal })
        else setState({ phase: 'locked' })
      })
      .catch((err) => {
        if (controller.signal.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setState({ phase: 'error' })
      })
    return () => controller.abort()
  }, [match.id, match.status, match.kickoffAt])

  if (state.phase === 'loading') {
    return (
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-16 animate-pulse rounded-xl bg-white/[0.05]" />
        ))}
      </div>
    )
  }

  if (state.phase === 'locked') {
    return (
      <div className="flex flex-col items-center rounded-xl border border-hairline bg-white/[0.02] px-4 py-8 text-center">
        <LockGlyph />
        <p className="mt-3 text-sm font-medium text-text">{t('competition.revealLocked')}</p>
        <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted">
          {t('competition.revealLockedBody')}
        </p>
      </div>
    )
  }

  if (state.phase === 'error') {
    return (
      <p className="rounded-xl border border-red-500/20 bg-red-500/[0.04] px-4 py-4 text-center text-xs text-muted">
        {t('competition.revealError')}
      </p>
    )
  }

  if (state.predictions.length === 0) {
    return (
      <p className="rounded-xl border border-hairline bg-white/[0.02] px-4 py-6 text-center text-xs text-muted">
        {t('competition.revealEmpty')}
      </p>
    )
  }

  const isKnockout = match.stage !== 'group'

  return (
    <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
      {state.predictions.map((p, i) => {
        const scored = typeof p.points === 'number'
        // For a knockout draw prediction, show who they picked to advance.
        const advTeam =
          isKnockout && p.home === p.away && p.winnerPickTeamId != null
            ? match.home?.id === p.winnerPickTeamId
              ? match.home
              : match.away?.id === p.winnerPickTeamId
                ? match.away
                : null
            : null
        return (
          <motion.div
            key={p.userId}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.25, delay: Math.min(i, 10) * 0.02 }}
            className="flex items-center gap-2 rounded-xl border border-hairline bg-white/[0.03] p-2.5"
          >
            <Avatar src={p.avatarUrl} nickname={p.nickname} className="h-7 w-7 shrink-0 text-xs" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-medium text-text">{p.nickname}</p>
              <p className="tabular-nums text-sm font-bold text-text">
                {p.home}–{p.away}
              </p>
              {advTeam && (
                <p className="mt-0.5 flex items-center gap-1 text-[0.62rem] text-muted/80">
                  <span aria-hidden>→</span>
                  <Flag code={advTeam.code} flagUrl={advTeam.flagUrl} label={advTeam.code} className="h-[0.6rem] w-[0.9rem]" />
                  <span className="font-medium">{advTeam.code}</span>
                </p>
              )}
            </div>
            {scored && (
              <span className="shrink-0 rounded-full border border-accent/40 bg-accent/10 px-1.5 py-0.5 text-[0.6rem] font-semibold tabular-nums text-accent">
                +{p.points}
              </span>
            )}
          </motion.div>
        )
      })}
    </div>
  )
}

function LockGlyph() {
  return (
    <span className="flex h-10 w-10 items-center justify-center rounded-full border border-hairline bg-white/[0.03]">
      <svg viewBox="0 0 24 24" className="h-4 w-4 text-accent" fill="none" aria-hidden="true">
        <rect x="5" y="11" width="14" height="9" rx="2" stroke="currentColor" strokeWidth="1.6" />
        <path
          d="M8 11V8a4 4 0 0 1 8 0v3"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
        />
      </svg>
    </span>
  )
}
