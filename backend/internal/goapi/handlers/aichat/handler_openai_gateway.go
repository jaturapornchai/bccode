package aichat

// OpenAI-compatible gateway สำหรับ OpenClaw (และ client อื่น ๆ ที่พูด OpenAI protocol)
//
// Endpoints:
//
//	POST /goapi/api/aichat/v1/chat/completions  — รับ OpenAI ChatCompletionRequest
//	                                              รัน น้องกุ้ง agent_loop_v2 แล้ว stream
//	                                              ผลลัพธ์กลับในรูป OpenAI SSE format
//	GET  /goapi/api/aichat/v1/models            — list models ที่ gateway นี้ให้บริการ
//
// Holding Code: OpenClaw ไม่รู้จัก holdingcode → ใช้ default จาก bootstrap.json
// (openclaw.default_holdingcode) ถ้าไม่มีจะใช้ constant fallback ด้านล่าง
//
// Session ID: รับจาก query `?session=xxx` หรือ `user` field ของ request
// ถ้าไม่มี generate ใหม่จาก timestamp (no memory cross-request)

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// truncateForLog — ตัดสตริงให้สั้นลงสำหรับ log (รองรับ rune หลาย byte)
func truncateForLog(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "...(truncated)"
}

// osGetenv — wrapper เพื่อให้ test mock ได้ (ตอนนี้ยังไม่ mock ใช้ os.Getenv ตรงๆ)
var osGetenv = os.Getenv

// defaultOpenclawHoldingCode — holdingcode ที่ OpenClaw gateway จะใช้
// (single-tenant dev mode — ถ้าต้องการเปลี่ยน ให้ override ผ่าน env OPENCLAW_HOLDING_CODE)
const defaultOpenclawHoldingCode = "3AEz8tu22GHPpAZ0XhwPFM4fjY9"

// getOpenclawHoldingCode — คืน holdingcode ที่จะใช้สำหรับ OpenClaw gateway
func getOpenclawHoldingCode() string {
	if v := strings.TrimSpace(osGetenv("OPENCLAW_HOLDING_CODE")); v != "" {
		return v
	}
	return defaultOpenclawHoldingCode
}

// ==================== OpenAI Types ====================

type openaiChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string | []contentPart
	Name    string `json:"name,omitempty"`
}

type openaiChatRequest struct {
	Model    string              `json:"model"`
	Messages []openaiChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	User     string              `json:"user,omitempty"` // ใช้เป็น sessionid
}

