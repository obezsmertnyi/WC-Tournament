#!/usr/bin/env node
// Doc-graph integrity: starting from the always-loaded bootstrap chain
// (CLAUDE.md → AGENTS.md → LOOP.md → …), verify every pointer to an OUR-repo path
// resolves. A dangling pointer = a dead end a fresh session would hit. A few
// external tooling paths are expected to be absent and are not flagged (REF_PREFIX).
//
//   node scripts/check-doc-links.mjs           # report
//   node scripts/check-doc-links.mjs --check   # exit 1 on genuine drift (gate mode)
import { readFileSync, existsSync, statSync, readdirSync } from 'node:fs'
import { resolve, dirname, join, basename } from 'node:path'

const REPO = process.env.CLAUDE_PROJECT_DIR || process.cwd()
const CHECK = process.argv.includes('--check')

const seed = [
  'CLAUDE.md', 'AGENTS.md', 'LOOP.md', 'WORKFLOW.md',
  'docs/REPO-MAP.md',
  '.claude/commands/verify.md', '.claude/commands/trace.md',
  '.claude/commands/review.md', '.claude/commands/new-capability.md',
].filter((f) => existsSync(join(REPO, f)))

// External/tooling paths intentionally absent from this repo — not flagged as broken.
const REF_PREFIX = []
const isRef = (p) => REF_PREFIX.some((r) => p === r || p.startsWith(r))
const asCmd = (t) => `.claude/commands/${t.slice(1)}.md`

const IGNORE = new Set(['node_modules', '.git', 'dist', 'build', '.next'])
const byBase = new Map()
;(function walk(dir) {
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    if (IGNORE.has(e.name)) continue
    const full = join(dir, e.name)
    if (e.isDirectory()) walk(full)
    else { const a = byBase.get(e.name) || []; a.push(full); byBase.set(e.name, a) }
  }
})(REPO)

function expandBraces(s) {
  const m = s.match(/\{([^}]*)\}/)
  if (!m) return [s]
  return m[1].split(',').flatMap((o) => expandBraces(s.replace(m[0], o)))
}

function refsIn(file) {
  const txt = readFileSync(join(REPO, file), 'utf8')
  const out = new Set()
  for (const m of txt.matchAll(/\]\(([^)]+)\)/g)) out.add(m[1]) // md links
  for (const m of txt.matchAll(/`([^`]+)`/g)) out.add(m[1]) // inline code
  return [...out]
    .map((s) => s.trim().split('#')[0])
    .filter(Boolean)
    .filter((s) => !/^https?:\/\//.test(s) && !s.startsWith('@'))
    .flatMap((s) => (/^\/[a-z-]+$/.test(s) ? [asCmd(s)] : expandBraces(s)))
    .filter((s) => s.includes('/') || /\.(md|mjs|sh|json|go|ts|tsx|yaml|yml|sql)$/.test(s))
    .filter((s) => !s.includes(' ') && s !== '/')
    .filter((s) => !/[<>$*…]/.test(s) && !/\d–\d/.test(s)) // placeholders/ranges/ellipsis
    .filter((s) => !/^(git|node|go|make|docker|npm|cd|sudo|echo|pgx)\b/.test(s))
}

function resolves(from, p) {
  const bases = [dirname(join(REPO, from)), REPO, join(REPO, 'scripts'),
    join(REPO, 'docs/qa'), join(REPO, 'docs'), join(REPO, 'backend'), join(REPO, 'frontend')]
  for (const b of bases) {
    const c = resolve(b, p)
    if (existsSync(c) && c.startsWith(REPO)) return true
  }
  if (byBase.has(basename(p))) return true // repo-wide basename fallback
  if (/\/adr\/\d{4}$/.test(p)) { // ADR-number shorthand → 0013-*.md
    const n = basename(p)
    if ([...byBase.keys()].some((k) => k.startsWith(n + '-'))) return true
  }
  return false
}

// A dangling ref WITH a file extension (docs/foo.md, scripts/x.mjs) is an
// unambiguous broken pointer → gated. A dangling ref without one (adr/, internal/api)
// is directory/prose shorthand → advisory only, never fails the gate.
const hasExt = (s) => /\.(md|mjs|sh|json|go|ts|tsx|yaml|yml|sql)$/.test(s)
const gated = [], advisory = [], ref = []
const seen = new Set()
let okCount = 0
for (const f of seed) {
  for (const p of refsIn(f)) {
    if (seen.has(p)) continue; seen.add(p)
    if (resolves(f, p)) { okCount++; continue }
    if (isRef(p)) ref.push(p)
    else if (hasExt(p)) gated.push({ p, f })
    else advisory.push({ p, f })
  }
}

console.log(`doc-graph: ${okCount} resolved · ${ref.length} reference-crosswalk · ${gated.length} broken file-pointer(s) · ${advisory.length} advisory`)
for (const d of gated) console.log(`  ✗ ${d.p}   (in ${d.f})`)
for (const d of advisory) console.log(`  ~ ${d.p}   (in ${d.f}) [advisory: dir/shorthand]`)
if (gated.length && CHECK) {
  console.error('\n✗ broken file-pointer(s) in the bootstrap doc-graph — fix the reference or the target.')
  process.exit(1)
}
if (!gated.length) console.log('✓ no broken file-pointers — the doc-graph is intact.')
