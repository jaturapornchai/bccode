package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp"
	"strings"
	"time"
)

const maxIterations = 10

// agentSystemPrompt — system prompt สำหรับ agent (ภาษาไทย)
func agentSystemPrompt() string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`คุณเป็นผู้ช่วย AI สำหรับระบบ POS/ERP ของร้านค้า
คุณมี tools สำหรับดึงข้อมูลจากระบบ — ใช้ tools เพื่อตอบคำถามของผู้ใช้

วันที่ปัจจุบัน: %s

กฎ:
- ใช้ tools ดึงข้อมูลจริงก่อนตอบเสมอ (ห้ามเดาข้อมูล)
- ตอบเป็นภาษาไทย ชัดเจน กระชับ
- ใส่ตัวเลขให้ครบ (จำนวน, เปอร์เซ็นต์, จำนวนเงินบาท)
- จัดรูปแบบตัวเลขให้อ่านง่าย เช่น 1,234,567.89 บาท
- ถ้าข้อมูลไม่พอ ให้บอกว่าต้องการข้อมูลอะไรเพิ่ม
- ถ้าต้องดูหลายด้าน ให้เรียก tools ทีละตัวแล้วสรุปรวม
- ถ้าผู้ใช้ถามวันนี้ ให้ใช้วันที่ %s
- ถ้าผู้ใช้ถามเดือนนี้ ไม่ต้องระบุ year/month (tools จะใช้ค่า default)`, today, today)
}

// callWithFallback เรียก AI พร้อม fallback — ถ้า provider แรก fail ให้ลองตัวถัดไป
func callWithFallback(ctx context.Context, providers []aiprovider.ToolCallingProvider, messages []aiprovider.OAIMessage, tools []aiprovider.OAITool, temperature float64) (*aiprovider.OAIResponse, string, error) {
	var lastErr error
	for _, p := range providers {
		resp, err := p.GenerateContentWithTools(ctx, messages, tools, temperature)
		if err == nil {
			return resp, p.Name(), nil
		}
		logger.Warn("[Agent] Provider %s failed: %v — trying next", p.Name(), err)
		lastErr = err
	}
	return nil, "", fmt.Errorf("ทุก provider ใช้ไม่ได้: %v", lastErr)
}

