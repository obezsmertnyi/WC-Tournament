# Agentic Engineering — submission checklist

This is the **loop driver** for taking WC-Tournament through the full agentic
engineering cycle (fwdays "Greenfield" homework). Each item maps to a graded
proof point. It is re-checked in-loop after every change: build → verify →
separate review → tick. `✅ done · 🔄 in progress · ☐ todo`.

> Rules live in **`AGENTS.md`** (single source of truth — Codex reads it
> natively); `CLAUDE.md` imports it via `@AGENTS.md` and `.codex/config.toml`
> documents it for Codex. One ruleset, both tools, no drift.

## Pass/fail (rubric)
- ✅ Real author name in the PR (`docs/SUBMISSION.md` — confirm spelling)
- ☐ 1–2 min demo video linked in the PR — *user records* (`docs/qa/demo-script.md`)
- ✅ Substantive description of agentic practices (`docs/agentic-engineering.md` + uk PR body)
- ✅ Deliverable finished and working (live in prod; CI green; 15 FE + Go tests + evals)

## Proof point 1 — Context engineering (static vs dynamic) ✅
- ✅ `AGENTS.md` — SSOT static rules: stack + pinned versions, guardrails, static/dynamic boundary
- ✅ `CLAUDE.md` — imports AGENTS.md (`@AGENTS.md`); `.codex/config.toml` — Codex pointer (cross-tool, no drift)
- ✅ `docs/memory/activeContext.md`, `docs/memory/progress.md` — dynamic context
- ✅ `current-state.md` — per-slice handoff

## Proof point 2 — Loop engineering (loops, not step prompting) ✅
- ✅ `WORKFLOW.md` + `LOOP.md` — the build→verify→review→commit loop (+ ASCII loop)
- ✅ `.claude/commands/{verify,trace,review,new-capability}.md` — phase loop commands
- ✅ `.claude/hooks/pre-commit.sh` — deterministic fast-gate guard
- ✅ `.github/workflows/ci.yml` — automated gate (+ traceability job, evals, vitest)

## Proof point 3 — Maker ≠ checker (separate review) ✅
- ✅ `.claude/agents/{scoring-correctness,security}-reviewer.md` — reviewer sub-agents
- ✅ adversarial review pass (2 separate agents) → `docs/qa/review-findings.json`; found+fixed C1/C2/C4/C5
- ☐ CodeRabbit enabled on the fork (external 2nd checker) — *user action*
- ⚠️ **S1: rotate the real secrets in `.env`** (gitignored, not committed — but exposed to review tooling) — *owner action*

## Proof point 4 — Verification (the 5-layer pyramid) 🔄
- ✅ Typecheck + lint (gofmt/vet, tsc, actionlint) in CI
- ✅ Unit tests — Go (22 tests)
- ✅ Unit tests — frontend (vitest: access@FR-031/032, bracket) — wired in CI
- ✅ Integration tests — Go against real Postgres (CI service)
- ✅ `evals/` — scoring golden-fixture evals (rubric + `@trace`), run in CI
- ✅ `docs/qa/` — generated traceability matrix + `trace/trace.json` + `--check` ratchet (CI job)
- ✅ `docs/qa/test-plan.md` + `demo-script.md`
- ☐ Real-behavior proof — the demo video (*user records*, see `demo-script.md`)

## Proof point 5 — Spec-driven development (SDD) ✅
- ✅ `docs/project-brief.md` (product brief)
- ✅ `docs/architecture.md` + `docs/diagrams/*.svg`
- ✅ `docs/adr/0001–0012` (MADR)
- ✅ `docs/requirements.md` — FR/NFR/TC/BC, numbered, MVP/Future
- ✅ `docs/mvp-capability-plan.md` — capabilities → slices owning FR-ids
- ✅ `docs/features/{predictions,scoring,demo-access,mcp}/spec.md` — Given/When/Then
- ✅ `docs/adr/0013` context-arch · `0014` requirement-id grammar · `0015` MCP

## Tools & MCP (bonus — visible engineering) ✅
- ✅ WC-Tournament read-only **MCP server** (`mcp/`): list_fixtures/group_standings/leaderboard/bracket/player_predictions
- ✅ MCP spec (`docs/features/mcp/spec.md`) + ADR-0015 + `.mcp.json` + README; zod-validated, read-only, no secrets, reveal-lock
- ✅ MCP evals (`mcp/evals/`, 10 tests) traced FR-070/FR-071; typecheck green

## GenUI (bonus) — AI match recap ✅
- ✅ `frontend/src/lib/recap.ts` — grounded recap generator + fact-grounding guardrail (provider-agnostic)
- ✅ `frontend/src/components/MatchRecap.tsx` — GenUI panel (rendered in reveal, post-kickoff)
- ✅ recap evals (`recap.test.ts`, 8 tests) — guardrail rejects hallucinated score/team; traced FR-080/FR-081
- ✅ `docs/features/recap/spec.md` + ADR-0016 (grounded generation behind a guardrail)

## Narrative & submission ✅
- ✅ `docs/agentic-engineering.md` — one section per proof point (EN, doubles as reference)
- ✅ `docs/SUBMISSION.md` — uk PR body (verbatim template) + handoff steps (fork/CodeRabbit/video/PR)
