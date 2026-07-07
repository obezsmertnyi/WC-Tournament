# Runbook: repeating the pool for a new edition (2026 → 2030 → …)

How to close out one World Cup edition and open the next, using the multi-edition
model from [ADR-0022](adr/0022-tournament-editions.md). Editions coexist in one
app: past editions are archived **read-only and stay browsable** (edition
switcher); exactly one edition is **active** and accepts play; `users` (the
friends) are shared across editions.

All management is done from the **admin panel** (admin-only, audited) — never by
hand in the DB.

---

## A. After the current tournament ends (archive the finished edition)

Do this once the final has been played. Nothing is deleted — archiving is just
making the edition read-only; its full history stays viewable.

1. **Confirm the final result is correct.** Check the final match score (and, for
   a knockout decided in extra time / penalties, that the stored scoreline is the
   regulation result with `winner_team_id` = the actual winner — ADR-0020).
2. **Resolve the tournament bonuses.** Champion / finalist / top-scorer are
   awarded only after the outcome exists. This runs automatically once the final
   is finished; to force it, run `server resolve-bonuses`. Verify each bonus’
   `awarded` flag flipped for the correct picks.
3. **Verify the final record that will be archived:** leaderboard totals, the
   knockout bracket, group standings, and top scorers all look correct. This is
   the frozen history players will browse later, so it must be right before you
   move on.
4. **(Optional) Post a season wrap-up** to the Telegram group — champion, pool
   winner, notable calls.
5. **Leave the edition as-is.** You do NOT archive it manually — activating the
   next edition (step B) automatically flips this one to read-only. Until then it
   remains the active edition.

> There is no data wipe and no export step: the edition’s rows stay in place,
> tagged with its `tournament_id`, and become read-only when the next edition is
> activated.

---

## B. Before the new tournament starts (open the next edition)

Do this when the next World Cup’s teams and schedule are published (the draw is
usually ~6 months before kick-off). All steps are in the admin panel →
**Tournaments** section.

1. **Create the edition.** New tournament with `code` (e.g. `WC2030`), `name`
   (`FIFA World Cup 2030`) and `year`. It is created **inactive** — invisible to
   players, so you can prepare it safely while the old edition is still live.
2. **Load fixtures.** Trigger the edition-scoped FIFA sync to populate that
   edition’s teams and fixtures (group stage first; the knockout bracket fills in
   as the draw resolves). Re-run it to pick up schedule/venue updates.
3. **Review the seeded data:** groups, teams, fixtures, kick-off times (timezone),
   and the bracket shape look right.
4. **Review the bonus rules/tiers** for the edition (champion/finalist/top-scorer
   point values and the tier windows) — adjust if the group agreed on changes.
   Remember the bonus **hard-lock is the first Round-of-16 kick-off**
   (ADR-0008), and the bonus-deadline reminder (`internal/remind/bonus.go`) will
   fire ~24h before it.
5. **Activate the new edition.** This atomically makes it the active edition and
   **archives the previous one** (which becomes read-only). Confirm the prompt —
   it changes who can play. From this point:
   - Players’ existing accounts carry over; their new predictions/bonuses go to
     the new edition (the old edition’s picks are untouched and read-only).
   - Predictions and bonus picks reopen for the new edition.
6. **Announce** to the Telegram group that the new season is open, with the link.

---

## What carries over vs. what is per-edition

| Carries over (shared) | Reset per edition (archived + fresh) |
| --- | --- |
| `users` (accounts, nicknames, Telegram links) | teams, matches/results/bracket |
| App configuration / AI assistant | predictions, points, leaderboard |
| | tournament bonuses (champion/finalist/top-scorer) + tiers |
| | group standings, top scorers, reminder state |

---

## Implementation status

- **Phase 1 (done):** schema — `tournaments` entity + `tournament_id` scoping,
  current data backfilled as the `WC2026` edition (active). No behavior change.
- **Phase 2 (planned):** edition-aware queries + a write-guard that blocks play on
  an archived edition; edition-qualified unique constraints.
- **Phase 3 (planned):** the admin-panel **Tournaments** section + endpoints used
  by section B (create / load-fixtures / activate-archive).
- **Phase 4 (planned):** the read-only edition switcher used to browse past
  editions.

Until Phases 3–4 land, sections A/B describe the intended flow; the create/load/
activate actions become available with Phase 3.
