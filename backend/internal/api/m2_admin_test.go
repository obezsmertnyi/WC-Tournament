package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeAdminStore is an in-memory AdminStore for handler tests.
type fakeAdminStore struct {
	users      []storage.User
	nextID     int64
	audits     []storage.AuditEntry
	cascadeIDs []int64
	createErr  error
}

func (f *fakeAdminStore) ListUsers(_ context.Context) ([]storage.User, error) {
	return f.users, nil
}

func (f *fakeAdminStore) CreatePlayer(_ context.Context, nickname string) (storage.User, error) {
	if f.createErr != nil {
		return storage.User{}, f.createErr
	}
	for _, u := range f.users {
		if u.Nickname == nickname {
			return storage.User{}, &pgconn.PgError{Code: "23505"}
		}
	}
	if f.nextID == 0 {
		f.nextID = 100
	}
	f.nextID++
	u := storage.User{ID: f.nextID, Nickname: nickname, Role: "player"}
	f.users = append(f.users, u)
	return u, nil
}

func (f *fakeAdminStore) GetUserByID(_ context.Context, id int64) (storage.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return storage.User{}, storage.ErrNotFound
}

func (f *fakeAdminStore) DeleteUserCascade(_ context.Context, id int64) error {
	for i, u := range f.users {
		if u.ID == id {
			f.users = append(f.users[:i], f.users[i+1:]...)
			f.cascadeIDs = append(f.cascadeIDs, id)
			return nil
		}
	}
	return storage.ErrNotFound
}

func (f *fakeAdminStore) AppendAudit(_ context.Context, e storage.AuditEntry) error {
	f.audits = append(f.audits, e)
	return nil
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

// adminReq builds an admin-authenticated request for the roster endpoints.
func adminReq(t *testing.T, method, path, body string, adminID int64) *http.Request {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("Cookie", sessionCookie(t, adminID, "admin"))
	return r
}

func TestAdminCreateUser_RequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	// Anonymous -> 401.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", strings.NewReader(`{"nickname":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anon create expected 401, got %d (%s)", rec.Code, rec.Body.String())
	}

	// Player -> 403.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/admin/users", strings.NewReader(`{"nickname":"x"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Cookie", sessionCookie(t, 9, "player"))
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("player create expected 403, got %d (%s)", rec2.Code, rec2.Body.String())
	}
	if len(store.users) != 0 {
		t.Fatalf("unauthorized create must not add a user, got %+v", store.users)
	}
}

func TestAdminCreateUser_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, adminReq(t, http.MethodPost, "/api/admin/users", `{"nickname":"  Charlie  "}`, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("create expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var got adminUserDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v (body=%s)", err, rec.Body.String())
	}
	if got.Nickname != "Charlie" || got.Role != "player" || got.ID == 0 {
		t.Fatalf("unexpected created DTO: %s", rec.Body.String())
	}
	if len(store.audits) != 1 || store.audits[0].Action != "admin_create_user" {
		t.Fatalf("expected admin_create_user audit, got %+v", store.audits)
	}
	if store.audits[0].TargetUserID == nil || *store.audits[0].TargetUserID != got.ID {
		t.Fatalf("audit target_user_id must equal new id, got %+v", store.audits[0])
	}
	if store.audits[0].ActorUserID == nil || *store.audits[0].ActorUserID != 1 {
		t.Fatalf("audit actor must be the admin (1), got %+v", store.audits[0])
	}
}

func TestAdminCreateUser_Duplicate409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{users: []storage.User{{ID: 5, Nickname: "dupe", Role: "player"}}}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, adminReq(t, http.MethodPost, "/api/admin/users", `{"nickname":"dupe"}`, 1))
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate nickname expected 409, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestAdminCreateUser_InvalidNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, adminReq(t, http.MethodPost, "/api/admin/users", `{"nickname":"  "}`, 1))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty nickname expected 400, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.users) != 0 {
		t.Fatalf("invalid create must not add a user")
	}
}

func TestAdminDeleteUser_PlayerCascade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{users: []storage.User{
		{ID: 1, Nickname: "Admin", Role: "admin"},
		{ID: 7, Nickname: "victim", Role: "player"},
	}}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, adminReq(t, http.MethodDelete, "/api/admin/users/7", "", 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete player expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.cascadeIDs) != 1 || store.cascadeIDs[0] != 7 {
		t.Fatalf("expected cascade delete of user 7, got %+v", store.cascadeIDs)
	}
	for _, u := range store.users {
		if u.ID == 7 {
			t.Fatalf("user 7 must be gone, got %+v", store.users)
		}
	}
	if len(store.audits) != 1 || store.audits[0].Action != "admin_delete_user" {
		t.Fatalf("expected admin_delete_user audit, got %+v", store.audits)
	}
	if store.audits[0].TargetUserID == nil || *store.audits[0].TargetUserID != 7 {
		t.Fatalf("audit target_user_id must be 7, got %+v", store.audits[0])
	}
}

func TestAdminDeleteUser_RefusesAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{users: []storage.User{
		{ID: 1, Nickname: "Admin", Role: "admin"},
		{ID: 2, Nickname: "Other", Role: "admin"},
	}}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, adminReq(t, http.MethodDelete, "/api/admin/users/2", "", 1))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("deleting an admin must be 403, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.cascadeIDs) != 0 {
		t.Fatalf("admin delete must not cascade, got %+v", store.cascadeIDs)
	}
	if len(store.audits) != 0 {
		t.Fatalf("refused delete must not audit, got %+v", store.audits)
	}
}

func TestAdminDeleteUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAdminStore{users: []storage.User{{ID: 1, Nickname: "Admin", Role: "admin"}}}
	r := gin.New()
	RegisterAdminRoutes(r, store)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, adminReq(t, http.MethodDelete, "/api/admin/users/999", "", 1))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown id expected 404, got %d (%s)", rec.Code, rec.Body.String())
	}
}
