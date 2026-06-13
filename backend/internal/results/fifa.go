package results

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const (
	fifaBaseURL       = "https://api.fifa.com/api/v3"
	fifaCompetitionID = "17"
	fifaSeasonID      = "285023"
	fifaPageCount     = 200
	fifaUserAgent     = "Mozilla/5.0"
	fifaHTTPTimeout   = 15 * time.Second
	// politeDelay throttles successive page requests to be a good citizen
	// against the Akamai-fronted, 5-min edge-cached API.
	politeDelay = 400 * time.Millisecond
	// maxPages caps pagination defensively against a non-terminating token.
	maxPages = 50
)

// FIFAClient is a ResultsProvider backed by the official FIFA API v3.
type FIFAClient struct {
	httpClient *http.Client
	baseURL    string
	now        func() time.Time
}

// FIFAOption configures a FIFAClient.
type FIFAOption func(*FIFAClient)

// WithBaseURL overrides the API base URL (used in tests).
func WithBaseURL(u string) FIFAOption { return func(c *FIFAClient) { c.baseURL = u } }

// WithHTTPClient injects a custom *http.Client (used in tests).
func WithHTTPClient(h *http.Client) FIFAOption { return func(c *FIFAClient) { c.httpClient = h } }

// WithClock injects a deterministic clock (used in tests).
func WithClock(now func() time.Time) FIFAOption { return func(c *FIFAClient) { c.now = now } }

// NewFIFAClient constructs a FIFA-backed provider with a shared *http.Client
// (cookie jar, 15s timeout) per the architecture notes.
func NewFIFAClient(opts ...FIFAOption) *FIFAClient {
	jar, _ := cookiejar.New(nil)
	c := &FIFAClient{
		httpClient: &http.Client{Timeout: fifaHTTPTimeout, Jar: jar},
		baseURL:    fifaBaseURL,
		now:        time.Now,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Fixtures walks the paginated calendar endpoint via ContinuationToken and
// returns all fixtures normalized to the source-agnostic Fixture type.
func (c *FIFAClient) Fixtures(ctx context.Context) ([]Fixture, error) {
	var all []Fixture
	var token *string

	for page := 0; page < maxPages; page++ {
		resp, err := c.fetchCalendarPage(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("fetch calendar page %d: %w", page, err)
		}

		all = append(all, parseCalendar(*resp, c.now().UTC())...)

		if resp.ContinuationToken == nil || *resp.ContinuationToken == "" {
			break
		}
		token = resp.ContinuationToken

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(politeDelay):
		}
	}

	return all, nil
}

func (c *FIFAClient) fetchCalendarPage(ctx context.Context, token *string) (*fifaCalendarResponse, error) {
	q := url.Values{}
	q.Set("idCompetition", fifaCompetitionID)
	q.Set("idSeason", fifaSeasonID)
	q.Set("language", "en")
	q.Set("count", fmt.Sprintf("%d", fifaPageCount))
	if token != nil && *token != "" {
		q.Set("continuationToken", *token)
	}

	endpoint := c.baseURL + "/calendar/matches?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", fifaUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out fifaCalendarResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
