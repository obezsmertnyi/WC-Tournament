import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { STARS, type Star } from '../lib/stars'
import Flag from './Flag'

/**
 * Subtle "stars" hero band for the Calendar landing — a row of real player
 * portraits in circular champagne-gold frames with a small country-flag badge
 * and a name caption. Portraits are locally-hosted CC-licensed photos (see
 * public/img/ATTRIBUTION.md). If an image fails to load we hide that portrait
 * entirely — never a broken <img>, never a monogram placeholder.
 */
function StarPortrait({ star }: { star: Star }) {
  const [imgFailed, setImgFailed] = useState(false)
  if (imgFailed || star.imageUrl.trim() === '') return null

  return (
    <li className="flex flex-col items-center gap-1.5">
      <div className="relative h-14 w-14 sm:h-16 sm:w-16">
        <img
          src={star.imageUrl}
          alt={star.name}
          loading="lazy"
          onError={() => setImgFailed(true)}
          className="h-full w-full rounded-full object-cover object-top shadow-[0_4px_14px_-4px_rgba(0,0,0,0.6)]"
        />
        {/* champagne-gold ring drawn on top so it always reads crisply */}
        <span className="pointer-events-none absolute inset-0 rounded-full ring-2 ring-accent/60 shadow-[0_0_16px_-4px_rgba(201,162,75,0.55)]" />
        {star.code && (
          <span className="absolute -bottom-1 left-1/2 -translate-x-1/2">
            <Flag
              code={star.code}
              flagUrl={undefined}
              label={star.code}
              className="h-[0.65rem] w-[1rem] ring-1 ring-bg"
            />
          </span>
        )}
      </div>
      <span className="max-w-[5rem] truncate text-center text-[0.6rem] font-medium text-muted/70">
        {star.name}
      </span>
    </li>
  )
}

export default function StarsBand() {
  const { t } = useTranslation()
  if (STARS.length === 0) return null

  return (
    <section
      aria-label={t('stars.heading')}
      className="mb-6 overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.04] to-white/[0.01] px-4 py-4 backdrop-blur-md"
    >
      <p className="mb-3 text-[0.6rem] font-semibold uppercase tracking-[0.2em] text-muted/70">
        {t('stars.heading')}
      </p>
      <ul className="flex flex-wrap items-start justify-center gap-x-5 gap-y-3 sm:justify-start">
        {STARS.map((star) => (
          <StarPortrait key={star.name} star={star} />
        ))}
      </ul>
    </section>
  )
}
