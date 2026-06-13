package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakeProfileStore is an in-memory ProfileStore for profile handler tests.
type fakeProfileStore struct {
	user      storage.User
	teamCodes map[string]bool
	updated   bool
}

func (f *fakeProfileStore) GetUserByID(_ context.Context, _ int64) (storage.User, error) {
	return f.user, nil
}

func (f *fakeProfileStore) UpdateUserProfile(_ context.Context, _ int64, nickname, favoriteTeamCode, avatarURL *string) (storage.User, error) {
	f.updated = true
	u := f.user
	if nickname != nil {
		u.Nickname = *nickname
	}
	if favoriteTeamCode != nil {
		u.FavoriteTeamCode = favoriteTeamCode
	}
	if avatarURL != nil {
		u.AvatarURL = avatarURL
	}
	f.user = u
	return u, nil
}

func (f *fakeProfileStore) TeamCodeExists(_ context.Context, code string) (bool, error) {
	return f.teamCodes[code], nil
}

func patchMe(t *testing.T, store ProfileStore, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterProfileRoutes(r, store)
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sessionCookie(t, 1, "player"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func patchMeJSON(t *testing.T, store ProfileStore, payload map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return patchMe(t, store, string(b))
}

func newProfileStore() *fakeProfileStore {
	return &fakeProfileStore{
		user:      storage.User{ID: 1, Nickname: "old", Role: "player"},
		teamCodes: map[string]bool{"BRA": true},
	}
}

func TestPatchMe_ValidNickname(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"nickname":"Cool_User-1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestPatchMe_RejectsEmptyNickname(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"nickname":"   "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty nickname must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_RejectsLongNickname(t *testing.T) {
	long := strings.Repeat("a", 33)
	rec := patchMeJSON(t, newProfileStore(), map[string]string{"nickname": long})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("33-char nickname must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_RejectsControlCharNickname(t *testing.T) {
	// A real control rune (U+0007 BEL) in the nickname must be rejected.
	bad := "bad" + string(rune(7)) + "name"
	rec := patchMeJSON(t, newProfileStore(), map[string]string{"nickname": bad})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("control-char nickname must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_RejectsPunctuationNickname(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"nickname":"drop;table"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("punctuation nickname must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_ValidAvatarURL(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"avatarUrl":"https://example.com/a.png"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid avatar must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestPatchMe_RejectsBadAvatarScheme(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"avatarUrl":"javascript:alert(1)"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("javascript scheme must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_RejectsHostlessAvatar(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"avatarUrl":"http:///nohost"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("hostless avatar must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_RejectsLongAvatar(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 600)
	rec := patchMeJSON(t, newProfileStore(), map[string]string{"avatarUrl": long})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("over-long avatar must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_RejectsUnknownTeamCode(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"favoriteTeamCode":"ZZZ"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown team code must be 400, got %d", rec.Code)
	}
}

func TestPatchMe_AcceptsKnownTeamCode(t *testing.T) {
	rec := patchMe(t, newProfileStore(), `{"favoriteTeamCode":"BRA"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("known team code must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}
