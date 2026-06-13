// Package notify sends messages to the pool's Telegram group. It is a thin,
// dependency-free wrapper over the Bot API sendMessage method, configured from
// the environment (TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID). When either is unset
// the notifier is disabled and Send becomes a no-op, so the app runs fine
// locally without Telegram credentials.
package notify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Telegram posts HTML messages to a fixed chat.
type Telegram struct {
	token  string
	chatID string
	client *http.Client
}

// TelegramFromEnv builds a notifier from TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID.
// Returns (nil, nil) when either is unset — callers treat a nil notifier as
// "disabled" and skip sending.
func TelegramFromEnv() (*Telegram, error) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	chatID := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))
	if token == "" || chatID == "" {
		return nil, nil
	}
	return &Telegram{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Enabled reports whether the notifier is configured. A nil receiver is safe and
// reports false, so callers can write `if n.Enabled()` without a nil check.
func (t *Telegram) Enabled() bool { return t != nil && t.token != "" && t.chatID != "" }

// Send posts an HTML-formatted message to the configured chat. Web-page preview
// is disabled to keep result posts compact. A no-op on a disabled notifier.
func (t *Telegram) Send(ctx context.Context, html string) error {
	if !t.Enabled() {
		return nil
	}
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	form := url.Values{}
	form.Set("chat_id", t.chatID)
	form.Set("text", html)
	form.Set("parse_mode", "HTML")
	form.Set("disable_web_page_preview", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("telegram send: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
