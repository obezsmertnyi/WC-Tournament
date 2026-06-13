/**
 * Featured-players ("stars") configuration for the Calendar hero band.
 *
 * Each entry renders as a refined circular portrait with a champagne-gold ring.
 * Portraits are real, locally-hosted photos from Wikimedia Commons under
 * CC BY / CC BY-SA licenses — see `public/img/ATTRIBUTION.md` for the source,
 * author, and license of every file. If a portrait fails to load at runtime the
 * UI hides that entry entirely (no placeholder, no monogram).
 *
 * To add a player: download a clean CC-licensed head/upper-body shot into
 * `public/img/stars/`, record it in ATTRIBUTION.md, and add an entry here.
 */

export interface Star {
  /** Display name, e.g. "Lionel Messi". */
  name: string
  /** Country code (FIFA-3) for the small flag accent; optional. */
  code?: string
  /** Local portrait path under /public (e.g. "/img/stars/messi.jpg"). */
  imageUrl: string
}

export const STARS: Star[] = [
  { name: 'Lionel Messi', code: 'ARG', imageUrl: '/img/stars/messi.jpg' },
  { name: 'Cristiano Ronaldo', code: 'POR', imageUrl: '/img/stars/ronaldo.jpg' },
  { name: 'Kylian Mbappé', code: 'FRA', imageUrl: '/img/stars/mbappe.jpg' },
  { name: 'Vinícius Júnior', code: 'BRA', imageUrl: '/img/stars/vinicius.jpg' },
  { name: 'Harry Kane', code: 'ENG', imageUrl: '/img/stars/kane.jpg' },
  { name: 'Erling Haaland', code: 'NOR', imageUrl: '/img/stars/haaland.jpg' },
]
