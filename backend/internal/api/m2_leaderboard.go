package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// LeaderboardStore is the storage capability the leaderboard endpoint depends on.
type LeaderboardStore interface {
	Leaderboard(ctx context.Context) ([]storage.LeaderboardRow, error)
}

type leaderboardRowDTO struct {
	UserID     int64   `json:"userId"`
	Nickname   string  `json:"nickname"`
	AvatarURL  *string `json:"avatarUrl"`
	Points     int     `json:"points"`
	ExactCount int     `json:"exactCount"`
	Played     int     `json:"played"`
}

// RegisterLeaderboardRoutes wires GET /api/leaderboard (public).
func RegisterLeaderboardRoutes(r gin.IRouter, store LeaderboardStore) {
	r.GET("/api/leaderboard", leaderboardHandler(store))
}

func leaderboardHandler(store LeaderboardStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := store.Leaderboard(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load leaderboard"})
			return
		}
		dtos := make([]leaderboardRowDTO, 0, len(rows))
		for _, r := range rows {
			dtos = append(dtos, leaderboardRowDTO{
				UserID:     r.UserID,
				Nickname:   r.Nickname,
				AvatarURL:  r.AvatarURL,
				Points:     r.Points,
				ExactCount: r.ExactCount,
				Played:     r.Played,
			})
		}
		c.JSON(http.StatusOK, dtos)
	}
}
