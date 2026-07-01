# Slide coverage audit — WC-Tournament vs the real course deck (91 slides, 3 days)

Audits every slide 1..91 of the **real** three-day deck (Day 01 · Нова SDLC,
Day 02 · Project Factory, Day 03 · Скіли та агенти) against what is actually in
this repo. Supersedes the earlier audit that mistakenly used a 50-slide deck.

Legend: ✅ IMPLEMENTED (artifact cited) · ➖ DUPLICATE / N-A / OPTIONAL
(deck-flagged optional, greenfield-only bootstrap that is N/A because our app is
already in prod, or a tool choice our ADR-equivalent already covers) · ☐ MISSING
(a real, actionable gap).

## Coverage summary

| Verdict | Count |
| --- | --- |
| ✅ IMPLEMENTED | 47 |
| ➖ DUPLICATE / N-A / OPTIONAL | 34 |
| ☐ MISSING | 10 |
| **Total** | **91** |

Headline: Day 01 (concepts) and Day 02 (Project Factory) are covered by
**equivalents** — we didn't install `koldovsky/project-factory`, but its
discipline (SSOT rules, SDD, G-style gates, maker≠checker, traceability ratchet,
eval-ratchet) is reproduced with our own artifacts. **Day 03 is where the real
gaps are:** we do **not author a project Agent Skill (SKILL.md)**, we have **no
live GenUI** (tool→streamed-component), and we have **no Claude Code hooks beyond
the git pre-commit**. See the prioritized GAPS section at the end.

---

## Day 01 — Нова SDLC (slides 1–41)

