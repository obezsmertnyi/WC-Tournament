#!/usr/bin/env node
// Generate the requirements-traceability matrix from docs/requirements.md +
// `@trace <FR-id>` annotations in tests/evals (ADR-0014). Deterministic (no
// timestamps) so `--check` can fail CI when the committed artifacts are stale.
//
//   node scripts/gen-traceability.mjs          # write matrix + trace.json
//   node scripts/gen-traceability.mjs --check  # fail (exit 1) if they'd change
import { readFileSync, writeFileSync, readdirSync, statSync, mkdirSync, existsSync } from 'node:fs'
import { join, relative } from 'node:path'

const ROOT = process.cwd()
const REQS = join(ROOT, 'docs/requirements.md')
const MATRIX = join(ROOT, 'docs/qa/requirements-traceability-matrix.md')
const TRACE = join(ROOT, 'trace/trace.json')

const ROOT_DIR = process.cwd()
const BASELINE = join(ROOT_DIR, 'quality/trace-baseline.json')
const SCAN_DIRS = ['backend', 'frontend/src', 'mcp', 'evals']
// Only TEST/EVAL files can contribute coverage — a `@trace` in ordinary source
// (a comment or string) must not count as a proving test (review finding C4).
const TEST_FILE = /(_test\.go|\.test\.[jt]sx?|\.eval\.[jt]sx?)$/
const IGNORE = new Set(['node_modules', 'dist', '.git', 'vendor'])
const FR_RE = /FR-\d+/g

// 1) Requirements (id, tag, one-line title) in document order.
function parseRequirements() {
  const text = readFileSync(REQS, 'utf8')
  const out = []
  for (const line of text.split('\n')) {
    const m = line.match(/^\s*-\s*\*\*(FR-\d+)\*\*\s*\(([^)]+)\)\s*(.+?)\s*$/)
    if (m) out.push({ id: m[1], tag: m[2].trim(), title: m[3].trim() })
  }
  return out
}

// 2) Walk source dirs, collect FR-id -> sorted unique file paths.
function collectTraces() {
  const map = new Map()
  const walk = (dir) => {
    let entries
    try { entries = readdirSync(dir) } catch { return }
    for (const name of entries) {
      if (IGNORE.has(name)) continue
      const full = join(dir, name)
      const st = statSync(full)
      if (st.isDirectory()) walk(full)
      else if (TEST_FILE.test(name)) {
        const text = readFileSync(full, 'utf8')
        // A line mentioning @trace contributes every FR-id on that line, so a
        // single `// @trace: FR-010, FR-011` annotation covers several reqs.
        for (const line of text.split('\n')) {
          if (!line.includes('@trace')) continue
          for (const m of line.matchAll(FR_RE)) {
            const id = m[0]
            if (!map.has(id)) map.set(id, new Set())
            map.get(id).add(relative(ROOT, full))
          }
        }
      }
    }
  }
  for (const d of SCAN_DIRS) walk(join(ROOT, d))
  return map
}

function build() {
  const reqs = parseRequirements()
  const traces = collectTraces()

  const rows = reqs.map((r) => {
    const files = [...(traces.get(r.id) ?? [])].sort()
    return { ...r, files, covered: files.length > 0 }
  })
  const covered = rows.filter((r) => r.covered).length
  const mvp = rows.filter((r) => r.tag.toUpperCase().includes('MVP'))
  const mvpCovered = mvp.filter((r) => r.covered).length

  // Matrix markdown (generated — do not edit by hand).
  const lines = []
  lines.push('# Requirements traceability matrix')
  lines.push('')
  lines.push('> **Generated** by `scripts/gen-traceability.mjs` from `docs/requirements.md`')
  lines.push('> + `@trace <FR-id>` annotations. Do not edit by hand — CI fails if stale')
  lines.push('> (`node scripts/gen-traceability.mjs --check`). See [ADR-0014](../adr/0014-requirement-id-grammar.md).')
  lines.push('')
  lines.push(`Coverage: **${covered}/${rows.length}** functional requirements traced ` +
    `(MVP: **${mvpCovered}/${mvp.length}**).`)
  lines.push('')
  lines.push('| FR | Tag | Requirement | Traced by |')
  lines.push('|----|-----|-------------|-----------|')
  for (const r of rows) {
    const by = r.files.length ? r.files.map((f) => `\`${f}\``).join('<br>') : '⚠️ **untraced**'
    lines.push(`| ${r.id} | ${r.tag} | ${r.title.replace(/\|/g, '\\|')} | ${by} |`)
  }
  lines.push('')
  const matrix = lines.join('\n') + '\n'

  // trace.json (deterministic, sorted).
  const trace = {
    requirements: rows.length,
    covered,
    items: rows.map((r) => ({ id: r.id, tag: r.tag, covered: r.covered, tracedBy: r.files })),
  }
  const traceJson = JSON.stringify(trace, null, 2) + '\n'

  return { matrix, traceJson, mvp, mvpCovered }
}

const { matrix, traceJson, mvp, mvpCovered } = build()
const check = process.argv.includes('--check')

const covered = JSON.parse(traceJson).covered
const baseline = existsSync(BASELINE) ? JSON.parse(readFileSync(BASELINE, 'utf8')).minCovered : 0

if (check) {
  const cur = (p) => (existsSync(p) ? readFileSync(p, 'utf8') : '')
  if (cur(MATRIX) !== matrix || cur(TRACE) !== traceJson) {
    console.error('✗ traceability artifacts are stale — run: node scripts/gen-traceability.mjs')
    process.exit(1)
  }
  // Non-regressing coverage ratchet: traced count may only go up.
  if (covered < baseline) {
    console.error(`✗ traceability regressed: ${covered} traced < baseline ${baseline}`)
    process.exit(1)
  }
  console.log(`✓ traceability fresh; ${covered} FR traced (baseline ${baseline}); ` +
    `MVP ${mvpCovered}/${mvp.length}`)
} else {
  mkdirSync(join(ROOT, 'docs/qa'), { recursive: true })
  mkdirSync(join(ROOT, 'trace'), { recursive: true })
  mkdirSync(join(ROOT, 'quality'), { recursive: true })
  writeFileSync(MATRIX, matrix)
  writeFileSync(TRACE, traceJson)
  // Ratchet the baseline up to current coverage (never down).
  if (covered > baseline) {
    writeFileSync(BASELINE, JSON.stringify({ minCovered: covered }, null, 2) + '\n')
  }
  console.log(`✓ wrote matrix + trace.json; coverage ${covered}, baseline ${Math.max(covered, baseline)}`)
}
