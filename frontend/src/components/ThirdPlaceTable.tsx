import { useTranslation } from 'react-i18next'
import type { ThirdPlaceRow } from '../types'
import { teamName } from '../lib/teamNames'
import Flag from './Flag'

interface ThirdPlaceTableProps {
  rows: ThirdPlaceRow[]
}

/**
 * Cross-group ranking of the third-placed teams. Under the 48-team WC2026
 * format the top 8 advance to the Round of 32 — those rows get a green
 * qualifying badge; the rest are dimmed. Tie-break order mirrors FIFA:
 * points → goal difference → goals scored (then a stable fallback).
 */
export default function ThirdPlaceTable({ rows }: ThirdPlaceTableProps) {
  const { t, i18n } = useTranslation()
  if (rows.length === 0) return null

  const cols: { key: string; label: string }[] = [
    { key: 'p', label: t('table.p') },
    { key: 'w', label: t('table.w') },
    { key: 'd', label: t('table.d') },
    { key: 'l', label: t('table.l') },
    { key: 'gd', label: t('table.gd') },
  ]

  return (
    <section className="mb-8">
      <div className="mb-3 flex items-baseline justify-between gap-3">
        <h2 className="text-base font-semibold text-text sm:text-lg">{t('thirdPlace.title')}</h2>
        <span className="text-[0.7rem] uppercase tracking-[0.12em] text-muted/70">
          {t('thirdPlace.qualifyHint')}
        </span>
      </div>

      <div className="overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] shadow-[0_8px_24px_-16px_rgba(0,0,0,0.8)] backdrop-blur-md">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b border-hairline text-[0.6rem] uppercase tracking-[0.12em] text-muted/70">
              <th className="py-2.5 pl-3 pr-1 text-left font-semibold sm:pl-4">#</th>
              <th className="py-2.5 pr-2 text-left font-semibold">{t('table.team')}</th>
              <th className="px-1.5 py-2.5 text-center font-semibold sm:px-2">{t('thirdPlace.group')}</th>
              {cols.map((c) => (
                <th key={c.key} className="px-1.5 py-2.5 text-center font-semibold tabular-nums sm:px-2">
                  {c.label}
                </th>
              ))}
              <th className="px-2 py-2.5 pr-3 text-center font-semibold text-accent/80 sm:pr-4">
                {t('table.pts')}
              </th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => (
              <tr
                key={r.teamId}
                className={`border-b border-hairline/60 transition-colors last:border-0 hover:bg-white/[0.03] ${
                  r.qualified ? '' : 'opacity-55'
                }`}
              >
                <td className="py-2.5 pl-3 pr-1 sm:pl-4">
                  <span
                    className={`inline-flex h-6 w-6 items-center justify-center rounded-full text-[0.7rem] font-bold tabular-nums ${
                      r.qualified
                        ? 'bg-emerald-500/20 text-emerald-300 ring-1 ring-emerald-400/40'
                        : 'text-muted'
                    }`}
                  >
                    {r.rank}
                  </span>
                </td>
                <td className="min-w-0 py-2.5 pr-2">
                  <div className="flex items-center gap-2.5">
                    <Flag code={r.code} flagUrl={r.flagUrl} label={r.name} className="h-[1.05rem] w-6" />
                    <span className="truncate font-medium text-text">
                      <span className="sm:hidden">{r.code}</span>
                      <span className="hidden sm:inline">
                        {teamName(r.code, r.name, i18n.resolvedLanguage)}
                      </span>
                    </span>
                  </div>
                </td>
                <td className="px-1.5 py-2.5 text-center font-medium tabular-nums text-muted sm:px-2">
                  {r.group}
                </td>
                <td className="px-1.5 py-2.5 text-center tabular-nums text-muted sm:px-2">{r.played}</td>
                <td className="px-1.5 py-2.5 text-center tabular-nums text-muted sm:px-2">{r.win}</td>
                <td className="px-1.5 py-2.5 text-center tabular-nums text-muted sm:px-2">{r.draw}</td>
                <td className="px-1.5 py-2.5 text-center tabular-nums text-muted sm:px-2">{r.loss}</td>
                <td className="px-1.5 py-2.5 text-center tabular-nums text-muted sm:px-2">
                  {r.gd > 0 ? `+${r.gd}` : r.gd}
                </td>
                <td className="px-2 py-2.5 pr-3 text-center font-bold tabular-nums text-accent sm:pr-4">
                  {r.points}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
