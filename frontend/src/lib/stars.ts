/**
 * Featured-players ("stars") configuration for the Calendar hero band.
 *
 * Each entry renders as a refined circular portrait with a champagne-gold ring.
 * When `imageUrl` is empty the UI shows a tasteful monogram placeholder built
 * from the player's initials — never a broken image — so the band looks polished
 * even with no photos configured.
 *
 * IMPORTANT: image URLs are intentionally left EMPTY. Do NOT scrape or hardcode
 * copyrighted photographs.
 *
 * TODO: owner to provide licensed player image URLs (or local files under
 * /public and reference them as e.g. "/stars/messi.jpg"). Leave empty to keep
 * the elegant monogram placeholders.
 */

export interface Star {
  /** Display name, e.g. "Lionel Messi". */
  name: string
  /** Country code (FIFA-3) for the small flag accent; optional. */
  code?: string
  /** Licensed portrait URL. Empty → monogram placeholder. */
  imageUrl: string
}

export const STARS: Star[] = [
  { name: 'Lionel Messi', code: 'ARG', imageUrl: '' },
  { name: 'Cristiano Ronaldo', code: 'POR', imageUrl: '' },
  { name: 'Kylian Mbappé', code: 'FRA', imageUrl: '' },
  { name: 'Vinícius Júnior', code: 'BRA', imageUrl: '' },
  { name: 'Jude Bellingham', code: 'ENG', imageUrl: '' },
]

/** Initials for the monogram placeholder, e.g. "Lionel Messi" → "LM". */
export function starInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '★'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}
