import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { Match, Team } from '../types'
import { formatKyivTime, formatKyivDayMonth, statusLabel } from '../lib/fixtures'
import { teamName } from '../lib/teamNames'
import { useMountAnimation } from '../lib/motion'
import Flag from './Flag'

/** Tiny status dot/pill — live pulses gold, finished is muted "FT". */
function TieStatus({ status }: { status: Match['status'] }) {
  useTranslation() // re-localize on language change
  if (status === 'live') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full border border-accent/45 bg-accent/10 px-1.5 py-0.5 text-[0.5rem] font-semibold uppercase tracking-[0.14em] text-accent">
        <span className="relative flex h-1 w-1">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent/70" />
          <span className="relative inline-flex h-1 w-1 rounded-full bg-accent" />
        </span>
        {statusLabel('live')}
      </span>
    )
  }
  if (status === 'finished') {
    return (
      <span className="rounded-full border border-hairline bg-white/[0.03] px-1.5 py-0.5 text-[0.5rem] font-semibold uppercase tracking-[0.14em] text-muted">
        {statusLabel('finished')}
      </span>
    )
  }
  return (
    <span className="text-[0.6rem] font-semibold tabular-nums tracking-wide text-text/80" />
  )
}

interface SideProps {
  team: Team
  placeholder: string | null
  score: number | null
  show: boolean
  winner: boolean
  loser: boolean
}

function Side({ team, placeholder, score, show, winner, loser }: SideProps) {
  const { t, i18n } = useTranslation()
  const isTbd = !team
  const name = team
    ? teamName(team.code, team.name, i18n.resolvedLanguage)
    : (placeholder ?? t('fixture.tbd'))

  return (
    <div className="flex items-center gap-2">
      <Flag
        code={team?.code}
        flagUrl={team?.flagUrl}
        label={name}
        className="h-[0.95rem] w-[1.45rem]"
      />
      <span
        className={`min-w-0 flex-1 truncate text-[0.82rem] leading-tight ${
          isTbd
            ? 'italic text-muted/80'
            : winner
              ? 'font-semibold text-accent'
              : loser
                ? 'text-muted'
                : 'font-medium text-text'
        }`}
      >
        {name}
      </span>
      {show && score !== null && (
        <span
          className={`shrink-0 tabular-nums text-sm font-bold leading-none ${
            winner ? 'text-accent' : loser ? 'text-muted' : 'text-text'
          }`}
        >
          {score}
        </span>
      )}
    </div>
  )
}

interface BracketTieProps {
  match: Match
  index?: number
  /** Highlight as a feature tie (the final). */
  emphasis?: boolean
}

/**
 * One knockout tie rendered as a compact two-row card — the leaf node of the
 * bracket tree. Shows both sides (flag + short name + score), kickoff time and
 * a status pill. Before the draw it renders the placeholder text in muted
 * italic so empty slots still read sensibly.
 */
export default function BracketTie({ match, index = 0, emphasis = false }: BracketTieProps) {
  const { home, away, homeScore, awayScore, status, placeholderHome, placeholderAway } = match
  const live = status === 'live'
  const finished = status === 'finished'
  const show = live || finished

  const decided = finished && homeScore !== null && awayScore !== null
  const homeWins = decided && homeScore! > awayScore!
  const awayWins = decided && awayScore! > homeScore!

  const mount = useMountAnimation(10, Math.min(index, 10) * 0.03)

  return (
    <motion.article
      initial={mount.initial}
      animate={mount.animate}
      transition={mount.transition}
      className={`relative overflow-hidden rounded-xl border p-2.5 backdrop-blur-md ${
        live
          ? 'border-accent/35 bg-gradient-to-b from-accent/[0.07] to-white/[0.02] shadow-[0_6px_22px_-12px_rgba(201,162,75,0.5)]'
          : emphasis
            ? 'border-accent/30 bg-gradient-to-b from-accent/[0.05] to-white/[0.015] shadow-[0_8px_28px_-14px_rgba(201,162,75,0.45)]'
            : 'border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] shadow-[0_6px_20px_-14px_rgba(0,0,0,0.8)]'
      }`}
    >
      <header className="mb-1.5 flex items-center justify-between gap-2">
        <time className="text-[0.6rem] font-semibold tabular-nums tracking-wide text-text/70">
          {match.kickoffAt ? `${formatKyivDayMonth(match.kickoffAt)} · ${formatKyivTime(match.kickoffAt)}` : ''}
        </time>
        <TieStatus status={status} />
      </header>
      <div className="space-y-1.5">
        <Side
          team={home}
          placeholder={placeholderHome}
          score={homeScore}
          show={show}
          winner={homeWins}
          loser={awayWins}
        />
        <Side
          team={away}
          placeholder={placeholderAway}
          score={awayScore}
          show={show}
          winner={awayWins}
          loser={homeWins}
        />
      </div>
    </motion.article>
  )
}

/** Resolve the winning Team of a finished tie, or null if undecided. */
export function tieWinner(match: Match): Team {
  if (match.status !== 'finished') return null
  const { homeScore: h, awayScore: a, home, away } = match
  if (h === null || a === null) return null
  if (h > a) return home
  if (a > h) return away
  return null
}
