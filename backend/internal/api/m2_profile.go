package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// ProfileStore is the storage capability the profile endpoints depend on.
type ProfileStore interface {
	GetUserByID(ctx context.Context, id int64) (storage.User, error)
	UpdateUserProfile(ctx context.Context, id int64, nickname, favoriteTeamCode, avatarURL *string) (storage.User, error)
}

type meDTO struct {
	ID               int64   `json:"id"`
	Nickname         string  `json:"nickname"`
	AvatarURL        *string `json:"avatarUrl"`
	FavoriteTeamCode *string `json:"favoriteTeamCode"`
	Role             string  `json:"role"`
}

func toMeDTO(u storage.User) meDTO {
	return meDTO{
		ID:               u.ID,
		Nickname:         u.Nickname,
		AvatarURL:        u.AvatarURL,
		FavoriteTeamCode: u.FavoriteTeamCode,
		Role:             u.Role,
	}
}

// RegisterProfileRoutes wires GET/PATCH /api/me (RequireUser).
func RegisterProfileRoutes(r gin.IRouter, store ProfileStore) {
	grp := r.Group("/api/me", auth.RequireUser())
	grp.GET("", meHandler(store))
	grp.PATCH("", patchMeHandler(store))
}

func meHandler(store ProfileStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		u, err := store.GetUserByID(c.Request.Context(), claims.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
			return
		}
		c.JSON(http.StatusOK, toMeDTO(u))
	}
}

func patchMeHandler(store ProfileStore) gin.HandlerFunc {
	type req struct {
		Nickname         *string `json:"nickname"`
		FavoriteTeamCode *string `json:"favoriteTeamCode"`
		AvatarURL        *string `json:"avatarUrl"`
	}
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if body.Nickname != nil {
			trimmed := strings.TrimSpace(*body.Nickname)
			if trimmed == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "nickname cannot be empty"})
				return
			}
			body.Nickname = &trimmed
		}

		u, err := store.UpdateUserProfile(c.Request.Context(), claims.Sub, body.Nickname, body.FavoriteTeamCode, body.AvatarURL)
		if err != nil {
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "nickname already taken"})
				return
			}
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
			return
		}
		c.JSON(http.StatusOK, toMeDTO(u))
	}
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
