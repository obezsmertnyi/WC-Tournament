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

## MCP server (agent tool)
Build it (`cd mcp && npm run build`), then connect via repo-root `.mcp.json`
(Claude Code auto-detects). Point `WC_API_BASE` at a read-reachable API (local
`docker compose up`); never bake a session token in. See `mcp/README.md`.

## Where things are
Product design & decisions: `docs/` (product-brief, architecture, ADRs). What to
build & why: `docs/requirements.md` + `docs/mvp-capability-plan.md` +
`docs/features/<cap>/spec.md`. How we work: `WORKFLOW.md` / `LOOP.md`. Live
status: `CHECKLIST.md`.
