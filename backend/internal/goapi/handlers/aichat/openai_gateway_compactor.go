package aichat

// LLM-based history compactor สำหรับ OpenClaw gateway
//
// แนวคิด:
//   - OpenClaw ส่ง full history ทุก request → ยิ่งคุยยิ่งใหญ่
//   - ถ้า history ยาวเกิน threshold → ยิงเข้า AI provider (ตัวเดียวกับที่ shop ใช้)
//     ให้สรุปข้อความเก่าเป็นย่อหน้าสั้น เก็บ ตัวเลข/รหัส/ชื่อ/ผลลัพธ์ tool
//   - เก็บ summary ใน in-memory cache ตาม sessionID + hash ของ messages ที่สรุปไปแล้ว
//   - คำตอบรอบหลัง: ใช้ summary + messages ใหม่หลัง cut-point เท่านั้น
//
// Cache invalidation: ถ้า hash เปลี่ยน (เช่น คนละ session) → สรุปใหม่
// TTL: 1 ชั่วโมง (ตัด session ค้าง)

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"sync"
	"time"
)

// ====== Config ======

const (
	// compactTriggerChars — ถ้า history (ไม่รวม latest) ยาวกว่านี้ → trigger compaction
	compactTriggerChars = 3000

	// compactKeepRecent — เก็บ message ล่าสุดไว้กี่ตัว (ไม่สรุป)
	compactKeepRecent = 4

	// compactSummaryMaxChars — ตัด summary ที่ AI ส่งกลับให้ไม่เกินนี้
	compactSummaryMaxChars = 1500

	// compactCacheTTL — TTL ของ summary cache
	compactCacheTTL = 1 * time.Hour

	// compactLLMTimeout — timeout เรียก AI สรุป
	compactLLMTimeout = 20 * time.Second
)

// ====== In-memory cache ======

type cachedSummary struct {
	summary      string
	coveredHash  string // hash ของ messages ที่สรุปไปแล้ว
	coveredCount int    // จำนวน messages ที่สรุปไปแล้ว (จากต้น)
	updatedAt    time.Time
}

var (
	summaryCache   = map[string]*cachedSummary{}
	summaryCacheMu sync.RWMutex
)

func getCachedSummary(sessionID string) *cachedSummary {
	summaryCacheMu.RLock()
	defer summaryCacheMu.RUnlock()
	c, ok := summaryCache[sessionID]
	if !ok {
		return nil
	}
	if time.Since(c.updatedAt) > compactCacheTTL {
		return nil
	}
	return c
}

func putCachedSummary(sessionID string, c *cachedSummary) {
	summaryCacheMu.Lock()
	defer summaryCacheMu.Unlock()
	c.updatedAt = time.Now()
	summaryCache[sessionID] = c
	// opportunistic cleanup
	if len(summaryCache) > 500 {
		cutoff := time.Now().Add(-compactCacheTTL)
		for k, v := range summaryCache {
			if v.updatedAt.Before(cutoff) {
				delete(summaryCache, k)
			}
		}
	}
}

// ====== Helpers ======

