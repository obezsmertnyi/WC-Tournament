# Domain glossary

Shared vocabulary for WC-Tournament — so specs, code, reviews, and the agent all
mean the same thing.

| Term | Meaning |
|------|---------|
| **Fixture** | A single scheduled match (two teams, a kickoff time, a stage). |
| **Prediction** | A player's guess of a fixture's regular-time scoreline (home, away), plus an advancer pick on a knockout draw. |
| **Regulation / regular time** | 90 minutes (+ stoppage), before any extra time or penalties. Scoring uses the regulation scoreline. |
| **Advancer** | The team that actually progresses from a knockout tie, resolved via extra time / penalties upstream — never part of the predicted scoreline. |
| **Regulation draw** | A knockout tie level at the end of regulation; only then is the advancer a separate question (and the +1 applies). |
| **Exact / Outcome** | Scoring buckets: exact scoreline = 3 pts; correct result (W/D/L) but wrong score = 1 pt; wrong = 0. |
| **Bonus (tiered)** | Tournament-wide champion / finalist / top-scorer picks; worth more the earlier they're set; awarded only if correct. |
| **Reveal / reveal-lock** | Other players' predictions for a match are hidden until it kicks off ("reveal-lock"), then revealed. |
| **Demo mode** | A global toggle that, when on, gates non-admin users by an access **tier**. |
| **Access tier** | Per-user level under demo mode: `none` (browse UI only), `ro` (also see others' data), `rw` (also participate). Admin / demo-off ⇒ `rw`. |
| **Standings / third-place ranking** | Derived group tables; the ranking of third-placed teams (best 8 advance to the round of 32). |
| **Bracket tree order** | Knockout ties ordered by their position in the tree (feeder→parent), not by FIFA match number, so connectors align. |
| **Capability (CAP-NN)** | A vertical feature slice (UI→API→domain→DB→tests→docs) that owns a disjoint set of FR ids. |
| **FR / NFR / TC / BC** | Requirement ids: Functional / Non-functional / Technical constraint / Business constraint (`docs/requirements.md`). |
| **@trace** | An annotation in a test/eval tying it to an `FR-id`; feeds the generated traceability matrix. |
| **Eval** | An offline golden-fixture check encoding the quality bar (scoring, MCP tools) — above unit tests in the pyramid. |
| **Maker ≠ checker** | The author never grades their own work; a separate reviewer (sub-agent + CodeRabbit) gates the change. |
| **MCP** | Model Context Protocol — our read-only server exposes the pool to agents as tools. |
| **Grounded generation / guardrail** | The AI recap is built from match facts and validated (no hallucinated score/team) before display. |
