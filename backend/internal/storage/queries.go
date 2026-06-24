package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

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
			fifa_id, fifa_stage_id, stage, group_label, match_number,
			home_team_id, away_team_id, kickoff_at, status,
			home_score, away_score,
			venue_stadium, venue_city, venue_country,
			placeholder_home, placeholder_away,
			result_source, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13, $14,
			$15, $16,
			'fifa', now()
		)
		ON CONFLICT (fifa_id) DO UPDATE SET
			fifa_stage_id    = EXCLUDED.fifa_stage_id,
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
		m.FifaID, m.FifaStageID, m.Stage, m.GroupLabel, m.MatchNumber,
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

// GetMatchByID returns the identifying fields of a single match needed to
// build the FIFA live URL. Returns ErrNotFound when no match with the id
// exists. Only the columns the match-detail endpoint consumes are populated.
func (s *Store) GetMatchByID(ctx context.Context, id int64) (Match, error) {
	const sql = `
		SELECT id, COALESCE(fifa_id, ''), COALESCE(fifa_stage_id, ''), stage, status
		FROM matches
		WHERE id = $1`
	var m Match
	err := s.pool.QueryRow(ctx, sql, id).Scan(&m.ID, &m.FifaID, &m.FifaStageID, &m.Stage, &m.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Match{}, ErrNotFound
	}
	if err != nil {
		return Match{}, fmt.Errorf("get match %d: %w", id, err)
	}
	return m, nil
}

