# Test plan — the verification pyramid

How "done" is proven, bottom-up. AI works the lower layers; the human owns the
top gate. Each layer is automated in CI except the last (the demo video).

| # | Layer | What | Where | Run |
|---|-------|------|-------|-----|
| 1 | **Static** | gofmt · go vet · tsc · actionlint · gitleaks · govulncheck · Trivy | CI `ci.yml` | `make ci` / per-tool |
| 2 | **Unit** | pure logic — Go (22 tests) + frontend (vitest) | `backend/**_test.go`, `frontend/src/**/*.test.ts` | `make test` · `cd frontend && npm test` |
| 3 | **Integration** | API + handlers against a **real Postgres 17** service | `backend/internal/{api,storage}/**_test.go` | `DATABASE_URL=… go test ./...` (CI runs a PG service) |
| 4 | **Evals** | golden-fixture quality bar for scoring (+ MCP tools) with `@trace` | `backend/internal/scoring/scoring_evals_test.go`, `mcp/evals/` | `go test -tags=evals ./internal/scoring/` · `cd mcp && npm run eval` |
| 5 | **Real behavior** | the 1–2 min demo video (the gate mocks never reach) | `docs/qa/demo-recordings/` (link in PR) | manual, see `demo-script.md` |

## Traceability
Every FR is proven by a test/eval carrying `@trace <FR-id>`; the matrix
(`docs/qa/requirements-traceability-matrix.md`) is **generated** and CI fails if
stale or if coverage regresses below `quality/trace-baseline.json`
(`node scripts/gen-traceability.mjs --check`). See
[ADR-0014](../adr/0014-requirement-id-grammar.md).

## Smoke (containers)
CI builds both images and smoke-tests them: backend boots (with a throwaway
`JWT_SECRET`) and serves `/healthz`; the frontend nginx image serves `/`.

## What is intentionally NOT auto-tested
- Visual/animation polish and real-device feel → covered by the demo video.
- Live FIFA-API shape drift → guarded by `internal/results` parse tests against
  captured fixtures, not by hitting the live API in CI.
