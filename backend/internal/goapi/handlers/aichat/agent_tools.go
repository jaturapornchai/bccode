package aichat

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"smlcloudplatform/internal/goapi/aiprovider"
)

// ====== Lazy-answer detection helpers ======
//
// บาง model (โดยเฉพาะ custom proxy / instruction-following ไม่ดี) ตอบ "กำลังค้นหาให้ค่ะ..."
// ลอยๆ โดยไม่เรียก tool — ทำให้ผู้ใช้ได้ข้อความว่างเปล่า
//
// helpers ด้านล่างใช้ตรวจจับและสร้าง prompt บังคับให้ AI ค้นหาจริง

// lazyAnswerKeywords — signs that AI is procrastinating or fallback-greeting
var lazyAnswerKeywords = []string{
	// Procrastinating phrases
	"กำลังค้น", "กำลังหา", "กำลังตรวจ", "กำลังดู", "กำลังเช็ค",
	"เดี๋ยวค้น", "เดี๋ยวหา", "เดี๋ยวดู", "เดี๋ยวเช็ค", "เดี๋ยวจะ",
	"ขอเวลา", "รอสักครู่", "หาให้", "ค้นให้", "เช็คให้",
	"searching", "looking up", "let me check", "one moment",
	// Greeting fallback (asks user what they want instead of answering)
	"ฉันคือ", "ฉันเป็น", "ฉันสามารถช่วย", "ผมคือ", "ผมสามารถช่วย",
	"กุ้งคือ", "กุ้งสามารถช่วย", "กุ้งเป็น",
	"คุณต้องการความช่วยเหลือ", "คุณต้องการอะไร", "มีอะไรให้ช่วย",
	"how can i help", "what would you like", "i am an assistant",
	// Refusing without trying
	"ไม่สามารถ", "ไม่มีข้อมูล", "ไม่รู้", "i don't know", "i cannot",
}

// nonDataQueryStarters — questions that legitimately don't need a database tool
// (still need a real answer though — Plan + Result + Analysis from own knowledge or websearch)
var nonDataQueryStarters = []string{
	"สูตร", "วิธีทำ", "ทำยังไง", "ทำอย่างไร", "อะไรคือ", "คืออะไร",
	"recipe", "how to", "what is", "explain",
}

