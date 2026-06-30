package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterHealthRoutes(r)

	cases := []struct {
		name string
		path string
	}{
		{name: "root healthz", path: "/healthz"},
		{name: "api healthz", path: "/api/healthz"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			var got healthResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
			}

			want := healthResponse{Status: "ok", Service: serviceName, Version: Version}
			if got != want {
				t.Fatalf("unexpected body: got %+v, want %+v", got, want)
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Fatalf("unexpected content-type: %q", ct)
			}
		})
	}
}
