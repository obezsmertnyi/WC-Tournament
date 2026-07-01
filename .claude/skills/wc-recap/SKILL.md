---
name: wc-recap
description: Generate a grounded, guardrailed natural-language recap of a finished WC-Tournament match from its facts. Use when an agent needs a match summary; safe by construction (never invents a score or team).
---

# wc-recap

Produce a short recap of a match from its facts (teams, scoreline, stage), with a
congrats line for exact-score guessers. This is the **portable** form of the
CAP-09 recap capability (see `docs/features/recap/spec.md`, ADR-0016): one core,
many surfaces — the same grounding + guardrail power the in-app `MatchRecap.tsx`
and this skill.

## Contract
- **Input:** a match JSON, either as `argv[2]` or on stdin, e.g.
  `{"home":"Brazil","away":"Argentina","homeScore":2,"awayScore":1,"stage":"r16","exactGuessers":["alice"]}`
- **Output (stdout):** one grounded recap sentence. Empty if there is no result yet.
- **Guarantee:** the text mentions only the given teams and only the scoreline
  numbers — `run.mjs` validates its own output and falls back to the grounded
  template on any violation. It never calls a model and never touches secrets.

## Use
```bash
node .claude/skills/wc-recap/run.mjs '{"home":"Brazil","away":"Argentina","homeScore":2,"awayScore":1,"stage":"r16"}'
# → Brazil beat Argentina 2:1 in the last sixteen.
echo '{"home":"Spain","away":"Italy","homeScore":1,"awayScore":1,"stage":"final","exactGuessers":["bob"]}' | node .claude/skills/wc-recap/run.mjs
```

## Notes
- Stages map to digit-free phrases so the scoreline is the only number a recap
  can contain (the anti-hallucination number check).
- To wire a live LLM later, generate a candidate and pass it through the same
  guardrail before printing (FR-082) — raw model output is never shown.
