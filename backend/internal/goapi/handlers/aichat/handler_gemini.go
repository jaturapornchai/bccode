package aichat

import (
	"context"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"github.com/labstack/echo/v4"
)

// ChatGemini handles chatbot requests with 2-step Gemini AI
// POST /api/v1/chatbot/chat-gemini
func ChatGemini(c echo.Context) error {
	var req ChatHTMLRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ChatHTMLResponse{
			Success:   false,
			Message:   "Invalid request format",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	// Validate request
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, ChatHTMLResponse{
			Success:   false,
			Message:   "holdingcode is required",
			Timestamp: time.Now(),
		})
	}

	if req.Question == "" {
		return c.JSON(http.StatusBadRequest, ChatHTMLResponse{
			Success:   false,
			Message:   "question is required",
			Timestamp: time.Now(),
		})
	}

	if req.FunctionName == "" {
		req.FunctionName = "product" // Default to product
	}

	ctx := context.Background()
	startTime := time.Now()

	logger.Info("[ChatGemini] Question: %s (shop: %s, function: %s)", req.Question, req.HoldingCode, req.FunctionName)

	// STEP 1: AI interprets question and generates SQL query
	logger.Info("[ChatGemini] Step 1: Generating query...")
	intent, err := GenerateQueryFromQuestion(ctx, req.Question, req.FunctionName)
	if err != nil {
		logger.Error("Failed to generate query: %v", err)
		return c.JSON(http.StatusInternalServerError, ChatHTMLResponse{
			Success:   false,
			Message:   "Failed to interpret question",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	logger.Info("[ChatGemini] Query Type: %s, Needs Data: %v", intent.QueryType, intent.NeedsData)
	if intent.SQL != "" {
		logger.Info("[ChatGemini] Generated SQL: %s", intent.SQL)
	}

	var queryResult *QueryResult

	// STEP 1.5: Execute query if needed
	if intent.NeedsData && intent.SQL != "" {
		logger.Info("[ChatGemini] Step 1.5: Executing query...")
		limit := intent.Limit
		if limit == 0 {
			limit = 100 // Default limit
		}

		queryResult, err = ExecuteQuery(ctx, req.HoldingCode, intent.SQL, limit)
		if err != nil {
			logger.Error("Failed to execute query: %v", err)
			return c.JSON(http.StatusInternalServerError, ChatHTMLResponse{
				Success:   false,
				Message:   "Failed to query database",
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
		}
		logger.Info("[ChatGemini] Query returned %d rows", queryResult.Count)
	}

	// STEP 2: AI generates natural language answer from data
	logger.Info("[ChatGemini] Step 2: Generating answer...")
	answer, html, suggestedQuestions, usage, err := GenerateAnswerFromData(ctx, req.Question, intent, queryResult)
	if err != nil {
		logger.Error("Failed to generate answer: %v", err)
		return c.JSON(http.StatusInternalServerError, ChatHTMLResponse{
			Success:   false,
			Message:   "Failed to generate response",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	// Calculate token usage
	var promptTokens, completionTokens, totalTokens int
	var modelName string
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
		totalTokens = usage.TotalTokens
		modelName = usage.Model
	}

	// Use default suggested questions if AI didn't provide any
	if len(suggestedQuestions) == 0 {
		suggestedQuestions = []string{
			"มีสินค้ากี่รายการทั้งหมด",
			"สินค้าที่ราคาสูงที่สุด 5 อันดับคืออะไร",
			"มีสินค้าแบรนด์ไหนบ้าง",
		}
	}

	duration := time.Since(startTime)
	logger.Info("[ChatGemini] Completed in %.2fs (tokens: %d)", duration.Seconds(), totalTokens)

	return c.JSON(http.StatusOK, ChatHTMLResponse{
		Success: true,
		Message: "Response generated successfully",
		Data: &ChatResponseData{
			Answer: answer,
			HTML:   html,
		},
		TokenUsage: &TokenUsage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
			CostUSD:          0.00,
			CostTHB:          0.00,
			Model:            modelName,
		},
		SuggestedQuestions: suggestedQuestions,
		Cached:             false,
		Timestamp:          time.Now(),
	})
}
