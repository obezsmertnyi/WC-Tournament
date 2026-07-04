# ADR-0021: Design-first UI workflow (design artifact before implementation)

**Status:** Accepted · **Date:** 2026-07-05 · **Related:** ADR-0019 (specs approach)

## Context
UI built by asking the coding agent to copy a reference "but a bit different"
comes out generated-looking and needs many rounds: the agent must interpret the
reference, infer the unstated intent, and produce code in one lossy step, with no
reviewable checkpoint before the code exists. This mirrors the ambiguity that
caused real bugs elsewhere in this repo (the AI reporting wrong results, ET-goal
mis-scoring) — the agent performs badly when left to infer, well when handed an
explicit spec.

## Decision
Adopt a **design-first** workflow for UI work: produce a precise, reviewable
**design artifact first**, approve it, then implement the code against that exact
artifact — the visual analog of our spec-driven flow. Two routes:

- **Route A — Artifact (default, no setup):** the agent authors a self-contained
  HTML design (via the `artifact-design` skill), honoring the app's real tokens
  and copy, showing all states in both themes; review; implement 1:1.
- **Route B — Claude Design + DesignSync:** run `/design-login` once, then create
  and sync a design-system project component-by-component (`/design-sync` +
  `DesignSync`); implement against the synced components.

Guardrails: honor the existing design system (precedence `user > project >
agent`), **no emoji** (SVG icons or typographic marks — in app *and* docs), both
themes, one accent, real content, no heavy assets the runtime wouldn't ship,
reduced-motion support, and nothing left for the agent to infer.

The full playbook (greenfield + brownfield, step by step) lives in
[`docs/design-first-workflow.md`](../design-first-workflow.md).

## Consequences
- A reviewable checkpoint (the artifact) before any code — cheap to change.
- More consistent, on-brand UI; fewer implementation rounds.
- Portable reference for future greenfield/brownfield projects (this repo is used
  as a reference).
- First applied to the AI assistant page refresh (Route A).

## Alternatives considered
- **Screenshot + "do it but different" (status quo).** Rejected — the failure
  mode this ADR exists to fix.
- **Design and code in one pass, no artifact.** Rejected — no review checkpoint;
  the coding context both invents and implements the design, losing the
  separation that removes ambiguity.
