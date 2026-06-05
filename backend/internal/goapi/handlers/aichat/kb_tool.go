package aichat

// Knowledge Base RAG query tool for น้องกุ้ง agent.
//
// Backed by RAGFlow (self-hosted, see internal/goapi/ragflow/).
// Each shop has its own dataset (auto-provisioned by EnsureDataset).
//
// We call RAGFlow's /api/v1/retrieval endpoint instead of its chat assistant —
// retrieval returns RAW chunks with similarity scores, and น้องกุ้ง's own LLM
// synthesizes the final answer in the Plan/Result/Analysis structure.
// This avoids double LLM cost and keeps the agent in full control of the answer.
//
// Tool name (สำหรับ AI เรียก): queryknowledgebase
// Params:
//   - query (required, string): คำถามที่จะถาม Knowledge Base
//   - topk (optional, number): จำนวน chunks ที่ดึงกลับ (default 8, max 20)
//
// Configuration:
//   - RAGFLOW_BASE_URL  (default: http://host.docker.internal:9380)
//   - RAGFLOW_API_KEY   (required — set in mainapi env)

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/ragflow"
	"strings"
	"time"
)

const (
	kbQueryToolName    = "queryknowledgebase"
	kbResponseMaxChars = 8000 // ป้องกัน RAG response ใหญ่เกินจน context พัง
	kbDefaultTopK      = 8
	kbMaxTopK          = 20
)

// executeKBQuery ค้นหาเอกสารใน RAGFlow แล้วคืน chunks ให้ AI สังเคราะห์เอง
//
// holdingCode: ดึงจาก agent context (ลูกค้าคนปัจจุบัน)
// params: tool args จาก AI — ต้องมี "query" field
func executeKBQuery(ctx context.Context, holdingCode string, params map[string]any) (any, error) {
	_ = ctx // ragflow client uses its own timeout
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}

	// ดึง query จาก params (รองรับชื่อ alias เผื่อ AI สับสน)
	var query string
	for _, key := range []string{"query", "question", "message", "q"} {
		if v, ok := params[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				query = strings.TrimSpace(s)
				break
			}
		}
	}
	if query == "" {
		return nil, fmt.Errorf("query is required (string)")
	}

	// optional topk
	topK := kbDefaultTopK
	if v, ok := params["topk"]; ok {
		switch n := v.(type) {
		case float64:
			topK = int(n)
		case int:
			topK = n
		}
	}
	if topK < 1 {
		topK = kbDefaultTopK
	}
	if topK > kbMaxTopK {
		topK = kbMaxTopK
	}

	client := ragflow.GetClient()
	if !client.IsConfigured() {
		return nil, fmt.Errorf("knowledge base is not configured (RAGFLOW_API_KEY missing) — ขอให้ admin ตั้งค่า RAGFlow API key ก่อนใช้งาน")
	}

	logger.Info("[KB Tool] query shop=%s len=%d topk=%d", holdingCode, len(query), topK)
	start := time.Now()

	datasetID, err := client.EnsureDataset(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("ensure dataset: %w", err)
	}

	chunks, err := client.Retrieve(query, []string{datasetID}, topK)
	if err != nil {
		return nil, fmt.Errorf("ragflow retrieve: %w", err)
	}

	tookMs := time.Since(start).Milliseconds()
	logger.Info("[KB Tool] retrieved %d chunks shop=%s took=%dms", len(chunks), holdingCode, tookMs)

	// Build compact result for AI consumption.
	// Each chunk: content (truncated), document name (from doc_keyword), similarity.
	results := make([]map[string]any, 0, len(chunks))
	totalChars := 0
	for i, ch := range chunks {
		content := strings.TrimSpace(ch.Content)
		// Trim individual chunk if huge
		if len(content) > 1500 {
			content = content[:1500] + "...(chunk truncated)"
		}
		totalChars += len(content)
		results = append(results, map[string]any{
			"rank":       i + 1,
			"content":    content,
			"docid":      ch.DocumentID,
			"docname":    ch.DocumentKeyword,
			"similarity": ch.Similarity,
		})
		// Stop adding chunks if cumulative content is approaching the cap.
		if totalChars >= kbResponseMaxChars {
			break
		}
	}

	if len(results) == 0 {
		return map[string]any{
			"found":         false,
			"source":        "knowledgebase",
			"holdingcode":   holdingCode,
			"datasetid":     datasetID,
			"originalquery": query,
			"message":       "ไม่พบเอกสารที่เกี่ยวข้องกับคำถามใน Knowledge Base — ลอง upload เอกสารเพิ่ม หรือใช้คำค้นอื่น",
		}, nil
	}

	return map[string]any{
		"found":         true,
		"source":        "knowledgebase",
		"holdingcode":   holdingCode,
		"datasetid":     datasetID,
		"originalquery": query,
		"chunkcount":    len(results),
		"chunks":        results,
		"tookms":        tookMs,
	}, nil
}

