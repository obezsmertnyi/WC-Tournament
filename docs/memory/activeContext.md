# Active context (dynamic)

Per-session working memory. Changes often; not part of the always-loaded static
budget (that's `CLAUDE.md`). Read this + `progress.md` + `current-state.md` to
resume fast.

## Current focus
Take WC-Tournament through the **full agentic engineering cycle** as the fwdays
"Greenfield" homework submission — augment the (already production-grade) app
with the engineering artifacts the rubric and bonus reward: context rules,
SDD specs, evals, traceability, a separate maker≠checker review, a documented
loop, and a read-only MCP server. Drive it from `CHECKLIST.md` in a loop.

## Constraints in effect
- **No `gh` access** — all repo writes go over git/SSH; GitHub-only steps (fork,
  enable CodeRabbit, record the 1–2 min video, open the PR) are handed to the user.
- Docs in English; PR body in Ukrainian (CodeRabbit reviews `uk-UA`).
- App is **live in prod** — keep `main` green; dep/code changes deploy on next release.

## Pointers
- Loop driver: `CHECKLIST.md` · loop state: `.workflow-state.toon`
- Conventions studied: `docs/agentic-hw-reference.md`
- Submission narrative: `docs/agentic-engineering.md` · PR body: `docs/SUBMISSION.md`
