# MVP acceptance report

Per-capability acceptance against `docs/requirements.md` and the capability
specs. "Verified by" cites the gate that proves it (see the generated
`requirements-traceability-matrix.md` for the FR→test mapping).

| Capability | FRs | Verified by | Accepted |
|------------|-----|-------------|----------|
| CAP-01 Predictions | FR-001..003 | Go handler tests (`m2_predictions_test.go`) · live in prod | ✅ |
| CAP-02 Scoring | FR-010..013, 015 | unit tests + golden-fixture evals (`scoring_evals_test.go`, `-tags=evals`) | ✅ |
| CAP-02 Bonus resolution | FR-014 | `resolve_test.go` | ✅ |
| CAP-05 Auth & demo access | FR-030..033 | `auth/access_test.go`, `jwt_test.go`; security-reviewer pass | ✅ (FR-033 covered by SQL default; not @trace'd) |
| CAP-08 MCP tools | FR-070, 071 | `mcp/evals` (10 cases); typecheck | ✅ |
| CAP-09 AI recap (GenUI) | FR-080, 081 | `recap.test.ts` (grounding + guardrail) | ✅ (live LLM = FR-082, deferred) |

## Verification summary
- 5-layer pyramid green (static · unit · integration-on-real-PG · evals · demo).
- Traceability: 17/19 FR traced (16/17 MVP); matrix generated, CI-checked, non-regressing.
- Ratchets green: traceability · eval-surface (28 cases) · backend coverage (≥34.5%).
- Separate review (maker≠checker) clean after fixing C1/C2/C4/C5; CodeRabbit pending on PR.
- Latest battery: `docs/qa/automated-verification-latest.md` (10/10).

## Open items (non-blocking for acceptance)
- Owner: rotate `.env` secrets (R-01 / review-findings S1).
- Owner: record the 1–2 min demo video (real-behavior proof, `demo-script.md`).
- FR-033 has a SQL-default implementation but no dedicated `@trace` test yet.

**Sign-off:** _pending owner_ — date: _____
