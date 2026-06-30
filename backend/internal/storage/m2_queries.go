package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a single-row lookup finds nothing.
var ErrNotFound = errors.New("not found")

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

const userCols = `id, google_sub, email, nickname, avatar_url, favorite_team_code,
	telegram_chat_id, role, approved, access_level, created_at`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.GoogleSub, &u.Email, &u.Nickname, &u.AvatarURL,
		&u.FavoriteTeamCode, &u.TelegramChatID, &u.Role, &u.Approved, &u.AccessLevel, &u.CreatedAt)
	return u, err
}

// CountUsers returns the total number of users (used to assign the first user
// the admin role).
func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return n, nil
}

// GetUserByID looks up a user by id. Returns ErrNotFound when absent.
func (s *Store) GetUserByID(ctx context.Context, id int64) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// UserExists reports whether a user with the given id exists. Used to validate
// an admin's on-behalf-of target before writing a prediction.
func (s *Store) UserExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("user exists: %w", err)
	}
	return exists, nil
}

// ListUsers returns all users ordered by nickname for the admin picker.
func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+userCols+` FROM users ORDER BY nickname ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// GetUserByNickname looks up a user by nickname. Returns ErrNotFound when absent.
func (s *Store) GetUserByNickname(ctx context.Context, nickname string) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE nickname = $1`, nickname))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by nickname: %w", err)
	}
	return u, nil
}

// GetUserByGoogleSub looks up a user by google_sub. Returns ErrNotFound when absent.
func (s *Store) GetUserByGoogleSub(ctx context.Context, sub string) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE google_sub = $1`, sub))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by google_sub: %w", err)
	}
	return u, nil
}

// CreateUser inserts a new user and returns the created row. The role is
// decided by the caller (first user = admin). When the caller leaves
// AccessLevel empty, the initial level is derived from the current demo mode:
// 'none' while demo mode is ON (so self-service Google sign-ups land in the
// browse-only tier until an admin grants access), otherwise 'rw'. An explicit
// AccessLevel (e.g. admin-provisioned players) is honoured as-is.
func (s *Store) CreateUser(ctx context.Context, u User) (User, error) {
	var access *string
	if u.AccessLevel != "" {
		access = &u.AccessLevel
	}
	const sql = `
		INSERT INTO users (google_sub, email, nickname, avatar_url, role, access_level)
		VALUES ($1, $2, $3, $4, $5,
			COALESCE($6, CASE WHEN (SELECT value FROM app_state WHERE key = 'demo_mode') = 'true'
				THEN 'none' ELSE 'rw' END))
		RETURNING ` + userCols
	out, err := scanUser(s.pool.QueryRow(ctx, sql,
		u.GoogleSub, u.Email, u.Nickname, u.AvatarURL, defaultRole(u.Role), access))
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return out, nil
}

func defaultRole(r string) string {
	if r == "" {
		return "player"
	}
	return r
}

// CreatePlayer inserts a new role='player' account by nickname and returns the
// created row. It is the admin-provisioned roster entry point (no google_sub,
// no email). A duplicate nickname surfaces as a Postgres unique violation
// (SQLSTATE 23505) for the handler to map to 409.
func (s *Store) CreatePlayer(ctx context.Context, nickname string) (User, error) {
	// Admin-provisioned players are deliberate roster entries: grant full
	// access regardless of demo mode.
	return s.CreateUser(ctx, User{Nickname: nickname, Role: "player", AccessLevel: "rw"})
}

// GetUserAccess returns a user's access_level. Used by the demo-mode gate on
// each guarded request (lightweight single-column read).
func (s *Store) GetUserAccess(ctx context.Context, id int64) (string, error) {
	var lvl string
	err := s.pool.QueryRow(ctx, `SELECT access_level FROM users WHERE id = $1`, id).Scan(&lvl)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get user access: %w", err)
	}
	return lvl, nil
}

// SetUserAccess updates a user's access_level. Returns ErrNotFound when no row
// matched. The caller must validate level ∈ {none, ro, rw}.
func (s *Store) SetUserAccess(ctx context.Context, id int64, level string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE users SET access_level = $2 WHERE id = $1`, id, level)
	if err != nil {
		return fmt.Errorf("set user access: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// IsDemoMode reports whether demo mode is enabled (app_state key 'demo_mode').
func (s *Store) IsDemoMode(ctx context.Context) (bool, error) {
	v, ok, err := s.GetAppState(ctx, "demo_mode")
	if err != nil {
		return false, err
	}
	return ok && v == "true", nil
}

// SetDemoMode enables or disables demo mode.
func (s *Store) SetDemoMode(ctx context.Context, on bool) error {
	v := "false"
	if on {
		v = "true"
	}
	return s.SetAppState(ctx, "demo_mode", v)
}

// DeleteUserCascade deletes a user and all of their derived rows (predictions,
// materialized points, and tournament picks) in a single transaction. It
// returns ErrNotFound when no user with the id exists. Callers must ensure the
// target is not an admin before invoking.
func (s *Store) DeleteUserCascade(ctx context.Context, id int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("delete user cascade: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, q := range []string{
		`DELETE FROM predictions WHERE user_id = $1`,
		`DELETE FROM points WHERE user_id = $1`,
		`DELETE FROM tournament_picks WHERE user_id = $1`,
	} {
		if _, err := tx.Exec(ctx, q, id); err != nil {
			return fmt.Errorf("delete user cascade: %w", err)
		}
	}

	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user cascade: users: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("delete user cascade: commit: %w", err)
	}
	return nil
}

// UpdateUserProfile patches nickname/favoriteTeamCode/avatarUrl when the
// corresponding pointer is non-nil. Returns the updated row.
func (s *Store) UpdateUserProfile(ctx context.Context, id int64, nickname, favoriteTeamCode, avatarURL *string) (User, error) {
	const sql = `
		UPDATE users SET
			nickname           = COALESCE($2, nickname),
			favorite_team_code = COALESCE($3, favorite_team_code),
			avatar_url         = COALESCE($4, avatar_url)
		WHERE id = $1
		RETURNING ` + userCols
	out, err := scanUser(s.pool.QueryRow(ctx, sql, id, nickname, favoriteTeamCode, avatarURL))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("update user profile: %w", err)
	}
	return out, nil
}

