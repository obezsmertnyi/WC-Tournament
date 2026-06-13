package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// AdminStore is the storage capability the admin endpoints depend on.
type AdminStore interface {
	ListUsers(ctx context.Context) ([]storage.User, error)
}

// adminUserDTO is the user picker shape returned to the admin UI.
type adminUserDTO struct {
	ID               int64   `json:"id"`
	Nickname         string  `json:"nickname"`
	AvatarURL        *string `json:"avatarUrl"`
	FavoriteTeamCode *string `json:"favoriteTeamCode"`
	Role             string  `json:"role"`
}

// RegisterAdminRoutes wires GET /api/admin/users (RequireAdmin).
func RegisterAdminRoutes(r gin.IRouter, store AdminStore) {
	r.GET("/api/admin/users", auth.RequireAdmin(), adminListUsersHandler(store))
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
