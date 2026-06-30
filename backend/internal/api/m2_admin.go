package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// AdminStore is the storage capability the admin endpoints depend on.
type AdminStore interface {
	ListUsers(ctx context.Context) ([]storage.User, error)
	CreatePlayer(ctx context.Context, nickname string) (storage.User, error)
	GetUserByID(ctx context.Context, id int64) (storage.User, error)
	DeleteUserCascade(ctx context.Context, id int64) error
	SetMatchResult(ctx context.Context, id int64, home, away int, status string) (storage.Match, error)
	SetUserAccess(ctx context.Context, id int64, level string) error
	IsDemoMode(ctx context.Context) (bool, error)
	SetDemoMode(ctx context.Context, on bool) error
	AppendAudit(ctx context.Context, e storage.AuditEntry) error
}

// adminUserDTO is the user picker shape returned to the admin UI.
type adminUserDTO struct {
	ID               int64   `json:"id"`
	Nickname         string  `json:"nickname"`
	AvatarURL        *string `json:"avatarUrl"`
	FavoriteTeamCode *string `json:"favoriteTeamCode"`
	Role             string  `json:"role"`
	AccessLevel      string  `json:"accessLevel"`
}

// RegisterAdminRoutes wires the admin endpoints (all RequireAdmin):
//
//	GET    /api/admin/users               list users
//	POST   /api/admin/users               provision a player by nickname
//	DELETE /api/admin/users/:id           delete a player and their derived rows
//	PUT    /api/admin/matches/:id/result  set/override a match result (manual)
//
// rc may be nil to disable the re-scoring hook (e.g. in tests that don't assert
// on recompute), though production always passes a live recomputer.
func RegisterAdminRoutes(r gin.IRouter, store AdminStore, rc Recomputer) {
	grp := r.Group("/api/admin/users", auth.RequireAdmin())
	grp.GET("", adminListUsersHandler(store))
	grp.POST("", adminCreateUserHandler(store))
	grp.DELETE("/:id", adminDeleteUserHandler(store))
	grp.PUT("/:id/access", adminSetAccessHandler(store))

	matches := r.Group("/api/admin/matches", auth.RequireAdmin())
	matches.PUT("/:id/result", adminSetMatchResultHandler(store, rc))

	demo := r.Group("/api/admin/demo", auth.RequireAdmin())
	demo.GET("", adminGetDemoHandler(store))
	demo.PUT("", adminSetDemoHandler(store))
}

// adminSetMatchResultHandler sets or overrides a match result by id. Per
// ADR-0006 a manual admin entry must win over a later FIFA sync, so the storage
// write marks result_source='manual'. After persisting, it recomputes points
// for the affected match so the leaderboard updates immediately, and writes a
// result_override audit row (match id only — never the score values, per policy).
func adminSetMatchResultHandler(store AdminStore, rc Recomputer) gin.HandlerFunc {
	type req struct {
		HomeScore int    `json:"homeScore"`
		AwayScore int    `json:"awayScore"`
		Status    string `json:"status"`
	}
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
			return
		}

		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if body.HomeScore < 0 || body.HomeScore > maxScore || body.AwayScore < 0 || body.AwayScore > maxScore {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scores must be between 0 and 30"})
			return
		}

		status := strings.TrimSpace(body.Status)
		if status == "" {
			status = "finished"
		}
		switch status {
		case "finished", "live", "scheduled":
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be finished, live or scheduled"})
			return
		}

		ctx := c.Request.Context()
		m, err := store.SetMatchResult(ctx, id, body.HomeScore, body.AwayScore, status)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set match result"})
			return
		}

		// Re-materialize points for this match so the leaderboard reflects the
		// override immediately.
		if rc != nil {
			_ = rc.RecomputeMatch(ctx, id)
		}

		// Audit (actions only — never the score values, per policy).
		claims, _ := auth.Current(c)
		actor := claims.Sub
		matchID := id
		_ = store.AppendAudit(ctx, storage.AuditEntry{
			ActorUserID: &actor,
			ActorRole:   claims.Role,
			Action:      "result_override",
			MatchID:     &matchID,
		})

		c.JSON(http.StatusOK, toMatchDTO(m))
	}
}

func adminListUsersHandler(store AdminStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := store.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
			return
		}
		dtos := make([]adminUserDTO, 0, len(rows))
		for _, u := range rows {
			dtos = append(dtos, adminUserDTO{
				ID:               u.ID,
				Nickname:         u.Nickname,
				AvatarURL:        u.AvatarURL,
				FavoriteTeamCode: u.FavoriteTeamCode,
				Role:             u.Role,
				AccessLevel:      u.AccessLevel,
			})
		}
		c.JSON(http.StatusOK, dtos)
	}
}

