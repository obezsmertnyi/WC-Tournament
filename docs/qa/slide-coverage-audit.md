# Slide-by-slide coverage audit — fwdays "Agentic Greenfield" deck vs WC-Tournament

Audits the public course deck, slide by slide, against what this repo implements.

## Source of truth for the slides

- **Deck used:** `github.com/koldovsky/2026-agentic-greenfield-vs-brownfield`
  (public), branch `main`. 50 slides = cover (`slides.md`) + 49 pages in
  `pages/act1..act6`, one file per slide, in the order wired by `slides.md`.
- **Note on the other URL:** the task also pointed at
  `koldovsky/2026-fwdays-agentic-greenfield-slidev`. That repo (and its
  `raw…/main/slides.md`) returns **HTTP 404** — it is private or does not exist,
  so it could not be read. The public `…greenfield-vs-brownfield` deck is the
  one the task itself flags as "the known-public deck source: ~50 slides, 6 acts,
  pages/ as markdown", and its content matches. The rendered SPA at
  `koldovsky.github.io/2026-fwdays-agentic-greenfield-slidev/…#/day02` is a
  *different* (brownfield-only) course; its resources slide (49) links it as a
  separate follow-on workshop. This audit covers the greenfield-vs-brownfield deck
  in full; if a private 100+-slide fwdays deck exists it was not reachable.
- **Framing that drives most verdicts:** WC-Tournament was **already a
  production app** before this homework (M0–M9 shipped, live in prod). The
  agentic homework *augments* it. So the deck's **greenfield-only, one-time
  bootstrap** slides (Phase 0, `create-next-app`, etc.) are legitimately
  **N/A / duplicate** for us — the deck's own Components-matrix (slide 33) marks
  them "⚪️ N/A" for brownfield. Where a slide's *underlying practice* applies to
  any mature project, we hold ourselves to it.

Legend: ✅ IMPLEMENTED · ➖ DUPLICATE / NOT-NEEDED · ☐ MISSING (real gap)

---

## Act 1 — Setup (slides 1–8)

| # | Slide | Practice / artifact implied | Verdict |
|---|-------|------------------------------|---------|
| 1 | Cover — Agentic Engineering: Greenfield vs Brownfield | Framing only | ➖ Title slide, no artifact. |
| 2 | About the speaker | — | ➖ Speaker bio; N/A. |
| 3 | Current state of AI in dev (May 2026) | Context/stat framing | ➖ Motivational stats; no artifact. |
| 4 | Bottleneck shifted → verification loop + strategic thinking | Verification-loop + planning as the real bottleneck | ✅ Whole loop is built for this: `WORKFLOW.md` / `LOOP.md` (SPEC→IMPLEMENT→TRACE→VERIFY→REVIEW), 5-layer gate in CI. |
| 5 | The myth "AI only works on greenfield" | Framing | ➖ Rhetorical; no artifact. |
| 6 | Greenfield vs Brownfield: definitions | Framing | ➖ Definitions; our project is the brownfield case (already in prod). |
| 7 | Why AI behaves differently (context extraction) | Context engineering matters | ✅ Embodied by `AGENTS.md` SSOT + `docs/memory/` static/dynamic split. |
| 8 | Talk map | Framing | ➖ Roadmap slide. |

---

## Act 2 — Greenfield flow (slides 9–20)

