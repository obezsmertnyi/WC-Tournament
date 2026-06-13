import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { FIFA_TO_ISO } from '../lib/flags'
import { teamName } from '../lib/teamNames'
import { useAuth } from '../auth/AuthContext'
import Avatar from './../components/Avatar'
import Flag from '../components/Flag'

/** Team codes offered as favorite picks — every code with a known flag. */
const TEAM_CODES = Object.keys(FIFA_TO_ISO).sort()

export default function Profile() {
  const { t, i18n } = useTranslation()
  const { user, status, updateProfile } = useAuth()
  const navigate = useNavigate()

  const [nickname, setNickname] = useState('')
  const [favorite, setFavorite] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (user) {
      setNickname(user.nickname)
      setFavorite(user.favoriteTeamCode)
    }
  }, [user])

  // Bounce anonymous visitors back to the calendar.
  useEffect(() => {
    if (status === 'anonymous') navigate('/', { replace: true })
  }, [status, navigate])

  const sortedCodes = useMemo(
    () =>
      [...TEAM_CODES].sort((a, b) =>
        teamName(a, a, i18n.resolvedLanguage).localeCompare(
          teamName(b, b, i18n.resolvedLanguage),
          i18n.resolvedLanguage === 'uk' ? 'uk' : 'en',
        ),
      ),
    [i18n.resolvedLanguage],
  )

  if (status === 'loading' || !user) {
    return (
      <div className="mx-auto w-full max-w-2xl">
        <div className="h-40 animate-pulse rounded-2xl border border-hairline bg-surface" />
      </div>
    )
  }

  async function onSave(e: React.FormEvent) {
    e.preventDefault()
    if (busy) return
    setBusy(true)
    setSaved(false)
    setError(false)
    try {
      await updateProfile({ nickname: nickname.trim(), favoriteTeamCode: favorite })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {
      setError(true)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-2xl">
      <header className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
          {t('profile.title')}
        </h1>
        <p className="mt-1 text-sm text-muted">{t('profile.subtitle')}</p>
      </header>

      <div className="rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] p-6 backdrop-blur-md">
        <div className="mb-6 flex items-center gap-4">
          <Avatar src={user.avatarUrl} nickname={user.nickname} className="h-16 w-16 text-xl" />
          <div className="min-w-0">
            <p className="truncate text-lg font-semibold text-text">{user.nickname}</p>
            <p className="text-xs uppercase tracking-[0.14em] text-muted/70">
              {user.role === 'admin' ? t('profile.roleAdmin') : t('profile.roleUser')}
            </p>
          </div>
        </div>

        <form onSubmit={onSave} className="space-y-5">
          <label className="block">
            <span className="mb-1.5 block text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80">
              {t('profile.nickname')}
            </span>
            <input
              type="text"
              value={nickname}
              maxLength={40}
              onChange={(e) => setNickname(e.target.value)}
              className="h-11 w-full rounded-xl border border-hairline bg-white/[0.04] px-3.5 text-sm text-text outline-none transition-colors focus:border-accent focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)]"
            />
          </label>

          <div>
            <span className="mb-2 block text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80">
              {t('profile.favoriteTeam')}
            </span>
            <div className="grid max-h-64 grid-cols-4 gap-1.5 overflow-y-auto rounded-xl border border-hairline bg-white/[0.02] p-2 sm:grid-cols-6">
              {sortedCodes.map((code) => {
                const active = favorite === code
                return (
                  <button
                    key={code}
                    type="button"
                    onClick={() => setFavorite(active ? null : code)}
                    title={teamName(code, code, i18n.resolvedLanguage)}
                    aria-pressed={active}
                    className={`flex flex-col items-center gap-1 rounded-lg border px-1 py-2 transition-colors ${
                      active
                        ? 'border-accent/60 bg-accent/10'
                        : 'border-transparent hover:bg-white/[0.05]'
                    }`}
                  >
                    <Flag
                      code={code}
                      flagUrl={undefined}
                      label={code}
                      className="h-[1.05rem] w-6"
                    />
                    <span
                      className={`text-[0.55rem] font-medium tabular-nums ${
                        active ? 'text-accent' : 'text-muted/70'
                      }`}
                    >
                      {code}
                    </span>
                  </button>
                )
              })}
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={busy || nickname.trim() === ''}
              className="h-11 rounded-xl bg-accent px-6 text-sm font-semibold uppercase tracking-[0.14em] text-bg transition-opacity hover:opacity-90 disabled:opacity-40"
            >
              {busy ? t('profile.saving') : t('profile.save')}
            </button>
            {saved && (
              <span className="text-xs font-semibold uppercase tracking-[0.14em] text-accent">
                ✓ {t('profile.saved')}
              </span>
            )}
            {error && (
              <span className="text-xs font-semibold uppercase tracking-[0.14em] text-red-400/90">
                {t('profile.error')}
              </span>
            )}
          </div>
        </form>
      </div>
    </div>
  )
}
