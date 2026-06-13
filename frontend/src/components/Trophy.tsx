/**
 * Champagne-gold trophy mark.
 *
 * - `variant="line"` (default) — restrained line-art cup, used as the faint
 *   ambient background watermark and other large, subtle placements.
 * - `variant="solid"` — the same silhouette with a soft gold fill + crisp
 *   stroke, sized for inline use beside the wordmark / page title where the
 *   trophy is meant to read as a deliberate, present element.
 *
 * `className` controls size/opacity from the caller.
 */
export default function Trophy({
  className = '',
  variant = 'line',
}: {
  className?: string
  variant?: 'line' | 'solid'
}) {
  const solid = variant === 'solid'
  return (
    <svg
      viewBox="0 0 100 120"
      fill="none"
      stroke="#C9A24B"
      strokeWidth={solid ? 3 : 2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden
    >
      {/* cup bowl */}
      <path
        d="M30 18 H70 V40 C70 56 61 66 50 66 C39 66 30 56 30 40 Z"
        fill={solid ? 'rgba(201,162,75,0.18)' : 'none'}
      />
      {/* left handle */}
      <path d="M30 24 C18 24 16 40 28 46" />
      {/* right handle */}
      <path d="M70 24 C82 24 84 40 72 46" />
      {/* stem */}
      <path d="M50 66 V82" />
      {/* base */}
      <path d="M38 82 H62" />
      <path
        d="M34 94 H66 L62 82 H38 Z"
        fill={solid ? 'rgba(201,162,75,0.18)' : 'none'}
      />
    </svg>
  )
}
