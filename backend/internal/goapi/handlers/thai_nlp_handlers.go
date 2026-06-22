package handlers

import (
	"net/http"
	"os"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

// ThaiNLPClient - HTTP client for Thai NLP service (ลด timeout เหลือ 500ms)
var thaiNLPClient = &http.Client{
	Timeout: 500 * time.Millisecond,
}

// Circuit breaker สำหรับ Thai NLP service
// ถ้า service ไม่ตอบ จะข้ามไปเลย 60 วินาที ไม่ต้องรอ timeout ซ้ำ
var (
	nlpCircuitOpen  bool
	nlpCircuitUntil time.Time
	nlpCircuitMu    sync.RWMutex
)

// nlpIsAvailable - ตรวจสอบว่า NLP service ใช้ได้หรือถูก circuit break อยู่
func nlpIsAvailable() bool {
	nlpCircuitMu.RLock()
	defer nlpCircuitMu.RUnlock()
	if !nlpCircuitOpen {
		return true
	}
	return time.Now().After(nlpCircuitUntil)
}

// nlpMarkFailed - บันทึกว่า NLP service ล้มเหลว → เปิด circuit breaker 60 วินาที
func nlpMarkFailed() {
	nlpCircuitMu.Lock()
	defer nlpCircuitMu.Unlock()
	nlpCircuitOpen = true
	nlpCircuitUntil = time.Now().Add(60 * time.Second)
	logger.Info("[ThaiNLP] Circuit breaker OPEN — skip NLP for 60s")
}

// nlpMarkSuccess - NLP service กลับมาทำงานได้ → ปิด circuit breaker
func nlpMarkSuccess() {
	nlpCircuitMu.Lock()
	defer nlpCircuitMu.Unlock()
	if nlpCircuitOpen {
		logger.Info("[ThaiNLP] Circuit breaker CLOSED — NLP service recovered")
	}
	nlpCircuitOpen = false
}

// getThaiNLPURL - Get Thai NLP service URL from environment
func getThaiNLPURL() string {
	url := os.Getenv("THAI_NLP_URL")
	if url == "" {
		url = "http://thai-nlp:8890" // Default Docker service name
	}
	return url
}

// TokenizeRequest - Request body for tokenization
type TokenizeRequest struct {
	Text string `json:"text"`
}

// TokenizeResponse - Response from PyThaiNLP service
type TokenizeResponse struct {
	Status string   `json:"status"`
	Tokens []string `json:"tokens"`
}
