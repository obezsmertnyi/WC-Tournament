# Evals — the executable quality bar

Evals are repeatable, offline checks that encode *what good looks like* for the
parts of the system where "the tests pass" isn't enough on its own. They sit
above unit tests in the [verification pyramid](../docs/qa/test-plan.md) and below
the manual real-behavior proof (the demo video).

Each eval case names the requirement it proves (`@trace <FR-id>`, see
[ADR-0014](../docs/adr/0014-requirement-id-grammar.md)) and runs without touching
prod or any secret.

## Suites

| Suite | Where | Kind | Run |
|---|---|---|---|
| **Scoring** | `backend/internal/scoring/scoring_evals_test.go` | deterministic golden fixtures (exact points + breakdown per spec scenario; purity asserted) | `cd backend && go test -tags=evals ./internal/scoring/` |
| **MCP tools** | `mcp/evals/` | golden-fixture tool I/O over a stubbed API (shape, reveal-lock, input-validation rejections) | `cd mcp && npm run eval` |

Why two kinds: scoring is a **pure deterministic** function, so its rubric *is*
the exact expected output — no LLM judge needed. The MCP suite checks tool
contracts and the safety invariants (read-only, validated input, reveal lock)
against golden fixtures. If a free-text / generative surface is added later, its
evals would add an LLM-as-judge rubric for subjective quality; we deliberately
avoid that until there's a non-deterministic output worth judging.

## Convention
- One case = one named spec scenario; golden expected value committed.
- `@trace <FR-id>` on every case → flows into the generated traceability matrix.
- Evals never mutate state and never read secrets.