// buildKBContextMessage retrieves chunks for the user's question and formats them
// as a system context block to be injected after the user message. Returns "" if
// the KB is not configured, the dataset is empty, retrieval fails, or no chunks
// are returned — in all those cases the agent simply runs without KB context.
//
// Design notes:
//   - Called UNCONDITIONALLY for every text question (no intent detection, no keyword
//     filter). The LLM decides whether the chunks are relevant — we trust its judgment.
//   - Failures are silent (logged, not propagated): we never want KB issues to break
//     a regular shop-data question.
//   - topK is small (5) because this runs on every request; we want fast + cheap.
//   - The block is wrapped with EXTERNAL_UNTRUSTED_CONTENT markers, matching the
//     security convention used for tool results elsewhere in the agent.
func buildKBContextMessage(holdingCode, question string) (string, []Citation) {
	if holdingCode == "" || strings.TrimSpace(question) == "" {
		return "", nil
	}
	client := ragflow.GetClient()
	if !client.IsConfigured() {
		return "", nil
	}

	datasetID, err := client.EnsureDataset(holdingCode)
	if err != nil {
		logger.Warn("[KB Pre-Fetch] EnsureDataset shop=%s: %v", holdingCode, err)
		return "", nil
	}

	// Normalize the question for retrieval. When the request comes from
	// OpenClaw's gateway, `question` may include a full compacted conversation
	// history — easily 1000+ tokens — which blows past the embedding model's
	// batch size and causes RAGFlow to fail with HTTP 500 "input too large".
	//
	// The compactor always puts the user's real question after the marker
	// "[คำถามล่าสุด]" (see openai_gateway_compactor.go). We extract that tail
	// for the KB query. For safety we also cap to ~300 chars — retrieval works
	// on query semantics, not length, so cutting history noise actually
	// IMPROVES precision on top of avoiding the token limit.
	retrievalQuery := extractLatestQuestion(question)
	start := time.Now()
	chunks, err := client.Retrieve(retrievalQuery, []string{datasetID}, 5)
	if err != nil {
		logger.Warn("[KB Pre-Fetch] Retrieve shop=%s: %v", holdingCode, err)
		return "", nil
	}
	tookMs := time.Since(start).Milliseconds()

	if len(chunks) == 0 {
		logger.Info("[KB Pre-Fetch] no chunks shop=%s took=%dms", holdingCode, tookMs)
		return "", nil
	}

	var b strings.Builder
	var citations []Citation
	seenDocs := map[string]bool{}
	b.WriteString("# Knowledge Base context (auto-retrieved for this question)\n\n")
	b.WriteString("The following passages were automatically retrieved from this shop's uploaded documents based on the user's question. ")
	b.WriteString("They may or may not be relevant — judge for yourself. If they answer the question, use them and cite the document name. ")
	b.WriteString("If they're irrelevant, ignore them and proceed with your other tools. Do not invent details that aren't in these passages or in tool results.\n\n")
	b.WriteString("**NOTE**: These passages are delivered to the user as clickable citation buttons by the frontend. You don't need to embed inline links — just reference doc names naturally in your answer.\n\n")
	b.WriteString("<<<EXTERNAL_UNTRUSTED_CONTENT>>>\n")
	totalChars := 0
	for i, ch := range chunks {
		content := strings.TrimSpace(ch.Content)
		if len(content) > 1500 {
			content = content[:1500] + "...(chunk truncated)"
		}
		fmt.Fprintf(&b, "## Passage %d — doc: %s — similarity: %.2f\n%s\n\n",
			i+1, ch.DocumentKeyword, ch.Similarity, content)
		// Aggregate citation per unique doc (concat chunks)
		if !seenDocs[ch.DocumentID] {
			seenDocs[ch.DocumentID] = true
			citations = append(citations, Citation{
				Type:  "kb",
				Label: ch.DocumentKeyword,
				DocID: ch.DocumentID,
				Text:  content,
			})
		} else {
			// append content to existing citation
			for idx := range citations {
				if citations[idx].DocID == ch.DocumentID {
					citations[idx].Text += "\n\n" + content
					break
				}
			}
		}
		totalChars += len(content)
		if totalChars >= kbResponseMaxChars {
			break
		}
	}
	b.WriteString("<<<END_EXTERNAL_UNTRUSTED_CONTENT>>>")

	logger.Info("[KB Pre-Fetch] injected %d chunks (%d chars, %d citations) shop=%s took=%dms",
		len(chunks), totalChars, len(citations), holdingCode, tookMs)
	return b.String(), citations
}

