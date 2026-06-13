import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import type { User } from '../types'
import { devLogin, fetchMe, logout as apiLogout, updateMe, type ProfilePatch } from '../lib/api'

type Status = 'loading' | 'authenticated' | 'anonymous'

interface AuthValue {
  user: User | null
  status: Status
  /** Dev login by nickname; resolves the session and stores the user. */
  loginDev: (nickname: string) => Promise<void>
  logout: () => Promise<void>
  /** Patch the current user's profile and update local state. */
  updateProfile: (patch: ProfilePatch) => Promise<User>
  /** Re-fetch /api/me (e.g. after returning from the Google redirect). */
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthValue | null>(null)

/**
 * Loads `GET /api/me` on mount to resolve the session. Every call goes through
 * lib/api, which sends the HttpOnly session cookie (`credentials: 'include'`).
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [status, setStatus] = useState<Status>('loading')

  const refresh = useCallback(async () => {
    try {
      const me = await fetchMe()
      setUser(me)
      setStatus(me ? 'authenticated' : 'anonymous')
    } catch {
      // Network/5xx: treat as anonymous so the UI stays usable.
      setUser(null)
      setStatus('anonymous')
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    fetchMe(controller.signal)
      .then((me) => {
        if (controller.signal.aborted) return
        setUser(me)
        setStatus(me ? 'authenticated' : 'anonymous')
      })
      .catch(() => {
        if (controller.signal.aborted) return
        setUser(null)
        setStatus('anonymous')
      })
    return () => controller.abort()
  }, [])

  const loginDev = useCallback(async (nickname: string) => {
    const me = await devLogin(nickname.trim())
    setUser(me)
    setStatus('authenticated')
  }, [])

  const logout = useCallback(async () => {
    try {
      await apiLogout()
    } finally {
      setUser(null)
      setStatus('anonymous')
    }
  }, [])

  const updateProfile = useCallback(async (patch: ProfilePatch) => {
    const me = await updateMe(patch)
    setUser(me)
    setStatus('authenticated')
    return me
  }, [])

  const value = useMemo<AuthValue>(
    () => ({ user, status, loginDev, logout, updateProfile, refresh }),
    [user, status, loginDev, logout, updateProfile, refresh],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider')
  return ctx
}
