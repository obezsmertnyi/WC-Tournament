/**
 * Minimal trophy mark — a restrained line-art cup in champagne gold.
 * Used as a faint ambient watermark, not a decorative blast.
 * `className` controls size/opacity from the caller.
 */
export default function Trophy({ className = '' }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 100 120"
      fill="none"
      stroke="#C9A24B"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden
    >
      {/* cup bowl */}
      <path d="M30 18 H70 V40 C70 56 61 66 50 66 C39 66 30 56 30 40 Z" />
      {/* left handle */}
      <path d="M30 24 C18 24 16 40 28 46" />
      {/* right handle */}
      <path d="M70 24 C82 24 84 40 72 46" />
      {/* stem */}
      <path d="M50 66 V82" />
      {/* base */}
      <path d="M38 82 H62" />
      <path d="M34 94 H66 L62 82 H38 Z" />
    </svg>
  )
}
