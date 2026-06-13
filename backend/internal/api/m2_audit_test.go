package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

type fakeAuditStore struct {
	rows []storage.AuditEntry
}

func (f *fakeAuditStore) ListAudit(_ context.Context, _ int) ([]storage.AuditEntry, error) {
	return f.rows, nil
}

func TestAudit_NeverLeaksValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mid := int64(42)
	store := &fakeAuditStore{rows: []storage.AuditEntry{
		{ID: 2, ActorNickname: "alice", ActorRole: "player", Action: "prediction_update", MatchID: &mid, CreatedAt: time.Unix(1700000000, 0).UTC()},
		{ID: 1, ActorNickname: "", ActorRole: "system", Action: "admin_override", CreatedAt: time.Unix(1699999999, 0).UTC()},
	}}

	r := gin.New()
	RegisterAuditRoutes(r, store)

	req := httptest.NewRequest(http.MethodGet, "/api/audit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()

	// The DTO must only expose actor/action/matchId/createdAt — no score fields.
	for _, forbidden := range []string{"home", "away", "score", "homePred", "awayPred", "points", "breakdown"} {
		if regexp.MustCompile(`"` + forbidden + `"`).MatchString(body) {
			t.Errorf("audit response leaked forbidden field %q: %s", forbidden, body)
		}
	}
	// System actor falls back to "system".
	if !regexp.MustCompile(`"actor":"system"`).MatchString(body) {
		t.Errorf("expected system actor fallback in %s", body)
	}
	if !regexp.MustCompile(`"actor":"alice"`).MatchString(body) {
		t.Errorf("expected alice actor in %s", body)
	}
}