type openaiChoice struct {
	Index        int                `json:"index"`
	Message      *openaiChatMessage `json:"message,omitempty"`
	Delta        *openaiChatMessage `json:"delta,omitempty"`
	FinishReason *string            `json:"finish_reason"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openaiChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
	Usage   *openaiUsage   `json:"usage,omitempty"`
}

type openaiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type openaiModelList struct {
	Object string        `json:"object"`
	Data   []openaiModel `json:"data"`
}

// ==================== extractLastUserMessage ====================

// messageContentToText แปลง content (string | []part) → plain text
func messageContentToText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if t, _ := m["type"].(string); t == "text" {
					if txt, _ := m["text"].(string); txt != "" {
						parts = append(parts, txt)
					}
				}
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// extractLastUserMessage ดึงข้อความ user สุดท้ายจาก messages
func extractLastUserMessage(messages []openaiChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "user" {
			continue
		}
		return messageContentToText(messages[i].Content)
	}
	return ""
}

// buildQuestionWithHistory สร้าง question ที่มี context ของบทสนทนาก่อนหน้า
// (OpenClaw ส่ง history ทุกครั้ง — เราเอามาใส่เป็น context block ให้ AI จำได้)
//
// Format:
//
//	[ประวัติสนทนา]
//	ผู้ใช้: ...
//	น้องกุ้ง: ...
//	...
//
//	[คำถามล่าสุด]
//	xxx
//
// ถ้าไม่มี history → คืนแค่คำถามล่าสุด
func buildQuestionWithHistory(messages []openaiChatMessage) string {
	// หา index ของ user message สุดท้าย
	lastUserIdx := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserIdx = i
			break
		}
	}
	if lastUserIdx < 0 {
		return ""
	}
	latest := messageContentToText(messages[lastUserIdx].Content)

	// รวม history ก่อนหน้า (ไม่รวม system + ไม่รวม latest)
	const maxHistoryChars = 4000
	var hist []string
	for i := 0; i < lastUserIdx; i++ {
		m := messages[i]
		if m.Role == "system" {
			continue
		}
		txt := strings.TrimSpace(messageContentToText(m.Content))
		if txt == "" {
			continue
		}
		label := "ผู้ใช้"
		if m.Role == "assistant" {
			label = "น้องกุ้ง"
		}
		hist = append(hist, fmt.Sprintf("%s: %s", label, txt))
	}
	if len(hist) == 0 {
		return latest
	}

	histBlock := strings.Join(hist, "\n")
	// ตัดท้ายสุด ๆ ถ้ายาวเกิน (เก็บส่วนล่าสุดไว้)
	if len(histBlock) > maxHistoryChars {
		histBlock = "...(ตัดบางส่วนทิ้ง)...\n" + histBlock[len(histBlock)-maxHistoryChars:]
	}

	return fmt.Sprintf("[ประวัติสนทนาก่อนหน้า]\n%s\n\n[คำถามล่าสุด]\n%s", histBlock, latest)
}

// ==================== Handler: /v1/models ====================

// OpenAIGatewayListModels handles GET /goapi/api/aichat/v1/models
func OpenAIGatewayListModels(c echo.Context) error {
	now := time.Now().Unix()
	return c.JSON(http.StatusOK, openaiModelList{
		Object: "list",
		Data: []openaiModel{
			{ID: "nongkung", Object: "model", Created: now, OwnedBy: "bc-account"},
			{ID: "nongkung-react", Object: "model", Created: now, OwnedBy: "bc-account"},
		},
	})
}

// ==================== Handler: /v1/chat/completions ====================

// OpenAIGatewayChatCompletions handles POST /goapi/api/aichat/v1/chat/completions
//
// รัน agent_loop_v2 (หรือ ReAct) แล้วส่งผลลัพธ์เป็น OpenAI format
// ถ้า stream=true → ส่ง SSE chunks
// ถ้า stream=false → ส่ง JSON เต็ม
func OpenAIGatewayChatCompletions(c echo.Context) error {
	// Log raw body ก่อน bind เพื่อ debug — อ่าน body แล้ว restore กลับ
	rawBody, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(rawBody))
	logger.Info("[OpenClaw Gateway] IN  raw body len=%d preview=%s",
		len(rawBody), truncateForLog(string(rawBody), 500))

	var req openaiChatRequest
	if err := c.Bind(&req); err != nil {
		logger.Error("[OpenClaw Gateway] bind failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "invalid request: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
	}

	logger.Info("[OpenClaw Gateway] parsed: model=%s stream=%v messages=%d user=%s",
		req.Model, req.Stream, len(req.Messages), req.User)
	for i, m := range req.Messages {
		preview := ""
		switch v := m.Content.(type) {
		case string:
			preview = truncateForLog(v, 200)
		default:
			b, _ := json.Marshal(v)
			preview = truncateForLog(string(b), 200)
		}
		logger.Info("[OpenClaw Gateway]   msg[%d] role=%s content=%s", i, m.Role, preview)
	}

	// Resolve shop + session ก่อนเพื่อให้ compactor ใช้
	holdingCode := getOpenclawHoldingCode()
	sessionID := strings.TrimSpace(req.User)
	if sessionID == "" {
		sessionID = strings.TrimSpace(c.QueryParam("session"))
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("openclaw-%d", time.Now().UnixNano())
	}

	compactCtx, compactCancel := context.WithTimeout(context.Background(), 25*time.Second)
	question := buildQuestionWithCompactedHistory(compactCtx, holdingCode, sessionID, req.Messages)
	compactCancel()
	if strings.TrimSpace(question) == "" {
		logger.Warn("[OpenClaw Gateway] no user message found in %d messages", len(req.Messages))
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "no user message found in request",
				"type":    "invalid_request_error",
			},
		})
	}

	agentReq := AgentV2Request{
		HoldingCode: holdingCode,
		SessionID:   sessionID,
		Question:    question,
		// External OpenAI-compatible clients (OpenClaw, ChatGPT-style UIs)
		// render Markdown — not HTML.
		OutputFormat: "markdown",
	}

	completionID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	createdAt := time.Now().Unix()
	modelName := req.Model
	if modelName == "" {
		modelName = "nongkung"
	}

	logger.Info("[OpenClaw Gateway] model=%s shop=%s session=%s stream=%v question=%q",
		modelName, holdingCode, sessionID, req.Stream, question)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// ========== Real SSE: early-open + keepalive ==========
	//
	// เดิม: รอ agent loop จบ 100% → ค่อย fake-stream ออก → TTFB นานเป็นวินาที (client อาจ timeout)
	// ใหม่: ถ้า stream=true → เปิด SSE stream ทันที + ส่ง role chunk แรก → client เห็น connection alive
	//       ระหว่าง agent ทำงาน → ส่ง SSE comments (`: ...\n\n`) ตาม tool event เป็น keepalive
	//       SSE comments เป็นมาตรฐาน W3C — OpenAI client ทุกตัว ignore เงียบๆ (ไม่กระทบ content)
	//       หลัง agent done → stream final answer เป็น content chunks ตามเดิม
	//
	// Safety: emit callback ถูกเรียกจาก parallel goroutines ใน tool execution → ใช้ mutex คุม

	var sseRespWriter *echo.Response
	var sseEmitMu sync.Mutex
	var sseOpened bool

	openSSEStream := func() {
		if !req.Stream || sseOpened {
			return
		}
		sseOpened = true
		sseRespWriter = c.Response()
		sseRespWriter.Header().Set("Content-Type", "text/event-stream")
		sseRespWriter.Header().Set("Cache-Control", "no-cache")
		sseRespWriter.Header().Set("Connection", "keep-alive")
		sseRespWriter.Header().Set("X-Accel-Buffering", "no")
		sseRespWriter.WriteHeader(http.StatusOK)
		sseRespWriter.Flush()

		// ส่ง role chunk แรกทันที → TTFB ~ms แทน ~s
		roleChunk := openaiChatResponse{
			ID:      completionID,
			Object:  "chat.completion.chunk",
			Created: createdAt,
			Model:   modelName,
			Choices: []openaiChoice{{
				Index: 0,
				Delta: &openaiChatMessage{Role: "assistant", Content: ""},
			}},
		}
		data, _ := json.Marshal(roleChunk)
		fmt.Fprintf(sseRespWriter, "data: %s\n\n", data)
		sseRespWriter.Flush()
	}

	// เปิด stream ก่อนเริ่ม agent loop (สำหรับ stream mode)
	openSSEStream()

	// streamingEmit — safe callback สำหรับ parallel goroutines
	// ส่งเป็น SSE comment (`: ...`) ซึ่ง client จะ ignore แต่รักษา connection alive
	streamingEmit := func(event SSEEvent) {
		logger.Info("[OpenClaw Gateway] SSE event: type=%s", event.Type)
		if !sseOpened || sseRespWriter == nil {
			return
		}
		sseEmitMu.Lock()
		defer sseEmitMu.Unlock()
		// Comment-style heartbeat (client-safe, ไม่กระทบ content)
		// Format: ": event=<type>\n\n"
		comment := fmt.Sprintf(": event=%s t=%d\n\n", event.Type, time.Now().UnixMilli())
		fmt.Fprint(sseRespWriter, comment)
		sseRespWriter.Flush()
	}

	loopStart := time.Now()
	resp, err := routeAgentLoop(ctx, agentReq, streamingEmit)
	logger.Info("[OpenClaw Gateway] agent loop done in %dms err=%v",
		time.Since(loopStart).Milliseconds(), err)
	if err != nil {
		logger.Error("[OpenClaw Gateway] agent loop failed: %v", err)
		if req.Stream {
			return streamOpenAIError(c, completionID, modelName, createdAt, err)
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{
				"message": err.Error(),
				"type":    "internal_error",
			},
		})
	}

	answer := ""
	toolsCount := 0
	if resp != nil && resp.Data != nil {
		answer = resp.Data.Answer
		toolsCount = len(resp.Data.ToolsUsed)
	}
	logger.Info("[OpenClaw Gateway] answer: len=%d tools=%d preview=%q",
		len(answer), toolsCount, truncateForLog(answer, 300))
	if answer == "" {
		logger.Warn("[OpenClaw Gateway] empty answer — returning fallback message")
		answer = "น้องกุ้งตอบไม่ได้ในตอนนี้ค่ะ"
	}

	if !req.Stream {
		// Non-streaming
		stop := "stop"
		return c.JSON(http.StatusOK, openaiChatResponse{
			ID:      completionID,
			Object:  "chat.completion",
			Created: createdAt,
			Model:   modelName,
			Choices: []openaiChoice{
				{
					Index: 0,
					Message: &openaiChatMessage{
						Role:    "assistant",
						Content: answer,
					},
					FinishReason: &stop,
				},
			},
			Usage: mapTokenUsage(resp),
		})
	}

	// Streaming: chunk คำตอบออกเป็น SSE events
	// ถ้า stream ถูกเปิดไปแล้วก่อน agent loop (openSSEStream) → ส่ง chunks ต่อเนื่อง ไม่ตั้ง headers ซ้ำ
	return streamOpenAIAnswer(c, completionID, modelName, createdAt, answer, sseOpened)
}

// mapTokenUsage แปลง TokenUsage ของเรา → openaiUsage
func mapTokenUsage(resp *AgentChatResponse) *openaiUsage {
	if resp == nil || resp.TokenUsage == nil {
		return &openaiUsage{}
	}
	return &openaiUsage{
		PromptTokens:     resp.TokenUsage.PromptTokens,
		CompletionTokens: resp.TokenUsage.CompletionTokens,
		TotalTokens:      resp.TokenUsage.TotalTokens,
	}
}

// ==================== SSE streaming ====================

// streamOpenAIAnswer ส่งคำตอบเป็น OpenAI SSE chunks
//
// เรา chunk เป็น runes (Unicode code points) เพื่อรองรับไทย — chunk ละ 32 runes
//
// alreadyOpen: ถ้า stream ถูกเปิดไว้แล้ว (openSSEStream ทำไปแล้ว) → ข้าม header setup + role chunk
// เพื่อไม่ให้เขียน WriteHeader ซ้ำ (จะ panic)
func streamOpenAIAnswer(c echo.Context, id, model string, created int64, answer string, alreadyOpen bool) error {
	// echo.Response implements http.ResponseWriter + http.Flusher — ใช้ตรงได้เลย
	resp := c.Response()
	if !alreadyOpen {
		resp.Header().Set("Content-Type", "text/event-stream")
		resp.Header().Set("Cache-Control", "no-cache")
		resp.Header().Set("Connection", "keep-alive")
		resp.Header().Set("X-Accel-Buffering", "no")
		resp.WriteHeader(http.StatusOK)
		resp.Flush()
	}

	chunkCount := 0
	writeChunk := func(delta openaiChatMessage, finish *string) {
		chunk := openaiChatResponse{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []openaiChoice{{
				Index:        0,
				Delta:        &delta,
				FinishReason: finish,
			}},
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(resp, "data: %s\n\n", data)
		resp.Flush()
		chunkCount++
	}

	// 1) role chunk — ข้ามถ้า openSSEStream ส่งไปแล้ว (กัน duplicate role chunk)
	if !alreadyOpen {
		writeChunk(openaiChatMessage{Role: "assistant", Content: ""}, nil)
	}

	// 2) content chunks — chunk ละ 32 runes
	runes := []rune(answer)
	const chunkSize = 32
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		writeChunk(openaiChatMessage{Content: string(runes[i:end])}, nil)
	}

	// 3) finish chunk
	stop := "stop"
	writeChunk(openaiChatMessage{}, &stop)

	// 4) [DONE] sentinel
	fmt.Fprint(resp, "data: [DONE]\n\n")
	resp.Flush()

	logger.Info("[OpenClaw Gateway] streamed %d chunks (%d runes total)", chunkCount, len(runes))
	return nil
}

// streamOpenAIError ส่ง error เป็น SSE
// ใช้ sync.Once pattern ภายใน caller เพื่อไม่เขียน header ซ้ำ — แต่ฟังก์ชันนี้ตั้ง header แบบ idempotent-safe:
// WriteHeader ที่ถูกเรียกซ้ำจะถูก echo ignore (echo.Response ใช้ sync.Once ภายใน) จึงไม่ panic
// Header().Set ก่อน WriteHeader ที่เกิดแล้วก็ไม่มีผล แต่ก็ไม่ error — safe
func streamOpenAIError(c echo.Context, id, model string, created int64, err error) error {
	resp := c.Response()
	if !resp.Committed {
		resp.Header().Set("Content-Type", "text/event-stream")
		resp.Header().Set("Cache-Control", "no-cache")
		resp.Header().Set("Connection", "keep-alive")
		resp.WriteHeader(http.StatusOK)
	}

	stop := "stop"
	chunk := openaiChatResponse{
		ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
		Choices: []openaiChoice{{
			Index:        0,
			Delta:        &openaiChatMessage{Role: "assistant", Content: "Error: " + err.Error()},
			FinishReason: &stop,
		}},
	}
	data, _ := json.Marshal(chunk)
	fmt.Fprintf(resp, "data: %s\n\n", data)
	fmt.Fprint(resp, "data: [DONE]\n\n")
	resp.Flush()
	return nil
}
