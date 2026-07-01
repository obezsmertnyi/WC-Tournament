# Spec — AI match recap (CAP-09, GenUI bonus)

A generated natural-language recap of a finished/kicked-off match, shown in the
UI. The engineering point is **safe generative UI**: the output is *grounded*
and passes a guardrail before display, so the feature can't hallucinate a wrong
score or team — regardless of whether a template or an LLM produced it.

Owns: FR-080, FR-081 (FR-082 Future). Decision: [ADR-0016](../../adr/0016-ai-recap-grounded-guardrail.md).

## Generation

### FR-080 — recap from facts
- **Given** a finished fixture `BRA 2:1 ARG` (R16)
- **When** the recap is generated
- **Then** it is a short natural-language summary mentioning only the real teams,
  the real scoreline, and the stage; if exact-score guessers are supplied it
  congratulates them.

### FR-080 — pre-result state
- **Given** a fixture with no recorded score yet
- **When** a recap is requested
- **Then** the component shows nothing (or a "not played yet" note) — no recap is fabricated.

## Guardrail (the safety contract)

### FR-081 — reject a hallucinated score
- **Given** the match facts are `BRA 2:1 ARG`
- **When** a candidate recap claims "BRA won 3:0"
- **Then** the guardrail flags it as ungrounded (a number not in the facts) and the recap is **not displayed** as-is.

### FR-081 — reject a hallucinated team
- **Given** the match is `BRA` vs `ARG`
- **When** a candidate recap mentions "Germany"
- **Then** the guardrail flags it ungrounded and rejects it.

### FR-081 — accept a grounded recap
- **Given** the match facts `BRA 2:1 ARG`
- **When** a candidate recap mentions only BRA, ARG, and 2/1
- **Then** the guardrail passes it.

### FR-081 — bound length & neutralize injection
- **Given** a candidate recap that is excessively long or contains control/markup/prompt-injection sequences
- **Then** the guardrail truncates to the bound and strips the unsafe content.

## Provider (Future)

### FR-082 — pluggable LLM behind the guardrail
- **Given** an optional LLM recap provider
- **When** it returns prose
- **Then** that prose is run through the FR-081 guardrail before display; raw LLM
  output is never shown. The default provider is a grounded template (no API key,
  works offline).
