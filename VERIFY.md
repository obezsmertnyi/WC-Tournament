# VERIFY.md — dev→prod verification, backup & rollback

Runbook for shipping a release, with the current change being the **football AI
assistant "Pitchside"** (ADR-0017). The assistant is **opt-in**: it stays OFF
unless `AI_ENABLED=true`, which only the prod overlay (`docker-compose.gemini.yml`)
sets. If the WIF credentials are absent it degrades to `503` and the rest of the
app is unaffected.

Verify by **running the commands**, not by assertion. Roll out and verify; keep the
backup until the new version has been healthy for a day.

---

## 0. Owner-only prerequisites (do first)

- **`.env` secrets** — not in git (gitignored, not in history). Rotation is
  **owner-deferred** and is *not* a deploy blocker (finding S1 noted it as advisable
  only because the values were readable by in-session review tooling; revisit later).
- **WIF host infra present** (for AI): `/opt/workload-identity/jwt_svid` kept fresh
  by Teleport `tbot`, and `/etc/gemini/cred.json` (external-account config). Confirm:
  `sudo gemini-check gemini-2.5-flash "test"` returns text on the host.

## 1. Pre-deploy — local proof (must be green)

```bash
# backend: fmt, vet, build, race+integration tests, scoring + guardrail evals
cd backend && gofmt -l . && go vet ./... && go build ./...
go test -race -cover ./...                      # needs DATABASE_URL + JWT_SECRET
go test -tags=evals ./internal/scoring/
go test ./internal/gemini/                       # AI guardrail evals (12 cases)
# frontend: typecheck + build + unit tests
cd ../frontend && npm run build && npm test
# repo gates
cd .. && node scripts/gen-traceability.mjs --check \
      && node scripts/check-eval-ratchet.mjs --check \
      && node scripts/check-doc-links.mjs --check
DATABASE_URL=... JWT_SECRET=... node scripts/check-coverage-ratchet.mjs --check
```

All green ⇒ proceed. A red gate is a **STOP** (fix the cause, never bypass).

## 2. Backup — MANDATORY before touching prod

On the **prod host**, before pulling the new images:

```bash
mkdir -p ~/wc2026-backup && cd ~/wc2026-backup
TS=$(date +%Y%m%d-%H%M%S)

# 2a. Save the CURRENTLY-RUNNING images (so we can roll back to the exact bits).
docker image save -o images-$TS.tar \
  "$(docker inspect --format '{{.Config.Image}}' wc2026-backend)" \
  "$(docker inspect --format '{{.Config.Image}}' wc2026-frontend)"
docker inspect --format '{{.Config.Image}}' wc2026-backend wc2026-frontend | tee images-$TS.txt

# 2b. Dump the database (custom format → pg_restore-able).
docker exec wc2026-db pg_dump -U wc -Fc wc2026 > db-$TS.dump
ls -lh images-$TS.tar db-$TS.dump    # sanity: non-zero sizes
```

Do not proceed until both files exist and are non-empty.

## 3. Deploy (rsync + build on prod — no GitHub CI)

Prod builds images locally from `/srv/WC-Tournament` (the running images are
`wc2026-*:latest`, not GHCR tags). We deploy by syncing the repo and rebuilding on
the host — **no `git push` / no `make release`** (that would tag → trigger paid
GitHub CI). Tagging a release is a *separate, optional* step done only **after**
this deploy is verified (§6).

```bash
# 3a. Sync local → prod, preserving prod's .env and git state.
rsync -az --delete \
  --exclude '.git' --exclude 'node_modules' --exclude '.env' \
  --exclude 'dist' --exclude '*.workflow-state.toon' \
  ./ <PROD_HOST>:/srv/WC-Tournament/

# 3b. On the prod host — validate nginx, then rebuild WITH the AI overlay.
ssh <PROD_HOST>
cd /srv/WC-Tournament
docker run --rm -v "$PWD/frontend/nginx.conf:/etc/nginx/conf.d/default.conf:ro" nginx:alpine nginx -t
docker compose -f docker-compose.yml -f docker-compose.gemini.yml up -d --build
docker compose -f docker-compose.yml -f docker-compose.gemini.yml ps   # all healthy
```

**Prereq (one-time, host):** create the WIF cred-config the overlay mounts (the
JWT-SVID exists but `/etc/gemini/cred.json` does not yet):
```bash
sudo mkdir -p /etc/gemini
sudo gcloud iam workload-identity-pools create-cred-config \
  projects/<GCP_PROJECT_NUMBER>/locations/global/workloadIdentityPools/<WIF_POOL>/providers/<WIF_PROVIDER> \
  --service-account=gemini-agent@<GCP_PROJECT>.iam.gserviceaccount.com \
  --credential-source-file=/opt/workload-identity/jwt_svid \
  --credential-source-type=text --output-file=/etc/gemini/cred.json
```

## 4. Post-deploy verification

```bash
# health
curl -fsS http://localhost:8081/healthz        # backend (direct)
curl -fsS http://localhost:8080/                # frontend (SPA)

# AI wired? expect {"available":true} (503/false ⇒ WIF creds not seen → check mounts)
curl -fsS https://wc2026.mtgrd-das.app/api/ai/status

# backend log line on boot should read: "gemini: AI assistant ENABLED"
docker logs wc2026-backend 2>&1 | grep -i 'gemini:'
```

In the UI (logged in, any tier incl. `none`): open the **AI** tab →
- on-topic ("Who are Argentina's key players for WC 2026?") → a streamed answer;
- off-topic ("write me some Python") → the single-sentence refusal, no answer;
- 🃏 card ("Messi") → a structured card with a confidence badge.

## 5. Rollback

If the new version misbehaves:

```bash
# App only (keep DB): restore previous images + restart without the overlay if AI is the culprit.
cd ~/wc2026-backup
docker load -i images-$TS.tar
# re-tag / re-pin compose to the previous image ids from images-$TS.txt, then:
docker compose -f docker-compose.yml up -d          # (omit the gemini overlay to disable AI)

# DB restore (only if data was affected — destructive; recreates schema):
docker exec -i wc2026-db pg_restore -U wc -d wc2026 --clean --if-exists < db-$TS.dump
```

Fastest AI-only kill switch: bring the stack up **without** `docker-compose.gemini.yml`
(or unset `AI_ENABLED`) → assistant returns 503, everything else keeps working.

---

## Per-component local tests (pre-deploy proof)

| Component | Command | Asserts |
|---|---|---|
| AI guardrail | `go test ./internal/gemini/` | input caps, off-topic refusal, classify fail-closed, card unmarshal, availability/503 (`@trace FR-090..093`) |
| Scoring | `go test -tags=evals ./internal/scoring/` | golden-fixture scoring invariants |
| Frontend | `cd frontend && npm test` | recap + UI unit tests |
| Doc graph | `node scripts/check-doc-links.mjs --check` | no bootstrap doc points at a missing file |
| Traceability | `node scripts/gen-traceability.mjs --check` | matrix fresh + non-regressing |
