# Dev setup

One page to get building and verifying locally. Rules/context: `AGENTS.md`.

## Prerequisites
- Go `1.26.4`, Node `26` (build) / `20+` (dev), Docker + docker-compose.
- `cp .env.example .env` and fill values (never commit `.env` — it's gitignored).

## Run the stack
```bash
docker compose up -d --build      # db + backend + frontend
# app:     http://localhost:8080
# backend: http://localhost:8081  (direct, for debugging)
```

## Build & verify (no Docker)
```bash
make build          # backend + frontend
make test           # backend tests (set DATABASE_URL for the integration tests)
make ci             # local equivalent of the CI gates
make help           # all targets

# component-level
cd backend  && go test ./... && go test -tags=evals ./internal/scoring/
cd frontend && npm ci && npm run build && npm test        # tsc + vite + vitest
cd mcp      && npm ci && npm run typecheck && npm run eval # MCP server + evals
node scripts/gen-traceability.mjs --check                 # traceability fresh + non-regressing
```

## Optional local guards
```bash
ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit   # fast gates on commit
pre-commit install                                               # if using .pre-commit-config.yaml
```

## Release
```bash
make release VERSION=v0.1.x   # tags + pushes → CI builds multi-arch GHCR images + a GitHub Release
```

## AI assistant (local)
Off by default. To try "Pitchside" locally, authenticate ADC and enable the flag:
```bash
gcloud auth application-default login          # keyless; no API/SA key
export AI_ENABLED=true GOOGLE_GENAI_USE_VERTEXAI=true \
       GOOGLE_CLOUD_PROJECT=<GCP_PROJECT> GOOGLE_CLOUD_LOCATION=us-central1
```
Leave `AI_ENABLED` unset → the assistant logs "DISABLED" and `/api/ai/*` returns 503
(the rest of the app is unaffected). Prod uses the `docker-compose.gemini.yml`
overlay instead (keyless WIF mounts). Guardrail evals: `go test ./internal/gemini/`.
See ADR-0017 and `docs/gemini-wif.md`.

## MCP server (agent tool)
Build it (`cd mcp && npm run build`), then connect via repo-root `.mcp.json`
(Claude Code auto-detects). Point `WC_API_BASE` at a read-reachable API (local
`docker compose up`); never bake a session token in. See `mcp/README.md`.

## Where things are
Product design & decisions: `docs/` (project-brief, architecture, ADRs). What to
build & why: `docs/requirements.md` + `docs/mvp-capability-plan.md` +
`docs/features/<cap>/spec.md`. How we work: `WORKFLOW.md` / `LOOP.md`. Live
status: `CHECKLIST.md`.
