import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { ApiError, GOOGLE_LOGIN_URL } from '../lib/api'
import { useAuth } from '../auth/AuthContext'

interface LoginModalProps {
  open: boolean
  onClose: () => void
}

/**
 * Small auth sheet: a dev-login (nickname → POST /api/auth/dev-login) plus a
 * "Continue with Google" link to the redirect endpoint. Closes on success,
 * Escape, or backdrop click.
 */
export default function LoginModal({ open, onClose }: LoginModalProps) {
  const { t } = useTranslation()
  const { loginDev, loginAdmin } = useAuth()
  const [nickname, setNickname] = useState('')
  const [busy, setBusy] = useState(false)
  // null = no error; 'notFound' = unknown nickname (404); 'failed' = other.
  const [error, setError] = useState<null | 'notFound' | 'failed'>(null)
  const [adminMode, setAdminMode] = useState(false)
  const [password, setPassword] = useState('')
  const [adminBusy, setAdminBusy] = useState(false)
  const [adminError, setAdminError] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    setError(null)
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    const id = setTimeout(() => inputRef.current?.focus(), 60)
    return () => {
      document.removeEventListener('keydown', onKey)
      clearTimeout(id)
    }
  }, [open, onClose])

  async function submitAdmin(e: React.FormEvent) {
    e.preventDefault()
    if (!password || adminBusy) return
    setAdminBusy(true)
    setAdminError(false)
    try {
      await loginAdmin(password)
      setPassword('')
      onClose()
    } catch {
      setAdminError(true)
    } finally {
      setAdminBusy(false)
    }
  }

  async function submitDev(e: React.FormEvent) {
    e.preventDefault()
    const name = nickname.trim()
    if (!name || busy) return
    setBusy(true)
    setError(null)
    try {
      await loginDev(name)
      setNickname('')
      onClose()
    } catch (err) {
      // 404 → nickname no longer auto-creates an account; ask an admin.
      setError(err instanceof ApiError && err.status === 404 ? 'notFound' : 'failed')
    } finally {
      setBusy(false)
    }
  }

  return createPortal(
    <AnimatePresence>
      {open && (
        <motion.div
          className="fixed inset-0 z-[100] flex items-end justify-center sm:items-center"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
        >
          <button
            type="button"
            aria-label={t('auth.close')}
            onClick={onClose}
            className="absolute inset-0 bg-black/60 backdrop-blur-sm"
          />
          <motion.div
            role="dialog"
            aria-modal="true"
            aria-label={t('auth.signIn')}
            initial={{ y: 24, opacity: 0, scale: 0.98 }}
            animate={{ y: 0, opacity: 1, scale: 1 }}
            exit={{ y: 24, opacity: 0, scale: 0.98 }}
            transition={{ type: 'spring', stiffness: 320, damping: 30 }}
            className="relative z-10 w-full max-w-sm rounded-t-3xl border border-hairline bg-gradient-to-b from-[#15171B] to-bg p-6 shadow-[0_-8px_40px_-12px_rgba(0,0,0,0.8)] sm:rounded-3xl sm:shadow-[0_24px_60px_-20px_rgba(0,0,0,0.9)]"
          >
            <header className="mb-5">
              <p className="text-[0.65rem] font-semibold uppercase tracking-[0.28em] text-accent">
                {t('auth.eyebrow')}
              </p>
              <h2 className="mt-1.5 text-lg font-semibold text-text">{t('auth.signIn')}</h2>
              <p className="mt-1 text-sm text-muted">{t('auth.subtitle')}</p>
            </header>

            <form onSubmit={submitDev} className="space-y-3">
              <label className="block">
                <span className="mb-1.5 block text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80">
                  {t('auth.nickname')}
                </span>
                <input
                  ref={inputRef}
                  type="text"
                  value={nickname}
                  maxLength={40}
                  onChange={(e) => setNickname(e.target.value)}
                  placeholder={t('auth.nicknamePlaceholder')}
                  className="h-11 w-full rounded-xl border border-hairline bg-white/[0.04] px-3.5 text-sm text-text outline-none transition-colors placeholder:text-muted/50 focus:border-accent focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)]"
                />
              </label>
              {error === 'notFound' && (
                <p className="text-xs text-accent/90">{t('auth.noSuchPlayer')}</p>
              )}
              {error === 'failed' && <p className="text-xs text-red-400/90">{t('auth.failed')}</p>}
              <button
                type="submit"
                disabled={busy || nickname.trim() === ''}
                className="h-11 w-full rounded-xl bg-accent text-sm font-semibold uppercase tracking-[0.14em] text-bg transition-opacity hover:opacity-90 disabled:opacity-40"
              >
                {busy ? t('auth.signingIn') : t('auth.devLogin')}
              </button>
            </form>

            <div className="my-4 flex items-center gap-3">
              <span className="h-px flex-1 bg-hairline" />
              <span className="text-[0.6rem] uppercase tracking-[0.16em] text-muted/60">
                {t('auth.or')}
              </span>
              <span className="h-px flex-1 bg-hairline" />
            </div>

            <a
              href={GOOGLE_LOGIN_URL}
              className="flex h-11 w-full items-center justify-center gap-2 rounded-xl border border-hairline bg-white/[0.04] text-sm font-medium text-text transition-colors hover:border-white/20 hover:bg-white/[0.07]"
            >
              <GoogleGlyph />
              {t('auth.google')}
            </a>

            {!adminMode ? (
              <button
                type="button"
                onClick={() => setAdminMode(true)}
                className="mt-4 block w-full text-center text-[0.7rem] uppercase tracking-[0.16em] text-muted/60 transition-colors hover:text-accent"
              >
                {t('auth.adminEntry')}
              </button>
            ) : (
              <form onSubmit={submitAdmin} className="mt-4 space-y-2 border-t border-hairline pt-4">
                <span className="block text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80">
                  {t('auth.adminPassword')}
                </span>
                <input
                  type="password"
                  value={password}
                  autoComplete="current-password"
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="h-11 w-full rounded-xl border border-hairline bg-white/[0.04] px-3.5 text-sm text-text outline-none transition-colors placeholder:text-muted/40 focus:border-accent focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)]"
                />
                {adminError && <p className="text-xs text-red-400/90">{t('auth.adminFailed')}</p>}
                <button
                  type="submit"
                  disabled={adminBusy || password === ''}
                  className="h-11 w-full rounded-xl border border-accent/50 bg-accent/10 text-sm font-semibold uppercase tracking-[0.14em] text-accent transition-opacity hover:bg-accent/20 disabled:opacity-40"
                >
                  {adminBusy ? t('auth.signingIn') : t('auth.adminLogin')}
                </button>
              </form>
            )}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>,
    document.body,
  )
}

function GoogleGlyph() {
  return (
    <svg viewBox="0 0 48 48" className="h-4 w-4" aria-hidden="true">
      <path
        fill="#EA4335"
        d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
      />
      <path
        fill="#4285F4"
        d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
      />
      <path
        fill="#FBBC05"
        d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
      />
      <path
        fill="#34A853"
        d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
      />
    </svg>
  )
}
