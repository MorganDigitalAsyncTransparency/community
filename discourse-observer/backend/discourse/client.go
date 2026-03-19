// Spec: specs/discourse/discourse-source-model.md
// Tests: backend/pipeline_test.go
package discourse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/code-community/discourse-observer/backend/model"
)

// Client fetches raw data from a Discourse-compatible HTTP API.
type Client struct {
	baseURL     string
	apiKey      string
	apiUsername string
	http        *http.Client
}

// NewClient creates a Discourse API client. apiKey and apiUsername may be
// empty for unauthenticated access (e.g. against a mock server).
func NewClient(baseURL, apiKey, apiUsername string) *Client {
	return &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		apiUsername: apiUsername,
		http:        &http.Client{},
	}
}

// latestResponse mirrors the Discourse /latest.json shape.
type latestResponse struct {
	TopicList struct {
		Topics []model.RawTopic `json:"topics"`
	} `json:"topic_list"`
}

// categoriesResponse mirrors the Discourse /categories.json shape.
type categoriesResponse struct {
	CategoryList struct {
		Categories []model.RawCategory `json:"categories"`
	} `json:"category_list"`
}

// FetchTopics retrieves all topics from the /latest.json endpoint.
func (c *Client) FetchTopics(ctx context.Context) ([]model.RawTopic, error) {
	var resp latestResponse
	if err := c.getJSON(ctx, "/latest.json", &resp); err != nil {
		return nil, fmt.Errorf("fetch topics: %w", err)
	}
	return resp.TopicList.Topics, nil
}

// FetchCategories retrieves all categories from the /categories.json endpoint.
func (c *Client) FetchCategories(ctx context.Context) ([]model.RawCategory, error) {
	var resp categoriesResponse
	if err := c.getJSON(ctx, "/categories.json", &resp); err != nil {
		return nil, fmt.Errorf("fetch categories: %w", err)
	}
	return resp.CategoryList.Categories, nil
}

func (c *Client) getJSON(ctx context.Context, path string, dst any) error {
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
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
