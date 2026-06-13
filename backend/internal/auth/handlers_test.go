package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeUserStore is an in-memory UserStore for auth handler tests.
type fakeUserStore struct {
	byNickname map[string]storage.User
	byGoogle   map[string]storage.User
	nextID     int64
	created    []storage.User
	audits     []storage.AuditEntry
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		byNickname: map[string]storage.User{},
		byGoogle:   map[string]storage.User{},
		nextID:     1,
	}
}

func (f *fakeUserStore) CountUsers(_ context.Context) (int, error) {
	return len(f.byNickname), nil
}

func (f *fakeUserStore) GetUserByNickname(_ context.Context, nickname string) (storage.User, error) {
	if u, ok := f.byNickname[nickname]; ok {
		return u, nil
	}
	return storage.User{}, storage.ErrNotFound
}

func (f *fakeUserStore) GetUserByGoogleSub(_ context.Context, sub string) (storage.User, error) {
	if u, ok := f.byGoogle[sub]; ok {
		return u, nil
	}
	return storage.User{}, storage.ErrNotFound
}

func (f *fakeUserStore) CreateUser(_ context.Context, u storage.User) (storage.User, error) {
	u.ID = f.nextID
	f.nextID++
	if u.Role == "" {
		u.Role = "player"
	}
	f.byNickname[u.Nickname] = u
	if u.GoogleSub != nil {
		f.byGoogle[*u.GoogleSub] = u
	}
	f.created = append(f.created, u)
	return u, nil
}

func (f *fakeUserStore) GetUserByID(_ context.Context, id int64) (storage.User, error) {
	for _, u := range f.byNickname {
		if u.ID == id {
			return u, nil
		}
	}
	return storage.User{}, storage.ErrNotFound
}

func (f *fakeUserStore) AppendAudit(_ context.Context, e storage.AuditEntry) error {
	f.audits = append(f.audits, e)
	return nil
}

func (f *fakeUserStore) seed(u storage.User) {
	if u.ID == 0 {
		u.ID = f.nextID
		f.nextID++
	}
	f.byNickname[u.Nickname] = u
	if u.GoogleSub != nil {
		f.byGoogle[*u.GoogleSub] = u
	}
}

func sp(s string) *string { return &s }

func newRouter(store UserStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, store)
	return r
}

func post(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestDevLogin_NewPlayer(t *testing.T) {
	setTestSecret(t)
	store := newFakeUserStore()
	rec := post(newRouter(store), "/api/auth/dev-login", `{"nickname":"newbie"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.created) != 1 || store.created[0].Role != "player" {
		t.Fatalf("new dev-login user must be a player, got %+v", store.created)
	}
}

func TestDevLogin_ExistingPlayerOK(t *testing.T) {
	setTestSecret(t)
	store := newFakeUserStore()
	store.seed(storage.User{Nickname: "alice", Role: "player"})
	rec := post(newRouter(store), "/api/auth/dev-login", `{"nickname":"alice"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("existing player should log in, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestDevLogin_RefusesAdmin(t *testing.T) {
	setTestSecret(t)
	store := newFakeUserStore()
	store.seed(storage.User{Nickname: "Admin", Role: "admin"})
	rec := post(newRouter(store), "/api/auth/dev-login", `{"nickname":"Admin"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("dev-login as admin must be 403, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestDevLogin_RefusesGoogleUser(t *testing.T) {
	setTestSecret(t)
	store := newFakeUserStore()
	store.seed(storage.User{Nickname: "gmail-bob", Role: "player", GoogleSub: sp("g-123")})
	rec := post(newRouter(store), "/api/auth/dev-login", `{"nickname":"gmail-bob"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("dev-login as Google user must be 403, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestAdminLogin_Disabled(t *testing.T) {
	setTestSecret(t)
	t.Setenv("ADMIN_PASSWORD", "")
	store := newFakeUserStore()
	rec := post(newRouter(store), "/api/auth/admin-login", `{"password":"anything"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("disabled admin-login must be 503, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestAdminLogin_Mismatch(t *testing.T) {
	setTestSecret(t)
	t.Setenv("ADMIN_PASSWORD", "s3cret-password-value")
	store := newFakeUserStore()
	rec := post(newRouter(store), "/api/auth/admin-login", `{"password":"wrong"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("password mismatch must be 401, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.created) != 0 {
		t.Fatalf("failed admin-login must not create a user")
	}
}

func TestAdminLogin_Match(t *testing.T) {
	setTestSecret(t)
	t.Setenv("ADMIN_PASSWORD", "s3cret-password-value")
	t.Setenv("ADMIN_NICKNAME", "Boss")
	store := newFakeUserStore()
	rec := post(newRouter(store), "/api/auth/admin-login", `{"password":"s3cret-password-value"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin-login match must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.created) != 1 || store.created[0].Role != "admin" || store.created[0].Nickname != "Boss" {
		t.Fatalf("admin-login must create the admin user, got %+v", store.created)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "admin_login" {
		t.Fatalf("admin-login must write admin_login audit, got %+v", store.audits)
	}
	// Session cookie issued.
	if !strings.Contains(rec.Header().Get("Set-Cookie"), CookieName) {
		t.Fatalf("admin-login must set the session cookie")
	}
}

func TestRoleForNewUser_AlwaysPlayer(t *testing.T) {
	store := newFakeUserStore()
	role, err := roleForNewUser(context.Background(), store)
	if err != nil || role != "player" {
		t.Fatalf("roleForNewUser must always be player, got %q err=%v", role, err)
	}
}
