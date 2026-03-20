// Spec: specs/discourse/discourse-source-model.md
// Tests: backend/discourse/client_test.go, backend/pipeline_test.go
package discourse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// Client fetches raw data from a Discourse-compatible HTTP API.
type Client struct {
	baseURL     string
	apiKey      string
	apiUsername string
	http        *http.Client
	pageCfg     PageConfig
}

// Option configures a Client.
type Option func(*Client)

// WithPageConfig sets the pagination timing configuration used by
// FetchTopicsPages when called through the observer interface (without
// an explicit PageConfig argument).
func WithPageConfig(cfg PageConfig) Option {
	return func(c *Client) { c.pageCfg = cfg }
}

// NewClient creates a Discourse API client. apiKey and apiUsername may be
// empty for unauthenticated access (e.g. against a mock server).
func NewClient(baseURL, apiKey, apiUsername string, opts ...Option) *Client {
	c := &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		apiUsername: apiUsername,
		http:        &http.Client{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// PageConfig controls pagination behavior for paginated fetches.
type PageConfig struct {
	Delay      time.Duration // delay between page requests
	MaxRetries int           // max consecutive retries per page on 5xx/network error
	RetryDelay time.Duration // wait between retries on 5xx/network error
	StartPage  int           // page to start from (for resume)
}

// latestResponse mirrors the Discourse /latest.json shape.
type latestResponse struct {
	TopicList struct {
		Topics        []model.RawTopic `json:"topics"`
		MoreTopicsURL string           `json:"more_topics_url,omitempty"`
	} `json:"topic_list"`
}

// categoriesResponse mirrors the Discourse /categories.json shape.
type categoriesResponse struct {
	CategoryList struct {
		Categories []model.RawCategory `json:"categories"`
	} `json:"category_list"`
}

// FetchTopics retrieves all topics by paginating /latest.json to exhaustion.
// It uses zero delay between pages (suitable for local/mock use).
// For controlled pagination with delays and retries, use FetchTopicsPagesWithConfig.
func (c *Client) FetchTopics(ctx context.Context) ([]model.RawTopic, error) {
	var all []model.RawTopic
	err := c.FetchTopicsPagesWithConfig(ctx, PageConfig{}, func(topics []model.RawTopic, _ int) error {
		all = append(all, topics...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fetch topics: %w", err)
	}
	return all, nil
}

// FetchTopicsPages paginates /latest.json from startPage using the client's
// stored PageConfig. This method satisfies the observer.FetchClient interface.
func (c *Client) FetchTopicsPages(ctx context.Context, startPage int, fn func(topics []model.RawTopic, page int) error) error {
	cfg := c.pageCfg
	cfg.StartPage = startPage
	return c.FetchTopicsPagesWithConfig(ctx, cfg, fn)
}

// FetchTopicsPagesWithConfig paginates /latest.json with explicit config.
// Pagination stops when the response lacks more_topics_url.
// The fn receives the topics from each page and the page number.
// If fn returns a non-nil error, pagination stops and that error is returned.
func (c *Client) FetchTopicsPagesWithConfig(ctx context.Context, cfg PageConfig, fn func(topics []model.RawTopic, page int) error) error {
	page := cfg.StartPage
	for {
		if page > cfg.StartPage {
			if err := sleepCtx(ctx, cfg.Delay); err != nil {
				return err
			}
		}

		path := fmt.Sprintf("/latest.json?page=%d", page)
		var resp latestResponse
		if err := c.getJSONRetry(ctx, path, &resp, cfg.MaxRetries, cfg.RetryDelay); err != nil {
			return fmt.Errorf("fetch page %d: %w", page, err)
		}

		if err := fn(resp.TopicList.Topics, page); err != nil {
			return err
		}

		if resp.TopicList.MoreTopicsURL == "" {
			return nil
		}
		page++
	}
}

// FetchCategories retrieves all categories from the /categories.json endpoint.
func (c *Client) FetchCategories(ctx context.Context) ([]model.RawCategory, error) {
	var resp categoriesResponse
	if err := c.getJSON(ctx, "/categories.json", &resp); err != nil {
		return nil, fmt.Errorf("fetch categories: %w", err)
	}
	return resp.CategoryList.Categories, nil
}

// getJSON performs a GET request and decodes the JSON response.
// It does not retry on errors.
func (c *Client) getJSON(ctx context.Context, path string, dst any) error {
	return c.getJSONRetry(ctx, path, dst, 0, 0)
}

// rateLimitFallback is the minimum wait when a 429 response lacks Retry-After
// and no retryDelay is configured. Prevents tight retry loops.
const rateLimitFallback = 10 * time.Second

// getJSONRetry performs a GET request with retry handling for 429 and 5xx.
func (c *Client) getJSONRetry(ctx context.Context, path string, dst any, maxRetries int, retryDelay time.Duration) error {
	retries := 0
	for {
		err := c.doGetJSON(ctx, path, dst)
		if err == nil {
			return nil
		}

		var httpErr *HTTPError
		if !errors.As(err, &httpErr) {
			// Network error — treat like 5xx.
			retries++
			if retries > maxRetries {
				return fmt.Errorf("after %d retries: %w", retries, err)
			}
			if err := sleepCtx(ctx, retryDelay); err != nil {
				return err
			}
			continue
		}

		switch {
		case httpErr.StatusCode == http.StatusTooManyRequests:
			wait := httpErr.RetryAfter
			if wait <= 0 {
				wait = retryDelay
			}
			if wait <= 0 {
				wait = rateLimitFallback
			}
			if err := sleepCtx(ctx, wait); err != nil {
				return err
			}
			// 429 retries do not count toward maxRetries.
			continue

		case httpErr.StatusCode >= 500:
			retries++
			if retries > maxRetries {
				return fmt.Errorf("after %d retries: %w", retries, err)
			}
			if err := sleepCtx(ctx, retryDelay); err != nil {
				return err
			}
			continue

		default:
			// 4xx (not 429) — not retryable.
			return err
		}
	}
}

// doGetJSON performs a single GET request and decodes JSON.
func (c *Client) doGetJSON(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, http.NoBody)
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		req.Header.Set("Api-Key", c.apiKey)
		req.Header.Set("Api-Username", c.apiUsername)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Drain body to allow connection reuse.
		_, _ = io.Copy(io.Discard, resp.Body)
		return newHTTPError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// HTTPError represents a non-200 HTTP response with optional Retry-After.
type HTTPError struct {
	StatusCode int
	RetryAfter time.Duration
	Path       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("unexpected status %d from %s", e.StatusCode, e.Path)
}

func newHTTPError(resp *http.Response) *HTTPError {
	e := &HTTPError{
		StatusCode: resp.StatusCode,
		Path:       resp.Request.URL.Path,
	}
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if seconds, err := strconv.Atoi(ra); err == nil {
			e.RetryAfter = time.Duration(seconds) * time.Second
		}
	}
	return e
}

// sleepCtx waits for the given duration or until ctx is canceled.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
