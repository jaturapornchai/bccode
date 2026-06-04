package aichat

import "time"

// AgentChatRequest — request สำหรับ agentic chatbot
type AgentChatRequest struct {
	HoldingCode string `json:"holding_code" validate:"required"`
	Question    string `json:"question" validate:"required"`
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
	Thinking   string          `json:"thinking,omitempty"`
	HTML       string          `json:"html,omitempty"`
	ToolsUsed  []ToolExecution `json:"tools_used"`
	Iterations int             `json:"iterations"`
	Citations  []Citation      `json:"citations,omitempty"`
}

// Citation — แหล่งอ้างอิงที่ frontend แสดงเป็นปุ่มได้
// Type: "kb" = KB document (มี text อ่านได้เลย), "web" = web search result (มี url ให้กดเปิด)
type Citation struct {
	Type  string `json:"type"`             // "kb" or "web"
	Label string `json:"label"`            // doc name or page title
	URL   string `json:"url,omitempty"`    // web URL (type=web)
	Text  string `json:"text,omitempty"`   // full text content (type=kb)
	DocID string `json:"doc_id,omitempty"` // KB document id (type=kb)
}

// ToolExecution — log ของ tool ที่ถูกเรียก
type ToolExecution struct {
	Tool       string      `json:"tool"`
	Params     interface{} `json:"params"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
	DurationMs int64       `json:"duration_ms"`
	Source     string      `json:"source,omitempty"` // "mcp" or "web_search"
}
