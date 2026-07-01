#!/usr/bin/env node
// Eval-surface ratchet: the number of eval cases per suite may only grow.
// Guards against silently deleting/weakening evals (the quality bar can't erode).
//   node scripts/check-eval-ratchet.mjs           # report counts
//   node scripts/check-eval-ratchet.mjs --check    # fail if below baseline
//   node scripts/check-eval-ratchet.mjs --update   # ratchet baseline up
import { readFileSync, writeFileSync, existsSync } from 'node:fs'

const BASELINE = 'quality/eval-baseline.json'

// suite -> [file, regex counting one eval case each]
const SUITES = {
  scoring: ['backend/internal/scoring/scoring_evals_test.go', /trace:\s*"FR-/g],
  recap: ['frontend/src/lib/recap.test.ts', /\bit\(/g],
  mcp: ['mcp/evals/tools.eval.test.ts', /\bit\(/g],
}

const counts = {}
for (const [suite, [file, re]] of Object.entries(SUITES)) {
  counts[suite] = existsSync(file) ? (readFileSync(file, 'utf8').match(re) ?? []).length : 0
}
const total = Object.values(counts).reduce((a, b) => a + b, 0)

const base = existsSync(BASELINE) ? JSON.parse(readFileSync(BASELINE, 'utf8')) : { suites: {}, total: 0 }
const mode = process.argv.includes('--check') ? 'check' : process.argv.includes('--update') ? 'update' : 'report'

if (mode === 'check') {
  const regressed = Object.entries(counts).filter(([s, n]) => n < (base.suites?.[s] ?? 0))
  if (regressed.length || total < (base.total ?? 0)) {
    console.error('✗ eval ratchet regressed:', JSON.stringify({ now: counts, baseline: base.suites }))
    process.exit(1)
  }
  console.log(`✓ eval ratchet: ${total} cases (baseline ${base.total}) — ${JSON.stringify(counts)}`)
} else if (mode === 'update') {
  writeFileSync(BASELINE, JSON.stringify({ suites: counts, total }, null, 2) + '\n')
  console.log(`✓ eval baseline set: ${total} — ${JSON.stringify(counts)}`)
} else {
  console.log(JSON.stringify({ suites: counts, total, baseline: base.total }, null, 2))
}
