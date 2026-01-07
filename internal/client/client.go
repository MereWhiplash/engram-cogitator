// internal/client/client.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/MereWhiplash/engram-cogitator/internal/apitypes"
	"github.com/MereWhiplash/engram-cogitator/internal/gitinfo"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// Client is an HTTP client for the central API
type Client struct {
	baseURL string
	gitInfo *gitinfo.Info
	http    *http.Client
}

// New creates a new API client
func New(baseURL string, gitInfo *gitinfo.Info) *Client {
	return &Client{
		baseURL: baseURL,
		gitInfo: gitInfo,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add git context headers
	if c.gitInfo != nil {
		if c.gitInfo.AuthorName != "" {
			req.Header.Set("X-EC-Author-Name", c.gitInfo.AuthorName)
		}
		if c.gitInfo.AuthorEmail != "" {
			req.Header.Set("X-EC-Author-Email", c.gitInfo.AuthorEmail)
		}
		if c.gitInfo.Repo != "" {
			req.Header.Set("X-EC-Repo", c.gitInfo.Repo)
		}
	}

	return c.http.Do(req)
}

// parseErrorResponse extracts error message from API response
func parseErrorResponse(resp *http.Response) error {
	var errResp apitypes.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil || errResp.Error == "" {
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}
	return fmt.Errorf("API error: %s", errResp.Error)
}

// Add creates a new memory
func (c *Client) Add(ctx context.Context, memType, area, content, rationale string) (*types.Memory, error) {
	req := apitypes.AddRequest{
		Type:      memType,
		Area:      area,
		Content:   content,
		Rationale: rationale,
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/memories", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseErrorResponse(resp)
	}

	var result apitypes.AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Memory, nil
}

// Search finds memories by query
func (c *Client) Search(ctx context.Context, query string, limit int, memType, area string) ([]types.Memory, error) {
	req := apitypes.SearchRequest{
		Query: query,
		Limit: limit,
		Type:  memType,
		Area:  area,
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/memories/search", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var result apitypes.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// List returns recent memories
func (c *Client) List(ctx context.Context, limit int, memType, area string, includeInvalid bool) ([]types.Memory, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if memType != "" {
		params.Set("type", memType)
	}
	if area != "" {
		params.Set("area", area)
	}
	if includeInvalid {
		params.Set("include_invalid", "true")
	}
	path := "/v1/memories?" + params.Encode()

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var result apitypes.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// Invalidate marks a memory as invalid
func (c *Client) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	req := apitypes.InvalidateRequest{
		SupersededBy: supersededBy,
	}

	path := fmt.Sprintf("/v1/memories/%d/invalidate", id)
	resp, err := c.doRequest(ctx, "PUT", path, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp)
	}

	return nil
}
