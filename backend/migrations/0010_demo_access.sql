-- Demo mode + per-user access level.
--
-- access_level gates what a user may do while demo mode is ON (see app_state
-- key 'demo_mode'). Existing users default to 'rw' so enabling demo mode never
-- locks out current participants; only NEW self-service (Google) sign-ups land
-- in the restricted tiers until an admin grants access.
--
--   none — browse the UI only (no other players' data, cannot participate)
--   ro   — also see other players' predictions, leaderboard, audit
--   rw   — also participate (submit predictions and bonus picks)
--
-- When demo mode is OFF the column is ignored and everyone has full access,
-- preserving the pre-demo behaviour.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS access_level TEXT NOT NULL DEFAULT 'rw'
        CHECK (access_level IN ('none', 'ro', 'rw'));
