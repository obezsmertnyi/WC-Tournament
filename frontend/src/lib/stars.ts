/**
 * Featured-players ("stars") configuration used for large, blended page artwork
 * (see `components/StarHero.tsx`). These are NOT rendered as a labelled avatar
 * widget any more — the portraits are incorporated into the page design as big,
 * desaturated, gradient-masked hero imagery that fades into the dark background.
 *
 * Portraits are real, locally-hosted photos from Wikimedia Commons under
 * CC BY / CC BY-SA licenses — see `public/img/ATTRIBUTION.md` for the source,
 * author, and license of every file. If a portrait fails to load at runtime the
 * UI hides that image entirely (no placeholder, no monogram).
 *
 * `teamCode` is the FIFA-3 code of the player's national team. It is used to map
 * a group (which contains that team) to "its" star for the GroupDetail hero —
 * e.g. the group containing ARG shows Messi, NOR shows Haaland, ENG shows Kane.
 *
 * To add a player: download a clean CC-licensed photo into `public/img/stars/`,
 * record it in ATTRIBUTION.md, and add an entry here.
 */

export interface Star {
  /** Display name, e.g. "Lionel Messi". */
  name: string
  /** FIFA-3 national-team code (e.g. "ARG"). Drives the group→star mapping. */
  teamCode: string
  /** Local portrait path under /public (e.g. "/img/stars/messi.jpg"). */
  imageUrl: string
}

export const STARS: Star[] = [
  { name: 'Lionel Messi', teamCode: 'ARG', imageUrl: '/img/stars/messi.jpg' },
  { name: 'Cristiano Ronaldo', teamCode: 'POR', imageUrl: '/img/stars/ronaldo.jpg' },
  { name: 'Kylian Mbappé', teamCode: 'FRA', imageUrl: '/img/stars/mbappe.jpg' },
  { name: 'Vinícius Júnior', teamCode: 'BRA', imageUrl: '/img/stars/vinicius.jpg' },
  { name: 'Harry Kane', teamCode: 'ENG', imageUrl: '/img/stars/kane.jpg' },
  { name: 'Erling Haaland', teamCode: 'NOR', imageUrl: '/img/stars/haaland.jpg' },
]

/** Fast lookup of a featured star by FIFA-3 team code (uppercased). */
const STARS_BY_TEAM: Record<string, Star> = STARS.reduce<Record<string, Star>>(
  (acc, star) => {
    acc[star.teamCode.toUpperCase()] = star
    return acc
  },
  {},
)

/**
 * Returns the featured star for a group, given the FIFA-3 codes of the teams in
 * that group, or `undefined` if none of the featured players belong to the
 * group. The first match (in team order) wins.
 */
export function starForTeams(teamCodes: Iterable<string>): Star | undefined {
  for (const code of teamCodes) {
    const star = STARS_BY_TEAM[code?.toUpperCase()]
    if (star) return star
  }
  return undefined
}
