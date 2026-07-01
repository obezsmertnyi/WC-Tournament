# ADR-0016: AI match recap — grounded generation behind a guardrail

**Status:** Accepted · **Date:** 2026-07-01

## Context
We want a small generative-UI touch: a natural-language recap of a finished
match. The risk with any generative feature in a live product is hallucination
(a recap claiming the wrong score/team) and prompt injection — and an LLM call
adds a provider key (a secret), cost, and latency on the request path. We want
the *feature* without importing those risks unguarded.

## Decision
Treat the recap as **grounded generation behind a guardrail**, provider-agnostic.

- A `RecapProvider` produces candidate prose. The **default provider is a
  grounded template** (`buildRecap(match)`) — deterministic, varied, no API key,
  works offline.
- Whatever the provider, output passes through **`validateRecap(text, match)`**
  before display: it rejects text that introduces a **team or number not present
  in the match facts** (anti-hallucination), bounds length, and strips
  control/markup/injection sequences. The UI shows only validated text; on a
  guardrail failure it falls back to the grounded template.
- A future LLM provider (FR-082) plugs in *behind* the same guardrail — its raw
  output is never shown. The LLM key, if ever added, lives in env/vault, never in
  the client or the prompt context.
- The guardrail is the eval'd unit (`recap.test.ts`, FR-080/FR-081): grounded
  recaps pass, hallucinated score/team are rejected.

## Consequences
- A real GenUI feature ships with **zero new secret** and zero live-LLM risk by
  default; the safety property (grounding) is tested, not assumed.
- Adding an LLM later is a contained change (a new provider) — the guardrail and
  its evals already constrain the blast radius.

## Alternatives considered
- **Call an LLM and render its output directly** — hallucination + injection +
  a provider key on the live path. Rejected.
- **No recap** — misses the GenUI demonstration the assignment rewards. Rejected.
- **Backend LLM endpoint** — more infra/secret for the same user-visible result;
  deferred until there's a real LLM provider to host (FR-082).
