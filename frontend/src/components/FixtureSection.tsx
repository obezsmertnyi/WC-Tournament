import type { Match } from '../types'
import { formatKyivDate, kyivDayKey } from '../lib/fixtures'
import FixtureCard from './FixtureCard'

interface FixtureSectionProps {
  title: string
  /** Optional small label above the title (e.g. phase name). */
  eyebrow?: string
  matches: Match[]
}

interface DayBucket {
  key: string
  label: string
  matches: Match[]
}

function bucketByDay(matches: Match[]): DayBucket[] {
  const map = new Map<string, Match[]>()
  for (const m of matches) {
    const key = kyivDayKey(m.kickoffAt)
    const arr = map.get(key) ?? []
    arr.push(m)
    map.set(key, arr)
  }
  return [...map.keys()]
    .sort()
    .map((key) => ({
      key,
      label: formatKyivDate(map.get(key)![0].kickoffAt),
      matches: map.get(key)!,
    }))
}

export default function FixtureSection({ title, eyebrow, matches }: FixtureSectionProps) {
  if (matches.length === 0) return null
  const days = bucketByDay(matches)

  return (
    <section className="scroll-mt-24">
      <header className="mb-4 border-b border-hairline pb-2.5">
        {eyebrow && (
          <p className="text-[0.6rem] font-semibold uppercase tracking-[0.24em] text-muted/60">
            {eyebrow}
          </p>
        )}
        <h2 className="text-xs font-semibold uppercase tracking-[0.2em] text-muted">
          {title}
        </h2>
      </header>

      <div className="space-y-6">
        {days.map((day) => (
          <div key={day.key}>
            <p className="mb-2.5 text-[0.7rem] font-medium capitalize tracking-wide text-muted/70">
              {day.label}
            </p>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {day.matches.map((m, i) => (
                <FixtureCard key={m.id} match={m} index={i} />
              ))}
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
