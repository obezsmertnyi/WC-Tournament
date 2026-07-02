# LOOP — the tight cycle

Quality here comes from **deterministic loops, not prompts**: every requirement
is traced, every gate is an exit code, and no agent grades its own work. This
file is the *map* of that loop system — the mechanics live in the files it links
to (`WORKFLOW.md`, `docs/qa/quality-gates.md`, `AGENTS.md`), not duplicated here.

## Three nested loops

```
 ┌─ OUTER · tournament / feedback (days) ─────────────────────────────┐
 │  FIFA results sync → score predictions → players report → new spec  │
 │  ┌─ SLICE · per capability (the core cycle) ────────────────────┐   │
 │  │  spec → implement → trace → verify → review → commit          │   │
 │  │  ┌─ INNER · per edit (seconds) ─────────────────────────┐     │   │
 │  │  │  gofmt on write · pre-commit: vet + trace + gitleaks  │     │   │
 │  │  └───────────────────────────────────────────────────────┘     │   │
 │  └───────────────────────────────────────────────────────────────┘   │
 └─────────────────────────────────────────────────────────────────────┘
```

- **Inner — per edit (seconds).** On every file write `gofmt` auto-runs
  (PostToolUse hook, `.claude/settings.json`). At commit, `.claude/hooks/pre-commit.sh`
  runs gofmt + `go vet` + the traceability freshness check + `gitleaks`, and
  `commit-msg.sh` checks the message type and nudges `Slice:`/`Refs:` trailers.
  A red hook stops the commit.
- **Slice — per capability.** The core cycle below: one PR = one capability.
- **Outer — tournament / feedback (days).** Live results sync from the FIFA API
  score predictions; players surface issues; a fix re-enters the slice loop as a
  new spec / regression. UAT is currently **informal** (ad-hoc reports) — gate G8
  in `docs/qa/quality-gates.md`.

## The slice loop

```
        ┌──────────────────────────────────────────────────────────┐
        │                                                          │
        ▼                                                          │
   ┌─────────┐   ┌───────────┐   ┌────────┐   ┌────────┐   ┌──────────┐
   │  SPEC   │──▶│ IMPLEMENT │──▶│ TRACE  │──▶│ VERIFY │──▶│  REVIEW  │
   │ FR + GWT│   │ smallest  │   │ @trace │   │ gates  │   │ maker ≠  │
   │  + ADR  │   │  change   │   │  FR-id │   │  green │   │ checker  │
   └─────────┘   └───────────┘   └────────┘   └────────┘   └────┬─────┘
        ▲                                          │ fail       │ findings
        │                                          └────────────┤
        │                                                       ▼
        │                                                  ┌─────────┐
        └──────────────── handoff (current-state) ─────────│ COMMIT  │
                                                           │ + trailers
                                                           └─────────┘
```

Repeat per capability slice. Never advance past a red gate; never mark a
requirement done until a `@trace`'d test proves it and a separate checker has
looked. Full detail in `WORKFLOW.md`; live status in `CHECKLIST.md`.

Commands (`.claude/commands/`): `/verify` (run the gates), `/trace` (regenerate +
check the traceability matrix), `/review` (dispatch the reviewer sub-agents),
`/new-capability` (scaffold a spec + FR ids).

## Hard rules (non-negotiable)

1. **Traceability** — every FR id → spec → `@trace`'d test → a **generated**
   matrix (`scripts/gen-traceability.mjs`; CI `--check` fails on stale/regressed).
   The matrix is generated, never hand-written.
2. **Deterministic checks before judgment** — hooks + CI (gofmt, `go vet`, `tsc`,
   gitleaks, govulncheck, Trivy) go green *before* any LLM review runs.
3. **Ratchets only tighten** — backend coverage (`scripts/check-coverage-ratchet.mjs`)
   and per-suite eval counts (`scripts/check-eval-ratchet.mjs`) may grow, never
   silently loosen. Baselines in `quality/`.
4. **Maker ≠ checker** — the implementer never reviews their own slice; fresh
   reviewer agents (`.claude/agents/scoring-correctness-reviewer.md`,
   `security-reviewer.md`) + CodeRabbit do. Findings land in
   `docs/qa/review-findings.json` and block until critical/high are fixed.
5. **Honest reporting** — a red gate is a **STOP**; fixing the check is allowed,
   weakening or bypassing it is not (`docs/qa/quality-gates.md`). No unverified
   success claims.
6. **Fix the root cause** — never disable/skip a failing gate or write a test that
   asserts buggy behavior; bump/patch and prove it green (`AGENTS.md`).

## Anti-patterns this loop counters

- **Self-grading** → structural maker ≠ checker + deterministic gates that can't
  be argued with.
- **Comprehension debt** → generated traceability matrix + Given/When/Then specs
  + ADRs keep intent legible instead of buried in diffs.
- **Cognitive surrender** → red gates fail loudly and mean STOP; no narrating
  around a failure to keep moving.

## Deliberately out of scope

Kept lean vs the project-factory template — this is a friends PoC, not an
agent factory. Not implemented (and intentionally so; see
`docs/qa/slide-coverage-audit.md`): a `check-trajectory` process eval, a
`uat-triage` agent + `@trace BUG-x` regression convention, scheduled
propose-only automations, and a conductor/orchestrator mode split. Revisit if the
project outgrows a PoC.
