/**
 * FIFA World Cup trophy mark — a gold trophy illustration on a TRANSPARENT
 * background (Wikimedia Commons, CC0; see public/img/ATTRIBUTION.md). Transparent
 * so it sits cleanly on the dark theme — no white box, no black-on-black blur.
 *
 * `className` controls the box size from the caller; the image is `object-contain`
 * so the whole trophy is visible and never cropped. `framed` is kept for API
 * compatibility but does nothing now (a transparent cutout needs no frame).
 */
export default function Trophy({
  className = '',
}: {
  className?: string
  framed?: boolean
}) {
  return (
    <img
      src="/img/trophy.png"
      alt=""
      aria-hidden
      loading="lazy"
      className={`shrink-0 object-contain ${className}`}
    />
  )
}
