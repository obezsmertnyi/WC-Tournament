# WC-Tournament

A friends-only World Cup 2026 score-prediction pool — predict match scores, earn points, climb a live leaderboard. **PoC** (proof it on the friends group, possibly grow into a real product later), built production-grade.

## Stack
- **Backend:** Go 1.24 + Gin + Postgres 17
- **Frontend:** React + Vite + TypeScript (Tailwind + shadcn/ui)
- **Deploy:** docker-compose (local-first)
- **Data:** official FIFA API (ESPN / football-data.org fallbacks)
- **Auth:** Google OAuth · **Notifications:** Telegram

## Scoring
Exact score **3** · correct outcome **1** · wrong **0**. Knockout adds **+1** for the correct advancer (penalties count for that pick only). Optional tournament bonuses (champion / finalist / top scorer). No fractions.

## Docs
Design and decisions live in [`docs/`](docs/): the [project brief](docs/project-brief.md), the [architecture](docs/architecture.md), and the [ADRs](docs/adr/README.md).

## Setup
```bash
cp .env.example .env   # fill in real values; .env is gitignored — never commit secrets
docker compose up
```

## Status
Design complete. Build proceeds in milestones M0→M6 (see architecture §9).
