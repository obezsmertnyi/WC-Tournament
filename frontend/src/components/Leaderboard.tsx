import { useTranslation } from 'react-i18next'
import type { LeaderboardEntry } from '../types'
import Avatar from './Avatar'

interface LeaderboardProps {
  entries: LeaderboardEntry[]
}

/** Ranked standings of players: rank, avatar+nickname, points, exact count. */
export default function Leaderboard({ entries }: LeaderboardProps) {
  const { t } = useTranslation()

  if (entries.length === 0) {
    return (
      <div className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center backdrop-blur-md">
        <p className="text-sm font-medium text-text">{t('competition.leaderboardEmpty')}</p>
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-hairline bg-gradient-to-b from-white/[0.05] to-white/[0.015] shadow-[0_8px_24px_-16px_rgba(0,0,0,0.8)] backdrop-blur-md">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-hairline text-[0.6rem] uppercase tracking-[0.12em] text-muted/70">
            <th className="py-2.5 pl-3 pr-1 text-left font-semibold sm:pl-4">#</th>
            <th className="py-2.5 pr-2 text-left font-semibold">{t('competition.player')}</th>
            <th className="px-1.5 py-2.5 text-center font-semibold tabular-nums sm:px-2">
              {t('competition.exact')}
            </th>
            <th className="px-2 py-2.5 pr-3 text-center font-semibold text-accent/80 sm:pr-4">
              {t('competition.points')}
            </th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e, i) => {
            const rank = i + 1
            const leader = rank === 1
            return (
              <tr
                key={e.userId}
                className="border-b border-hairline/60 transition-colors last:border-0 hover:bg-white/[0.03]"
              >
                <td
                  className={`py-3 pl-3 pr-1 text-center font-bold tabular-nums sm:pl-4 ${
                    leader ? 'text-accent' : 'text-muted'
                  }`}
                >
                  {rank}
                </td>
                <td className="min-w-0 py-3 pr-2">
                  <div className="flex items-center gap-2.5">
                    <Avatar
                      src={e.avatarUrl}
                      nickname={e.nickname}
                      gold={leader}
                      className="h-7 w-7 text-xs"
                    />
                    <span
                      className={`truncate font-medium ${leader ? 'text-accent' : 'text-text'}`}
                    >
                      {e.nickname}
                    </span>
                  </div>
                </td>
                <td className="px-1.5 py-3 text-center tabular-nums text-muted sm:px-2">
                  {e.exactCount}
                </td>
                <td
                  className={`px-2 py-3 pr-3 text-center font-bold tabular-nums sm:pr-4 ${
                    leader ? 'text-accent' : 'text-text'
                  }`}
                >
                  {e.points}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
