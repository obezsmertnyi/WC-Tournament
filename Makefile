# WC-Tournament — developer tasks. Run `make help` for the list.
.DEFAULT_GOAL := help
SHELL := /usr/bin/env bash

BACKEND  := backend
FRONTEND := frontend

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ── Backend ──────────────────────────────────────────────────────────────────
.PHONY: build-backend
build-backend: ## Build the backend binary
	cd $(BACKEND) && go build ./...

.PHONY: test-backend
test-backend: ## Run backend tests (set DATABASE_URL for integration tests)
	cd $(BACKEND) && go test -race -cover ./...

.PHONY: vet
vet: ## go vet the backend
	cd $(BACKEND) && go vet ./...

.PHONY: fmt
fmt: ## gofmt the backend
	cd $(BACKEND) && gofmt -w .

.PHONY: vuln
vuln: ## Run govulncheck on the backend
	cd $(BACKEND) && go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# ── Frontend ─────────────────────────────────────────────────────────────────
.PHONY: build-frontend
build-frontend: ## Typecheck + build the frontend
	cd $(FRONTEND) && npm ci && npm run build

# ── Aggregate ────────────────────────────────────────────────────────────────
.PHONY: build
build: build-backend build-frontend ## Build backend + frontend

.PHONY: test
test: test-backend ## Run all tests

.PHONY: ci
ci: vet test-backend build-frontend ## Run the local equivalent of CI gates

.PHONY: qa
qa: ## Run the full verification battery + write the evidence report
	node scripts/qa-verify.mjs

.PHONY: gates
gates: ## Gate roll-up: traceability + eval + coverage ratchets (PASS/FAIL/SKIP)
	node scripts/gate-status.mjs

.PHONY: trace
trace: ## Regenerate the traceability matrix + trace.json
	node scripts/gen-traceability.mjs

# ── Docker / local stack ─────────────────────────────────────────────────────
.PHONY: up
up: ## Start the local stack (docker compose)
	docker compose up -d --build

.PHONY: down
down: ## Stop the local stack
	docker compose down

.PHONY: logs
logs: ## Tail stack logs
	docker compose logs -f

# ── Release ──────────────────────────────────────────────────────────────────
.PHONY: release
release: ## Tag and push a release: make release VERSION=v0.1.0
	@[ -n "$(VERSION)" ] || { echo "usage: make release VERSION=v0.1.0"; exit 1; }
	@echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+' || { echo "VERSION must be vX.Y.Z"; exit 1; }
	git tag -a "$(VERSION)" -m "Release $(VERSION)"
	git push origin "$(VERSION)"
	@echo "Pushed $(VERSION) → the release workflow will build images and cut the GitHub Release."
