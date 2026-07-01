# Risk register

Living list of the project's top risks and their mitigations. L = likelihood,
I = impact (1 low – 3 high). Reviewed when a capability or dependency changes.

| ID | Risk | L | I | Mitigation | Status |
|----|------|---|---|------------|--------|
| R-01 | Secret exposure (JWT/OAuth/Telegram/admin) | 2 | 3 | `.env` gitignored + `.aiexclude`; CI gitleaks (full history) + Trivy; fail-closed `JWT_SECRET`. **Open action:** rotate the `.env` values exposed to review tooling (review-findings S1). | mitigated · rotate pending |
| R-02 | FIFA API shape drift breaks sync/scoring | 2 | 3 | Parse tests against captured fixtures (`internal/results`); manual admin override as outage fallback (ADR-0006); status/score never trusted blindly (memory: FIFA status unreliable). | mitigated |
| R-03 | Scoring bug (wrong points) | 1 | 3 | Pure deterministic `Score()`; unit tests + golden-fixture evals (ADR-0008); scoring-correctness reviewer sub-agent; invariants need an ADR to change. | mitigated |
| R-04 | Private pool leaks to strangers (public Google sign-in) | 2 | 2 | Demo-mode access tiers (none/ro/rw), `DemoGate` server-side; new self-service users default `none` (ADR-0012); security-reviewer verified. | mitigated |
| R-05 | AI recap hallucinates a score/team | 1 | 2 | Grounded generation + `validateRecap` guardrail (ADR-0016); recap evals; live LLM only behind the guardrail (FR-082, no client key). | mitigated |
| R-06 | prod amd64 vs local arm64 image mismatch | 1 | 2 | Multi-arch GHCR images (amd64+arm64) from CI; prod pulls a tagged release. | mitigated |
| R-07 | Silent quality erosion (tests/coverage/evals dropped) | 2 | 2 | Non-regressing ratchets: traceability, eval-surface, backend coverage — CI fails on regression. | mitigated |
| R-08 | Comprehension/intent debt (agent-generated code) | 2 | 2 | SDD specs + 16 ADRs record the "why"; traceability FR→spec→test; AGENTS.md as intent ledger. | mitigated |