func hashMessages(msgs []openaiChatMessage) string {
	h := sha1.New()
	for _, m := range msgs {
		h.Write([]byte(m.Role))
		h.Write([]byte{0})
		h.Write([]byte(messageContentToText(m.Content)))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// hashMessagePrefix คำนวณ hash ของ messages[:n] แบบเร็ว (ไม่ alloc slice ใหม่)
func hashMessagePrefix(msgs []openaiChatMessage, n int) string {
	if n < 0 {
		n = 0
	}
	if n > len(msgs) {
		n = len(msgs)
	}
	h := sha1.New()
	for i := 0; i < n; i++ {
		h.Write([]byte(msgs[i].Role))
		h.Write([]byte{0})
		h.Write([]byte(messageContentToText(msgs[i].Content)))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func renderHistoryPlain(msgs []openaiChatMessage) string {
	var lines []string
	for _, m := range msgs {
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
		lines = append(lines, fmt.Sprintf("%s: %s", label, txt))
	}
	return strings.Join(lines, "\n")
}

func totalContentChars(msgs []openaiChatMessage) int {
	n := 0
	for _, m := range msgs {
		n += len(messageContentToText(m.Content))
	}
	return n
}

// ====== Main entry ======

// buildQuestionWithCompactedHistory สร้าง question พร้อม history ที่ compact แล้ว
//
// Strategy:
//  1. หา last user message
//  2. แบ่ง messages เป็น 2 ส่วน: olderToCompact (ก่อน) + recentRaw (หลัง compactKeepRecent ตัว) + latest
//  3. ถ้า olderToCompact ยาวเกิน threshold → ยิง AI สรุป (ใช้ cache ถ้า hash ตรงกัน)
//  4. ประกอบ:  [สรุป] + [recent raw] + [latest question]
//
// ถ้า AI provider ไม่พร้อมหรือ fail → fallback เป็น hard-truncate (เหมือนเดิม)
func buildQuestionWithCompactedHistory(ctx context.Context, holdingCode, sessionID string, messages []openaiChatMessage) string {
	// 1. หา last user message
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

	// history = ทุกอย่างก่อน latest (ไม่รวม latest)
	history := messages[:lastUserIdx]
	if len(history) == 0 {
		return latest
	}

	// ถ้า history สั้น → ไม่ต้อง compact ใช้ plain แบบเดิม
	if totalContentChars(history) < compactTriggerChars {
		return fmt.Sprintf("[ประวัติสนทนาก่อนหน้า]\n%s\n\n[คำถามล่าสุด]\n%s",
			renderHistoryPlain(history), latest)
	}

	// === Rolling window cache lookup ===
	//
	// แนวคิด: cache เก็บ summary ที่ครอบ messages[:coveredCount] พร้อม hash ของ prefix นั้น
	// ถ้ารอบถัดมา prefix เดิม hash ตรง → ใช้ summary เดิม + messages[coveredCount:lastUserIdx] เป็น recent
	// ไม่ต้อง re-summarize ทุกครั้งที่มี message ใหม่ (ซึ่งเดิมทำ hash เปลี่ยนทุกรอบ)
	//
	// เงื่อนไข re-summarize:
	//  1) ไม่มี cache
	//  2) cached.coveredHash ไม่ตรงกับ hash prefix (session ต่าง / history ถูกแก้)
	//  3) recent block ยาวเกิน compactTriggerChars → ขยาย coverage ใหม่
	var summary string
	var olderCount int // จำนวน messages ใน prefix ที่ summary ครอบอยู่
	var recent []openaiChatMessage
	var needResummarize bool

	cached := getCachedSummary(sessionID)
	if cached != nil && cached.coveredCount > 0 && cached.coveredCount <= len(history) {
		// Verify prefix hash ตรง
		prefixHash := hashMessagePrefix(messages, cached.coveredCount)
		if prefixHash == cached.coveredHash {
			// Cache HIT — ใช้ summary เดิม, recent = messages หลัง prefix ถึงก่อน latest
			recent = history[cached.coveredCount:]
			summary = cached.summary
			olderCount = cached.coveredCount

			// ถ้า recent ยาวเกิน threshold → ขยาย coverage (re-summarize)
			if totalContentChars(recent) >= compactTriggerChars {
				logger.Info("[OpenClaw Compactor] cache HIT but recent block too large (%d chars) — re-summarize session=%s",
					totalContentChars(recent), sessionID)
				needResummarize = true
			} else {
				logger.Info("[OpenClaw Compactor] rolling cache HIT session=%s coveredCount=%d recentCount=%d",
					sessionID, olderCount, len(recent))
			}
		} else {
			logger.Info("[OpenClaw Compactor] cache prefix mismatch (session=%s) — re-summarize", sessionID)
			needResummarize = true
		}
	} else {
		needResummarize = true
	}

	if needResummarize {
		// หา cut point: ทิ้ง recent ท้ายไว้ compactKeepRecent ตัว, สรุปส่วนที่เหลือ
		cut := len(history) - compactKeepRecent
		if cut < 1 {
			cut = 1
		}
		older := history[:cut]
		recent = history[cut:]

		// ยิง AI สรุป
		s, err := summarizeWithAI(ctx, holdingCode, older)
		if err != nil || strings.TrimSpace(s) == "" {
			logger.Warn("[OpenClaw Compactor] summarize failed (session=%s): %v — fallback hard truncate", sessionID, err)
			// fallback: ตัดดื้อๆ
			plain := renderHistoryPlain(history)
			const maxFallback = 4000
			if len(plain) > maxFallback {
				plain = "...(ตัดบางส่วนทิ้ง)...\n" + plain[len(plain)-maxFallback:]
			}
			return fmt.Sprintf("[ประวัติสนทนาก่อนหน้า]\n%s\n\n[คำถามล่าสุด]\n%s", plain, latest)
		}
		// ตัด summary ถ้ายาวเกิน
		if len(s) > compactSummaryMaxChars {
			s = s[:compactSummaryMaxChars] + "..."
		}
		summary = s
		olderCount = cut

		// บันทึก cache ด้วย prefix hash (ไม่ใช่ hash ของ older block แบบเดิม — ให้ prefix match ได้ใน rolling window)
		putCachedSummary(sessionID, &cachedSummary{
			summary:      summary,
			coveredHash:  hashMessagePrefix(messages, cut),
			coveredCount: cut,
		})
		logger.Info("[OpenClaw Compactor] summarized session=%s older=%d msgs → summaryLen=%d",
			sessionID, cut, len(summary))
	}

	// 5. ประกอบผลลัพธ์
	recentPlain := renderHistoryPlain(recent)
	var sb strings.Builder
	sb.WriteString("[สรุปบทสนทนาก่อนหน้า — compacted by AI]\n")
	sb.WriteString(summary)
	if recentPlain != "" {
		sb.WriteString("\n\n[ข้อความล่าสุดก่อนคำถามนี้]\n")
		sb.WriteString(recentPlain)
	}
	sb.WriteString("\n\n[คำถามล่าสุด]\n")
	sb.WriteString(latest)
	return sb.String()
}

// summarizeWithAI ยิง provider ของ shop ไปสรุป history
func summarizeWithAI(parentCtx context.Context, holdingCode string, older []openaiChatMessage) (string, error) {
	providers := aiprovider.GetShopAIProviders(holdingCode)
	if len(providers) == 0 {
		return "", fmt.Errorf("no AI providers available for shop=%s", holdingCode)
	}

	plain := renderHistoryPlain(older)
	if plain == "" {
		return "", fmt.Errorf("nothing to summarize")
	}

	// LANGUAGE RULE: prompt is English; AI replies in Thai (summary contains Thai content)
	systemPrompt := `You are a Thai conversation summarizer. Summarize the dialogue between "ผู้ใช้" (user) and "น้องกุ้ง" (AI business assistant) concisely.

Rules:
- Output in Thai. Use bullets, max 8.
- Preserve key facts: numbers, product/customer codes, proper names, query results
- Keep both the user's questions and the AI's prior answers (briefly)
- Do not invent or interpret — summarize only what is present
- No HTML / no Markdown tables — plain bullets only
- Total length must not exceed 1000 characters

CRITICAL: Reply in Thai (the dialogue is in Thai, the summary must be in Thai).`

	userPrompt := "Dialogue to summarize:\n\n" + plain

	ctx, cancel := context.WithTimeout(parentCtx, compactLLMTimeout)
	defer cancel()

	// ลอง provider ทีละตัว
	var lastErr error
	for _, p := range providers {
		resp, err := p.GenerateContent(ctx, aiprovider.ChatRequest{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			Temperature:  0.2,
			MaxTokens:    800,
		})
		if err != nil {
			logger.Warn("[OpenClaw Compactor] provider failed, trying next: %v", err)
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
	return "", fmt.Errorf("all providers returned empty")
}
