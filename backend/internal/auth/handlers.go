package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// UserStore is the storage capability the auth handlers depend on.
type UserStore interface {
	CountUsers(ctx context.Context) (int, error)
	GetUserByNickname(ctx context.Context, nickname string) (storage.User, error)
	GetUserByGoogleSub(ctx context.Context, sub string) (storage.User, error)
	CreateUser(ctx context.Context, u storage.User) (storage.User, error)
	GetUserByID(ctx context.Context, id int64) (storage.User, error)
}

// userDTO is the public user shape returned by auth + profile endpoints.
type userDTO struct {
	ID               int64   `json:"id"`
	Nickname         string  `json:"nickname"`
	AvatarURL        *string `json:"avatarUrl"`
	FavoriteTeamCode *string `json:"favoriteTeamCode"`
	Role             string  `json:"role"`
}

// ToUserDTO maps a storage.User to the public DTO.
func ToUserDTO(u storage.User) userDTO {
	return userDTO{
		ID:               u.ID,
		Nickname:         u.Nickname,
		AvatarURL:        u.AvatarURL,
		FavoriteTeamCode: u.FavoriteTeamCode,
		Role:             u.Role,
	}
}

// RegisterRoutes wires the auth endpoints. Google OAuth routes are only active
// when GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET are set.
func RegisterRoutes(r gin.IRouter, store UserStore) {
	r.POST("/api/auth/dev-login", devLoginHandler(store))
	r.POST("/api/auth/logout", logoutHandler())

	if oauthEnabled() {
		r.GET("/api/auth/google/login", googleLoginHandler())
		r.GET("/api/auth/google/callback", googleCallbackHandler(store))
	}
}

// roleForNewUser assigns 'admin' to the very first user, 'player' otherwise.
func roleForNewUser(ctx context.Context, store UserStore) (string, error) {
	n, err := store.CountUsers(ctx)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "admin", nil
	}
	return "player", nil
}

// devLoginHandler finds or creates a user by nickname, sets the session cookie
// and returns the user. Works without Google credentials (PoC).
func devLoginHandler(store UserStore) gin.HandlerFunc {
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
		if nickname == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nickname required"})
			return
		}

		ctx := c.Request.Context()
		user, err := store.GetUserByNickname(ctx, nickname)
		if errors.Is(err, storage.ErrNotFound) {
			role, rerr := roleForNewUser(ctx, store)
			if rerr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
				return
			}
			user, err = store.CreateUser(ctx, storage.User{Nickname: nickname, Role: role})
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}

		if err := issueAndSet(c, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}
		c.JSON(http.StatusOK, ToUserDTO(user))
	}
}

func issueAndSet(c *gin.Context, u storage.User) error {
	token, err := IssueToken(u.ID, u.Role, u.Nickname)
	if err != nil {
		return err
	}
	SetSessionCookie(c, token)
	return nil
}

// logoutHandler clears the session cookie.
func logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ClearSessionCookie(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// randomState returns a URL-safe random string for the OAuth state cookie.
func randomState() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
