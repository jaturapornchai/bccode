# API Request: AI Provider Auto-Fallback with Cooldown

## สิ่งที่ต้องการ
ปรับ `GetProvider()` ให้ **ทดสอบ provider อัตโนมัติ** — เอาเฉพาะตัวที่มี API key แล้วลองทีละตัว ถ้าตัวไหนใช้ไม่ได้ให้ข้ามไปตัวถัดไป ตัวที่ fail ให้ cooldown 1 ชั่วโมงก่อนลองใหม่

## สถานะปัจจุบัน
- `GetProvider()` ใน `internal/goapi/aiprovider/provider.go` ใช้ `AI_PROVIDER` env var เลือก provider **เดียว**
- ถ้า provider ที่เลือกไม่มี API key → fallback ไป Gemini
- ถ้า Gemini ไม่มี API key → **error ทันที** (ไม่ลองตัวอื่น)
- ไม่มี auto-retry หรือ cooldown mechanism

## สิ่งที่ต้องทำ

### 1. สร้าง `internal/goapi/aiprovider/fallback.go` — Auto-Fallback wrapper

```go
package aiprovider

import (
    "context"
    "os"
    "sync"
    "time"
)

// providerCooldown เก็บเวลาที่ provider fail ล่าสุด
var (
    cooldownMap   = make(map[string]time.Time) // provider name → fail time
    cooldownMu    sync.RWMutex
    cooldownDuration = 1 * time.Hour
)

// providerEntry เก็บข้อมูล provider ที่พร้อมใช้
type providerEntry struct {
    name     string
    provider AIProvider
}

// markFailed บันทึกว่า provider นี้ fail — cooldown 1 ชม.
func markFailed(name string) {
    cooldownMu.Lock()
    defer cooldownMu.Unlock()
    cooldownMap[name] = time.Now()
    logger.Warn("[AIProvider] %s ถูก cooldown 1 ชั่วโมง", name)
}

// isCoolingDown ตรวจว่า provider ยัง cooldown อยู่หรือไม่
func isCoolingDown(name string) bool {
    cooldownMu.RLock()
    defer cooldownMu.RUnlock()
    failTime, exists := cooldownMap[name]
    if !exists {
        return false
    }
    if time.Since(failTime) > cooldownDuration {
        // หมด cooldown แล้ว → ลบออก
        delete(cooldownMap, name)
        return false
    }
    return true
}

// getAvailableProviders รวบรวม providers ที่มี API key + ไม่ได้ cooldown
// เรียงตาม priority: openrouter → groq → deepseek → gemini
func getAvailableProviders() []providerEntry {
    var providers []providerEntry

    // Priority order ตาม business requirement
    candidates := []struct {
        name      string
        envKey    string
        envModel  string
        defModel  string
        baseURL   string
    }{
        {"openrouter", "OPENROUTER_API_KEY", "OPENROUTER_MODEL", "google/gemini-2.0-flash-exp:free", "https://openrouter.ai/api/v1/chat/completions"},
        {"groq", "GROQ_API_KEY", "GROQ_MODEL", "llama-3.3-70b-versatile", "https://api.groq.com/openai/v1/chat/completions"},
        {"deepseek", "DEEPSEEK_API_KEY", "DEEPSEEK_MODEL", "deepseek-chat", "https://api.deepseek.com/v1/chat/completions"},
    }

    // ถ้า AI_PROVIDER ถูกตั้งค่าไว้ → ให้ตัวนั้นขึ้นก่อน
    preferred := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))

    // เพิ่ม preferred ก่อน (ถ้ามี key + ไม่ cooldown)
    for _, c := range candidates {
        if c.name == preferred {
            apiKey := os.Getenv(c.envKey)
            if apiKey != "" && !isCoolingDown(c.name) {
                model := os.Getenv(c.envModel)
                if model == "" { model = c.defModel }
                providers = append(providers, providerEntry{
                    name:     c.name,
                    provider: newOpenAICompatProvider(c.name, c.baseURL, apiKey, model),
                })
            }
            break
        }
    }

    // เพิ่มที่เหลือตาม priority (ข้าม preferred ที่เพิ่มแล้ว)
    for _, c := range candidates {
        if c.name == preferred { continue }
        apiKey := os.Getenv(c.envKey)
        if apiKey != "" && !isCoolingDown(c.name) {
            model := os.Getenv(c.envModel)
            if model == "" { model = c.defModel }
            providers = append(providers, providerEntry{
                name:     c.name,
                provider: newOpenAICompatProvider(c.name, c.baseURL, apiKey, model),
            })
        }
    }

    // Gemini เป็น fallback สุดท้าย
    geminiKey := os.Getenv("GEMINI_API_KEY")
    if geminiKey != "" && !isCoolingDown("gemini") {
        providers = append(providers, providerEntry{
            name:     "gemini",
            provider: newGeminiProvider(),
        })
    }

    return providers
}
```

