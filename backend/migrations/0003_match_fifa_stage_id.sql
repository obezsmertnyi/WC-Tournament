-- M3: store the FIFA IdStage per match so the live match-detail endpoint can
-- build the live football URL (.../live/football/17/285023/{idStage}/{idMatch}).
-- The calendar response already carries IdStage per match; the sync now captures
-- it. Nullable: pre-existing rows backfill on the next sync.
ALTER TABLE matches ADD COLUMN IF NOT EXISTS fifa_stage_id TEXT;
