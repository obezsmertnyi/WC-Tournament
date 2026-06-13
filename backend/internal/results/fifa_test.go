package results

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TestFIFAClient_Fixtures drives the HTTP client against an httptest server
// serving the checked-in sample (no real network). It also asserts the
// User-Agent / Accept headers and single-page termination on a null token.
func TestFIFAClient_Fixtures(t *testing.T) {
	body, err := os.ReadFile("testdata/fifa_calendar_sample.json")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}

	var gotUA, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		if r.URL.Path != "/calendar/matches" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := NewFIFAClient(
		WithBaseURL(srv.URL),
		WithClock(func() time.Time { return fixedNow }),
	)

	fixtures, err := client.Fixtures(context.Background())
	if err != nil {
		t.Fatalf("Fixtures: %v", err)
	}
	if len(fixtures) != 3 {
		t.Fatalf("expected 3 fixtures, got %d", len(fixtures))
	}
	if gotUA != fifaUserAgent {
		t.Errorf("User-Agent: got %q want %q", gotUA, fifaUserAgent)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept: got %q", gotAccept)
	}
	if fixtures[0].Home.Name != "Mexico" {
		t.Errorf("first fixture home: got %q", fixtures[0].Home.Name)
	}
}
