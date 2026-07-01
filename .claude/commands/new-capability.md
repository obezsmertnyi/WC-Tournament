---
description: Scaffold a new capability slice — spec, requirement ids, and a plan row — before writing code (SDD).
argument-hint: <capability-name>
---

Start a new capability slice **spec-first** (SDD), for `$ARGUMENTS`:

1. Allocate the next free `FR`/`NFR` ids in `docs/requirements.md` (never reuse
   or renumber; tag each MVP/Future; one testable behavior each — ADR-0014).
2. Add a row to `docs/mvp-capability-plan.md`: capability name, slice summary,
   the FR ids it owns (disjoint from other capabilities), and the spec path.
3. Create `docs/features/$ARGUMENTS/spec.md` with Given/When/Then scenarios, each
   citing its FR id. Add an ADR under `docs/adr/` if the approach is architectural.
4. Only then implement, annotating the proving test with `@trace <FR-id>`.
5. Run `/verify` and `/trace`, then `/review` (maker≠checker) before committing
   with `Slice: $ARGUMENTS` / `Refs: FR-…` trailers.

Do not write implementation code before the spec + FR ids exist.
