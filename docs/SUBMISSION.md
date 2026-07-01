# Здача ДЗ — матеріали для PR

Готовий текст для Pull Request у форк `koldovsky/2026-fwdays-agentic-greenfield-task`
(мова — українська, бо CodeRabbit рев'юить `uk-UA`). Заповни `<...>` перед
відправкою.

---

## Автор
Олександр Безсмертний <!-- підтвердь написання свого справжнього імені -->

## Проєкт
**WC-Tournament** — приватний пул прогнозів на Кубок світу 2026 для друзів:
прогнозуй рахунки, отримуй бали, живий лідерборд + Telegram-бот. Стек: Go 1.26 +
Gin + Postgres 17 (бекенд), React + Vite + TS (фронт), docker-compose, CI/CD з
релізами в GHCR. Код лежить в окремому репозиторії (посилання нижче).

## Відео-демо (1–2 хв)
Video: `<ВСТАВ ПОСИЛАННЯ — YouTube/Loom/Drive>`
<!-- сценарій запису: docs/qa/demo-script.md -->

## Які практики Agentic Engineering застосовано
Повний розбір із посиланнями на артефакти — `docs/agentic-engineering.md`. Стисло:

- **Контекст-інженерія (static vs dynamic):** правила в `AGENTS.md` (SSOT,
  пінновані версії, guardrails), `CLAUDE.md` імпортує його через `@AGENTS.md`,
  `.codex/config.toml` — для Codex; динамічний контекст у `docs/memory/` +
  `current-state.md` (ADR-0013). Один набір правил, обидва інструменти, без дрейфу.
- **Loop engineering:** цикл spec→implement→trace→verify→review→commit
  (`WORKFLOW.md`/`LOOP.md`), драйвиться `CHECKLIST.md` у лупі; команди
  `.claude/commands/` + guard `.claude/hooks/pre-commit.sh`. Не покроковий промптинг.
- **Maker ≠ checker:** два рев'юер-субагенти (`.claude/agents/`) окремим проходом
  знайшли й виправили **реальні баги** (мертвий guardrail рекапу, витік цифр
  стадії, підробний трейс) → `docs/qa/review-findings.json`. CodeRabbit —
  зовнішній другий чекер.
- **Верифікація:** 5-рівнева піраміда в CI (static → unit → integration на
  **реальному Postgres** → **evals** golden-fixtures для scoring і MCP → відео).
  Кожна вимога доведена тестом `@trace <FR-id>`; матриця трасування **генерується**,
  CI падає якщо застаріла/регресує (ADR-0014). Покриття 17/19 FR.
- **SDD:** `docs/requirements.md` (FR/NFR/TC/BC) → `mvp-capability-plan.md` →
  `docs/features/<cap>/spec.md` (Given/When/Then) → 16 ADR.
- **Інструменти & MCP:** read-only **MCP-сервер** (`mcp/`, security-first, zod,
  reveal-lock, ADR-0015) + **GenUI** AI-рекап матчу з anti-hallucination guardrail
  (ADR-0016). Скіли: `mcp-secure-server-dev`, `svg-diagram`, bmad-review,
  workflow-memory.
- **Я вирішував / агент робив:** я — продукт, інваріанти (scoring/приватність),
  вибір компонентів, тріаж кожної знахідки рев'ю; агент — ресерч, чернетки
  специфікацій/ADR/тестів/CI, дифи, трасування, і окремим агентом — adversarial
  рев'ю.

## (Опційно) Посилання на код
https://github.com/obezsmertnyi/WC-Tournament

### Чекліст
- [x] Вказано справжнє імʼя
- [ ] Додано посилання на відео-демо (1–2 хв) — *встав лінк вище*
- [x] Описано застосовані практики Agentic Engineering
- [x] Результат робочий і доведений до кінця

---

## Кроки здачі (виконує власник — потрібен GitHub/веб, не gh CLI)
1. **Ротуй секрети** з `.env` (знахідка S1): Telegram-токен (BotFather), Google
   OAuth secret (GCP), перегенеруй `JWT_SECRET` і `ADMIN_PASSWORD`. У репо їх нема
   (gitignored, історія чиста), але вони були відкриті рев'ю-тулінгу.
2. **Fork** `koldovsky/2026-fwdays-agentic-greenfield-task` (приносить PR-шаблон +
   `.coderabbit.yaml`).
3. **Увімкни CodeRabbit** на форку (безкоштовно для публічних репо).
4. Поклади проєкт на окрему гілку у форку **або** лишай WC-Tournament окремим репо
   і дай посилання в PR (вже вставлено вище).
5. **Запиши відео** (1–2 хв) за `docs/qa/demo-script.md`, встав лінк у PR.
6. **Відкрий PR**, встав цей текст, познач чеклист, прочитай фідбек CodeRabbit,
   поітеруй — і надішли посилання на PR як здачу.
