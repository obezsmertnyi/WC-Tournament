import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { AdminPlayer } from '../types'
import { ApiError, createPlayer, deletePlayer, fetchAdminUsers } from '../lib/api'
import Avatar from './Avatar'

/**
 * Admin-only player roster. Lists `GET /api/admin/users`, adds a player by
 * nickname (POST), and removes a player (DELETE, confirmed). Rendered on the
 * Profile page below the profile card; never shown to regular players.
 */
export default function AdminPlayers() {
  const { t } = useTranslation()
  const [players, setPlayers] = useState<AdminPlayer[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)

  const [nickname, setNickname] = useState('')
  const [adding, setAdding] = useState(false)
  const [addError, setAddError] = useState(false)
  const [removingId, setRemovingId] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const controller = new AbortController()
    fetchAdminUsers(controller.signal)
      .then((list) => {
        if (controller.signal.aborted) return
        setPlayers(list)
        setLoading(false)
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
        setLoadError(true)
        setLoading(false)
      })
    return () => controller.abort()
  }, [])

  async function onAdd(e: React.FormEvent) {
    e.preventDefault()
    const name = nickname.trim()
    if (!name || adding) return
    setAdding(true)
    setAddError(false)
    try {
      const created = await createPlayer(name)
      setPlayers((prev) =>
        prev.some((p) => p.id === created.id) ? prev : [...prev, created],
      )
      setNickname('')
      inputRef.current?.focus()
    } catch {
      setAddError(true)
    } finally {
      setAdding(false)
    }
  }

  async function onRemove(player: AdminPlayer) {
    if (removingId) return
    if (!window.confirm(t('admin.removeConfirm', { player: player.nickname }))) return
    setRemovingId(player.id)
    try {
      await deletePlayer(player.id)
      setPlayers((prev) => prev.filter((p) => p.id !== player.id))
    } catch (err) {
      // Surface 404 as already-gone; otherwise alert generically.
      if (err instanceof ApiError && err.status === 404) {
        setPlayers((prev) => prev.filter((p) => p.id !== player.id))
      } else {
        window.alert(t('admin.removeError'))
      }
    } finally {
      setRemovingId(null)
    }
  }

  return (
    <section className="mt-6 rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] p-6 backdrop-blur-md">
      <header className="mb-4">
        <p className="text-[0.65rem] font-semibold uppercase tracking-[0.28em] text-accent">
          {t('admin.eyebrow')}
        </p>
        <h2 className="mt-1.5 text-lg font-semibold text-text">{t('admin.players')}</h2>
        <p className="mt-1 text-sm text-muted">{t('admin.playersSubtitle')}</p>
      </header>

      <form onSubmit={onAdd} className="mb-5 flex items-end gap-2">
        <label className="block flex-1">
          <span className="mb-1.5 block text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80">
            {t('admin.addPlayer')}
          </span>
          <input
            ref={inputRef}
            type="text"
            value={nickname}
            maxLength={40}
            onChange={(e) => setNickname(e.target.value)}
            placeholder={t('admin.nicknamePlaceholder')}
            className="h-11 w-full rounded-xl border border-hairline bg-white/[0.04] px-3.5 text-sm text-text outline-none transition-colors placeholder:text-muted/50 focus:border-accent focus:shadow-[0_0_0_3px_rgba(201,162,75,0.18)]"
          />
        </label>
        <button
          type="submit"
          disabled={adding || nickname.trim() === ''}
          className="h-11 shrink-0 rounded-xl bg-accent px-5 text-sm font-semibold uppercase tracking-[0.14em] text-bg transition-opacity hover:opacity-90 disabled:opacity-40"
        >
          {adding ? t('admin.adding') : t('admin.add')}
        </button>
      </form>
      {addError && (
        <p className="-mt-3 mb-4 text-xs text-red-400/90">{t('admin.addError')}</p>
      )}

      {loading ? (
        <div className="h-24 animate-pulse rounded-xl border border-hairline bg-white/[0.02]" />
      ) : loadError ? (
        <p className="text-sm text-red-400/90">{t('admin.loadError')}</p>
      ) : players.length === 0 ? (
        <p className="text-sm text-muted/70">{t('admin.empty')}</p>
      ) : (
        <ul className="divide-y divide-hairline rounded-xl border border-hairline bg-white/[0.02]">
          {players.map((p) => (
            <li key={p.id} className="flex items-center gap-3 px-3 py-2.5">
              <Avatar src={p.avatarUrl} nickname={p.nickname} className="h-8 w-8 text-xs" />
              <span className="min-w-0 flex-1 truncate text-sm font-medium text-text">
                {p.nickname}
              </span>
              {p.role === 'admin' && (
                <span className="rounded-full border border-accent/40 bg-accent/10 px-2 py-0.5 text-[0.55rem] font-semibold uppercase tracking-[0.14em] text-accent">
                  {t('profile.roleAdmin')}
                </span>
              )}
              <button
                type="button"
                onClick={() => onRemove(p)}
                disabled={removingId === p.id}
                aria-label={t('admin.removePlayer', { player: p.nickname })}
                className="shrink-0 rounded-lg border border-hairline px-2.5 py-1 text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80 transition-colors hover:border-red-400/40 hover:text-red-400/90 disabled:opacity-40"
              >
                {removingId === p.id ? t('admin.removing') : t('admin.remove')}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
