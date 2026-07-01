import { useTranslation } from 'react-i18next'
import type { Match } from '../types'
import { recap } from '../lib/recap'

/**
 * Generative-UI panel: a short AI-style recap of a finished match. The text
 * comes from a grounded generator and passes the anti-hallucination guardrail
 * (ADR-0016) before it is rendered — a wrong score/team can never reach the UI.
 * Renders nothing until the match has a result.
 */
export default function MatchRecap({ match, exactGuessers }: { match: Match; exactGuessers?: string[] }) {
  const { t } = useTranslation()
  const text = recap(match, { exactGuessers })
  if (!text) return null
  return (
    <div className="mb-3 rounded-xl border border-accent/25 bg-accent/[0.06] px-3.5 py-2.5 backdrop-blur-md">
      <p className="mb-1 text-[0.6rem] font-semibold uppercase tracking-[0.14em] text-accent/70">
        {t('recap.title')}
      </p>
      <p className="text-sm leading-relaxed text-text/90">{text}</p>
    </div>
  )
}