func toAdminUserDTO(u storage.User) adminUserDTO {
	return adminUserDTO{
		ID:               u.ID,
		Nickname:         u.Nickname,
		AvatarURL:        u.AvatarURL,
		FavoriteTeamCode: u.FavoriteTeamCode,
		Role:             u.Role,
		AccessLevel:      u.AccessLevel,
	}
}

// adminSetAccessHandler sets a player's demo access level (none/ro/rw). It
// refuses (403) to change an admin account (admins are always rw), returns 404
// for an unknown id, and writes an admin_set_access audit row on success.
func adminSetAccessHandler(store AdminStore) gin.HandlerFunc {
	type req struct {
		Level string `json:"level"`
	}
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		level := strings.TrimSpace(body.Level)
		switch level {
		case auth.AccessNone, auth.AccessRO, auth.AccessRW:
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "level must be none, ro or rw"})
			return
		}

		ctx := c.Request.Context()
		target, err := store.GetUserByID(ctx, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set access"})
			return
		}
		if target.Role == "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot change an admin's access"})
			return
		}

		if err := store.SetUserAccess(ctx, id, level); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set access"})
			return
		}

		claims, _ := auth.Current(c)
		actor := claims.Sub
		t := id
		_ = store.AppendAudit(ctx, storage.AuditEntry{
			ActorUserID:  &actor,
			ActorRole:    claims.Role,
			Action:       "admin_set_access",
			TargetUserID: &t,
		})

		target.AccessLevel = level
		c.JSON(http.StatusOK, toAdminUserDTO(target))
	}
}

// adminGetDemoHandler reports whether demo mode is enabled.
func adminGetDemoHandler(store AdminStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		on, err := store.IsDemoMode(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read demo mode"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"enabled": on})
	}
}

// adminSetDemoHandler enables or disables demo mode and writes an
// admin_demo_mode audit row.
func adminSetDemoHandler(store AdminStore) gin.HandlerFunc {
	type req struct {
		Enabled bool `json:"enabled"`
	}
	return func(c *gin.Context) {
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		ctx := c.Request.Context()
		if err := store.SetDemoMode(ctx, body.Enabled); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set demo mode"})
			return
		}
		claims, _ := auth.Current(c)
		actor := claims.Sub
		_ = store.AppendAudit(ctx, storage.AuditEntry{
			ActorUserID: &actor,
			ActorRole:   claims.Role,
			Action:      "admin_demo_mode",
		})
		c.JSON(http.StatusOK, gin.H{"enabled": body.Enabled})
	}
}

// adminCreateUserHandler provisions a role='player' account by nickname. The
// nickname is trimmed and validated against the same safe charset / length
// rules as the profile editor. A duplicate nickname returns 409. On success it
// writes an admin_create_user audit row (target_user_id = the new user id).
func adminCreateUserHandler(store AdminStore) gin.HandlerFunc {
	type req struct {
		Nickname string `json:"nickname"`
	}
	return func(c *gin.Context) {
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		nickname := strings.TrimSpace(body.Nickname)
		if msg := validateNickname(nickname); msg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		ctx := c.Request.Context()
		u, err := store.CreatePlayer(ctx, nickname)
		if err != nil {
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "nickname already taken"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		claims, _ := auth.Current(c)
		actor := claims.Sub
		target := u.ID
		_ = store.AppendAudit(ctx, storage.AuditEntry{
			ActorUserID:  &actor,
			ActorRole:    claims.Role,
			Action:       "admin_create_user",
			TargetUserID: &target,
		})

		c.JSON(http.StatusOK, toAdminUserDTO(u))
	}
}

// adminDeleteUserHandler deletes a player and all of their derived rows
// (predictions, points, tournament picks). It refuses (403) to delete an admin
// account, returns 404 for an unknown id, and writes an admin_delete_user audit
// row (target_user_id = the deleted user id) on success.
func adminDeleteUserHandler(store AdminStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		ctx := c.Request.Context()
		target, err := store.GetUserByID(ctx, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
			return
		}
		if target.Role == "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete an admin account"})
			return
		}

		if err := store.DeleteUserCascade(ctx, id); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
			return
		}

		claims, _ := auth.Current(c)
		actor := claims.Sub
		t := id
		_ = store.AppendAudit(ctx, storage.AuditEntry{
			ActorUserID:  &actor,
			ActorRole:    claims.Role,
			Action:       "admin_delete_user",
			TargetUserID: &t,
		})

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