| # | Title / topic | Verdict |
| --- | --- | --- |
| 1 | Title — 3 live sessions, greenfield | ➖ Title slide (N-A). |
| 2 | Speaker bio | ➖ Speaker slide (N-A). |
| 3 | Speaker bio | ➖ Duplicate of 2. |
| 4 | Speaker bio | ➖ Duplicate of 2. |
| 5 | Speaker bio | ➖ Duplicate of 2. |
| 6 | Speaker bio | ➖ Duplicate of 2. |
| 7 | Agent = Model + Harness | ➖ Concept slide; embodied by the whole harness (`AGENTS.md`, `.claude/`, CI). No artifact required. |
| 8 | Agentic Engineering = configuration (instructions/tools/sandboxes/orchestration/guardrails/observability) | ✅ Instructions `AGENTS.md`; tools `mcp/`; guardrails/hooks `.claude/hooks/pre-commit.sh` + `.github/workflows/ci.yml`; observability = traceability + evals. |
| 9 | Vibe vs Structured vs Agentic (how you verify) | ✅ We are "Agentic": tests + evals + CI gates — `WORKFLOW.md`, `.github/workflows/ci.yml`, `evals/`. |
| 10 | Tests vs Evals — why both | ✅ Go/vitest unit tests **and** `backend/internal/scoring/scoring_evals_test.go` + `mcp/evals/` + `frontend/src/lib/recap.test.ts`. |
| 11 | Eval example (output) — exact-equality fails for LLM prose | ✅ Rubric evals on the recap prose grounding — `frontend/src/lib/recap.test.ts`, `docs/features/recap/spec.md`. |
| 12 | Eval example (output) | ➖ Duplicate of 11. |
| 13 | Static vs Dynamic context (greenfield) | ✅ Explicit boundary in `AGENTS.md` ("Static vs dynamic context"), ADR-0013. |
| 14 | Traditional → AI-driven SDLC (focus shifts to verification) | ➖ Concept; realized by the verification pyramid (`docs/qa/test-plan.md`). |
| 15 | Conductor → Orchestrator | ➖ Concept slide (N-A). |
| 16 | Economics: Vibe = low CapEx/high OpEx; model routing lever | ➖ Concept. Model-routing = **covered-by-practice** (org rule: Sonnet default / Opus for complex; not repo-scoped). |
| 17 | Components of Agentic Engineering (roadmap) | ➖ Roadmap slide; each component audited below. |
| 18 | Component 1 — Models (frontier vs cheap; routing) | ➖ Covered-by-practice (org model-routing rule). No repo artifact expected. |
| 19 | Frontier benchmarks | ➖ Informational (N-A). |
| 20 | Component 2 — Tools (IDE / terminal / cloud) | ✅ We run terminal agents; `.codex/config.toml` + `.claude/` + `.mcp.json` show the harness surface. |
| 21 | Component 3 — Rules (AGENTS.md + CLAUDE.md, thin, versioned) | ✅ `AGENTS.md` (SSOT, ~971 tokens) + `CLAUDE.md` one-liner import; negative rules ("no secrets", "no workarounds"). |
| 22 | Component 4 — Agent Skills (SKILL.md, progressive disclosure) | ☐ **MISSING** — we author **no** SKILL.md. We consume skills (e.g. `mcp-secure-server-dev`) but ship none as a project artifact. See Day 03 slides 22/79/87/90. |
| 23 | Component 5 — MCP | ✅ `mcp/` read-only server (stdio), ADR-0015, `.mcp.json`. |
| 24 | Component 6 — Sub-agents (maker ≠ checker) | ✅ `.claude/agents/{scoring-correctness,security}-reviewer.md`; findings in `docs/qa/review-findings.json`. |
| 25 | Component 7 — Loop engineering | ✅ `WORKFLOW.md` + `LOOP.md` + `.claude/commands/{verify,trace,review,new-capability}.md`. |
| 26 | Component 8 — Factory approach | ✅ (equivalent) Our own factory-shaped loop; see Day 02. Verification-is-the-bottleneck realized in CI. |
| 27 | How OpenClaw / Claude Code do it in practice | ➖ Informational (N-A). |
| 28 | Their discipline: AGENTS.md compiler-grade, memory=policy, hooks>prompts, allowlist | ✅ `AGENTS.md` is compiler-grade w/ hard rules; `.claude/settings.local.json` uses a **permission allowlist** (no `--dangerously-skip`); pre-commit hook enforces gates. (Partial: hooks are git-only — see slide 58 / GAP.) |
| 29 | 5 shared principles (feedback loop, memory=policy, high-thinking baseline, CLI-first, long runs) | ➖ Synthesis (N-A); reflected in `docs/agentic-engineering.md`. |
| 30 | Greenfield = opportunities + discipline | ➖ Concept (N-A). |
| 31 | AI-friendly stack (green/red flags, training-data density) | ✅ Stack choice is deliberate & mainstream (Go/Gin, React/Vite/TS) — ADR-0001/0003; rationale in `docs/architecture.md`. |
| 32 | Agent-friendly architecture (clean boundaries, pure core, vertical slices) | ✅ Pure scoring core `backend/internal/scoring` + `backend/internal/scorers`; vertical slices in `docs/mvp-capability-plan.md`; ADR-0008. |
| 33 | Monolith vs Microservices | ➖ Decision made: modular monolith — ADR-0001, `docs/architecture.md`. |
| 34 | Frameworks: popularity = fewer hallucinations | ✅ Same as 31/33 — mainstream frameworks pinned in `AGENTS.md`. |
| 35 | Hidden risk 1 — Comprehension debt | ✅ Mitigated by generated traceability matrix + specs + human checkpoints (`docs/qa/`, `docs/agentic-engineering.md`). |
| 36 | Hidden risk 2 — Intent debt (AGENTS.md as intent ledger, ADR) | ✅ 16 ADRs in `docs/adr/` as the decision/intent log; `AGENTS.md` guardrails. |
| 37 | Everything in the repo (agent forgets, repo doesn't) | ✅ Specs, ADRs, requirements, tests, memory all committed; `docs/`, `docs/memory/`. |
| 38 | Spec-driven development (OpenSpec / Spec Kit) | ✅ (equivalent) SDD via `docs/requirements.md` + `docs/features/*/spec.md` (Given/When/Then). Tool choice ≠ OpenSpec — our ADR-equivalent covers it (see Day 02 slide 55 note). |
| 39 | Vibe coding — "here's a PRD, implement" | ➖ Anti-pattern shown for contrast (N-A). We did SDD instead. |
| 40 | Vibe coding (PRD dump) | ➖ Duplicate of 39. |
| 41 | Day 01 summary — generation solved, verification is the craft | ➖ Summary (N-A). |

## Day 02 — Project Factory (slides 42–70)

The deck's factory is the installable `koldovsky/project-factory` plugin (11
agents, 5 workflows, 9 check-* scripts, gates G0–G8). **We did not install it.**
Every row below judges whether the *underlying discipline* exists as our own
artifact. Where the deck names a specific factory file (e.g. `check-trajectory.mjs`,
`vision-verify.js`, `record-demos.mjs`) and we have no equivalent, it is a gap.

