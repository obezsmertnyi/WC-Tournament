package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

const (
	maxNicknameRunes = 32
	maxAvatarBytes   = 512
)

// ProfileStore is the storage capability the profile endpoints depend on.
type ProfileStore interface {
	GetUserByID(ctx context.Context, id int64) (storage.User, error)
	UpdateUserProfile(ctx context.Context, id int64, nickname, favoriteTeamCode, avatarURL *string) (storage.User, error)
	TeamCodeExists(ctx context.Context, code string) (bool, error)
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
		ctx := c.Request.Context()

		if body.Nickname != nil {
			trimmed := strings.TrimSpace(*body.Nickname)
			if msg := validateNickname(trimmed); msg != "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
			body.Nickname = &trimmed
		}

		if body.AvatarURL != nil && *body.AvatarURL != "" {
			if msg := validateAvatarURL(*body.AvatarURL); msg != "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
		}

		if body.FavoriteTeamCode != nil && *body.FavoriteTeamCode != "" {
			ok, err := store.TeamCodeExists(ctx, *body.FavoriteTeamCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate team"})
				return
			}
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unknown favorite team code"})
				return
			}
		}

		u, err := store.UpdateUserProfile(ctx, claims.Sub, body.Nickname, body.FavoriteTeamCode, body.AvatarURL)
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

// validateNickname returns an error message (empty when valid) for a trimmed
// nickname: 1..32 runes; only letters, digits, space, underscore or hyphen;
// no control characters.
func validateNickname(s string) string {
	if s == "" {
		return "nickname cannot be empty"
	}
	if utf8.RuneCountInString(s) > maxNicknameRunes {
		return "nickname must be at most 32 characters"
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return "nickname contains invalid characters"
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '_' || r == '-' {
			continue
		}
		return "nickname contains invalid characters"
	}
	return ""
}

// validateAvatarURL returns an error message (empty when valid) for a non-empty
// avatar URL: parseable, http/https scheme, non-empty host, at most 512 bytes.
func validateAvatarURL(s string) string {
	if len(s) > maxAvatarBytes {
		return "avatarUrl too long"
	}
	u, err := url.Parse(s)
	if err != nil {
		return "invalid avatarUrl"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "avatarUrl must be http or https"
	}
	if u.Host == "" {
		return "avatarUrl must have a host"
	}
	return ""
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
