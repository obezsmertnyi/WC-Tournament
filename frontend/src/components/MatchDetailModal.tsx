import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type {
  Match,
  MatchDetail,
  DetailSide,
  DetailLineup,
  DetailGoal,
  DetailCard,
  DetailSubstitution,
  DetailOfficial,
} from '../types'
import { fetchMatchDetail } from '../lib/api'
import { teamName } from '../lib/teamNames'
import { formatKyivTime, statusLabel, stageLabel } from '../lib/fixtures'
import Flag from './Flag'
import MatchRevealPanel from './MatchRevealPanel'

type State =
  | { phase: 'loading' }
  | { phase: 'ready'; detail: MatchDetail }
  | { phase: 'error' }

interface MatchDetailModalProps {
  match: Match
  open: boolean
  onClose: () => void
}

/**
 * Drill-in for a single fixture: FIFA statistics (score, possession, goals,
 * cards, subs, lineups, officials, venue) plus the existing "who predicted
 * what" reveal — all in one premium dark + champagne-gold sheet. Rendered via
 * createPortal to escape transformed/`overflow-hidden` ancestors (the cards
 * live inside animated containers); centered on desktop, a full-height bottom
 * sheet on mobile. Closes on backdrop click or Escape.
 */
export default function MatchDetailModal({ match, open, onClose }: MatchDetailModalProps) {
  const { t, i18n } = useTranslation()
  const [state, setState] = useState<State>({ phase: 'loading' })
  const [reloadKey, setReloadKey] = useState(0)

  // Escape to close + lock body scroll while open.
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = prevOverflow
    }
  }, [open, onClose])

  // Load the detail when opened (and on retry). Aborts on close/unmount.
  useEffect(() => {
    if (!open) return
    const controller = new AbortController()
    setState({ phase: 'loading' })
    fetchMatchDetail(match.id, controller.signal)
      .then((detail) => {
        if (controller.signal.aborted) return
        setState({ phase: 'ready', detail })
      })
      .catch((err) => {
        if (controller.signal.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setState({ phase: 'error' })
      })
    return () => controller.abort()
  }, [open, match.id, reloadKey])

  const lang = i18n.resolvedLanguage
  const homeName = match.home
    ? teamName(match.home.code, match.home.name, lang)
    : (match.placeholderHome ?? t('fixture.tbd'))
  const awayName = match.away
    ? teamName(match.away.code, match.away.name, lang)
    : (match.placeholderAway ?? t('fixture.tbd'))

  const badge =
    match.stage === 'group'
      ? match.group
        ? t('calendar.groupNamed', { letter: match.group })
        : null
      : stageLabel(match.stage)

  return createPortal(
    <AnimatePresence>
      {open && (
        <motion.div
          className="fixed inset-0 z-[100] flex items-end justify-center sm:items-center sm:p-4"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
        >
          <button
            type="button"
            aria-label={t('auth.close')}
            onClick={onClose}
            className="absolute inset-0 bg-black/60 backdrop-blur-sm"
          />
          <motion.div
            role="dialog"
            aria-modal="true"
            aria-label={t('matchDetail.title')}
            initial={{ y: 28, opacity: 0, scale: 0.985 }}
            animate={{ y: 0, opacity: 1, scale: 1 }}
            exit={{ y: 28, opacity: 0, scale: 0.985 }}
            transition={{ type: 'spring', stiffness: 320, damping: 32 }}
            className="relative z-10 flex max-h-[92vh] w-full max-w-xl flex-col overflow-hidden rounded-t-3xl border border-hairline bg-gradient-to-b from-[#15171B] to-bg shadow-[0_-8px_40px_-12px_rgba(0,0,0,0.85)] sm:max-h-[90vh] sm:rounded-3xl sm:shadow-[0_24px_60px_-20px_rgba(0,0,0,0.9)]"
          >
            {/* Sticky header: eyebrow + close */}
            <header className="flex shrink-0 items-start justify-between gap-3 border-b border-hairline px-5 pb-3.5 pt-5">
              <div className="min-w-0">
                <p className="text-[0.6rem] font-semibold uppercase tracking-[0.24em] text-accent">
                  {t('matchDetail.title')}
                </p>
                <p className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[0.7rem] text-muted/80">
                  {badge && (
                    <span className="rounded-md border border-hairline bg-white/[0.03] px-1.5 py-0.5 text-[0.55rem] font-semibold uppercase tracking-[0.12em] text-muted/80">
                      {badge}
                    </span>
                  )}
                  <span className="tabular-nums">{formatKyivTime(match.kickoffAt)}</span>
                  <span>· {statusLabel(match.status)}</span>
                </p>
              </div>
              <button
                type="button"
                aria-label={t('auth.close')}
                onClick={onClose}
                className="-mr-1 -mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-hairline bg-white/[0.03] text-muted transition-colors hover:border-white/20 hover:text-text"
              >
                <CloseGlyph />
              </button>
            </header>

            {/* Scrollable body */}
            <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
              {state.phase === 'loading' && <DetailSkeleton />}

              {state.phase === 'error' && (
                <div className="rounded-2xl border border-red-500/20 bg-red-500/[0.04] px-4 py-8 text-center">
                  <p className="text-sm font-medium text-text">{t('matchDetail.errorTitle')}</p>
                  <p className="mt-1 text-xs text-muted">{t('matchDetail.errorBody')}</p>
                  <button
                    type="button"
                    onClick={() => setReloadKey((k) => k + 1)}
                    className="mt-4 rounded-lg border border-accent/50 bg-accent/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.14em] text-accent transition-colors hover:bg-accent/20"
                  >
                    {t('matchDetail.retry')}
                  </button>
                </div>
              )}

              {state.phase === 'ready' && (
                <DetailBody
                  match={match}
                  detail={state.detail}
                  homeName={homeName}
                  awayName={awayName}
                />
              )}
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>,
    document.body,
  )
}

interface DetailBodyProps {
  match: Match
  detail: MatchDetail
  homeName: string
  awayName: string
}

/** Flag identity for each side, passed to event rows so it's clear whose event it is. */
export interface SideFlags {
  home: { code?: string; flagUrl?: string; name: string }
  away: { code?: string; flagUrl?: string; name: string }
}

function DetailBody({ match, detail, homeName, awayName }: DetailBodyProps) {
  const { t } = useTranslation()
  const sides: SideFlags = {
    home: { code: match.home?.code, flagUrl: match.home?.flagUrl, name: homeName },
    away: { code: match.away?.code, flagUrl: match.away?.flagUrl, name: awayName },
  }

  if (!detail.available) {
    return (
      <>
        <ScoreHeader match={match} homeName={homeName} awayName={awayName} />
        <div className="mt-5 flex flex-col items-center rounded-2xl border border-hairline bg-white/[0.02] px-4 py-10 text-center">
          <StatsGlyph />
          <p className="mt-3 text-sm font-medium text-text">
            {t('matchDetail.notAvailable')}
          </p>
          <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted">
            {t('matchDetail.notAvailableBody')}
          </p>
        </div>
        <Section title={t('matchDetail.predictions')}>
          <MatchRevealPanel match={match} />
        </Section>
      </>
    )
  }

  // Winner highlighting is driven by the actual scoreline in ScoreHeader (the
  // Team model here carries no id to compare against `detail.winnerTeamId`, and
  // FIFA stats always agree with the scoreline).
  return (
    <>
      <ScoreHeader match={match} detail={detail} homeName={homeName} awayName={awayName} />

      {detail.possession && (
        <Section title={t('matchDetail.possession')}>
          <PossessionBar home={detail.possession.home} away={detail.possession.away} />
        </Section>
      )}

      {detail.goals.length > 0 && (
        <Section title={t('matchDetail.goals')}>
          <ul className="space-y-1.5">
            {detail.goals.map((g, i) => (
              <GoalRow key={i} goal={g} sides={sides} homeName={homeName} awayName={awayName} />
            ))}
          </ul>
        </Section>
      )}

      {detail.cards.length > 0 && (
        <Section title={t('matchDetail.cards')}>
          <ul className="space-y-1.5">
            {detail.cards.map((c, i) => (
              <CardRow key={i} card={c} sides={sides} homeName={homeName} awayName={awayName} />
            ))}
          </ul>
        </Section>
      )}

      {detail.substitutions.length > 0 && (
        <Section title={t('matchDetail.subs')}>
          <ul className="space-y-1.5">
            {detail.substitutions.map((s, i) => (
              <SubRow key={i} sub={s} sides={sides} homeName={homeName} awayName={awayName} />
            ))}
          </ul>
        </Section>
      )}

      {(detail.homeLineup || detail.awayLineup) && (
        <Section title={t('matchDetail.lineups')}>
          <div className="grid gap-4 sm:grid-cols-2">
            {detail.homeLineup && (
              <LineupColumn lineup={detail.homeLineup} fallbackName={homeName} />
            )}
            {detail.awayLineup && (
              <LineupColumn lineup={detail.awayLineup} fallbackName={awayName} />
            )}
          </div>
        </Section>
      )}

      {detail.officials.length > 0 && (
        <Section title={t('matchDetail.officials')}>
          <ul className="space-y-1.5">
            {detail.officials.map((o, i) => (
              <OfficialRow key={i} official={o} />
            ))}
          </ul>
        </Section>
      )}

      {(detail.attendance || detail.stadium) && (
        <Section title={t('matchDetail.venue')}>
          <dl className="space-y-2 text-sm">
            {detail.stadium && (
              <div className="flex items-center justify-between gap-3">
                <dt className="text-xs uppercase tracking-[0.12em] text-muted/80">
                  {t('matchDetail.stadium')}
                </dt>
                <dd className="min-w-0 truncate text-right font-medium text-text">
                  {detail.stadium}
                </dd>
              </div>
            )}
            {detail.attendance && (
              <div className="flex items-center justify-between gap-3">
                <dt className="text-xs uppercase tracking-[0.12em] text-muted/80">
                  {t('matchDetail.attendance')}
                </dt>
                <dd className="tabular-nums font-medium text-text">
                  {formatAttendance(detail.attendance)}
                </dd>
              </div>
            )}
          </dl>
        </Section>
      )}

      <Section title={t('matchDetail.predictions')}>
        <MatchRevealPanel match={match} />
      </Section>
    </>
  )
}

// ── Score header ─────────────────────────────────────────────────────────────

interface ScoreHeaderProps {
  match: Match
  detail?: Extract<MatchDetail, { available: true }>
  homeName: string
  awayName: string
}

function ScoreHeader({ match, detail, homeName, awayName }: ScoreHeaderProps) {
  const { t } = useTranslation()
  const homeScore = match.homeScore
  const awayScore = match.awayScore
  const hasScore = homeScore !== null && awayScore !== null
  const homeWins = hasScore && homeScore! > awayScore!
  const awayWins = hasScore && awayScore! > homeScore!

  const penalties =
    detail && detail.homePenaltyScore != null && detail.awayPenaltyScore != null
      ? `(${detail.homePenaltyScore}–${detail.awayPenaltyScore} ${t('matchDetail.penalties')})`
      : null
  const aggregate =
    detail && detail.aggregateHomeScore != null && detail.aggregateAwayScore != null
      ? t('matchDetail.aggregate', {
          home: detail.aggregateHomeScore,
          away: detail.aggregateAwayScore,
        })
      : null

  return (
    <div className="rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] p-4">
      <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-2">
        <TeamHead
          code={match.home?.code}
          flagUrl={match.home?.flagUrl}
          name={homeName}
          winner={homeWins}
        />
        <div className="flex flex-col items-center px-2">
          <div className="flex items-baseline gap-1.5 tabular-nums">
            <span
              className={`text-3xl font-bold leading-none tracking-tight ${homeWins ? 'text-accent' : 'text-text'}`}
            >
              {hasScore ? homeScore : '–'}
            </span>
            <span className="text-xl font-semibold text-muted/60">:</span>
            <span
              className={`text-3xl font-bold leading-none tracking-tight ${awayWins ? 'text-accent' : 'text-text'}`}
            >
              {hasScore ? awayScore : '–'}
            </span>
          </div>
          {detail?.matchTime && (
            <span className="mt-1 rounded-full border border-hairline bg-white/[0.03] px-2 py-0.5 text-[0.6rem] font-semibold tabular-nums tracking-wide text-muted">
              {detail.matchTime}
            </span>
          )}
          {penalties && (
            <span className="mt-1 text-[0.6rem] font-medium tabular-nums text-accent/90">
              {penalties}
            </span>
          )}
        </div>
        <TeamHead
          code={match.away?.code}
          flagUrl={match.away?.flagUrl}
          name={awayName}
          winner={awayWins}
        />
      </div>
      {aggregate && (
        <p className="mt-3 border-t border-hairline pt-2 text-center text-[0.65rem] uppercase tracking-[0.12em] text-muted/70">
          {aggregate}
        </p>
      )}
    </div>
  )
}

function TeamHead({
  code,
  flagUrl,
  name,
  winner,
}: {
  code?: string
  flagUrl?: string
  name: string
  winner: boolean
}) {
  return (
    <div className="flex min-w-0 flex-col items-center gap-1.5">
      <Flag code={code} flagUrl={flagUrl} label={name} className="h-7 w-10" />
      <span
        className={`line-clamp-2 text-center text-[0.8rem] leading-tight ${winner ? 'font-semibold text-text' : 'text-muted'}`}
      >
        {name}
      </span>
    </div>
  )
}

// ── Possession ───────────────────────────────────────────────────────────────

function PossessionBar({ home, away }: { home: number; away: number }) {
  const total = home + away || 100
  const homePct = Math.max(0, Math.min(100, (home / total) * 100))
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between text-xs font-semibold tabular-nums">
        <span className="text-accent">{home}%</span>
        <span className="text-muted">{away}%</span>
      </div>
      <div className="flex h-2.5 overflow-hidden rounded-full bg-white/[0.06]">
        <motion.div
          className="h-full rounded-l-full bg-accent"
          initial={{ width: 0 }}
          animate={{ width: `${homePct}%` }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
        />
      </div>
    </div>
  )
}

// ── Event rows ───────────────────────────────────────────────────────────────

function sideName(team: DetailSide, homeName: string, awayName: string): string {
  return team === 'home' ? homeName : awayName
}

function Minute({ value }: { value: string }) {
  return (
    <span className="w-9 shrink-0 text-right text-xs font-semibold tabular-nums text-accent">
      {value}
      {/\d$/.test(value) ? "'" : ''}
    </span>
  )
}

function EventRow({
  children,
  team,
  sides,
}: {
  children: React.ReactNode
  team: DetailSide
  sides: SideFlags
}) {
  const s = team === 'home' ? sides.home : sides.away
  return (
    <li className="flex items-center gap-3 rounded-xl border border-hairline bg-white/[0.025] px-3 py-2">
      {children}
      {/* Whose event it is — the team's flag + code (clear at a glance) */}
      <span className="ml-auto flex shrink-0 items-center gap-1.5" title={s.name}>
        <Flag code={s.code} flagUrl={s.flagUrl} label={s.name} className="h-[0.85rem] w-[1.25rem]" />
        <span className="text-[0.6rem] font-semibold uppercase tracking-[0.08em] text-muted/75">
          {s.code ?? (team === 'home' ? 'H' : 'A')}
        </span>
      </span>
    </li>
  )
}

function GoalRow({
  goal,
  sides,
  homeName,
  awayName,
}: {
  goal: DetailGoal
  sides: SideFlags
  homeName: string
  awayName: string
}) {
  return (
    <EventRow team={goal.team} sides={sides}>
      <Minute value={goal.minute} />
      <BallGlyph />
      <span className="min-w-0 flex-1 truncate text-sm text-text">
        <span className="font-medium">{goal.scorer}</span>
        {goal.assist && <span className="text-muted"> ({goal.assist})</span>}
        {typeof goal.type === 'string' && goal.type.toLowerCase() !== 'goal' && goal.type !== '' && (
          <span className="text-muted/70"> · {goal.type}</span>
        )}
        <span className="sr-only">{sideName(goal.team, homeName, awayName)}</span>
      </span>
    </EventRow>
  )
}

function CardRow({
  card,
  sides,
  homeName,
  awayName,
}: {
  card: DetailCard
  sides: SideFlags
  homeName: string
  awayName: string
}) {
  // FIFA `card` is numeric (1=yellow, 2/3=red) but may be a string elsewhere.
  const red =
    typeof card.card === 'string'
      ? card.card.toLowerCase().includes('red')
      : Number(card.card) >= 2
  return (
    <EventRow team={card.team} sides={sides}>
      <Minute value={card.minute} />
      <span
        className={`h-3.5 w-2.5 shrink-0 rounded-[2px] shadow-sm ${red ? 'bg-red-500' : 'bg-yellow-400'}`}
        aria-hidden="true"
      />
      <span className="min-w-0 flex-1 truncate text-sm font-medium text-text">
        {card.player}
        <span className="sr-only">{sideName(card.team, homeName, awayName)}</span>
      </span>
    </EventRow>
  )
}

function SubRow({
  sub,
  sides,
  homeName,
  awayName,
}: {
  sub: DetailSubstitution
  sides: SideFlags
  homeName: string
  awayName: string
}) {
  return (
    <EventRow team={sub.team} sides={sides}>
      <Minute value={sub.minute} />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5 text-sm">
        <span className="flex items-center gap-1.5 truncate text-emerald-400/90">
          <ArrowGlyph dir="up" />
          <span className="truncate text-text">{sub.playerIn}</span>
        </span>
        <span className="flex items-center gap-1.5 truncate text-muted">
          <ArrowGlyph dir="down" />
          <span className="truncate">{sub.playerOut}</span>
        </span>
      </div>
      <span className="sr-only">{sideName(sub.team, homeName, awayName)}</span>
    </EventRow>
  )
}

function OfficialRow({ official }: { official: DetailOfficial }) {
  return (
    <li className="flex items-center justify-between gap-3 rounded-xl border border-hairline bg-white/[0.025] px-3 py-2 text-sm">
      <span className="min-w-0 truncate font-medium text-text">{official.name}</span>
      <span className="shrink-0 text-xs uppercase tracking-[0.1em] text-muted/80">
        {official.type}
      </span>
    </li>
  )
}

// ── Lineups ──────────────────────────────────────────────────────────────────

function LineupColumn({
  lineup,
  fallbackName,
}: {
  lineup: DetailLineup
  fallbackName: string
}) {
  const { t } = useTranslation()
  return (
    <div className="rounded-2xl border border-hairline bg-white/[0.02] p-3">
      <header className="mb-2.5 flex items-baseline justify-between gap-2 border-b border-hairline pb-2">
        <span className="min-w-0 truncate text-sm font-semibold text-text">
          {lineup.teamName || fallbackName}
        </span>
        {lineup.formation && (
          <span className="shrink-0 rounded-md border border-accent/40 bg-accent/10 px-1.5 py-0.5 text-[0.6rem] font-semibold tabular-nums tracking-wide text-accent">
            {lineup.formation}
          </span>
        )}
      </header>
      <span className="sr-only">{t('matchDetail.formation')}</span>
      <ul className="space-y-1">
        {lineup.players.map((p, i) => (
          <li key={i} className="flex items-center gap-2 text-sm">
            <span className="w-5 shrink-0 text-right text-xs font-semibold tabular-nums text-muted/80">
              {p.shirtNumber}
            </span>
            <span className="min-w-0 flex-1 truncate text-text">
              {p.name}
              {p.captain && (
                <span
                  className="ml-1 text-[0.6rem] font-bold text-accent"
                  title="Captain"
                  aria-label="Captain"
                >
                  ©
                </span>
              )}
            </span>
            {p.position && (
              <span className="shrink-0 text-[0.6rem] font-medium uppercase tracking-wide text-muted/70">
                {p.position}
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}

// ── Layout helpers ───────────────────────────────────────────────────────────

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mt-5">
      <h3 className="mb-2.5 text-[0.65rem] font-semibold uppercase tracking-[0.18em] text-muted/80">
        {title}
      </h3>
      {children}
    </section>
  )
}

function formatAttendance(raw: string): string {
  const n = Number(raw)
  if (!Number.isFinite(n) || n <= 0) return raw
  return new Intl.NumberFormat().format(n)
}

function DetailSkeleton() {
  return (
    <div className="space-y-5">
      <div className="h-28 animate-pulse rounded-2xl bg-white/[0.05]" />
      <div className="h-12 animate-pulse rounded-2xl bg-white/[0.04]" />
      <div className="grid grid-cols-1 gap-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-11 animate-pulse rounded-xl bg-white/[0.04]" />
        ))}
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="h-48 animate-pulse rounded-2xl bg-white/[0.04]" />
        <div className="h-48 animate-pulse rounded-2xl bg-white/[0.04]" />
      </div>
    </div>
  )
}

// ── Glyphs ───────────────────────────────────────────────────────────────────

function CloseGlyph() {
  return (
    <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" aria-hidden="true">
      <path
        d="M6 6l12 12M18 6L6 18"
        stroke="currentColor"
        strokeWidth="1.7"
        strokeLinecap="round"
      />
    </svg>
  )
}

function BallGlyph() {
  return (
    <svg viewBox="0 0 24 24" className="h-3.5 w-3.5 shrink-0 text-text/70" fill="none" aria-hidden="true">
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.5" />
      <path d="M12 7l3 2-1 3.5h-4L9 9z" fill="currentColor" />
    </svg>
  )
}

function ArrowGlyph({ dir }: { dir: 'up' | 'down' }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className="h-3 w-3 shrink-0"
      fill="none"
      aria-hidden="true"
    >
      {dir === 'up' ? (
        <path d="M12 19V5M6 11l6-6 6 6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      ) : (
        <path d="M12 5v14M6 13l6 6 6-6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      )}
    </svg>
  )
}

function StatsGlyph() {
  return (
    <span className="flex h-11 w-11 items-center justify-center rounded-full border border-hairline bg-white/[0.03]">
      <svg viewBox="0 0 24 24" className="h-5 w-5 text-accent" fill="none" aria-hidden="true">
        <path d="M5 19V11M12 19V5M19 19v-6" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
      </svg>
    </span>
  )
}
