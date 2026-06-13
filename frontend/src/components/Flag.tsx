import { useState } from 'react'
import { fifaToIso } from '../lib/flags'

interface FlagProps {
  code: string | undefined
  flagUrl: string | undefined
  /** Accessible name, typically the team name. */
  label: string
  className?: string
}

/**
 * Renders a country flag with a graceful fallback chain:
 *   1. `flag-icons` CSS sprite — when the FIFA-3 (or ISO-2) code maps to a
 *      known ISO code (see lib/flags.ts). Crisp and consistent.
 *   2. The API-provided <img src={flagUrl}> — when no ISO mapping exists.
 *   3. A neutral monogram chip showing the code — last resort.
 *
 * Flags are a primary source of color in the UI, so we lean hard on the sprite
 * (which always renders) before falling back.
 */
export default function Flag({ code, flagUrl, label, className = '' }: FlagProps) {
  const iso = fifaToIso(code)
  const [imgFailed, setImgFailed] = useState(false)

  if (iso) {
    return (
      <span
        className={`fi fi-${iso} inline-block shrink-0 overflow-hidden rounded-[4px] bg-white/5 bg-cover bg-center shadow-[0_1px_2px_rgba(0,0,0,0.45)] ring-1 ring-white/15 ${className}`}
        role="img"
        aria-label={label}
      />
    )
  }

  if (flagUrl && !imgFailed) {
    return (
      <img
        src={flagUrl}
        alt={label}
        loading="lazy"
        onError={() => setImgFailed(true)}
        className={`inline-block shrink-0 rounded-[4px] object-cover shadow-[0_1px_2px_rgba(0,0,0,0.45)] ring-1 ring-white/15 ${className}`}
      />
    )
  }

  // Last-resort neutral monogram chip (no ISO mapping, no usable image).
  const monogram = (code ?? '').trim().slice(0, 3).toUpperCase()
  return (
    <span
      className={`inline-flex shrink-0 items-center justify-center rounded-[4px] bg-gradient-to-br from-white/[0.12] to-white/[0.04] text-[0.5rem] font-semibold uppercase tracking-tight text-muted ring-1 ring-white/15 ${className}`}
      role="img"
      aria-label={label}
    >
      {monogram || '·'}
    </span>
  )
}
