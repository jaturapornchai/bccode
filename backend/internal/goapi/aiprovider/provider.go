package aiprovider

import (
	"context"
	"fmt"
)

// AIProvider interface สำหรับทุก AI provider
type AIProvider interface {
	// GenerateContent ส่ง prompt ไป AI แล้ว return text response
	GenerateContent(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	// Name ชื่อ provider (สำหรับ log)
	Name() string
}

// ChatRequest — unified request format สำหรับทุก provider
type ChatRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	TopP         float64
	MaxTokens    int
}

// ChatResponse — unified response format
type ChatResponse struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Model            string
}

// GetProvider returns auto-fallback provider
// ลอง provider ที่มี API key ทีละตัวตาม priority
// ถ้าตัวไหน fail → cooldown 1 ชม. แล้วข้ามไปตัวถัดไป
func GetProvider() AIProvider {
	return &fallbackProvider{}
}

// ToolCallingProvider — provider ที่รองรับ OpenAI function calling
type ToolCallingProvider interface {
	GenerateContentWithTools(ctx context.Context, messages []OAIMessage, tools []OAITool, temperature float64) (*OAIResponse, error)
	Name() string
}

// GetToolCallingProvider returns provider ตัวแรกที่มี API key + รองรับ tool calling
func GetToolCallingProvider() (ToolCallingProvider, error) {
	providers := getAvailableProviders()
	for _, p := range providers {
		if tc, ok := p.provider.(ToolCallingProvider); ok {
			return tc, nil
		}
	}
	return nil, fmt.Errorf("ไม่มี AI Provider ที่รองรับ tool calling")
}

// GetAllToolCallingProviders returns ทุก provider ที่รองรับ tool calling (สำหรับ fallback)
func GetAllToolCallingProviders() []ToolCallingProvider {
	providers := getAvailableProviders()
	var result []ToolCallingProvider
	for _, p := range providers {
		if tc, ok := p.provider.(ToolCallingProvider); ok {
			result = append(result, tc)
		}
	}
	return result
}
