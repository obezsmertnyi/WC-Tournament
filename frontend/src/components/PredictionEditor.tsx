import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { AdminPlayer, Match } from '../types'
import { hasKickedOff } from '../lib/fixtures'
import { teamName } from '../lib/teamNames'
import { fetchAdminUsers } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
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
 *
 * Admins additionally get an "as [player ▾]" selector that lets them write a
 * prediction on behalf of any player (bypassing the kickoff lock). While a
 * player is selected the editor edits that player's pick — never the admin's.
 */
export default function PredictionEditor({ match }: PredictionEditorProps) {
  const { t, i18n } = useTranslation()
  const { user } = useAuth()
  const { byMatch, save, saveStateOf } = usePredictions()
  const isAdmin = user?.role === 'admin'

  // ── Admin "predict as player" selection ────────────────────────────────────
  const [players, setPlayers] = useState<AdminPlayer[]>([])
  // '' means the admin is editing their own prediction (default).
  const [asPlayerId, setAsPlayerId] = useState<string>('')

  useEffect(() => {
    if (!isAdmin) return
    const controller = new AbortController()
    fetchAdminUsers(controller.signal)
      .then((list) => {
        if (controller.signal.aborted) return
        // Hide the acting admin from the roster of pickable players. IDs come
        // off the wire as numbers; compare as strings to stay type-agnostic.
        setPlayers(list.filter((p) => p.role !== 'admin' || String(p.id) !== String(user?.id)))
      })
      .catch(() => {
        /* leave empty; selector simply offers no players */
      })
    return () => controller.abort()
  }, [isAdmin, user?.id])

  // The roster comes off the wire with numeric ids, while the <select> value is
  // always a string — compare as strings so the lookup actually matches.
  const actingFor = asPlayerId
    ? players.find((p) => String(p.id) === asPlayerId) ?? null
    : null
  // Send the target id to the backend as a number (it binds `*int64`).
  const forUserId = actingFor ? Number(actingFor.id) : undefined

  // The admin's own prediction lives in `byMatch`; another player's does not
  // (we never prefetch it), so the editor starts blank when a player is chosen.
  const own = byMatch.get(match.id)

  const [home, setHome] = useState<number | null>(own?.home ?? null)
  const [away, setAway] = useState<number | null>(own?.away ?? null)
  const [winner, setWinner] = useState<number | null>(own?.winnerPickTeamId ?? null)

  // Reconcile when the source prediction changes: either the admin's own
  // prediction refreshes from the server, or the acting-for player switches.
  useEffect(() => {
    if (forUserId) {
      setHome(null)
      setAway(null)
      setWinner(null)
    } else {
      setHome(own?.home ?? null)
      setAway(own?.away ?? null)
      setWinner(own?.winnerPickTeamId ?? null)
    }
  }, [forUserId, own?.home, own?.away, own?.winnerPickTeamId])

  const kickedOff = hasKickedOff(match)
  // Admins are never locked — they may set/correct any prediction (their own or,
  // via the selector, any player's) at any time; the backend audits it. Regular
  // players lock at kickoff.
  const locked = kickedOff && !isAdmin
  const isKnockout = match.stage !== 'group'
  const saveState = saveStateOf(match.id, forUserId)

  // A knockout draw is incomplete until you also pick who advances (ET/pens).
  const needsAdvancer =
    isKnockout && home !== null && away !== null && home === away && winner === null

  const homeId = match.home?.id ?? null
  const awayId = match.away?.id ?? null
  const homeLabel = match.home
    ? teamName(match.home.code, match.home.name, i18n.resolvedLanguage)
    : t('fixture.tbd')
  const awayLabel = match.away
    ? teamName(match.away.code, match.away.name, i18n.resolvedLanguage)
    : t('fixture.tbd')

  function commit(nextHome: number | null, nextAway: number | null, nextWinner: number | null) {
    if (nextHome === null || nextAway === null) return // need both scores to save
    // Knockout draw: don't save until the advancer is chosen (the hint shows).
    if (isKnockout && nextHome === nextAway && nextWinner === null) return
    save(match.id, {
      home: nextHome,
      away: nextAway,
      winnerPickTeamId: isKnockout ? nextWinner : null,
      forUserId,
    })
  }

  // Compact admin selector — "as [player ▾]".
  const adminSelector = isAdmin && (
    <div className="mb-2 flex items-center gap-2">
      <span className="text-[0.6rem] font-semibold uppercase tracking-[0.14em] text-accent/70">
        {t('predict.asPlayer')}
      </span>
      <select
        aria-label={t('predict.selectPlayer')}
        value={asPlayerId}
        onChange={(e) => setAsPlayerId(e.target.value)}
        className="h-7 max-w-[10rem] flex-1 truncate rounded-md border border-hairline bg-white/[0.04] px-2 text-xs text-text outline-none transition-colors focus:border-accent"
      >
        <option value="">{t('predict.asMyself')}</option>
        {players.map((p) => (
          <option key={p.id} value={p.id}>
            {p.nickname}
          </option>
        ))}
      </select>
    </div>
  )

  // ── Locked (kicked off, editing own pick): read-only summary ───────────────
  if (locked) {
    const scored = own && typeof (own as { points?: number }).points === 'number'
    return (
      <div className="mt-3 border-t border-hairline pt-2.5">
        {adminSelector}
        {!own ? (
          <p className="text-[0.65rem] uppercase tracking-[0.14em] text-muted/60">
            {t('predict.noPick')}
          </p>
        ) : (
          <div className="flex items-center justify-between gap-2">
            <span className="text-[0.6rem] font-semibold uppercase tracking-[0.14em] text-muted/70">
              {t('predict.yourPick')}
            </span>
            <span className="flex items-center gap-2">
              <span className="tabular-nums text-sm font-bold text-text">
                {own.home}–{own.away}
              </span>
              {scored && (
                <span className="rounded-full border border-accent/40 bg-accent/10 px-2 py-0.5 text-[0.6rem] font-semibold text-accent">
                  ✓ {t('predict.pointsBadge', { points: (own as { points?: number }).points })}
                </span>
              )}
            </span>
          </div>
        )}
      </div>
    )
  }

  // ── Editable (before kickoff, or admin acting for a player) ────────────────
  return (
    <div className="mt-3 border-t border-hairline pt-3">
      {adminSelector}

      <div className="flex items-center justify-between gap-2">
        <span className="truncate text-[0.6rem] font-semibold uppercase tracking-[0.14em] text-accent/80">
          {actingFor
            ? t('predict.editingFor', { player: actingFor.nickname })
            : t('predict.yourPick')}
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

      {isKnockout && home !== null && away !== null && home === away && (
        <div className="mt-3">
          <p
            className={`mb-1.5 text-center text-[0.6rem] uppercase tracking-[0.14em] ${
              needsAdvancer ? 'font-semibold text-accent' : 'text-muted/70'
            }`}
          >
            {needsAdvancer ? t('predict.advancerRequired') : t('predict.advancer')}
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
