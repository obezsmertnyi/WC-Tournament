/**
 * Golden Boot (top-scorer) candidates offered as quick picks in the bonus panel.
 * Ordered by 2026 World Cup top-goalscorer betting odds (Mbappé favourite down).
 * The free-text field still accepts any name — this is just a discoverable
 * shortlist so the picker isn't empty. `teamCode` is the FIFA-3 code for a flag.
 *
 * Sources (June 2026): covers.com, si.com, goal.com top-goalscorer markets.
 */
export interface ScorerCandidate {
  name: string
  teamCode: string
}

export const TOP_SCORERS: ScorerCandidate[] = [
  { name: 'Kylian Mbappé', teamCode: 'FRA' },
  { name: 'Harry Kane', teamCode: 'ENG' },
  { name: 'Mikel Oyarzabal', teamCode: 'ESP' },
  { name: 'Erling Haaland', teamCode: 'NOR' },
  { name: 'Lionel Messi', teamCode: 'ARG' },
  { name: 'Cristiano Ronaldo', teamCode: 'POR' },
  { name: 'Julián Álvarez', teamCode: 'ARG' },
  { name: 'Raphinha', teamCode: 'BRA' },
  { name: 'Lamine Yamal', teamCode: 'ESP' },
  { name: 'Kai Havertz', teamCode: 'GER' },
  { name: 'Michael Olise', teamCode: 'FRA' },
  { name: 'Vinícius Júnior', teamCode: 'BRA' },
  { name: 'Lautaro Martínez', teamCode: 'ARG' },
  { name: 'Cody Gakpo', teamCode: 'NED' },
  { name: 'Romelu Lukaku', teamCode: 'BEL' },
  { name: 'Ousmane Dembélé', teamCode: 'FRA' },
]
