# Jead Skill — Global AI Rules

## Who is Jead?
ดู `identity.md` สำหรับตัวตนและสไตล์ของ Jead
AI ทุกตัวที่ใช้ skill นี้ต้องปฏิบัติตาม identity.md เสมอ

## AI Behavior Rules (สำคัญมาก)

### อ่าน skill ใหม่เสมอ ก่อนทำอะไร
**ทุกครั้งที่เริ่ม session ใหม่ หรือเปลี่ยน task → อ่าน bcai-claude-skills ใหม่ก่อนเสมอ**
เพราะ skill อาจถูก update จาก session อื่น / AI ตัวอื่น / Jead แก้เอง

1. **อ่าน `identity.md`** — เข้าใจ Jead + Lessons Learned ล่าสุด
2. **อ่าน `CLAUDE.md`** — กฏ + rules ล่าสุด
3. **อ่าน rules ที่เกี่ยวข้อง** — ตาม project/task ที่ทำงานอยู่
4. **อ่าน references ที่เกี่ยวข้อง** — domain knowledge + code templates
5. **สื่อสารภาษาไทย** — ตอบภาษาไทยเสมอ
6. **ลงมือทำเลย** — ถ้าทำได้ อย่าถามมาก

### Personality ที่ AI ต้องแสดง
- **กระชับ ตรงประเด็น** — ไม่ยืดเยื้อ ไม่ต้องอธิบายเยอะ
- **Action-oriented** — ลงมือทำก่อน ค่อยอธิบายทีหลัง
- **Pragmatic** — ทำให้ work ก่อน สวยทีหลัง
- **Proactive** — เสนอสิ่งที่ควรทำถัดไป ไม่รอให้ถาม

### ต้องทำทุกครั้งเมื่อจบ session
1. **สรุปงาน** — ทำอะไรไปบ้าง
2. **ตรวจ skill update** — มีอะไรใหม่ที่ควรจำ?
3. **update bcai-claude-skills** — ถ้ามีสิ่งใหม่ให้จำ เขียนลงทันที

## Directory Structure

```
bcai-claude-skills/
├── CLAUDE.md              ← ไฟล์นี้ — global rules
├── identity.md            ← ตัวตน/สไตล์ของ Jead
├── rules/
│   ├── coding-style.md    ← Go + Flutter conventions
│   ├── mcp-first.md       ← MCP-First development rules
│   ├── erp-conventions.md ← ERP system conventions
│   ├── security.md        ← Security rules
│   ├── workflow.md        ← Development workflow + deploy pipeline
│   └── manual-writing.md  ← กฏการเขียนคู่มือ
├── skills/                ← Slash commands (user-invocable)
│   ├── api-search/SKILL.md
│   ├── api-spec/SKILL.md
│   ├── model-gen/SKILL.md
│   ├── enum-list/SKILL.md
│   ├── mcp-check/SKILL.md
│   └── master-data/SKILL.md
├── references/            ← Domain knowledge + code templates
│   ├── core/
│   │   ├── business-rules.md     ← Validation, checklist, common mistakes
│   │   ├── document-flows.md     ← State machine ทุก document type
│   │   ├── system-flow.md        ← System overview (Mermaid diagrams)
│   │   ├── database-schema.md    ← PostgreSQL/MongoDB/ClickHouse tables
│   │   ├── schema-migration.md   ← กฎเมื่อ DB schema เปลี่ยน
│   │   └── cross-system-workflow.md ← Frontend ↔ Backend workflow
│   ├── flutter/
│   │   ├── bloc-pattern.md       ← BLoC code templates เต็มรูปแบบ
│   │   ├── api-client.md         ← Dio, ApiResponse, GoAPI URLs
│   │   └── ui-components.md      ← UI patterns, status badges, forms
│   └── go/
│       ├── handlers.md           ← Handler templates (list/save/approve)
│       ├── database-queries.md   ← SQL patterns (PG/Mongo/ClickHouse)
│       └── mcp-tools.md          ← วิธีสร้าง + register MCP tools
└── docs/
    └── mcp-tools-guide.md        ← MCP tools reference (41 tools)
```

## Rules (ต้องปฏิบัติตามทุกข้อ)

### 1. ภาษาสื่อสาร
- **สื่อสารกับ Jead เป็นภาษาไทยเสมอ**
- Code comments/logs ใช้ภาษาไทยได้
- Variable names, function names ใช้ภาษาอังกฤษ

### 2. Coding Style
ดู `rules/coding-style.md`

### 3. MCP-First
ดู `rules/mcp-first.md`

### 4. ERP Conventions
ดู `rules/erp-conventions.md`

### 5. Security
ดู `rules/security.md`

### 6. Workflow & Deploy
ดู `rules/workflow.md`

### 7. Manual Writing (การเขียนคู่มือ)
ดู `rules/manual-writing.md`

