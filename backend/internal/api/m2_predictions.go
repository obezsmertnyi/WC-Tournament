package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

const maxScore = 30

// PredictionStore is the storage capability the prediction endpoints depend on.
type PredictionStore interface {
	GetMatchForScoring(ctx context.Context, id int64) (storage.MatchScoringRow, error)
	UpsertPrediction(ctx context.Context, p storage.Prediction) (storage.Prediction, error)
	ListPredictionsByUser(ctx context.Context, userID int64) ([]storage.Prediction, error)
	ListPredictionsByMatch(ctx context.Context, matchID int64) ([]storage.MatchPrediction, error)
	UserExists(ctx context.Context, id int64) (bool, error)
	AppendAudit(ctx context.Context, e storage.AuditEntry) error
}

// Recomputer recomputes materialized points for a match after a write. The
// concrete *scoring.Recomputer satisfies this; nil disables the hook.
type Recomputer interface {
	RecomputeMatch(ctx context.Context, matchID int64) error
}

type predictionDTO struct {
	MatchID          int64  `json:"matchId"`
	Home             int    `json:"home"`
	Away             int    `json:"away"`
	WinnerPickTeamID *int64 `json:"winnerPickTeamId"`
}

// RegisterPredictionRoutes wires the prediction read/write endpoints.
func RegisterPredictionRoutes(r gin.IRouter, store PredictionStore, rc Recomputer) {
	r.GET("/api/predictions/me", auth.RequireUser(), myPredictionsHandler(store))
	r.PUT("/api/predictions/:matchId", auth.RequireUser(), upsertPredictionHandler(store, rc))
	// Reveal endpoint: only logged-in users may see revealed predictions.
	r.GET("/api/matches/:id/predictions", auth.RequireUser(), matchPredictionsHandler(store))
}

func myPredictionsHandler(store PredictionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		rows, err := store.ListPredictionsByUser(c.Request.Context(), claims.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load predictions"})
			return
		}
		dtos := make([]predictionDTO, 0, len(rows))
		for _, p := range rows {
			dtos = append(dtos, predictionDTO{
				MatchID:          p.MatchID,
				Home:             p.HomePred,
				Away:             p.AwayPred,
				WinnerPickTeamID: p.WinnerPickTeamID,
			})
		}
		c.JSON(http.StatusOK, gin.H{"predictions": dtos})
	}
}

func upsertPredictionHandler(store PredictionStore, rc Recomputer) gin.HandlerFunc {
	type req struct {
		Home             int    `json:"home"`
		Away             int    `json:"away"`
		WinnerPickTeamID *int64 `json:"winnerPickTeamId"`
		ForUserID        *int64 `json:"forUserId"`
	}
	return func(c *gin.Context) {
		claims, _ := auth.Current(c)
		ctx := c.Request.Context()

		matchID, err := strconv.ParseInt(c.Param("matchId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
			return
		}

		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if body.Home < 0 || body.Home > maxScore || body.Away < 0 || body.Away > maxScore {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scores must be between 0 and 30"})
			return
		}

		isAdmin := claims.Role == "admin"

		// Admin on-behalf-of: when forUserId is present the caller MUST be admin
		// and the prediction is written for that target user. The target must
		// exist. Non-admins are forbidden from setting forUserId.
		targetUserID := claims.Sub
		onBehalf := false
		if body.ForUserID != nil {
			if !isAdmin {
				c.JSON(http.StatusForbidden, gin.H{"error": "admin required to predict for another user"})
				return
			}
			exists, err := store.UserExists(ctx, *body.ForUserID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate target user"})
				return
			}
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
				return
			}
			targetUserID = *body.ForUserID
			onBehalf = true
		}

		match, err := store.GetMatchForScoring(ctx, matchID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load match"})
			return
		}

		// Winner pick: only for knockout matches and must be one of the two teams.
		if body.WinnerPickTeamID != nil {
			if match.Stage == "group" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "winner pick only allowed on knockout matches"})
				return
			}
			if !isOneOfTeams(*body.WinnerPickTeamID, match) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "winner pick must be one of the two teams"})
				return
			}
		}

		// Lock: now (UTC) >= kickoff => 409 unless admin (logged override).
		// Admins may write regardless of the kickoff lock; this also covers the
		// admin on-behalf-of path.
		locked := match.KickoffAt != nil && !time.Now().UTC().Before(match.KickoffAt.UTC())
		action := "prediction_update"
		switch {
		case onBehalf:
			action = "prediction_for_user"
		case locked && !isAdmin:
			c.JSON(http.StatusConflict, gin.H{"error": "predictions are locked for this match"})
			return
		case locked:
			action = "admin_override"
		}

		_, err = store.UpsertPrediction(ctx, storage.Prediction{
			UserID:           targetUserID,
			MatchID:          matchID,
			HomePred:         body.Home,
			AwayPred:         body.Away,
			WinnerPickTeamID: body.WinnerPickTeamID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save prediction"})
			return
		}

		// Audit (actions only — never the predicted values). For on-behalf-of
		// writes record actor=admin and target_user_id=forUserId.
		actorUser := claims.Sub
		entry := storage.AuditEntry{
			ActorUserID: &actorUser,
			ActorRole:   claims.Role,
			Action:      action,
			MatchID:     &matchID,
		}
		if onBehalf {
			t := targetUserID
			entry.TargetUserID = &t
		}
		_ = store.AppendAudit(ctx, entry)

		// Recompute points if the match already has a result (admin override
		// after kickoff, or late result correction).
		if rc != nil {
			_ = rc.RecomputeMatch(ctx, matchID)
		}

		c.JSON(http.StatusOK, predictionDTO{
			MatchID:          matchID,
			Home:             body.Home,
			Away:             body.Away,
			WinnerPickTeamID: body.WinnerPickTeamID,
		})
	}
}

func isOneOfTeams(teamID int64, m storage.MatchScoringRow) bool {
	return (m.HomeTeamID != nil && *m.HomeTeamID == teamID) ||
		(m.AwayTeamID != nil && *m.AwayTeamID == teamID)
}

type revealPredictionDTO struct {
	UserID           int64   `json:"userId"`
	Nickname         string  `json:"nickname"`
	AvatarURL        *string `json:"avatarUrl"`
	Home             int     `json:"home"`
	Away             int     `json:"away"`
	WinnerPickTeamID *int64  `json:"winnerPickTeamId"`
	Points           int     `json:"points"`
}

// matchPredictionsHandler reveals all predictions for a match only once it has
// kicked off (server-side gating, never trusts the client). Before kickoff it
// returns {locked:true, predictions:[]} and never leaks any values.
func matchPredictionsHandler(store PredictionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		matchID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
			return
		}

		match, err := store.GetMatchForScoring(ctx, matchID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load match"})
			return
		}

		kickedOff := match.KickoffAt != nil && !time.Now().UTC().Before(match.KickoffAt.UTC())
		if !kickedOff {
			c.JSON(http.StatusOK, gin.H{"locked": true, "predictions": []revealPredictionDTO{}})
			return
		}

		rows, err := store.ListPredictionsByMatch(ctx, matchID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load predictions"})
			return
		}
		dtos := make([]revealPredictionDTO, 0, len(rows))
		for _, p := range rows {
			dtos = append(dtos, revealPredictionDTO{
				UserID:           p.UserID,
				Nickname:         p.Nickname,
				AvatarURL:        p.AvatarURL,
				Home:             p.HomePred,
				Away:             p.AwayPred,
				WinnerPickTeamID: p.WinnerPickTeamID,
				Points:           p.Points,
			})
		}
		c.JSON(http.StatusOK, dtos)
	}
}
