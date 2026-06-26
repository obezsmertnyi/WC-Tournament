-- The actual advancing team for a knockout match, resolved from FIFA's live
-- detail (WinnerTeamID) so extra-time / penalty results are correct even when
-- the 90-minute scoreline is a draw. Used by the knockout "+1 advancer" scoring.
-- NULL until resolved (and always NULL for group-stage matches).
ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS winner_team_id BIGINT REFERENCES teams (id);
