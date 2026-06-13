import { useState } from 'react'

interface FlagProps {
  code: string | undefined
  flagUrl: string | undefined
  /** Accessible name, typically the team name. */
  label: string
  className?: string
}

/**
 * Renders a country flag via the `flag-icons` CSS package (ISO 3166-1 alpha-2),
 * falling back to the API-provided <img> when no matching icon class exists or
 * the CSS sprite fails to load.
 *
 * flag-icons expects lowercase two-letter ISO codes. FIFA three-letter codes
 * (e.g. "BRA") won't resolve, so we only attempt the icon for 2-char codes and
 * otherwise go straight to the image fallback.
 */
export default function Flag({ code, flagUrl, label, className = '' }: FlagProps) {
  const iso = code?.trim().toLowerCase()
  const canUseIcon = !!iso && iso.length === 2
  const [imgFailed, setImgFailed] = useState(false)

  if (canUseIcon) {
    return (
      <span
        className={`fi fi-${iso} inline-block overflow-hidden rounded-[3px] ring-1 ring-white/10 ${className}`}
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
        className={`inline-block rounded-[3px] object-cover ring-1 ring-white/10 ${className}`}
      />
    )
  }

  // Last-resort neutral placeholder (no code, no image).
  return (
    <span
      className={`inline-block rounded-[3px] bg-white/10 ring-1 ring-white/10 ${className}`}
      role="img"
      aria-label={label}
    />
  )
}