// TeamCodeExists reports whether a team with the given code exists. Used to
// validate the profile favorite-team allowlist.
func (s *Store) TeamCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM teams WHERE code = $1)`, code).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("team code exists: %w", err)
	}
	return exists, nil
}

// ---------------------------------------------------------------------------
// Matches (M2 helpers)
// ---------------------------------------------------------------------------

// MatchScoringRow is the minimal match data the scoring/lock logic needs.
type MatchScoringRow struct {
	ID           int64
	Stage        string
	KickoffAt    *time.Time
	Status       string
	HomeTeamID   *int64
	AwayTeamID   *int64
	HomeScore    *int
	AwayScore    *int
	WinnerTeamID *int64 // actual advancer for knockouts (incl. ET/pens), else nil
}

// GetMatchForScoring loads the scoring-relevant fields of a single match.
func (s *Store) GetMatchForScoring(ctx context.Context, id int64) (MatchScoringRow, error) {
	const sql = `
		SELECT id, stage, kickoff_at, status, home_team_id, away_team_id, home_score, away_score, winner_team_id
		FROM matches WHERE id = $1`
	var m MatchScoringRow
	err := s.pool.QueryRow(ctx, sql, id).Scan(
		&m.ID, &m.Stage, &m.KickoffAt, &m.Status, &m.HomeTeamID, &m.AwayTeamID, &m.HomeScore, &m.AwayScore, &m.WinnerTeamID)
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchScoringRow{}, ErrNotFound
	}
	if err != nil {
		return MatchScoringRow{}, fmt.Errorf("get match for scoring: %w", err)
	}
	return m, nil
}

// GroupStageLastKickoff returns the latest kickoff among group-stage matches,
// used as the champion-pick tier window boundary. Returns (nil,nil) when none.
func (s *Store) GroupStageLastKickoff(ctx context.Context) (*time.Time, error) {
	var t *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT max(kickoff_at) FROM matches WHERE stage = 'group'`).Scan(&t)
	if err != nil {
		return nil, fmt.Errorf("group stage last kickoff: %w", err)
	}
	return t, nil
}

// FirstKnockoutKickoff returns the earliest kickoff among knockout matches
// (R32 / "round of 32"), used as the bonus tier boundary: a pick made before it
// (group stage still running) earns the higher tier, after it the lower tier.
// Returns (nil,nil) when none.
func (s *Store) FirstKnockoutKickoff(ctx context.Context) (*time.Time, error) {
	var t *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT min(kickoff_at) FROM matches WHERE stage <> 'group'`).Scan(&t)
	if err != nil {
		return nil, fmt.Errorf("first knockout kickoff: %w", err)
	}
	return t, nil
}

// FirstRoundOf16Kickoff returns the earliest Round-of-16 ("1/8") kickoff — the
// HARD lock for all tournament bonus picks (champion/finalist/top scorer): once
// the R16 starts no pick may be set or changed. Returns (nil,nil) when none.
func (s *Store) FirstRoundOf16Kickoff(ctx context.Context) (*time.Time, error) {
	var t *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT min(kickoff_at) FROM matches WHERE stage = 'r16'`).Scan(&t)
	if err != nil {
		return nil, fmt.Errorf("first round-of-16 kickoff: %w", err)
	}
	return t, nil
}

