import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { GroupStanding } from '../types'
import { useMountAnimation } from '../lib/motion'
import Flag from './Flag'

interface GroupCardProps {
  standing: GroupStanding
  index?: number
}

/** Compact group card for the Groups index — header + mini standings. */
export default function GroupCard({ standing, index = 0 }: GroupCardProps) {
  const { t } = useTranslation()
  const rows = standing.rows.slice(0, 4)
  const mount = useMountAnimation(14, Math.min(index, 12) * 0.03)

  return (
    <motion.div
      initial={mount.initial}
      animate={mount.animate}
      transition={mount.transition}
      whileHover={{ y: -3 }}
    >
      <Link
        to={`/groups/${standing.group}`}
        className="group block overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.055] to-white/[0.015] p-4 shadow-[0_8px_24px_-16px_rgba(0,0,0,0.8)] backdrop-blur-md transition-colors hover:border-accent/30 hover:from-white/[0.08]"
      >
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-sm font-semibold uppercase tracking-[0.14em] text-text">
            {t('calendar.groupNamed', { letter: standing.group })}
          </h3>
          <span className="text-muted/50 transition-colors group-hover:text-accent">
            <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden>
              <path d="M9 18l6-6-6-6" />
            </svg>
          </span>
        </div>

        <ul className="space-y-1.5">
          {rows.map((r) => (
            <li key={r.teamId} className="flex items-center gap-2.5 text-sm">
              <span className="w-3 shrink-0 text-center text-[0.7rem] font-medium tabular-nums text-muted/60">
                {r.rank}
              </span>
              <Flag code={r.code} flagUrl={r.flagUrl} label={r.name} className="h-[0.95rem] w-5" />
              <span className="min-w-0 flex-1 truncate font-medium text-text/90">{r.code}</span>
              <span className="tabular-nums text-xs font-bold text-accent">{r.points}</span>
            </li>
          ))}
          {rows.length === 0 && (
            <li className="py-2 text-center text-xs text-muted/60">{t('groups.empty')}</li>
          )}
        </ul>
      </Link>
    </motion.div>
  )
}