### 2. สร้าง `fallbackProvider` struct — implement AIProvider interface

```go
// fallbackProvider wraps หลาย providers — ลองทีละตัว
type fallbackProvider struct{}

func (f *fallbackProvider) Name() string {
    return "auto-fallback"
}

func (f *fallbackProvider) GenerateContent(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    providers := getAvailableProviders()

    if len(providers) == 0 {
        return nil, fmt.Errorf("ไม่มี AI Provider ที่พร้อมใช้งาน — กรุณาตั้งค่า API Key ใน Setup → Integrations")
    }

    var lastErr error
    for _, p := range providers {
        logger.Info("[AIProvider] กำลังลอง %s...", p.name)

        resp, err := p.provider.GenerateContent(ctx, req)
        if err == nil {
            logger.Info("[AIProvider] ✅ %s สำเร็จ", p.name)
            return resp, nil
        }

        // fail → mark cooldown + ลองตัวถัดไป
        logger.Warn("[AIProvider] ❌ %s ล้มเหลว: %v — ลองตัวถัดไป", p.name, err)
        markFailed(p.name)
        lastErr = err
    }

    return nil, fmt.Errorf("AI Provider ทุกตัวใช้ไม่ได้: %v", lastErr)
}
```

### 3. แก้ `GetProvider()` ใน `provider.go`

```go
// GetProvider returns auto-fallback provider
// ลอง provider ที่มี API key ทีละตัวตาม priority
// ถ้าตัวไหน fail → cooldown 1 ชม. แล้วข้ามไปตัวถัดไป
func GetProvider() AIProvider {
    return &fallbackProvider{}
}
```

### 4. (Optional) เพิ่ม health endpoint — ดูสถานะ providers

```
GET /goapi/api/v1/chatbot/providers
```

Response:
```json
{
    "success": true,
    "providers": [
        {"name": "openrouter", "has_key": true, "status": "active", "model": "google/gemini-2.0-flash-exp:free"},
        {"name": "groq", "has_key": true, "status": "cooldown", "cooldown_until": "2026-03-01T15:00:00Z"},
        {"name": "deepseek", "has_key": false, "status": "no_key"},
        {"name": "gemini", "has_key": true, "status": "active", "model": "gemini-2.0-flash-exp"}
    ],
    "active_count": 2
}
```

## ไฟล์ที่ต้องแก้/สร้าง

| ไฟล์ | Action | รายละเอียด |
|------|--------|-----------|
| `internal/goapi/aiprovider/fallback.go` | **สร้างใหม่** | Auto-fallback logic + cooldown map |
| `internal/goapi/aiprovider/provider.go` | **แก้ไข** | `GetProvider()` return `&fallbackProvider{}` |
| `internal/goapi/handlers/aichat/handler_gemini.go` | ไม่ต้องแก้ | ใช้ `GetProvider()` เดิม — auto-fallback ทำงานโปร่งใส |

## Behavior สรุป

```
Request เข้ามา
  → GetProvider() return fallbackProvider
  → GenerateContent() เริ่มลอง:
    1. openrouter (มี key + ไม่ cooldown) → ลอง
       - สำเร็จ → return ✅
       - fail → markFailed("openrouter") + ลองตัวถัดไป
    2. groq (มี key + ไม่ cooldown) → ลอง
       - สำเร็จ → return ✅
       - fail → markFailed("groq") + ลองตัวถัดไป
    3. deepseek (ไม่มี key) → ข้าม
    4. gemini (มี key + ไม่ cooldown) → ลอง
       - สำเร็จ → return ✅
       - fail → return error "AI Provider ทุกตัวใช้ไม่ได้"

หลังจาก 1 ชม. → cooldown หมด → กลับมาลองอีกครั้ง
```

## Use Case
- User ตั้ง API key หลายตัว → ระบบใช้ตัวแรกที่พร้อม
- Groq rate limit → อัตโนมัติข้ามไปใช้ OpenRouter
- ทุกตัว fail → แสดง error ที่ชัดเจน
- หลัง 1 ชม. → provider ที่ fail กลับมาลองอีกครั้ง

## Note สำหรับ UnifiedAPIServer
`unified_server.go` line 107 cache `GetProvider()` ตอน init → ต้องเปลี่ยนให้เรียก `GetProvider()` ทุก request (หรือใช้ `fallbackProvider` ที่ไม่ cache ภายใน)
