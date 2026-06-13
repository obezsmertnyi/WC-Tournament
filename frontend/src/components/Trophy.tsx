/**
 * Champagne-gold trophy mark — a real, locally-hosted photo of a generic gold
 * cup (Pexels, no-attribution-required; see public/img/ATTRIBUTION.md). It is a
 * generic trophy, NOT the official FIFA World Cup Trophy.
 *
 * `className` controls the box size from the caller. The image is masked into a
 * rounded frame and `object-top` keeps the gold cup bowl in view at small sizes
 * (e.g. the 22px wordmark mark in the app bar).
 *
 * `framed` (default true) adds the champagne-gold ring + soft shadow so the mark
 * reads as a deliberate, integrated element. Set `framed={false}` for ambient
 * watermark use where a bare, low-opacity image is wanted.
 */
export default function Trophy({
  className = '',
  framed = true,
}: {
  className?: string
  framed?: boolean
}) {
  return (
    <span className={`relative inline-block shrink-0 overflow-hidden ${className}`} aria-hidden>
      <img
        src="/img/trophy.jpg"
        alt=""
        loading="lazy"
        className="h-full w-full rounded-[inherit] object-cover object-top"
      />
      {framed && (
        <span className="pointer-events-none absolute inset-0 rounded-[inherit] ring-1 ring-accent/50 shadow-[0_0_14px_-4px_rgba(201,162,75,0.5)]" />
      )}
    </span>
  )
}
