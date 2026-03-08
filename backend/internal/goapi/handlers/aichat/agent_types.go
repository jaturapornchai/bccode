package aichat

import "time"

// AgentChatRequest — request สำหรับ agentic chatbot
type AgentChatRequest struct {
	ShopID   string `json:"shop_id" validate:"required"`
	Question string `json:"question" validate:"required"`
}

// AgentChatResponse — response จาก agent loop
type AgentChatResponse struct {
	Success            bool               `json:"success"`
	Message            string             `json:"message"`
	Data               *AgentResponseData `json:"data,omitempty"`
	Error              string             `json:"error,omitempty"`
	Timestamp          time.Time          `json:"timestamp"`
	TokenUsage         *TokenUsage        `json:"token_usage,omitempty"`
	SuggestedQuestions []string           `json:"suggested_questions,omitempty"`
}

// AgentResponseData — ข้อมูลคำตอบจาก agent
type AgentResponseData struct {
	Answer     string          `json:"answer"`
	HTML       string          `json:"html,omitempty"`
	ToolsUsed  []ToolExecution `json:"tools_used"`
	Iterations int             `json:"iterations"`
}

// ToolExecution — log ของ tool ที่ถูกเรียก
type ToolExecution struct {
	Tool       string      `json:"tool"`
	Params     interface{} `json:"params"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
	DurationMs int64       `json:"duration_ms"`
}
