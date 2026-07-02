import { useTranslation } from 'react-i18next'
import type { AiCard as Card } from '../lib/aiApi'

/**
 * Structured club/player card from the AI assistant (ADR-0017 / FR-092). The JSON
 * is already validated server-side; when confidence < high we flag it as possibly
 * outdated (model knowledge, not live data).
 */
export default function AiCard({ card }: { card: Card }) {
  const { t } = useTranslation()
  const row = (label: string, value?: string) =>
    value ? (
      <div className="flex justify-between gap-3 text-sm">
        <dt className="text-muted/70">{label}</dt>
        <dd className="text-right font-medium text-text">{value}</dd>
      </div>
    ) : null

  return (
    <article className="rounded-2xl border border-accent/30 bg-gradient-to-b from-accent/[0.07] to-white/[0.02] p-4 backdrop-blur-md">
      <header className="mb-2 flex items-center gap-3">
        {card.imageUrl && (
          <img
            src={card.imageUrl}
            alt={card.name}
            loading="lazy"
            className="h-14 w-14 shrink-0 rounded-xl border border-hairline object-cover"
            onError={(e) => {
              e.currentTarget.style.display = 'none'
            }}
          />
        )}
        <h3 className="min-w-0 flex-1 truncate text-base font-bold text-text">{card.name}</h3>
        {card.confidence !== 'high' && (
          <span
            className="shrink-0 rounded-full border border-amber-400/40 bg-amber-400/10 px-2 py-0.5 text-[0.55rem] font-semibold uppercase tracking-[0.12em] text-amber-300/90"
            title={t('ai.mayBeOutdatedHint')}
          >
            {t('ai.mayBeOutdated')}
          </span>
        )}
      </header>
      <dl className="space-y-1">
        {row(t('ai.card.fullName'), card.full_name)}
        {row(t('ai.card.country'), card.country)}
        {row(t('ai.card.club'), card.club)}
        {row(t('ai.card.position'), card.position)}
      </dl>
      {card.stats && card.stats.length > 0 && (
        <div className="mt-3 grid grid-cols-3 gap-2">
          {card.stats.slice(0, 6).map((s, i) => (
            <div
              key={i}
              className="rounded-xl border border-hairline bg-white/[0.03] px-2 py-1.5 text-center"
            >
              <div className="tabular-nums text-sm font-bold text-text">{s.value}</div>
              <div className="text-[0.6rem] leading-tight text-muted/70">{s.label}</div>
            </div>
          ))}
        </div>
      )}
      {card.achievements && card.achievements.length > 0 && (
        <ul className="mt-2 flex flex-wrap gap-1.5">
          {card.achievements.slice(0, 6).map((a, i) => (
            <li
              key={i}
              className="rounded-full border border-hairline bg-white/[0.04] px-2 py-0.5 text-[0.62rem] text-muted/85"
            >
              {a}
            </li>
          ))}
        </ul>
      )}
      <p className="mt-2.5 text-sm leading-relaxed text-text/90">{card.summary}</p>
    </article>
  )
}
