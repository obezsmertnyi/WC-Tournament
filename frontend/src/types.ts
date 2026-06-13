export type Team = { name: string; code: string; flagUrl: string } | null

export type Stage = 'group' | 'r32' | 'r16' | 'qf' | 'sf' | 'final' | 'third'

export type MatchStatus = 'scheduled' | 'live' | 'finished'

export interface Venue {
  stadium: string
  city: string
  country: string
}

export interface Match {
  id: number
  stage: Stage
  group: string | null
  matchNumber: number
  kickoffAt: string // UTC ISO
  status: MatchStatus
  home: Team
  away: Team
  homeScore: number | null
  awayScore: number | null
  venue: Venue
  placeholderHome: string | null
  placeholderAway: string | null
}

export interface StandingRow {
  teamId: number
  name: string
  code: string
  flagUrl: string
  played: number
  win: number
  draw: number
  loss: number
  gf: number
  ga: number
  gd: number
  points: number
  rank: number
}

export interface GroupStanding {
  group: string
  rows: StandingRow[]
}
