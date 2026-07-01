# Repo map — where everything is and what you actually need

Lost in the files? This is the index. Everything is grouped by **purpose**. The
`Submission?` column flags what matters for the fwdays homework grade.

> **Start here:** [`docs/SUBMISSION.md`](SUBMISSION.md) (ready-to-paste PR text) →
> [`docs/agentic-engineering.md`](agentic-engineering.md) (the 5 proof points) →
> [`CHECKLIST.md`](../CHECKLIST.md) (live status).

## 🎮 The product (the app)
| Path | What |
|------|------|
| `backend/` | Go service — API + background jobs (scoring, FIFA sync, Telegram) |
| `frontend/` | React SPA (predictions, leaderboard, bracket, recap) |
| `docker-compose.yml`, `.env.example` | Run the stack locally |

## 🧠 Context / rules (proof point 1)
| Path | What | Submission? |
|------|------|:--:|
| **`AGENTS.md`** | Single source of truth: stack, guardrails, **bug-learned rules** | ⭐ |
| `CLAUDE.md` | One line — imports `@AGENTS.md` (no drift) | ⭐ |
| `.codex/config.toml`, `.aiexclude` | Codex pointer; agent context border | ✓ |
| `docs/memory/`, `current-state.md` | Dynamic context / fast handoff | ✓ |

## 📐 What & why — SDD (proof point 5)
| Path | What | Submission? |
|------|------|:--:|
| `docs/product-brief.md` | Narrative scope | ✓ |
| **`docs/requirements.md`** | FR/NFR/TC/BC (numbered, MVP/Future) | ⭐ |
| `docs/mvp-capability-plan.md` | Capabilities → FR ownership | ✓ |
| `docs/features/*/spec.md` | Given/When/Then per capability | ⭐ |
| `docs/adr/0001–0016` | Decisions (why it won) | ⭐ |
| `docs/domain-glossary.md`, `docs/dev-setup.md` | Vocabulary; how to build | |

## ✅ Verification (proof point 4)
| Path | What | Submission? |
|------|------|:--:|
| `backend/**_test.go`, `frontend/src/**/*.test.ts` | Unit/integration tests | ⭐ |
| `backend/internal/scoring/scoring_evals_test.go`, `mcp/evals/`, `evals/README.md` | **Evals** (golden fixtures) | ⭐ |
| `scripts/gen-traceability.mjs` → `docs/qa/requirements-traceability-matrix.md`, `trace/trace.json` | FR→test map (generated) | ⭐ |
| `scripts/check-eval-ratchet.mjs`, `check-coverage-ratchet.mjs`, `quality/*-baseline.json` | Quality-only-tightens ratchets | ✓ |
| `scripts/qa-verify.mjs` (`make qa`), `gate-status.mjs` (`make gates`) | Battery runner + gate roll-up | ✓ |
| `docs/qa/` | **QA pack** — see [`docs/qa/README.md`](qa/README.md) | ⭐ |

## 🔁 Loop / process (proof point 2)
| Path | What |
|------|------|
| `WORKFLOW.md`, `LOOP.md` | The build→verify→review→commit loop |
| `.claude/commands/*.md` | `/verify` `/trace` `/review` `/new-capability` |
| `.claude/hooks/{pre-commit,commit-msg}.sh`, `.claude/settings.json` | Deterministic guards + Claude Code hooks |
| `CHECKLIST.md` | Live loop driver (status of everything) |

## 👀 Maker ≠ checker (proof point 3)
| Path | What | Submission? |
|------|------|:--:|
| `.claude/agents/*-reviewer.md` | Reviewer sub-agents (separate context) | ⭐ |
| `docs/qa/review-findings.json` | Adversarial review results (found+fixed bugs) | ⭐ |
| `.coderabbit.yaml` | External reviewer config (uk-UA) | ✓ |

## 🎁 Bonus — tools & GenUI
| Path | What | Submission? |
|------|------|:--:|
| `mcp/`, `.mcp.json` | Read-only MCP server + evals | ⭐ |
| `.claude/skills/wc-recap/` | Authored Agent Skill | ✓ |
| `frontend/src/lib/recap.ts`, `components/MatchRecap.tsx` | GenUI AI recap (grounded + guardrail) | ⭐ |

## 🚀 CI/CD & release
| Path | What |
|------|------|
| `.github/workflows/{ci,release}.yml` | CI gates + multi-arch GHCR release |
| `.github/dependabot.yml`, `Makefile`, `docs/diagrams/*.svg` | Deps; dev tasks; architecture diagrams |

## 📄 Submission artifacts
| Path | What | Submission? |
|------|------|:--:|
| **`docs/SUBMISSION.md`** | Ready uk PR text + handoff steps | ⭐ |
| **`docs/agentic-engineering.md`** | Maps to the 5 graded proof points | ⭐ |
| `docs/qa/slide-coverage-audit.md` | 91-slide deck coverage + adoption notes | ✓ |
| `README.md` | Badges, diagrams, overview | ✓ |

## 👉 What only YOU can do (no gh access for the agent)
1. **Rotate the 4 secrets in `.env`** (Telegram / Google OAuth / `JWT_SECRET` / `ADMIN_PASSWORD`) — risk R-01.
2. **Fork** the task repo → enable **CodeRabbit** → open the **PR** (paste `docs/SUBMISSION.md`).
3. **Record the 1–2 min demo video** (`docs/qa/demo-script.md`) and link it in the PR.