// looksLikeLazyAnswer — ตรวจว่า answer text เป็นการ procrastinate ไม่มีข้อมูลจริง
func looksLikeLazyAnswer(text string) bool {
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	for _, kw := range lazyAnswerKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// looksLikeRealQuestion — checks whether the user is asking something that needs a substantive answer
// (not just a greeting). True for almost everything except trivial hi/hello.
//
// Renamed from looksLikeDataQuery to reflect the new policy: every real question deserves
// a Plan+Result+Analysis answer, regardless of whether it needs a DB tool or just AI knowledge.
func looksLikeRealQuestion(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	// Pure greeting (very short, no question mark, no digits)
	if len([]rune(q)) <= 3 {
		return false
	}
	// Common Thai/English greetings — let model handle these naturally
	lower := strings.ToLower(q)
	greetings := []string{"hi", "hello", "hey", "สวัสดี", "หวัดดี", "ดี"}
	for _, g := range greetings {
		if lower == g {
			return false
		}
	}
	// Anything else = real question deserving a real answer
	return true
}

// Backward-compatible alias — keep old name working until callers migrate
func looksLikeDataQuery(question string) bool {
	return looksLikeRealQuestion(question)
}

// buildLazyAnswerReminder — build a system message that forces the AI to actually call a tool.
// Includes a concrete example matching the question pattern.
// LANGUAGE RULE: this reminder goes to the LLM → English (token-efficient).
func buildLazyAnswerReminder(question string) string {
	q := strings.TrimSpace(question)

	var hint strings.Builder
	hint.WriteString("[SYSTEM-CRITICAL] You answered without calling any tool — this violates Rule #1.\n\n")
	hint.WriteString("You MUST call a tool **right now** to fetch real data. Do not say \"searching\" or \"let me check\".\n\n")
	hint.WriteString("Concrete example for this question:\n\n")

	// detect pattern
	digitCount := 0
	for _, r := range q {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}

	if digitCount == 13 {
		// taxid
		hint.WriteString(fmt.Sprintf("```javascript\n// Question is exactly 13 digits = Thai tax ID\nconst id = \"%s\";\nlet r = querymongo(\"debtors\", {\"taxid\": id}, 5);\nif (r.length === 0) r = querymongo(\"creditors\", {\"taxid\": id}, 5);\nif (r.length === 0) r = querymongo(\"productbarcodes\", {\"barcode\": id}, 5);\nreturn {found: r.length, items: r};\n```\n\n", q))
	} else if digitCount >= 8 && digitCount <= 14 {
		// barcode-ish
		hint.WriteString(fmt.Sprintf("```javascript\n// %d-digit number = likely a barcode\nconst rows = querymongo(\"productbarcodes\", {\"barcode\":\"%s\"}, 5);\nreturn rows;\n```\n\n", digitCount, q))
	} else {
		// generic name search
		hint.WriteString("```javascript\n// Search operational data in MongoDB source-of-truth collections\nlet r = querymongo(\"productbarcodes\", {\"names.name\":{\"$regex\":\"KEYWORD\",\"$options\":\"i\"}}, 5);\nif (r.length === 0) r = querymongo(\"debtors\", {\"names.name\":{\"$regex\":\"KEYWORD\",\"$options\":\"i\"}}, 5);\nif (r.length === 0) r = querymongo(\"creditors\", {\"names.name\":{\"$regex\":\"KEYWORD\",\"$options\":\"i\"}}, 5);\nreturn r;\n```\n\n")
	}

	hint.WriteString("**Rule:** If you still don't call a tool this round, the system will force-retry. Please call `executejs` or `querymongodb` **NOW**. Reply in Thai when you have data.")
	return hint.String()
}

// truncateToolResultForUI — แปลง tool result เป็น compact preview สำหรับส่งให้ frontend
//
// ตัวอย่าง use case: querypg คืน 100 rows × 50 columns = ~50KB JSON
// frontend ไม่ต้องการข้อมูลทั้งหมด แค่ preview ก่อน — ถ้าผู้ใช้อยากดูเต็มก็ไปดู log
//
// ขั้นตอน:
//  1. ถ้า result เป็น array → เก็บ first N items + รายงาน total count
//  2. Marshal เป็น JSON แล้วตัดถ้ายาวเกิน maxChars
func truncateToolResultForUI(result any, maxChars int) any {
	if result == nil {
		return nil
	}
	if maxChars <= 0 {
		maxChars = 1500
	}

	// Optimize: ถ้าเป็น slice ใหญ่ → เก็บแค่ 5 ตัวแรก + meta
	switch v := result.(type) {
	case []any:
		if len(v) > 5 {
			return map[string]any{
				"_preview":     v[:5],
				"_total_items": len(v),
				"_truncated":   true,
				"_note":        "showing first 5 of " + jsonItoa(len(v)) + " items",
			}
		}
	case []map[string]any:
		if len(v) > 5 {
			return map[string]any{
				"_preview":     v[:5],
				"_total_items": len(v),
				"_truncated":   true,
				"_note":        "showing first 5 of " + jsonItoa(len(v)) + " items",
			}
		}
	}

	// Marshal เพื่อตรวจขนาด — ถ้าเล็กกว่า maxChars ใช้ได้เลย
	b, err := json.Marshal(result)
	if err != nil || len(b) <= maxChars {
		return result
	}

	// Result ใหญ่เกิน — ส่ง string preview แทน
	return map[string]any{
		"_string_preview": string(b[:maxChars]) + "...(truncated)",
		"_original_size":  len(b),
		"_truncated":      true,
	}
}

// jsonItoa — int → string เลี่ยง strconv import
func jsonItoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// ====== Cached tool definitions ======
//
// AgentToolDefs() ถูกเรียกทุก agent iteration → สร้าง object graph ใหม่ทุกครั้ง (~200 lines worth)
// เนื้อหา tool defs คงที่ตลอด runtime → cache ไว้เลย
// ใช้ sync.Once เพื่อ thread-safe init
var (
	cachedAgentToolDefs []aiprovider.OAITool
	agentToolDefsOnce   sync.Once
)

// agentToolNames — whitelist ของ tools ที่ agent ใช้ได้ (readonly business tools เท่านั้น)
//
// 2026-04-07 ลด tools จาก 27 → 8 ตัว เพื่อลด payload + ป้องกัน 503 timeout จาก upstream
// AI ใช้ executejs เขียน JS ที่ query ข้อมูลเองได้ — ไม่ต้องมี wrapper เยอะๆ
//
// ที่เก็บไว้:
//   - executejs — เครื่องมือหลัก (AI เขียน script + query เอง)
//   - querypostgresql/querymongodb/queryclickhouse — ใช้ตรงเมื่อ query สั้นๆ ไม่ต้องเขียน script
//   - listmongodbcollections/listclickhousetables — สำรวจ schema
//   - aggregatemongodb — pipeline ที่ JS เขียนยาก
//   - websearch — ข้อมูลภายนอก
var agentToolNames = map[string]bool{
	// Code interpreters — AI เขียน + ทดสอบ + รัน เอง
	// Python เป็น primary (LLM เขียน Python เก่งสุด), JS เป็น fallback
	"executepython": true,
	"executejs":     true,
	// Database query tools (readonly — สำหรับ query สั้นๆ)
	"querypostgresql":        true,
	"querymongodb":           true,
	"queryclickhouse":        true,
	"aggregatemongodb":       true,
	"listmongodbcollections": true,
	"listclickhousetables":   true,
	// External
	"websearch":          true,
	"queryknowledgebase": true,
}

// IsAgentTool ตรวจว่า tool อยู่ใน whitelist หรือไม่
func IsAgentTool(name string) bool {
	return agentToolNames[name]
}

// dedupeCitations — รวม citations ที่ซ้ำ (เทียบด้วย URL สำหรับ web, DocID สำหรับ kb)
func dedupeCitations(in []Citation) []Citation {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []Citation
	for _, c := range in {
		key := c.Type + "|"
		if c.Type == "web" {
			key += c.URL
		} else {
			key += c.DocID
			if c.DocID == "" {
				key += c.Label
			}
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

// extractCitationsFromTool ดึง citations ออกจากผลลัพธ์ของ tool
//   - websearch → citation type="web" (label=title, url=URL)
//   - queryknowledgebase → citation type="kb" (label=docname, text=content)
func extractCitationsFromTool(toolName string, result any) []Citation {
	if result == nil {
		return nil
	}
	m, ok := result.(map[string]any)
	if !ok {
		return nil
	}
	switch toolName {
	case "websearch":
		// Expected shape: {"results": [{"title","url","snippet"}, ...]}
		raw, ok := m["results"].([]any)
		if !ok {
			return nil
		}
		var out []Citation
		for _, item := range raw {
			r, ok := item.(map[string]any)
			if !ok {
				continue
			}
			url, _ := r["url"].(string)
			title, _ := r["title"].(string)
			if url == "" {
				continue
			}
			if title == "" {
				title = url
			}
			out = append(out, Citation{Type: "web", Label: title, URL: url})
		}
		return out
	case "queryknowledgebase":
		// Expected shape: {"results": [{"docid","docname","content"}, ...]}
		raw, ok := m["results"].([]any)
		if !ok {
			return nil
		}
		seen := map[string]int{}
		var out []Citation
		for _, item := range raw {
			r, ok := item.(map[string]any)
			if !ok {
				continue
			}
			docID, _ := r["docid"].(string)
			docName, _ := r["docname"].(string)
			content, _ := r["content"].(string)
			if docName == "" && docID == "" {
				continue
			}
			if idx, exists := seen[docID]; exists && docID != "" {
				out[idx].Text += "\n\n" + content
				continue
			}
			seen[docID] = len(out)
			out = append(out, Citation{Type: "kb", Label: docName, DocID: docID, Text: content})
		}
		return out
	}
	return nil
}

// filterAgentTools กรอง tools ตาม search flags ของ request
// nil flag = default true = เปิดใช้งาน (OpenClaw / clients ที่ไม่ส่ง flag ค้นหาทั้งหมด)
var dbToolNames = map[string]bool{
	"executepython": true, "executejs": true,
	"querypostgresql": true, "querymongodb": true, "queryclickhouse": true,
	"aggregatemongodb": true, "listmongodbcollections": true, "listclickhousetables": true,
}

func filterAgentTools(all []aiprovider.OAITool, req AgentV2Request) []aiprovider.OAITool {
	useDB := req.SearchDatabase == nil || *req.SearchDatabase
	useKB := req.SearchKB == nil || *req.SearchKB
	useWeb := req.SearchInternet == nil || *req.SearchInternet
	// ถ้าเปิดทั้งหมด ไม่ต้องกรอง
	if useDB && useKB && useWeb {
		return all
	}
	out := make([]aiprovider.OAITool, 0, len(all))
	for _, t := range all {
		name := t.Function.Name
		if !useDB && dbToolNames[name] {
			continue
		}
		if !useKB && name == "queryknowledgebase" {
			continue
		}
		if !useWeb && name == "websearch" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// AgentToolDefs returns OpenAI function calling format tool definitions
// ผลลัพธ์ถูก cache ไว้ตลอด process lifetime (เนื้อหาคงที่) — safe เพราะ OpenAI client
// ทำ marshal เป็น JSON ก่อนส่ง ไม่ได้ mutate slice
func AgentToolDefs() []aiprovider.OAITool {
	agentToolDefsOnce.Do(func() {
		cachedAgentToolDefs = buildAgentToolDefs()
	})
	return cachedAgentToolDefs
}

// buildAgentToolDefs — สร้าง tool defs จริง (เรียกครั้งเดียวจาก AgentToolDefs)
func buildAgentToolDefs() []aiprovider.OAITool {
	return []aiprovider.OAITool{
		// ==================== Database Query Tools (readonly — เขียนโปรแกรมเอง) ====================
		// MongoDB Query
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "querymongodb",
				Description: "ค้นหาข้อมูลใน MongoDB ด้วย filter (readonly) — ใช้เมื่อ tools สำเร็จรูปไม่มีข้อมูลที่ต้องการ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"collection": map[string]interface{}{
							"type":        "string",
							"description": "ชื่อ collection เช่น productbarcodes, debtors, creditors, transactionSaleInvoice",
						},
						"filter": map[string]interface{}{
							"type":        "string",
							"description": "JSON filter (bson.M format) เช่น {\"itemcode\":\"P001\"} หรือ {\"docdate\":{\"$gte\":\"2026-01-01\"}}",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวน documents สูงสุด (default: 20, max: 100)",
						},
					},
					"required": []string{"collection"},
				},
			},
		},
		// MongoDB Aggregate
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "aggregatemongodb",
				Description: "รัน MongoDB aggregation pipeline (readonly) — สำหรับวิเคราะห์ข้อมูลซับซ้อน group/sum/count/avg",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"collection": map[string]interface{}{
							"type":        "string",
							"description": "ชื่อ collection",
						},
						"pipeline": map[string]interface{}{
							"type":        "string",
							"description": "JSON array ของ pipeline stages เช่น [{\"$match\":{\"docdate\":{\"$gte\":\"2026-01-01\"}}},{\"$group\":{\"_id\":\"$itemcode\",\"total\":{\"$sum\":\"$amount\"}}}]",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนผลลัพธ์สูงสุด (default: 100, max: 100)",
						},
					},
					"required": []string{"collection", "pipeline"},
				},
			},
		},
		// List MongoDB Collections
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "listmongodbcollections",
				Description: "แสดงรายชื่อ collections ทั้งหมดใน MongoDB database",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		// ClickHouse Query
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "queryclickhouse",
				Description: "รัน SQL query บน ClickHouse (readonly SELECT เท่านั้น) — สำหรับ BI/analytics/reporting ข้อมูลขนาดใหญ่",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "SQL SELECT query เช่น SELECT date, sum(amount) FROM sales GROUP BY date ORDER BY date DESC LIMIT 30",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนแถวสูงสุด (default: 100)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		// List ClickHouse Tables
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "listclickhousetables",
				Description: "แสดงรายชื่อ tables ทั้งหมดใน ClickHouse database",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		// PostgreSQL Query
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "querypostgresql",
				Description: "รัน SQL query บน PostgreSQL (readonly SELECT เท่านั้น) — สำหรับ relational processing/projection results เช่น postings, balances, tax/VAT, AR/AP, GL ไม่ใช่ CRUD source",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "SQL SELECT query",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนแถวสูงสุด (default: 100)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		// Python Code Interpreter — **PRIMARY TOOL** สำหรับงาน data/query
		// (LLM เขียน Python เก่งที่สุด — prefer Python over JS whenever possible)
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name: "executepython",
				Description: "**PREFERRED TOOL for data queries and analysis.** Run Python 3 in a readonly sandbox. " +
					"Write idiomatic Python — list comprehensions, dicts, f-strings, loops — exactly as you would in a Jupyter notebook. " +
					"Sandbox helpers (already imported): querymongo(collection, filter=None, limit=200) for MongoDB operational data, querypg(sql, limit=200) for PostgreSQL relational projections, " +
					"querych(sql, limit=200), log(*args). " +
					"**Assign your final answer to the variable `resultvalue`** (no `return` at top level — this runs as a script, not a function). " +
					"Example: `rows = querymongo(\"productbarcodes\", {\"names.name\":{\"$regex\":\"coffee\",\"$options\":\"i\"}}, 10); " +
					"resultvalue = {\"count\": len(rows), \"items\": rows}`. " +
					"Use try/except to handle query errors and fall back to other searches. Timeout: 30s.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code": map[string]interface{}{
							"type":        "string",
							"description": "Python 3 code. Must assign final answer to resultvalue. No INSERT/UPDATE/DELETE/DROP. No network / file I/O. Timeout 30s.",
						},
					},
					"required": []string{"code"},
				},
			},
		},
		// JS Code Interpreter — fallback เมื่อต้องการ specific JS behavior
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name: "executejs",
				Description: "Fallback JavaScript sandbox (use executepython instead when possible — Python is preferred). " +
					"Readonly Goja sandbox. Helpers: querymongo(collection, filter?, limit?) for MongoDB operational data, querypg(sql, limit?) for PostgreSQL relational projections, " +
					"querych(sql, limit?), log(...). Script must `return` a value. Timeout 30s.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code": map[string]interface{}{
							"type":        "string",
							"description": "JavaScript code (ES5+ via Goja). ต้อง return ค่าออกมา. ห้าม INSERT/UPDATE/DELETE/DROP. Timeout 30 วินาที.",
						},
					},
					"required": []string{"code"},
				},
			},
		},
		// Knowledge Base RAG — query เอกสารภายใน shop ที่อัปโหลดไว้
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "queryknowledgebase",
				Description: "Search the shop's Knowledge Base — internal documents the shop has uploaded itself (PDF/Word/Excel/Markdown/Text) via RAG retrieval. This is the shop's own private library: handbooks, policies, manuals, contracts, internal FAQs, uploaded reports. CALL THIS TOOL whenever the user's question is plausibly answered by something the shop wrote down for itself — anything shop-specific that is NOT raw transactional data (products/sales/customers) and NOT general world knowledge. If you find yourself about to answer a shop-specific factual question (a number, a procedure, a contact, a rule) from your training data, you MUST call this tool first instead. Returning 'not found' from this tool is a valid and useful answer; fabricating an answer is not. holdingcode is set automatically — pass query as a natural language question (Thai or English).",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "คำถามที่จะค้นหาใน Knowledge Base (ภาษาไทยหรืออังกฤษ)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		// Web Search (ค้นหาจาก internet)
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "websearch",
				Description: "ค้นหาข้อมูลจาก internet เมื่อต้องการข้อมูลภายนอก เช่น อัตราภาษี กฎหมาย ข้อมูลตลาด เทรนด์ ราคาวัตถุดิบ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "คำค้นหา (ไทยหรืออังกฤษ)",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนผลลัพธ์สูงสุด (default: 5, max: 10)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
	}
}