| # | Slide | Practice / artifact implied | Verdict |
|---|-------|------------------------------|---------|
| 9 | Divider — Greenfield | — | ➖ Section divider. |
| 10 | Mental model: speed = inference + hard thinking; **give the agent a way to verify its work** | A machine-checkable verification path | ✅ `make ci`, `go test -tags=evals`, vitest, `gen-traceability --check`, container smoke — the agent can self-verify every layer. |
| 11 | Greenfield workflow: 9 phases (0 bootstrap … 8 parallel worktrees) | A named, repeatable phase model incl. **Phase 8 parallel git worktrees** | ✅ *for the loop* (phases map to `WORKFLOW.md`). ☐ *for Phase 8 only* — see GAPS: no documented parallel-worktree practice. Phases 0/3/4 (bootstrap/PRD/DESIGN) are greenfield-one-time → ➖ (covered below). |
| 12 | Phase 0: deterministic bootstrap (`create-next-app`, pinned versions) | Human-owned deterministic scaffold before the agent starts | ➖ NOT-NEEDED — greenfield-only ("⚪️ N/A" on slide 33). Our stack was scaffolded pre-homework; equivalent intent (pinned versions, no hallucinated structure) is met by pinned stack in `AGENTS.md` + `go.mod`/`package.json`. |
| 13 | Phase 1: MCP servers as baseline (`.mcp.json` shared in git) | A committed `.mcp.json` so the team shares tools | ✅ `.mcp.json` at repo root (our own `wc-tournament` read-only server). Note: deck lists `context7`/`vercel`/`chrome-devtools` as *examples*; we ship our project MCP instead, which is the stronger signal. |
| 14 | Capability slicing — vertical slice (UI+API+domain+DB+tests+docs), 1 PR = 1 capability = 1 demo | Vertical capability slices | ✅ `docs/mvp-capability-plan.md` (CAP-01…CAP-09, each owns disjoint FR ids); "one PR = one capability = one demo" is stated in `mvp-capability-plan.md` + `WORKFLOW.md`. |
| 15 | Optimal slice size (1–3 days; 5–7 caps per MVP) | Sizing discipline | ✅ Reflected in the CAP plan (5 spec'd caps on the submission surface); guidance, not a separate artifact. |
| 16 | Antipattern: "build the whole MVP in one prompt" | Avoid big-bang generation | ✅ Avoided by design — SDD + per-capability specs + loop; documented in `docs/agentic-engineering.md`. |
| 17 | OpenSpec cycle: proposal → design → tasks → **spec.md (Given/When/Then)** → impl → lint/test/build → validate/archive → update handoff | Spec-before-code, per-slice cycle with a persisted handoff | ✅ Equivalent implemented **without** the OpenSpec tool: `docs/features/<cap>/spec.md` (Given/When/Then) + `docs/requirements.md` (FR/NFR) + `docs/adr/` + `current-state.md` handoff. `.claude/commands/new-capability.md` scaffolds a spec. OpenSpec-the-tool itself → ➖ (a tool choice; our SDD chain is the substitute and is ADR-backed, ADR-0014). |
| 18 | Verification-first pyramid (typecheck→unit→integration→E2E→**real-behavior proof / screen recording**) | The 5-layer pyramid incl. a real-behavior video | ✅ Pyramid implemented in `ci.yml` + `docs/qa/test-plan.md`. Layer 5 (video) → artifact exists as `docs/qa/demo-script.md`; the **recording itself is a user action** (deck marks real-behavior proof as the top gate; CHECKLIST tracks it as ☐ user-records). Counts as covered (script present, recording is out-of-repo by nature). |
| 19 | Real numbers from a case (commits, days, caps, FR, coverage) | Quantified outcome reporting | ✅ `docs/agentic-engineering.md` + `docs/SUBMISSION.md` report our equivalents (ADR count, test count, releases, FR coverage). |
| 20 | Greenfield: what gave speed (spec-before-code, dep-ordered slices, smoke on real DB, `current-state.md` handoff, static+dynamic gates, **separate QA pack: traceability matrix / test plan / demo script**) | The full "engineering system" checklist | ✅ Every item present: spec-before-code (`docs/features/`), smoke on real PG (CI integration job), `current-state.md`, gates (`ci.yml`), and the QA pack: `docs/qa/{requirements-traceability-matrix.md,test-plan.md,demo-script.md}`. |

---

## Act 3 — Brownfield flow (slides 21–32)

This is our real regime (we're augmenting a live app). Held to a higher bar.

| # | Slide | Practice / artifact implied | Verdict |
|---|-------|------------------------------|---------|
| 21 | Divider — Brownfield | — | ➖ Section divider. |
| 22 | Reality of brownfield (low coverage, coupling, implicit contracts) | Framing | ➖ Framing; N/A. |
| 23 | What does NOT work (ad-hoc prompts, whole-repo-in-context, generic advice) | Anti-patterns to avoid | ✅ Avoided via `AGENTS.md` rules + scoped context budget; not an artifact but a documented discipline (ADR-0013). |
| 24 | What WORKS (Memory Bank, context borders, rules hierarchy, skills, MCP, SDD) | The brownfield toolkit | ✅ Covered item-by-item by rows below. |
| 25 | Code archaeology pattern (broad→narrow→deep, **verify against source**) | Progressive investigation + verification of AI claims | ➖ NOT-NEEDED as a standing artifact — it's a one-time reverse-engineering ritual; we *authored* the outputs (architecture.md, ADRs) rather than keep the process. Acceptable, but see slide 28. |
| 26 | Token economy: `.cursorignore` + Repomix (measure token size, set context borders) | A committed context-border file + a token measurement | ☐ **MISSING** — no `.cursorignore` / `.aiexclude` / ignore-for-agents file and no recorded Repomix token measurement. Our context discipline lives in prose (ADR-0013 budget) but there is no machine-enforced context border. See GAPS. |
| 27 | Memory Bank — 7 files (projectbrief, productContext, techContext, systemPatterns, activeContext, progress, decisionLog) | A persistent external-memory set | ✅ Functionally complete, mapped onto our own layout: projectbrief→`docs/project-brief.md`; productContext→`docs/project-brief.md`+`requirements.md` (BC); techContext→`AGENTS.md` (pinned stack/commands); systemPatterns→`docs/architecture.md`; activeContext→`docs/memory/activeContext.md`; progress→`docs/memory/progress.md`; decisionLog→`docs/adr/`. ADR-0013 documents this static/dynamic split deliberately. Not the literal Cline "memory-bank/" folder, but every function is present. |
| 28 | Technical & Product docs layers (architecture, dev-setup, data-model, api-and-actions, integrations; domain-glossary, reverse-eng PRD, user journeys, feature inventory) | A layered doc set generated from the code | ✅ mostly: architecture (`docs/architecture.md` + diagrams), data-model + api (in architecture.md + ADRs 0001/0002/0007), integrations (ADRs 0002/0011), reverse-eng PRD (`docs/project-brief.md` + `requirements.md`), feature inventory (`docs/BACKLOG.md` + `mvp-capability-plan.md`). ☐ minor: no standalone `dev-setup.md` and no `domain-glossary.md` — see GAPS (low priority; dev-setup is partly in README/Makefile). |
| 29 | Rules hierarchy (global → project → module → file) | Layered agent rules | ✅ Present at the levels that apply: project rules (`AGENTS.md`/`CLAUDE.md`), path-scoped review rules (`.coderabbit.yaml` path_instructions), tool rules (`.claude/agents/*`, `.codex/config.toml`). Module/file glob rules are a Cursor-specific mechanism → ➖ (not-needed; our repo is small and single-domain, path_instructions cover the one place it matters: scoring/access invariants). |
| 30 | Impact analysis before changes (blast radius, contract vs internal change) | A pre-change impact ritual | ✅ Enforced structurally: invariant guardrails in `AGENTS.md` ("don't change scoring/access without an ADR"), maker≠checker reviewers (`.claude/agents/scoring-correctness-reviewer.md`) whose job is exactly contract/blast-radius review, and ADR-gating. |
| 31 | SDD instead of TDD for legacy (spec = contract) | Spec-driven change | ✅ Core of our approach — `docs/features/<cap>/spec.md` (Given/When/Then) as the contract, traced to tests. ADR-0014. Deck says SDD *complements* TDD; we do both (specs + Go/vitest tests). |
| 32 | Strangler Fig pattern (replace use-case by use-case) | Incremental migration pattern | ➖ NOT-NEEDED — we are not migrating/replacing a legacy module; we add net-new capabilities (MCP, recap) to a healthy codebase. Deck ties Strangler-Fig to legacy sunset, which doesn't apply. |

---

## Act 4 — Synthesis (slides 33–40)

| # | Slide | Practice / artifact implied | Verdict |
|---|-------|------------------------------|---------|
| 33 | **Components matrix** (Greenfield vs Brownfield: which components Core/Light/Skip) | The central decision artifact | ✅ We implement the brownfield "Core" column: MCP ✅, AGENTS.md ✅, OpenSpec/SDD ✅, verification pyramid ✅, rules ✅, impact analysis ✅. The two brownfield-Core items we *don't* have machine-enforced are `.cursorignore`+Repomix and Memory-Bank-as-folder (both discussed above). No matrix artifact of our own is required. |
| 34 | Token spend profile | Cost framing | ➖ Framing; no artifact. Related to the missing token measurement (slide 26). |
| 35 | Velocity curve | Framing | ➖ Framing; no artifact. |
| 36 | Multi-agent orchestration ladder (single → conductor → heavy) | Choose orchestration level | ✅ We sit at "conductor": maker + separate reviewer sub-agents (`.claude/agents/`). Deck explicitly says most teams should stop at conductor — heavy orchestration N/A. |
| 37 | BMAD (full role cycle) | Heavy framework — optional | ➖ NOT-NEEDED — deck marks BMAD for "large teams / exceptional cases"; solo/small project. Explicitly optional. |
| 38 | OpenAI Symphony (Linear→Codex→PR factory) | Optional framework | ➖ NOT-NEEDED — deck-flagged niche (Linear-standardized teams). Optional. |
| 39 | Minions (local-model decomposition) | Optional, cost/privacy niche | ➖ NOT-NEEDED — deck calls it niche/experimental. Optional. |
| 40 | Decision rubric (which orchestration level) | Decision guide | ✅ Our choice (single-agent loop + conductor + SDD) is a valid rubric outcome; documented in `docs/agentic-engineering.md`. |

---

## Act 5 — Teams & process (slides 41–46)

| # | Slide | Practice / artifact implied | Verdict |
|---|-------|------------------------------|---------|
| 41 | How the engineer's role changes (architecture/spec = main verification) | Framing | ➖ Framing; embodied by our SDD/spec-first flow. |
| 42 | What the team must learn (write specs, rules/skills, token budget, calibrate trust, impact analysis, structured verification) | Team-skills framing | ✅ mostly demonstrated (specs, rules, verification, impact analysis). Token-budget skill ties back to the missing measurement (slide 26). |
| 43 | Hiring signals (systematic AI use, own rules/skills) | Framing | ➖ Framing; N/A. |
| 44 | Changes in code review (AI first pass — **CodeRabbit / BugBot**; human on architecture/contracts; **spec → traceability matrix**) | AI reviewer + traceability | ✅ Both: `.coderabbit.yaml` (AI first-pass, uk mentor tone) **and** our own reviewer sub-agents; traceability via `scripts/gen-traceability.mjs` → `docs/qa/requirements-traceability-matrix.md` (req↔code↔test↔demo). |
| 45 | ROI: how to calculate | Framing | ➖ Framing; no artifact. |
| 46 | Anti-patterns (AI-in-legacy without Memory Bank; **OpenSpec without verification loop**; rules-on-everything; token economy without measurement; one-tool-for-all; vibe-coding on prod) | Pitfalls to avoid | ✅ Avoided: we pair SDD *with* a verification loop (not spec-without-verify), keep rules lean. ⚠️ Two we brush against: "token economy without measurement" (we have no measurement at all — slide 26 gap) and the practice is honored elsewhere. |

---

## Act 6 — Closing (slides 47–50)

| # | Slide | Practice / artifact implied | Verdict |
|---|-------|------------------------------|---------|
| 47 | Final decision rubric (task X in project Y) | Decision compass | ✅ Our project = brownfield → we run the brownfield column (Memory-Bank-equivalent + rules + SDD + verification). Documented. |
| 48 | Takeaways | Summary | ➖ Summary; no artifact. |
| 49 | Resources & next steps (links to DOU / fwdays workshops, OpenSpec, BMAD, Symphony, Minions) | Reference links | ➖ Reference; N/A. |
| 50 | Thank you / QR | — | ➖ Closing slide. |

---

## GAPS — only the real, actionable ☐ items, prioritized

Everything else is either implemented or is a deck-flagged-optional / greenfield-only /
tool-choice item that our equivalent already covers. Only four gaps are real, and
all four are small:

1. **Context-border file + token measurement (slide 26; reinforced by 34, 42, 46).**
   *Highest value, lowest effort.* We have a prose "context budget" (ADR-0013) but
   **no machine-enforced context border** and **no recorded token measurement**.
   Add a repo-root ignore-for-agents file (`.cursorignore` / `.aiexclude`,
   excluding `**/dist`, `**/node_modules`, `frontend/dist`, `mcp/dist`, generated
   SVGs, `*.toon`, lockfiles) and record a one-line Repomix token count of the
   repo in `docs/agentic-engineering.md`. Directly closes a brownfield-**Core**
   cell on the slide-33 matrix and neutralizes the slide-46 "token economy without
   measurement" anti-pattern. ~30 min.

2. **Parallel-worktree practice (slide 11, Phase 8).** No documented use of
   `git worktree` for parallel slices. For a solo submission this is genuinely
   optional (deck presents it as an advanced greenfield phase), but if we want to
   claim full phase coverage, add a short `WORKFLOW.md` note on when/how we'd
   fan out slices into worktrees (the `Agent … isolation: worktree` capability
   already exists in our harness). Low priority.

3. **`dev-setup.md` doc layer (slide 28).** The "technical docs layers" set expects
   a standalone dev-setup doc; ours is scattered across `README.md` + `Makefile` +
   `AGENTS.md` commands. Consolidating a short `docs/dev-setup.md` (or explicitly
   pointing the doc-layer list at README) would fully satisfy the layer. Low
   priority — the information exists, it's just not one file.

4. **`domain-glossary.md` (slide 28, product docs layer).** No glossary of domain
   terms (fixture, stage, advancer, regulation draw, tier, reveal-lock). A one-page
   glossary would help agents disambiguate and complete the product-docs layer.
   Lowest priority — terms are defined inline across specs/ADRs.

**Not gaps (explicitly cleared):** OpenSpec-the-tool, BMAD, Symphony, Minions,
module/file-glob Cursor rules, Strangler-Fig, Phase-0 bootstrap, code-archaeology
as a standing process, and heavy orchestration — each is either deck-flagged
optional, greenfield-only (N/A for a live app), or a tool choice our own
ADR-backed equivalent already satisfies. The real-behavior **demo video** is not a
code gap — the script (`docs/qa/demo-script.md`) exists; recording is an
out-of-repo user action already tracked in `CHECKLIST.md`.

---

## Resolution (post-audit, 2026-07-01)
All actionable gaps from the audit are closed:

1. ✅ **Context border + token measurement** — added [`.aiexclude`](../../.aiexclude)
   (machine-enforced border) and recorded measured token counts (AGENTS.md ≈971
   tokens; full pack ≈431k/239 files) in `docs/agentic-engineering.md` §1.
2. ✅ **Parallel-worktree note** — added to `WORKFLOW.md` (loop-engineering notes).
3. ✅ **`docs/dev-setup.md`** — consolidated local setup/build/verify/release.
4. ✅ **`docs/domain-glossary.md`** — shared vocabulary added.

Remaining non-items: greenfield-only bootstrap slides (N/A — WC-Tournament was
already in prod), deck-flagged-optional tooling (BMAD/Symphony/OpenSpec-tool), and
the demo video (out-of-repo owner action, tracked in `CHECKLIST.md`).
