# WORKFLOW — how we build WC-Tournament agentically

The engineering loop this repo runs. It replaces step-by-step prompting with a
repeating, verifiable cycle; a **checker that is not the maker** gates each
change; and every requirement is traced to a test. See also `LOOP.md` (the tight
cycle), `CHECKLIST.md` (the live driver), and `AGENTS.md` (rules/context).

## Roles
- **Maker** — the agent implementing a slice (writes code + tests).
- **Checker** — a *separate* pass that never wrote the code: the reviewer
  sub-agents in `.claude/agents/` (`scoring-correctness-reviewer`,
  `security-reviewer`) and **CodeRabbit** on the PR. Findings persist to
  `docs/qa/review-findings.json`.

## The cycle (one capability slice)
1. **Spec** — add/refine `docs/features/<cap>/spec.md` (Given/When/Then) and the
   FR/NFR ids in `docs/requirements.md`; record a decision in `docs/adr/` if the
   choice is architectural.
2. **Implement** — smallest correct change; match surrounding style.
3. **Trace** — annotate the proving test/eval with `@trace <FR-id>`.
4. **Verify** — run the pyramid (`make ci`, evals, `node scripts/gen-traceability.mjs --check`).
5. **Review (maker≠checker)** — dispatch the reviewer sub-agents; address real
   findings; CodeRabbit reviews the PR.
6. **Commit** — conventional message + `Slice:` / `Refs:` trailers; push; CI is
   the gate to `main`.
7. **Handoff** — update `current-state.md` + `docs/memory/progress.md`.

## Verification gates (must be green before merge — see `docs/qa/test-plan.md`)
`gofmt · go vet · tsc · actionlint` → `go test` (unit + integration on real PG) +
`vitest` → `go test -tags=evals` (scoring) + `npm run eval` (MCP) →
`gen-traceability --check` (fresh + non-regressing) → `gitleaks` + `govulncheck` +
Trivy → container smoke. The top gate — real behavior — is the demo video
(`docs/qa/demo-script.md`).

## Loop-engineering notes
- **One PR = one capability = one demo.** Slices are vertical (spec→UI/API→tests→docs).
- **Plan + Verify** over freehand prompting: the CHECKLIST is re-checked in-loop
  after every change; nothing is "done" until its gate is green and its FR is traced.
- Dynamic state (`current-state.md`, `docs/memory/`) carries context between
  sessions so a fresh agent resumes in ~30s; static rules stay in `AGENTS.md`.
