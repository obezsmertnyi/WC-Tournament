# Repo map — where everything is

Lost in the files? This is the index. Everything is grouped by **purpose**.

## How this repo is organized
1. **Product** — the running app: `backend/`, `frontend/`, `docker-compose.yml`.
2. **Project docs** — what/why/how of the product: `docs/` (brief, requirements,
   architecture, ADRs, features, dev-setup, glossary) + `README.md`.
3. **Engineering process** — rules (`AGENTS.md`/`CLAUDE.md`), the build→verify→review
   loop (`WORKFLOW.md`/`LOOP.md`/`.claude/`), and verification (`scripts/`,
   `quality/`, `docs/qa/`, `evals/`).

## 🎮 The product (the app)
| Path | What |
|------|------|
| `backend/` | Go service — API + background jobs (scoring, FIFA sync, Telegram) |
| `frontend/` | React SPA (predictions, leaderboard, bracket, recap) |
| `docker-compose.yml`, `.env.example` | Run the stack locally |

## 🧠 Context / rules
| Path | What |
|------|------|
| **`AGENTS.md`** | Single source of truth: stack, guardrails, **bug-learned rules** |
| `CLAUDE.md` | One line — imports `@AGENTS.md` (no drift) |
| `.codex/config.toml`, `.aiexclude` | Codex pointer; agent context border |
| `docs/memory/` | Dynamic context / handoff between sessions |

## 📐 What & why — SDD
| Path | What |
|------|------|
| `docs/project-brief.md` | Narrative scope |
| **`docs/requirements.md`** | FR/NFR/TC/BC (numbered, MVP/Future) |
| `docs/mvp-capability-plan.md` | Capabilities → FR ownership |
| `docs/features/*/spec.md` | Given/When/Then per capability |
| `docs/adr/0001–0018` | Decisions (why it won) |
| `docs/domain-glossary.md`, `docs/dev-setup.md` | Vocabulary; how to build |

## ✅ Verification
| Path | What |
|------|------|
| `backend/**_test.go`, `frontend/src/**/*.test.ts` | Unit/integration tests |
| `backend/internal/scoring/scoring_evals_test.go`, `mcp/evals/`, `evals/README.md` | **Evals** (golden fixtures) |
| `scripts/gen-traceability.mjs` → `docs/qa/requirements-traceability-matrix.md`, `trace/trace.json` | FR→test map (generated) |
| `scripts/check-eval-ratchet.mjs`, `check-coverage-ratchet.mjs`, `quality/*-baseline.json` | Quality-only-tightens ratchets |
| `scripts/qa-verify.mjs` (`make qa`), `gate-status.mjs` (`make gates`) | Battery runner + gate roll-up |
| `docs/qa/` | **QA pack** — see [`docs/qa/README.md`](qa/README.md) |

## 🔁 Loop / process
| Path | What |
|------|------|
| `WORKFLOW.md`, `LOOP.md` | The build→verify→review→commit loop |
| `.claude/commands/*.md` | `/verify` `/trace` `/review` `/new-capability` |
| `.claude/hooks/{pre-commit,commit-msg}.sh`, `.claude/settings.json` | Deterministic guards + Claude Code hooks |

## 👀 Maker ≠ checker
| Path | What |
|------|------|
| `.claude/agents/*-reviewer.md` | Reviewer sub-agents (separate context) |
| `docs/qa/review-findings.json` | Adversarial review results (found+fixed bugs) |
| `.coderabbit.yaml` | External reviewer config (uk-UA) |

## 🎁 Tools & GenUI
| Path | What |
|------|------|
| `mcp/`, `.mcp.json` | Read-only MCP server + evals |
| `.claude/skills/wc-recap/` | Authored Agent Skill |
| `frontend/src/lib/recap.ts`, `components/MatchRecap.tsx` | GenUI AI recap (grounded + guardrail) |

## 🤖 AI assistant "Pitchside" (ADR-0017)
| Path | What |
|------|------|
| `backend/internal/gemini/gemini.go` | Guardrailed façade (L0–L3, fail-closed) — football-only |
| `backend/internal/gemini/vertex.go` | Vertex AI generator; `AI_ENABLED` opt-in + graceful 503 |
| `backend/internal/gemini/gemini_test.go` | Guardrail + grounding evals (`@trace FR-090..093, FR-100`) |
| `backend/internal/gemini/grounding.go` | Tools interface + fact structs + tool dispatch (grounding, ADR-0018) |
| `backend/internal/api/ai_tools.go` | Tools impl over storage — live overview/results/standings/leaderboard |
| `backend/internal/api/ai.go` | `/api/ai/{status,chat,card}` — auth-only, rate-limited, SSE |
| `frontend/src/pages/AI.tsx`, `lib/aiApi.ts`, `components/AiCard.tsx` | Chat tab + card UI |
| `docker-compose.gemini.yml` | Prod overlay: mounts keyless WIF creds + turns AI on |
| `docs/gemini-wif.md`, `docs/adr/0017-*.md` | WIF host config; the decision record |

## 🚀 CI/CD & release
| Path | What |
|------|------|
| `.github/workflows/{ci,release}.yml` | CI gates + multi-arch GHCR release |
| `VERIFY.md` | dev→prod verification, backup & rollback runbook |
| `scripts/check-doc-links.mjs` | Doc-graph integrity — every bootstrap pointer resolves |
| `.github/dependabot.yml`, `Makefile`, `docs/diagrams/*.svg` | Deps; dev tasks; architecture diagrams |
