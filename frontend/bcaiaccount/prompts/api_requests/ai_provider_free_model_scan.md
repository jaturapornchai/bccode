# API Request: AI Provider Free Model Auto-Scan

## สิ่งที่ต้องการ
Backend ต้องมีระบบ scan หารายการ model ฟรีจาก provider ต่างๆ อัตโนมัติ แล้ว cache ไว้ เพื่อ:
1. เมื่อ `model` field ว่าง → backend เลือก free model ที่ใช้งานได้จาก cache ให้อัตโนมัติ
2. Frontend สามารถดึงรายการ free models ล่าสุดมาแสดงได้

## ฟีเจอร์หลัก

### 1. Auto-scan free models (background)
- Scan ครั้งแรกตอน server start
- Scan ซ้ำทุก 6 ชม. (configurable)
- Scan ใหม่เมื่อ frontend request (ถ้า cache เก่ากว่า 1 ชม.)
- Cache ผลไว้ใน memory (ไม่ต้องลง DB)

### 2. Provider-specific scan logic
| Provider | API Endpoint | Auth | Filter |
|----------|-------------|------|--------|
| OpenRouter | `GET https://openrouter.ai/api/v1/models` | ไม่ต้อง (public) | id ลงท้าย `:free` |
| Groq | `GET https://api.groq.com/openai/v1/models` | Bearer `GROQ_API_KEY` | ทุก model (ฟรีหมด), exclude whisper/guard/tool-use |
| DeepSeek | ไม่มี free tier → skip | - | - |
| Gemini | `GET https://generativelanguage.googleapis.com/v1beta/models?key=GEMINI_API_KEY` | Query param | filter models ที่ free tier ใช้ได้ |

### 3. Smart model selection ใน fallbackProvider
เมื่อ provider ถูกเลือกแต่ model ว่าง:
- ดูจาก cached free models → เลือกตัวแรกที่ available
- ถ้า cache ว่าง → ใช้ default model เดิม (defModel ใน candidate)

## Endpoint ที่ต้องการ

### GET /goapi/api/setup/ai-models
- **Auth:** Setup password (เหมือน endpoint setup อื่นๆ)
- **Description:** ดึงรายการ free models ของทุก provider (trigger scan ถ้า cache เก่า)

## Request Parameters
ไม่มี (GET request)

## Expected Response
```json
{
  "success": true,
  "data": {
    "openrouter": {
      "free_models": ["google/gemma-3n-e4b-it:free", "meta-llama/llama-3.3-70b-instruct:free", "..."],
      "last_scanned": "2026-03-02T10:30:00Z",
      "total": 25
    },
    "groq": {
      "free_models": ["llama-3.3-70b-versatile", "qwen/qwen-3-32b", "..."],
      "last_scanned": "2026-03-02T10:30:00Z",
      "total": 12
    },
    "gemini": {
      "free_models": ["gemini-2.5-flash", "gemini-2.5-flash-lite"],
      "last_scanned": "2026-03-02T10:30:00Z",
      "total": 2
    }
  }
}
```

## Use Case (Frontend)
- หน้าจอ: Setup → Integrations → AI Providers
- การใช้งาน:
  1. User เปิด checkbox "ใช้โมเดลฟรีอัตโนมัติ" → model field ว่าง → backend เลือกเอง
  2. กดปุ่ม "Scan ล่าสุด" → เรียก GET /goapi/api/setup/ai-models → แสดง preview รายการ model ฟรี
  3. User ไม่ต้องรู้ว่า model ไหนฟรี model ไหนหมดอายุ — backend จัดการให้

## Implementation Notes

### ใน `fallback.go` — แก้ `buildProvider()`:
```go
func buildProvider(c providerCandidate) *providerEntry {
    // ... existing cooldown + key check ...
    model := os.Getenv(c.envModel)
    if model == "" {
        // ลองดึงจาก cached free models
        model = getFirstFreeModel(c.name)
    }
    if model == "" {
        model = c.defModel // fallback to default
    }
    // ...
}
```

### สร้างไฟล์ใหม่ `model_scanner.go`:
```go
package aiprovider

// FreeModelCache — cache รายการ model ฟรีของแต่ละ provider
type FreeModelCache struct {
    mu          sync.RWMutex
    cache       map[string]FreeModelEntry
    scanInterval time.Duration // default 6h
    minCacheAge  time.Duration // default 1h
}

type FreeModelEntry struct {
    Models     []string
    ScannedAt  time.Time
}

// StartBackgroundScan — เริ่ม goroutine scan ทุก scanInterval
func (c *FreeModelCache) StartBackgroundScan()

// ScanAll — scan ทุก provider ที่มี key
func (c *FreeModelCache) ScanAll()

// ScanProvider — scan provider เฉพาะตัว
func (c *FreeModelCache) ScanProvider(name string) ([]string, error)

// GetFreeModels — ดึง cached models (trigger scan ถ้าเก่า)
func (c *FreeModelCache) GetFreeModels(name string) []string

// GetAllCached — return ทั้งหมด (สำหรับ API endpoint)
func (c *FreeModelCache) GetAllCached() map[string]FreeModelEntry
```

## Database Tables ที่เกี่ยวข้อง
ไม่ต้องใช้ DB — cache ใน memory เพียงพอ (scan ใหม่ทุกครั้งที่ restart)

## Priority
Medium — frontend ทำงานได้แล้วโดยใช้ default model, แต่ feature นี้จะทำให้ระบบ resilient ขึ้นเมื่อ provider เปลี่ยน model list
