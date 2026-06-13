import { useEffect, useRef, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import Avatar from './Avatar'
import LoginModal from './LoginModal'

/**
 * App-bar auth affordance. Anonymous → "Sign in" button opening the login
 * sheet. Authenticated → avatar + nickname with a Profile / Logout menu.
 */
export default function AuthControl() {
  const { t } = useTranslation()
  const { user, status, logout } = useAuth()
  const navigate = useNavigate()
  const [modalOpen, setModalOpen] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuOpen) return
    const onClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuOpen(false)
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuOpen(false)
    }
    document.addEventListener('mousedown', onClick)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onClick)
      document.removeEventListener('keydown', onKey)
    }
  }, [menuOpen])

  if (status === 'loading') {
    return <div className="h-8 w-8 animate-pulse rounded-full bg-white/10" aria-hidden="true" />
  }

  if (status !== 'authenticated' || !user) {
    return (
      <>
        <button
          type="button"
          onClick={() => setModalOpen(true)}
          className="rounded-full border border-accent/40 bg-accent/10 px-3 py-1.5 text-[0.7rem] font-semibold uppercase tracking-[0.14em] text-accent transition-colors hover:bg-accent/15"
        >
          {t('auth.signIn')}
        </button>
        <LoginModal open={modalOpen} onClose={() => setModalOpen(false)} />
      </>
    )
  }

  return (
    <div className="relative" ref={menuRef}>
      <button
        type="button"
        onClick={() => setMenuOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={menuOpen}
        className="flex items-center gap-2 rounded-full border border-hairline bg-surface py-0.5 pl-0.5 pr-2.5 backdrop-blur-md transition-colors hover:border-white/20"
      >
        <Avatar src={user.avatarUrl} nickname={user.nickname} className="h-7 w-7 text-xs" />
        <span className="hidden max-w-[8rem] truncate text-xs font-medium text-text sm:inline">
          {user.nickname}
        </span>
      </button>

      <AnimatePresence>
        {menuOpen && (
          <motion.div
            role="menu"
            initial={{ opacity: 0, y: -6, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -6, scale: 0.98 }}
            transition={{ duration: 0.16 }}
            className="absolute right-0 top-full z-50 mt-2 w-44 overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-[#15171B] to-bg p-1.5 shadow-[0_16px_40px_-16px_rgba(0,0,0,0.9)]"
          >
            <button
              type="button"
              role="menuitem"
              onClick={() => {
                setMenuOpen(false)
                navigate('/profile')
              }}
              className="block w-full rounded-xl px-3 py-2 text-left text-sm text-text transition-colors hover:bg-white/[0.06]"
            >
              {t('auth.profile')}
            </button>
            <button
              type="button"
              role="menuitem"
              onClick={() => {
                setMenuOpen(false)
                void logout().then(() => navigate('/'))
              }}
              className="block w-full rounded-xl px-3 py-2 text-left text-sm text-muted transition-colors hover:bg-white/[0.06] hover:text-text"
            >
              {t('auth.logout')}
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
