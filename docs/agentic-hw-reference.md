# Agentic Greenfield HW — Reference Repo Study

Study of the three reference repos for the fwdays "Agentic Engineering: Greenfield"
homework, to mirror a full-marks + bonus submission. Date: 2026-07-01.

Sources:
- Task: `koldovsky/2026-fwdays-agentic-greenfield-task` (+ `-live` onboarded variant)
- Framework: `koldovsky/project-factory`
- Real run: `koldovsky/project-factory-test-run-claude-code`
- Reference submission: `shchadyloTaras/gitwarden`

---

## 1. TASK repo

Root tree: `.github/pull_request_template.md`, `.coderabbit.yaml`, `.gitignore`, `README.md`.
PR-template path is **`.github/pull_request_template.md`** (lowercase). No CI, no scaffold —
intentionally minimal (it just ships the CodeRabbit config + PR template).

### PR template (verbatim — fill ALL sections)

```
<!-- Домашнє завдання — Agentic Engineering: Greenfield.
     Заповни всі розділи. Стек — будь-який. -->

## Автор
<!-- Твоє справжнє імʼя -->


## Проєкт
<!-- 1–2 речення: що ти зробив(ла) і в якому стеку (будь-який) -->


## Відео-демо (1–2 хв)
<!-- Обовʼязково: посилання на YouTube / Loom / Google Drive з демонстрацією -->
Video:


## Які практики Agentic Engineering застосовано
<!-- Опиши КОНКРЕТНО:
     - контекст-інженерія (AGENTS.md / правила; статичний vs динамічний контекст)
     - цикли (loop engineering) замість покрокового промптингу
     - maker ≠ checker / суб-агенти / окреме рев'ю
     - верифікація: тести / evals / перевірки
     - специфікації наперед (SDD), якщо було
     - які інструменти та MCP використав(ла)
     - що вирішував(ла) ти, а що — агент -->


## (Опційно) Посилання на код
<!-- Якщо проєкт живе в окремому репозиторії — встав посилання -->


---

### Чекліст
- [ ] Вказано справжнє імʼя
- [ ] Додано посилання на відео-демо (1–2 хв)
- [ ] Описано застосовані практики Agentic Engineering
- [ ] Результат робочий і доведений до кінця
```

Five sections to fill: **Автор** (real name), **Проєкт** (1–2 sentences, any
stack), **Відео-демо** (required 1–2 min link), **Які практики Agentic Engineering
застосовано** (concrete: context-eng, loops, maker≠checker/sub-agents/separate
review, verification tests/evals, SDD specs, tools+MCP, you-vs-agent split),
**(Optional) code link**. 4-item checklist at the bottom.

### .coderabbit.yaml (key settings)

- `language: uk-UA` — reviews in Ukrainian, "course mentor" tone (friendly but demanding).
- `reviews.profile: assertive`; `request_changes_workflow: false` (advisory, non-blocking);
  `high_level_summary: true`; auto-review enabled, drafts excluded.
- Path-instructions apply to **all paths** and assess 4 proof points:
  (1) real name, (2) 1–2 min video link, (3) substantive description of agentic
  practices (context-eng, agent loops, maker≠checker, sub-agents, specs/tests/evals,
  verification, tools & MCP, problem scope), (4) deliverable finished end-to-end.
- Code review scope: substance only (logic errors, secrets, clarity); style left to linters.

### Grading (README)

Pass = real name ✅ + video ✅ + substantive agentic-practices description ✅ +
finished-not-abandoned ✅. **Bonus = visible artifacts**: rules/AGENTS.md, specs,
tests/evals, verification traces, separate review, demo recordings.
Process over product size.

### `-live` variant

