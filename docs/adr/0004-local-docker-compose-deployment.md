# ADR-0004: Local-first deployment via docker-compose

**Status:** Accepted · **Date:** 2026-06-13

## Context
This is a friends project under tight time pressure (tournament already live). A GCP project is available if needed later, but the immediate goal is a working local solution.

## Decision
Deploy with **`docker-compose`** — three services: `db` (postgres:17, named volume, healthcheck), `backend` (Go, exposes `:8080`), `frontend` (nginx serving the SPA, proxies `/api`). Config via `.env`; no secrets committed. Single host, local-first.

Database migrations run as a **dedicated one-shot step / `migrate` container**, not silently on every backend boot, to avoid races on restart/scale.

## Consequences
- One `docker-compose up` brings the whole stack online → low operational overhead.
- Cheap and reproducible; easy to hand to a friend to run.
- A future GCP path (Cloud Run backend + Cloud SQL + GCS/CDN frontend) remains open but is explicitly out of scope for this iteration.

## Alternatives considered
- **Direct GCP deploy now (Cloud Run + Cloud SQL)** — more setup and cost than warranted before the app even works; deferred.
- **Kubernetes** — vastly over-engineered for ≤15 users.
