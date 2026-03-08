package aichat

import "time"

// ChatHTMLRequest represents the request structure for chat-html endpoint
type ChatHTMLRequest struct {
	ShopID       string `json:"shop_id" validate:"required"`
	Question     string `json:"question" validate:"required"`
	FunctionName string `json:"function_name" validate:"required"` // "product" or "customer"
}

// ChatHTMLResponse represents the response structure
type ChatHTMLResponse struct {
	Success            bool              `json:"success"`
	Message            string            `json:"message"`
	Data               *ChatResponseData `json:"data,omitempty"`
	Error              string            `json:"error,omitempty"`
	Cached             bool              `json:"cached"`
	Timestamp          time.Time         `json:"timestamp"`
	TokenUsage         *TokenUsage       `json:"token_usage,omitempty"`
	SuggestedQuestions []string          `json:"suggested_questions,omitempty"`
}

// ChatResponseData contains the actual answer
type ChatResponseData struct {
	Answer string `json:"answer"`         // Plain text answer
	HTML   string `json:"html,omitempty"` // HTML formatted answer (optional)
}

// TokenUsage represents token consumption and cost
type TokenUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	CostTHB          float64 `json:"cost_thb"`
	Model            string  `json:"model"`
}

// StockData represents product/customer data information
type StockData struct {
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	BarcodeList string `json:"barcode_list"`
	UnitStruct  string `json:"unit_structure"`
	StockQty    string `json:"stock_qty"`
}

// PromptCache stores cached prompt data
type PromptCache struct {
	Hash      string
	StockData []StockData
	Timestamp time.Time
}

const (
	CacheDuration = 15 * time.Minute // Cache expires after 15 minutes

	// Exchange rate (approximate)
	USDToTHB = 35.50 // 1 USD = 35.50 THB

	// Minimax M2 pricing (free tier has $0 cost)
	MinimaxM2InputCostPer1M  = 0.00 // $0 per 1M input tokens (free)
	MinimaxM2OutputCostPer1M = 0.00 // $0 per 1M output tokens (free)
)
