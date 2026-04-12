package aiprovider

import (
	"context"
	"fmt"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"sync"
	"time"
)

// cooldown state — provider ที่ fail จะถูก cooldown
var (
	cooldownMap       = make(map[string]cooldownEntry)
	cooldownMu        sync.RWMutex
	defaultCooldown   = 5 * time.Minute // config ผิด, 401, 404 → cooldown moderate (เดิม 1h — มากเกินเมื่อ upstream flap)
	rateLimitCooldown = 1 * time.Minute // 429 / 503 server_overloaded / transient → cooldown สั้น
)

type cooldownEntry struct {
	failTime time.Time
	duration time.Duration
}

// providerCandidate กำหนด provider ที่รองรับ
type providerCandidate struct {
	name     string
	envKey   string
	envModel string
	defModel string
	baseURL  string
}

// candidates ทั้งหมด (ไม่รวม Gemini ซึ่งเป็น fallback สุดท้าย)
var candidates = []providerCandidate{
	{"openrouter", "OPENROUTER_API_KEY", "OPENROUTER_MODEL", "google/gemini-2.0-flash-exp:free", "https://openrouter.ai/api/v1/chat/completions"},
	{"groq", "GROQ_API_KEY", "GROQ_MODEL", "llama-3.3-70b-versatile", "https://api.groq.com/openai/v1/chat/completions"},
	{"deepseek", "DEEPSEEK_API_KEY", "DEEPSEEK_MODEL", "deepseek-chat", "https://api.deepseek.com/v1/chat/completions"},
}

// markFailed บันทึกว่า provider fail → cooldown ตาม duration
func markFailed(name string, duration time.Duration) {
	cooldownMu.Lock()
	defer cooldownMu.Unlock()
	cooldownMap[name] = cooldownEntry{failTime: time.Now(), duration: duration}
	logger.Warn("[AIProvider] %s ถูก cooldown %v", name, duration)
}

// isRateLimitError ตรวจว่า error เป็น transient failure ที่จะหายเองในไม่กี่นาที
// รวมทั้ง 429 rate limit, 503 server overloaded, 500 timeout, service capacity exceeded, etc.
// เหตุผล: cooldown ยาว (1h) ไม่เหมาะกับ upstream ที่ flap — ใช้ cooldown สั้น (1min) ให้มีโอกาสรีทรายเร็วๆ
func isRateLimitError(err error) bool {
	s := err.Error()
	return strings.Contains(s, "status 429") ||
		strings.Contains(s, "status 503") ||
		strings.Contains(s, "status 500") ||
		strings.Contains(s, "server_overloaded") ||
		strings.Contains(s, "rate limit") ||
		strings.Contains(s, "rate_limit") ||
		strings.Contains(s, "TimeoutError") ||
		strings.Contains(s, "service_tier_capacity_exceeded") ||
		strings.Contains(s, "capacity exceeded")
}

// isCoolingDown ตรวจว่า provider ยัง cooldown อยู่หรือไม่
func isCoolingDown(name string) bool {
	cooldownMu.RLock()
	entry, exists := cooldownMap[name]
	cooldownMu.RUnlock()
	if !exists {
		return false
	}
	if time.Since(entry.failTime) > entry.duration {
		// หมด cooldown → ลบออก
		cooldownMu.Lock()
		delete(cooldownMap, name)
		cooldownMu.Unlock()
		return false
	}
	return true
}

// providerEntry — provider ที่พร้อมใช้
type providerEntry struct {
	name     string
	provider AIProvider
}

// getAvailableProviders รวบรวม providers ที่มี API key + ไม่ได้ cooldown
// เรียงตาม priority: preferred (จาก AI_PROVIDER) → ที่เหลือตามลำดับ → gemini
func getAvailableProviders() []providerEntry {
	var providers []providerEntry

	preferred := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))

	// เพิ่ม preferred ก่อน
	for _, c := range candidates {
		if c.name == preferred {
			if p := buildProvider(c); p != nil {
				providers = append(providers, *p)
			}
			break
		}
	}

	// เพิ่มที่เหลือตาม priority (ข้าม preferred ที่เพิ่มแล้ว)
	for _, c := range candidates {
		if c.name == preferred {
			continue
		}
		if p := buildProvider(c); p != nil {
			providers = append(providers, *p)
		}
	}

	// Gemini เป็น fallback สุดท้าย
	if !isCoolingDown("gemini") {
		geminiKey := os.Getenv("GEMINI_API_KEY")
		if geminiKey != "" {
			providers = append(providers, providerEntry{
				name:     "gemini",
				provider: newGeminiProvider(),
			})
		}
	}

	return providers
}

