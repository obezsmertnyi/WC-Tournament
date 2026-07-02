package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/gemini"
)

// slowCardAssistant models the real assistant: Card() blocks for `delay`
// (grounded Google-Search + enrich routinely takes 10-20s in prod) then returns
// a valid card. ImageURL is pre-set so the handler's Wikipedia enrich is skipped
// (keeps the test hermetic — no outbound network).
type slowCardAssistant struct{ delay time.Duration }

func (slowCardAssistant) Available() bool { return true }
func (slowCardAssistant) StreamChat(context.Context, []gemini.Turn, string, func(string) error) error {
	return nil
}
func (s slowCardAssistant) Card(ctx context.Context, _ string) (*gemini.Card, bool, error) {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
	return &gemini.Card{Name: "Мессі", Summary: "тест", Confidence: "high", ImageURL: "https://example.test/x.png"}, true, nil
}

// serveReal runs handler on a real TCP listener with the given WriteTimeout and
// returns the base URL + cleanup. Unlike httptest.NewRecorder (which bypasses
// the socket), this actually enforces WriteTimeout — required to reproduce the
// bug where a slow response is reset before its body reaches the client.
func serveReal(t *testing.T, handler http.Handler, wt time.Duration) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: handler, WriteTimeout: wt}
	go func() { _ = srv.Serve(ln) }()
	return "http://" + ln.Addr().String(), func() { _ = srv.Close() }
}

// TestCardHandlerBeatsWriteTimeout proves — with a REAL end-to-end HTTP call —
// that /api/ai/card delivers its full JSON body to the client even when card
// generation (600ms here) outlasts the server's WriteTimeout (200ms). Without
// the handler's SetWriteDeadline extension the connection is reset mid-write and
// the client read fails, which is the prod "cards never render" bug: the server
// logs 200 but the browser's fetch throws and the UI shows the generic error.
// The control sub-test proves the same server truly truncates a slow write, so
// the main assertion is meaningful and would fail if the fix regressed.
func TestCardHandlerBeatsWriteTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Control: a slow handler WITHOUT the write-deadline extension must be cut
	// off by the 200ms WriteTimeout — confirms the test server enforces it.
	t.Run("control_writetimeout_truncates_slow_write", func(t *testing.T) {
		r := gin.New()
		r.GET("/slow", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"name": "Мессі"})
		})
		base, stop := serveReal(t, r, 200*time.Millisecond)
		defer stop()

		resp, err := http.Get(base + "/slow")
		if err != nil {
			return // request-level failure — the write was truncated, as expected
		}
		_, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if rerr == nil {
			t.Fatal("expected the 200ms WriteTimeout to truncate the 500ms write, but got a clean body")
		}
	})

	// The real card handler WITH the fix must deliver the full card body despite
	// generation (600ms) exceeding the 200ms WriteTimeout.
	t.Run("card_handler_delivers_full_body", func(t *testing.T) {
		r := gin.New()
		RegisterAIRoutes(r, slowCardAssistant{delay: 600 * time.Millisecond})
		base, stop := serveReal(t, r, 200*time.Millisecond)
		defer stop()

		resp, err := http.Post(base+"/api/ai/card", "application/json",
			bytes.NewReader([]byte(`{"query":"Мессі"}`)))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body failed (write deadline not extended — connection reset): %v", err)
		}
		var card gemini.Card
		if err := json.Unmarshal(raw, &card); err != nil {
			t.Fatalf("unmarshal card: %v (body=%q)", err, raw)
		}
		if card.Name != "Мессі" {
			t.Fatalf("card.Name = %q, want Мессі", card.Name)
		}
	})
}
