import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Standings } from '../types'
import { fetchStandings } from '../lib/api'
import GroupCard from '../components/GroupCard'
import ThirdPlaceTable from '../components/ThirdPlaceTable'
import StarHero from '../components/StarHero'
import { ErrorState } from '../components/states'

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; standings: Standings }

/** Skeleton grid mirroring the 12 group cards. */
function GroupsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 12 }).map((_, i) => (
        <div
          key={i}
          className="rounded-2xl border border-hairline bg-surface p-4 backdrop-blur-md"
        >
          <div className="mb-3 h-4 w-16 animate-pulse rounded-full bg-white/10" />
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, j) => (
              <div key={j} className="flex items-center gap-2.5">
                <div className="h-3.5 w-5 animate-pulse rounded-[3px] bg-white/10" />
                <div className="h-3 flex-1 animate-pulse rounded-full bg-white/10" />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

export default function Groups() {
  const { t } = useTranslation()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback((signal?: AbortSignal) => {
    setState({ phase: 'loading' })
    fetchStandings(signal)
      .then((standings) => setState({ phase: 'ready', standings }))
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
      <header className="relative mb-6 -mx-4 overflow-hidden rounded-b-3xl px-4 pb-6 pt-4 sm:-mx-6 sm:px-6">
        <StarHero variant="band" />
        <div className="relative">
          <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
            {t('groups.title')}
          </h1>
          <p className="mt-1 text-sm text-muted">{t('groups.subtitle')}</p>
        </div>
      </header>

      {state.phase === 'loading' && <GroupsSkeleton />}
      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}
      {state.phase === 'ready' && (
        <>
          <ThirdPlaceTable rows={state.standings.thirdPlace} />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[...state.standings.groups]
              .sort((a, b) => a.group.localeCompare(b.group))
              .map((g, i) => (
                <GroupCard key={g.group} standing={g} index={i} />
              ))}
          </div>
        </>
      )}
    </div>
  )
}
