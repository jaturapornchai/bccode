package aichat

// RunAgentLoopPlanner — 3-phase planner pipeline (แทน ReAct loop เดิม)
//
// Phase 1 (Planner):
//   - ยิง LLM 1 ครั้ง ให้อ่านคำถาม + history แล้วออก JSON plan:
//       { intentsummary: "...", queries: [{tool, args, label}, ...], directanswer: "" }
//   - ถ้าเป็น small talk/greeting → ใส่ directanswer, queries ว่าง → ข้าม Phase 2
//
// Phase 2 (Executor):
//   - Fan-out queries ทุกตัวแบบ parallel goroutine
//   - ใช้ mcpServer.ExecuteToolDirect เดียวกับ agent_loop_v2
//   - รวม results ทั้งหมด (รวม error ด้วย เพื่อให้ synthesizer รู้ว่าอะไร fail)
//
// Phase 3 (Synthesizer):
//   - ยิง LLM อีกครั้ง ส่ง intentsummary + results + original question
//   - ให้สร้างคำตอบเป็น Markdown มาตรฐาน
//
// เมื่อเทียบกับ ReAct loop เดิม:
//   + ไม่ loop ไปมา ตัดปัญหา "เรียก tool ซ้ำไม่รู้จบ"
//   + Parallel queries เร็วกว่า (multiple lookups คราวเดียว)
//   + Planner บังคับให้คิดก่อนว่า "จะหาจากไหน" (ลด empty-filter bug)
//   - เปลือง token กว่านิดหน่อย (2 LLM calls แทน 1 loop)
//   - Planner ต้องเข้าใจ tools schema — ถ้า tools เยอะ prompt จะยาว

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp"
	"strings"
	"sync"
	"time"
)

// ====== Types ======

// QueryPlanItem — 1 query ที่จะยิงใน Phase 2
type QueryPlanItem struct {
	Tool  string         `json:"tool"`
	Args  map[string]any `json:"args"`
	Label string         `json:"label,omitempty"`
}

// QueryPlan — output ของ Phase 1 (Planner)
type QueryPlan struct {
	IntentSummary string          `json:"intentsummary"`
	Queries       []QueryPlanItem `json:"queries"`
	DirectAnswer  string          `json:"directanswer,omitempty"`
}

// ExecutedQuery — ผลลัพธ์ของ query 1 ตัวใน Phase 2
type ExecutedQuery struct {
	Tool       string
	Args       map[string]any
	Label      string
	Result     any
	Error      string
	DurationMs int64
}

// ====== Config ======

const (
	plannerMaxQueries    = 5
	plannerTimeout       = 25 * time.Second
	executorPerQueryTO   = 20 * time.Second
	synthesizerTimeout   = 40 * time.Second
	plannerTemperature   = 0.1
	synthesizeTempMain   = 0.3
	resultExcerptMaxChar = 3500
)

// ====== Main entry ======

