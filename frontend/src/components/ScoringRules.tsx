import { useTranslation } from 'react-i18next'

/** A clear, always-visible explainer of how points are awarded. */
export default function ScoringRules() {
  const { t } = useTranslation()

  const rows: { pts: string; label: string }[] = [
    { pts: '3', label: t('rules.exact') },
    { pts: '1', label: t('rules.outcome') },
    { pts: '0', label: t('rules.wrong') },
    { pts: '+1', label: t('rules.knockout') },
  ]

  return (
    <section className="mb-5 rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] p-4 backdrop-blur-md">
      <p className="text-[0.6rem] font-semibold uppercase tracking-[0.28em] text-accent/80">
        {t('rules.eyebrow')}
      </p>
      <h2 className="mt-1 text-base font-semibold text-text">{t('rules.title')}</h2>

      <ul className="mt-3 space-y-2">
        {rows.map((r) => (
          <li key={r.label} className="flex items-start gap-3">
            <span className="mt-0.5 inline-flex h-6 min-w-[1.75rem] items-center justify-center rounded-md border border-accent/40 bg-accent/10 px-1.5 text-xs font-bold tabular-nums text-accent">
              {r.pts}
            </span>
            <span className="text-sm leading-snug text-muted">{r.label}</span>
          </li>
        ))}
      </ul>

      <p className="mt-3 border-t border-hairline pt-3 text-xs leading-relaxed text-muted/80">
        {t('rules.knockoutExample')}
      </p>
      <p className="mt-2 text-xs leading-relaxed text-muted/80">{t('rules.championBonus')}</p>
    </section>
  )
}
