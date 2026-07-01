# ADR-0014: Requirement-id grammar and the trace chain

**Status:** Accepted · **Date:** 2026-07-01

## Context
"Done" must mean *verifiably* done — every requirement provable by a test, and
every test attributable to a requirement. Free-form requirements and hand-written
"we tested it" matrices rot: ids get reused, matrices drift from reality, and
nobody notices coverage gaps.

## Decision
Adopt a stable id grammar and a generated trace chain.

- **Grammar:** `FR` (functional), `NFR` (non-functional), `TC` (technical
  constraint), `BC` (business constraint). Ids are zero-padded, **stable, never
  renumbered**. Each requirement states **one** testable behavior and is tagged
  **MVP** or **Future**. Lives in `docs/requirements.md`.
- **Ownership:** every `FR` is owned by **exactly one** capability slice in
  `docs/mvp-capability-plan.md` and cited in that capability's `spec.md`
  (Given/When/Then).
- **Trace chain:** `FR-id` → spec → owning slice → `@trace <FR-id>` annotation in
  the proving test/eval (`// @trace: FR-013` in Go, `// @trace FR-013` in TS) →
  a row in the **generated** `docs/qa/requirements-traceability-matrix.md` →
  `Slice:` / `Refs:` commit trailers.
- **Generated, checked:** the matrix is produced by `scripts/gen-traceability.mjs`
  by scanning requirements + `@trace` annotations; CI runs it with `--check` and
  fails if the committed matrix is stale. The matrix is never hand-edited.

## Consequences
- Coverage gaps are visible (an FR with no `@trace` shows as untraced).
- The matrix can't silently drift — CI regenerates and diffs it.
- Tiny per-test cost (one `@trace` comment) buys end-to-end traceability.

## Alternatives considered
- **Hand-written traceability matrix** — drifts immediately. Rejected.
- **No ids, prose requirements** — untestable/unattributable. Rejected.
