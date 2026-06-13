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
	AppendAudit(ctx context.Context, e storage.AuditEntry) error
}

// adminUserDTO is the user picker shape returned to the admin UI.
type adminUserDTO struct {
	ID               int64   `json:"id"`
	Nickname         string  `json:"nickname"`
	AvatarURL        *string `json:"avatarUrl"`
	FavoriteTeamCode *string `json:"favoriteTeamCode"`
	Role             string  `json:"role"`
}

// RegisterAdminRoutes wires the admin user-roster endpoints (all RequireAdmin):
//
//	GET    /api/admin/users        list users
//	POST   /api/admin/users        provision a player by nickname
//	DELETE /api/admin/users/:id    delete a player and their derived rows
func RegisterAdminRoutes(r gin.IRouter, store AdminStore) {
	grp := r.Group("/api/admin/users", auth.RequireAdmin())
	grp.GET("", adminListUsersHandler(store))
	grp.POST("", adminCreateUserHandler(store))
	grp.DELETE("/:id", adminDeleteUserHandler(store))
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
