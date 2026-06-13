import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { Match, Team } from '../types'
import { formatKyivTime, statusLabel } from '../lib/fixtures'
import Flag from './Flag'

function StatusChip({ status }: { status: Match['status'] }) {
  useTranslation()
  if (status === 'live') {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full border border-accent/40 bg-accent/10 px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.16em] text-accent">
        <span className="relative flex h-1.5 w-1.5">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent/70" />
          <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-accent" />
        </span>
        {statusLabel('live')}
      </span>
    )
  }

  if (status === 'finished') {
    return (
      <span className="rounded-full border border-hairline bg-white/[0.03] px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.16em] text-muted">
        {statusLabel('finished')}
      </span>
    )
  }

  return (
    <span className="rounded-full border border-hairline bg-white/[0.03] px-2.5 py-1 text-[0.6rem] font-medium uppercase tracking-[0.16em] text-muted/80">
      {statusLabel('scheduled')}
    </span>
  )
}

interface TeamRowProps {
  team: Team
  placeholder: string | null
  score: number | null
  live: boolean
  finished: boolean
  /** True for the side with the higher score (only when finished). */
  winner: boolean
}

function TeamRow({ team, placeholder, score, live, finished, winner }: TeamRowProps) {
  const { t } = useTranslation()
  const showScore = (finished || live) && score !== null
  const name = team?.name ?? placeholder ?? t('fixture.tbd')
  const isTbd = !team

  return (
    <div className="flex items-center gap-3">
      <Flag
        code={team?.code}
        flagUrl={team?.flagUrl}
        label={name}
        className="h-4 w-6 shrink-0"
      />
      <span
        className={`min-w-0 flex-1 truncate text-[0.95rem] ${
          isTbd
            ? 'italic text-muted'
            : finished && !winner
              ? 'text-muted'
              : 'font-medium text-text'
        }`}
      >
        {name}
      </span>
      {showScore && (
        <span
          className={`tabular-nums text-lg font-semibold tracking-tight ${
            live ? 'text-accent' : winner ? 'text-text' : 'text-muted'
          }`}
        >
          {score}
        </span>
      )}
    </div>
  )
}

interface FixtureCardProps {
  match: Match
  index?: number
}

export default function FixtureCard({ match, index = 0 }: FixtureCardProps) {
  const { home, away, homeScore, awayScore, status, venue, placeholderHome, placeholderAway } =
    match

  const live = status === 'live'
  const finished = status === 'finished'

  const homeWins = finished && homeScore !== null && awayScore !== null && homeScore > awayScore
  const awayWins = finished && homeScore !== null && awayScore !== null && awayScore > homeScore

  return (
    <motion.article
      initial={{ opacity: 0, y: 14 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: '-40px' }}
      transition={{ duration: 0.45, delay: Math.min(index, 8) * 0.04, ease: [0.22, 1, 0.36, 1] }}
      whileHover={{ y: -3 }}
      className={`group rounded-2xl border bg-surface p-4 backdrop-blur-md transition-colors ${
        live
          ? 'border-accent/30 shadow-[0_0_22px_-6px_rgba(201,162,75,0.45)]'
          : 'border-hairline hover:border-white/15'
      }`}
    >
      {/* Header: kickoff time + status */}
      <header className="mb-3 flex items-center justify-between">
        <time className="text-xs font-medium tabular-nums tracking-wide text-muted">
          {formatKyivTime(match.kickoffAt)}
        </time>
        <StatusChip status={status} />
      </header>

      {/* Teams */}
      <div className="space-y-2.5">
        <TeamRow
          team={home}
          placeholder={placeholderHome}
          score={homeScore}
          live={live}
          finished={finished}
          winner={homeWins}
        />
        <TeamRow
          team={away}
          placeholder={placeholderAway}
          score={awayScore}
          live={live}
          finished={finished}
          winner={awayWins}
        />
      </div>

      {/* Venue */}
      <footer className="mt-3.5 border-t border-hairline pt-2.5">
        <p className="truncate text-[0.7rem] uppercase tracking-[0.1em] text-muted/70">
          {venue.city} · {venue.stadium}
        </p>
      </footer>
    </motion.article>
  )
}
