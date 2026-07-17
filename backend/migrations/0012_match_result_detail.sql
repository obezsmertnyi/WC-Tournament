-- 0012: how a knockout that was level after 90' was actually decided, so the AI
-- assistant can say it precisely instead of guessing ("probably ET or penalties").
-- ADR-0020 stores the REGULATION score as home_score/away_score and discards the
-- aet/penalty detail; this keeps that detail for display/grounding without
-- affecting scoring. Compact code set by winners.Run from the FIFA live detail:
--   'et:H:A'  = won in extra time, H:A is the aet (full-time incl. ET) score
--   'pen:H:A' = won on a penalty shootout, H:A is the shootout score
--   NULL/''   = decided in normal time (no extra detail)
-- Nullable, additive, non-breaking. Applied idempotently on boot.

ALTER TABLE matches ADD COLUMN IF NOT EXISTS result_detail TEXT;
