package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// TopScorersStore is the storage capability the top-scorers endpoint depends on.
type TopScorersStore interface {
	ListTopScorers(ctx context.Context, limit int) ([]storage.ScorerRow, error)
}

type topScorerDTO struct {
	Rank     int    `json:"rank"`
	Name     string `json:"name"`
	TeamCode string `json:"teamCode"`
	Goals    int    `json:"goals"`
}

// RegisterTopScorersRoutes wires GET /api/top-scorers?limit=N (default 10).
func RegisterTopScorersRoutes(r gin.IRouter, store TopScorersStore) {
	r.GET("/api/top-scorers", topScorersHandler(store))
}

func topScorersHandler(store TopScorersStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 10
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		rows, err := store.ListTopScorers(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load top scorers"})
			return
		}
		dtos := make([]topScorerDTO, 0, len(rows))
		for i, r := range rows {
			dtos = append(dtos, topScorerDTO{Rank: i + 1, Name: r.Name, TeamCode: r.TeamCode, Goals: r.Goals})
		}
		c.JSON(http.StatusOK, gin.H{"scorers": dtos})
	}
}