### 8. การทำงานข้าม Project
- **Backend (Go)** `D:\bcdev\backend` — อ่าน+แก้ได้
- **Frontend (Flutter)** `D:\bcdev\frontend\bcaiaccount` — อ่าน+แก้ได้
- **สามารถตรวจ source code ข้ามไปมาได้** และแก้ไขได้ทั้งคู่ถ้าจำเป็น
- **bcai-claude-skills** `D:\bcdev\bcai-claude-skills` — อ่าน+update ได้ (เป็นหน้าที่ของ AI)
- **ระวังโครงสร้างข้อมูล** — แก้ฝั่งหนึ่งต้องตรวจอีกฝั่ง (ดู `rules/erp-conventions.md`)

### 9. References — ต้องอ่านก่อนเริ่มงาน

| งานที่ทำ | อ่าน references |
|---------|----------------|
| **ทุกงาน** | `references/core/business-rules.md` — checklist + common mistakes |
| **งานเอกสาร (PO/SO/Invoice)** | `references/core/document-flows.md` — state transitions |
| **เชื่อมต่อ frontend-backend** | `references/core/cross-system-workflow.md` — MCP-first, API spec |
| **DB schema เปลี่ยน** | `references/core/schema-migration.md` — migration checklist |
| **Flutter frontend** | `references/flutter/bloc-pattern.md`, `api-client.md`, `ui-components.md` |
| **Go backend** | `references/go/handlers.md`, `database-queries.md`, `mcp-tools.md` |
| **ดู DB structure** | `references/core/database-schema.md` — tables + columns |
| **ดู MCP tools** | `docs/mcp-tools-guide.md` — 41 tools reference |

### 10. สรุปท้ายงานทุกครั้ง
เมื่อทำงานเสร็จ ต้องสรุป:
- ทำอะไรไปบ้าง
- ต้องแก้ Backend หรือไม่
- มี skill/rule ที่ควร update หรือไม่

## Auto-Update Rule (สำคัญมาก)

**AI ต้อง update skill เสมอ** เมื่อเรียนรู้สิ่งใหม่จาก Jead:

1. **เมื่อ Jead บอกว่า "จำไว้" / "remember" / "ตั้งกฏ"** → เขียนลง bcai-claude-skills ทันที
2. **เมื่อพบ pattern ที่ Jead ใช้ซ้ำ** → เพิ่มเข้า rules/ หรือ references/
3. **เมื่อ Jead แก้ไข AI** (เช่น "ไม่ใช่แบบนี้") → update rule ที่เกี่ยวข้อง
4. **ทุกครั้งที่ทำงานจบ** → ตรวจสอบว่ามี skill/rule ใหม่ที่ควร update หรือไม่

### วิธี Update
- **เพิ่ม rule ใหม่**: เขียนไฟล์ใน `rules/` แล้ว update `CLAUDE.md`
- **เพิ่ม skill ใหม่**: สร้าง folder ใน `skills/{name}/SKILL.md`
- **เพิ่ม reference ใหม่**: เขียนใน `references/{domain}/`
- **แก้ไข identity**: update `identity.md`
- **ทุกการ update**: ต้องแจ้ง Jead ว่า update อะไร

### Skill Evolution

| เหตุการณ์ | สิ่งที่ต้องทำ |
|----------|-------------|
| แก้ bug สำเร็จ | เพิ่ม pattern/lesson ใน `identity.md` → Lessons Learned |
| พบ pitfall ใหม่ | เพิ่มใน `rules/coding-style.md` หรือ rule ที่เกี่ยวข้อง |
| Jead สอนวิธีใหม่ | update rule/identity ที่เกี่ยวข้อง |
| ใช้ MCP tool ใหม่ | update `docs/mcp-tools-guide.md` |
| สร้าง workflow ใหม่ | สร้าง skill ใหม่ใน `skills/` |
| Jead บอก "จำไว้" | เขียนลง identity.md หรือ rules/ ทันที |
| พบ code pattern ใหม่ | เพิ่มใน `references/{domain}/` |

**เป้าหมาย:** ยิ่งทำงานกับ Jead มาก → bcai-claude-skills ยิ่งฉลาดขึ้น → AI ตัวใหม่เข้าใจ Jead เร็วขึ้น

## How to Use (สำหรับทุก Project)

### วิธีเชื่อม bcai-claude-skills เข้า project
เพิ่มบรรทัดนี้ใน CLAUDE.md ของแต่ละ project:

```markdown
## Jead Skill Library
อ่านและปฏิบัติตาม rules ทั้งหมดใน `D:\bcdev\bcai-claude-skills\`:
- `CLAUDE.md` — global rules
- `identity.md` — ตัวตนและสไตล์ (ต้องปฏิบัติตามเสมอ)
- `rules/*.md` — กฏการทำงานทั้งหมด
- `references/**/*.md` — domain knowledge + code templates
- `skills/*/SKILL.md` — slash commands ที่ใช้ได้
```

### Supported AI Tools
- Claude Code (CLI + VSCode Extension)
- Cursor
- Claude Desktop (via MCP)
- Windsurf
- ทุก AI ที่อ่าน CLAUDE.md หรือ rules file ได้
