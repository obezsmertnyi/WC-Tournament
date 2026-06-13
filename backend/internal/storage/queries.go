package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// UpsertTeam inserts or updates a team keyed by fifa_id and returns its
// local id. Idempotent: re-running with the same fifa_id updates in place.
func (s *Store) UpsertTeam(ctx context.Context, q pgx.Tx, t UpsertTeam) (int64, error) {
	const sql = `
		INSERT INTO teams (fifa_id, name, code, flag_url, group_label)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (fifa_id) DO UPDATE SET
			name        = EXCLUDED.name,
			code        = EXCLUDED.code,
			flag_url    = EXCLUDED.flag_url,
			group_label = EXCLUDED.group_label
		RETURNING id`
	var id int64
	if err := q.QueryRow(ctx, sql, t.FifaID, t.Name, t.Code, t.FlagURL, t.GroupLabel).Scan(&id); err != nil {
		return 0, fmt.Errorf("upsert team %s: %w", t.FifaID, err)
	}
	return id, nil
}

// teamIDByFifaID resolves a team's local id from its fifa_id within a tx.
// Returns (0,false,nil) when fifaID is empty (unassigned slot).
func (s *Store) teamIDByFifaID(ctx context.Context, q pgx.Tx, fifaID string) (int64, bool, error) {
	if fifaID == "" {
		return 0, false, nil
	}
	var id int64
	err := q.QueryRow(ctx, `SELECT id FROM teams WHERE fifa_id = $1`, fifaID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("lookup team %s: %w", fifaID, err)
	}
	return id, true, nil
}

// UpsertMatch inserts or updates a match keyed by fifa_id. It NEVER overwrites
// a row whose result_source = 'manual' (admin outage fallback wins per
// ADR-0006). Idempotent on fifa_id.
func (s *Store) UpsertMatch(ctx context.Context, q pgx.Tx, m UpsertMatch) error {
	homeID, homeOK, err := s.teamIDByFifaID(ctx, q, m.HomeFifaID)
	if err != nil {
		return err
	}
	awayID, awayOK, err := s.teamIDByFifaID(ctx, q, m.AwayFifaID)
	if err != nil {
		return err
	}

	var homePtr, awayPtr *int64
	if homeOK {
		homePtr = &homeID
	}
	if awayOK {
		awayPtr = &awayID
	}

	const sql = `
		INSERT INTO matches (
			fifa_id, stage, group_label, match_number,
			home_team_id, away_team_id, kickoff_at, status,
			home_score, away_score,
			venue_stadium, venue_city, venue_country,
			placeholder_home, placeholder_away,
			result_source, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10,
			$11, $12, $13,
			$14, $15,
			'fifa', now()
		)
		ON CONFLICT (fifa_id) DO UPDATE SET
			stage            = EXCLUDED.stage,
			group_label      = EXCLUDED.group_label,
			match_number     = EXCLUDED.match_number,
			home_team_id     = EXCLUDED.home_team_id,
			away_team_id     = EXCLUDED.away_team_id,
			kickoff_at       = EXCLUDED.kickoff_at,
			status           = EXCLUDED.status,
			home_score       = EXCLUDED.home_score,
			away_score       = EXCLUDED.away_score,
			venue_stadium    = EXCLUDED.venue_stadium,
			venue_city       = EXCLUDED.venue_city,
			venue_country    = EXCLUDED.venue_country,
			placeholder_home = EXCLUDED.placeholder_home,
			placeholder_away = EXCLUDED.placeholder_away,
			result_source    = 'fifa',
			updated_at       = now()
		WHERE matches.result_source <> 'manual'`
	_, err = q.Exec(ctx, sql,
		m.FifaID, m.Stage, m.GroupLabel, m.MatchNumber,
		homePtr, awayPtr, m.KickoffAt, m.Status,
		m.HomeScore, m.AwayScore,
		m.VenueStadium, m.VenueCity, m.VenueCountry,
		m.PlaceholderHome, m.PlaceholderAway,
	)
	if err != nil {
		return fmt.Errorf("upsert match %s: %w", m.FifaID, err)
	}
	return nil
}

