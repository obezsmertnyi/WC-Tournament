# Project Brief — WC2026 Prediction Pool

**Status:** Draft v0.1 (scoping) · **Date:** 2026-06-13 · **Owner:** o.bezsmertnyi

## Executive summary
A friends-only score-prediction pool for the 2026 World Cup. We are porting the mechanics of the WC2022 Excel sheet ("guess the score → earn points → leaderboard") into a football-themed **web app** for the same group of 5–15 people. The 2022 sheet is used **only** to learn two things: **who plays** (the participants) and **how points were counted** (the scoring method). The tournament structure itself (teams, groups, knockout rounds) is **not hardcoded** — it is ingested from the data source.

**This is a PoC** — we run it on the friends group for WC2026 to prove the idea, then may grow it into a polished public product later. Build PoC-simple, but keep clean extension points (pluggable data providers, versioned rules, decoupled notifier) so the future product isn't a rewrite.

**Delivery decided:** web app (React frontend + Go/Gin backend + Postgres), local-first via `docker-compose`. Match calendar and results are scraped from sport.ua with an admin manual override; the knockout bracket auto-populates after the group stage or is entered by the admin. Technical design in [architecture.md](architecture.md).

## Problem / motivation
- The old Excel sheet worked, but: points were tallied by hand, match results entered manually, easy to make mistakes, awkward to enter predictions from a phone, no live leaderboard.
- We want the same game for 2026, automated: pull fixtures/results from a data source, auto-score, live leaderboard — instead of a manually maintained spreadsheet.

## Goals
- G1. Each participant enters a prediction (score) before a match starts — simple, mobile-friendly.
- G2. Points are awarded automatically based on match results.
- G3. Live leaderboard visible to all participants.
- G4. Minimal manual admin work during the tournament.

## Non-goals (for the friends-only variant)
- Money betting / payments.
- Public sign-up, anti-fraud, scaling to hundreds/thousands.
- Integration with Everstake corporate systems.

## Users
- **Players (5–15):** the same group as in 2022 (Sanya B, Petrovych, Zheka, Sanya M, Dimon, + newcomers). Action: enter a prediction, view the leaderboard.
- **Admin (1):** sets up the match calendar, enters results (or an auto-feed), opens/closes rounds.

## Scoring (ported from 2022 — needs confirmation)
Working baseline (typical "guess the score" rules):
- Exact score — N points.
- Correct outcome (W/D/L) + goal difference — M points.
- Correct winner/draw only — K points.
- Separate bonus for knockout stage / tournament winner (?).
> ⚠️ The exact 2022 rules were not preserved in the formulas — recover them from participants' memory or define fresh.

## Data / dependencies
- WC2026 calendar and results: manual entry by admin OR an external football-data API (e.g. football-data.org, API-Football).
- Old 2022 data — reference for structure only, not migrated.

## Open questions
1. Exact point-scoring rules (2022 or new)?
2. Prediction cutoff: how long before a match do we lock entries?
3. Do we predict the knockout bracket (who advances) separately from scores?
4. Match results: manual admin entry or auto-feed via API?
5. Who hosts (if web/bot) and who is the admin?

## Constraints / context
- Tournament starts 2026-06-11 → little development time remains; priority is to be ready before the group stage.
- Budget/infra: a friends project, keep cost and maintenance minimal.

## Candidate solutions (decision deferred)
- **A. Telegram bot** — lowest entry barrier for players, leaderboard in the chat, but needs hosting + state.
- **B. Web app** — most flexible (leaderboard, history), more work.
- **C. Google Sheets 2.0** — fast, formulas auto-tally, but entering predictions from a phone is awkward.
