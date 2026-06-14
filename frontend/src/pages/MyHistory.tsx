import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { HistoryMatch, MyHistory } from '../types'
import { fetchMyHistory } from '../lib/api'
import { teamName } from '../lib/teamNames'
import { formatKyivDayMonth, stageLabel } from '../lib/fixtures'
import { useAuth } from '../auth/AuthContext'
import Flag from '../components/Flag'
import { ErrorState } from '../components/states'

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; data: MyHistory }

/** Localized label for a bonus kind. */
function bonusLabel(t: (k: string) => string, kind: string): string {
  switch (kind) {
    case 'champion':
      return t('bonus.champion')
    case 'finalist':
      return t('bonus.finalist')
    case 'top_scorer':
      return t('bonus.topScorer')
    default:
      return kind
  }
}

/**
 * Personal results history: one list of every match the player predicted —
 * their pick, the actual result, points earned and a running cumulative total —
 * plus their tournament bonus picks. The place to track your points dynamics.
 */
export default function MyHistory() {
  const { t, i18n } = useTranslation()
  const { user } = useAuth()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback((signal?: AbortSignal) => {
    setState({ phase: 'loading' })
    fetchMyHistory(signal)
      .then((data) => {
        if (signal?.aborted) return
        setState({ phase: 'ready', data })
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

  const lang = i18n.resolvedLanguage

  return (
    <div className="mx-auto w-full max-w-2xl">
      <header className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">{t('history.title')}</h1>
        <p className="mt-1 text-sm text-muted">
          {user ? t('history.subtitleNamed', { name: user.nickname }) : t('history.subtitle')}
        </p>
      </header>

      {state.phase === 'loading' && (
        <div className="space-y-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-16 animate-pulse rounded-xl bg-white/[0.05]" />
          ))}
        </div>
      )}

      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}

      {state.phase === 'ready' && <HistoryBody data={state.data} lang={lang} />}
    </div>
  )
}

