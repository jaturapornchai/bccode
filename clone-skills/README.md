# Jead Skill Library

AI Skill Library for BC AI Cloud — enables any AI agent to work correctly with Jead.

## Structure

```
clone-skills/
├── CLAUDE.md              ← Entry point (AI reads this first)
├── identity.md            ← Identity, style, Lessons Learned
├── rules/                 ← Rules AI must follow
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

## Usage

### Link to a project
Add to each project's CLAUDE.md:

```markdown
## Jead Skill Library
Read and follow all rules in `D:\bcdev\clone-skills\`:
- `CLAUDE.md` — global rules
- `identity.md` — identity and style
- `rules/*.md` — work rules
- `references/**/*.md` — domain knowledge
- `skills/*/SKILL.md` — slash commands
```

### Install Script (Multi-tool)

```bash
bash install.sh
```

Copies skills to: `.claude/`, `.agents/`, `.kilocode/`, `.kiro/`, `.agent/`

## Supported AI Tools

- Claude Code (CLI + VSCode Extension)
- Cursor
- Claude Desktop (via MCP)
- Windsurf
- Any AI that can read CLAUDE.md

## Slash Commands

| Command | Description |
|---------|-------------|
| `/api-search <keyword>` | Search API endpoints |
| `/api-spec <method> <path>` | View API specification |
| `/model-gen <model_name>` | View/generate model schema |
| `/enum-list [keyword]` | View enum values |
| `/mcp-check` | Check MCP server status |
| `/master-data <action> <type>` | Explore master data |

## Auto-Update Rule

All AI agents must update the skill library when learning something new from Jead.
More work -> smarter skills -> faster onboarding for new AI agents.
