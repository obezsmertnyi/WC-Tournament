package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// contextKey is the gin context key under which the current user's claims are
// stored.
const contextKey = "currentUser"

// CookieMaxAge is the session cookie lifetime in seconds.
const CookieMaxAge = 30 * 24 * 60 * 60

// SetSessionCookie writes the session JWT as an HttpOnly cookie. Secure is set
// when COOKIE_SECURE=true (production behind TLS); SameSite=Lax for OAuth
// redirect compatibility.
func SetSessionCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CookieName, token, CookieMaxAge, "/", "", cookieSecure(), true)
}

// ClearSessionCookie expires the session cookie.
func ClearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CookieName, "", -1, "/", "", cookieSecure(), true)
}

func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

// claimsFrom extracts and verifies the session claims from the request cookie.
func claimsFrom(c *gin.Context) (Claims, bool) {
	tok, err := c.Cookie(CookieName)
	if err != nil || tok == "" {
		return Claims{}, false
	}
	claims, err := ParseToken(tok)
	if err != nil {
		return Claims{}, false
	}
	return claims, true
}

// Optional loads the current user into the context when a valid session exists
// but never rejects the request. Useful for endpoints whose behavior differs
// for admins (e.g. the prediction lock override).
func Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, ok := claimsFrom(c); ok {
			c.Set(contextKey, claims)
		}
		c.Next()
	}
}

// RequireUser rejects unauthenticated requests with 401 and otherwise stores
// the current user's claims in the context.
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := claimsFrom(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Set(contextKey, claims)
		c.Next()
	}
}

// RequireAdmin rejects non-admins with 403 (and anon with 401).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := claimsFrom(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		if claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		c.Set(contextKey, claims)
		c.Next()
	}
}

// Current returns the current user's claims from the context, if present.
func Current(c *gin.Context) (Claims, bool) {
	v, ok := c.Get(contextKey)
	if !ok {
		return Claims{}, false
	}
	claims, ok := v.(Claims)
	return claims, ok
}

// IsAdmin reports whether the current request is by an admin.
func IsAdmin(c *gin.Context) bool {
	claims, ok := Current(c)
	return ok && claims.Role == "admin"
}