// RunAgentLoop — ReAct loop: AI เรียก tools ซ้ำๆ จนได้คำตอบ
func RunAgentLoop(ctx context.Context, shopID string, question string) (*AgentChatResponse, error) {
	startTime := time.Now()

	// หา providers ที่รองรับ tool calling (ทุกตัวสำหรับ fallback)
	providers := aiprovider.GetAllToolCallingProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("ไม่มี AI Provider ที่รองรับ tool calling")
	}

	mcpServer := mcp.GetDefaultServer()
	if mcpServer == nil {
		return nil, fmt.Errorf("MCP server ยังไม่พร้อม")
	}

	tools := AgentToolDefs()

	// สร้าง messages
	messages := []aiprovider.OAIMessage{
		{Role: "system", Content: agentSystemPrompt()},
		{Role: "user", Content: question},
	}

	var toolsUsed []ToolExecution
	var totalPromptTokens, totalCompletionTokens, totalTokens int
	var modelName string
	iterations := 0

	for i := 0; i < maxIterations; i++ {
		iterations = i + 1
		logger.Info("[Agent] Iteration %d — sending %d messages to AI", iterations, len(messages))

		// เรียก AI พร้อม tools (fallback ข้าม provider ถ้า rate limit)
		resp, providerName, err := callWithFallback(ctx, providers, messages, tools, 0.3)
		if err != nil {
			logger.Error("[Agent] AI call failed at iteration %d: %v", iterations, err)
			return nil, fmt.Errorf("AI ตอบไม่ได้: %w", err)
		}
		_ = providerName

		// สะสม token usage
		totalPromptTokens += resp.Usage.PromptTokens
		totalCompletionTokens += resp.Usage.CompletionTokens
		totalTokens += resp.Usage.TotalTokens
		if modelName == "" {
			modelName = resp.Model
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("AI ไม่มี response")
		}

		assistantMsg := resp.Choices[0].Message
		messages = append(messages, assistantMsg)

		// ถ้าไม่มี tool_calls → คำตอบสุดท้าย
		if len(assistantMsg.ToolCalls) == 0 {
			logger.Info("[Agent] Final answer at iteration %d (tools used: %d)", iterations, len(toolsUsed))

			return &AgentChatResponse{
				Success: true,
				Message: "ตอบสำเร็จ",
				Data: &AgentResponseData{
					Answer:     assistantMsg.Content,
					ToolsUsed:  toolsUsed,
					Iterations: iterations,
				},
				TokenUsage: &TokenUsage{
					PromptTokens:     totalPromptTokens,
					CompletionTokens: totalCompletionTokens,
					TotalTokens:      totalTokens,
					Model:            modelName,
				},
				Timestamp: time.Now(),
			}, nil
		}

		// Execute each tool call
		logger.Info("[Agent] AI เรียก %d tools", len(assistantMsg.ToolCalls))

		for _, tc := range assistantMsg.ToolCalls {
			toolStart := time.Now()
			toolName := tc.Function.Name

			// ตรวจ whitelist
			if !IsAgentTool(toolName) {
				logger.Warn("[Agent] AI พยายามเรียก tool ที่ไม่อนุญาต: %s", toolName)
				toolResultMsg := aiprovider.OAIMessage{
					Role:       "tool",
					Content:    fmt.Sprintf("Error: tool '%s' is not available", toolName),
					ToolCallID: tc.ID,
				}
				messages = append(messages, toolResultMsg)
				toolsUsed = append(toolsUsed, ToolExecution{
					Tool:       toolName,
					Error:      fmt.Sprintf("tool '%s' is not available", toolName),
					DurationMs: time.Since(toolStart).Milliseconds(),
				})
				continue
			}

			// Parse arguments
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				logger.Error("[Agent] Failed to parse tool args for %s: %v", toolName, err)
				toolResultMsg := aiprovider.OAIMessage{
					Role:       "tool",
					Content:    fmt.Sprintf("Error: invalid arguments: %v", err),
					ToolCallID: tc.ID,
				}
				messages = append(messages, toolResultMsg)
				toolsUsed = append(toolsUsed, ToolExecution{
					Tool:       toolName,
					Error:      fmt.Sprintf("invalid arguments: %v", err),
					DurationMs: time.Since(toolStart).Milliseconds(),
				})
				continue
			}

			// Inject shop_id
			params["shop_id"] = shopID

			logger.Info("[Agent] Executing tool: %s", toolName)

			// Execute tool
			result, toolErr := mcpServer.ExecuteToolDirect(ctx, toolName, params)

			durationMs := time.Since(toolStart).Milliseconds()

			var toolContent string
			if toolErr != nil {
				logger.Error("[Agent] Tool %s failed: %v", toolName, toolErr)
				toolContent = fmt.Sprintf("Error: %s", toolErr.Error())
				toolsUsed = append(toolsUsed, ToolExecution{
					Tool:       toolName,
					Params:     params,
					Error:      toolErr.Error(),
					DurationMs: durationMs,
				})
			} else {
				// แปลง result เป็น JSON string
				resultJSON, _ := json.Marshal(result)
				toolContent = string(resultJSON)

				// จำกัดขนาด result ที่ส่งกลับ AI (ไม่เกิน 8000 chars)
				if len(toolContent) > 8000 {
					toolContent = toolContent[:8000] + "...(truncated)"
				}

				logger.Info("[Agent] Tool %s สำเร็จ (%dms, %d chars)", toolName, durationMs, len(toolContent))
				toolsUsed = append(toolsUsed, ToolExecution{
					Tool:       toolName,
					Params:     params,
					DurationMs: durationMs,
				})
			}

			// Append tool result message
			toolResultMsg := aiprovider.OAIMessage{
				Role:       "tool",
				Content:    toolContent,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolResultMsg)
		}
	}

	// ถ้า loop ครบ maxIterations แล้วยังไม่ได้คำตอบ
	logger.Warn("[Agent] Reached max iterations (%d) — forcing final answer", maxIterations)

	// หาคำตอบจาก message สุดท้ายที่เป็น assistant
	var lastAnswer string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" && messages[i].Content != "" {
			lastAnswer = messages[i].Content
			break
		}
	}
	if lastAnswer == "" {
		lastAnswer = "ขออภัย ไม่สามารถสรุปคำตอบได้ในเวลาที่กำหนด"
	}

	// ตัด tool result ออกจาก answer ถ้ามีปนมา
	lastAnswer = strings.TrimSpace(lastAnswer)

	duration := time.Since(startTime)
	logger.Info("[Agent] Completed in %.2fs (iterations: %d, tools: %d, tokens: %d)",
		duration.Seconds(), iterations, len(toolsUsed), totalTokens)

	return &AgentChatResponse{
		Success: true,
		Message: "ตอบสำเร็จ (ใช้ครบ iterations)",
		Data: &AgentResponseData{
			Answer:     lastAnswer,
			ToolsUsed:  toolsUsed,
			Iterations: iterations,
		},
		TokenUsage: &TokenUsage{
			PromptTokens:     totalPromptTokens,
			CompletionTokens: totalCompletionTokens,
			TotalTokens:      totalTokens,
			Model:            modelName,
		},
		Timestamp: time.Now(),
	}, nil
}
