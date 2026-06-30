import type { User } from '../types'

// Demo-mode access gating, mirrored from the backend. The server already folds
// admin + demo-off into access='rw', so the UI can drive everything off the
// single `access` field; `demoMode` only controls whether we explain the
// restriction to the user.

/** Can the user participate (submit predictions / bonus picks)? */
export function canParticipate(user: User | null): boolean {
  return user?.access === 'rw'
}

/** Can the user see other players' data (reveals, leaderboard, audit, scorers)? */
export function canSeeOthers(user: User | null): boolean {
  return user?.access === 'ro' || user?.access === 'rw'
}

/** Is the user in a restricted demo tier (so we should show an explainer)? */
export function isRestricted(user: User | null): boolean {
  return !!user?.demoMode && user.access !== 'rw'
}
