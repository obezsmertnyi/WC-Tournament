package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
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
	AppendAudit(ctx context.Context, e storage.AuditEntry) error
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
	r.POST("/api/auth/admin-login", adminLoginHandler(store))
	r.POST("/api/auth/logout", logoutHandler())

	if oauthEnabled() {
		r.GET("/api/auth/google/login", googleLoginHandler())
		r.GET("/api/auth/google/callback", googleCallbackHandler(store))
	}
}

// roleForNewUser always returns 'player'. Admin accounts are never provisioned
// implicitly (no admin-by-count); they exist only via the password-gated
// admin-login path. Used by the Google OAuth upsert path.
func roleForNewUser(_ context.Context, _ UserStore) (string, error) {
	return "player", nil
}

// devLoginHandler logs in an EXISTING player by nickname (Friends-PoC trust
// model). It is LOGIN-ONLY: it never creates an account, so random names and
// typos can no longer spawn junk accounts — the player roster is provisioned
// by an admin via POST /api/admin/users. An unknown nickname returns 404. It
// REFUSES (403) to log in as any account that is an admin or is linked to
// Google, so this path can never impersonate a privileged or federated
// identity and never issues an admin token.
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
		switch {
		case errors.Is(err, storage.ErrNotFound):
			// No auto-create: unknown nicknames are rejected so an admin must
			// add the player to the roster first.
			c.JSON(http.StatusNotFound, gin.H{"error": "no such player — ask an admin to add you"})
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		default:
			// Existing account: only plain player accounts may use dev-login.
			// Admin or Google-linked accounts are protected from impersonation.
			if user.Role != "player" || user.GoogleSub != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "this account cannot use nickname login"})
				return
			}
		}

		if err := issueAndSet(c, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}
		c.JSON(http.StatusOK, ToUserDTO(user))
	}
}

// adminLoginHandler authenticates the single admin account with a password
// compared in constant time against ADMIN_PASSWORD. When ADMIN_PASSWORD is
// unset the endpoint is disabled (503). On success it finds-or-creates the
// admin user (nickname from ADMIN_NICKNAME, default "Admin"), issues its
// session cookie and writes an admin_login audit row. On mismatch it returns
// 401 with no user enumeration.
func adminLoginHandler(store UserStore) gin.HandlerFunc {
	type req struct {
		Password string `json:"password"`
	}
	return func(c *gin.Context) {
		want := os.Getenv("ADMIN_PASSWORD")
		if want == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "admin login disabled"})
			return
		}

		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		// Constant-time compare to avoid leaking the password via timing.
		if subtle.ConstantTimeCompare([]byte(body.Password), []byte(want)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		ctx := c.Request.Context()
		nickname := envDefault("ADMIN_NICKNAME", "Admin")

		user, err := store.GetUserByNickname(ctx, nickname)
		if errors.Is(err, storage.ErrNotFound) {
			user, err = store.CreateUser(ctx, storage.User{Nickname: nickname, Role: "admin"})
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}
		// Defensive: if an account already exists under the admin nickname but
		// is not actually an admin, do not silently elevate it.
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin account misconfigured"})
			return
		}

		if err := issueAndSet(c, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}

		actor := user.ID
		_ = store.AppendAudit(ctx, storage.AuditEntry{
			ActorUserID: &actor,
			ActorRole:   "admin",
			Action:      "admin_login",
		})

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