// SetMatchResult overrides a match result by id (admin manual entry). It sets
// home_score/away_score/status, marks result_source='manual' so a later FIFA
// sync will not silently overwrite it (ADR-0006), and bumps updated_at. The
// updated row is returned with home/away team data joined when assigned.
// Returns ErrNotFound when no match with the id exists.
func (s *Store) SetMatchResult(ctx context.Context, id int64, home, away int, status string) (Match, error) {
	const sql = `
		WITH upd AS (
			UPDATE matches SET
				home_score    = $2,
				away_score    = $3,
				status        = $4,
				result_source = 'manual',
				updated_at    = now()
			WHERE id = $1
			RETURNING id, fifa_id, stage, group_label, match_number,
			          kickoff_at, status, home_score, away_score,
			          venue_stadium, venue_city, venue_country,
			          placeholder_home, placeholder_away,
			          result_source, updated_at, home_team_id, away_team_id
		)
		SELECT
			upd.id, upd.fifa_id, upd.stage, COALESCE(upd.group_label, ''), upd.match_number,
			upd.kickoff_at, upd.status, upd.home_score, upd.away_score,
			COALESCE(upd.venue_stadium, ''), COALESCE(upd.venue_city, ''), COALESCE(upd.venue_country, ''),
			COALESCE(upd.placeholder_home, ''), COALESCE(upd.placeholder_away, ''),
			upd.result_source, upd.updated_at,
			ht.id, COALESCE(ht.name, ''), COALESCE(ht.code, ''), COALESCE(ht.flag_url, ''),
			at.id, COALESCE(at.name, ''), COALESCE(at.code, ''), COALESCE(at.flag_url, '')
		FROM upd
		LEFT JOIN teams ht ON ht.id = upd.home_team_id
		LEFT JOIN teams at ON at.id = upd.away_team_id`

	var (
		m                   Match
		homeID, awayID      *int64
		hName, hCode, hFlag string
		aName, aCode, aFlag string
	)
	err := s.pool.QueryRow(ctx, sql, id, home, away, status).Scan(
		&m.ID, &m.FifaID, &m.Stage, &m.GroupLabel, &m.MatchNumber,
		&m.KickoffAt, &m.Status, &m.HomeScore, &m.AwayScore,
		&m.VenueStadium, &m.VenueCity, &m.VenueCountry,
		&m.PlaceholderHome, &m.PlaceholderAway,
		&m.ResultSource, &m.UpdatedAt,
		&homeID, &hName, &hCode, &hFlag,
		&awayID, &aName, &aCode, &aFlag,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Match{}, ErrNotFound
	}
	if err != nil {
		return Match{}, fmt.Errorf("set match result: %w", err)
	}
	if homeID != nil {
		m.Home = &Team{ID: *homeID, Name: hName, Code: hCode, FlagURL: hFlag, GroupLabel: m.GroupLabel}
	}
	if awayID != nil {
		m.Away = &Team{ID: *awayID, Name: aName, Code: aCode, FlagURL: aFlag, GroupLabel: m.GroupLabel}
	}
	return m, nil
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

// ListUnannouncedFinishedMatches returns finished matches with both teams and
// both scores that have NOT yet been posted to Telegram (announced_at IS NULL),
// with team data joined. Used by the result announcer. Oldest first so messages
// arrive in chronological order.
func (s *Store) ListUnannouncedFinishedMatches(ctx context.Context) ([]Match, error) {
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
		WHERE m.status = 'finished'
		  AND m.announced_at IS NULL
		  AND m.home_team_id IS NOT NULL
		  AND m.away_team_id IS NOT NULL
		  AND m.home_score IS NOT NULL
		  AND m.away_score IS NOT NULL
		ORDER BY m.kickoff_at ASC NULLS LAST, m.match_number ASC NULLS LAST, m.id ASC`

	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query unannounced matches: %w", err)
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
			return nil, fmt.Errorf("scan unannounced match: %w", err)
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

// ListUpcomingUnremindedMatches returns scheduled matches with both teams known
// whose kickoff is within the next `window` (and still in the future), that have
// not yet had a pre-match reminder sent. Team data is joined. Oldest kickoff first.
func (s *Store) ListUpcomingUnremindedMatches(ctx context.Context, window time.Duration) ([]Match, error) {
	cutoff := time.Now().UTC().Add(window)
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
		WHERE m.status = 'scheduled'
		  AND m.reminded_at IS NULL
		  AND m.home_team_id IS NOT NULL
		  AND m.away_team_id IS NOT NULL
		  AND m.kickoff_at IS NOT NULL
		  AND m.kickoff_at > now()
		  AND m.kickoff_at <= $1
		ORDER BY m.kickoff_at ASC, m.id ASC`

	rows, err := s.pool.Query(ctx, sql, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query upcoming unreminded matches: %w", err)
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
			return nil, fmt.Errorf("scan upcoming match: %w", err)
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

// ListUsersMissingPrediction returns the nicknames of non-admin users who have
// NOT submitted a prediction for the given match. Used by the pre-match reminder.
func (s *Store) ListUsersMissingPrediction(ctx context.Context, matchID int64) ([]string, error) {
	const sql = `
		SELECT u.nickname
		FROM users u
		WHERE u.role <> 'admin'
		  AND NOT EXISTS (
		      SELECT 1 FROM predictions p WHERE p.user_id = u.id AND p.match_id = $1
		  )
		ORDER BY u.nickname ASC`
	rows, err := s.pool.Query(ctx, sql, matchID)
	if err != nil {
		return nil, fmt.Errorf("query users missing prediction: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("scan missing predictor: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// GetAppState reads a key from the app_state kv table. Returns ("",false,nil)
// when the key is absent.
func (s *Store) GetAppState(ctx context.Context, key string) (string, bool, error) {
	var v string
	err := s.pool.QueryRow(ctx, `SELECT value FROM app_state WHERE key = $1`, key).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get app state %q: %w", key, err)
	}
	return v, true, nil
}

// SetAppState upserts a key/value into the app_state kv table.
func (s *Store) SetAppState(ctx context.Context, key, value string) error {
	const sql = `
		INSERT INTO app_state (key, value, updated_at) VALUES ($1, $2, now())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`
	if _, err := s.pool.Exec(ctx, sql, key, value); err != nil {
		return fmt.Errorf("set app state %q: %w", key, err)
	}
	return nil
}

// MarkMatchReminded stamps reminded_at=now() so the pre-match reminder fires once.
func (s *Store) MarkMatchReminded(ctx context.Context, id int64) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE matches SET reminded_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("mark match reminded %d: %w", id, err)
	}
	return nil
}

// MarkMatchAnnounced stamps announced_at=now() so the result is not posted again.
func (s *Store) MarkMatchAnnounced(ctx context.Context, id int64) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE matches SET announced_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("mark match announced %d: %w", id, err)
	}
	return nil
}

// TeamIDByFifaID resolves a team's local id from its FIFA id. (false, nil) when absent.
func (s *Store) TeamIDByFifaID(ctx context.Context, fifaID string) (int64, bool, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM teams WHERE fifa_id = $1`, fifaID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("team by fifa id %q: %w", fifaID, err)
	}
	return id, true, nil
}

// FinishedFifaRef identifies a finished match for live-detail fetches.
type FinishedFifaRef struct {
	ID          int64
	FifaID      string
	FifaStageID string
}

// ListFinishedFifaRefs returns the FIFA id/stage of every finished match, for
// aggregating goal data (e.g. resolving the top scorer).
func (s *Store) ListFinishedFifaRefs(ctx context.Context) ([]FinishedFifaRef, error) {
	const sql = `
		SELECT id, COALESCE(fifa_id, ''), COALESCE(fifa_stage_id, '')
		FROM matches WHERE status = 'finished' AND fifa_id <> '' ORDER BY kickoff_at ASC NULLS LAST`
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("list finished fifa refs: %w", err)
	}
	defer rows.Close()
	var out []FinishedFifaRef
	for rows.Next() {
		var r FinishedFifaRef
		if err := rows.Scan(&r.ID, &r.FifaID, &r.FifaStageID); err != nil {
			return nil, fmt.Errorf("scan fifa ref: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// AwardBonusByKind marks the tournament picks of a kind awarded iff their
// pick_ref equals the resolved correct value (idempotent: re-running re-derives
// awarded for every pick of that kind). Used for champion/finalist.
func (s *Store) AwardBonusByKind(ctx context.Context, kind, correctPickRef string) error {
	const sql = `UPDATE tournament_picks SET awarded = (pick_ref = $2) WHERE kind = $1`
	if _, err := s.pool.Exec(ctx, sql, kind, correctPickRef); err != nil {
		return fmt.Errorf("award bonus %q: %w", kind, err)
	}
	return nil
}

// ListTournamentPicksByKind returns all picks of one kind across users (for the
// top-scorer name match, where awarding can't be done by an exact SQL equality).
func (s *Store) ListTournamentPicksByKind(ctx context.Context, kind string) ([]TournamentPick, error) {
	const sql = `
		SELECT id, user_id, kind, pick_ref, locked_at, tier_points, awarded, created_at, updated_at
		FROM tournament_picks WHERE kind = $1`
	rows, err := s.pool.Query(ctx, sql, kind)
	if err != nil {
		return nil, fmt.Errorf("list picks by kind %q: %w", kind, err)
	}
	defer rows.Close()
	var out []TournamentPick
	for rows.Next() {
		var p TournamentPick
		if err := rows.Scan(&p.ID, &p.UserID, &p.Kind, &p.PickRef, &p.LockedAt,
			&p.TierPoints, &p.Awarded, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pick: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SetPickAwarded sets the awarded flag on a single tournament pick.
func (s *Store) SetPickAwarded(ctx context.Context, id int64, awarded bool) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE tournament_picks SET awarded = $2 WHERE id = $1`, id, awarded); err != nil {
		return fmt.Errorf("set pick %d awarded: %w", id, err)
	}
	return nil
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