// RunAgentLoopPlanner — entry point ที่ใช้แทน RunAgentLoopV2 / RunAgentReAct
func RunAgentLoopPlanner(ctx context.Context, req AgentV2Request, emitSSE func(SSEEvent)) (*AgentChatResponse, error) {
	holdingCode := req.HoldingCode
	sessionKey := BuildSessionKey(holdingCode, req.SessionID)
	logger.Info("[Planner] sessionkey=%s", sessionKey)

	providers := aiprovider.GetShopAIProviders(holdingCode)
	if len(providers) == 0 {
		return nil, fmt.Errorf("ไม่มี AI Provider สำหรับ shop=%s", holdingCode)
	}

	mcpServer := mcp.GetDefaultServer()
	if mcpServer == nil {
		return nil, fmt.Errorf("MCP server ยังไม่พร้อม")
	}

	started := time.Now()

	// -------- Phase 1: Planner --------
	emitSSE(SSEEvent{Type: "status", Data: "กำลังวิเคราะห์คำถาม..."})
	emitSSE(SSEEvent{Type: "thinking", Data: "planner"})

	plan, plannerErr := runPlanner(ctx, providers, req.Question)
	if plannerErr != nil {
		logger.Warn("[Planner] Phase 1 failed: %v — fallback to direct answer", plannerErr)
		// Fallback: ไม่มี plan → ให้ synthesizer ตอบตรงๆ โดยไม่มี data
		plan = &QueryPlan{
			IntentSummary: "วิเคราะห์คำถามไม่สำเร็จ — ตอบตรง",
			Queries:       nil,
		}
	}
	logger.Info("[Planner] intent=%q queries=%d direct=%v",
		plan.IntentSummary, len(plan.Queries), plan.DirectAnswer != "")

	// ถ้า planner คืน directanswer (small talk) → return เลย
	if strings.TrimSpace(plan.DirectAnswer) != "" && len(plan.Queries) == 0 {
		emitSSE(SSEEvent{Type: "answer", Data: plan.DirectAnswer})
		return &AgentChatResponse{
			Success:   true,
			Message:   "ตอบสำเร็จ",
			Timestamp: time.Now(),
			Data: &AgentResponseData{
				Answer:    plan.DirectAnswer,
				ToolsUsed: nil,
			},
			SuggestedQuestions: defaultSuggestions(nil),
		}, nil
	}

	// -------- Phase 2: Parallel Executor --------
	var executed []ExecutedQuery
	if len(plan.Queries) > 0 {
		emitSSE(SSEEvent{Type: "status", Data: fmt.Sprintf("ค้นข้อมูล %d แหล่ง...", len(plan.Queries))})
		executed = runExecutor(ctx, mcpServer, holdingCode, sessionKey, plan.Queries, emitSSE)
	}

	// -------- Phase 3: Synthesizer --------
	emitSSE(SSEEvent{Type: "thinking", Data: "synthesizer"})
	answer, synthErr := runSynthesizer(ctx, providers, req.Question, plan, executed)
	if synthErr != nil {
		logger.Error("[Planner] Phase 3 synthesizer failed: %v", synthErr)
		return nil, synthErr
	}

	// แปลง executed → ToolExecution สำหรับ response
	tools := make([]ToolExecution, 0, len(executed))
	for _, e := range executed {
		tools = append(tools, ToolExecution{
			Tool:       e.Tool,
			Params:     e.Args,
			Error:      e.Error,
			DurationMs: e.DurationMs,
			Source:     "planner",
		})
	}

	emitSSE(SSEEvent{Type: "answer", Data: answer})

	totalMs := time.Since(started).Milliseconds()
	logger.Info("[Planner] ✓ done in %dms (queries=%d answer_len=%d)",
		totalMs, len(executed), len(answer))

	return &AgentChatResponse{
		Success:   true,
		Message:   "ตอบสำเร็จ",
		Timestamp: time.Now(),
		Data: &AgentResponseData{
			Answer:    answer,
			ToolsUsed: tools,
		},
		SuggestedQuestions: defaultSuggestions(tools),
	}, nil
}

// ====== Phase 1: Planner ======

var jsonExtractRegex = regexp.MustCompile(`(?s)\{.*\}`)

