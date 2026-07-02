# Active context (dynamic)

Per-session working memory. Changes often; not part of the always-loaded static
budget (that's `CLAUDE.md`). Read this + `progress.md` to resume fast.

## Current focus
WC-Tournament is a production-grade, friends-only FIFA World Cup 2026 prediction
pool, **live at `wc2026.mtgrd-das.app`**. Latest work: the football AI assistant
"Pitchside" (grounded chat + player/club cards, keyless WIF) and ongoing polish.

## Constraints in effect
- **No `gh` access** — repo writes go over git/SSH; GitHub-UI steps are handed to the user.
- Docs in English.
- App is **live in prod** — keep `main` green; dep/code changes deploy on next release.

## Pointers
- Map of the whole repo: `docs/REPO-MAP.md` · the loop: `WORKFLOW.md` / `LOOP.md`
- Decisions: `docs/adr/` · specs: `docs/features/*/spec.md` · requirements: `docs/requirements.md`
