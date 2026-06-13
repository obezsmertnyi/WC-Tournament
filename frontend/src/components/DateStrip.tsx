import { useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import type { MatchDay } from '../lib/fixtures'
import { formatKyivWeekday, formatKyivDayMonth, todayKyivKey } from '../lib/fixtures'

interface DateStripProps {
  days: MatchDay[]
  selected: string
  onSelect: (key: string) => void
}

function ArrowButton({
  dir,
  disabled,
  onClick,
  label,
}: {
  dir: 'prev' | 'next'
  disabled: boolean
  onClick: () => void
  label: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
      className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-hairline bg-white/[0.03] text-muted transition-colors hover:border-accent/40 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:border-hairline disabled:hover:text-muted"
    >
      <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden>
        {dir === 'prev' ? <path d="M15 18l-6-6 6-6" /> : <path d="M9 18l6-6-6-6" />}
      </svg>
    </button>
  )
}

export default function DateStrip({ days, selected, onSelect }: DateStripProps) {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const today = todayKyivKey()

  const index = days.findIndex((d) => d.key === selected)
  const hasToday = days.some((d) => d.key === today)

  // Keep the selected chip scrolled into view.
  useEffect(() => {
    const el = scrollRef.current?.querySelector<HTMLElement>(`[data-day="${selected}"]`)
    el?.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' })
  }, [selected])

  const goPrev = () => index > 0 && onSelect(days[index - 1].key)
  const goNext = () => index < days.length - 1 && onSelect(days[index + 1].key)

  return (
    <div className="sticky top-14 z-20 -mx-4 mb-6 border-b border-hairline bg-bg/70 px-4 py-3 backdrop-blur-xl sm:-mx-6 sm:px-6">
      <div className="mx-auto flex max-w-5xl items-center gap-2">
        <ArrowButton dir="prev" disabled={index <= 0} onClick={goPrev} label={t('dateStrip.prev')} />

        <div
          ref={scrollRef}
          className="flex min-w-0 flex-1 gap-2 overflow-x-auto scroll-smooth [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
        >
          {days.map((day) => {
            const active = day.key === selected
            const isToday = day.key === today
            return (
              <button
                key={day.key}
                type="button"
                data-day={day.key}
                onClick={() => onSelect(day.key)}
                aria-pressed={active}
                className={`relative flex shrink-0 flex-col items-center rounded-xl border px-3 py-2 transition-colors ${
                  active
                    ? 'border-accent/40 text-text'
                    : 'border-hairline bg-white/[0.02] text-muted hover:border-white/15 hover:text-text'
                }`}
              >
                {active && (
                  <motion.span
                    layoutId="dateStripActive"
                    className="absolute inset-0 -z-10 rounded-xl bg-accent/[0.12] shadow-[0_0_18px_-4px_rgba(201,162,75,0.55)] ring-1 ring-accent/30"
                    transition={{ type: 'spring', stiffness: 380, damping: 32 }}
                  />
                )}
                <span
                  className={`text-[0.6rem] font-semibold uppercase tracking-[0.14em] ${
                    active ? 'text-accent' : isToday ? 'text-accent/70' : 'text-muted/70'
                  }`}
                >
                  {formatKyivWeekday(day.iso)}
                </span>
                <span className="mt-0.5 whitespace-nowrap text-sm font-semibold tabular-nums">
                  {formatKyivDayMonth(day.iso)}
                </span>
                <span className="mt-1 flex items-center gap-1">
                  <span
                    className={`h-1 w-1 rounded-full ${active ? 'bg-accent' : 'bg-muted/40'}`}
                  />
                  <span className="text-[0.55rem] font-medium tabular-nums text-muted/60">
                    {day.matches.length}
                  </span>
                </span>
              </button>
            )
          })}
        </div>

        <ArrowButton dir="next" disabled={index >= days.length - 1} onClick={goNext} label={t('dateStrip.next')} />

        <button
          type="button"
          onClick={() => hasToday && onSelect(today)}
          disabled={!hasToday}
          className="ml-1 hidden shrink-0 rounded-full border border-hairline bg-white/[0.03] px-3 py-1.5 text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-muted transition-colors hover:border-accent/40 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:border-hairline disabled:hover:text-muted sm:block"
        >
          {t('dateStrip.today')}
        </button>
      </div>
    </div>
  )
}
