# Jead Skill — Global AI Rules

## Who is Jead?
See `identity.md` for Jead's identity and style.
All AI agents using this skill must follow `identity.md` at all times.

## AI Behavior Rules (Critical)

### Re-read skills before every task
**At every new session or task switch, re-read clone-skills first.**
Skills may have been updated by another session, another AI, or Jead directly.

1. **Read `identity.md`** — understand Jead + latest Lessons Learned
2. **Read `CLAUDE.md`** — latest rules
3. **Read relevant `rules/`** — based on current project/task
4. **Read relevant `references/`** — domain knowledge + code templates
5. **Communicate in Thai** — always respond in Thai
6. **Act immediately** — if you can do it, don't ask

### Required AI Personality
- **Concise and direct** — no verbose explanations
- **Action-oriented** — do first, explain later
- **Pragmatic** — make it work first, polish later
- **Proactive** — suggest next steps without being asked

### End-of-session checklist
1. **Summarize work** — what was done
2. **Check for skill updates** — anything new worth remembering?
3. **Update clone-skills** — write new learnings immediately

## Directory Structure

```
clone-skills/
├── CLAUDE.md              ← This file — global rules
├── identity.md            ← Jead's identity/style
├── rules/
│   ├── coding-style.md    ← Go + Flutter conventions
│   ├── mcp-first.md       ← MCP-First development rules
│   ├── erp-conventions.md ← ERP system conventions
│   ├── security.md        ← Security rules
│   ├── workflow.md        ← Development workflow + deploy pipeline
│   └── manual-writing.md  ← Documentation writing rules
├── skills/                ← Slash commands (user-invocable)
│   ├── accounting/SKILL.md       ← Accounting / GL / Journal
│   ├── api-search/SKILL.md
│   ├── api-spec/SKILL.md
│   ├── approval-workflow/SKILL.md ← Approval workflow
│   ├── auto-packing/SKILL.md    ← Pick&Pack, BOM, auto packing, shipping label
│   ├── coupon-promotion/SKILL.md ← Coupon & Promotion
│   ├── creditor-debtor/SKILL.md  ← Creditor / Debtor management
│   ├── dashboard-report/SKILL.md ← Dashboard & Reports
│   ├── data-list/SKILL.md
│   ├── date-time-picker/SKILL.md
│   ├── enum-list/SKILL.md
│   ├── formdesign/SKILL.md
│   ├── import-export/SKILL.md    ← Import / Export data
│   ├── lineoa-chatbot/SKILL.md   ← LINE OA Chatbot
│   ├── master-data/SKILL.md
│   ├── mcp-check/SKILL.md
│   ├── model-gen/SKILL.md
│   ├── payment/SKILL.md          ← Payment / PromptPay / QR
│   ├── procurement/SKILL.md      ← PR/RFQ/PO feature tracker + competitive analysis
│   ├── product/SKILL.md          ← Product CRUD, Barcode, BOM, Options
│   ├── quotation/SKILL.md        ← Quotation / ใบเสนอราคา
│   ├── restaurant-pos/SKILL.md   ← Restaurant POS, Zone, Table, Kitchen
│   ├── sales-transaction/SKILL.md ← Sales/Purchase 40+ doc types
│   ├── stock-inventory/SKILL.md  ← Warehouse, Stock, Transfer, Costing
│   ├── tab-focus/SKILL.md
│   └── theming/SKILL.md
├── references/            ← Domain knowledge + code templates
│   ├── core/
│   │   ├── business-rules.md     ← Validation, checklist, common mistakes
│   │   ├── document-flows.md     ← State machine for every document type
│   │   ├── system-flow.md        ← System overview (Mermaid diagrams)
│   │   ├── database-schema.md    ← PostgreSQL/MongoDB/ClickHouse tables
│   │   ├── schema-migration.md   ← Rules for DB schema changes
│   │   ├── cross-system-workflow.md ← Frontend <-> Backend workflow
│   │   └── procurement-flow.md   ← PR -> RFQ -> PO flow (API, DB, Kafka, Approval)
│   ├── flutter/
│   │   ├── bloc-pattern.md       ← Full BLoC code templates
│   │   ├── api-client.md         ← Dio, ApiResponse, GoAPI URLs
│   │   └── ui-components.md      ← UI patterns, status badges, forms
│   └── go/
│       ├── handlers.md           ← Handler templates (list/save/approve)
│       ├── database-queries.md   ← SQL patterns (PG/Mongo/ClickHouse)
│       ├── mcp-tools.md          ← How to create + register MCP tools
│       └── inventory-costing.md  ← Inventory Costing System (7 methods)
└── docs/
    └── mcp-tools-guide.md        ← MCP tools reference (41 tools)
```

## Rules (mandatory)

