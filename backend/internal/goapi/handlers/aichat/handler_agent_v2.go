package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// routeAgentLoop — เลือก agent loop ตาม provider
// ถ้า provider แรกเป็น ollama → ใช้ ReAct mode (local models ไม่รองรับ native tool_calls ดี)
// อื่นๆ → ใช้ v2 (OpenAI function calling)
func routeAgentLoop(ctx context.Context, req AgentV2Request, emitSSE func(SSEEvent)) (*AgentChatResponse, error) {
	providers := aiprovider.GetShopToolCallingProviders(req.ShopID)
	if len(providers) > 0 && strings.EqualFold(providers[0].Name(), "ollama") {
		logger.Info("[น้องกุ้ง] Provider=ollama → ใช้ ReAct mode")
		return RunAgentReAct(ctx, req, emitSSE)
	}
	return RunAgentLoopV2(ctx, req, emitSSE)
}

// ChatAgentV2 handles น้องกุ้ง agent requests with SSE streaming
// POST /api/v1/chatbot/chat-agent-v2
func ChatAgentV2(c echo.Context) error {
	var req AgentV2Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success: false, Message: "Invalid request format", Error: err.Error(), Timestamp: time.Now(),
		})
	}

	if req.ShopID == "" || req.Question == "" {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success: false, Message: "shop_id and question are required", Timestamp: time.Now(),
		})
	}

	// Flutter overlay renders via flutter_html → ask AI to emit HTML
	if req.OutputFormat == "" {
		req.OutputFormat = "html"
	}

	logger.Info("[น้องกุ้ง] Question: %s (shop: %s, session: %s, format: %s)", req.Question, req.ShopID, req.SessionID, req.OutputFormat)

	// Check SSE support — ต้องดึง underlying http.Flusher จาก echo.Response.Writer
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		// fallback: sync mode
		logger.Warn("[น้องกุ้ง] SSE not supported (underlying writer), falling back to sync")
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		noopEmit := func(event SSEEvent) {}
		resp, err := routeAgentLoop(ctx, req, noopEmit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AgentChatResponse{
				Success: false, Message: "น้องกุ้งทำงานไม่สำเร็จ", Error: err.Error(), Timestamp: time.Now(),
			})
		}
		return c.JSON(http.StatusOK, resp)
	}

	// SSE supported — setup headers
	rw := c.Response().Writer
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("X-Accel-Buffering", "no")
	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	logger.Info("[น้องกุ้ง] SSE streaming enabled")

	// SSE emit function
	emitSSE := func(event SSEEvent) {
		data, _ := json.Marshal(event)
		fmt.Fprintf(rw, "data: %s\n\n", data)
		flusher.Flush()
	}

	// 180s timeout for v2 (longer because more iterations)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 180*time.Second)
	defer cancel()

	resp, err := routeAgentLoop(ctx, req, emitSSE)
	if err != nil {
		logger.Error("[น้องกุ้ง] Agent loop failed: %v", err)
		emitSSE(SSEEvent{Type: "error", Data: err.Error()})
		emitSSE(SSEEvent{Type: "done", Data: nil})
		return nil
	}

	// Send final complete response
	emitSSE(SSEEvent{Type: "done", Data: resp})
	return nil
}

// ChatAgentV2Sync — non-streaming version (fallback สำหรับ client ที่ไม่รองรับ SSE)
// POST /api/v1/chatbot/chat-agent-v2-sync
func ChatAgentV2Sync(c echo.Context) error {
	var req AgentV2Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success: false, Message: "Invalid request format", Error: err.Error(), Timestamp: time.Now(),
		})
	}

	if req.ShopID == "" || req.Question == "" {
		return c.JSON(http.StatusBadRequest, AgentChatResponse{
			Success: false, Message: "shop_id and question are required", Timestamp: time.Now(),
		})
	}

	if req.OutputFormat == "" {
		req.OutputFormat = "html"
	}

	logger.Info("[น้องกุ้ง-sync] Question: %s (shop: %s, format: %s)", req.Question, req.ShopID, req.OutputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// No-op emitter for sync mode
	noopEmit := func(event SSEEvent) {}

	resp, err := routeAgentLoop(ctx, req, noopEmit)
	if err != nil {
		logger.Error("[น้องกุ้ง-sync] Agent loop failed: %v", err)
		ocErr := classifyAgentError(err)
		body := ocErr.AsResponseBody()
		body["session_key"] = BuildSessionKey(req.ShopID, req.SessionID)
		body["timestamp"] = time.Now()
		return c.JSON(http.StatusInternalServerError, body)
	}

	return c.JSON(http.StatusOK, resp)
}

// ClearChatSession — ลบ session history
// POST /api/v1/chatbot/clear-session
func ClearChatSession(c echo.Context) error {
	var req struct {
		ShopID    string `json:"shop_id"`
		SessionID string `json:"session_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
	}
	if req.ShopID == "" || req.SessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "message": "shop_id and session_id required"})
	}

	if err := clearSession(req.SessionID, req.ShopID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{"success": true, "message": "ล้างประวัติสนทนาแล้ว"})
}
