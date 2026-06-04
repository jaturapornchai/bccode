package aichat

import (
	"context"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"github.com/labstack/echo/v4"
)

// ChatAgent handles agentic chatbot requests with ReAct loop
// POST /api/v1/chatbot/chat-agent
func ChatAgent(c echo.Context) error {
	var req AgentChatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success:   false,
			Message:   "Invalid request format",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success:   false,
			Message:   "holding_code is required",
			Timestamp: time.Now(),
		})
	}

	if req.Question == "" {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success:   false,
			Message:   "question is required",
			Timestamp: time.Now(),
		})
	}

	logger.Info("[ChatAgent] Question: %s (shop: %s)", req.Question, req.HoldingCode)

	// 120s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := RunAgentLoop(ctx, req.HoldingCode, req.Question)
	if err != nil {
		logger.Error("[ChatAgent] Agent loop failed: %v", err)
		return c.JSON(http.StatusInternalServerError, AgentChatResponse{
			Success:   false,
			Message:   "Agent ทำงานไม่สำเร็จ",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
