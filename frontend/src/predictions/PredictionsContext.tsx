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
  saveStateOf: (matchId: number) => SaveState
  /** Queue a debounced save for a match (called on input change/blur). */
  save: (matchId: number, input: PredictionInput) => void
}

const PredictionsContext = createContext<PredictionsValue | null>(null)

const DEBOUNCE_MS = 600

/**
 * Holds the signed-in user's predictions and debounced PUT saves. Prefilled
 * from `GET /api/predictions/me`. When anonymous it stays empty and inert.
 */
export function PredictionsProvider({ children }: { children: ReactNode }) {
  const { status } = useAuth()
  const [byMatch, setByMatch] = useState<Map<number, MyPrediction>>(new Map())
  const [saveStates, setSaveStates] = useState<Map<number, SaveState>>(new Map())

  const timers = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map())
  const savedTimers = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map())

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

  const setSaveState = useCallback((matchId: number, s: SaveState) => {
    setSaveStates((prev) => {
      const next = new Map(prev)
      next.set(matchId, s)
      return next
    })
  }, [])

  const flush = useCallback(
    (matchId: number, input: PredictionInput) => {
      setSaveState(matchId, 'saving')
      savePrediction(matchId, input)
        .then(() => {
          setSaveState(matchId, 'saved')
          // Auto-fade the "saved" affordance after a moment.
          const prev = savedTimers.current.get(matchId)
          if (prev) clearTimeout(prev)
          savedTimers.current.set(
            matchId,
            setTimeout(() => setSaveState(matchId, 'idle'), 1800),
          )
        })
        .catch((err) => {
          if (err instanceof ApiError && err.status === 409) {
            setSaveState(matchId, 'locked')
          } else {
            setSaveState(matchId, 'error')
          }
        })
    },
    [setSaveState],
  )

  const save = useCallback(
    (matchId: number, input: PredictionInput) => {
      // Optimistic local update so inputs stay responsive.
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

      const existing = timers.current.get(matchId)
      if (existing) clearTimeout(existing)
      timers.current.set(
        matchId,
        setTimeout(() => {
          timers.current.delete(matchId)
          flush(matchId, input)
        }, DEBOUNCE_MS),
      )
    },
    [flush],
  )

  const saveStateOf = useCallback(
    (matchId: number): SaveState => saveStates.get(matchId) ?? 'idle',
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
