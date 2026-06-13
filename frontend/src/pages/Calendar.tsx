import { useCallback, useEffect, useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { Match } from '../types'
import { fetchMatches } from '../lib/api'
import { buildMatchDays, defaultMatchDayKey, formatKyivFullDate } from '../lib/fixtures'
import FixtureCard from '../components/FixtureCard'
import DateStrip from '../components/DateStrip'
import StarHero from '../components/StarHero'
import Trophy from '../components/Trophy'
import { EmptyState, ErrorState, FixturesSkeleton } from '../components/states'
import { useMountAnimation } from '../lib/motion'

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; matches: Match[] }

export default function Calendar() {
  const { t } = useTranslation()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback((signal?: AbortSignal) => {
    setState({ phase: 'loading' })
    fetchMatches(signal)
      .then((matches) => setState({ phase: 'ready', matches }))
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

  return (
    <div className="mx-auto w-full max-w-5xl">
      {/* Hero: large blended player artwork behind the trophy mark + title. */}
      <header className="relative mb-6 -mx-4 overflow-hidden rounded-b-3xl px-4 pb-6 pt-4 sm:-mx-6 sm:px-6">
        <StarHero variant="band" />
        <div className="relative flex items-center gap-3 sm:gap-4">
          <Trophy className="h-12 w-12 rounded-xl sm:h-16 sm:w-16" />
          <div className="min-w-0">
            <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
              {t('calendar.title')}
            </h1>
            <p className="mt-1 text-sm text-muted">{t('calendar.subtitle')}</p>
          </div>
        </div>
      </header>

      {state.phase === 'loading' && <FixturesSkeleton />}
      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}
      {state.phase === 'ready' && <DayView matches={state.matches} />}
    </div>
  )
}

function DayView({ matches }: { matches: Match[] }) {
  const { t } = useTranslation()
  const days = useMemo(() => buildMatchDays(matches), [matches])
  const [selected, setSelected] = useState<string | undefined>(undefined)
  const mount = useMountAnimation(8)

  // Initialize / reconcile the selected day whenever the day set changes.
  useEffect(() => {
    setSelected((prev) =>
      prev && days.some((d) => d.key === prev) ? prev : defaultMatchDayKey(days),
    )
  }, [days])

  if (days.length === 0 || !selected) return <EmptyState />

  const current = days.find((d) => d.key === selected) ?? days[0]

  return (
    <div>
      <DateStrip days={days} selected={current.key} onSelect={setSelected} />

      <motion.div
        key={current.key}
        initial={mount.initial}
        animate={mount.animate}
        transition={mount.transition}
      >
        <div className="mb-4 flex items-baseline justify-between gap-3">
          <h2 className="text-base font-semibold capitalize tracking-tight text-text">
            {formatKyivFullDate(current.iso)}
          </h2>
          <span className="shrink-0 text-xs font-medium uppercase tracking-[0.14em] text-muted/70">
            {t('calendar.matchCount', { count: current.matches.length })}
          </span>
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {current.matches.map((m, i) => (
            <FixtureCard key={m.id} match={m} index={i} />
          ))}
        </div>
      </motion.div>
    </div>
  )
}
