-- Tracks whether the pre-match "you haven't predicted yet" Telegram reminder has
-- already been sent for a match, so it fires at most once. NULL = not reminded.
ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS reminded_at TIMESTAMPTZ;
