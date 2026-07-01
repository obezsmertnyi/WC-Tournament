#!/usr/bin/env node
// Gate roll-up: runs the deterministic check-* gates and prints PASS/FAIL/SKIP.
// Turns hand-maintained CHECKLIST ticks into a generated status line.
import { execSync } from 'node:child_process'

const GATES = [
  ['G-trace', 'traceability fresh + non-regressing', 'node scripts/gen-traceability.mjs --check'],
  ['G-eval', 'eval-surface ratchet', 'node scripts/check-eval-ratchet.mjs --check'],
  ['G-cover', 'backend coverage ratchet', 'node scripts/check-coverage-ratchet.mjs --check'],
]

let failed = 0
console.log('Gate status')
console.log('-----------')
for (const [id, desc, cmd] of GATES) {
  let status = 'PASS'
  try {
    execSync(cmd, { stdio: 'pipe' })
  } catch (e) {
    // exit code 2 = could-not-measure (e.g. coverage needs DATABASE_URL) → SKIP
    status = e.status === 2 ? 'SKIP' : 'FAIL'
    if (status === 'FAIL') failed++
  }
  console.log(`  ${status.padEnd(4)}  ${id.padEnd(9)} ${desc}`)
}
console.log('-----------')
console.log(failed ? `✗ ${failed} gate(s) failed` : '✓ all gates pass (skips noted)')
process.exit(failed ? 1 : 0)
