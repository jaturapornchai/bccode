package aichat

import "time"

// ChatHTMLRequest represents the request structure for chat-html endpoint
type ChatHTMLRequest struct {
	HoldingCode  string `json:"holdingcode" validate:"required"`
	Question     string `json:"question" validate:"required"`
	FunctionName string `json:"functionname" validate:"required"` // "product" or "customer"
}

// ChatHTMLResponse represents the response structure
type ChatHTMLResponse struct {
	Success            bool              `json:"success"`
	Message            string            `json:"message"`
	Data               *ChatResponseData `json:"data,omitempty"`
	Error              string            `json:"error,omitempty"`
	Cached             bool              `json:"cached"`
	Timestamp          time.Time         `json:"timestamp"`
	TokenUsage         *TokenUsage       `json:"tokenusage,omitempty"`
	SuggestedQuestions []string          `json:"suggestedquestions,omitempty"`
}

// ChatResponseData contains the actual answer
type ChatResponseData struct {
	Answer string `json:"answer"`         // Plain text answer
	HTML   string `json:"html,omitempty"` // HTML formatted answer (optional)
}

// TokenUsage represents token consumption and cost
type TokenUsage struct {
	PromptTokens     int     `json:"prompttokens"`
	CompletionTokens int     `json:"completiontokens"`
	TotalTokens      int     `json:"totaltokens"`
	CostUSD          float64 `json:"costusd"`
	CostTHB          float64 `json:"costthb"`
	Model            string  `json:"model"`
	HasThinking      bool    `json:"hasthinking"`
}

// StockData represents product/customer data information
type StockData struct {
	ProductCode string `json:"productcode"`
	ProductName string `json:"productname"`
	BarcodeList string `json:"barcodelist"`
	UnitStruct  string `json:"unitstructure"`
	StockQty    string `json:"stockqty"`
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
