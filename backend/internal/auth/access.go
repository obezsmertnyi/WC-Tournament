package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Access levels, ordered none < ro < rw. They gate what a user may do while
// demo mode is enabled. See migration 0010 and ADR notes.
const (
	AccessNone = "none" // browse the UI only
	AccessRO   = "ro"   // also see other players' data
	AccessRW   = "rw"   // also participate (predictions + bonus picks)
)

// accessRank maps a level to a comparable rank; unknown levels rank as none.
func accessRank(level string) int {
	switch level {
	case AccessRW:
		return 2
	case AccessRO:
		return 1
	default:
		return 0
	}
}

// AccessStore is the storage capability the demo gate depends on.
type AccessStore interface {
	IsDemoMode(ctx context.Context) (bool, error)
	GetUserAccess(ctx context.Context, id int64) (string, error)
}

// ComputeAccess resolves a user's effective access level. Admins always have
// full access, and when demo mode is OFF everyone has full access (preserving
// the pre-demo behaviour). Only when demo mode is ON does a non-admin user's
// stored access_level apply.
func ComputeAccess(ctx context.Context, store AccessStore, claims Claims) (string, error) {
	if claims.Role == "admin" {
		return AccessRW, nil
	}
	on, err := store.IsDemoMode(ctx)
	if err != nil {
		return "", err
	}
	if !on {
		return AccessRW, nil
	}
	lvl, err := store.GetUserAccess(ctx, claims.Sub)
	if err != nil {
		return "", err
	}
	return lvl, nil
}

// routeAccess maps "METHOD FULLPATH" to the minimum access level required while
// demo mode is on. Routes not listed here are open to any authenticated user
// (their own/public data — calendar, groups, standings, profile, own history).
// Reads of OTHER players' data require ro; participation requires rw.
var routeAccess = map[string]string{
	"GET /api/leaderboard":             AccessRO,
	"GET /api/audit":                   AccessRO,
	"GET /api/top-scorers":             AccessRO,
	"GET /api/matches/:id/predictions": AccessRO,

	"PUT /api/predictions/:matchId": AccessRW,
	"PUT /api/bonus/champion":       AccessRW,
	"PUT /api/bonus/finalist":       AccessRW,
	"PUT /api/bonus/top-scorer":     AccessRW,
}

// DemoGate enforces per-user access levels for the guarded routes above. It is
// installed once as global middleware: Gin resolves the route before running
// the chain, so c.FullPath() is populated here. For open routes it is a no-op.
// For a guarded route it reads the session cookie directly (it runs before the
// route's own RequireUser); a missing/invalid session is left for RequireUser
// to reject with 401, while an authenticated user lacking the required level
// gets 403 with a machine-readable body so the SPA can show a "request access"
// prompt.
func DemoGate(store AccessStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		need, guarded := routeAccess[c.Request.Method+" "+c.FullPath()]
		if !guarded {
			c.Next()
			return
		}
		claims, ok := claimsFrom(c)
		if !ok {
			c.Next() // RequireUser will 401
			return
		}
		access, err := ComputeAccess(c.Request.Context(), store, claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve access"})
			return
		}
		if accessRank(access) < accessRank(need) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":  "demo_locked",
				"access": access,
				"need":   need,
			})
			return
		}
		c.Set(contextKey, claims)
		c.Next()
	}
}