// ---------------------------------------------------------------------------
// Predictions
// ---------------------------------------------------------------------------

// UpsertPrediction inserts or updates a prediction keyed by (user_id, match_id)
// and returns the stored row.
func (s *Store) UpsertPrediction(ctx context.Context, p Prediction) (Prediction, error) {
	const sql = `
		INSERT INTO predictions (user_id, match_id, home_pred, away_pred, winner_pick_team_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (user_id, match_id) DO UPDATE SET
			home_pred           = EXCLUDED.home_pred,
			away_pred           = EXCLUDED.away_pred,
			winner_pick_team_id = EXCLUDED.winner_pick_team_id,
			updated_at          = now()
		RETURNING id, user_id, match_id, home_pred, away_pred, winner_pick_team_id, created_at, updated_at`
	var out Prediction
	err := s.pool.QueryRow(ctx, sql, p.UserID, p.MatchID, p.HomePred, p.AwayPred, p.WinnerPickTeamID).Scan(
		&out.ID, &out.UserID, &out.MatchID, &out.HomePred, &out.AwayPred, &out.WinnerPickTeamID,
		&out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return Prediction{}, fmt.Errorf("upsert prediction: %w", err)
	}
	return out, nil
}

// ListPredictionsByUser returns all predictions for one user.
func (s *Store) ListPredictionsByUser(ctx context.Context, userID int64) ([]Prediction, error) {
	const sql = `
		SELECT id, user_id, match_id, home_pred, away_pred, winner_pick_team_id, created_at, updated_at
		FROM predictions WHERE user_id = $1 ORDER BY match_id`
	rows, err := s.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("list predictions by user: %w", err)
	}
	defer rows.Close()
	var out []Prediction
	for rows.Next() {
		var p Prediction
		if err := rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.HomePred, &p.AwayPred,
			&p.WinnerPickTeamID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan prediction: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListPredictionsByMatch returns all predictions for one match joined with the
// author's nickname/avatar and their materialized points for that match.
func (s *Store) ListPredictionsByMatch(ctx context.Context, matchID int64) ([]MatchPrediction, error) {
	const sql = `
		SELECT p.user_id, u.nickname, u.avatar_url, p.home_pred, p.away_pred,
		       p.winner_pick_team_id, COALESCE(pt.points, 0)
		FROM predictions p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN points pt ON pt.user_id = p.user_id AND pt.match_id = p.match_id
		WHERE p.match_id = $1
		ORDER BY COALESCE(pt.points, 0) DESC, u.nickname ASC`
	rows, err := s.pool.Query(ctx, sql, matchID)
	if err != nil {
		return nil, fmt.Errorf("list predictions by match: %w", err)
	}
	defer rows.Close()
	var out []MatchPrediction
	for rows.Next() {
		var mp MatchPrediction
		if err := rows.Scan(&mp.UserID, &mp.Nickname, &mp.AvatarURL, &mp.HomePred,
			&mp.AwayPred, &mp.WinnerPickTeamID, &mp.Points); err != nil {
			return nil, fmt.Errorf("scan match prediction: %w", err)
		}
		out = append(out, mp)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Points (materialized)
// ---------------------------------------------------------------------------

// UpsertPoints materializes a per-match points row keyed by (user_id, match_id).
func (s *Store) UpsertPoints(ctx context.Context, p PointRow) error {
	const sql = `
		INSERT INTO points (user_id, match_id, points, breakdown_json)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, match_id) DO UPDATE SET
			points         = EXCLUDED.points,
			breakdown_json = EXCLUDED.breakdown_json`
	if _, err := s.pool.Exec(ctx, sql, p.UserID, p.MatchID, p.Points, p.BreakdownJSON); err != nil {
		return fmt.Errorf("upsert points: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Bonus rules / champion tiers
// ---------------------------------------------------------------------------

// GetBonusRule fetches a bonus rule by kind. Returns (zero,false,nil) when absent.
func (s *Store) GetBonusRule(ctx context.Context, kind string) (BonusRule, bool, error) {
	var b BonusRule
	err := s.pool.QueryRow(ctx,
		`SELECT kind, enabled, pts FROM bonus_rules WHERE kind = $1`, kind).Scan(&b.Kind, &b.Enabled, &b.Pts)
	if errors.Is(err, pgx.ErrNoRows) {
		return BonusRule{}, false, nil
	}
	if err != nil {
		return BonusRule{}, false, fmt.Errorf("get bonus rule: %w", err)
	}
	return b, true, nil
}

// ListChampionTiers returns the configured champion-pick tiers ordered by
// window_end ascending (nulls last).
func (s *Store) ListChampionTiers(ctx context.Context) ([]ChampionTier, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT window_end, pts FROM champion_tiers ORDER BY window_end ASC NULLS LAST`)
	if err != nil {
		return nil, fmt.Errorf("list champion tiers: %w", err)
	}
	defer rows.Close()
	var out []ChampionTier
	for rows.Next() {
		var t ChampionTier
		if err := rows.Scan(&t.WindowEnd, &t.Pts); err != nil {
			return nil, fmt.Errorf("scan champion tier: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Tournament picks
// ---------------------------------------------------------------------------

// UpsertTournamentPick inserts or updates a tournament pick keyed by
// (user_id, kind) and returns the stored row.
func (s *Store) UpsertTournamentPick(ctx context.Context, p TournamentPick) (TournamentPick, error) {
	const sql = `
		INSERT INTO tournament_picks (user_id, kind, pick_ref, locked_at, tier_points, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (user_id, kind) DO UPDATE SET
			pick_ref    = EXCLUDED.pick_ref,
			locked_at   = EXCLUDED.locked_at,
			tier_points = EXCLUDED.tier_points,
			updated_at  = now()
		RETURNING id, user_id, kind, pick_ref, locked_at, tier_points, created_at, updated_at`
	var out TournamentPick
	err := s.pool.QueryRow(ctx, sql, p.UserID, p.Kind, p.PickRef, p.LockedAt, p.TierPoints).Scan(
		&out.ID, &out.UserID, &out.Kind, &out.PickRef, &out.LockedAt, &out.TierPoints,
		&out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return TournamentPick{}, fmt.Errorf("upsert tournament pick: %w", err)
	}
	return out, nil
}

// ListTournamentPicksByUser returns all tournament picks for one user.
func (s *Store) ListTournamentPicksByUser(ctx context.Context, userID int64) ([]TournamentPick, error) {
	const sql = `
		SELECT id, user_id, kind, pick_ref, locked_at, tier_points, awarded, created_at, updated_at
		FROM tournament_picks WHERE user_id = $1 ORDER BY kind`
	rows, err := s.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("list tournament picks: %w", err)
	}
	defer rows.Close()
	var out []TournamentPick
	for rows.Next() {
		var p TournamentPick
		if err := rows.Scan(&p.ID, &p.UserID, &p.Kind, &p.PickRef, &p.LockedAt,
			&p.TierPoints, &p.Awarded, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tournament pick: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListUserHistory returns a player's per-match results history: their prediction,
// the actual result, and points earned, ordered by kickoff. Includes upcoming
// matches they've predicted (scored=false, points pending).
func (s *Store) ListUserHistory(ctx context.Context, userID int64) ([]UserHistoryRow, error) {
	const sql = `
		SELECT m.id, m.stage, COALESCE(m.group_label, ''), m.kickoff_at, m.status,
		       COALESCE(ht.code,''), COALESCE(ht.name,''), COALESCE(ht.flag_url,''),
		       COALESCE(at.code,''), COALESCE(at.name,''), COALESCE(at.flag_url,''),
		       m.home_score, m.away_score,
		       p.home_pred, p.away_pred, p.winner_pick_team_id,
		       COALESCE(pt.points, 0),
		       COALESCE((pt.breakdown_json->>'exact')::boolean, false),
		       (pt.user_id IS NOT NULL) AS scored
		FROM predictions p
		JOIN matches m ON m.id = p.match_id
		LEFT JOIN teams ht ON ht.id = m.home_team_id
		LEFT JOIN teams at ON at.id = m.away_team_id
		LEFT JOIN points pt ON pt.user_id = p.user_id AND pt.match_id = p.match_id
		WHERE p.user_id = $1
		ORDER BY m.kickoff_at ASC NULLS LAST, m.id ASC`
	rows, err := s.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("list user history: %w", err)
	}
	defer rows.Close()
	var out []UserHistoryRow
	for rows.Next() {
		var r UserHistoryRow
		if err := rows.Scan(&r.MatchID, &r.Stage, &r.GroupLabel, &r.KickoffAt, &r.Status,
			&r.HomeCode, &r.HomeName, &r.HomeFlag, &r.AwayCode, &r.AwayName, &r.AwayFlag,
			&r.HomeScore, &r.AwayScore, &r.PredHome, &r.PredAway, &r.WinnerPickTeamID,
			&r.Points, &r.Exact, &r.Scored); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Audit log
// ---------------------------------------------------------------------------

// AppendAudit inserts an immutable audit row. Actions only — never score values.
func (s *Store) AppendAudit(ctx context.Context, e AuditEntry) error {
	const sql = `
		INSERT INTO audit_log (actor_user_id, actor_role, action, match_id, target_user_id)
		VALUES ($1, $2, $3, $4, $5)`
	if _, err := s.pool.Exec(ctx, sql, e.ActorUserID, e.ActorRole, e.Action, e.MatchID, e.TargetUserID); err != nil {
		return fmt.Errorf("append audit: %w", err)
	}
	return nil
}

// ListAudit returns the newest-first public audit feed (no score values).
func (s *Store) ListAudit(ctx context.Context, limit int) ([]AuditEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	const sql = `
		SELECT a.id, a.actor_user_id, COALESCE(a.actor_role, ''), a.action, a.match_id,
		       a.target_user_id, a.created_at, COALESCE(u.nickname, '')
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.actor_user_id
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $1`
	rows, err := s.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.ActorUserID, &e.ActorRole, &e.Action, &e.MatchID,
			&e.TargetUserID, &e.CreatedAt, &e.ActorNickname); err != nil {
			return nil, fmt.Errorf("scan audit: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Leaderboard
// ---------------------------------------------------------------------------

// Leaderboard aggregates per-match points plus any awarded tournament-pick
// tier points, sorted by total points desc then exact-count desc.
func (s *Store) Leaderboard(ctx context.Context) ([]LeaderboardRow, error) {
	const sql = `
		SELECT
			u.id,
			u.nickname,
			u.avatar_url,
			COALESCE(mp.match_points, 0) AS match_points,
			COALESCE(tp.bonus_points, 0) AS bonus_points,
			COALESCE(mp.match_points, 0) + COALESCE(tp.bonus_points, 0) AS total_points,
			COALESCE(mp.exact_count, 0) AS exact_count,
			COALESCE(mp.played, 0)      AS played
		FROM users u
		LEFT JOIN (
			SELECT user_id,
			       sum(points) AS match_points,
			       count(*) FILTER (WHERE (breakdown_json->>'exact')::boolean) AS exact_count,
			       count(*) AS played
			FROM points
			GROUP BY user_id
		) mp ON mp.user_id = u.id
		-- The organizer (admin) is not a competitor — keep them off the board.
		LEFT JOIN (
			-- Bonus points count only once a pick is resolved correct (awarded=true);
			-- a pending or wrong pick contributes 0. tier_points is the potential award.
			SELECT user_id, sum(tier_points) AS bonus_points
			FROM tournament_picks
			WHERE tier_points IS NOT NULL AND awarded IS TRUE
			GROUP BY user_id
		) tp ON tp.user_id = u.id
		WHERE u.role <> 'admin'
		ORDER BY total_points DESC, exact_count DESC, u.nickname ASC`
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("leaderboard: %w", err)
	}
	defer rows.Close()
	var out []LeaderboardRow
	for rows.Next() {
		var r LeaderboardRow
		if err := rows.Scan(&r.UserID, &r.Nickname, &r.AvatarURL, &r.MatchPoints, &r.BonusPoints, &r.Points, &r.ExactCount, &r.Played); err != nil {
			return nil, fmt.Errorf("scan leaderboard: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListPredictionsForMatchRaw returns the raw predictions for a match for
// recompute (no joins).
func (s *Store) ListPredictionsForMatchRaw(ctx context.Context, matchID int64) ([]Prediction, error) {
	const sql = `
		SELECT id, user_id, match_id, home_pred, away_pred, winner_pick_team_id, created_at, updated_at
		FROM predictions WHERE match_id = $1`
	rows, err := s.pool.Query(ctx, sql, matchID)
	if err != nil {
		return nil, fmt.Errorf("list predictions for match: %w", err)
	}
	defer rows.Close()
	var out []Prediction
	for rows.Next() {
		var p Prediction
		if err := rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.HomePred, &p.AwayPred,
			&p.WinnerPickTeamID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan prediction: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListMatchIDsWithResults returns ids of matches that have both scores set
// (used by recompute-scores to iterate scoreable matches).
func (s *Store) ListMatchIDsWithResults(ctx context.Context) ([]int64, error) {
	const sql = `SELECT id FROM matches WHERE home_score IS NOT NULL AND away_score IS NOT NULL`
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("list match ids with results: %w", err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan match id: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