Same README + same PR template + same `.coderabbit.yaml`. Difference: it is the
**project-factory-onboarded** version — ships the entire `.project-factory/` scaffold
(LOOP, MASTER-PROMPT, quality-gates, skills), `.claude/` (11 agents, 6 workflows,
5 opsx commands, openspec skills), `.cursor/` + `.codex/` + `.github/` multi-tool
prompts, a full **design-system** under `docs/design-system/` (tokens, DS components,
brand guidelines, prompt.md per component), `openspec/`, `evals/`, `trace/`,
`automations/`, and `scripts/check-*.mjs`. I.e. the live repo = the task scaffolded
through project-factory. Good blueprint for "what an onboarded repo looks like".

---

## 2. PROJECT-FACTORY (the framework)

Claude Code plugin (+ Cursor/Copilot/Codex) for multi-agent spec-driven delivery.
`/project-factory:init` (greenfield) or `:onboard` (existing). Three nested loops
(LOOP.md): inner = per-edit hooks (ESLint/tsc/secrets/trace/trailer); slice =
spec→tasks→tests(red)→implement(green)→battery→review-gate(maker≠checker)→archive;
outer = recordings→UAT triage→fix→regression→customer report. Gates **G0–G8** are
commands with exit codes, not prose.

### Canonical files it ships / generates

| Path | Contents |
|---|---|
| `AGENTS.md`, `CLAUDE.md` | Static context layer (CLAUDE.md references @AGENTS.md). |
| `LOOP.md` | The 3-loop rationale + loop-engineering primitive mapping + trace chain. |
| `MASTER-PROMPT.md` | Orchestrator phases 1–8. |
| `checklists/quality-gates.md` | **G0–G8** hard exit criteria; each gate = command set exiting 0. |
| `agents/*.md` (11) | requirements-analyst, spec-writer, capability-implementer, code-reviewer, security-reviewer, spec-compliance-auditor, test-engineer, eval-judge, vision-judge, qa-documenter, bug-triage-analyst. Maker≠checker: implementer never reviews. |
| `.claude/workflows/*.js` (6) | eval-suite, review-gate, spec-pipeline, trajectory-eval, uat-triage, vision-verify. |
| `commands/{init,onboard}.md` | Entry commands. |
| `templates/docs/*.template.md` | requirements, mvp-capability-plan, current-state, context-architecture, qa-pack. |
| `templates/adr.template.md` | MADR-style ADR (Status/Date/Deciders/Context/Decision/Alternatives table/Consequences). |
| `templates/openspec/change.template.md` | proposal.md + design.md + tasks.md skeletons per slice. |
| `templates/AGENTS.template.md` | Static-rules skeleton. |
| `templates/ci/*.yml`, `templates/hooks/*` | CI + git/Claude hooks (pre-commit, commit-msg, claude-code-hooks.json). |
| `evals/cases/*.eval.ts` + `evals/README.md` | Graded qualitative evals (rubric + @trace), eval ratchet. |
| `scripts/check-*.mjs` | traceability, coverage-ratchet, eval-ratchet, trajectory, recordings, a11y, qa-verify, record-demos, gate-status. |
| `skills/project-factory/SKILL.md` + references/ | Self-contained standalone-skill playbook. |
| `automations/registry.json` + `scripts/automations/` | Off-by-default scheduled watchers (drift/CI-triage/dep-audit), propose-only. |

### Requirements grammar

Numbered, never renumbered, one testable behavior each:
**FR-x** (functional), **NFR-x** (non-functional), **TC-x** (technical constraint),
**BC-x** (business constraint). Phase-tagged MVP/Future.

### Trace chain (the differentiator)

