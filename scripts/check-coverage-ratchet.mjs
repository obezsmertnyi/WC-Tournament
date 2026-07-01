#!/usr/bin/env node
// Backend test-coverage ratchet: total Go coverage may only go up. Uses Go's
// built-in coverprofile (no extra deps). Frontend coverage can be added later
// via @vitest/coverage-v8; kept backend-only to stay dependency-light.
//   node scripts/check-coverage-ratchet.mjs           # measure + print
//   node scripts/check-coverage-ratchet.mjs --check    # fail if below baseline (minus epsilon)
//   node scripts/check-coverage-ratchet.mjs --update   # ratchet baseline up
import { execSync } from 'node:child_process'
import { readFileSync, writeFileSync, existsSync, mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const BASELINE = 'quality/coverage-baseline.json'
const EPS = 0.5 // allow tiny noise; a real drop still trips

function measure() {
  const out = mkdtempSync(join(tmpdir(), 'wccov-'))
  const prof = join(out, 'c.out')
  execSync(`go test -coverprofile=${prof} ./... >/dev/null 2>&1 || true`, { cwd: 'backend', stdio: 'ignore' })
  const total = execSync(`go tool cover -func=${prof} | tail -1`, { cwd: 'backend' }).toString()
  const m = total.match(/([\d.]+)%/)
  return m ? parseFloat(m[1]) : NaN
}

const pct = measure()
if (Number.isNaN(pct)) {
  console.error('✗ could not measure coverage (set DATABASE_URL for integration tests)')
  process.exit(2)
}
const base = existsSync(BASELINE) ? JSON.parse(readFileSync(BASELINE, 'utf8')) : { backendPct: 0 }
const mode = process.argv.includes('--check') ? 'check' : process.argv.includes('--update') ? 'update' : 'report'

if (mode === 'check') {
  if (pct + EPS < base.backendPct) {
    console.error(`✗ coverage regressed: ${pct.toFixed(1)}% < baseline ${base.backendPct}%`)
    process.exit(1)
  }
  console.log(`✓ backend coverage ${pct.toFixed(1)}% (baseline ${base.backendPct}%)`)
} else if (mode === 'update') {
  writeFileSync(BASELINE, JSON.stringify({ backendPct: Math.round(pct * 10) / 10 }, null, 2) + '\n')
  console.log(`✓ coverage baseline set: ${pct.toFixed(1)}%`)
} else {
  console.log(`backend coverage ${pct.toFixed(1)}% (baseline ${base.backendPct ?? 'unset'}%)`)
}
