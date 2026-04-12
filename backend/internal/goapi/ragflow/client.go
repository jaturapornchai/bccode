package ragflow

// RAGFlow Go client — minimal HTTP wrapper for the RAGFlow REST API.
//
// Docs: https://ragflow.io/docs/dev/category/http-api-reference
//
// Configuration via environment variables:
//   - RAGFLOW_BASE_URL  (default: http://host.docker.internal:9380)
//   - RAGFLOW_API_KEY   (required at runtime — get from RAGFlow Web UI → Profile → API)
//
// The default client is process-wide and lazily initialized via GetClient().
// If RAGFLOW_API_KEY is empty, client.IsConfigured() returns false and tools
// using it should fall back gracefully.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseURL = "http://host.docker.internal:9380"
	defaultTimeout = 60 * time.Second
)

// Client — RAGFlow HTTP client (thread-safe, single shared instance)
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

var (
	defaultClient     *Client
	defaultClientOnce sync.Once
)

// GetClient returns the process-wide RAGFlow client.
// Reads RAGFLOW_BASE_URL + RAGFLOW_API_KEY from env on first call.
func GetClient() *Client {
	defaultClientOnce.Do(func() {
		baseURL := strings.TrimSpace(os.Getenv("RAGFLOW_BASE_URL"))
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		defaultClient = &Client{
			BaseURL: strings.TrimRight(baseURL, "/"),
			APIKey:  strings.TrimSpace(os.Getenv("RAGFLOW_API_KEY")),
			HTTP:    &http.Client{Timeout: defaultTimeout},
		}
	})
	return defaultClient
}

// IsConfigured reports whether an API key is set (cannot call RAGFlow without one)
func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.APIKey) != ""
}

// doJSON sends a JSON request and decodes a JSON response into out.
// Returns an error for non-2xx HTTP status (with response snippet).
func (c *Client) doJSON(method, path string, payload any, out any) error {
	if !c.IsConfigured() {
		return fmt.Errorf("ragflow not configured (RAGFLOW_API_KEY missing)")
	}
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("ragflow unreachable: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := string(respBody)
		if len(snippet) > 800 {
			snippet = snippet[:800] + "...(truncated)"
		}
		return fmt.Errorf("ragflow HTTP %d: %s", resp.StatusCode, snippet)
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w (body: %s)", err, truncate(string(respBody), 300))
	}
	return nil
}

// doMultipart sends a multipart/form-data request (for file uploads) and decodes the response.
func (c *Client) doMultipart(method, path string, contentType string, body io.Reader, out any) error {
	if !c.IsConfigured() {
		return fmt.Errorf("ragflow not configured (RAGFLOW_API_KEY missing)")
	}
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("ragflow unreachable: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := string(respBody)
		if len(snippet) > 800 {
			snippet = snippet[:800] + "...(truncated)"
		}
		return fmt.Errorf("ragflow HTTP %d: %s", resp.StatusCode, snippet)
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w (body: %s)", err, truncate(string(respBody), 300))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
