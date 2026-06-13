-- Small key/value store for bot/runtime state that must survive restarts,
-- e.g. the last day the morning digest was posted (so it fires once per day).
CREATE TABLE IF NOT EXISTS app_state (
    key        TEXT PRIMARY KEY,
    value      TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
