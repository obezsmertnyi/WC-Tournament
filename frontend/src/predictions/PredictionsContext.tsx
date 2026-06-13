import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import type { MyPrediction } from '../types'
import { ApiError, fetchMyPredictions, savePrediction, type PredictionInput } from '../lib/api'
import { useAuth } from '../auth/AuthContext'

/** Per-match save lifecycle, surfaced as a subtle inline state on the card. */
export type SaveState = 'idle' | 'saving' | 'saved' | 'locked' | 'error'

interface PredictionsValue {
  /** matchId → the user's current prediction (optimistic). */
  byMatch: Map<number, MyPrediction>
  saveStateOf: (matchId: number, forUserId?: number) => SaveState
  /**
   * Queue a debounced save for a match (called on input change/blur).
   * When `input.forUserId` is set (admin acting on behalf of a player), the
   * save is sent with that id, the kickoff lock is bypassed by the backend,
   * and the admin's own optimistic `byMatch` is left untouched.
   */
  save: (matchId: number, input: PredictionInput) => void
}

const PredictionsContext = createContext<PredictionsValue | null>(null)

const DEBOUNCE_MS = 600

/** Save-state / timer key — per match, and per target user when admin-writing. */
function keyOf(matchId: number, forUserId?: number): string {
  return forUserId ? `${matchId}:${forUserId}` : `${matchId}`
}

/**
 * Holds the signed-in user's predictions and debounced PUT saves. Prefilled
 * from `GET /api/predictions/me`. When anonymous it stays empty and inert.
 */
export function PredictionsProvider({ children }: { children: ReactNode }) {
  const { status } = useAuth()
  const [byMatch, setByMatch] = useState<Map<number, MyPrediction>>(new Map())
  const [saveStates, setSaveStates] = useState<Map<string, SaveState>>(new Map())

  const timers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())
  const savedTimers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())

  // Load (or clear) predictions when auth status changes.
  useEffect(() => {
    if (status !== 'authenticated') {
      setByMatch(new Map())
      setSaveStates(new Map())
      return
    }
    const controller = new AbortController()
    fetchMyPredictions(controller.signal)
      .then((preds) => {
        if (controller.signal.aborted) return
        setByMatch(new Map(preds.map((p) => [p.matchId, p])))
      })
      .catch(() => {
        /* leave empty; cards still render read-only / editable as appropriate */
      })
    return () => controller.abort()
  }, [status])

  // Clean up any pending timers on unmount.
  useEffect(() => {
    const pending = timers.current
    const settled = savedTimers.current
    return () => {
      pending.forEach((t) => clearTimeout(t))
      settled.forEach((t) => clearTimeout(t))
    }
  }, [])

  const setSaveState = useCallback((key: string, s: SaveState) => {
    setSaveStates((prev) => {
      const next = new Map(prev)
      next.set(key, s)
      return next
    })
  }, [])

  const flush = useCallback(
    (matchId: number, input: PredictionInput) => {
      const key = keyOf(matchId, input.forUserId)
      setSaveState(key, 'saving')
      savePrediction(matchId, input)
        .then(() => {
          setSaveState(key, 'saved')
          // Auto-fade the "saved" affordance after a moment.
          const prev = savedTimers.current.get(key)
          if (prev) clearTimeout(prev)
          savedTimers.current.set(
            key,
            setTimeout(() => setSaveState(key, 'idle'), 1800),
          )
        })
        .catch((err) => {
          if (err instanceof ApiError && err.status === 409) {
            setSaveState(key, 'locked')
          } else {
            setSaveState(key, 'error')
          }
        })
    },
    [setSaveState],
  )

  const save = useCallback(
    (matchId: number, input: PredictionInput) => {
      // Optimistic local update so inputs stay responsive — but only for the
      // signed-in user's own predictions. Admin writes for another player must
      // not clobber the admin's own pick in `byMatch`.
      if (!input.forUserId) {
        setByMatch((prev) => {
          const next = new Map(prev)
          next.set(matchId, {
            matchId,
            home: input.home,
            away: input.away,
            winnerPickTeamId: input.winnerPickTeamId ?? null,
          })
          return next
        })
      }

      const key = keyOf(matchId, input.forUserId)
      const existing = timers.current.get(key)
      if (existing) clearTimeout(existing)
      timers.current.set(
        key,
        setTimeout(() => {
          timers.current.delete(key)
          flush(matchId, input)
        }, DEBOUNCE_MS),
      )
    },
    [flush],
  )

  const saveStateOf = useCallback(
    (matchId: number, forUserId?: number): SaveState =>
      saveStates.get(keyOf(matchId, forUserId)) ?? 'idle',
    [saveStates],
  )

  const value = useMemo<PredictionsValue>(
    () => ({ byMatch, saveStateOf, save }),
    [byMatch, saveStateOf, save],
  )

  return <PredictionsContext.Provider value={value}>{children}</PredictionsContext.Provider>
}

export function usePredictions(): PredictionsValue {
  const ctx = useContext(PredictionsContext)
  if (!ctx) throw new Error('usePredictions must be used within a PredictionsProvider')
  return ctx
}
