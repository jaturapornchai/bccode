package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"io"
	"net/http"
	"os"
	"time"
)

// GeminiClient handles communication with Google Gemini API
type GeminiClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []Content          `json:"contents"`
	GenerationConfig GenerationConfig   `json:"generation_config,omitempty"`
	SystemInstruction *SystemInstruction `json:"system_instruction,omitempty"`
	CachedContent string             `json:"cached_content,omitempty"`
}

// Content represents message content
type Content struct {
	Role string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part represents a message part
type Part struct {
	Text string `json:"text"`
}

// SystemInstruction represents system-level instructions
type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

// GenerationConfig represents generation parameters
type GenerationConfig struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP float64 `json:"top_p,omitempty"`
	TopK int     `json:"top_k,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usage_metadata,omitempty"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content Content        `json:"content"`
	FinishReason string         `json:"finish_reason"`
	SafetyRatings []SafetyRating `json:"safety_ratings"`
}

// SafetyRating represents content safety rating
type SafetyRating struct {
	Category string `json:"category"`
	Probability string `json:"probability"`
}

// UsageMetadata contains token usage information
type UsageMetadata struct {
	PromptTokenCount int `json:"prompt_token_count"`
	CandidatesTokenCount int `json:"candidates_token_count"`
	TotalTokenCount int `json:"total_token_count"`
	CachedContentTokenCount int `json:"cached_content_token_count,omitempty"`
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient() *GeminiClient {
	apiKey := os.Getenv("GEMINI_API_KEY")
	model := os.Getenv("GEMINI_MODEL")

	if model == "" {
		model = "gemini-2.0-flash-exp"
	}

	return &GeminiClient{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateContent sends a request to Gemini API
func (c *GeminiClient) GenerateContent(ctx context.Context, request GeminiRequest) (*GeminiResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured")
	}

	// Build API URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.Model, c.APIKey)

	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Info("[Gemini] Sending request to: %s", c.Model)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Error("[Gemini] API error: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Unmarshal response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Log usage
	if geminiResp.UsageMetadata.TotalTokenCount > 0 {
		logger.Info("[Gemini] Tokens used - Prompt: %d, Response: %d, Total: %d, Cached: %d",
			geminiResp.UsageMetadata.PromptTokenCount,
			geminiResp.UsageMetadata.CandidatesTokenCount,
			geminiResp.UsageMetadata.TotalTokenCount,
			geminiResp.UsageMetadata.CachedContentTokenCount,
		)
	}

	return &geminiResp, nil
}

// ExtractTextFromResponse extracts text from the first candidate
func ExtractTextFromResponse(resp *GeminiResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return ""
	}

	return candidate.Content.Parts[0].Text
}
