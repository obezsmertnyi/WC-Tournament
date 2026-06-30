import { useCallback, useEffect, useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { BonusPick, TeamOption } from '../types'
import {
  ApiError,
  fetchMyBonus,
  fetchTeams,
  saveChampionBonus,
  saveFinalistBonus,
  saveTopScorerBonus,
} from '../lib/api'
import { teamName } from '../lib/teamNames'
import { useAuth } from '../auth/AuthContext'
import { canParticipate } from '../lib/access'
import DemoLocked from './DemoLocked'
import { TOP_SCORERS } from '../lib/topScorers'
import { formatKyivDayMonth, formatKyivTime } from '../lib/fixtures'
import Flag from './Flag'
import { ErrorState } from './states'

function pickOf(picks: BonusPick[], kind: string): BonusPick | null {
  return picks.find((p) => p.kind === kind) ?? null
}

/**
 * Tournament-wide bonus picks: champion, losing finalist (both teams), and the
 * top scorer (free-text player). Each pick is time-tiered — choosing during the
 * group stage is worth more — and all lock at the Round of 16. Awarding is gated
 * on the pick proving correct, so the points shown are the *potential* award.
 */
export default function BonusPanel() {
  const { t } = useTranslation()
  const { status, user } = useAuth()

  const [teams, setTeams] = useState<TeamOption[] | null>(null)
  const [picks, setPicks] = useState<Record<string, BonusPick | null>>({})
  const [loadError, setLoadError] = useState(false)

  const load = useCallback((signal?: AbortSignal) => {
    setLoadError(false)
    Promise.all([fetchTeams(signal), fetchMyBonus(signal)])
      .then(([teamRows, bonusPicks]) => {
        if (signal?.aborted) return
        setTeams(teamRows)
        setPicks({
          champion: pickOf(bonusPicks, 'champion'),
          finalist: pickOf(bonusPicks, 'finalist'),
          top_scorer: pickOf(bonusPicks, 'top_scorer'),
        })
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

  const onSaved = useCallback((kind: string, pick: BonusPick) => {
    setPicks((prev) => ({ ...prev, [kind]: pick }))
  }, [])

  if (status !== 'authenticated') {
    return (
      <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
        {t('bonus.signInPrompt')}
      </p>
    )
  }

  // Restricted demo tier (none/ro): bonus picks are participation — locked.
  if (!canParticipate(user)) return <DemoLocked reason="participate" />

  if (loadError && teams === null) return <ErrorState onRetry={() => load()} />

  if (teams === null) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-28 animate-pulse rounded-2xl bg-white/[0.05]" />
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <TeamPickCard
        kind="champion"
        title={t('bonus.champion')}
        tierValue="10 / 6"
        teams={teams}
        pick={picks.champion ?? null}
        save={saveChampionBonus}
        onSaved={onSaved}
      />
      <TeamPickCard
        kind="finalist"
        title={t('bonus.finalist')}
        tierValue="5 / 3"
        teams={teams}
        pick={picks.finalist ?? null}
        save={saveFinalistBonus}
        onSaved={onSaved}
      />
      <TopScorerCard
        title={t('bonus.topScorer')}
        tierValue="5 / 2"
        pick={picks.top_scorer ?? null}
        onSaved={onSaved}
      />
    </div>
  )
}

// ── Tier chip shared by all cards ────────────────────────────────────────────
function TierChip({ value }: { value: string }) {
  return (
    <span className="shrink-0 rounded-full border border-accent/40 bg-accent/10 px-2 py-0.5 text-[0.6rem] font-bold tabular-nums tracking-wide text-accent">
      {value}
    </span>
  )
}

// ── Team-referenced pick (champion / finalist) ───────────────────────────────
interface TeamPickCardProps {
  kind: 'champion' | 'finalist'
  title: string
  tierValue: string
  teams: TeamOption[]
  pick: BonusPick | null
  save: (teamId: number) => Promise<BonusPick>
  onSaved: (kind: string, pick: BonusPick) => void
}

function TeamPickCard({ kind, title, tierValue, teams, pick, save, onSaved }: TeamPickCardProps) {
  const { t, i18n } = useTranslation()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [busyId, setBusyId] = useState<number | null>(null)
  const [locked, setLocked] = useState(false)
  const [saveError, setSaveError] = useState(false)

  const currentTeamId = pick ? Number(pick.pickRef) : null
  const currentTeam = currentTeamId !== null ? teams.find((tm) => tm.id === currentTeamId) : undefined
  const currentName = currentTeam
    ? teamName(currentTeam.code, currentTeam.name, i18n.resolvedLanguage)
    : null

  const sorted = useMemo(() => {
    const lang = i18n.resolvedLanguage
    const collator = lang === 'uk' ? 'uk' : 'en'
    return [...teams].sort((a, b) =>
      teamName(a.code, a.name, lang).localeCompare(teamName(b.code, b.name, lang), collator),
    )
  }, [teams, i18n.resolvedLanguage])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return sorted
    return sorted.filter((team) => {
      const name = teamName(team.code, team.name, i18n.resolvedLanguage).toLowerCase()
      return name.includes(q) || team.code.toLowerCase().includes(q)
    })
  }, [sorted, query, i18n.resolvedLanguage])

  const choose = useCallback(
    async (team: TeamOption) => {
      if (busyId !== null || locked || team.id === currentTeamId) return
      setBusyId(team.id)
      setSaveError(false)
      try {
        const saved = await save(team.id)
        onSaved(kind, saved)
        setOpen(false)
      } catch (err) {
        if (err instanceof ApiError && err.status === 409) setLocked(true)
        else setSaveError(true)
      } finally {
        setBusyId(null)
      }
    },
    [busyId, locked, currentTeamId, save, onSaved, kind],
  )

  return (
    <div className="rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.06] to-white/[0.015] p-4 backdrop-blur-md">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-semibold uppercase tracking-[0.14em] text-muted/80">{title}</h3>
          <TierChip value={tierValue} />
        </div>
        {locked ? (
          <span className="shrink-0 rounded-full bg-white/[0.06] px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.12em] text-muted">
            {t('bonus.locked')}
          </span>
        ) : (
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            className="shrink-0 rounded-full border border-hairline px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.12em] text-muted/80 transition-colors hover:border-accent/40 hover:text-accent"
          >
            {open ? t('bonus.close') : currentTeam ? t('bonus.change') : t('bonus.pick')}
          </button>
        )}
      </div>

      {currentTeam && currentName ? (
        <div className="mt-3 flex items-center gap-3">
          <Flag code={currentTeam.code} flagUrl={currentTeam.flagUrl} label={currentName} className="h-6 w-9" />
          <div className="min-w-0 flex-1">
            <p className="truncate text-base font-bold text-text">{currentName}</p>
            <div className="mt-0.5 flex flex-wrap items-center gap-x-3 text-xs text-muted/80">
              {pick?.tierPoints != null && (
                <span className="font-semibold text-accent">{t('bonus.points', { points: pick.tierPoints })}</span>
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

      {saveError && (
        <p className="mt-2 text-xs font-semibold uppercase tracking-[0.14em] text-red-400/90">{t('bonus.saveError')}</p>
      )}

      {open && !locked && (
        <div className="mt-3 space-y-2 border-t border-hairline pt-3">
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('bonus.searchPlaceholder')}
            className="h-10 w-full rounded-xl border border-hairline bg-white/[0.04] px-3 text-sm text-text outline-none transition-colors placeholder:text-muted/50 focus:border-accent"
          />
          {filtered.length === 0 ? (
            <p className="px-4 py-5 text-center text-sm text-muted">{t('bonus.noResults')}</p>
          ) : (
            <div className="grid max-h-72 grid-cols-3 gap-1.5 overflow-y-auto rounded-xl border border-hairline bg-white/[0.02] p-2 sm:grid-cols-4">
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
                      active ? 'border-accent/60 bg-accent/10' : 'border-transparent hover:bg-white/[0.05]'
                    } ${busyId !== null && !busy ? 'opacity-40' : ''}`}
                  >
                    {active && (
                      <motion.span
                        layoutId={`bonus-${kind}-active`}
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
        </div>
      )}
    </div>
  )
}

// ── Top-scorer pick (free-text player, with star suggestions) ────────────────
interface TopScorerCardProps {
  title: string
  tierValue: string
  pick: BonusPick | null
  onSaved: (kind: string, pick: BonusPick) => void
}

function TopScorerCard({ title, tierValue, pick, onSaved }: TopScorerCardProps) {
  const { t } = useTranslation()
  const [value, setValue] = useState(pick?.pickRef ?? '')
  const [busy, setBusy] = useState(false)
  const [locked, setLocked] = useState(false)
  const [saveError, setSaveError] = useState(false)

  // Keep the input in sync if the pick loads/changes from the server.
  useEffect(() => {
    setValue(pick?.pickRef ?? '')
  }, [pick?.pickRef])

  const current = pick?.pickRef ?? ''
  const dirty = value.trim() !== '' && value.trim() !== current

  const commit = useCallback(
    async (name: string) => {
      const trimmed = name.trim()
      if (!trimmed || busy || locked) return
      setBusy(true)
      setSaveError(false)
      try {
        const saved = await saveTopScorerBonus(trimmed)
        onSaved('top_scorer', saved)
        setValue(trimmed)
      } catch (err) {
        if (err instanceof ApiError && err.status === 409) setLocked(true)
        else setSaveError(true)
      } finally {
        setBusy(false)
      }
    },
    [busy, locked, onSaved],
  )

  const submit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      void commit(value)
    },
    [commit, value],
  )

  return (
    <div className="rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.06] to-white/[0.015] p-4 backdrop-blur-md">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-semibold uppercase tracking-[0.14em] text-muted/80">{title}</h3>
          <TierChip value={tierValue} />
        </div>
        {locked && (
          <span className="shrink-0 rounded-full bg-white/[0.06] px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.12em] text-muted">
            {t('bonus.locked')}
          </span>
        )}
      </div>

      {pick?.pickRef ? (
        <div className="mt-2 flex flex-wrap items-center gap-x-3 text-xs text-muted/80">
          <span className="text-base font-bold text-text">⚽ {pick.pickRef}</span>
          {pick.tierPoints != null && (
            <span className="font-semibold text-accent">{t('bonus.points', { points: pick.tierPoints })}</span>
          )}
          {pick.lockedAt && (
            <span className="tabular-nums">
              {t('bonus.lockedAt', {
                date: formatKyivDayMonth(pick.lockedAt),
                time: formatKyivTime(pick.lockedAt),
              })}
            </span>
          )}
        </div>
      ) : (
        <p className="mt-2 text-sm text-muted">{t('bonus.noPick')}</p>
      )}

      {!locked && (
        <div className="mt-3 border-t border-hairline pt-3">
          {/* Quick-pick chips — top Golden Boot contenders; tap to select & save */}
          <div className="flex flex-wrap gap-1.5">
            {TOP_SCORERS.map((s) => {
              const active = current === s.name
              return (
                <button
                  key={s.name}
                  type="button"
                  onClick={() => void commit(s.name)}
                  disabled={busy}
                  aria-pressed={active}
                  className={`flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors disabled:opacity-40 ${
                    active
                      ? 'border-accent/60 bg-accent/10 text-accent'
                      : 'border-hairline bg-white/[0.03] text-muted/85 hover:border-white/20 hover:text-text'
                  }`}
                >
                  <Flag code={s.teamCode} flagUrl={undefined} label={s.name} className="h-[0.8rem] w-[1.15rem]" />
                  {s.name}
                </button>
              )
            })}
          </div>

          {/* Free-text entry for anyone not in the shortlist */}
          <form onSubmit={submit} className="mt-2.5 flex gap-2">
            <input
              type="text"
              list="topscorer-suggestions"
              value={value}
              maxLength={80}
              onChange={(e) => setValue(e.target.value)}
              placeholder={t('bonus.scorerPlaceholder')}
              className="h-10 min-w-0 flex-1 rounded-xl border border-hairline bg-white/[0.04] px-3 text-sm text-text outline-none transition-colors placeholder:text-muted/50 focus:border-accent"
            />
            <datalist id="topscorer-suggestions">
              {TOP_SCORERS.map((s) => (
                <option key={s.name} value={s.name} />
              ))}
            </datalist>
            <button
              type="submit"
              disabled={!dirty || busy}
              className="h-10 shrink-0 rounded-xl bg-accent px-4 text-sm font-semibold uppercase tracking-[0.12em] text-bg transition-opacity hover:opacity-90 disabled:opacity-40"
            >
              {busy ? t('bonus.saving') : t('bonus.save')}
            </button>
          </form>
        </div>
      )}

      {saveError && (
        <p className="mt-2 text-xs font-semibold uppercase tracking-[0.14em] text-red-400/90">{t('bonus.saveError')}</p>
      )}
    </div>
  )
}
