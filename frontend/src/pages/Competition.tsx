import { useCallback, useEffect, useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { LeaderboardEntry, Match } from '../types'
import { fetchLeaderboard, fetchMatches } from '../lib/api'
import { formatKyivTime, formatKyivDayMonth, hasKickedOff, stageLabel } from '../lib/fixtures'
import { teamName } from '../lib/teamNames'
import Leaderboard from '../components/Leaderboard'
import MatchRevealPanel from '../components/MatchRevealPanel'
import Flag from '../components/Flag'
import StarHero from '../components/StarHero'
import { ErrorState } from '../components/states'

const POLL_MS = 30_000

type Sub = 'leaderboard' | 'reveals'

export default function Competition() {
  const { t } = useTranslation()
  const [sub, setSub] = useState<Sub>('leaderboard')

  return (
    <div className="mx-auto w-full max-w-3xl">
      <header className="relative mb-6 -mx-4 overflow-hidden rounded-b-3xl px-4 pb-6 pt-4 sm:-mx-6 sm:px-6">
        <StarHero variant="band" />
        <div className="relative">
          <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
            {t('competition.title')}
          </h1>
          <p className="mt-1 text-sm text-muted">{t('competition.subtitle')}</p>
        </div>
      </header>

      <div className="mb-5 inline-flex rounded-full border border-hairline bg-surface p-0.5 backdrop-blur-md">
        {(['leaderboard', 'reveals'] as const).map((key) => {
          const active = sub === key
          return (
            <button
              key={key}
              type="button"
              onClick={() => setSub(key)}
              className={`relative rounded-full px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.12em] transition-colors ${
                active ? 'text-bg' : 'text-muted hover:text-text'
              }`}
            >
              {active && (
                <motion.span
                  layoutId="competition-subtab"
                  className="absolute inset-0 rounded-full bg-accent shadow-[0_0_10px_1px_rgba(201,162,75,0.5)]"
                  transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                />
              )}
              <span className="relative">{t(`competition.tab.${key}`)}</span>
            </button>
          )
        })}
      </div>

      {sub === 'leaderboard' ? <LeaderboardSection /> : <RevealsSection />}
    </div>
  )
}

function LeaderboardSection() {
  const [entries, setEntries] = useState<LeaderboardEntry[] | null>(null)
  const [error, setError] = useState(false)

  const load = useCallback((signal?: AbortSignal) => {
    fetchLeaderboard(signal)
      .then((rows) => {
        if (signal?.aborted) return
        setEntries(rows)
        setError(false)
      })
      .catch((err) => {
        if (signal?.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError(true)
      })
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    // Live-ish polling.
    const id = setInterval(() => load(), POLL_MS)
    return () => {
      controller.abort()
      clearInterval(id)
    }
  }, [load])

  if (error && entries === null) return <ErrorState onRetry={() => load()} />
  if (entries === null) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="h-12 animate-pulse rounded-xl bg-white/[0.05]" />
        ))}
      </div>
    )
  }
  return <Leaderboard entries={entries} />
}

function RevealsSection() {
  const { t, i18n } = useTranslation()
  const [matches, setMatches] = useState<Match[] | null>(null)
  const [error, setError] = useState(false)
  const [openId, setOpenId] = useState<number | null>(null)

  const load = useCallback((signal?: AbortSignal) => {
    fetchMatches(signal)
      .then((m) => {
        if (signal?.aborted) return
        setMatches(m)
        setError(false)
      })
      .catch((err) => {
        if (signal?.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError(true)
      })
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    return () => controller.abort()
  }, [load])

  // Most-recent first: kicked-off matches lead, scheduled ones follow.
  const ordered = useMemo(() => {
    if (!matches) return []
    return [...matches]
      .filter((m) => m.home && m.away)
      .sort((a, b) => new Date(b.kickoffAt).getTime() - new Date(a.kickoffAt).getTime())
  }, [matches])

  if (error && matches === null) return <ErrorState onRetry={() => load()} />
  if (matches === null) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="h-14 animate-pulse rounded-xl bg-white/[0.05]" />
        ))}
      </div>
    )
  }

  if (ordered.length === 0) {
    return (
      <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
        {t('competition.revealNoMatches')}
      </p>
    )
  }

  return (
    <div className="space-y-2">
      {ordered.map((m) => {
        const open = openId === m.id
        const kicked = hasKickedOff(m)
        const homeName = teamName(m.home!.code, m.home!.name, i18n.resolvedLanguage)
        const awayName = teamName(m.away!.code, m.away!.name, i18n.resolvedLanguage)
        const badge = m.stage === 'group' && m.group ? t('calendar.groupNamed', { letter: m.group }) : stageLabel(m.stage)
        return (
          <div
            key={m.id}
            className="overflow-hidden rounded-xl border border-hairline bg-gradient-to-b from-white/[0.045] to-white/[0.01] backdrop-blur-md"
          >
            <button
              type="button"
              onClick={() => setOpenId(open ? null : m.id)}
              aria-expanded={open}
              className="flex w-full items-center gap-3 px-3.5 py-3 text-left transition-colors hover:bg-white/[0.03]"
            >
              <div className="flex min-w-0 flex-1 items-center gap-2">
                <Flag code={m.home!.code} flagUrl={m.home!.flagUrl} label={homeName} className="h-[0.95rem] w-5" />
                <span className="truncate text-sm font-medium text-text">{homeName}</span>
                <span className="shrink-0 text-xs text-muted/60">
                  {m.status === 'finished' && m.homeScore !== null && m.awayScore !== null
                    ? `${m.homeScore}–${m.awayScore}`
                    : t('competition.vs')}
                </span>
                <span className="truncate text-sm font-medium text-text">{awayName}</span>
                <Flag code={m.away!.code} flagUrl={m.away!.flagUrl} label={awayName} className="h-[0.95rem] w-5" />
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <span className="hidden text-[0.6rem] uppercase tracking-[0.12em] text-muted/60 sm:inline">
                  {badge}
                </span>
                <span className="text-[0.65rem] tabular-nums text-muted/70">
                  {kicked ? formatKyivDayMonth(m.kickoffAt) : formatKyivTime(m.kickoffAt)}
                </span>
                <Chevron open={open} />
              </div>
            </button>
            {open && (
              <div className="border-t border-hairline px-3.5 py-3.5">
                <MatchRevealPanel match={m} />
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}

function Chevron({ open }: { open: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={`h-4 w-4 text-muted transition-transform ${open ? 'rotate-180' : ''}`}
      fill="none"
      aria-hidden="true"
    >
      <path d="M6 9l6 6 6-6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}
