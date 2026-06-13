import { useState } from 'react'
import { STARS, type Star } from '../lib/stars'

/**
 * Large, blended editorial player artwork — NOT a labelled avatar widget.
 *
 * Player portraits are rendered big and treated as page decoration: desaturated
 * toward charcoal, tinted with a champagne-gold duotone, and faded into the dark
 * background with layered gradient masks so titles / controls sit cleanly on top.
 * The whole band is `pointer-events-none` and `aria-hidden` — it is purely
 * decorative and never intercepts interaction or hurts readability.
 *
 * Two layouts:
 *   - `variant="band"` (default): a horizontal strip of several portraits, used
 *     as a top hero on the Calendar / Groups index.
 *   - `variant="portrait"`: one large portrait anchored to the right, used on a
 *     GroupDetail page for that group's star.
 *
 * If an image fails to load it is dropped silently (no broken <img>).
 */

const DUOTONE_FILTER = 'grayscale(1) contrast(1.05) brightness(0.62) sepia(0.45)'

function HeroImage({
  star,
  className,
  imgClassName,
  style,
}: {
  star: Star
  className?: string
  imgClassName?: string
  style?: React.CSSProperties
}) {
  const [failed, setFailed] = useState(false)
  if (failed || star.imageUrl.trim() === '') return null
  return (
    <div className={className} style={style}>
      <img
        src={star.imageUrl}
        alt=""
        loading="lazy"
        onError={() => setFailed(true)}
        className={imgClassName}
        style={{ filter: DUOTONE_FILTER }}
      />
    </div>
  )
}

/** Champagne-gold tint + readability scrim layered over the imagery. */
function HeroOverlays() {
  return (
    <>
      {/* Gold duotone wash — multiply keeps the charcoal shadows, lifts gold. */}
      <span
        aria-hidden
        className="pointer-events-none absolute inset-0 mix-blend-multiply"
        style={{
          background:
            'linear-gradient(115deg, rgba(201,162,75,0.34) 0%, rgba(201,162,75,0.10) 38%, rgba(11,12,14,0) 70%)',
        }}
      />
      {/* Bottom + left scrim so any text above remains legible. */}
      <span
        aria-hidden
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            'linear-gradient(to top, #0B0C0E 4%, rgba(11,12,14,0.55) 42%, rgba(11,12,14,0) 100%), linear-gradient(to right, #0B0C0E 0%, rgba(11,12,14,0.35) 34%, rgba(11,12,14,0) 78%)',
        }}
      />
      {/* Hairline gold seam at the very bottom for a premium edge. */}
      <span
        aria-hidden
        className="pointer-events-none absolute inset-x-0 bottom-0 h-px"
        style={{
          background:
            'linear-gradient(to right, rgba(201,162,75,0) 0%, rgba(201,162,75,0.4) 45%, rgba(201,162,75,0) 100%)',
        }}
      />
    </>
  )
}

export interface StarHeroProps {
  /** Layout. "band" = strip of several; "portrait" = one large image. */
  variant?: 'band' | 'portrait'
  /** Override the players shown. Defaults to all featured STARS. */
  stars?: Star[]
  /** Extra classes on the outer wrapper (e.g. height tweaks). */
  className?: string
}

/**
 * Decorative hero artwork. Absolutely positioned by default so callers can layer
 * page content on top within a `relative` container; the band has a fixed height
 * and clips its imagery.
 */
export default function StarHero({
  variant = 'band',
  stars = STARS,
  className = '',
}: StarHeroProps) {
  if (stars.length === 0) return null

  if (variant === 'portrait') {
    const star = stars[0]
    return (
      <div
        aria-hidden
        className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`}
        // Fade the portrait out toward the left so left-aligned text stays clean.
        style={{
          WebkitMaskImage:
            'linear-gradient(to left, #000 0%, rgba(0,0,0,0.85) 30%, rgba(0,0,0,0) 72%)',
          maskImage:
            'linear-gradient(to left, #000 0%, rgba(0,0,0,0.85) 30%, rgba(0,0,0,0) 72%)',
        }}
      >
        <HeroImage
          star={star}
          className="absolute right-0 top-1/2 h-[150%] w-[58%] max-w-[420px] -translate-y-1/2 sm:w-[46%]"
          imgClassName="h-full w-full object-cover object-top opacity-[0.5]"
        />
        <HeroOverlays />
      </div>
    )
  }

  // "band": a row of large portraits bleeding across the top.
  return (
    <div
      aria-hidden
      className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`}
      style={{
        WebkitMaskImage:
          'linear-gradient(to bottom, #000 0%, rgba(0,0,0,0.9) 45%, rgba(0,0,0,0) 100%)',
        maskImage:
          'linear-gradient(to bottom, #000 0%, rgba(0,0,0,0.9) 45%, rgba(0,0,0,0) 100%)',
      }}
    >
      <div className="absolute inset-0 flex justify-end">
        {stars.map((star, i) => (
          <HeroImage
            key={star.name}
            star={star}
            className="relative h-full flex-1"
            imgClassName="h-full w-full object-cover object-top"
            // Stagger opacity so the band reads as a soft, layered collage,
            // brighter toward the right edge, faintest under the title.
            style={{ opacity: 0.16 + i * 0.05 }}
          />
        ))}
      </div>
      <HeroOverlays />
    </div>
  )
}
