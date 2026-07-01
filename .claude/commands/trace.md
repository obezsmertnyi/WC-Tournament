---
description: Regenerate the requirements-traceability matrix and verify freshness + non-regressing coverage.
---

Regenerate the traceability artifacts from `docs/requirements.md` + the `@trace
<FR-id>` annotations, then verify (ADR-0014):

```bash
node scripts/gen-traceability.mjs          # write matrix + trace.json (+ ratchet baseline up)
node scripts/gen-traceability.mjs --check  # must be fresh and not below baseline
```

Then report the coverage line and list any **untraced MVP** requirements. For
each untraced FR, either (a) add a test/eval that proves it and annotate it with
`// @trace <FR-id>`, or (b) confirm it's intentionally deferred (Future) and say
why. Commit the regenerated `docs/qa/requirements-traceability-matrix.md`,
`trace/trace.json`, and `quality/trace-baseline.json` together with the code.
