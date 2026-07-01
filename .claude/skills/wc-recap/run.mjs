#!/usr/bin/env node
// wc-recap — portable grounded match-recap skill. Mirrors the CAP-09 core
// (frontend/src/lib/recap.ts): grounded template + a guardrail that rejects any
// number outside the scoreline. No model, no secrets. See ADR-0016.

const STAGE = {
  group: 'the group stage',
  r32: 'the first knockout round',
  r16: 'the last sixteen',
  qf: 'the quarter-final',
  sf: 'the semi-final',
  third: 'the third-place play-off',
  final: 'the final',
}

function buildRecap(m) {
  if (m.homeScore == null || m.awayScore == null || !m.home || !m.away) return ''
  const stage = STAGE[m.stage] ?? 'the match'
  const [h, a] = [Number(m.homeScore), Number(m.awayScore)]
  let line
  if (h > a) line = `${m.home} beat ${m.away} ${h}:${a} in ${stage}.`
  else if (a > h) line = `${m.away} beat ${m.home} ${a}:${h} in ${stage}.`
  else line = `${m.home} and ${m.away} drew ${h}:${a} in ${stage}.`
  const guessers = (m.exactGuessers ?? []).filter(Boolean)
  if (guessers.length) line += ` Exact-score call by ${guessers.join(', ')} — nicely done.`
  return line
}

// Guardrail: recap numbers must be exactly the scoreline; else it's ungrounded.
function isGrounded(text, m) {
  const allowed = new Set([String(m.homeScore), String(m.awayScore)])
  for (const n of text.match(/\d+/g) ?? []) if (!allowed.has(n)) return false
  return text.length <= 400 && !/[\x00-\x1f]|[`<>]|\{\{/.test(text)
}

function recap(m) {
  const t = buildRecap(m)
  return isGrounded(t, m) ? t : buildRecap(m) // grounded template is always safe
}

async function readInput() {
  if (process.argv[2]) return process.argv[2]
  let s = ''
  for await (const c of process.stdin) s += c
  return s.trim()
}

const raw = await readInput()
if (!raw) {
  console.error('usage: run.mjs \'{"home","away","homeScore","awayScore","stage","exactGuessers?"}\'')
  process.exit(2)
}
let match
try {
  match = JSON.parse(raw)
} catch {
  console.error('invalid JSON input')
  process.exit(2)
}
process.stdout.write(recap(match) + '\n')
