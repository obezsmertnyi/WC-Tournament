import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Match } from '../types'
import { fetchMatches } from '../lib/api'
import { groupFixtures, stageLabel } from '../lib/fixtures'
import FixtureSection from '../components/FixtureSection'
import { EmptyState, ErrorState, FixturesSkeleton } from '../components/states'

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
      <header className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
          {t('calendar.title')}
        </h1>
        <p className="mt-1 text-sm text-muted">{t('calendar.subtitle')}</p>
      </header>

      {state.phase === 'loading' && <FixturesSkeleton />}

      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}

      {state.phase === 'ready' && <Fixtures matches={state.matches} />}
    </div>
  )
}

function Fixtures({ matches }: { matches: Match[] }) {
  const { t } = useTranslation()
  if (matches.length === 0) return <EmptyState />

  const { groupStage, knockout } = groupFixtures(matches)

  return (
    <div className="space-y-10">
      {groupStage.map((section) => (
        <FixtureSection
          key={`group-${section.key}`}
          eyebrow={t('calendar.phaseGroup')}
          title={
            section.letter === '—'
              ? t('calendar.group')
              : t('calendar.groupNamed', { letter: section.letter })
          }
          matches={section.matches}
        />
      ))}

      {knockout.map((section) => (
        <FixtureSection
          key={`ko-${section.key}`}
          eyebrow={t('calendar.phaseKnockout')}
          title={stageLabel(section.key)}
          matches={section.matches}
        />
      ))}
    </div>
  )
}