### 1. Communication Language
- **Always communicate with Jead in Thai**
- Code comments/logs may use Thai
- Variable names and function names must use English

### 2. Coding Style
See `rules/coding-style.md`

### 3. MCP-First
See `rules/mcp-first.md`

### 4. ERP Conventions
See `rules/erp-conventions.md`

### 5. Security
See `rules/security.md`

### 6. Workflow & Deploy
See `rules/workflow.md`

### 7. Manual Writing
See `rules/manual-writing.md`

### 8. Cross-Project Work
- **Backend (Go)** `D:\bcdev\backend` — read + edit allowed
- **Frontend (Flutter)** `D:\bcdev\frontend\bcaiaccount` — read + edit allowed
- **clone-skills** `D:\bcdev\clone-skills` — read + update (AI's responsibility)
- **Cross-check data structures** — changes on one side require verification on the other (see `rules/erp-conventions.md`)

### 9. References — read before starting work

| Task | Read these references |
|------|----------------------|
| **All tasks** | `references/core/business-rules.md` — checklist + common mistakes |
| **Documents (PO/SO/Invoice)** | `references/core/document-flows.md` — state transitions |
| **Procurement (PR/RFQ/PO)** | `references/core/procurement-flow.md` — PR->RFQ->PO API/DB/Kafka |
| **Frontend-backend integration** | `references/core/cross-system-workflow.md` — MCP-first, API spec |
| **DB schema changes** | `references/core/schema-migration.md` — migration checklist |
| **Flutter frontend** | `references/flutter/bloc-pattern.md`, `api-client.md`, `ui-components.md` |
| **Go backend** | `references/go/handlers.md`, `database-queries.md`, `mcp-tools.md` |
| **Inventory Costing** | `references/go/inventory-costing.md` — 7 costing methods, API, DB schema |
| **DB structure** | `references/core/database-schema.md` — tables + columns |
| **MCP tools** | `docs/mcp-tools-guide.md` — 41 tools reference |

### 10. Subagent Model Rule
- **Subagents ใช้ Sonnet เสมอ** — ทุกครั้งที่ spawn Agent tool ต้องระบุ `model: "sonnet"`
- Opus ใช้เฉพาะ main conversation เท่านั้น
- ยกเว้น: งาน architecture decisions หรือ complex debugging ที่ต้องการคุณภาพ Opus จริงๆ
- เหตุผล: ประหยัด "All models" quota — Sonnet เพียงพอสำหรับงาน boilerplate/สร้างไฟล์

### 11. End-of-task summary
Always summarize when finishing work:
- What was done
- Whether backend changes are needed
- Whether any skill/rule should be updated

## Auto-Update Rule (Critical)

**AI must always update skills** when learning something new from Jead:

1. **When Jead says "remember" / "set rule"** — write to clone-skills immediately
2. **When a repeated pattern is found** — add to `rules/` or `references/`
3. **When Jead corrects the AI** — update the relevant rule
4. **At end of every task** — check if any skill/rule needs updating

### How to Update
- **New rule**: write file in `rules/`, update `CLAUDE.md`
- **New skill**: create folder `skills/{name}/SKILL.md`
- **New reference**: write in `references/{domain}/`
- **Identity change**: update `identity.md`
- **Every update**: notify Jead what was changed

### Skill Evolution

| Event | Action |
|-------|--------|
| Bug fixed successfully | Add pattern/lesson to `identity.md` -> Lessons Learned |
| New pitfall found | Add to `rules/coding-style.md` or relevant rule |
| Jead teaches new method | Update relevant rule/identity |
| New MCP tool used | Update `docs/mcp-tools-guide.md` |
| New workflow created | Create new skill in `skills/` |
| Jead says "remember" | Write to `identity.md` or `rules/` immediately |
| New code pattern found | Add to `references/{domain}/` |
| **PR/RFQ/PO code changed** | **Update `skills/procurement/references/feature-matrix.md` immediately** |
| **Packing/BOM code changed** | **Update `skills/auto-packing/references/feature-status.md` immediately** |

**Goal:** More work with Jead -> smarter clone-skills -> faster onboarding for new AI agents.

## How to Use (for any project)

### Link clone-skills to a project
Add this to each project's CLAUDE.md:

```markdown
## Jead Skill Library
Read and follow all rules in `D:\bcdev\clone-skills\`:
- `CLAUDE.md` — global rules
- `identity.md` — identity and style (always follow)
- `rules/*.md` — all work rules
- `references/**/*.md` — domain knowledge + code templates
- `skills/*/SKILL.md` — available slash commands
```

### Supported AI Tools
- Claude Code (CLI + VSCode Extension)
- Cursor
- Claude Desktop (via MCP)
- Windsurf
- Any AI that can read CLAUDE.md or rules files
