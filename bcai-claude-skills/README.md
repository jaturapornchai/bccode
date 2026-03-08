# Jead Skill Library

AI Skill Library สำหรับ BC AI Cloud — ช่วยให้ AI ทุกตัวทำงานกับ Jead ได้อย่างถูกต้อง

## โครงสร้าง

```
jeadbcdev/
├── CLAUDE.md              ← Entry point (AI อ่านไฟล์นี้ก่อน)
├── identity.md            ← ตัวตน สไตล์ Lessons Learned
├── rules/                 ← กฏที่ AI ต้องปฏิบัติตาม
│   ├── coding-style.md    ← Go + Flutter conventions
│   ├── erp-conventions.md ← ERP system conventions
│   ├── mcp-first.md       ← MCP development rules
│   ├── security.md        ← Security rules
│   ├── workflow.md        ← Development pipeline
│   └── manual-writing.md  ← Documentation standards
├── skills/                ← Slash commands (user-invocable)
│   ├── api-search/        ← /api-search <keyword>
│   ├── api-spec/          ← /api-spec <method> <path>
│   ├── model-gen/         ← /model-gen <model_name>
│   ├── enum-list/         ← /enum-list [keyword]
│   ├── mcp-check/         ← /mcp-check
│   └── master-data/       ← /master-data <action> <type>
├── references/            ← Domain knowledge + code templates
│   ├── core/              ← Business rules, document flows, DB schema
│   ├── flutter/           ← BLoC pattern, API client, UI components
│   └── go/                ← Handlers, queries, MCP tools
└── docs/
    └── mcp-tools-guide.md ← MCP tools reference (41 tools)
```

## วิธีใช้

### เชื่อมเข้า Project (CLAUDE.md)

เพิ่มใน CLAUDE.md ของแต่ละ project:

```markdown
## Jead Skill Library
อ่านและปฏิบัติตาม rules ทั้งหมดใน `D:\bcdev\bcai-claude-skills\`:
- `CLAUDE.md` — global rules
- `identity.md` — ตัวตนและสไตล์
- `rules/*.md` — กฏการทำงาน
- `references/**/*.md` — domain knowledge
- `skills/*/SKILL.md` — slash commands
```

### Install Script (Multi-tool)

```bash
bash install.sh
```

Copy skills ไปยัง: `.claude/`, `.agents/`, `.kilocode/`, `.kiro/`, `.agent/`

## Supported AI Tools

- Claude Code (CLI + VSCode Extension)
- Cursor
- Claude Desktop (via MCP)
- Windsurf
- ทุก AI ที่อ่าน CLAUDE.md ได้

## Slash Commands

| Command | Description |
|---------|-------------|
| `/api-search <keyword>` | ค้นหา API endpoints |
| `/api-spec <method> <path>` | ดู API specification |
| `/model-gen <model_name>` | ดู/สร้าง model schema |
| `/enum-list [keyword]` | ดู enum values |
| `/mcp-check` | ตรวจสอบ MCP server |
| `/master-data <action> <type>` | สำรวจ master data |

## Auto-Update Rule

AI ทุกตัวต้อง update skill library เมื่อเรียนรู้สิ่งใหม่จาก Jead
ยิ่งทำงานมาก → skill ยิ่งฉลาด → AI ตัวใหม่เข้าใจ Jead เร็วขึ้น
