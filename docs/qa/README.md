# QA pack — index

Every artifact that turns "seems to work" into checkable evidence. See
[`test-plan.md`](test-plan.md) for the verification pyramid and
[`quality-gates.md`](quality-gates.md) for the G0–G8 command gates.

| Artifact | Purpose |
|----------|---------|
| [`quality-gates.md`](quality-gates.md) | G0–G8 deterministic command gates mapped to our pipeline |
| [`test-plan.md`](test-plan.md) | The 5-layer verification pyramid + how to run each layer |
| [`manual-test-plan.md`](manual-test-plan.md) | Numbered manual test cases (non-developer executable, incl. negatives) |
| [`requirements-traceability-matrix.md`](requirements-traceability-matrix.md) | **Generated** FR → proving test map (CI-checked, non-regressing) |
| [`demo-script.md`](demo-script.md) | Beat sheet for the 1–2 min real-behavior demo video |
| [`risk-register.md`](risk-register.md) | Top risks, likelihood/impact, mitigations, status |
| [`mvp-acceptance-report.md`](mvp-acceptance-report.md) | Per-capability acceptance sign-off table |
| [`review-findings.json`](review-findings.json) | maker≠checker adversarial review results (clean flag) |
| [`automated-verification-latest.md`](automated-verification-latest.md) | Latest `make qa` battery run (generated evidence) |
| [`slide-coverage-audit.md`](slide-coverage-audit.md) | Course-deck (91 slides) coverage audit + adoption notes |

Baselines that ratchet (in [`../../quality/`](../../quality/)): traceability
(`trace-baseline.json`), eval-surface (`eval-baseline.json`), backend coverage
(`coverage-baseline.json`) — quality may only tighten.
