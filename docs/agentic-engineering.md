# Agentic engineering — how WC-Tournament was built

This document maps 1:1 to the fwdays "Agentic Greenfield" proof points. It is the
evidence-of-process record for the homework and doubles as the PR narrative.
WC-Tournament is a production, live friends-only World Cup 2026 prediction pool
(Go + React + Postgres); the agentic engineering below is what makes "done" mean
*verifiably* done rather than "seems to work".

## 1. Context engineering (static vs dynamic)
- **Static, single source of truth:** [`AGENTS.md`](../AGENTS.md) — stack with
  **pinned versions**, repo layout, commands, and guardrails (no secrets; fix the
  root cause; scoring/access invariants need an ADR). Codex reads it natively;
  [`CLAUDE.md`](../CLAUDE.md) imports it via `@AGENTS.md`; `.codex/config.toml`
  documents it. **One ruleset, both tools, no drift** — a lesson learned when a
  Codex export once copied CLAUDE.md → AGENTS.md and immediately diverged
  ([ADR-0013](adr/0013-context-architecture.md)).
- **Dynamic, per-session:** [`docs/memory/`](memory/) (activeContext + progress),
  [`current-state.md`](../current-state.md) (a ~30-second handoff), and
  [`CHECKLIST.md`](../CHECKLIST.md) / `.workflow-state.toon` (loop state). The
  boundary is explicit so the always-loaded budget stays lean.
- **Context border & token economy (measured, not vibes):** a machine-enforced
  [`.aiexclude`](../.aiexclude) keeps generated/vendored/binary/secret files out
  of AI context. Measured with `wc`: the always-loaded static rules (`AGENTS.md`)
  are **≈971 tokens** — comfortably under the ≲4k budget ADR-0013 sets; the full
  post-border repo pack is **≈431k tokens across 239 files** (so we never blindly
  paste the whole tree — the border + the static/dynamic split are what keep the
  working context small).

## 2. Loop engineering (loops, not step-by-step prompting)
The work ran as a repeating cycle, not hand-held prompts — see
[`WORKFLOW.md`](../WORKFLOW.md) / [`LOOP.md`](../LOOP.md):
**spec → implement → trace → verify → review → commit → handoff.**
- [`CHECKLIST.md`](../CHECKLIST.md) is the live driver, re-checked in-loop after
  every change; nothing is "done" until its gate is green and its FR is traced.
- Reusable loop steps are `.claude/commands/` (`/verify`, `/trace`, `/review`,
  `/new-capability`) and a deterministic `.claude/hooks/pre-commit.sh` guard.
- The pattern is **Plan + Verify**: external checks gate each step, so quality
  comes from checkpoints, not from typing more prompts.

## 3. Maker ≠ checker (separate review)
- Two reviewer **sub-agents** in [`.claude/agents/`](../.claude/agents/)
  (`scoring-correctness-reviewer`, `security-reviewer`) run as a *separate
  context that did not write the code*.
- The adversarial pass **found and fixed real bugs** — a dead team-guardrail on
  the recap's live path, stage-digit leakage in number-grounding, and a
  traceability scanner that could be satisfied by a comment rather than a test.
  Findings are persisted in [`docs/qa/review-findings.json`](qa/review-findings.json).
- **CodeRabbit** is the external second checker on the PR (uk-UA, mentor tone;
  [`.coderabbit.yaml`](../.coderabbit.yaml) filters generated/scaffolding noise).

## 4. Verification (tests / evals / traces, not "seems to work")
A 5-layer pyramid ([`docs/qa/test-plan.md`](qa/test-plan.md)), all gated in CI:
1. static (gofmt/vet, tsc, actionlint, gitleaks, govulncheck, Trivy);
2. unit (Go + vitest); 3. integration (Go against a **real Postgres** service);
4. **evals** — golden-fixture quality bars: scoring
   (`backend/internal/scoring/scoring_evals_test.go`) and the MCP tools
   (`mcp/evals/`); 5. real-behavior proof — the demo video.
- **Traceability:** every FR is proven by a `@trace <FR-id>` test; the matrix
  ([`docs/qa/requirements-traceability-matrix.md`](qa/requirements-traceability-matrix.md))
  is **generated** by `scripts/gen-traceability.mjs` and CI fails if it is stale
  or coverage regresses below `quality/trace-baseline.json`
  ([ADR-0014](adr/0014-requirement-id-grammar.md)). Current: 17/19 FR (16/17 MVP).

## 5. Spec-driven development (SDD)
Specs precede code: [`docs/product-brief.md`](project-brief.md) →
[`docs/requirements.md`](requirements.md) (FR/NFR/TC/BC, numbered, MVP/Future) →
[`docs/mvp-capability-plan.md`](mvp-capability-plan.md) (each capability owns
disjoint FR ids) → per-capability Given/When/Then specs in
[`docs/features/`](features/) → 16 ADRs in [`docs/adr/`](adr/). New behavior
starts from a spec + FR id, then a test that `@trace`s it.

## Tools & MCP (bonus)
- A read-only **MCP server** ([`mcp/`](../mcp/), [ADR-0015](adr/0015-mcp-read-only-tools.md))
  exposes the pool to agents (fixtures/standings/leaderboard/bracket/player
  predictions) — built security-first (zod-validated, no secrets, reveal-lock),
  with its own evals.
- A **generative-UI** AI match recap ([`frontend/src/lib/recap.ts`](../frontend/src/lib/recap.ts),
  [ADR-0016](adr/0016-ai-recap-grounded-guardrail.md)): grounded generation
  behind an anti-hallucination guardrail — it cannot render a wrong score/team,
  and a future LLM provider plugs in *behind* the same guardrail.

## What I decided vs what the agent did
- **I (human) decided:** the product and scope, the invariants (scoring rules,
  demo-mode privacy model, manual-override precedence), which bonus components to
  build (MCP first, then a guarded GenUI recap), and every accept/fix triage on
  review findings.
- **The agent did:** research (course conventions, reference repos), drafting the
  specs/ADRs/tests/evals/CI from those decisions, the implementation diffs, the
  traceability tooling, and — as a *separate* agent — the adversarial review that
  caught the bugs above. Tools used: Claude Code with sub-agents, the
  `mcp-secure-server-dev` / `svg-diagram` / bmad-review skills, and workflow-memory
  for loop state.
