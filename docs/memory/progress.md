# Progress (dynamic)

Authoritative phase log for the agentic-HW augmentation. Source of truth for
"what's done"; mirror of `CHECKLIST.md` status. Newest at top.

## Done
- Product app: M0–M9 shipped, live in prod (calendar, groups, predictions,
  scoring, bracket, bonuses, leaderboard, history, top-scorers, Telegram bot,
  Google OAuth, demo-mode access tiers).
- CI/CD: `ci.yml` (gofmt/vet/race-tests-on-PG/build · tsc+vite · govulncheck ·
  gitleaks · docker build+trivy+smoke) + `release.yml` (multi-arch GHCR + SBOM +
  GitHub Release). 12 ADRs. 22 Go tests. Releases v0.1.0–v0.1.3.
- Agentic context: `CLAUDE.md`, `AGENTS.md`, `docs/memory/`, `current-state.md`.

## In progress
- SDD: `requirements.md`, `mvp-capability-plan.md`, per-capability specs, ADR 0013/0014.

## Next
- Verification: scoring evals + frontend vitest + traceability matrix + CI wiring.
- MCP server (read-only) + spec/tests/evals.
- maker≠checker reviewer agents + adversarial pass.
- Loop docs (`WORKFLOW.md`/`LOOP.md`, `.claude/commands`, hooks).
- Narrative `docs/agentic-engineering.md` + uk `docs/SUBMISSION.md`.
