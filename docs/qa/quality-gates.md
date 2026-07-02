# Quality gates G0–G8

Gates are **deterministic command-exit criteria** — judgment (reviews, sign-offs)
sits *atop* green commands, never instead. A red command is a STOP: fixing the
check is allowed, **weakening or bypassing it is not** (AGENTS.md).

A G0–G8 gate model for our stack (Go/React/Postgres, `docs/features/*` specs, our
`scripts/check-*`). Most gates map onto `.github/workflows/ci.yml` + `make ci` +
`make qa`/`make gates`.

| Gate | What it proves | Commands (must exit 0) |
|------|----------------|------------------------|
| **G0 Scaffold & env** | Builds; hooks + CI + check-scripts wired; `AGENTS.md`/`CLAUDE.md`/ADRs present; `.env.example` complete | `make build` · `gofmt -l backend` (empty) · `actionlint .github/workflows/*.yml` |
| **G1 Product framing** | Product brief + numbered `FR/NFR/TC/BC` (MVP/Future) exist | review `docs/product-brief.md` + `docs/requirements.md` (checkpoint: scope sign-off) |
| **G2 Baseline specs** | One Given/When/Then spec per capability; every MVP FR traced once | `node scripts/gen-traceability.mjs --check` |
| **G3 Capability plan** | `mvp-capability-plan.md`: slices own disjoint FR ids | review `docs/mvp-capability-plan.md` (checkpoint: plan sign-off) |
| **G4 Per slice** | test-first `@trace`; lint/test/build green; separate review clean; `Slice:`/`Refs:` trailers | `cd backend && go vet ./... && go test -race ./...` · `cd frontend && npm test` · `node scripts/gen-traceability.mjs --check` · reviewer sub-agents → `docs/qa/review-findings.json` |
| **G5 Cross-cutting hardening** | Integration tests on a real Postgres; coverage ratcheted | `DATABASE_URL=… go test ./...` · `node scripts/check-coverage-ratchet.mjs --update` |
| **G6 QA proof pack** | Traceability matrix + manual test plan + risk register + acceptance report; eval baseline; battery all-green | `make qa` (→ `docs/qa/automated-verification-latest.md`) · `node scripts/check-eval-ratchet.mjs --check` |
| **G7 Global review & release** | Security clean; CI green; multi-arch images + SBOM; live smoke | `node scripts/gen-traceability.mjs --check` · `cd backend && govulncheck ./...` · gitleaks · Trivy · tag `v*` → `release.yml` → live `/healthz` |
| **G8 UAT round** | Each bug → verdict + regression test `@trace BUG-x` | *(informal for a friends pool — no formal UAT; bugs reported ad-hoc, fixed via G4)* |

## Roll-up
`make gates` runs the deterministic subset (G-trace/G-eval/G-cover) and prints
PASS/FAIL/SKIP. `make qa` runs the full battery and writes the evidence report.

## Deviations from the reference (deliberate)
- **Specs, not a spec CLI** — we keep specs as `docs/features/<cap>/spec.md` (Given/When/Then) + ADRs rather than adopting a dedicated spec-management CLI such as **OpenSpec**; validated via our `gen-traceability --check`.
- **G6 headless recordings + a11y/vision** deferred — the narrated demo video is the real-behavior proof (`demo-script.md`); browser-automation infra is disproportionate for a single-audience app.
- **G8** informal — no external QA round.