`FR-id → cited in spec.md → owned by exactly one slice in capability-plan →
`// @trace FR-id` in *.test.ts → proof in qa/**/manifest.json → `Slice:`/`Refs:`
commit trailers`. Validator `check-traceability.mjs` walks it; matrix is
**generated, never hand-written**; CI `--check-fresh` fails on stale.

---

## 3. PROJECT-FACTORY TEST-RUN (a real output) — "Weather Explorer"

Next.js app, 33 MVP FRs, 9 capabilities. The concrete shape of a graded repo:

- `docs/product-brief.md`, `docs/requirements.md` (FR/NFR/TC/BC), `docs/estimation.md`
- `docs/mvp-capability-plan.md`, `docs/context-architecture.md`, `docs/current-state.md`
- `docs/delivery-report.md` (exec summary, gates G0–G6 passed), `docs/project-factory-retrospective.md`
- `docs/adr/ADR-0001..0008` — incl. ADR-0002 context-architecture (static ≤4k tokens vs dynamic),
  ADR-0003 requirement-id-grammar, ADR-0005 traceability-gate-calibration, ADR-0006 automated-e2e-recordings
- `docs/technical/{architecture,data-model,integrations,operations,testing}.md`
- `docs/qa/` — **requirements-traceability-matrix.md** (human), traceability-report.md (generated),
  manual-test-plan.md, risk-register.md, mvp-acceptance-report.md, eval-report.md,
  trajectory-eval-report.md, global-review.md, ux-defects.md, demo-script.md,
  automated-verification-latest.md, and **demo-recordings/** (per-scenario .webm clip +
  .png frame + manifest.json)
- `openspec/` — `project.md`, `specs/<capability>/spec.md` (baseline) and
  `changes/archive/<date>-add-<cap>/{proposal,design,tasks,review-findings.json}.md`
  + `specs/<cap>/spec.md` delta per slice. review-findings.json carries `clean:true`.
- `evals/cases/*.eval.ts` (rubric-graded, @trace, judged by eval-suite; maker≠checker),
  `evals/results/*.json`, `quality/{coverage,eval}-baseline.json` (ratchets)
- `trace/{trace.json,trajectory.json}` (machine-readable)
- `tests/` co-located unit `.test.ts` per lib module + `tests/integration/` + `e2e/*.spec.ts`
- `.claude/{agents,commands/opsx,skills,workflows}`, `.githooks/{pre-commit,commit-msg}`,
  `.github/workflows/ci.yml`, `automations/registry.json`

Representative artifact contents confirmed: proposal.md cites FR ids + degradation
behavior; ADR uses MADR table; eval `.eval.ts` exports `cases[]` with `{id, trace[],
dimension, capability, scenario, produce(), rubric[]}` and is written against the
not-yet-built module (eval = the bar); matrix quotes `@trace` verbatim per FR row.

---

## 4. GITWARDEN (reference submission) — Electron Git GUI

A hand-rolled (not project-factory) but rigorously agentic submission. Tree:

- Root: `AGENTS.md` (SSOT), `CLAUDE.md` (@AGENTS.md), `WORKFLOW.md`, `DECISIONS.md`,
  `SECURITY.md` (threat model), `CONTRIBUTING.md`, `CHANGELOG.md`, `README.md`
- `docs/agentic-engineering.md` — **the "how I built this agentically" narrative**,
  structured exactly as the 5 PR-template proof points: Context engineering /
  Loops not step-by-step / Maker≠checker / Verification / Specs upfront (SDD)
- `docs/adr/0001..0009` (MADR) + `docs/adr/README.md`
- `docs/plans/*.md` (per-feature plans) + `docs/prompts/*.md` (per-phase prompts) — SDD track
- `docs/features/gitwarden/{spec.md,CONTEXT.md}`, `docs/progress-log.md` (authoritative Phase Checklist)
- `docs/architecture/` (excalidraw + png + svg diagram), `docs/roadmap.md`, `docs/release-checklist.md`
- `.claude/agents/{core-purity-reviewer,safety-reviewer}.md` — **2 reviewer sub-agents (maker≠checker)**
- `.claude/commands/{new-phase,verify-phase,log-phase,commit-phase,run-track,release}.md` — the **4-command loop**
- `.claude/hooks/*.sh` (commit-needs-log, core-purity, execfile-guard, no-global-git-config) — deterministic guards
- `.claude/skills/new-feature/` (SKILL.md + references)
- `.mcp.json` (MCP wired)
- Tests: `tests/unit/` (~70 files), `tests/integration/`, `tests/e2e/` (~29 specs),
  **`tests/evals/`** (golden-fixture AI evals, offline fake-adapter, `GITWARDEN_EVAL_LIVE=1` opt-in)
- CodeRabbit = the external maker≠checker pass on the PR

Loop (WORKFLOW.md): `/new-phase → implement → /verify-phase → review(sub-agent) →
/log-phase → /commit-phase`; phase not closed until exit criteria green; push always
manual. Seven always-valid invariants; never commit before tests green + checklist ticked.

---

## SYNTHESIS — what OUR submission needs

We reuse **WC-Tournament** (own repo, linked in PR). Since we're on Claude not Codex,
context lives in **CLAUDE.md** (gitwarden uses both; AGENTS.md is optional — a thin
AGENTS.md that points to CLAUDE.md is the portable move and costs nothing).

### Files/artifacts to add (paths relative to WC-Tournament repo root)

Context engineering
- `CLAUDE.md` — lean static rules (stack lock, module conventions, correctness/validation
  cadence, test-first, handoff protocol). Document a static-vs-dynamic boundary + token budget.
- (optional) `AGENTS.md` one-liner → `@CLAUDE.md` for tool portability.

Specs / SDD
- `docs/product-brief.md`, `docs/requirements.md` (FR/NFR/TC/BC numbered, MVP/Future tagged)
- `docs/mvp-capability-plan.md` (slice table + acyclic dep graph + FR-coverage table)
- `openspec/` OR `docs/features/<cap>/spec.md` — per-capability specs with proposal/design/tasks
- `docs/adr/ADR-0001..N.md` (MADR): at minimum stack, context-architecture, requirement-id-grammar,
  scoring/bracket-ordering decisions (we already have wc2026 memories for these)

Verification
- Frontend tests (currently **0** → must add): unit `.test.ts` per lib module, integration, e2e (Playwright)
- `evals/cases/*.eval.ts` — rubric-graded qualitative evals (e.g. leaderboard copy / prediction
  validation messages) with `@trace` ids; `evals/README.md`; an eval ratchet baseline
- `trace/trace.json` + a generated traceability matrix in `docs/qa/` (or a check-traceability script)

Maker ≠ checker
- `.claude/agents/<reviewer>.md` — ≥1 reviewer sub-agent (e.g. security/correctness reviewer);
  run an adversarial review pass, persist findings (review-findings.json, clean:true)
- CodeRabbit on the PR = external second checker (free on public repo)

Loop engineering
- `.claude/commands/*.md` for the phase loop (new-phase/verify-phase/log-phase/commit-phase) and/or
  `.claude/hooks/*` deterministic guards + `.github/workflows/ci.yml` running lint+test+build+checks
- `WORKFLOW.md` or `docs/agentic-engineering.md` documenting the loop

The narrative doc (highest ROI for the grade)
- `docs/agentic-engineering.md` — mirror gitwarden: one section per PR proof point
  (Context engineering / Loops / Maker≠checker / Verification / SDD / Tools & MCP /
  you-vs-agent split). This doubles as the body of the PR description.

Bonus engineering flourishes (optional, only if they land rigorously)
- WC-Tournament **MCP server** (read-only: fixtures/standings/leaderboard/my-predictions/bracket)
  built via `mcp-secure-server-dev`, with its own spec + tests + evals + `.mcp.json`
- Architecture diagram (svg) under `docs/architecture/` (use svg-diagram skill)
- Demo recordings (Playwright) under `docs/qa/demo-recordings/` with manifest.json

### Exact PR fields to fill (see verbatim template §1)
Автор (real name) · Проєкт (1–2 sentences, "WC-Tournament — …, stack …") ·
Відео-демо (1–2 min link — mandatory) · Які практики Agentic Engineering застосовано
(the 7 bullets from docs/agentic-engineering.md) · (optional) code link (WC-Tournament repo URL).
Tick all 4 checklist boxes. Write it in Ukrainian (CodeRabbit reviews in uk-UA).

### Submission mechanics (user-owned, no gh access)
Fork task repo → enable CodeRabbit on fork → push WC-Tournament work to a branch (or
link the separate repo) → open PR filling template → iterate on CodeRabbit feedback.