// WithTx runs fn inside a transaction, committing on success and rolling back
// on error.
func (s *Store) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ListMatches returns all matches ordered by kickoff_at (nulls last), with
// home/away team data joined when assigned.
func (s *Store) ListMatches(ctx context.Context) ([]Match, error) {
	const sql = `
		SELECT
			m.id, m.fifa_id, m.stage, COALESCE(m.group_label, ''), m.match_number,
			m.kickoff_at, m.status, m.home_score, m.away_score,
			COALESCE(m.venue_stadium, ''), COALESCE(m.venue_city, ''), COALESCE(m.venue_country, ''),
			COALESCE(m.placeholder_home, ''), COALESCE(m.placeholder_away, ''),
			m.result_source, m.updated_at,
			ht.id, COALESCE(ht.name, ''), COALESCE(ht.code, ''), COALESCE(ht.flag_url, ''),
			at.id, COALESCE(at.name, ''), COALESCE(at.code, ''), COALESCE(at.flag_url, '')
		FROM matches m
		LEFT JOIN teams ht ON ht.id = m.home_team_id
		LEFT JOIN teams at ON at.id = m.away_team_id
		ORDER BY m.kickoff_at ASC NULLS LAST, m.match_number ASC NULLS LAST, m.id ASC`

	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query matches: %w", err)
	}
	defer rows.Close()

	var out []Match
	for rows.Next() {
		var (
			m                   Match
			homeID, awayID      *int64
			hName, hCode, hFlag string
			aName, aCode, aFlag string
		)
		if err := rows.Scan(
			&m.ID, &m.FifaID, &m.Stage, &m.GroupLabel, &m.MatchNumber,
			&m.KickoffAt, &m.Status, &m.HomeScore, &m.AwayScore,
			&m.VenueStadium, &m.VenueCity, &m.VenueCountry,
			&m.PlaceholderHome, &m.PlaceholderAway,
			&m.ResultSource, &m.UpdatedAt,
			&homeID, &hName, &hCode, &hFlag,
			&awayID, &aName, &aCode, &aFlag,
		); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		if homeID != nil {
			m.Home = &Team{ID: *homeID, Name: hName, Code: hCode, FlagURL: hFlag, GroupLabel: m.GroupLabel}
		}
		if awayID != nil {
			m.Away = &Team{ID: *awayID, Name: aName, Code: aCode, FlagURL: aFlag, GroupLabel: m.GroupLabel}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ListFinishedGroupMatches returns finished group-stage matches that have both
// teams assigned and both scores recorded. Used to compute group standings.
func (s *Store) ListFinishedGroupMatches(ctx context.Context) ([]Match, error) {
	const sql = `
		SELECT m.id, m.home_team_id, m.away_team_id, m.home_score, m.away_score
		FROM matches m
		WHERE m.status = 'finished'
		  AND m.stage = 'group'
		  AND m.home_team_id IS NOT NULL
		  AND m.away_team_id IS NOT NULL
		  AND m.home_score IS NOT NULL
		  AND m.away_score IS NOT NULL
		ORDER BY m.id ASC`

	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query finished group matches: %w", err)
	}
	defer rows.Close()

	var out []Match
	for rows.Next() {
		var (
			m              Match
			homeID, awayID *int64
		)
		if err := rows.Scan(&m.ID, &homeID, &awayID, &m.HomeScore, &m.AwayScore); err != nil {
			return nil, fmt.Errorf("scan finished group match: %w", err)
		}
		if homeID != nil {
			m.Home = &Team{ID: *homeID}
		}
		if awayID != nil {
			m.Away = &Team{ID: *awayID}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ListTeams returns all teams ordered by group then name.
func (s *Store) ListTeams(ctx context.Context) ([]Team, error) {
	const sql = `
		SELECT id, COALESCE(fifa_id, ''), name, COALESCE(code, ''),
		       COALESCE(flag_url, ''), COALESCE(group_label, '')
		FROM teams
		ORDER BY group_label ASC NULLS LAST, name ASC`

	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query teams: %w", err)
	}
	defer rows.Close()

	var out []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.FifaID, &t.Name, &t.Code, &t.FlagURL, &t.GroupLabel); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
