package aichat

import (
	"context"
	"fmt"
)

// ──────────────────────────────────────────────────────────────────────────
// MCP ถูกถอดออกจาก project ชั่วคราว (จะทำใหม่หลังโปรแกรมหลักเสร็จ).
// agentToolServer เป็น stub แทน *mcp.MCPServer เดิม เพื่อให้ agentic chat
// (RunAgentLoop / V2 / ReAct / Planner / openai gateway) ยัง compile + ทำงาน
// ได้ แต่ tool execution คืน error — agent loop จึงไม่มี tool-calling.
// ──────────────────────────────────────────────────────────────────────────

type agentToolServer struct{}

// getAgentToolServer แทน mcp.GetDefaultServer() เดิม (ไม่คืน nil).
func getAgentToolServer() *agentToolServer {
	return &agentToolServer{}
}

// ExecuteToolDirect — stub: MCP ถอดออก, ทุก tool call คืน error.
// signature ตรงกับ mcp.MCPServer.ExecuteToolDirect เดิม จึงใช้แทนได้ทันที
// (รวมการส่งเป็น argument ให้ dispatchAgentTool).
func (*agentToolServer) ExecuteToolDirect(ctx context.Context, name string, params map[string]any) (any, error) {
	return nil, fmt.Errorf("AI tools ปิดใช้งานชั่วคราว: ระบบ MCP ถูกถอดออก (รอเปิดใหม่)")
}
