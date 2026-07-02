# Repo map — where everything is and what you actually need

Lost in the files? This is the index. Everything is grouped by **purpose**. The
`Submission?` column flags what matters for the fwdays homework grade.

> **Start here:** [`docs/SUBMISSION.md`](SUBMISSION.md) (ready-to-paste PR text) →
> [`docs/agentic-engineering.md`](agentic-engineering.md) (the 5 proof points) →
> [`CHECKLIST.md`](../CHECKLIST.md) (live status).

## How this repo is organized (three layers)
1. **Product** — the running app: `backend/`, `frontend/`, `docker-compose.yml`.
2. **Project docs** — what/why/how of the product: `docs/` (brief, requirements,
   architecture, ADRs, features, dev-setup, glossary) + `README.md`.
3. **Engineering process & submission** — the agentic layer: rules (`AGENTS.md`/
   `CLAUDE.md`), the loop (`WORKFLOW.md`/`LOOP.md`/`CHECKLIST.md`/`.claude/`),
   verification (`scripts/`, `quality/`, `docs/qa/`, `evals/`), and the homework
   deliverables (`docs/agentic-engineering.md`, `docs/SUBMISSION.md`).

Reference material (the course deck, project-factory) is **not** committed — only
our *decisions* about it live here (crosswalk below; `docs/qa/slide-coverage-audit.md`).

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
| `docs/project-brief.md` | Narrative scope | ✓ |
| **`docs/requirements.md`** | FR/NFR/TC/BC (numbered, MVP/Future) | ⭐ |
| `docs/mvp-capability-plan.md` | Capabilities → FR ownership | ✓ |
| `docs/features/*/spec.md` | Given/When/Then per capability | ⭐ |
| `docs/adr/0001–0018` | Decisions (why it won) | ⭐ |
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

## 📄 Submission artifacts
| Path | What | Submission? |
|------|------|:--:|
| **`docs/SUBMISSION.md`** | Ready uk PR text + handoff steps | ⭐ |
| **`docs/agentic-engineering.md`** | Maps to the 5 graded proof points | ⭐ |
| `docs/qa/slide-coverage-audit.md` | 91-slide deck coverage + adoption notes | ✓ |
| `README.md` | Badges, diagrams, overview | ✓ |

## 🔀 Naming crosswalk — reference (project-factory / deck) → ours
We deliberately kept our own naming (we're an app, not the framework). This maps
their term to our file so nothing looks "missing" — full rationale in
[`docs/qa/slide-coverage-audit.md`](qa/slide-coverage-audit.md) (adoption section)
and [`docs/qa/quality-gates.md`](qa/quality-gates.md) (deviations).

| Reference | Ours | Note |
|-----------|------|------|
| `openspec/specs/<cap>/spec.md` | `docs/features/<cap>/spec.md` | Given/When/Then; no OpenSpec CLI |
| `openspec/changes/<slice>/{proposal,design,tasks}.md` | `docs/adr/*` + `docs/mvp-capability-plan.md` | decisions in ADRs; slices in the plan |
| `MASTER-PROMPT.md` | `WORKFLOW.md` + `LOOP.md` | orchestration + the loop |
| `checklists/quality-gates.md` (G0–G8) | `docs/qa/quality-gates.md` | same gates, mapped to our commands |
| `scripts/check-traceability.mjs` | `scripts/gen-traceability.mjs` | generates matrix + `trace/trace.json` + `--check` |
| `check-coverage-ratchet` / `check-eval-ratchet` / `qa-verify` / `gate-status` | **same names** in `scripts/` | adopted as-is |
| `check-trajectory` / `check-recordings` / `check-a11y` / `vision-verify` | — | skipped (see audit) |
| `templates/docs/requirements.template.md` | `docs/requirements.md` | FR/NFR/TC/BC |
| `templates/docs/mvp-capability-plan.template.md` | `docs/mvp-capability-plan.md` | FR ownership |
| `templates/docs/current-state.template.md` | `current-state.md` + `docs/memory/` | handoff |
| `templates/docs/context-architecture.template.md` | `docs/adr/0013` + `.aiexclude` | static/dynamic + border |
| `templates/adr.template.md` | `docs/adr/*.md` | MADR format |
| `templates/docs/qa-pack.template.md` | `docs/qa/{README,test-plan,manual-test-plan,requirements-traceability-matrix,risk-register,mvp-acceptance-report}.md` | the QA pack |
| `agents/*.md` (11 roles) | `.claude/agents/{scoring-correctness,security}-reviewer.md` | 2 reviewers; other 9 = orchestration, skipped |
| `.claude/workflows/*.js` (fan-out) | `.claude/commands/{verify,trace,review,new-capability}.md` | our loop steps |
| `skills/project-factory/` | `.claude/skills/wc-recap/` | we author one skill |
| `evals/cases/*.eval.ts` | `scoring_evals_test.go` (`-tags=evals`), `mcp/evals/`, `recap.test.ts` | golden fixtures |
| `quality/{coverage,eval}-baseline.json` | **same** | ratchets |
| `docs/qa/demo-recordings/` (headless) | `docs/qa/demo-script.md` (manual video) | automation deferred |
| `AGENTS.md` (SSOT) + `CLAUDE.md`→`@AGENTS.md` | **same** | one ruleset, both tools |

## 👉 What only YOU can do (no gh access for the agent)
1. **Rotate the 4 secrets in `.env`** (Telegram / Google OAuth / `JWT_SECRET` / `ADMIN_PASSWORD`) — risk R-01.
2. **Fork** the task repo → enable **CodeRabbit** → open the **PR** (paste `docs/SUBMISSION.md`).
3. **Record the 1–2 min demo video** (`docs/qa/demo-script.md`) and link it in the PR.