function HistoryBody({ data, lang }: { data: MyHistory; lang: string | undefined }) {
  const { t } = useTranslation()

  if (data.matches.length === 0 && data.bonuses.length === 0) {
    return (
      <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
        {t('history.empty')}
      </p>
    )
  }

  // Running cumulative total over scored matches (chronological order).
  let cumulative = 0

  return (
    <div className="space-y-5">
      {/* Totals summary */}
      <div className="grid grid-cols-3 gap-2">
        <TotalCard label={t('history.matchPoints')} value={data.matchPoints} />
        <TotalCard label={t('history.bonusPoints')} value={data.bonusPoints} accent />
        <TotalCard label={t('history.total')} value={data.total} strong />
      </div>

      {/* Per-match list */}
      <div className="overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] backdrop-blur-md">
        <ul className="divide-y divide-hairline/60">
          {data.matches.map((m) => {
            if (m.scored) cumulative += m.points
            return (
              <HistoryRow key={m.matchId} m={m} lang={lang} cumulative={m.scored ? cumulative : null} />
            )
          })}
        </ul>
      </div>

      {/* Bonus picks */}
      {data.bonuses.length > 0 && (
        <div>
          <h2 className="mb-2 text-[0.65rem] font-semibold uppercase tracking-[0.18em] text-muted/80">
            {t('history.bonusesTitle')}
          </h2>
          <div className="overflow-hidden rounded-2xl border border-hairline bg-white/[0.02]">
            <ul className="divide-y divide-hairline/60">
              {data.bonuses.map((b) => (
                <li key={b.kind} className="flex items-center justify-between gap-3 px-4 py-3 text-sm">
                  <div className="min-w-0">
                    <p className="font-medium text-text">{bonusLabel(t, b.kind)}</p>
                    <p className="flex items-center gap-1.5 truncate text-xs text-muted/80">
                      {b.team && (
                        <Flag
                          code={b.team.code}
                          flagUrl={b.team.flagUrl}
                          label={b.team.name}
                          className="h-[0.7rem] w-[1.05rem]"
                        />
                      )}
                      <span className="truncate">
                        {b.team ? teamName(b.team.code, b.team.name, lang) : b.pickRef}
                      </span>
                    </p>
                  </div>
                  <span className="shrink-0 text-right">
                    {b.awarded ? (
                      <span className="font-bold tabular-nums text-accent">+{b.tierPoints ?? 0}</span>
                    ) : (
                      <span className="text-[0.65rem] uppercase tracking-[0.12em] text-muted/70">
                        {b.tierPoints != null
                          ? t('history.bonusPending', { points: b.tierPoints })
                          : t('history.bonusPendingNoPts')}
                      </span>
                    )}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </div>
  )
}

function TotalCard({
  label,
  value,
  accent,
  strong,
}: {
  label: string
  value: number
  accent?: boolean
  strong?: boolean
}) {
  return (
    <div className="rounded-xl border border-hairline bg-white/[0.03] px-3 py-2.5 text-center">
      <p className="text-[0.55rem] font-semibold uppercase tracking-[0.12em] text-muted/70">{label}</p>
      <p
        className={`mt-0.5 text-xl font-bold tabular-nums ${
          strong ? 'text-accent' : accent ? 'text-accent/90' : 'text-text'
        }`}
      >
        {value}
      </p>
    </div>
  )
}

function HistoryRow({
  m,
  lang,
  cumulative,
}: {
  m: HistoryMatch
  lang: string | undefined
  cumulative: number | null
}) {
  const { t } = useTranslation()
  const home = teamName(m.home.code, m.home.name, lang) || t('fixture.tbd')
  const away = teamName(m.away.code, m.away.name, lang) || t('fixture.tbd')
  const badge = m.stage === 'group' ? m.group || stageLabel(m.stage) : stageLabel(m.stage)

  return (
    <li className="px-3 py-2.5 sm:px-4">
      <div className="mb-1 flex items-center gap-2 text-[0.6rem] uppercase tracking-[0.1em] text-muted/60">
        {m.kickoffAt && <span className="tabular-nums">{formatKyivDayMonth(m.kickoffAt)}</span>}
        {badge && <span className="truncate">· {badge}</span>}
      </div>
      <div className="flex items-center gap-3">
        {/* Teams + prediction vs result */}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <Flag code={m.home.code} flagUrl={m.home.flagUrl} label={home} className="h-[0.85rem] w-[1.25rem]" />
            <span className="min-w-0 flex-1 truncate text-sm text-text">{home}</span>
            <span className="tabular-nums text-sm font-semibold text-muted">
              {m.scored && m.homeScore !== null ? m.homeScore : '–'}
            </span>
          </div>
          <div className="mt-1 flex items-center gap-2">
            <Flag code={m.away.code} flagUrl={m.away.flagUrl} label={away} className="h-[0.85rem] w-[1.25rem]" />
            <span className="min-w-0 flex-1 truncate text-sm text-text">{away}</span>
            <span className="tabular-nums text-sm font-semibold text-muted">
              {m.scored && m.awayScore !== null ? m.awayScore : '–'}
            </span>
          </div>
        </div>

        {/* My pick + points */}
        <div className="shrink-0 border-l border-hairline pl-3 text-right">
          <p className="text-[0.55rem] uppercase tracking-[0.1em] text-muted/60">{t('history.yourPick')}</p>
          <p className="tabular-nums text-sm font-semibold text-text/90">
            {m.predHome}–{m.predAway}
          </p>
          {m.scored ? (
            <p className="mt-0.5">
              <span
                className={`rounded-full px-1.5 py-0.5 text-[0.6rem] font-bold tabular-nums ${
                  m.points > 0 ? 'bg-accent/15 text-accent' : 'bg-white/[0.05] text-muted/70'
                }`}
              >
                +{m.points}
              </span>
              {cumulative !== null && (
                <span className="ml-1.5 text-[0.6rem] tabular-nums text-muted/60">
                  {t('history.runningTotal', { total: cumulative })}
                </span>
              )}
            </p>
          ) : (
            <p className="mt-0.5 text-[0.55rem] uppercase tracking-[0.1em] text-muted/50">
              {t('history.pending')}
            </p>
          )}
        </div>
      </div>
    </li>
  )
}