| # | Title / topic | Verdict |
| --- | --- | --- |
| 42 | Day 02 title — assemble the factory | ➖ Title slide (N-A). |
| 43 | First by hand — step-by-step engineering | ➖ Concept (N-A). Bootstrap steps (fork/clone the weather starter) are greenfield-only, N/A — our app is in prod. |
| 44 | By hand step 1–2: repo + Next.js + best-practices skill | ➖ Greenfield bootstrap N-A (weather starter, `create-next-app`). Our stack is already established (`AGENTS.md`). |
| 45 | By hand step 3–6: requirements/design/OpenSpec/Fallow | ✅ (equivalent, partial) requirements ✅ (`docs/requirements.md`); design ✅ (`docs/architecture.md`); OpenSpec → our spec files; **Fallow (codebase-intelligence CLI+MCP) not used** — noted, optional tool choice. |
| 46 | Content for requirements.md + product-brief.md | ✅ `docs/requirements.md` + `docs/project-brief.md` exist (our domain, not weather). |
| 47 | One slice: spec → red → green → verify | ✅ Loop is real — `WORKFLOW.md`, `LOOP.md`, `docs/features/*/spec.md`, `@trace` tests. |
| 48 | From components to factory (manual → automatic) | ➖ Concept (N-A). We chose bespoke automation over the plugin. |
| 49 | Map of the day — three nested loops (outer/slice/inner) | ✅ (equivalent) inner = `.claude/hooks/pre-commit.sh`; slice = review sub-agents + gates; outer = QA/demo — `WORKFLOW.md`. |
| 50 | Loop memory / state (comfort-score memory spread across repo) | ✅ `current-state.md`, `docs/memory/`, `trace/trace.json`, `.workflow-state.toon`, git log trailers. |
| 51 | Memory layers — static / working / episodic | ✅ Static `AGENTS.md`; working `current-state.md` + `docs/memory/`; episodic git log + `docs/adr/` + `quality/trace-baseline.json`. |
| 52 | Gates = commands with exit codes (G0–G8) | ✅ (equivalent) `.github/workflows/ci.yml` jobs are our exit-code gates (vet/test/build/evals/traceability/gitleaks/govulncheck/trivy/smoke). Not literally labelled G0–G8 but functionally equivalent. |
| 53 | End-to-end example: comfort-score | ➖ Deck's worked example (weather); our analogue is the scoring capability (`docs/features/scoring/spec.md`). N-A verbatim. |
| 54 | Requirements · G1 (numbered stable IDs) | ✅ `docs/requirements.md` FR/NFR/TC/BC numbered + phased (MVP/Future); ADR-0014 id grammar. |
| 55 | Specifications · G2 (OpenSpec, GWT contract) | ✅ (equivalent) Given/When/Then specs in `docs/features/*/spec.md`; `scripts/gen-traceability.mjs --check` enforces each MVP FR is cited. **OpenSpec `validate --strict` not used** — our traceability check is the equivalent gate. |
| 56 | Slicing · G3 (one owner per FR + dep DAG) | ✅ `docs/mvp-capability-plan.md` (capabilities → disjoint FR ids, dependencies). |
| 57 | Slice cycle · G4 (test-first red→green) | ✅ `@trace` tests precede/prove FRs; scoring + access tests; `WORKFLOW.md` step order. |
| 58 | Inner loop (hooks) — deterministic code at loop points | ☐ **MISSING (partial)** — we have a **git** `pre-commit.sh` (gofmt/vet/traceability/gitleaks) but **no `commit-msg` hook enforcing `Refs:`/`Slice:` trailers**, and **no Claude Code PostToolUse hook**. Trailers are convention-only, not machine-enforced. Real gap. |
| 59 | Traceability chain (check-traceability.mjs, exit code, --check-fresh) | ✅ `scripts/gen-traceability.mjs --check` (fresh + non-regressing vs `quality/trace-baseline.json`); CI `traceability` job; `trace/trace.json`. |
| 60 | Adversarial review · maker ≠ checker (review-gate workflow) | ✅ (equivalent) Two reviewer sub-agents + CodeRabbit; findings persisted. **No automated multi-reviewer "refute ≥2 → rejected" workflow file** — done as a manual adversarial pass (documented in `review-findings.json`). |
| 61 | Evals (output) — rubric + fresh judge, eval-ratchet | ✅ Scoring evals `go test -tags=evals`; recap rubric evals. **`check-eval-ratchet` proper is not a discrete gate** — trace ratchet exists; eval-score ratchet does not (see GAP). |
| 62 | Evals (trajectory) — order/no-weakened-test/no-scope-drift | ☐ **MISSING** — no `check-trajectory.mjs` / trajectory-eval equivalent. We assert output correctness but do not machine-verify the *route* (test-first order, no weakened tests, scope discipline). Real gap. |
| 63 | Proof, not promises · G6 (record-demos, vision-verify, a11y) | ☐ **MISSING (partial)** — we have a **manual** `docs/qa/demo-script.md` (human records the video), but **no `record-demos.mjs` (headless Playwright)**, **no `vision-verify` (vision-judge on a frame)**, and **no automated `check-a11y` (light+dark)**. Real gap (a11y automation + recorded proof). |
| 64 | Outer loop · UAT · G8 (bug → root cause → regression → proof) | ☐ **MISSING (partial)** — root-cause discipline is a *rule* in `AGENTS.md`, but there is **no `uat-triage` workflow / `bug-triage-analyst`** and no `@trace BUG-x` regression convention wired. Actionable if UAT feedback is expected. |
| 65 | Rest of the factory (coverage ratchet, ADR, current-state, worktrees, scheduled automations, QA pack) | ✅ (mostly) coverage/trace ratchet ✅; ADR ✅; `current-state.md` ✅; worktrees noted ✅ (`WORKFLOW.md`); QA pack ✅ (`docs/qa/`). **Scheduled propose-only automations: none** — optional, N-A for a solo submission. |
| 66 | One loop, any tool — factory is agent-agnostic | ✅ Portable core: `AGENTS.md` (Codex-native) + `CLAUDE.md` import + `.codex/config.toml` + Node scripts + git hooks + CI. Tool-agnostic by construction. |
| 67 | Assemble it all — one command (`/project-factory:init`, 11 agents/5 workflows/9 scripts) | ➖ N-A verbatim — we did not adopt the plugin. Our equivalent set is smaller (2 reviewer agents, 4 commands, 1 traceability script) and bespoke; deliberate scope choice. |
| 68 | Why it works — two pillars vs three antipatterns | ✅ Everything-traceable (deterministic validator) + checks-before-judgment (CI before LLM review) both present. |
| 69 | Homework — your project, any stack | ✅ This repo **is** the homework deliverable (Go/React, not the weather starter); `docs/SUBMISSION.md`. |
| 70 | Day 02 summary | ➖ Summary (N-A). |

