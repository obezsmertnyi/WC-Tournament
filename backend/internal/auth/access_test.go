package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeAccessStore is an in-memory AccessStore for gate tests.
type fakeAccessStore struct {
	demo   bool
	levels map[int64]string
	err    error
}

func (f *fakeAccessStore) IsDemoMode(context.Context) (bool, error) { return f.demo, f.err }
func (f *fakeAccessStore) GetUserAccess(_ context.Context, id int64) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.levels[id], nil
}

func TestComputeAccess(t *testing.T) {
	cases := []struct {
		name  string
		role  string
		demo  bool
		level string
		want  string
	}{
		{"admin always rw (demo on)", "admin", true, "none", AccessRW},
		{"admin always rw (demo off)", "admin", false, "", AccessRW},
		{"demo off → everyone rw", "player", false, "none", AccessRW},
		{"demo on → none stays none", "player", true, AccessNone, AccessNone},
		{"demo on → ro stays ro", "player", true, AccessRO, AccessRO},
		{"demo on → rw stays rw", "player", true, AccessRW, AccessRW},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeAccessStore{demo: tc.demo, levels: map[int64]string{7: tc.level}}
			got, err := ComputeAccess(context.Background(), store, Claims{Sub: 7, Role: tc.role})
			if err != nil {
				t.Fatalf("ComputeAccess: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

// gateRouter builds a router with DemoGate installed and the guarded routes
// registered with a trivial 200 handler behind RequireUser (mirrors prod).
func gateRouter(store AccessStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DemoGate(store))
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	r.GET("/api/leaderboard", RequireUser(), ok)             // ro
	r.PUT("/api/predictions/:matchId", RequireUser(), ok)    // rw
	r.GET("/api/matches/:id/predictions", RequireUser(), ok) // ro
	r.GET("/api/matches", RequireUser(), ok)                 // open
	return r
}

func reqWithUser(t *testing.T, method, path string, uid int64, role string) *http.Request {
	t.Helper()
	tok, err := IssueToken(uid, role, "n")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: tok})
	return req
}

func TestDemoGate(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-test-secret-test-secret-123456")

	cases := []struct {
		name   string
		demo   bool
		level  string // for uid 7
		role   string
		method string
		path   string
		want   int
	}{
		{"demo off: none-user can read others", false, AccessNone, "player", http.MethodGet, "/api/leaderboard", http.StatusOK},
		{"demo on: none blocked from leaderboard", true, AccessNone, "player", http.MethodGet, "/api/leaderboard", http.StatusForbidden},
		{"demo on: none blocked from reveals", true, AccessNone, "player", http.MethodGet, "/api/matches/12/predictions", http.StatusForbidden},
		{"demo on: none blocked from participating", true, AccessNone, "player", http.MethodPut, "/api/predictions/12", http.StatusForbidden},
		{"demo on: none can still browse open route", true, AccessNone, "player", http.MethodGet, "/api/matches", http.StatusOK},
		{"demo on: ro can read others", true, AccessRO, "player", http.MethodGet, "/api/leaderboard", http.StatusOK},
		{"demo on: ro cannot participate", true, AccessRO, "player", http.MethodPut, "/api/predictions/12", http.StatusForbidden},
		{"demo on: rw can participate", true, AccessRW, "player", http.MethodPut, "/api/predictions/12", http.StatusOK},
		{"demo on: admin bypasses all", true, AccessNone, "admin", http.MethodPut, "/api/predictions/12", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeAccessStore{demo: tc.demo, levels: map[int64]string{7: tc.level}}
			r := gateRouter(store)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, reqWithUser(t, tc.method, tc.path, 7, tc.role))
			if rec.Code != tc.want {
				t.Fatalf("%s %s: got %d want %d (body %s)", tc.method, tc.path, rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestDemoGateAnonymousFallsThroughTo401(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-test-secret-test-secret-123456")
	store := &fakeAccessStore{demo: true, levels: map[int64]string{}}
	r := gateRouter(store)
	rec := httptest.NewRecorder()
	// No cookie: the gate must not 403; RequireUser owns the 401.
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/leaderboard", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anon guarded route: got %d want 401", rec.Code)
	}
}
