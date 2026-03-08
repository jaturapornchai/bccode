package mythaitokenizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// ThaiTokenizerServiceURL - URL ของ Thai Tokenizer service
	ThaiTokenizerServiceURL = "http://thai-tokenizer:8000"
)

// TokenizeRequest - request สำหรับ tokenize endpoint
type TokenizeRequest struct {
	Texts  []string `json:"texts"`
	Engine string   `json:"engine"` // "newmm", "attacut", "deepcut"
}

// TokenizeResponse - response จาก tokenize endpoint
type TokenizeResponse struct {
	Tokenized []string `json:"tokenized"`
	Count     int      `json:"count"`
}

// Client - Thai Tokenizer client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient - สร้าง Thai Tokenizer client ใหม่
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: ThaiTokenizerServiceURL,
	}
}

// NewClientWithURL - สร้าง client พร้อม custom URL
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Tokenize - แบ่งคำภาษาไทย
func (c *Client) Tokenize(texts []string, engine string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	// Default engine
	if engine == "" {
		engine = "newmm"
	}

	// สร้าง request
	reqBody := TokenizeRequest{
		Texts:  texts,
		Engine: engine,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// เรียก API
	url := fmt.Sprintf("%s/tokenize", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tokenizer service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tokenizer service returned error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenizeResp TokenizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenizeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tokenizeResp.Tokenized, nil
}

// TokenizeSingle - แบ่งคำภาษาไทยสำหรับข้อความเดียว (helper function)
func (c *Client) TokenizeSingle(text string, engine string) (string, error) {
	if text == "" {
		return "", nil
	}

	results, err := c.Tokenize([]string{text}, engine)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no tokenization result returned")
	}

	return results[0], nil
}

// HealthCheck - ตรวจสอบว่า service ทำงานปกติหรือไม่
func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