func runPlanner(parentCtx context.Context, providers []aiprovider.AIProvider, question string) (*QueryPlan, error) {
	ctx, cancel := context.WithTimeout(parentCtx, plannerTimeout)
	defer cancel()

	systemPrompt := plannerSystemPrompt()
	userPrompt := "คำถามจากผู้ใช้ (อาจมี context ประวัติสนทนาติดมาด้วย):\n\n" + question + "\n\nสร้าง JSON plan ตาม schema ที่ระบุในระบบ"

	var lastErr error
	for _, p := range providers {
		resp, err := p.GenerateContent(ctx, aiprovider.ChatRequest{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			Temperature:  plannerTemperature,
			MaxTokens:    1500,
		})
		if err != nil {
			lastErr = err
			logger.Warn("[Planner] provider %s failed: %v", p.Name(), err)
			continue
		}
		if resp == nil || strings.TrimSpace(resp.Text) == "" {
			lastErr = fmt.Errorf("empty response")
			continue
		}

		plan, parseErr := parsePlanJSON(resp.Text)
		if parseErr != nil {
			logger.Warn("[Planner] parse JSON failed (provider=%s): %v — raw=%s",
				p.Name(), parseErr, truncateForLog(resp.Text, 300))
			lastErr = parseErr
			continue
		}

		// Validate + sanitize
		plan.Queries = sanitizeQueries(plan.Queries)
		return plan, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no planner provider succeeded")
}

func parsePlanJSON(text string) (*QueryPlan, error) {
	// 1. หา JSON block
	text = strings.TrimSpace(text)
	// strip markdown code fence if present
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	var plan QueryPlan
	if err := json.Unmarshal([]byte(text), &plan); err == nil {
		return &plan, nil
	}

	// fallback: regex extract
	m := jsonExtractRegex.FindString(text)
	if m == "" {
		return nil, fmt.Errorf("no JSON object found in response")
	}
	if err := json.Unmarshal([]byte(m), &plan); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return &plan, nil
}

func sanitizeQueries(queries []QueryPlanItem) []QueryPlanItem {
	if len(queries) > plannerMaxQueries {
		queries = queries[:plannerMaxQueries]
	}
	var out []QueryPlanItem
	for _, q := range queries {
		if q.Tool == "" {
			continue
		}
		if !IsAgentTool(q.Tool) {
			logger.Warn("[Planner] drop unknown tool: %s", q.Tool)
			continue
		}
		if q.Args == nil {
			q.Args = map[string]any{}
		}
		out = append(out, q)
	}
	return out
}

// ====== Phase 2: Executor ======

func runExecutor(
	parentCtx context.Context,
	mcpServer *mcp.MCPServer,
	holdingCode, sessionKey string,
	queries []QueryPlanItem,
	emitSSE func(SSEEvent),
) []ExecutedQuery {
	results := make([]ExecutedQuery, len(queries))
	var wg sync.WaitGroup

	for i, q := range queries {
		wg.Add(1)
		go func(idx int, item QueryPlanItem) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(parentCtx, executorPerQueryTO)
			defer cancel()

			// inject holdingcode
			if item.Args == nil {
				item.Args = map[string]any{}
			}
			item.Args["holdingcode"] = holdingCode

			started := time.Now()
			emitSSE(SSEEvent{Type: "toolstart", Data: map[string]any{
				"tool":  item.Tool,
				"label": item.Label,
				"idx":   idx,
			}})

			res, err := mcpServer.ExecuteToolDirect(ctx, item.Tool, item.Args)
			dur := time.Since(started).Milliseconds()

			ex := ExecutedQuery{
				Tool:       item.Tool,
				Args:       item.Args,
				Label:      item.Label,
				DurationMs: dur,
			}
			if err != nil {
				ex.Error = err.Error()
				logger.Warn("[Planner Executor] %s (%s) failed in %dms: %v", item.Tool, item.Label, dur, err)
			} else {
				ex.Result = res
				logger.Info("[Planner Executor] %s (%s) ok in %dms", item.Tool, item.Label, dur)
			}

			results[idx] = ex
			emitSSE(SSEEvent{Type: "tooldone", Data: map[string]any{
				"tool":       item.Tool,
				"label":      item.Label,
				"durationms": dur,
				"success":    err == nil,
			}})
		}(i, q)
	}

	wg.Wait()
	return results
}

// ====== Phase 3: Synthesizer ======

