-- Tournament-wide bonus picks (champion / finalist / top scorer) are credited to
-- the leaderboard ONLY once their outcome is resolved AND the pick was correct.
-- Until then `tier_points` is just the *potential* award; `awarded` gates whether
-- it actually counts. Defaults to FALSE so no bonus is granted before resolution.
ALTER TABLE tournament_picks
    ADD COLUMN IF NOT EXISTS awarded BOOLEAN NOT NULL DEFAULT FALSE;