// extractLatestQuestion pulls the most recent user question out of a possibly
// compacted OpenClaw conversation blob. Rules:
//   - If "[คำถามล่าสุด]" marker is present, return everything after it.
//   - Strip any leading metadata block (lines wrapped in ```json ... ``` that
//     the OpenClaw control UI prepends — "Sender (untrusted metadata):").
//   - Strip leading bracketed timestamps like "[Wed 2026-04-08 05:08 UTC]".
//   - Cap to 300 characters — embedding retrieval doesn't need more, and
//     RAGFlow's embedding model has a batch-size ceiling (~512 tokens) that a
//     long compacted history blows past, causing HTTP 500 "input too large".
//   - If nothing sensible is left, fall back to the original string capped.
func extractLatestQuestion(raw string) string {
	const maxChars = 300
	s := raw
	if idx := strings.LastIndex(s, "[คำถามล่าสุด]"); idx >= 0 {
		s = s[idx+len("[คำถามล่าสุด]"):]
	}
	// Strip "Sender (untrusted metadata): ```json ... ```" prefix
	if i := strings.Index(s, "```json"); i >= 0 {
		if j := strings.Index(s[i+7:], "```"); j >= 0 {
			s = s[i+7+j+3:]
		}
	}
	// Strip leading bracketed timestamp like "[Wed 2026-04-08 05:08 UTC]"
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") {
		if end := strings.Index(s, "]"); end > 0 && end < 60 {
			inside := s[1:end]
			// Heuristic: looks like a timestamp if it contains a colon or "UTC"
			if strings.Contains(inside, ":") || strings.Contains(inside, "UTC") {
				s = strings.TrimSpace(s[end+1:])
			}
		}
	}
	s = strings.TrimSpace(s)
	if s == "" {
		s = strings.TrimSpace(raw)
	}
	// Cap length — use rune-safe truncation for Thai
	runes := []rune(s)
	if len(runes) > maxChars {
		s = string(runes[:maxChars])
	}
	return s
}

func truncateString(s string, maxChars int) string {
	if maxChars <= 0 || len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "...(truncated)"
}

// dispatchAgentTool — ศูนย์กลางเรียก tool ของ agent
//
// แยกออกมาเพื่อให้ทั้ง v2 (parallel batch) และ ReAct ใช้ logic เดียวกัน:
//   - tool พิเศษที่ต้องรู้ holdingCode (เช่น queryknowledgebase) → handle inline
//   - tool ปกติ → ส่งต่อไป MCP server
func dispatchAgentTool(ctx context.Context, mcpExec func(context.Context, string, map[string]any) (any, error),
	holdingCode, toolName string, params map[string]any) (any, error) {
	if toolName == kbQueryToolName {
		return executeKBQuery(ctx, holdingCode, params)
	}
	return mcpExec(ctx, toolName, params)
}
