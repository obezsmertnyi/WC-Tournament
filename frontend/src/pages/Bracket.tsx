import { useCallback, useEffect, useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { Match, Stage, Team } from '../types'
import { fetchMatches } from '../lib/api'
import { groupFixtures, stageLabel, stageShortLabel } from '../lib/fixtures'
import { teamName } from '../lib/teamNames'
import { useMountAnimation } from '../lib/motion'
import BracketTie, { tieWinner } from '../components/BracketTie'
import Flag from '../components/Flag'
import Trophy from '../components/Trophy'
import StarHero from '../components/StarHero'
import { ErrorState, FixturesSkeleton } from '../components/states'

type LoadState =
  | { phase: 'loading' }
  | { phase: 'error' }
  | { phase: 'ready'; matches: Match[] }

/** Main-line knockout rounds, left → right. `third` is rendered separately. */
const TREE_ORDER: Stage[] = ['r32', 'r16', 'qf', 'sf', 'final']

interface KnockoutData {
  /** Rounds present in the dataset, in tree order, each with its ties. */
  rounds: { stage: Stage; matches: Match[] }[]
  /** The single final match, if present. */
  final: Match | null
  /** The third-place play-off, if present. */
  third: Match | null
  /** Champion (winner of the final), or null until decided. */
  champion: Team
}

function buildKnockout(matches: Match[]): KnockoutData {
  const { knockout } = groupFixtures(matches)
  const byStage = new Map<Stage, Match[]>()
  for (const section of knockout) byStage.set(section.key, section.matches)

  const rounds = TREE_ORDER.filter((s) => byStage.has(s)).map((stage) => ({
    stage,
    matches: byStage.get(stage) ?? [],
  }))
  const final = byStage.get('final')?.[0] ?? null
  const third = byStage.get('third')?.[0] ?? null
  const champion = final ? tieWinner(final) : null
  return { rounds, final, third, champion }
}

export default function Bracket() {
  const { t } = useTranslation()
  const [state, setState] = useState<LoadState>({ phase: 'loading' })

  const load = useCallback((signal?: AbortSignal) => {
    setState({ phase: 'loading' })
    fetchMatches(signal)
      .then((matches) => setState({ phase: 'ready', matches }))
      .catch((err) => {
        if (signal?.aborted) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setState({ phase: 'error' })
      })
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    return () => controller.abort()
  }, [load])

  return (
    <div className="mx-auto w-full max-w-6xl">
      <header className="relative mb-6 -mx-4 overflow-hidden rounded-b-3xl px-4 pb-6 pt-4 sm:-mx-6 sm:px-6">
        <StarHero variant="band" />
        <div className="relative flex items-center gap-3 sm:gap-4">
          <Trophy className="h-12 w-12 rounded-xl sm:h-16 sm:w-16" />
          <div className="min-w-0">
            <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">
              {t('bracket.title')}
            </h1>
            <p className="mt-1 text-sm text-muted">{t('bracket.subtitle')}</p>
          </div>
        </div>
      </header>

      {state.phase === 'loading' && <FixturesSkeleton />}
      {state.phase === 'error' && <ErrorState onRetry={() => load()} />}
      {state.phase === 'ready' && <BracketView matches={state.matches} />}
    </div>
  )
}

function BracketView({ matches }: { matches: Match[] }) {
  const { t } = useTranslation()
  const data = useMemo(() => buildKnockout(matches), [matches])

  if (data.rounds.length === 0) {
    return (
      <div className="rounded-2xl border border-hairline bg-surface px-6 py-16 text-center backdrop-blur-md">
        <p className="text-sm text-muted">{t('bracket.empty')}</p>
      </div>
    )
  }

  return (
    <>
      {/* Mobile (< lg): round selector + single stacked column */}
      <div className="lg:hidden">
        <MobileBracket data={data} />
      </div>

      {/* Desktop (>= lg): full horizontally-scrollable tree */}
      <div className="hidden lg:block">
        <DesktopTree data={data} />
      </div>
    </>
  )
}

// ── Desktop tree ─────────────────────────────────────────────────────────────

/**
 * The full knockout tree. Each round is a flex column whose ties are spread
 * with `justify-around`, so a tie in round N sits vertically centred against
 * the pair that feeds it in round N+1 — the classic bracket geometry. Connector
 * elbows are drawn with CSS borders in a thin column between rounds, so the
 * tree reads as a bracket and renders identically in headless captures (no
 * runtime measurement needed). The champion + third-place sit beside the final.
 */
function DesktopTree({ data }: { data: KnockoutData }) {
  const { t } = useTranslation()
  return (
    <div className="overflow-x-auto pb-4 [-ms-overflow-style:none] [scrollbar-width:thin] [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-white/10 [&::-webkit-scrollbar]:h-1.5">
      <div className="flex min-w-max items-stretch gap-0">
        {data.rounds.map((round, ri) => {
          const isLast = ri === data.rounds.length - 1
          return (
            <div key={round.stage} className="flex items-stretch">
              <div className="flex w-[14.5rem] flex-col">
                <RoundHeader stage={round.stage} count={round.matches.length} />
                <div className="flex flex-1 flex-col justify-around gap-3 py-1">
                  {round.matches.map((m, i) => (
                    <BracketTie
                      key={m.id}
                      match={m}
                      index={ri * 4 + i}
                      emphasis={round.stage === 'final'}
                    />
                  ))}
                </div>
              </div>
              {/* Connector column between this round and the next */}
              {!isLast && <Connectors count={round.matches.length} />}
            </div>
          )
        })}

        {/* Champion + third place rail beside the final */}
        <div className="flex w-[14.5rem] flex-col justify-around pl-0">
          <div className="space-y-4 self-center">
            <ChampionCard champion={data.champion} />
            {data.third && (
              <div>
                <p className="mb-1.5 px-0.5 text-[0.6rem] font-semibold uppercase tracking-[0.16em] text-muted/70">
                  {t('bracket.thirdPlace')}
                </p>
                <BracketTie match={data.third} index={20} />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function RoundHeader({ stage, count }: { stage: Stage; count: number }) {
  return (
    <div className="mb-2 flex items-baseline justify-between gap-2 px-0.5">
      <h2 className="text-[0.7rem] font-bold uppercase tracking-[0.16em] text-accent">
        {stageLabel(stage)}
      </h2>
      <span className="text-[0.6rem] font-medium tabular-nums text-muted/60">{count}</span>
    </div>
  )
}

/**
 * Connector elbows for one round → next. For every pair of ties we draw a
 * horizontal stub out of each tie, a vertical link joining the pair, and a
 * horizontal stub into the merged tie of the next round. Pure CSS borders, so
 * it's crisp and animation-free.
 */
function Connectors({ count }: { count: number }) {
  const pairs = Math.floor(count / 2)
  const hasOdd = count % 2 === 1
  return (
    <div className="flex w-7 flex-col justify-around py-1">
      {Array.from({ length: pairs }).map((_, i) => (
        <div key={i} className="flex flex-1 items-stretch">
          {/* left stubs + vertical join */}
          <div className="flex w-1/2 flex-col">
            <span className="flex-1 border-b border-r border-hairline" />
            <span className="flex-1 border-t border-r border-hairline" />
          </div>
          {/* stub into next round, centred */}
          <span className="my-auto block h-px w-1/2 bg-hairline" />
        </div>
      ))}
      {hasOdd && <div className="flex-1" />}
    </div>
  )
}

function ChampionCard({ champion }: { champion: Team }) {
  const { t, i18n } = useTranslation()
  const name = champion
    ? teamName(champion.code, champion.name, i18n.resolvedLanguage)
    : t('bracket.championTbd')
  return (
    <div className="rounded-2xl border border-accent/40 bg-gradient-to-b from-accent/[0.1] to-white/[0.02] p-4 text-center shadow-[0_10px_36px_-16px_rgba(201,162,75,0.6)]">
      <p className="text-[0.6rem] font-semibold uppercase tracking-[0.2em] text-accent">
        {t('bracket.champion')}
      </p>
      <Trophy className="mx-auto my-3 h-12 w-12 rounded-xl" />
      {champion ? (
        <div className="flex items-center justify-center gap-2">
          <Flag
            code={champion.code}
            flagUrl={champion.flagUrl}
            label={name}
            className="h-[1rem] w-[1.5rem]"
          />
          <span className="truncate text-sm font-bold text-text">{name}</span>
        </div>
      ) : (
        <p className="text-xs italic text-muted">{name}</p>
      )}
    </div>
  )
}

// ── Mobile: round chips + stacked ties ───────────────────────────────────────

function MobileBracket({ data }: { data: KnockoutData }) {
  const { t } = useTranslation()
  const mount = useMountAnimation(8)

  // Chips in chronological order. The third-place play-off is played BEFORE the
  // final, so its chip sits right before "final" (not tacked on at the end).
  const chips: Stage[] = useMemo(() => {
    const main = data.rounds.map((r) => r.stage)
    if (!data.third) return main
    const fi = main.indexOf('final')
    return fi === -1 ? [...main, 'third'] : [...main.slice(0, fi), 'third', ...main.slice(fi)]
  }, [data])

  const matchesForStage = useCallback(
    (stage: Stage): Match[] =>
      stage === 'third'
        ? data.third
          ? [data.third]
          : []
        : (data.rounds.find((r) => r.stage === stage)?.matches ?? []),
    [data],
  )

  // Land on the round that's up next: the earliest round with an unfinished
  // match; if everything is played, the final.
  const defaultStage = useMemo<Stage>(() => {
    for (const s of chips) {
      const ms = matchesForStage(s)
      if (ms.length > 0 && ms.some((m) => m.status !== 'finished')) return s
    }
    return chips[chips.length - 1] ?? 'final'
  }, [chips, matchesForStage])

  const [selected, setSelected] = useState<Stage>(defaultStage)

  // Keep the user's manual choice; only re-default if it's no longer valid.
  useEffect(() => {
    setSelected((prev) => (chips.includes(prev) ? prev : defaultStage))
  }, [chips, defaultStage])

  const ties = matchesForStage(selected)

  const showChampion = selected === 'final'

  return (
    <div>
      {/* Round selector chips */}
      <div
        className="sticky top-14 z-20 -mx-4 mb-5 flex gap-2 overflow-x-auto border-b border-hairline bg-bg/70 px-4 py-3 backdrop-blur-xl [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden sm:-mx-6 sm:px-6"
        role="tablist"
        aria-label={t('bracket.selectRound')}
      >
        {chips.map((stage) => {
          const active = stage === selected
          return (
            <button
              key={stage}
              type="button"
              role="tab"
              aria-selected={active}
              onClick={() => setSelected(stage)}
              className={`relative shrink-0 rounded-full border px-3.5 py-1.5 text-xs font-semibold uppercase tracking-[0.12em] transition-colors ${
                active
                  ? 'border-accent/40 text-accent'
                  : 'border-hairline bg-white/[0.02] text-muted hover:border-white/15 hover:text-text'
              }`}
            >
              {active && (
                <motion.span
                  layoutId="bracketRoundActive"
                  className="absolute inset-0 -z-10 rounded-full bg-accent/[0.12] shadow-[0_0_18px_-4px_rgba(201,162,75,0.55)] ring-1 ring-accent/30"
                  transition={{ type: 'spring', stiffness: 380, damping: 32 }}
                />
              )}
              {stageShortLabel(stage)}
            </button>
          )
        })}
      </div>

      <motion.div
        key={selected}
        initial={mount.initial}
        animate={mount.animate}
        transition={mount.transition}
      >
        <div className="mb-3 flex items-baseline justify-between gap-3">
          <h2 className="text-base font-semibold tracking-tight text-text">
            {selected === 'third' ? t('bracket.thirdPlace') : stageLabel(selected)}
          </h2>
          <span className="text-xs font-medium uppercase tracking-[0.14em] text-muted/70">
            {ties.length}
          </span>
        </div>

        {showChampion && (
          <div className="mb-4">
            <ChampionCard champion={data.champion} />
          </div>
        )}

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {ties.map((m, i) => (
            <BracketTie key={m.id} match={m} index={i} emphasis={selected === 'final'} />
          ))}
        </div>
      </motion.div>
    </div>
  )
}
