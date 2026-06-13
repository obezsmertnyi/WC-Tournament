-- Tracks whether a finished match's result has already been posted to the
-- Telegram group, so the announcer never double-posts. NULL = not yet announced.
ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS announced_at TIMESTAMPTZ;