// buildProvider สร้าง provider จาก candidate (ถ้ามี key + ไม่ cooldown)
func buildProvider(c providerCandidate) *providerEntry {
	if isCoolingDown(c.name) {
		return nil
	}
	apiKey := os.Getenv(c.envKey)
	if apiKey == "" {
		return nil
	}
	model := os.Getenv(c.envModel)
	if model == "" {
		model = c.defModel
	}
	return &providerEntry{
		name:     c.name,
		provider: newOpenAICompatProvider(c.name, c.baseURL, apiKey, model),
	}
}

// fallbackProvider ลอง providers ทีละตัว — ตัวไหน fail ก็ข้ามไปตัวถัดไป
type fallbackProvider struct{}

func (f *fallbackProvider) Name() string {
	return "auto-fallback"
}

func (f *fallbackProvider) GenerateContent(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	providers := getAvailableProviders()

	if len(providers) == 0 {
		return nil, fmt.Errorf("ไม่มี AI Provider ที่พร้อมใช้งาน — กรุณาตั้งค่า API Key")
	}

	var lastErr error
	for _, p := range providers {
		logger.Info("[AIProvider] กำลังลอง %s...", p.name)

		resp, err := p.provider.GenerateContent(ctx, req)
		if err == nil {
			logger.Info("[AIProvider] %s สำเร็จ", p.name)
			if resp.Model == "" {
				resp.Model = p.name
			}
			return resp, nil
		}

		// เลือก cooldown duration ตาม error type
		cd := defaultCooldown
		if isRateLimitError(err) {
			cd = rateLimitCooldown
		}

		logger.Warn("[AIProvider] %s ล้มเหลว: %v", p.name, err)
		markFailed(p.name, cd)
		lastErr = err
	}

	return nil, fmt.Errorf("AI Provider ทุกตัวใช้ไม่ได้: %v", lastErr)
}

// ProviderStatus สถานะของแต่ละ provider (สำหรับ health endpoint)
type ProviderStatus struct {
	Name          string  `json:"name"`
	HasKey        bool    `json:"has_key"`
	Status        string  `json:"status"` // "active", "cooldown", "no_key"
	Model         string  `json:"model,omitempty"`
	CooldownUntil *string `json:"cooldown_until,omitempty"`
}

// GetProviderStatuses คืนสถานะทุก provider (สำหรับ health/debug endpoint)
func GetProviderStatuses() []ProviderStatus {
	var statuses []ProviderStatus

	for _, c := range candidates {
		s := ProviderStatus{Name: c.name}
		apiKey := os.Getenv(c.envKey)
		s.HasKey = apiKey != ""

		if !s.HasKey {
			s.Status = "no_key"
		} else if isCoolingDown(c.name) {
			s.Status = "cooldown"
			cooldownMu.RLock()
			if entry, ok := cooldownMap[c.name]; ok {
				until := entry.failTime.Add(entry.duration).Format(time.RFC3339)
				s.CooldownUntil = &until
			}
			cooldownMu.RUnlock()
		} else {
			s.Status = "active"
			model := os.Getenv(c.envModel)
			if model == "" {
				model = c.defModel
			}
			s.Model = model
		}
		statuses = append(statuses, s)
	}

	// Gemini
	geminiStatus := ProviderStatus{Name: "gemini"}
	geminiKey := os.Getenv("GEMINI_API_KEY")
	geminiStatus.HasKey = geminiKey != ""
	if !geminiStatus.HasKey {
		geminiStatus.Status = "no_key"
	} else if isCoolingDown("gemini") {
		geminiStatus.Status = "cooldown"
		cooldownMu.RLock()
		if entry, ok := cooldownMap["gemini"]; ok {
			until := entry.failTime.Add(entry.duration).Format(time.RFC3339)
			geminiStatus.CooldownUntil = &until
		}
		cooldownMu.RUnlock()
	} else {
		geminiStatus.Status = "active"
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = "gemini-2.0-flash-exp"
		}
		geminiStatus.Model = model
	}
	statuses = append(statuses, geminiStatus)

	return statuses
}
