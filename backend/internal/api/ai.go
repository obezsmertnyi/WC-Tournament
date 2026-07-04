package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/gemini"
)

const (
	aiMaxHistoryTurns = 12
	aiTimeout         = 30 * time.Second
)

// AIAssistant is the capability the AI handlers depend on (satisfied by
// *gemini.Assistant; stubbed in tests).
type AIAssistant interface {
	Available() bool
	StreamChat(ctx context.Context, history []gemini.Turn, msg string, onToken func(string) error) error
	Card(ctx context.Context, query string) (*gemini.Card, bool, error)
}

// perUserLimiter is a per-authenticated-user token bucket for the AI endpoints.
type perUserLimiter struct {
	mu sync.Mutex
	m  map[int64]*rate.Limiter
	r  rate.Limit
	b  int
}

func newPerUserLimiter(perMinute, burst int) *perUserLimiter {
	return &perUserLimiter{m: map[int64]*rate.Limiter{}, r: rate.Every(time.Minute / time.Duration(perMinute)), b: burst}
}

func (l *perUserLimiter) allow(uid int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.m[uid]
	if !ok {
		lim = rate.NewLimiter(l.r, l.b)
		l.m[uid] = lim
	}
	return lim.Allow()
}

var wikiClient = &http.Client{Timeout: 4 * time.Second}

// wikiImage best-effort fetches a photo/crest thumbnail for a football entity from
// Wikipedia REST (keyless, no API key). Tries Ukrainian then English; returns "" on
// any miss so the card renders fine without an image.
func wikiImage(ctx context.Context, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	for _, lang := range []string{"uk", "en"} {
		u := "https://" + lang + ".wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(name)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "WC-Tournament/1.0 (friends prediction pool)")
		resp, err := wikiClient.Do(req)
		if err != nil {
			continue
		}
		var d struct {
			Thumbnail struct {
				Source string `json:"source"`
			} `json:"thumbnail"`
		}
		good := resp.StatusCode == http.StatusOK && json.NewDecoder(resp.Body).Decode(&d) == nil
		resp.Body.Close()
		if good && d.Thumbnail.Source != "" && !isPlaceholderThumb(d.Thumbnail.Source) {
			return d.Thumbnail.Source
		}
	}
	return ""
}

// isPlaceholderThumb rejects Wikipedia infobox kit-template images
// (Kit_body/Kit_shorts/Kit_socks/Kit_left_arm/…) that the summary API returns as
// the lead thumbnail for many national-team and club articles — they render as a
// meaningless kit diagram, worse than showing no image at all.
func isPlaceholderThumb(u string) bool {
	return strings.Contains(strings.ToLower(u), "/kit_")
}

// RegisterAIRoutes wires the football AI assistant (ADR-0017). Mount on the
// RequireUser group so it's auth-only; it is deliberately NOT in DemoGate's
// routeAccess, so every logged-in user (incl. the `none` demo tier) can use it.
func RegisterAIRoutes(r gin.IRouter, a AIAssistant) {
	lim := newPerUserLimiter(20, 5) // ~20 req/min/user, burst 5
	grp := r.Group("/api/ai")
	grp.GET("/status", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"available": a.Available()}) })
	grp.POST("/chat", aiChatHandler(a, lim))
	grp.POST("/card", aiCardHandler(a, lim))
}

func aiChatHandler(a AIAssistant, lim *perUserLimiter) gin.HandlerFunc {
	type req struct {
		Message string        `json:"message"`
		History []gemini.Turn `json:"history"`
	}
	return func(c *gin.Context) {
		if !a.Available() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai_unavailable"})
			return
		}
		claims, _ := auth.Current(c)
		if !lim.allow(claims.Sub) {
			c.Header("Retry-After", "5")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited"})
			return
		}
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		// Pre-stream input hygiene via the SAME validator StreamChat uses, so an
		// invalid message returns a clean 400 BEFORE we open the SSE stream (never
		// mid-stream as an opaque error frame).
		if _, err := gemini.Sanitize(body.Message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message is empty or too long"})
			return
		}
		history := body.History
		if len(history) > aiMaxHistoryTurns {
			history = history[len(history)-aiMaxHistoryTurns:]
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), aiTimeout)
		defer cancel()
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no") // disable nginx proxy buffering
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "stream unsupported"})
			return
		}
		// The shared server has a 10s WriteTimeout, which would truncate a normal
		// streamed answer mid-flight. Extend the write deadline for THIS response
		// only (to the AI context budget + slack); other endpoints keep the 10s cap.
		_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(aiTimeout + 5*time.Second))

		err := a.StreamChat(ctx, history, body.Message, func(tok string) error {
			b, _ := json.Marshal(tok)
			if _, werr := fmt.Fprintf(c.Writer, "data: %s\n\n", b); werr != nil {
				return werr
			}
			flusher.Flush()
			return nil
		})
		if err != nil {
			// Log the real cause (server-side only); surface a generic SSE error to
			// the client (never leak internals).
			slog.Warn("ai chat failed", slog.Any("error", err))
			fmt.Fprint(c.Writer, "event: error\ndata: \"ai_error\"\n\n")
			flusher.Flush()
			return
		}
		fmt.Fprint(c.Writer, "event: done\ndata: {}\n\n")
		flusher.Flush()
	}
}

func aiCardHandler(a AIAssistant, lim *perUserLimiter) gin.HandlerFunc {
	type req struct {
		Query string `json:"query"`
	}
	return func(c *gin.Context) {
		if !a.Available() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai_unavailable"})
			return
		}
		claims, _ := auth.Current(c)
		if !lim.allow(claims.Sub) {
			c.Header("Retry-After", "5")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited"})
			return
		}
		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		// Card generation (grounded Google-Search + Wikipedia enrich) routinely
		// takes 10-20s, well past the server's 10s WriteTimeout — which resets the
		// connection before c.JSON writes the body, so the client's fetch fails
		// even though the handler completes and the server logs 200 (the card
		// never renders). Extend the write deadline for THIS response only to the
		// AI budget + slack, mirroring the chat handler.
		_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(aiTimeout + 10*time.Second))

		ctx, cancel := context.WithTimeout(c.Request.Context(), aiTimeout)
		defer cancel()

		card, ok, err := a.Card(ctx, body.Query)
		switch {
		case errors.Is(err, gemini.ErrInputInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"error": "query is empty or too long"})
		case errors.Is(err, gemini.ErrUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai_unavailable"})
		case err != nil:
			slog.Warn("ai card failed", slog.Any("error", err))
			c.JSON(http.StatusBadGateway, gin.H{"error": "ai_error"})
		case !ok:
			// Off-topic or failed validation — fail closed with the canned refusal,
			// localized to the query language (UA/EN).
			c.JSON(http.StatusOK, gin.H{"refused": true, "message": gemini.RefusalFor(body.Query)})
		case strings.TrimSpace(card.Clarify) != "":
			// Ambiguous name (e.g. "Роналду"): ask which entity instead of guessing.
			c.JSON(http.StatusOK, gin.H{"clarify": true, "message": card.Clarify})
		default:
			// Best-effort enrich with a Wikipedia photo/crest (keyless); never blocks.
			// Prefer the model's disambiguated wiki_title (e.g. "Кріштіану Роналду")
			// over the possibly-ambiguous display name ("Роналду" → Brazilian R9).
			if card.ImageURL == "" {
				q := strings.TrimSpace(card.WikiTitle)
				if q == "" {
					q = card.Name
				}
				card.ImageURL = wikiImage(ctx, q)
			}
			c.JSON(http.StatusOK, card)
		}
	}
}
