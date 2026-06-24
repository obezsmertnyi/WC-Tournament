-- Aggregated goal tally per scorer, rebuilt periodically from finished-match
-- detail data (the calendar has only scores; scorer names live in match detail).
-- Drives the public "top scorers" board and the top-scorer bonus resolution.
CREATE TABLE IF NOT EXISTS top_scorers (
    name       TEXT PRIMARY KEY,
    team_code  TEXT        NOT NULL DEFAULT '',
    goals      INTEGER     NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
