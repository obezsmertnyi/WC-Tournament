# ADR-0011: Notifications via Telegram, friendly Ukrainian copy

**Status:** Accepted · **Date:** 2026-06-13

## Context
The top pre-mortem risk is players forgetting to submit a prediction before kickoff. The owner has a Telegram channel and wants reminders there, written in **Ukrainian with a friendly, fun tone** — not dry system messages. This is a PoC, so keep it simple but behind a clean interface for a future product.

## Decision
**Telegram only — no email, no web-push** (owner decision). A decoupled `Notifier` interface keeps a future channel possible without touching callers, but the only implementation we build is the **Telegram bot**.

- **Notification set (kept deliberately small so it stays fun, not spammy):**
  | Type | Target | When | Example (Ukrainian) |
  |---|---|---|---|
  | Nudge | DM | ~1h before kickoff, only if no prediction yet | «⚽️ За годину Франція–Бразилія, а ти ще ні 👀» |
  | Round start | channel | start of a matchday | «🏟 Тур 3 пішов — кидайте прогнози!» |
  | Round summary | channel | after the day's last match | «📊 Тур 3: 🥇 Дімон +7, в лідерах Жека» |
  | Bonus deadline | channel + DM | before knockout | «⏳ Завтра закриваємо вибір чемпіона» |
  Channel sees only round start / summary (+ rare deadline); personal nudges are DM-only.
- **Direct messages** require each player to `/start` the bot once → we store `telegram_chat_id`. The shared channel id comes from `TELEGRAM_CHAT_ID`.
- **Copy is Ukrainian and playful** (templated, with light variation so it doesn't feel robotic) — e.g. "⚽️ За годину Аргентина–Бразилія, а ти ще не поставив! Не лажай 😏". Templates live in code/config, not the docs.
- **Secrets:** `TELEGRAM_BOT_TOKEN` and the channel/chat ID come from `.env` / Secret Manager — never committed, never pasted into issues or chat. If a token leaks, rotate via @BotFather.
- **Delivery is recorded** in `notifications` (kind, channel, sent_at, status) to avoid duplicate nudges and for the audit/debug trail.
- **Scheduling:** the existing cron (`internal/scheduler`) computes "who is missing a prediction for matches kicking off in the next N hours" and dispatches.

## Consequences
- Directly mitigates the forgotten-prediction risk; feels personal and fun, fitting a friends pool.
- DM reminders require a one-time `/start` per player (a profile step); without it, that player only gets channel broadcasts.
- Rate-limit outbound Telegram calls to stay within bot API limits.

## Alternatives considered
- **Email only** — higher-friction, easy to ignore; Telegram is where this group already is.
- **No reminders** — leaves the #1 pre-mortem risk unmitigated. Rejected.
- **English/neutral copy** — owner explicitly wants Ukrainian and playful. Rejected.