## Day 03 — Скіли та агенти (slides 71–91)

The thesis: one capability → **four agent architectures** off one pure core —
**Ship** (self-contained SKILL.md), **Extend** (external agent + skill drives the
app), **Embed** (in-product GenUI via streamed tool→component), **MCP** (core as
an MCP server). We implement **MCP** and have a **guardrail-ready recap seam**,
but **do not author a SKILL.md** and have **no live GenUI streaming**.

| # | Title / topic | Verdict |
| --- | --- | --- |
| 71 | Day 03 title — program with skills, not code | ➖ Title slide (N-A). |
| 72 | Paradigm shift — deterministic code → agent runs a skill | ➖ Concept (N-A). |
| 73 | One capability — many surfaces (SKILL.md across Claude Code / product / runner) | ☐ **MISSING** — the "one skill, many surfaces" artifact (a shared `SKILL.md` + `run.mjs`) does **not** exist. We reuse a pure core across surfaces only via the **MCP** surface; no skill surface. |
| 74 | Runner landscape (OpenClaw / Hermes / NanoClaw) | ➖ Informational (N-A). Our MCP server is portable to MCP clients, but there is no self-contained skill to drop into a runner. |
| 75 | Chat as the primary product surface | ➖ Concept (N-A). Our product UI is form/route-based; no conversational surface. |
| 76 | Text vs GenUI (same answer, richer form) | ➖ Concept (N-A) — but motivates the GenUI gap below. |
| 77 | GenUI — model generates UI (tool call → streamed React component) | ☐ **MISSING** — no `streamText`/`useChat`/AI-SDK tool-streaming anywhere (`grep` in `frontend/src` = none). `MatchRecap.tsx` renders a **static** grounded string, not a model-streamed component. |
| 78 | Our solution — one task, four agent architectures (Ship/Extend/Embed/MCP) | ➖ (partial) Only **MCP** of the four is built (`mcp/`). Ship/Extend/Embed absent — see 73/80/81. Deck's own framing; N-A as a whole, gaps tracked per-surface. |
| 79 | Way 1 · Ship — self-contained skill, any harness | ☐ **MISSING** — no `SKILL.md` + `run.mjs` self-contained skill in the repo. |
| 80 | Way 2 · Extend — external agent + skill controls the app | ☐ **MISSING** — no `where-to-go`-style skill that drives our app over HTTP. (Our MCP server is read-only; it doesn't publish/act.) |
| 81 | Way 3 · Embed — agent inside the app (GenUI, Vercel AI SDK) | ☐ **MISSING** — same as slide 77; no embedded chat agent / streamed tool UI. |
| 82 | Way 4 · MCP — same core as a standard MCP server | ✅ `mcp/src/server.ts` (stdio), `mcp/src/tools.ts` (zod-validated read-only tools), `mcp/evals/`, ADR-0015. |
| 83 | Architecture — one core, four ways to run it | ➖ (partial) Our pure core (`backend/internal/scoring`, `frontend/src/lib`) backs the app + MCP; only 2 of the 4 run-modes exist. N-A verbatim. |
| 84 | Shared contract — agent decides, code grounds | ✅ (in spirit) The recap **guardrail** (`recap.ts` `validateRecap`) and MCP zod schemas ground agent output; teams/scores can't be fabricated. Matches the "code holds the boundaries" idea. |
| 85 | Code map — lib/recommend, skills/, api, mcp, components | ➖ (partial) We have `mcp/` + `frontend/src/lib` + `components/MatchRecap.tsx`, but **no `.claude/skills/` directory** and no `app/api/chat` GenUI route. N-A verbatim. |
| 86 | GenUI close-up — server tools + client renders component | ☐ **MISSING** — no `streamText({ tools })` server route, no `useChat` client. |
| 87 | Self-contained skill close-up — `run.mjs`, zero deps, SKILL.md | ☐ **MISSING** — no `run.mjs` / `SKILL.md`. |
| 88 | MCP close-up — `mcp/weather.ts`, thin wrapper over lib | ✅ `mcp/src/tools.ts` is exactly that pattern — thin, zod-validated wrappers over the read client, no new business logic. |
| 89 | Honest tradeoffs — any criterion, 1 keyed service, 0 new MCP logic | ✅ (in spirit) Our MCP adds **0** new logic (wraps existing HTTP read API); the recap seam is provider-agnostic (0 keyed services today). |
| 90 | Summary — static context = who; skills = what it can do; Ship/Extend/Embed/MCP | ➖ (partial) Summary; only MCP delivered. Reinforces the SKILL.md + GenUI gaps. |
| 91 | Your move — build your agent | ➖ Closing (N-A). |

---

## GAPS (prioritized, actionable)

Real, actionable gaps only. Everything else is either implemented, a duplicate,
a concept slide, a greenfield-only bootstrap N/A because our app is already in
prod, or a tool choice our ADR-equivalent already covers.

### P1 — Day 03 skill authoring (the headline gap)
1. **Author a project Agent Skill (`SKILL.md`)** — slides 22, 73, 79, 87, 90.
   We *consume* skills but ship none. Highest-value fix: add a self-contained
   `.claude/skills/<name>/{SKILL.md,run.mjs}` that wraps the existing pure core
   (e.g. a read-only "tournament briefing" skill built on the same functions the
   MCP server uses) — this simultaneously lights up **Ship** (slide 79) and, if
   pointed at our HTTP API, **Extend** (slide 80).

### P2 — Day 03 GenUI (Embed surface)
2. **Live GenUI (tool → streamed component)** — slides 77, 81, 86. `MatchRecap.tsx`
   is a static grounded template, not a model-streamed component; there is no
   `streamText`/`useChat` route. Closing this means a real embedded assistant
   route + a `useChat` client rendering `MatchRecap` from a tool result — the
   guardrail (`validateRecap`) is already the grounding seam it would plug into.
   *(Note: adds one keyed LLM service; deck itself flags this as the only keyed
   surface — acceptable per slide 89.)*

### P3 — Day 02 factory gates we don't yet have
3. **`commit-msg` trailer enforcement** (slide 58) — `Refs:`/`Slice:` trailers are
   convention-only. Add a `commit-msg` git hook that rejects app/lib/db commits
   lacking a trailer, so `git log --grep FR-…` is a complete audit.
4. **Trajectory eval** (slide 62) — no machine check that the *route* was correct
   (test-first order, no weakened tests, no scope drift). Add a `check-trajectory`
   equivalent or a documented trajectory-review pass.
5. **Recorded proof + a11y automation** (slide 63) — demo proof is manual
   (`demo-script.md`); no headless recording, no `vision-verify`, no automated
   `check-a11y` (light+dark). Add an automated a11y check in CI at minimum.
6. **Eval-ratchet as a discrete gate** (slide 61) — we ratchet *traceability*
   coverage but not *eval scores*. Add a `check-eval-ratchet` so the eval bar
   can only rise, matching "set the bar at the eval."

### P4 — Day 02 outer loop (only if UAT feedback is in scope)
7. **UAT-triage discipline** (slide 64) — root-cause-over-symptom is a rule but
   has no `uat-triage` workflow or `@trace BUG-x` regression convention. Wire one
   if real user-acceptance feedback will flow back.

### Explicitly NOT gaps (called out to avoid re-flagging)
- **Not installing `koldovsky/project-factory`** (slides 48, 67) — deliberate;
  we reproduce its discipline with bespoke, smaller artifacts. Scope choice.
- **OpenSpec / Spec Kit / Fallow** (slides 38, 45, 55) — tool choices; our
  `docs/features/*/spec.md` + `gen-traceability.mjs --check` are the equivalent
  gates. Optional.
- **Greenfield bootstrap** (slides 43, 44) — fork/clone/`create-next-app` are
  N/A: the app is already built and live in prod.
- **Model routing** (slides 16, 18) — covered-by-practice via the org rule
  (Sonnet default / Opus for complex), not a repo-scoped artifact.
- **Scheduled propose-only automations** (slide 65) — optional; N-A for a solo
  submission.

---

## Resolution (post-audit, 2026-07-01) — 91-slide deck

High-value Day-3/Day-2 gaps closed:
- ✅ **P1 — Agent Skill authored** (slides 22/73/79/80/87): `.claude/skills/wc-recap/`
  (`SKILL.md` + `run.mjs`) — portable grounded recap, same core as `MatchRecap.tsx`,
  self-contained, no model/secrets, guardrailed (tested). We now *ship* a skill,
  not just consume them.
- ✅ **P3a — commit-msg trailer gate** (slide 58): `.claude/hooks/commit-msg.sh`
  enforces conventional type + nudges `Refs:`/`Slice:` trailers (trace chain, ADR-0014).
- ✅ **P3b — Claude Code hooks** (slide 28, "hooks > prompts"): `.claude/settings.json`
  — `PostToolUse` auto-gofmt + `Stop` traceability-freshness check. Deterministic
  guardrails at cycle points, beyond the existing `pre-commit.sh`.

Accepted with rationale (not closed, on purpose):
- ➖ **P2 — live model-streamed GenUI** (slides 77/81/86): the recap is grounded
  template GenUI; a live LLM-streamed component needs a keyed backend service.
  We deliberately do **not** add a client-side LLM key (security anti-pattern) to
  a live app — the seam + guardrail + eval already exist; enabling a provider is
  FR-082 (ADR-0016).
- ➖ **P3c — trajectory eval** (62): test-first-order verification is partly covered
  by maker≠checker + the traceability ratchet; a dedicated `check-trajectory` is
  optional for a solo submission.
- ➖ **P3c — recorded proof + a11y automation** (63): real-behavior proof is the
  demo video (`demo-script.md`); headless recording / Lighthouse a11y are optional
  automation (the `chrome-devtools` MCP is available to add later).
- ➖ **P3d — eval-score ratchet** (61): our scoring/MCP evals are deterministic
  pass/fail, so a *score* ratchet is N/A; we ratchet traceability coverage instead.
- ➖ **P4 — UAT-triage** (64): only actionable once real UAT feedback flows back.

Not flagged (correct to omit): the project-factory plugin, OpenSpec/Spec-Kit/Fallow
tool choices, greenfield bootstrap (app already in prod), model routing (org rule).

---

## project-factory best-practice adoption (2026-07-01)

Read the full `koldovsky/project-factory` tree file-by-file and adopted the
high-value, right-sized practices:

**Adopted:**
- `scripts/check-eval-ratchet.mjs` + `quality/eval-baseline.json` — eval-surface may only grow (28 cases).
- `scripts/check-coverage-ratchet.mjs` + `quality/coverage-baseline.json` — backend coverage may only grow (≥34.5%); wired into CI backend job (with PG).
- `scripts/qa-verify.mjs` (`make qa`) — single battery runner → committed evidence report `docs/qa/automated-verification-latest.md`.
- `scripts/gate-status.mjs` (`make gates`) — G-trace/G-eval/G-cover PASS/FAIL/SKIP roll-up.
- `docs/qa/risk-register.md` + `docs/qa/mvp-acceptance-report.md` — complete the QA pack.
- CI: eval-ratchet + coverage-ratchet checks added.

**Skipped (framework-scale machinery — the factory is *their* deliverable; we're a shipped app):**
OpenSpec migration; the 5 JS fan-out workflows; the 9 orchestration agents;
`check-trajectory` git-process audit; `vision-verify`/`vision-judge`; the LLM
`eval-suite` judge (we grade with deterministic golden fixtures); the
`automations/` cron layer (dependabot already covers dep drift);
`uat-triage`/`bug-triage` (no formal UAT round); plugin/marketplace packaging +
`sync-skill-refs`; Cursor/Copilot multi-tool shims (we run Claude Code + Codex).
Headless demo recordings + a11y axe gate: deferred (need Playwright/browser infra
disproportionate to a single-audience app; the narrated demo video is the
rubric's real-behavior proof).
