import { useState } from 'react'

interface AvatarProps {
  src: string | null | undefined
  nickname: string
  className?: string
  /** Highlight ring for the leader (#1). */
  gold?: boolean
}

/**
 * Round user avatar with a graceful fallback to an initial monogram on a glass
 * chip. Matches the Flag component's defensive image-error handling.
 */
export default function Avatar({ src, nickname, className = '', gold = false }: AvatarProps) {
  const [failed, setFailed] = useState(false)
  const initial = (nickname || '?').trim().charAt(0).toUpperCase() || '?'
  const ring = gold
    ? 'ring-2 ring-accent shadow-[0_0_12px_-2px_rgba(201,162,75,0.6)]'
    : 'ring-1 ring-white/15'

  if (src && !failed) {
    return (
      <img
        src={src}
        alt={nickname}
        loading="lazy"
        onError={() => setFailed(true)}
        className={`inline-block shrink-0 rounded-full object-cover ${ring} ${className}`}
      />
    )
  }

  return (
    <span
      role="img"
      aria-label={nickname}
      className={`inline-flex shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-white/[0.14] to-white/[0.04] font-semibold uppercase text-text ${ring} ${className}`}
    >
      {initial}
    </span>
  )
}
