import { useTranslation } from 'react-i18next'
import { useAuth } from '../auth/AuthContext'
import { isRestricted } from '../lib/access'

/**
 * Thin banner shown to users in a restricted demo tier, explaining that they
 * are in preview mode and an organizer must grant access. Copy differs for the
 * browse-only ('none') and read-only ('ro') tiers. Renders nothing for full
 * ('rw') users and when demo mode is off.
 */
export default function DemoBanner() {
  const { t } = useTranslation()
  const { user } = useAuth()
  if (!isRestricted(user)) return null
  const key = user!.access === 'ro' ? 'demo.banner.ro' : 'demo.banner.none'
  return (
    <div className="mb-4 flex items-center gap-2.5 rounded-xl border border-accent/30 bg-accent/[0.08] px-3.5 py-2.5 text-xs text-text/90 backdrop-blur-md">
      <span className="text-base leading-none">👋</span>
      <span>{t(key)}</span>
    </div>
  )
}
