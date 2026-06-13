import { useCallback, useEffect, useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { BonusPick, TeamOption } from '../types'
import { ApiError, fetchMyBonus, fetchTeams, saveChampionBonus } from '../lib/api'
import { teamName } from '../lib/teamNames'
import { useAuth } from '../auth/AuthContext'
import { formatKyivDayMonth, formatKyivTime } from '../lib/fixtures'
import Flag from './Flag'
import { ErrorState } from './states'

/** Pull the champion pick (if any) out of the bonus-picks list. */
function championPick(picks: BonusPick[]): BonusPick | null {
  return picks.find((p) => p.kind === 'champion') ?? null
}

/**
 * Champion bonus picker. Lets the signed-in user choose the team they think will
 * win the tournament; shows the current pick, its locked tier points and lock
 * time, and explains that picking during the group stage is worth more than
 * after it. Hard-locks (read-only) once the knockout stage starts (API 409).
 */
export default function BonusChampion() {
  const { t, i18n } = useTranslation()
  const { status } = useAuth()

  const [teams, setTeams] = useState<TeamOption[] | null>(null)
  const [pick, setPick] = useState<BonusPick | null>(null)
  const [loadError, setLoadError] = useState(false)
  const [query, setQuery] = useState('')
  const [busyId, setBusyId] = useState<number | null>(null)
  const [locked, setLocked] = useState(false)
  const [saveError, setSaveError] = useState(false)

  const load = useCallback((signal?: AbortSignal) => {
    setLoadError(false)
    Promise.all([fetchTeams(signal), fetchMyBonus(signal)])
      .then(([teamRows, picks]) => {
        if (signal?.aborted) return
        setTeams(teamRows)
        setPick(championPick(picks))
      })
      .catch((err) => {
        if (signal?.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setLoadError(true)
      })
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    return () => controller.abort()
  }, [load])

  const currentTeamId = pick ? Number(pick.pickRef) : null

  const sortedTeams = useMemo(() => {
    if (!teams) return []
    const lang = i18n.resolvedLanguage
    const collator = lang === 'uk' ? 'uk' : 'en'
    return [...teams].sort((a, b) =>
      teamName(a.code, a.name, lang).localeCompare(teamName(b.code, b.name, lang), collator),
    )
  }, [teams, i18n.resolvedLanguage])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return sortedTeams
    return sortedTeams.filter((team) => {
      const name = teamName(team.code, team.name, i18n.resolvedLanguage).toLowerCase()
      return name.includes(q) || team.code.toLowerCase().includes(q)
    })
  }, [sortedTeams, query, i18n.resolvedLanguage])

  const choose = useCallback(
    async (team: TeamOption) => {
      if (busyId !== null || locked || team.id === currentTeamId) return
      setBusyId(team.id)
      setSaveError(false)
      try {
        const saved = await saveChampionBonus(team.id)
        setPick(saved)
      } catch (err) {
        if (err instanceof ApiError && err.status === 409) {
          setLocked(true)
        } else {
          setSaveError(true)
        }
      } finally {
        setBusyId(null)
      }
    },
    [busyId, locked, currentTeamId],
  )

  if (status !== 'authenticated') {
    return (
      <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
        {t('bonus.signInPrompt')}
      </p>
    )
  }

  if (loadError && teams === null) return <ErrorState onRetry={() => load()} />

  if (teams === null) {
    return (
      <div className="space-y-3">
        <div className="h-24 animate-pulse rounded-2xl bg-white/[0.05]" />
        <div className="h-11 animate-pulse rounded-xl bg-white/[0.05]" />
        <div className="grid grid-cols-3 gap-1.5 sm:grid-cols-4">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="h-16 animate-pulse rounded-lg bg-white/[0.05]" />
          ))}
        </div>
      </div>
    )
  }

  const currentTeam = currentTeamId !== null ? teams.find((tm) => tm.id === currentTeamId) : undefined
  const currentName = currentTeam
    ? teamName(currentTeam.code, currentTeam.name, i18n.resolvedLanguage)
    : null

  return (
    <div className="space-y-4">
      {/* Current pick + tier summary */}
      <div className="rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.06] to-white/[0.015] p-5 backdrop-blur-md">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-muted/80">
            {t('bonus.champion')}
          </h2>
          {locked && (
            <span className="shrink-0 rounded-full bg-white/[0.06] px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.12em] text-muted">
              {t('bonus.locked')}
            </span>
          )}
        </div>

        {currentTeam && currentName ? (
          <div className="mt-3 flex items-center gap-3">
            <Flag
              code={currentTeam.code}
              flagUrl={currentTeam.flagUrl}
              label={currentName}
              className="h-7 w-10"
            />
            <div className="min-w-0 flex-1">
              <p className="truncate text-lg font-bold text-text">{currentName}</p>
              <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs text-muted/80">
                {pick?.tierPoints != null && (
                  <span className="font-semibold text-accent">
                    {t('bonus.points', { points: pick.tierPoints })}
                  </span>
                )}
                {pick?.lockedAt && (
                  <span className="tabular-nums">
                    {t('bonus.lockedAt', {
                      date: formatKyivDayMonth(pick.lockedAt),
                      time: formatKyivTime(pick.lockedAt),
                    })}
                  </span>
                )}
              </div>
            </div>
          </div>
        ) : (
          <p className="mt-2 text-sm text-muted">{t('bonus.noPick')}</p>
        )}

        <p className="mt-3 text-xs leading-relaxed text-muted/70">{t('bonus.tierHint')}</p>

        {saveError && (
          <p className="mt-3 text-xs font-semibold uppercase tracking-[0.14em] text-red-400/90">
            {t('bonus.saveError')}
          </p>
        )}
      </div>

      {/* Picker — searchable list of the 48 teams */}
      {!locked && (
        <>
          <label className="block">
            <span className="sr-only">{t('bonus.searchPlaceholder')}</span>
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t('bonus.searchPlaceholder')}
              className="h-11 w-full rounded-xl border border-hairline bg-white/[0.04] px-3.5 text-sm text-text outline-none transition-colors placeholder:text-muted/50 focus:border-accent focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)]"
            />
          </label>

          {filtered.length === 0 ? (
            <p className="rounded-xl border border-hairline bg-white/[0.02] px-4 py-6 text-center text-sm text-muted">
              {t('bonus.noResults')}
            </p>
          ) : (
            <div className="grid max-h-[26rem] grid-cols-3 gap-1.5 overflow-y-auto rounded-2xl border border-hairline bg-white/[0.02] p-2 sm:grid-cols-4">
              {filtered.map((team) => {
                const active = team.id === currentTeamId
                const name = teamName(team.code, team.name, i18n.resolvedLanguage)
                const busy = busyId === team.id
                return (
                  <button
                    key={team.id}
                    type="button"
                    onClick={() => void choose(team)}
                    disabled={busyId !== null}
                    aria-pressed={active}
                    title={name}
                    className={`relative flex flex-col items-center gap-1.5 rounded-lg border px-1.5 py-2.5 transition-colors disabled:cursor-not-allowed ${
                      active
                        ? 'border-accent/60 bg-accent/10'
                        : 'border-transparent hover:bg-white/[0.05]'
                    } ${busyId !== null && !busy ? 'opacity-40' : ''}`}
                  >
                    {active && (
                      <motion.span
                        layoutId="bonus-champion-active"
                        className="absolute inset-0 -z-10 rounded-lg ring-1 ring-accent/40"
                        transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                      />
                    )}
                    <Flag code={team.code} flagUrl={team.flagUrl} label={name} className="h-[1.1rem] w-6" />
                    <span
                      className={`line-clamp-2 text-center text-[0.62rem] font-medium leading-tight ${
                        active ? 'text-accent' : 'text-muted/80'
                      }`}
                    >
                      {busy ? t('bonus.saving') : name}
                    </span>
                  </button>
                )
              })}
            </div>
          )}
        </>
      )}
    </div>
  )
}
