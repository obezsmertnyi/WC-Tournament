import { useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { Match, Team } from '../types'
import { formatKyivTime, statusLabel, stageLabel, venueCaption } from '../lib/fixtures'
import { teamName } from '../lib/teamNames'
import { useMountAnimation } from '../lib/motion'
import { useAuth } from '../auth/AuthContext'
import Flag from './Flag'
import PredictionEditor from './PredictionEditor'
import MatchRevealPanel from './MatchRevealPanel'

function StatusChip({ status }: { status: Match['status'] }) {
  // Subscribe to language changes so labels re-localize.
  useTranslation()
  if (status === 'live') {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full border border-accent/45 bg-accent/10 px-2.5 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.16em] text-accent shadow-[0_0_12px_-2px_rgba(201,162,75,0.5)]">
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
    <span className="rounded-full border border-hairline bg-white/[0.02] px-2.5 py-1 text-[0.6rem] font-medium uppercase tracking-[0.16em] text-muted/75">
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
  const { t, i18n } = useTranslation()
  const showScore = (finished || live) && score !== null
  const name = team
    ? teamName(team.code, team.name, i18n.resolvedLanguage)
    : (placeholder ?? t('fixture.tbd'))
  const isTbd = !team

  return (
    <div className="flex items-center gap-3">
      <Flag
        code={team?.code}
        flagUrl={team?.flagUrl}
        label={name}
        className="h-[1.15rem] w-7"
      />
      <span
        className={`min-w-0 flex-1 truncate text-[0.975rem] leading-tight ${
          isTbd
            ? 'italic text-muted'
            : finished && !winner
              ? 'text-muted'
              : 'font-semibold text-text'
        }`}
      >
        {name}
      </span>
      {showScore && (
        <span
          className={`tabular-nums text-xl font-bold leading-none tracking-tight ${
            live
              ? 'animate-pulse text-accent'
              : winner
                ? 'text-text'
                : 'text-muted'
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
  /** Show a faint group/stage badge in the header (default true). */
  showBadge?: boolean
}

export default function FixtureCard({ match, index = 0, showBadge = true }: FixtureCardProps) {
  const { t } = useTranslation()
  const { status: authStatus } = useAuth()

  // Expandable "who predicted what" panel (reveal is kickoff-gated server-side).
  const [showReveal, setShowReveal] = useState(false)

  // Results come exclusively from the FIFA sync — never set manually here.
  const { home, away, venue, placeholderHome, placeholderAway } = match
  const homeScore = match.homeScore
  const awayScore = match.awayScore
  const status = match.status

  const live = status === 'live'
  const finished = status === 'finished'

  const homeWins = finished && homeScore !== null && awayScore !== null && homeScore > awayScore
  const awayWins = finished && homeScore !== null && awayScore !== null && awayScore > homeScore

  const badge =
    match.stage === 'group'
      ? match.group
        ? t('calendar.groupNamed', { letter: match.group })
        : null
      : stageLabel(match.stage)

  // Self-completing mount animation; collapses to the final state under
  // prefers-reduced-motion so cards always render (headless/screenshots).
  const mount = useMountAnimation(14, Math.min(index, 8) * 0.04)

  return (
    <motion.article
      initial={mount.initial}
      animate={mount.animate}
      transition={mount.transition}
      whileHover={{ y: -3 }}
      className={`group relative overflow-hidden rounded-2xl border p-4 backdrop-blur-md transition-colors ${
        live
          ? 'border-accent/30 bg-gradient-to-b from-accent/[0.06] to-white/[0.02] shadow-[0_8px_30px_-12px_rgba(201,162,75,0.45)]'
          : 'border-hairline bg-gradient-to-b from-white/[0.055] to-white/[0.015] shadow-[0_8px_24px_-16px_rgba(0,0,0,0.8)] hover:border-white/15 hover:from-white/[0.08]'
      }`}
    >
      {/* Header: kickoff time + status, with optional group/stage badge */}
      <header className="mb-3 flex items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <time className="text-xs font-semibold tabular-nums tracking-wide text-text/90">
            {formatKyivTime(match.kickoffAt)}
          </time>
          {showBadge && badge && (
            <span className="truncate rounded-md border border-hairline bg-white/[0.03] px-1.5 py-0.5 text-[0.55rem] font-semibold uppercase tracking-[0.12em] text-muted/80">
              {badge}
            </span>
          )}
        </div>
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
        <div className="h-px bg-gradient-to-r from-transparent via-hairline to-transparent" />
        <TeamRow
          team={away}
          placeholder={placeholderAway}
          score={awayScore}
          live={live}
          finished={finished}
          winner={awayWins}
        />
      </div>

      {/* Prediction entry / read-only pick — only for signed-in users with both
          real teams known (no TBD placeholders to predict against). Admins also
          get an "as [player]" selector here to edit any player's prediction. */}
      {authStatus === 'authenticated' && home && away && (
        <PredictionEditor match={match} />
      )}

      {/* Expand to see everyone's predictions (revealed after kickoff). */}
      {authStatus === 'authenticated' && home && away && (
        <div className="mt-3">
          <button
            type="button"
            onClick={() => setShowReveal((v) => !v)}
            aria-expanded={showReveal}
            className="flex w-full items-center justify-center gap-1.5 rounded-lg border border-hairline bg-white/[0.02] py-1.5 text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted/80 transition-colors hover:border-white/15 hover:text-text"
          >
            {t('fixture.whoPredicted')}
            <span className={`transition-transform ${showReveal ? 'rotate-180' : ''}`}>▾</span>
          </button>
          {showReveal && (
            <div className="mt-2">
              <MatchRevealPanel match={{ ...match, homeScore, awayScore, status }} />
            </div>
          )}
        </div>
      )}

      {/* Venue — host-country flag + caption */}
      <footer className="mt-3.5 flex items-center gap-1.5 border-t border-hairline pt-2.5">
        <Flag
          code={venue.country}
          flagUrl={undefined}
          label={venue.country}
          className="h-[0.7rem] w-[1.05rem]"
        />
        <p className="min-w-0 truncate text-[0.7rem] uppercase tracking-[0.1em] text-muted/70">
          {venueCaption(venue.city, venue.stadium)}
        </p>
      </footer>
    </motion.article>
  )
}
