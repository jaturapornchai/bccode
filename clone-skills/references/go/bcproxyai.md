# bcproxyai — AI Provider Gateway

bcproxyai เป็น **OpenAI-compatible reverse proxy** ที่ route request ไปยัง upstream AI providers (OpenRouter, Groq, Cerebras, Mistral, Ollama, SambaNova) ผ่าน endpoint เดียว

## Base URL
- **Local (dev):** `http://localhost:3333/v1`
- **From Docker container:** `http://host.docker.internal:3333/v1`

## Endpoints
| Path | Method | Description |
|---|---|---|
| `/v1/chat/completions` | POST | OpenAI-compatible chat (รองรับ tools + vision) |
| `/v1/models` | GET | List available models |

## Authentication
ใช้ `Authorization: Bearer <API_KEY>` — key ถูกตั้งใน BC Account shop config (`aiProviderConfigs` collection) ไม่ใช่ใน bootstrap.json

## Virtual Models (Auto-Routing)
bcproxyai มี virtual models ที่ route อัตโนมัติไปยัง upstream ที่เสถียร/เหมาะกับงาน:

| Virtual Model | Purpose |
|---|---|
| `bcproxy/auto` | Auto-select — คุณภาพดีสุด แต่อาจช้าเมื่อ upstream busy |
| `bcproxy/fast` | Fast — ใช้ model เร็วอย่าง Cerebras/Groq |
| `bcproxy/tools` | Specialized สำหรับ tool calling (default ของ custom provider ใน BC Account) |
| `bcproxy/thai` | Optimized สำหรับภาษาไทย |
| `bcproxy/consensus` | Multi-model consensus (ช้า, คุณภาพสูงสุด) |

**สำคัญ:** `bcproxy/auto` อาจ timeout/503 เมื่อ upstream pool overload — bcproxy จะ fallback ภายในเอง ไม่ต้อง hardcode fallback list ใน Go code

## Direct Upstream Models
นอกจาก virtual models ยังเรียก upstream model ตรงๆ ได้:

- **Cerebras** (เร็วสุด): `cerebras/qwen-3-235b-a22b-instruct-2507`, `cerebras/gpt-oss-120b`, `cerebras/zai-glm-4.7`
- **Groq**: `groq/llama-3.3-70b-versatile`, `groq/moonshotai/kimi-k2-instruct-0905`, `groq/openai/gpt-oss-120b`
- **Mistral**: `mistral/mistral-large-latest`, `mistral/magistral-medium-latest`, `mistral/pixtral-large-latest`
- **OpenRouter** (free tier): `openrouter/google/gemma-3-27b-it:free`, `openrouter/meta-llama/llama-3.3-70b-instruct:free`
- **SambaNova**: `sambanova/DeepSeek-V3.2`, `sambanova/Meta-Llama-3.3-70B-Instruct`
- **Ollama** (local): `ollama/qwen2.5:7b`, `ollama/gemma3:12b`, `ollama/qwen2.5vl:7b` (vision)

ดูรายการเต็ม: `curl http://localhost:3333/v1/models`

## Response Headers
bcproxyai ส่ง custom headers กลับมาเพื่อบอกว่า route ไปที่ไหน:
- `X-BCProxy-Model` — ชื่อ actual model ที่ใช้
- `X-BCProxy-Provider` — ชื่อ upstream provider

Go code เก็บไว้ใน `oaiResp.Model` และ `oaiResp.Provider` ใน `openai_compat.go:callOnce()`

## Call Example (Go)
```go
// ใน backend — ใช้ผ่าน aiprovider package
providers := aiprovider.GetShopToolCallingProviders(shopID)
// provider name = "custom-x1", base_url = "http://host.docker.internal:3333/v1"
// model = ตามที่ shop ตั้งใน DB (default "bcproxy/tools")
resp, err := providers[0].GenerateContentWithTools(ctx, messages, tools, 0.7)
```

## Call Example (curl)
```bash
curl -s http://localhost:3333/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-placeholder" \
  -d '{
    "model": "bcproxy/tools",
    "messages": [{"role":"user","content":"hello"}],
    "temperature": 0.7
  }'
```

## Shop Config (MongoDB)
BC Account เก็บ provider config ใน collection `aiProviderConfigs`:
```js
{
  shopid: "3AEz8tu22GHPpAZ0XhwPFM4fjY9",
  providername: "custom-x1",       // custom provider name (prefix "custom")
  baseurl: "http://host.docker.internal:3333/v1",
  apikey: "sk-placeholder",         // bcproxy key
  model: "bcproxy/tools",           // virtual or upstream model
  capabilities: ["tools","vision","thinking"],
  isactive: true,
  priority: 1
}
```

API endpoints สำหรับจัดการ:
- `POST /goapi/api/v1/ai-provider/list` — list providers ของ shop
- `POST /goapi/api/v1/ai-provider/save` — upsert provider
- `POST /goapi/api/v1/ai-provider/test` — test connection
- `POST /goapi/api/v1/ai-provider/models` — list models จาก provider

## Troubleshooting

### 503 server_overloaded
```
All N models from M providers failed: ...
```
→ bcproxy upstream pool ล่ม retry logic (3 attempts + exponential backoff) จะทำงานอัตโนมัติ ถ้ายังไม่ผ่าน = bcproxy มีปัญหาจริงๆ ต้องดู bcproxy log

### Timeout (คำตอบช้ามาก)
- Cloud timeout = **60s** (ตั้งใน `openai_compat.go:newOpenAICompatProvider()`)
- Ollama local timeout = 300s
- ถ้า bcproxy/auto ช้าเกิน 60s → abort แล้วเข้า retry
- อย่าเพิ่ม timeout โดยไม่จำเป็น — upstream busy ต้อง fail fast

### HTTP 400 จาก bcproxy
- มักเกิดเมื่อ upstream model ไม่รองรับ payload เช่น ส่ง image ไป text-only model
- Go code มี `stripImageContent()` fallback ใน `agent_loop.go` แล้ว

## Key Files
- [`internal/goapi/aiprovider/openai_compat.go`](../../../backend/internal/goapi/aiprovider/openai_compat.go) — HTTP client + timeout config
- [`internal/goapi/aiprovider/shop_providers.go`](../../../backend/internal/goapi/aiprovider/shop_providers.go) — shop config loading + cooldown
- [`internal/goapi/handlers/aichat/agent_loop.go`](../../../backend/internal/goapi/handlers/aichat/agent_loop.go) — retry + fallback orchestration
