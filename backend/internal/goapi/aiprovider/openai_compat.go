package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"time"
)

// openAICompatProvider — shared client สำหรับ OpenRouter / Groq / DeepSeek
// ทั้ง 3 ใช้ OpenAI-compatible chat completions API format เหมือนกัน
type openAICompatProvider struct {
	name       string
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func newOpenAICompatProvider(name, baseURL, apiKey, model string) AIProvider {
	return &openAICompatProvider{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
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
	Content    string        `json:"content"`
	ToolCalls  []OAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
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
}

// OAIResponse — full API response
type OAIResponse struct {
	Choices []OAIChoice `json:"choices"`
	Usage   OAIUsage    `json:"usage"`
	Model   string      `json:"model"`
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
	if o.apiKey == "" {
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
		Text:             oaiResp.Choices[0].Message.Content,
		PromptTokens:     oaiResp.Usage.PromptTokens,
		CompletionTokens: oaiResp.Usage.CompletionTokens,
		TotalTokens:      oaiResp.Usage.TotalTokens,
		Model:            oaiResp.Model,
	}, nil
}

// GenerateContentWithTools — เรียก API พร้อม tool definitions (function calling)
// Return OAIResponse โดยตรงเพราะ agent loop ต้องการ tool_calls field
func (o *openAICompatProvider) GenerateContentWithTools(ctx context.Context, messages []OAIMessage, tools []OAITool, temperature float64) (*OAIResponse, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("%s API key not configured", o.name)
	}

	oaiReq := oaiRequest{
		Model:       o.model,
		Messages:    messages,
		Temperature: temperature,
		Tools:       tools,
	}

	jsonData, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Info("[%s] Sending tool-calling request to model: %s (messages: %d, tools: %d)", o.name, o.model, len(messages), len(tools))

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

	logger.Info("[%s] Tool-calling tokens - Prompt: %d, Completion: %d, Total: %d",
		o.name,
		oaiResp.Usage.PromptTokens,
		oaiResp.Usage.CompletionTokens,
		oaiResp.Usage.TotalTokens,
	)

	return &oaiResp, nil
}
