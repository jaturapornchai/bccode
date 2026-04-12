package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"
)

// BCProxyComplaint — ร้องเรียน model ผ่าน bcproxyai POST /api/complaint
type BCProxyComplaint struct {
	ModelID          string `json:"model_id"`
	Category         string `json:"category"`                    // wrong_answer, gibberish, wrong_language, refused, hallucination, too_short, irrelevant
	Reason           string `json:"reason,omitempty"`             // เหตุผลเพิ่มเติม
	UserMessage      string `json:"user_message,omitempty"`       // คำถามของผู้ใช้
	AssistantMessage string `json:"assistant_message,omitempty"`  // คำตอบของ AI
	Source           string `json:"source,omitempty"`             // "api" or "auto"
}

// ComplaintCategory constants
const (
	ComplaintWrongAnswer   = "wrong_answer"
	ComplaintGibberish     = "gibberish"
	ComplaintWrongLanguage = "wrong_language"
	ComplaintRefused       = "refused"
	ComplaintHallucination = "hallucination"
	ComplaintTooShort      = "too_short"
	ComplaintIrrelevant    = "irrelevant"
)

// SendBCProxyComplaint — ส่งร้องเรียนไป bcproxyai
func SendBCProxyComplaint(baseURL string, complaint BCProxyComplaint) error {
	if baseURL == "" || complaint.ModelID == "" {
		return fmt.Errorf("baseURL and model_id are required")
	}

	// ตัด /v1/chat/completions ออก ให้เหลือแค่ base
	complaintURL := strings.TrimRight(baseURL, "/")
	if idx := strings.Index(complaintURL, "/v1"); idx > 0 {
		complaintURL = complaintURL[:idx]
	}
	complaintURL += "/api/complaint"

	jsonData, err := json.Marshal(complaint)
	if err != nil {
		return fmt.Errorf("marshal complaint: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", complaintURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create complaint request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send complaint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 512)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("complaint API status %d: %s", resp.StatusCode, string(body[:n]))
	}

	logger.Info("[BCProxy] Complaint sent: model=%s category=%s reason=%s", complaint.ModelID, complaint.Category, complaint.Reason)
	return nil
}

// AutoAnalyzeQuality — วิเคราะห์คุณภาพคำตอบครบทุกมิติ
// คืน (category, reason) ถ้าพบปัญหา หรือ ("", "") ถ้าผ่าน
func AutoAnalyzeQuality(question, answer string, usedTools bool) (category string, reason string) {
	answerLen := len([]rune(answer))

	// 1. ตอบสั้นเกินไป
	if answerLen < 15 {
		return ComplaintTooShort, fmt.Sprintf("คำตอบสั้นเกินไป (%d ตัวอักษร)", answerLen)
	}

	// 2. ตรวจภาษา — ควรตอบไทย (ถ้าถามเป็นไทย)
	thaiInQ := countThai(question)
	thaiInA := countThai(answer)
	if thaiInQ > 3 && thaiInA == 0 && answerLen > 50 {
		return ComplaintWrongLanguage, "ถามเป็นภาษาไทยแต่ตอบเป็นภาษาอังกฤษ"
	}

	// 3. ตรวจ gibberish — ตัวอักษรแปลกๆ มากเกิน
	nonPrintable := 0
	sample := []rune(answer)
	if len(sample) > 200 {
		sample = sample[:200]
	}
	for _, r := range sample {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			nonPrintable++
		}
	}
	if len(sample) > 0 && float64(nonPrintable)/float64(len(sample)) > 0.3 {
		return ComplaintGibberish, "คำตอบมีตัวอักษรแปลกๆ มากเกิน"
	}

	// 4. ตรวจว่าปฏิเสธไม่ตอบ
	refusalPhrases := []string{
		"ไม่สามารถตอบ", "ไม่สามารถช่วย", "I cannot", "I can't",
		"I'm sorry, I", "ขออภัย ไม่สามารถ", "ฉันไม่สามารถ",
	}
	lowerAnswer := strings.ToLower(answer)
	for _, phrase := range refusalPhrases {
		if strings.Contains(lowerAnswer, strings.ToLower(phrase)) {
			return ComplaintRefused, "AI ปฏิเสธไม่ยอมตอบคำถาม"
		}
	}

	// 5. ใช้ tools แล้วแต่ตอบสั้น (น่าจะมีข้อมูลเยอะกว่านี้)
	if usedTools && answerLen < 80 {
		return ComplaintTooShort, fmt.Sprintf("ดึงข้อมูลจาก tools แล้วแต่ตอบแค่ %d ตัวอักษร", answerLen)
	}

	return "", ""
}

// countThai — นับจำนวนอักษรไทย
func countThai(text string) int {
	count := 0
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F {
			count++
		}
	}
	return count
}
