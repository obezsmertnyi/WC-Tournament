import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ApiError, GOOGLE_LOGIN_URL } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
import { STARS } from '../lib/stars'
import LanguageSwitcher from './LanguageSwitcher'

/**
 * Full-screen sign-in wall. Anonymous users see ONLY this (no app chrome) — the
 * whole app is gated behind login. Doubles as the entry hero: the featured
 * players are shown LARGE and clearly visible across the top (this is the place
 * the photos finally get real space), with the login card below.
 */
export default function LoginScreen() {
  const { t } = useTranslation()
  const { loginDev, loginAdmin } = useAuth()
  const [nickname, setNickname] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<null | 'notFound' | 'failed'>(null)
  const [adminMode, setAdminMode] = useState(false)
  const [password, setPassword] = useState('')
  const [adminBusy, setAdminBusy] = useState(false)
  const [adminError, setAdminError] = useState(false)

  async function submitDev(e: React.FormEvent) {
    e.preventDefault()
    const name = nickname.trim()
    if (!name || busy) return
    setBusy(true)
    setError(null)
    try {
      await loginDev(name)
    } catch (err) {
      setError(err instanceof ApiError && err.status === 404 ? 'notFound' : 'failed')
    } finally {
      setBusy(false)
    }
  }

  async function submitAdmin(e: React.FormEvent) {
    e.preventDefault()
    if (!password || adminBusy) return
    setAdminBusy(true)
    setAdminError(false)
    try {
      await loginAdmin(password)
    } catch {
      setAdminError(true)
    } finally {
      setAdminBusy(false)
    }
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-b from-bg to-bg-end">
      {/* Players hero — large, visible portraits across the top, fading into bg. */}
      <div className="absolute inset-x-0 top-0 h-[42vh] min-h-[240px] overflow-hidden sm:h-[46vh]">
        <div className="flex h-full w-full">
          {STARS.map((s) => (
            <div key={s.name} className="relative h-full flex-1">
              <img
                src={s.imageUrl}
                alt=""
                aria-hidden
                className="h-full w-full object-cover object-top"
                style={{ filter: 'grayscale(0.25) contrast(1.05) brightness(0.82)' }}
              />
            </div>
          ))}
        </div>
        {/* fade the photos into the page so the card below sits cleanly */}
        <span
          aria-hidden
          className="pointer-events-none absolute inset-0"
          style={{
            background:
              'linear-gradient(to bottom, rgba(11,12,14,0.15) 0%, rgba(11,12,14,0.55) 55%, #0B0C0E 100%), linear-gradient(115deg, rgba(201,162,75,0.12) 0%, rgba(11,12,14,0) 55%)',
          }}
        />
      </div>

      {/* language switch, top-right */}
      <div className="absolute right-4 top-4 z-10">
        <LanguageSwitcher />
      </div>

      {/* Login card */}
      <div className="relative z-10 flex min-h-screen items-end justify-center px-4 pb-10 pt-[36vh] sm:items-center sm:pt-0">
        <div className="w-full max-w-sm rounded-3xl border border-hairline bg-gradient-to-b from-[#15171B] to-bg p-6 shadow-[0_24px_60px_-20px_rgba(0,0,0,0.9)] backdrop-blur-xl">
          <header className="mb-5 text-center">
            <p className="text-[0.6rem] font-semibold uppercase tracking-[0.28em] text-muted/70">
              {t('auth.eyebrow')}
            </p>
            <h1 className="mt-1 text-2xl font-bold tracking-tight text-text">
              WORLD CUP <span className="text-accent tabular-nums">2026</span>
            </h1>
            <p className="mt-1.5 text-sm text-muted">{t('auth.subtitle')}</p>
          </header>

          <form onSubmit={submitDev} className="space-y-3">
            <input
              type="text"
              value={nickname}
              maxLength={40}
              onChange={(e) => setNickname(e.target.value)}
              placeholder={t('auth.nicknamePlaceholder')}
              className="h-12 w-full rounded-xl border border-hairline bg-white/[0.04] px-3.5 text-sm text-text outline-none transition-colors placeholder:text-muted/50 focus:border-accent focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)]"
            />
            {error === 'notFound' && (
              <p className="text-xs text-accent/90">{t('auth.noSuchPlayer')}</p>
            )}
            {error === 'failed' && <p className="text-xs text-red-400/90">{t('auth.failed')}</p>}
            <button
              type="submit"
              disabled={busy || nickname.trim() === ''}
              className="h-12 w-full rounded-xl bg-accent text-sm font-semibold uppercase tracking-[0.14em] text-bg transition-opacity hover:opacity-90 disabled:opacity-40"
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
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl border border-hairline bg-white/[0.04] text-sm font-medium text-text transition-colors hover:border-white/20 hover:bg-white/[0.07]"
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
        </div>
      </div>
    </div>
  )
}

function GoogleGlyph() {
  return (
    <svg viewBox="0 0 48 48" className="h-4 w-4" aria-hidden="true">
      <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
      <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z" />
      <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z" />
      <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z" />
    </svg>
  )
}
