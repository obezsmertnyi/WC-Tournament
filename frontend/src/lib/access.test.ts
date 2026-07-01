import { describe, it, expect } from 'vitest'
import type { User, AccessLevel } from '../types'
import { canParticipate, canSeeOthers, isRestricted } from './access'

// Frontend mirror of the demo-access contract (docs/features/demo-access/spec.md).
function user(access: AccessLevel, demoMode: boolean): User {
  return {
    id: '1',
    nickname: 'n',
    avatarUrl: null,
    favoriteTeamCode: null,
    role: 'user',
    demoMode,
    access,
  }
}

describe('demo-mode access gating (frontend)', () => {
  // @trace FR-031
  it('none → browse only: cannot see others, cannot participate', () => {
    const u = user('none', true)
    expect(canSeeOthers(u)).toBe(false)
    expect(canParticipate(u)).toBe(false)
    expect(isRestricted(u)).toBe(true)
  })

  // @trace FR-031
  it('ro → sees others but cannot participate', () => {
    const u = user('ro', true)
    expect(canSeeOthers(u)).toBe(true)
    expect(canParticipate(u)).toBe(false)
    expect(isRestricted(u)).toBe(true)
  })

  // @trace FR-031
  it('rw → full access, not restricted', () => {
    const u = user('rw', true)
    expect(canSeeOthers(u)).toBe(true)
    expect(canParticipate(u)).toBe(true)
    expect(isRestricted(u)).toBe(false)
  })

  // @trace FR-032
  it('demo off (server returns rw) → never restricted', () => {
    const u = user('rw', false)
    expect(canParticipate(u)).toBe(true)
    expect(isRestricted(u)).toBe(false)
  })

  // @trace FR-031
  it('anonymous (null user) → no access', () => {
    expect(canParticipate(null)).toBe(false)
    expect(canSeeOthers(null)).toBe(false)
    expect(isRestricted(null)).toBe(false)
  })
})
