package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// oauthStateCookie is the short-lived CSRF state cookie for the OAuth dance.
const oauthStateCookie = "wc_oauth_state"

// oauthEnabled reports whether Google OAuth is configured.
func oauthEnabled() bool {
	return os.Getenv("GOOGLE_OAUTH_CLIENT_ID") != "" && os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET") != ""
}

// oauthConfig builds the oauth2 config from the environment.
func oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  envDefault("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// googleUserinfo is the subset of the Google userinfo response we consume.
type googleUserinfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func googleLoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		state := randomState()
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(oauthStateCookie, state, 600, "/", "", cookieSecure(), true)
		url := oauthConfig().AuthCodeURL(state, oauth2.AccessTypeOnline)
		c.Redirect(http.StatusFound, url)
	}
}

func googleCallbackHandler(store UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		wantState, err := c.Cookie(oauthStateCookie)
		if err != nil || wantState == "" || c.Query("state") != wantState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
			return
		}
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}

		cfg := oauthConfig()
		tok, err := cfg.Exchange(ctx, code)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "token exchange failed"})
			return
		}

		info, err := fetchGoogleUserinfo(ctx, cfg, tok)
		if err != nil || info.Sub == "" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "userinfo fetch failed"})
			return
		}

		user, err := upsertGoogleUser(ctx, store, info)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}

		if err := issueAndSet(c, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}
		c.Redirect(http.StatusFound, envDefault("OAUTH_SUCCESS_REDIRECT", "/"))
	}
}

func fetchGoogleUserinfo(ctx context.Context, cfg *oauth2.Config, tok *oauth2.Token) (googleUserinfo, error) {
	client := cfg.Client(ctx, tok)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return googleUserinfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return googleUserinfo{}, errors.New("userinfo non-200")
	}
	var info googleUserinfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserinfo{}, err
	}
	return info, nil
}

// upsertGoogleUser finds the user by google_sub, creating one on first login.
// New Google users are always players; admin is provisioned only via the
// password-gated admin-login path.
func upsertGoogleUser(ctx context.Context, store UserStore, info googleUserinfo) (storage.User, error) {
	user, err := store.GetUserByGoogleSub(ctx, info.Sub)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return storage.User{}, err
	}

	role, rerr := roleForNewUser(ctx, store)
	if rerr != nil {
		return storage.User{}, rerr
	}

	nickname := info.Name
	if nickname == "" {
		nickname = info.Email
	}
	nickname = uniqueNickname(ctx, store, nickname, info.Sub)

	sub := info.Sub
	email := info.Email
	var avatar *string
	if info.Picture != "" {
		avatar = &info.Picture
	}
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	return store.CreateUser(ctx, storage.User{
		GoogleSub: &sub,
		Email:     emailPtr,
		Nickname:  nickname,
		AvatarURL: avatar,
		Role:      role,
	})
}

// uniqueNickname appends a short suffix when the desired nickname is taken.
func uniqueNickname(ctx context.Context, store UserStore, desired, sub string) string {
	if _, err := store.GetUserByNickname(ctx, desired); errors.Is(err, storage.ErrNotFound) {
		return desired
	}
	suffix := sub
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	return desired + "-" + suffix
}
