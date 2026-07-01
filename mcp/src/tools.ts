import { z } from 'zod'
import { type ReadClient, asMatches, kickedOff } from './api.js'

// Every tool declares a strict zod schema (.strict() rejects unknown params),
// validating before any fetch — the structured-invocation control that closes
// the injection vector (OWASP MCP Framework 4). All tools are read-only.

const DATE = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'date must be YYYY-MM-DD')
const STAGE = z.enum(['group', 'r32', 'r16', 'qf', 'sf', 'final'])
const GROUP = z.string().regex(/^[A-L]$/i, 'group must be a single letter A–L')

export interface Tool {
  name: string
  title: string
  description: string
  schema: z.AnyZodObject
  // args are zod-validated by the schema before run is called.
  run(args: Record<string, unknown>, client: ReadClient): Promise<unknown>
}

function team(m: Record<string, any>, side: 'home' | 'away') {
  return m[side]?.code ?? m[`placeholder${side === 'home' ? 'Home' : 'Away'}`] ?? null
}

export const listFixtures: Tool = {
  name: 'list_fixtures',
  title: 'List fixtures',
  description: 'Fixtures (teams, kickoff, status, score), optionally filtered by date (YYYY-MM-DD) or stage.',
  schema: z.object({ date: DATE.optional(), stage: STAGE.optional() }).strict(),
  async run(args, client) {
    let rows = asMatches(await client.get('/api/matches'))
    if (args.stage) rows = rows.filter((m) => m.stage === args.stage)
    if (args.date) rows = rows.filter((m) => String(m.kickoffAt ?? '').startsWith(args.date as string))
    return rows.slice(0, 200).map((m: Record<string, any>) => ({
      matchNumber: m.matchNumber ?? null,
      stage: m.stage,
      home: team(m, 'home'),
      away: team(m, 'away'),
      kickoffAt: m.kickoffAt ?? null,
      status: m.status,
      score: m.homeScore != null && m.awayScore != null ? `${m.homeScore}:${m.awayScore}` : null,
    }))
  },
}

export const groupStandings: Tool = {
  name: 'group_standings',
  title: 'Group standings',
  description: 'Derived group tables (P/W/D/L/GF/GA/GD/Pts), optionally a single group A–L.',
  schema: z.object({ group: GROUP.optional() }).strict(),
  async run(args, client) {
    const data = (await client.get('/api/standings')) as { groups?: Record<string, any>[] }
    let groups = data.groups ?? []
    if (args.group) {
      const g = (args.group as string).toUpperCase()
      groups = groups.filter((x) => String(x.group ?? x.letter ?? '').toUpperCase() === g)
    }
    return groups
  },
}

export const leaderboard: Tool = {
  name: 'leaderboard',
  title: 'Leaderboard',
  description: 'Players ranked by points (match + bonus split). Admins excluded by the API.',
  schema: z.object({ limit: z.number().int().min(1).max(100).optional() }).strict(),
  async run(args, client) {
    const data = (await client.get('/api/leaderboard')) as unknown
    const rows = Array.isArray(data) ? data : []
    const limit = (args.limit as number | undefined) ?? 50
    return rows.slice(0, limit)
  },
}

export const bracket: Tool = {
  name: 'bracket',
  title: 'Knockout bracket',
  description: 'The knockout ties (R32→final) with resolved teams or placeholders.',
  schema: z.object({}).strict(),
  async run(_args, client) {
    const rows = asMatches(await client.get('/api/matches'))
    return rows
      .filter((m) => m.stage !== 'group')
      .map((m: Record<string, any>) => ({
        matchNumber: m.matchNumber ?? null,
        stage: m.stage,
        home: team(m, 'home'),
        away: team(m, 'away'),
        status: m.status,
        score: m.homeScore != null && m.awayScore != null ? `${m.homeScore}:${m.awayScore}` : null,
      }))
  },
}

export const playerPredictions: Tool = {
  name: 'player_predictions',
  title: "A player's revealed predictions",
  description: "A named player's predictions — ONLY for matches that have kicked off (reveal lock).",
  schema: z.object({ nickname: z.string().min(1).max(40), limit: z.number().int().min(1).max(50).optional() }).strict(),
  async run(args, client) {
    const nickname = args.nickname as string
    const limit = (args.limit as number | undefined) ?? 20
    // Reveal lock (FR-072 spirit): only query matches that have kicked off; the
    // reveal endpoint itself returns a locked marker before kickoff.
    const revealed = asMatches(await client.get('/api/matches')).filter(kickedOff).slice(0, limit)
    const out: unknown[] = []
    for (const m of revealed) {
      const data = (await client.get(`/api/matches/${m.matchNumber ?? m.id}/predictions`)) as unknown
      if (!Array.isArray(data)) continue // locked shape => skip
      const pick = (data as Record<string, any>[]).find((p) => p.nickname === nickname)
      if (pick) {
        out.push({ matchNumber: m.matchNumber ?? null, home: pick.home, away: pick.away, points: pick.points ?? null })
      }
    }
    return { nickname, revealed: out }
  },
}

export const TOOLS: Tool[] = [listFixtures, groupStandings, leaderboard, bracket, playerPredictions]
