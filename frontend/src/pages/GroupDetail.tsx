import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { GroupStanding, Match } from '../types'
import { fetchMatches, fetchStandings } from '../lib/api'
import StandingsTable from '../components/StandingsTable'
import FixtureCard from '../components/FixtureCard'
import StarHero from '../components/StarHero'
import { ErrorState } from '../components/states'
import { useMountAnimation } from '../lib/motion'
import { starForTeams } from '../lib/stars'

interface Data {
  standing: GroupStanding | undefined
  matches: Match[]
}

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; data: Data }

function byKickoff(a: Match, b: Match): number {
  const ta = a.kickoffAt ? new Date(a.kickoffAt).getTime() : Infinity
  const tb = b.kickoffAt ? new Date(b.kickoffAt).getTime() : Infinity
  if (ta !== tb) return ta - tb
  return a.matchNumber - b.matchNumber
}

export default function GroupDetail() {
  const { t } = useTranslation()
  const { letter = '' } = useParams()
  const group = letter.toUpperCase()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback(
    (signal?: AbortSignal) => {
      setState({ phase: 'loading' })
      Promise.all([fetchStandings(signal), fetchMatches(signal)])
        .then(([standings, matches]) => {
          const standing = standings.find((g) => g.group.toUpperCase() === group)
          const groupMatches = matches
            .filter((m) => m.stage === 'group' && (m.group ?? '').toUpperCase() === group)
            .sort(byKickoff)
          setState({ phase: 'ready', data: { standing, matches: groupMatches } })
        })
        .catch((err) => {
          if (signal?.aborted) return
          if (err instanceof DOMException && err.name === 'AbortError') return
          setState({ phase: 'error' })
        })
    },
    [group],
  )

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    return () => controller.abort()
  }, [load])

  const title = useMemo(() => t('calendar.groupNamed', { letter: group }), [t, group])
  const mount = useMountAnimation(10)

  // The featured star (if any) whose national team plays in this group. Codes are
  // collected from standings rows and, as a fallback, from the group's fixtures —
  // so the portrait shows even before standings populate.
  const groupStar = useMemo(() => {
    if (state.phase !== 'ready') return undefined
    const codes = new Set<string>()
    state.data.standing?.rows.forEach((r) => r.code && codes.add(r.code))
    state.data.matches.forEach((m) => {
      if (m.home?.code) codes.add(m.home.code)
      if (m.away?.code) codes.add(m.away.code)
    })
    return starForTeams(codes)
  }, [state])

  return (
    <div className="mx-auto w-full max-w-5xl">
      <Link
        to="/groups"
        className="mb-4 inline-flex items-center gap-1.5 text-xs font-medium uppercase tracking-[0.14em] text-muted transition-colors hover:text-accent"
      >
        <svg viewBox="0 0 24 24" className="h-3.5 w-3.5" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden>
          <path d="M15 18l-6-6 6-6" />
        </svg>
        {t('groups.back')}
      </Link>

      <header className="relative mb-6 -mx-4 overflow-hidden rounded-b-3xl px-4 pb-6 pt-2 sm:-mx-6 sm:px-6">
        {groupStar && <StarHero variant="portrait" stars={[groupStar]} />}
        <div className="relative">
          <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">{title}</h1>
          <p className="mt-1 text-sm text-muted">{t('groups.detailSubtitle')}</p>
        </div>
      </header>

      {state.phase === 'loading' && (
        <div className="space-y-8">
          <div className="h-64 animate-pulse rounded-2xl border border-hairline bg-surface" />
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-40 animate-pulse rounded-2xl border border-hairline bg-surface" />
            ))}
          </div>
        </div>
      )}

      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}

      {state.phase === 'ready' && (
        <motion.div
          initial={mount.initial}
          animate={mount.animate}
          transition={mount.transition}
          className="space-y-10"
        >
          <section>
            <h2 className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-muted">
              {t('table.standings')}
            </h2>
            {state.data.standing && state.data.standing.rows.length > 0 ? (
              <StandingsTable rows={state.data.standing.rows} />
            ) : (
              <p className="rounded-2xl border border-hairline bg-surface px-4 py-8 text-center text-sm text-muted backdrop-blur-md">
                {t('groups.empty')}
              </p>
            )}
          </section>

          <section>
            <h2 className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-muted">
              {t('groups.matches')}
            </h2>
            {state.data.matches.length > 0 ? (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {state.data.matches.map((m, i) => (
                  <FixtureCard key={m.id} match={m} index={i} showBadge={false} />
                ))}
              </div>
            ) : (
              <p className="rounded-2xl border border-hairline bg-surface px-4 py-8 text-center text-sm text-muted backdrop-blur-md">
                {t('groups.noMatches')}
              </p>
            )}
          </section>
        </motion.div>
      )}
    </div>
  )
}
