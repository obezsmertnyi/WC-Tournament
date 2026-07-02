# Progress (dynamic)

Phase log. Source of truth for "what's done". Newest at top.

## Done
- Product app: M0–M9 shipped, live in prod (calendar, groups, predictions,
  scoring, bracket, bonuses, leaderboard, history, top-scorers, Telegram bot,
  Google OAuth, demo-mode access tiers).
- CI/CD: `ci.yml` (gofmt/vet/race-tests-on-PG/build · tsc+vite · govulncheck ·
  gitleaks · docker build+trivy+smoke) + `release.yml` (multi-arch GHCR + SBOM +
  GitHub Release). 18 ADRs.
- Engineering layer: `AGENTS.md`/`CLAUDE.md`, SDD (`requirements.md`,
  `mvp-capability-plan.md`, per-capability specs), verification (scoring evals,
  vitest, traceability matrix, quality ratchets), maker≠checker reviewer agents,
  loop docs (`WORKFLOW.md`/`LOOP.md`, `.claude/`), read-only MCP server.
- AI assistant "Pitchside" (ADR-0017/0018): grounded chat (DB + web) + player/club
  cards (photo, stats, ask-when-ambiguous), keyless WIF. Recap rework + AI-page
  motion. Security: grpc CVE bump; pre-commit gofmt/vet.

## Next
- Iterate on AI quality + any prod issues surfaced by use.
