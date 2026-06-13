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

  // Tournament-wide bonus picks. Each is awarded ONLY if correct; the value
  // depends on WHEN it was picked (group stage = more, after groups = less),
  // and picks lock at the Round of 16 (1/8) kickoff.
  const bonusRows: { label: string; group: number; post: number }[] = [
    { label: t('rules.bonusChampion'), group: 10, post: 6 },
    { label: t('rules.bonusFinalist'), group: 5, post: 3 },
    { label: t('rules.bonusTopScorer'), group: 5, post: 2 },
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

      {/* Tournament bonus picks — values by pick window */}
      <div className="mt-4 border-t border-hairline pt-3">
        <h3 className="text-sm font-semibold text-text">{t('rules.bonusTitle')}</h3>
        <div className="mt-2.5 overflow-hidden rounded-xl border border-hairline">
          <table className="w-full border-collapse text-sm">
            <thead>
              <tr className="border-b border-hairline text-[0.58rem] uppercase tracking-[0.1em] text-muted/70">
                <th className="py-2 pl-3 pr-2 text-left font-semibold">{t('rules.bonusPick')}</th>
                <th className="px-2 py-2 text-center font-semibold">{t('rules.tierGroupCol')}</th>
                <th className="px-2 py-2 pr-3 text-center font-semibold">{t('rules.tierPostCol')}</th>
              </tr>
            </thead>
            <tbody>
              {bonusRows.map((b) => (
                <tr key={b.label} className="border-b border-hairline/60 last:border-0">
                  <td className="py-2 pl-3 pr-2 text-muted">{b.label}</td>
                  <td className="px-2 py-2 text-center">
                    <span className="font-bold tabular-nums text-accent">{b.group}</span>
                  </td>
                  <td className="px-2 py-2 pr-3 text-center">
                    <span className="font-bold tabular-nums text-text/80">{b.post}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="mt-2.5 text-xs leading-relaxed text-muted/80">{t('rules.bonusNote')}</p>
      </div>
    </section>
  )
}