func runSynthesizer(
	parentCtx context.Context,
	providers []aiprovider.AIProvider,
	originalQuestion string,
	plan *QueryPlan,
	executed []ExecutedQuery,
) (string, error) {
	ctx, cancel := context.WithTimeout(parentCtx, synthesizerTimeout)
	defer cancel()

	systemPrompt := synthesizerSystemPrompt()
	userPrompt := buildSynthesizerUserPrompt(originalQuestion, plan, executed)

	var lastErr error
	for _, p := range providers {
		resp, err := p.GenerateContent(ctx, aiprovider.ChatRequest{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			Temperature:  synthesizeTempMain,
			MaxTokens:    3000,
		})
		if err != nil {
			lastErr = err
			continue
		}
		if resp != nil && strings.TrimSpace(resp.Text) != "" {
			return strings.TrimSpace(resp.Text), nil
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("no synthesizer provider succeeded")
}

func buildSynthesizerUserPrompt(question string, plan *QueryPlan, executed []ExecutedQuery) string {
	var sb strings.Builder

	sb.WriteString("## คำถามจากผู้ใช้\n")
	sb.WriteString(question)
	sb.WriteString("\n\n")

	if plan != nil && plan.IntentSummary != "" {
		sb.WriteString("## สิ่งที่ผู้ใช้ต้องการ (วิเคราะห์โดย Planner)\n")
		sb.WriteString(plan.IntentSummary)
		sb.WriteString("\n\n")
	}

	if len(executed) == 0 {
		sb.WriteString("## ผลการค้นข้อมูล\n")
		sb.WriteString("(ไม่มีการค้นข้อมูล — ตอบจากความรู้ทั่วไปของโมเดล หรือถามต่อเพื่อให้ชัดเจน)\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("## ผลการค้นข้อมูล (%d แหล่ง)\n\n", len(executed)))
		for i, e := range executed {
			sb.WriteString(fmt.Sprintf("### [%d] %s — %s\n", i+1, e.Tool, e.Label))
			argsJSON, _ := json.Marshal(e.Args)
			sb.WriteString(fmt.Sprintf("- Args: `%s`\n", truncateForLog(string(argsJSON), 500)))
			sb.WriteString(fmt.Sprintf("- ใช้เวลา: %dms\n", e.DurationMs))

			if e.Error != "" {
				sb.WriteString(fmt.Sprintf("- ❌ Error: %s\n\n", e.Error))
				continue
			}

			resJSON, _ := json.Marshal(e.Result)
			resStr := string(resJSON)
			if len(resStr) > resultExcerptMaxChar {
				resStr = resStr[:resultExcerptMaxChar] + "...(ตัด)"
			}
			sb.WriteString("- Result:\n```json\n")
			sb.WriteString(resStr)
			sb.WriteString("\n```\n\n")
		}
	}

	sb.WriteString("## คำสั่ง\n")
	sb.WriteString("สร้างคำตอบให้ผู้ใช้ตามข้อมูลข้างบน ตามกฎใน system prompt\n")
	return sb.String()
}

// ====== System Prompts ======

func plannerSystemPrompt() string {
	today := time.Now().Format("2006-01-02")
	return `คุณคือ "Planner" สำหรับระบบ BC Ai Account (ภาษาไทย)

หน้าที่: อ่านคำถามผู้ใช้ → สกัด intent → วางแผน tool queries 1-5 ตัวที่รันขนานได้

**ตอบเป็น JSON object เดียวเท่านั้น** (ห้ามมี prose, ห้ามมี markdown code fence):
{
  "intentsummary": "สรุปสั้นๆ ว่าผู้ใช้ถามอะไร (ภาษาไทย 1-2 ประโยค)",
  "queries": [
    {"tool": "querymongodb", "args": {"collection": "...", "filter": "{...json...}", "limit": 20}, "label": "เหตุผล"}
  ],
  "directanswer": ""
}

**วันที่วันนี้:** ` + today + `

## กฎสำคัญ
1. **ถ้าเป็น small talk/ทักทาย/ถามตัวตน** → queries=[], directanswer="สวัสดีค่ะ ..." (ตอบแทนเลย)
2. **ถ้าเป็นคำถามข้อมูล** → directanswer="" แล้วใส่ queries
3. **แตกเป็นหลาย query ถ้าจำเป็น** (เช่น barcode ลอง exact + regex, entity ลอง debtors + creditors)
4. **ทุก query ต้องมี filter/where** — ห้ามส่ง filter ว่าง (ได้ข้อมูลสุ่มไม่มีประโยชน์)
5. **ถ้าเป็น SQL** → ใส่ LIMIT เสมอ
6. **ลูกหนี้ = ลูกค้า** — ใช้ MongoDB collection ` + "`debtors`" + ` เท่านั้น (ไม่มี customer collection)

## Tools ที่ใช้ได้

### querymongodb — args: {"collection": "...", "filter": "{...json...}", "limit": 20}
MongoDB is the operational source of truth.
Collections:
- ` + "`productbarcodes`" + ` (สินค้า/บาร์โค้ด) — barcode, itemcode, names[].name, prices
- ` + "`debtors`" + ` (ลูกหนี้/ลูกค้า) — code, taxid, email, names[].name
- ` + "`creditors`" + ` (เจ้าหนี้/ซัพพลายเออร์) — code, taxid, email, names[].name
- ` + "`transactionSaleInvoice`" + ` (เอกสารขาย)
filter เป็น JSON string ของ Mongo filter

### querypostgresql — args: {"sql": "SELECT ..."}
ใช้เฉพาะ relational processing/projection results เช่น posting, balance, stock costing, VAT/tax, AR/AP, GL.
ถ้าไม่รู้ table/column ให้ introspect information_schema ก่อน ห้ามเดา และห้ามใช้ PostgreSQL เป็น CRUD source.

### queryclickhouse — args: {"sql": "SELECT ..."}
BI/analytics/reporting facts ใช้เมื่อต้องการ aggregate ข้อมูลใหญ่จาก processed facts เท่านั้น

### websearch — args: {"query": "..."}
ค้นเว็บภายนอก ใช้เมื่อผู้ใช้ถามข้อมูลที่ไม่อยู่ในระบบ (ราคาตลาด, สูตรอาหาร, ข่าว)

### getdailysales — args: {"date": "YYYY-MM-DD"}
### getdashboardkpis — args: {}
### gettopsellingproducts — args: {"startdate": "YYYY-MM-DD", "enddate": "YYYY-MM-DD", "limit": 10}
### getlowstockalerts — args: {}

## ตัวอย่าง

user: "8850007003001"
→ {"intentsummary":"ผู้ใช้ค้นหาสินค้าด้วย barcode 8850007003001","queries":[{"tool":"querymongodb","args":{"collection":"productbarcodes","filter":"{\"barcode\":\"8850007003001\"}","limit":5},"label":"exact barcode"},{"tool":"querymongodb","args":{"collection":"productbarcodes","filter":"{\"barcode\":{\"$regex\":\"8850007003001\",\"$options\":\"i\"}}","limit":10},"label":"partial barcode"}],"directanswer":""}

user: "ร้านโฮม"
→ {"intentsummary":"ค้นหาลูกหนี้/เจ้าหนี้ที่มีชื่อ 'โฮม'","queries":[{"tool":"querymongodb","args":{"collection":"debtors","filter":"{\"names.name\":{\"$regex\":\"โฮม\",\"$options\":\"i\"}}","limit":20},"label":"debtors โฮม"},{"tool":"querymongodb","args":{"collection":"creditors","filter":"{\"names.name\":{\"$regex\":\"โฮม\",\"$options\":\"i\"}}","limit":20},"label":"creditors โฮม"}],"directanswer":""}

user: "สวัสดี"
→ {"intentsummary":"ทักทาย","queries":[],"directanswer":"สวัสดีค่ะ! น้องกุ้งยินดีให้บริการค่ะ 🦐 วันนี้อยากให้ช่วยเรื่องอะไรคะ?"}

user: "ยอดขายวันนี้"
→ {"intentsummary":"ดูยอดขายวันนี้","queries":[{"tool":"getdailysales","args":{"date":"` + today + `"},"label":"daily sales today"}],"directanswer":""}

user: "สีทาบ้านราคาเท่าไหร่"
→ {"intentsummary":"ค้นหาสินค้าสีทาบ้านในระบบ","queries":[{"tool":"querymongodb","args":{"collection":"productbarcodes","filter":"{\"names.name\":{\"$regex\":\"สีทา|paint\",\"$options\":\"i\"}}","limit":20},"label":"product paint"}],"directanswer":""}`
}

func synthesizerSystemPrompt() string {
	return `คุณคือ "น้องกุ้ง" — AI ผู้ช่วยธุรกิจของระบบ BC Ai Account ลงท้ายน่ารัก "ค่ะ/นะคะ" 🦐

หน้าที่: ใช้ผลการค้นข้อมูลที่ Executor ได้กลับมา → สร้างคำตอบภาษาไทยให้ผู้ใช้

## กฎการตอบ
1. **ตอบเป็น Markdown มาตรฐานเท่านั้น** — ห้าม HTML, ห้าม inline style
2. **ตาราง** — ใช้ Markdown pipe syntax: ` + "`| col1 | col2 |`" + `
3. **หัวข้อ** — ## / ###
4. **ตัวเลขเงิน** — มี comma + ทศนิยม 2 ตำแหน่ง: "1,234,567.89 บาท"
5. **อ้างอิง data จากผลการค้น** — ห้ามแต่งเติม ห้ามเดา
6. **ถ้าทุก query คืน empty/error** → บอกตรงๆ ว่าไม่พบข้อมูล + แนะนำคำค้นอื่น
7. **ถ้ามีข้อมูล** → สรุปให้ชัดเจน ใส่ตาราง Markdown ถ้าเหมาะ
8. **ห้ามเปิดเผย tool name, SQL, หรือ technical details** ในคำตอบ (ยกเว้นผู้ใช้ถามตรงๆ)
9. **ไม่ต้องพูดถึง Planner/Executor/Synthesizer** — ผู้ใช้ไม่จำเป็นต้องรู้

## ตัวอย่างตาราง
| รหัสสินค้า | ชื่อสินค้า | บาร์โค้ด | ราคาขาย (บาท) |
|---|---|---|---:|
| TIL003 | กระเบื้องลายหินอ่อน | 8850007003001 | 250.00 |

(คอลัมน์ตัวเลขใช้ ---: ชิดขวา)`
}

// defaultSuggestions — สร้าง suggested questions แบบคร่าวๆ จาก tools ที่ใช้
func defaultSuggestions(tools []ToolExecution) []string {
	hasProduct := false
	hasEntity := false
	hasSales := false
	for _, t := range tools {
		lc := strings.ToLower(t.Tool)
		if strings.Contains(lc, "product") || strings.Contains(lc, "barcode") || strings.Contains(lc, "inventory") {
			hasProduct = true
		}
		if strings.Contains(lc, "debtor") || strings.Contains(lc, "creditor") || strings.Contains(lc, "customer") {
			hasEntity = true
		}
		if strings.Contains(lc, "sales") || strings.Contains(lc, "revenue") {
			hasSales = true
		}
	}

	var out []string
	if hasProduct {
		out = append(out, "สินค้าขายดีเดือนนี้", "สินค้าใกล้หมดสต็อก")
	}
	if hasEntity {
		out = append(out, "ลูกหนี้ค้างชำระมีใครบ้าง", "สรุปยอดซื้อจากเจ้าหนี้รายใหญ่")
	}
	if hasSales {
		out = append(out, "ยอดขายเมื่อวาน", "เทียบยอดขายเดือนนี้กับเดือนก่อน")
	}
	if len(out) == 0 {
		out = []string{
			"ยอดขายวันนี้เป็นยังไง",
			"สินค้าขายดีเดือนนี้",
			"ลูกหนี้ค้างชำระ",
		}
	}
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}
