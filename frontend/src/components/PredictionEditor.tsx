import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Match } from '../types'
import { hasKickedOff } from '../lib/fixtures'
import { teamName } from '../lib/teamNames'
import { usePredictions, type SaveState } from '../predictions/PredictionsContext'

/** Clamp a raw input value into the 0–30 score range, or null when empty. */
function clampScore(raw: string): number | null {
  if (raw.trim() === '') return null
  const n = Math.floor(Number(raw))
  if (Number.isNaN(n)) return null
  return Math.max(0, Math.min(30, n))
}

function SaveBadge({ state }: { state: SaveState }) {
  const { t } = useTranslation()
  if (state === 'idle') return null

  const map: Record<Exclude<SaveState, 'idle'>, { label: string; cls: string }> = {
    saving: { label: t('predict.saving'), cls: 'text-muted' },
    saved: { label: t('predict.saved'), cls: 'text-accent' },
    locked: { label: t('predict.locked'), cls: 'text-red-400/80' },
    error: { label: t('predict.error'), cls: 'text-red-400/80' },
  }
  const { label, cls } = map[state]
  return (
    <span className={`text-[0.6rem] font-semibold uppercase tracking-[0.14em] ${cls}`}>
      {state === 'saved' && '✓ '}
      {label}
    </span>
  )
}

interface ScoreInputProps {
  value: number | null
  onChange: (v: number | null) => void
  onCommit: () => void
  ariaLabel: string
  disabled: boolean
}

function ScoreInput({ value, onChange, onCommit, ariaLabel, disabled }: ScoreInputProps) {
  return (
    <input
      type="number"
      inputMode="numeric"
      min={0}
      max={30}
      step={1}
      disabled={disabled}
      aria-label={ariaLabel}
      value={value ?? ''}
      onChange={(e) => onChange(clampScore(e.target.value))}
      onBlur={onCommit}
      className="h-9 w-12 rounded-lg border border-hairline bg-white/[0.04] text-center text-base font-bold tabular-nums text-text outline-none transition-colors focus:border-accent focus:bg-accent/[0.06] focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)] disabled:opacity-50"
    />
  )
}

interface PredictionEditorProps {
  match: Match
}

/**
 * Inline prediction entry shown on a FixtureCard for signed-in users when the
 * match hasn't kicked off. Score inputs (0–30) plus, for knockout ties, an
 * advancer pick. Saves are debounced via the predictions context. Once the
 * match kicks off this collapses to a read-only summary with a points badge.
 */
export default function PredictionEditor({ match }: PredictionEditorProps) {
  const { t, i18n } = useTranslation()
  const { byMatch, save, saveStateOf } = usePredictions()
  const existing = byMatch.get(match.id)

  const [home, setHome] = useState<number | null>(existing?.home ?? null)
  const [away, setAway] = useState<number | null>(existing?.away ?? null)
  const [winner, setWinner] = useState<number | null>(existing?.winnerPickTeamId ?? null)

  // Reconcile when predictions arrive/refresh from the server.
  useEffect(() => {
    setHome(existing?.home ?? null)
    setAway(existing?.away ?? null)
    setWinner(existing?.winnerPickTeamId ?? null)
  }, [existing?.home, existing?.away, existing?.winnerPickTeamId])

  const locked = hasKickedOff(match)
  const isKnockout = match.stage !== 'group'
  const saveState = saveStateOf(match.id)

  const homeId = (match.home as { id?: number } | null)?.id ?? null
  const awayId = (match.away as { id?: number } | null)?.id ?? null
  const homeLabel = match.home
    ? teamName(match.home.code, match.home.name, i18n.resolvedLanguage)
    : t('fixture.tbd')
  const awayLabel = match.away
    ? teamName(match.away.code, match.away.name, i18n.resolvedLanguage)
    : t('fixture.tbd')

  function commit(nextHome: number | null, nextAway: number | null, nextWinner: number | null) {
    if (nextHome === null || nextAway === null) return // need both scores to save
    save(match.id, {
      home: nextHome,
      away: nextAway,
      winnerPickTeamId: isKnockout ? nextWinner : null,
    })
  }

  // ── Locked (kicked off): read-only summary of the user's own pick ──────────
  if (locked) {
    if (!existing) {
      return (
        <div className="mt-3 border-t border-hairline pt-2.5">
          <p className="text-[0.65rem] uppercase tracking-[0.14em] text-muted/60">
            {t('predict.noPick')}
          </p>
        </div>
      )
    }
    const scored = existing && typeof (existing as { points?: number }).points === 'number'
    return (
      <div className="mt-3 flex items-center justify-between gap-2 border-t border-hairline pt-2.5">
        <span className="text-[0.6rem] font-semibold uppercase tracking-[0.14em] text-muted/70">
          {t('predict.yourPick')}
        </span>
        <span className="flex items-center gap-2">
          <span className="tabular-nums text-sm font-bold text-text">
            {existing.home}–{existing.away}
          </span>
          {scored && (
            <span className="rounded-full border border-accent/40 bg-accent/10 px-2 py-0.5 text-[0.6rem] font-semibold text-accent">
              ✓ {t('predict.pointsBadge', { points: (existing as { points?: number }).points })}
            </span>
          )}
        </span>
      </div>
    )
  }

  // ── Editable (before kickoff) ──────────────────────────────────────────────
  return (
    <div className="mt-3 border-t border-hairline pt-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[0.6rem] font-semibold uppercase tracking-[0.14em] text-accent/80">
          {t('predict.yourPick')}
        </span>
        <SaveBadge state={saveState} />
      </div>

      <div className="mt-2 flex items-center justify-center gap-3">
        <ScoreInput
          value={home}
          ariaLabel={t('predict.scoreFor', { team: homeLabel })}
          disabled={false}
          onChange={(v) => setHome(v)}
          onCommit={() => commit(home, away, winner)}
        />
        <span className="text-sm font-semibold text-muted/60">:</span>
        <ScoreInput
          value={away}
          ariaLabel={t('predict.scoreFor', { team: awayLabel })}
          disabled={false}
          onChange={(v) => setAway(v)}
          onCommit={() => commit(home, away, winner)}
        />
      </div>

      {isKnockout && (
        <div className="mt-3">
          <p className="mb-1.5 text-center text-[0.6rem] uppercase tracking-[0.14em] text-muted/70">
            {t('predict.advancer')}
          </p>
          <div className="flex gap-2">
            {[
              { id: homeId, label: homeLabel },
              { id: awayId, label: awayLabel },
            ].map(({ id, label }) => {
              const active = id !== null && winner === id
              return (
                <button
                  key={label}
                  type="button"
                  disabled={id === null}
                  onClick={() => {
                    const next = id
                    setWinner(next)
                    commit(home, away, next)
                  }}
                  className={`flex-1 truncate rounded-lg border px-2 py-1.5 text-xs font-medium transition-colors disabled:opacity-40 ${
                    active
                      ? 'border-accent/50 bg-accent/10 text-accent'
                      : 'border-hairline bg-white/[0.03] text-muted hover:text-text'
                  }`}
                >
                  {label}
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
