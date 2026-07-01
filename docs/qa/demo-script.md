# Demo script — 1–2 min video (real-behavior proof)

The top of the verification pyramid: a short screen recording that proves the
product actually works end-to-end (layer 5, `test-plan.md`). Also the homework's
required video demo. Record at `wc2026.mtgrd-das.app` (or a local `docker compose
up`), save to `docs/qa/demo-recordings/`, and link it in the PR.

Keep it to ~90 seconds. Suggested beat sheet:

1. **(0:00–0:15) What it is.** One line: "private World Cup 2026 prediction pool
   for friends — predict scores, earn points, climb a live leaderboard."
2. **(0:15–0:40) Predict + lock.** Open a fixture, enter a score; on a knockout
   tie predict a draw and show the advancer picker appears (FR-003). Show that a
   kicked-off match is locked (FR-002).
3. **(0:40–1:00) Scoring + leaderboard.** Open the Competition tab: leaderboard
   with the match/bonus split, a revealed match showing who predicted what, the
   bracket in correct tree order.
4. **(1:00–1:20) Demo-mode gate.** As a fresh/`none` user, show the preview
   banner and a locked panel; as `rw`, show full participation (FR-031).
5. **(1:20–1:40) Agentic angle (the point of the homework).** Show the MCP server
   answering a question about the pool from an agent (e.g. "who's leading?"),
   and flash the artifacts: `CHECKLIST.md`, `docs/qa/requirements-traceability-matrix.md`,
   the CI run green, a reviewer sub-agent's `review-findings.json`.

Narration focus: not just *what* the product does, but *how it was built
agentically* — specs → tests/evals → loop → separate review.
