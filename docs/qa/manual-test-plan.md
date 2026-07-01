# Manual test plan

Executable by a non-developer against the live app (`wc2026.mtgrd-das.app`) or a
local `docker compose up`. Each case cites the requirement it exercises; tick
Pass/Fail. Includes negative cases (locks, access denial, no-hallucination).

| # | Case | Steps | Expected | FR | P/F |
|---|------|-------|----------|----|-----|
| MT-01 | Submit a prediction | Sign in → open an upcoming fixture → enter e.g. `2:1` | Saved; editable until kickoff | FR-001 | ☐ |
| MT-02 | Knockout draw needs advancer | On a knockout tie predict `1:1` | An "who advances?" picker appears; can't save until picked | FR-003 | ☐ |
| MT-03 | Kickoff lock (negative) | Open a match that has kicked off | Cannot create/change a prediction (locked) | FR-002 | ☐ |
| MT-04 | Leaderboard & reveal | Competition tab after a match kicks off | Leaderboard (match+bonus split); revealed picks with points | FR-011 | ☐ |
| MT-05 | Bracket order | Open the bracket | Ties in tree order (R32→final); connectors align | — | ☐ |
| MT-06 | Demo-mode preview (negative) | As a fresh/`none` user with demo mode on | Preview banner; leaderboard/reveals show a locked card; cannot submit | FR-031 | ☐ |
| MT-07 | Access granted | Admin sets a user to `rw` → user reloads | User can now see others and submit predictions | FR-031 | ☐ |
| MT-08 | AI recap grounding | Open a finished match's reveal | Recap names only the two real teams + real score; never a wrong team/score | FR-080, FR-081 | ☐ |
| MT-09 | Telegram result post | After a match finishes | Bot posts result + congratulates exact-score guessers (in the group) | — | ☐ |
| MT-10 | Admin override | Admin sets a result manually | Scores recompute; action appears in the audit feed | FR-015 | ☐ |
| MT-11 | Bilingual UI | Toggle UA/EN | All visible strings switch language | — | ☐ |
| MT-12 | Bad login (negative) | Try dev-login with an unknown nickname | "No such player" — no session issued | FR-030 | ☐ |

**Tester:** ______  **Date:** ______  **Env:** prod / local
