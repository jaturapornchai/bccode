package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"
)

// openAICompatProvider — shared client สำหรับ OpenRouter / Groq / DeepSeek / custom proxy
// ทุกตัวใช้ OpenAI-compatible chat completions API format เหมือนกัน
type openAICompatProvider struct {
	name       string
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func newOpenAICompatProvider(name, baseURL, apiKey, model string) AIProvider {
	// Per-provider timeout:
	//   - ollama (local, port 11434) → 300s (local GPU/CPU models ช้า)
	//   - cloud (openrouter, groq, deepseek, custom/bcproxy) → 60s
	//     ถ้าช้ากว่านี้ = upstream overload, retry/fallback เร็วดีกว่ารอ
	timeout := 60 * time.Second
	lname := strings.ToLower(name)
	lurl := strings.ToLower(baseURL)
	if lname == "ollama" || strings.Contains(lurl, ":11434") {
		timeout = 300 * time.Second
	}
	return &openAICompatProvider{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (o *openAICompatProvider) Name() string {
	return o.name
}

// OpenAI-compatible request/response structs

// OAIMessage — chat message (supports tool calls)
type OAIMessage struct {
	Role       string        `json:"role"`
	Content    interface{}   `json:"content"`             // string or []ContentPart
	Reasoning  string        `json:"reasoning,omitempty"` // Ollama thinking models (gemma4, etc.)
	ToolCalls  []OAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

// ContentPart — multimodal content (text or image)
type ContentPart struct {
	Type     string    `json:"type"` // "text" or "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL — image URL or base64 data
type ImageURL struct {
	URL string `json:"url"` // "data:image/jpeg;base64,..." or URL
}

// GetContentString extracts text content from OAIMessage.Content (string or []ContentPart)
func GetContentString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		for _, part := range v {
			if m, ok := part.(map[string]interface{}); ok {
				if m["type"] == "text" {
					if text, ok := m["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return fmt.Sprintf("%v", content)
}

// OAITool — function tool definition
type OAITool struct {
	Type     string      `json:"type"` // "function"
	Function OAIFunction `json:"function"`
}

// OAIFunction — function metadata
type OAIFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// OAIToolCall — tool call from assistant
type OAIToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"` // "function"
	Function OAIToolCallFunc `json:"function"`
}

// OAIToolCallFunc — function name + arguments
type OAIToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type oaiRequest struct {
	Model       string       `json:"model"`
	Messages    []OAIMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	TopP        float64      `json:"top_p,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Tools       []OAITool    `json:"tools,omitempty"`
	Stream      bool         `json:"stream"`
}

// OAIResponse — full API response
type OAIResponse struct {
	Choices  []OAIChoice `json:"choices"`
	Usage    OAIUsage    `json:"usage"`
	Model    string      `json:"model"`
	Provider string      `json:"provider,omitempty"` // จาก X-BCProxy-Provider header
}

// OAIChoice — single choice
type OAIChoice struct {
	Message OAIMessage `json:"message"`
}

// OAIUsage — token usage
type OAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (o *openAICompatProvider) GenerateContent(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if o.apiKey == "" && !strings.HasPrefix(o.name, "custom") && o.name != "ollama" {
		return nil, fmt.Errorf("%s API key not configured", o.name)
	}

	// สร้าง messages array
	messages := make([]OAIMessage, 0, 2)
	if req.SystemPrompt != "" {
		messages = append(messages, OAIMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, OAIMessage{Role: "user", Content: req.UserPrompt})

	oaiReq := oaiRequest{
		Model:       o.model,
		Messages:    messages,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
	}

	jsonData, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Info("[%s] Sending request to model: %s", o.name, o.model)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("[%s] API error (status %d): %s", o.name, resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var oaiResp OAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response from %s", o.name)
	}

	logger.Info("[%s] Tokens used - Prompt: %d, Completion: %d, Total: %d",
		o.name,
		oaiResp.Usage.PromptTokens,
		oaiResp.Usage.CompletionTokens,
		oaiResp.Usage.TotalTokens,
	)

	return &ChatResponse{
		Text:             GetContentString(oaiResp.Choices[0].Message.Content),
		PromptTokens:     oaiResp.Usage.PromptTokens,
		CompletionTokens: oaiResp.Usage.CompletionTokens,
		TotalTokens:      oaiResp.Usage.TotalTokens,
		Model:            oaiResp.Model,
	}, nil
}

// GenerateContentWithTools — เรียก API พร้อม tool definitions (function calling)
// Return OAIResponse โดยตรงเพราะ agent loop ต้องการ tool_calls field
func (o *openAICompatProvider) GenerateContentWithTools(ctx context.Context, messages []OAIMessage, tools []OAITool, temperature float64) (*OAIResponse, error) {
	if o.apiKey == "" && !strings.HasPrefix(o.name, "custom") && o.name != "ollama" {
		return nil, fmt.Errorf("%s API key not configured", o.name)
	}
	// bcproxyai (custom provider) routes internally — แค่ส่ง model ที่ user ตั้งไว้
	// ถ้า upstream ตัวนั้น overload → bcproxy จะเลือก fallback ให้เอง
	return o.callOnce(ctx, o.model, messages, tools, temperature)
}

// callOnce — ยิง HTTP request ครั้งเดียวด้วย model ที่กำหนด
func (o *openAICompatProvider) callOnce(ctx context.Context, model string, messages []OAIMessage, tools []OAITool, temperature float64) (*OAIResponse, error) {
	oaiReq := oaiRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		Tools:       tools,
	}

	jsonData, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Info("[%s] Sending tool-calling request to model: %s (messages: %d, tools: %d)", o.name, model, len(messages), len(tools))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("[%s] API error (status %d): %s", o.name, resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var oaiResp OAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response from %s", o.name)
	}

	// เก็บ actual model จาก bcproxyai header (ถ้ามี)
	if proxyModel := resp.Header.Get("X-BCProxy-Model"); proxyModel != "" {
		oaiResp.Model = proxyModel
		logger.Info("[%s] BCProxy routed to model: %s", o.name, proxyModel)
	} else {
		oaiResp.Model = model
	}
	if proxyProvider := resp.Header.Get("X-BCProxy-Provider"); proxyProvider != "" {
		oaiResp.Provider = proxyProvider
	}

	logger.Info("[%s] Tool-calling tokens - Prompt: %d, Completion: %d, Total: %d",
		o.name,
		oaiResp.Usage.PromptTokens,
		oaiResp.Usage.CompletionTokens,
		oaiResp.Usage.TotalTokens,
	)

	// Debug: ถ้า completion>0 แต่ parse ได้ content+tool_calls ว่าง → dump raw body
	if oaiResp.Usage.CompletionTokens > 0 {
		msg := oaiResp.Choices[0].Message
		contentStr := GetContentString(msg.Content)
		if contentStr == "" && len(msg.ToolCalls) == 0 && msg.Reasoning == "" {
			logger.Warn("[%s] Empty parse despite %d completion tokens — raw body: %s",
				o.name, oaiResp.Usage.CompletionTokens, string(body))
		}
	}

	return &oaiResp, nil
}
