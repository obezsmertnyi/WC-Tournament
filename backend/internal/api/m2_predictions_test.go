package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// fakePredStore is an in-memory PredictionStore for handler tests.
type fakePredStore struct {
	match    storage.MatchScoringRow
	matchErr error
	byUser   []storage.Prediction
	byMatch  []storage.MatchPrediction
	upserts  []storage.Prediction
	audits   []storage.AuditEntry
}

func (f *fakePredStore) GetMatchForScoring(_ context.Context, _ int64) (storage.MatchScoringRow, error) {
	return f.match, f.matchErr
}
func (f *fakePredStore) UpsertPrediction(_ context.Context, p storage.Prediction) (storage.Prediction, error) {
	f.upserts = append(f.upserts, p)
	p.ID = int64(len(f.upserts))
	return p, nil
}
func (f *fakePredStore) ListPredictionsByUser(_ context.Context, _ int64) ([]storage.Prediction, error) {
	return f.byUser, nil
}
func (f *fakePredStore) ListPredictionsByMatch(_ context.Context, _ int64) ([]storage.MatchPrediction, error) {
	return f.byMatch, nil
}
func (f *fakePredStore) AppendAudit(_ context.Context, e storage.AuditEntry) error {
	f.audits = append(f.audits, e)
	return nil
}

// setTestSecret installs an ephemeral random 32+ byte JWT_SECRET for the test.
func setTestSecret(t *testing.T) {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	t.Setenv("JWT_SECRET", hex.EncodeToString(b))
}

// sessionCookie builds a valid wc_session cookie header value for a user. It
// also installs an ephemeral JWT_SECRET so issued tokens verify in-process.
func sessionCookie(t *testing.T, id int64, role string) string {
	t.Helper()
	setTestSecret(t)
	tok, err := auth.IssueToken(id, role, "tester")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return auth.CookieName + "=" + tok
}

func i64p(v int64) *int64       { return &v }
func tp(t time.Time) *time.Time { return &t }

func TestPutPrediction_LockedReturns409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	past := time.Now().UTC().Add(-time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 5, Stage: "group", KickoffAt: tp(past)}}

	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	body := strings.NewReader(`{"home":1,"away":0}`)
	req := httptest.NewRequest(http.MethodPut, "/api/predictions/5", body)
	req.Header.Set("Cookie", sessionCookie(t, 1, "player"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	if len(store.upserts) != 0 {
		t.Errorf("locked prediction must not be persisted")
	}
}

func TestPutPrediction_AdminOverrideAfterLock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	past := time.Now().UTC().Add(-time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 5, Stage: "group", KickoffAt: tp(past)}}

	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/predictions/5", strings.NewReader(`{"home":2,"away":1}`))
	req.Header.Set("Cookie", sessionCookie(t, 1, "admin"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("admin override expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.upserts) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(store.upserts))
	}
	if len(store.audits) != 1 || store.audits[0].Action != "admin_override" {
		t.Fatalf("expected admin_override audit, got %+v", store.audits)
	}
}

func TestPutPrediction_OpenSucceedsAndAudits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	future := time.Now().UTC().Add(time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 9, Stage: "group", KickoffAt: tp(future)}}

	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/predictions/9", strings.NewReader(`{"home":1,"away":1}`))
	req.Header.Set("Cookie", sessionCookie(t, 7, "player"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(store.audits) != 1 || store.audits[0].Action != "prediction_update" {
		t.Fatalf("expected prediction_update audit, got %+v", store.audits)
	}
}

func TestPutPrediction_RejectsOutOfRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	future := time.Now().UTC().Add(time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 1, Stage: "group", KickoffAt: tp(future)}}
	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/predictions/1", strings.NewReader(`{"home":99,"away":0}`))
	req.Header.Set("Cookie", sessionCookie(t, 1, "player"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range, got %d", rec.Code)
	}
}

func TestPutPrediction_WinnerPickGroupRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	future := time.Now().UTC().Add(time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 1, Stage: "group", KickoffAt: tp(future), HomeTeamID: i64p(10), AwayTeamID: i64p(20)}}
	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/predictions/1", strings.NewReader(`{"home":1,"away":1,"winnerPickTeamId":10}`))
	req.Header.Set("Cookie", sessionCookie(t, 1, "player"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("winner pick on group should be 400, got %d", rec.Code)
	}
}

func TestPutPrediction_WinnerPickMustBeOneOfTeams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	future := time.Now().UTC().Add(time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 1, Stage: "r32", KickoffAt: tp(future), HomeTeamID: i64p(10), AwayTeamID: i64p(20)}}
	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	// Pick a team not in the match.
	req := httptest.NewRequest(http.MethodPut, "/api/predictions/1", strings.NewReader(`{"home":1,"away":1,"winnerPickTeamId":99}`))
	req.Header.Set("Cookie", sessionCookie(t, 1, "player"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid winner pick should be 400, got %d", rec.Code)
	}
}

func TestPutPrediction_AnonymousRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakePredStore{}
	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)
	req := httptest.NewRequest(http.MethodPut, "/api/predictions/1", strings.NewReader(`{"home":1,"away":0}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous should be 401, got %d", rec.Code)
	}
}

func TestMatchPredictions_RevealGating(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Pre-kickoff: must NOT leak any prediction values.
	future := time.Now().UTC().Add(time.Hour)
	store := &fakePredStore{
		match: storage.MatchScoringRow{ID: 3, Stage: "group", KickoffAt: tp(future)},
		byMatch: []storage.MatchPrediction{
			{UserID: 1, Nickname: "a", HomePred: 2, AwayPred: 1, Points: 3},
		},
	}
	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	cookie := sessionCookie(t, 1, "player")

	req := httptest.NewRequest(http.MethodGet, "/api/matches/3/predictions", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var locked struct {
		Locked      bool              `json:"locked"`
		Predictions []json.RawMessage `json:"predictions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &locked); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !locked.Locked || len(locked.Predictions) != 0 {
		t.Fatalf("pre-kickoff must be locked with empty predictions, got %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"home"`) || strings.Contains(rec.Body.String(), "homePred") {
		t.Fatalf("pre-kickoff response leaked prediction values: %s", rec.Body.String())
	}

	// Post-kickoff: reveal values.
	past := time.Now().UTC().Add(-time.Hour)
	store.match = storage.MatchScoringRow{ID: 3, Stage: "group", KickoffAt: tp(past)}
	req2 := httptest.NewRequest(http.MethodGet, "/api/matches/3/predictions", nil)
	req2.Header.Set("Cookie", cookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	var revealed []map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &revealed); err != nil {
		t.Fatalf("decode reveal: %v (body=%s)", err, rec2.Body.String())
	}
	if len(revealed) != 1 || revealed[0]["home"] != float64(2) || revealed[0]["points"] != float64(3) {
		t.Fatalf("post-kickoff reveal mismatch: %s", rec2.Body.String())
	}
}

func TestMatchPredictions_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	past := time.Now().UTC().Add(-time.Hour)
	store := &fakePredStore{match: storage.MatchScoringRow{ID: 3, Stage: "group", KickoffAt: tp(past)}}
	r := gin.New()
	RegisterPredictionRoutes(r, store, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/matches/3/predictions", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reveal must require auth (401), got %d (%s)", rec.Code, rec.Body.String())
	}
}
