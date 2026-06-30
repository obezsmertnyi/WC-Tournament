import { useTranslation } from 'react-i18next'

/**
 * Placeholder shown in place of gated content while the signed-in user is in a
 * restricted demo tier. `reason` selects the explanatory copy: 'seeOthers' for
 * read-gated panels (leaderboard, reveals, audit, scorers), 'participate' for
 * write-gated ones (bonus picks). The nav stays clickable — only the content
 * behind it is locked — so a previewer can still explore the whole UI.
 */
export default function DemoLocked({ reason }: { reason: 'seeOthers' | 'participate' }) {
  const { t } = useTranslation()
  const key = reason === 'participate' ? 'demo.lockedParticipate' : 'demo.lockedSeeOthers'
  return (
    <div className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center backdrop-blur-md">
      <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full border border-accent/30 bg-accent/10 text-2xl">
        🔒
      </div>
      <p className="text-sm font-semibold text-text">{t(key)}</p>
      <p className="mx-auto mt-1.5 max-w-sm text-xs text-muted">{t('demo.lockedHint')}</p>
    </div>
  )
}
