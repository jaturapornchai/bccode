# Cross-System Workflow — Frontend <-> Backend

## Architecture Overview

```
┌─────────────────────┐   REST API (Dio)   ┌──────────────────────┐
│   Flutter App        │ <───────────────>  │   Backend (Go)        │
│   (bcaiaccount)      │   port 8888        │   MainAPI + GoAPI     │
│                      │                    │                        │
│  AI ทำงานที่นี่       │   MCP              │  AI อ่าน+แก้ได้       │
│  อ่าน/แก้ code ได้   │ <───────────────>  │                        │
│                      │  General: /mcp/    │                        │
│                      │  Dev: /mcp/dev/    │                        │
└─────────────────────┘                    └──────────────────────┘
          │
          │ ถ้า API ไม่มี
          ▼
┌─────────────────────┐
│ prompts/api_requests/ │  ส่งให้ทีม backend
│ {feature}.md          │  ────────────────>  backend สร้าง API
└─────────────────────┘
```

## GoAPI URL Mapping

GoAPI รวมเข้า MainAPI (port 8888) แล้ว — ใช้ prefix `/goapi`

| เดิม | ใหม่ |
|------|------|
| `goapi:9091/api/health` | `localhost:8888/goapi/api/health` |
| `goapi:9091/get` | `localhost:8888/goapi/get` |
| `goapi:9091/genpdf` | `localhost:8888/goapi/genpdf` |
| `goapi:9091/mcp/sse` | `localhost:8888/goapi/mcp/sse` (General) |
| `goapi:9091/mcp/dev/sse` | `localhost:8888/goapi/mcp/dev/sse` (Dev) |

**MainAPI routes (ไม่มี /goapi prefix):** `/login`, `/register`, `/shop`, `/product/*`, `/healthz`

## MCP-First Rule (สำหรับ Flutter dev)

**ก่อนเขียนโค้ดทุกครั้ง** ให้ใช้ MCP ตรวจสอบว่า API มีอยู่แล้วไหม

### MCP มี 2 ชุด endpoint:

| Endpoint | URL | เห็น Tools |
|----------|-----|-----------|
| **General** | `/goapi/mcp/sse` | เฉพาะ business data |
| **Dev** | `/goapi/mcp/dev/sse` | ทุก tools (General + database + API catalog) |

- Frontend dev → ใช้ General (`/goapi/mcp/sse`)
- Backend dev → ใช้ Dev (`/goapi/mcp/dev/sse`)
- Auth: `x-api-key` header

### ลำดับการหาข้อมูล
1. **ใช้ MCP tools ก่อนเสมอ** — เร็วกว่า, ได้ข้อมูล runtime จริง
2. ถ้า MCP ไม่เพียงพอ → อ่าน backend source code ได้
3. ห้ามสร้าง Dart model โดยเดาจาก field names

### เมื่อต้องสร้าง Model ใหม่
```
1. ใช้ MCP (Dev): get_database_schema → ดู columns + types
2. ใช้ MCP (Dev): get_table_sample → ดูข้อมูลตัวอย่าง
3. สร้าง Dart model ตาม schema จริง
```

## เมื่อ API ที่ต้องการยังไม่มี

**ห้ามสร้าง API เอง** — ต้องเขียน spec ส่งให้ backend

### API Specification Prompt Format

บันทึกไว้ที่ `bcaiaccount/prompts/api_requests/{feature}.md`:

```markdown
# API Request: {ชื่อ feature}

## สิ่งที่ต้องการ
{อธิบายว่าต้องการ API อะไร ทำไม}

## Endpoint
- Method: POST
- Path: /goapi/api/{resource}/{action}
- Auth: Bearer token (JWT)

## Request Body
{json schema}

## Expected Response
{json schema}

## Use Case
{หน้าจอไหน ทำอะไร}

## Tables ที่เกี่ยวข้อง
{จาก MCP get_database_schema}
```

## Security Rules (บังคับ)

- **ห้าม hardcode AI API key** ใน Flutter — ต้องผ่าน backend เสมอ
- ห้าม Flutter เรียก `api.groq.com`, `openrouter.ai`, `generativelanguage.googleapis.com` โดยตรง
- ห้าม Flutter เรียก MCP endpoint โดยตรง — MCP สำหรับ AI dev tools เท่านั้น

## สรุปท้ายงาน (บังคับทุกครั้ง)

หลังทำงาน feature หรือ fix bug เสร็จ ต้องบอก user เสมอว่า:
- **ไม่ต้องแก้ Backend** — frontend แก้ครบ + backend มี API/fields พร้อมแล้ว
- **ต้องแก้ Backend** — API ยังไม่มี หรือ backend ต้องเพิ่ม field/logic → spec อยู่ที่ `prompts/api_requests/{feature}.md`
