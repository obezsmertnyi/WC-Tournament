package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// AuditStore is the storage capability the audit endpoint depends on.
type AuditStore interface {
	ListAudit(ctx context.Context, limit int) ([]storage.AuditEntry, error)
}

// auditDTO is the public audit feed shape. It NEVER carries score values.
type auditDTO struct {
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	MatchID   *int64 `json:"matchId"`
	CreatedAt string `json:"createdAt"`
}

// RegisterAuditRoutes wires GET /api/audit (public).
func RegisterAuditRoutes(r gin.IRouter, store AuditStore) {
	r.GET("/api/audit", auditHandler(store))
}

func auditHandler(store AuditStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 200
		if q := c.Query("limit"); q != "" {
			if n, err := strconv.Atoi(q); err == nil {
				limit = n
			}
		}
		rows, err := store.ListAudit(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load audit log"})
			return
		}
		dtos := make([]auditDTO, 0, len(rows))
		for _, e := range rows {
			actor := e.ActorNickname
			if actor == "" {
				actor = "system"
			}
			dtos = append(dtos, auditDTO{
				Actor:     actor,
				Action:    e.Action,
				MatchID:   e.MatchID,
				CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		c.JSON(http.StatusOK, dtos)
	}
}
