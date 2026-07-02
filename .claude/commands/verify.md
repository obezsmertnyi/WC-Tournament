---
description: Run the full verification pyramid (the CI-equivalent gate) and report pass/fail per layer.
---

Run every verification gate for WC-Tournament and report a concise pass/fail per
layer (see `docs/qa/test-plan.md`). Stop and surface the first failure with its
output; do not "fix by disabling".

```bash
# 1 static
cd backend && gofmt -l . && go vet ./...
cd ../frontend && npm run build          # tsc + vite
cd .. && actionlint .github/workflows/*.yml
# 2 unit + 3 integration (needs DATABASE_URL for the integration tests)
cd backend && go test -race ./... && cd ../frontend && npm test
# 4 evals
cd ../backend && go test -tags=evals ./internal/scoring/
cd ../mcp && npm run eval
# traceability (fresh + non-regressing) + doc-graph integrity
cd .. && node scripts/gen-traceability.mjs --check
node scripts/check-doc-links.mjs --check   # no bootstrap doc points at a missing file
```

Report which layers passed. If a gate is red, propose a root-cause fix (never a
workaround) and, once fixed, re-run to prove green.
