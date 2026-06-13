import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Match } from '../types'
import { teamName } from '../lib/teamNames'
import { setMatchResult } from '../lib/api'

type SaveState = 'idle' | 'saving' | 'saved' | 'error'

/** Clamp a raw input value into the 0–30 score range, or null when empty. */
function clampScore(raw: string): number | null {
  if (raw.trim() === '') return null
  const n = Math.floor(Number(raw))
  if (Number.isNaN(n)) return null
  return Math.max(0, Math.min(30, n))
}

interface ScoreInputProps {
  value: number | null
  onChange: (v: number | null) => void
  ariaLabel: string
  disabled: boolean
}

function ScoreInput({ value, onChange, ariaLabel, disabled }: ScoreInputProps) {
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
      className="h-9 w-12 rounded-lg border border-accent/30 bg-accent/[0.04] text-center text-base font-bold tabular-nums text-text outline-none transition-colors focus:border-accent focus:bg-accent/[0.08] focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)] disabled:opacity-50"
    />
  )
}

interface AdminResultEditorProps {
  match: Match
  /** Called with the new scores after a successful save (optimistic update). */
  onSaved?: (homeScore: number, awayScore: number) => void
}

/**
 * Admin-only mini-panel to set or correct the ACTUAL result of a match. Visually
 * distinct from the prediction editor (gold "Admin · result" header, collapsible)
 * so it's unmistakable that this writes the real score, not a prediction. Works
 * for any match including past/finished ones — the whole point is letting an
 * admin set or fix a result after the fact. Saves with status 'finished'.
 */
export default function AdminResultEditor({ match, onSaved }: AdminResultEditorProps) {
  const { t, i18n } = useTranslation()
  const [open, setOpen] = useState(false)
  const [home, setHome] = useState<number | null>(match.homeScore)
  const [away, setAway] = useState<number | null>(match.awayScore)
  const [state, setState] = useState<SaveState>('idle')

  // Reconcile local inputs when the match's scores change underneath us.
  useEffect(() => {
    setHome(match.homeScore)
    setAway(match.awayScore)
  }, [match.homeScore, match.awayScore])

  const homeLabel = match.home
    ? teamName(match.home.code, match.home.name, i18n.resolvedLanguage)
    : match.placeholderHome ?? t('fixture.tbd')
  const awayLabel = match.away
    ? teamName(match.away.code, match.away.name, i18n.resolvedLanguage)
    : match.placeholderAway ?? t('fixture.tbd')

  const canSave = home !== null && away !== null && state !== 'saving'

  async function handleSave() {
    if (home === null || away === null) return
    setState('saving')
    try {
      await setMatchResult(match.id, home, away, 'finished')
      setState('saved')
      onSaved?.(home, away)
    } catch {
      setState('error')
    }
  }

  return (
    <div className="mt-3 rounded-xl border border-accent/30 bg-accent/[0.045] p-2.5">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        className="flex w-full items-center justify-between gap-2"
      >
        <span className="flex items-center gap-1.5 text-[0.6rem] font-semibold uppercase tracking-[0.16em] text-accent">
          <span className="rounded bg-accent/15 px-1.5 py-0.5 text-[0.55rem]">
            {t('adminResult.eyebrow')}
          </span>
          {t('adminResult.title')}
        </span>
        <span className="text-[0.7rem] text-accent/70">{open ? '▾' : '▸'}</span>
      </button>

      {open && (
        <div className="mt-2.5">
          <div className="flex items-center justify-center gap-3">
            <ScoreInput
              value={home}
              ariaLabel={t('adminResult.scoreFor', { team: homeLabel })}
              disabled={state === 'saving'}
              onChange={setHome}
            />
            <span className="text-sm font-semibold text-muted/60">:</span>
            <ScoreInput
              value={away}
              ariaLabel={t('adminResult.scoreFor', { team: awayLabel })}
              disabled={state === 'saving'}
              onChange={setAway}
            />
          </div>

          <div className="mt-2.5 flex items-center justify-between gap-2">
            <span className="text-[0.6rem] font-semibold uppercase tracking-[0.14em]">
              {state === 'saving' && <span className="text-muted">{t('adminResult.saving')}</span>}
              {state === 'saved' && <span className="text-accent">✓ {t('adminResult.saved')}</span>}
              {state === 'error' && (
                <span className="text-red-400/80">{t('adminResult.error')}</span>
              )}
            </span>
            <button
              type="button"
              disabled={!canSave}
              onClick={handleSave}
              className="rounded-lg border border-accent/50 bg-accent/15 px-3 py-1.5 text-xs font-semibold text-accent transition-colors hover:bg-accent/25 disabled:cursor-not-allowed disabled:opacity-40"
            >
              {t('adminResult.save')}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
