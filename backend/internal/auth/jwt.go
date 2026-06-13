// Package auth provides JWT-cookie session authentication for the
// WC-Tournament backend: a minimal HS256 token implementation (no external
// dependency), Gin middleware (RequireUser / RequireAdmin), and the auth
// HTTP handlers (dev-login, optional Google OAuth, logout).
//
// The session is a JWT stored in an HttpOnly cookie named "wc_session".
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// CookieName is the session cookie name.
const CookieName = "wc_session"

// sessionTTL is how long an issued session is valid.
const sessionTTL = 30 * 24 * time.Hour

// minSecretLen is the minimum acceptable JWT_SECRET length in bytes.
const minSecretLen = 32

// Claims is the JWT payload we issue.
type Claims struct {
	Sub      int64  `json:"sub"` // user id
	Role     string `json:"role"`
	Nickname string `json:"nickname"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// secret returns the signing secret from JWT_SECRET. It never falls back to a
// default: callers must have validated the secret at boot via ValidateSecret.
func secret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// ValidateSecret reports whether JWT_SECRET is set and long enough to be safe.
// It must be called at boot (main.go) and the process must refuse to start when
// it returns an error — there is no insecure default.
func ValidateSecret() error {
	s := os.Getenv("JWT_SECRET")
	if len(s) < minSecretLen {
		return fmt.Errorf("JWT_SECRET must be set and at least %d bytes (got %d)", minSecretLen, len(s))
	}
	return nil
}

var b64 = base64.RawURLEncoding

// IssueToken creates a signed HS256 JWT for the given user.
func IssueToken(userID int64, role, nickname string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Sub:      userID,
		Role:     role,
		Nickname: nickname,
		Iat:      now.Unix(),
		Exp:      now.Add(sessionTTL).Unix(),
	}
	header := `{"alg":"HS256","typ":"JWT"}`
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	signingInput := b64.EncodeToString([]byte(header)) + "." + b64.EncodeToString(payload)
	sig := sign(signingInput)
	return signingInput + "." + sig, nil
}

func sign(input string) string {
	mac := hmac.New(sha256.New, secret())
	mac.Write([]byte(input))
	return b64.EncodeToString(mac.Sum(nil))
}

// ErrInvalidToken is returned for any malformed/expired/unsigned token.
var ErrInvalidToken = errors.New("invalid session token")

// ParseToken verifies the signature + expiry and returns the claims.
func ParseToken(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	expected := sign(signingInput)
	// Constant-time compare.
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := b64.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if c.Exp != 0 && time.Now().UTC().Unix() >= c.Exp {
		return Claims{}, ErrInvalidToken
	}
	return c, nil
}
