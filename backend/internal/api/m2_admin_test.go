package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeAdminStore is an in-memory AdminStore for handler tests.
type fakeAdminStore struct {
	users []storage.User
}

func (f *fakeAdminStore) ListUsers(_ context.Context) ([]storage.User, error) {
	return f.users, nil
}

func strp(s string) *string { return &s }

func TestAdminListUsers_RequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{users: []storage.User{
		{ID: 1, Nickname: "alice", AvatarURL: strp("a.png"), FavoriteTeamCode: strp("ARG"), Role: "admin"},
		{ID: 2, Nickname: "bob", Role: "player"},
	}}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	// Anonymous -> 401.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anon expected 401, got %d (%s)", rec.Code, rec.Body.String())
	}

	// Non-admin -> 403.
	req2 := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req2.Header.Set("Cookie", sessionCookie(t, 2, "player"))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("player expected 403, got %d (%s)", rec2.Code, rec2.Body.String())
	}

	// Admin -> 200 with the picker shape.
	req3 := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req3.Header.Set("Cookie", sessionCookie(t, 1, "admin"))
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("admin expected 200, got %d (%s)", rec3.Code, rec3.Body.String())
	}
	var got []adminUserDTO
	if err := json.Unmarshal(rec3.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v (body=%s)", err, rec3.Body.String())
	}
	if len(got) != 2 || got[0].Nickname != "alice" || got[0].Role != "admin" {
		t.Fatalf("unexpected users payload: %s", rec3.Body.String())
	}
	if got[0].FavoriteTeamCode == nil || *got[0].FavoriteTeamCode != "ARG" {
		t.Fatalf("expected favoriteTeamCode ARG, got %s", rec3.Body.String())
	}
}
