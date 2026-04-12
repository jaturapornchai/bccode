# Jead Skill — Global AI Rules

See `identity.md` for Jead's identity, clone team, and critical pitfalls.

## Required Reads per Session
1. `identity.md` — who Jead is, style, permanent pitfalls
2. This file — rules
3. Relevant `rules/*.md` — based on current task (load lazily)
4. Relevant `references/**/*.md` — domain knowledge (load lazily)

Skills (`skills/*/SKILL.md`) auto-load via description metadata — trigger by keyword.

## Rules (mandatory)

### 1. Language
- **Always respond in Thai** — Jead is Thai
- Variable/function names in English
- Log messages/comments: Thai allowed (for human readers)
- System prompts sent to LLMs: **English** (see `rules/ai-prompt-language.md`)

### 2. Workflow
- **Act immediately** — don't ask if you can just do it
- **Read before editing** — understand code first
- **MCP-first** — check backend via MCP before writing frontend code
- **No over-engineering** — only what's asked
- **End-of-task summary** — what was done, backend changes needed, skill updates

### 3. Cross-Project Scope
| Repo | Path | Access |
|---|---|---|
| Backend (Go) | `D:\bcdev\backend` | read + edit |
| Frontend (Flutter) | `D:\bcdev\frontend\bcaiaccount` | read + edit |
| clone-skills | `D:\bcdev\clone-skills` | read + update |

Changes on one side require cross-check on the other (see `rules/erp-conventions.md`).

### 4. Subagent Model
- Spawned Agents always use `model: "sonnet"`
- Opus reserved for main conversation
- Exception: architecture decisions / complex debugging that need Opus quality
- Reason: save "All models" quota; Sonnet handles boilerplate fine

### 5. AI Prompt Language (set 2026-04-08)
- All system prompts to LLMs: **English** (token efficiency, better instruction-following)
- AI reply to user: **Thai** — enforce via `CRITICAL: Always reply in Thai` at end of prompt
- Applies to: agent system prompts, tool descriptions, mid-conversation injections, summarizer prompts
- Details: `rules/ai-prompt-language.md`

### 6. Other rule files (load when relevant)
| File | When |
|---|---|
| `rules/coding-style.md` | Writing Go or Flutter code |
| `rules/mcp-first.md` | Backend API work, new endpoints |
| `rules/erp-conventions.md` | Document flow, multi-tenant, sync |
| `rules/security.md` | Auth, permissions, sensitive data |
| `rules/workflow.md` | Deploy pipeline, Docker, CI/CD |
| `rules/manual-writing.md` | User-facing docs |
| `rules/chatbot-answer-format.md` | Nong Kung chatbot output |

### 7. References (load when relevant)
| Task | Reference |
|---|---|
| Any task | `references/core/business-rules.md` — checklist + common mistakes |
| Documents (PO/SO/Invoice) | `references/core/document-flows.md` |
| Procurement (PR/RFQ/PO) | `references/core/procurement-flow.md` + `references/modules/procurement/` |
| Frontend↔backend integration | `references/core/cross-system-workflow.md` |
| DB schema changes | `references/core/schema-migration.md` |
| Flutter | `references/flutter/{bloc-pattern,api-client,ui-components}.md` |
| Go | `references/go/{handlers,database-queries,mcp-tools,inventory-costing,bcproxyai}.md` |
| AI Provider (bcproxyai) | `references/go/bcproxyai.md` — virtual models, routing, troubleshooting |
| DB structure | `references/core/database-schema.md` |
| MCP tools | `docs/mcp-tools-guide.md` — 41 tools reference |
| RAGFlow / Knowledge Base | `references/infra/ragflow-setup.md` |
| ERP module details | `references/modules/<name>.md` — accounting, payment, product, quotation, sales-transaction, stock-inventory, creditor-debtor, approval-workflow, coupon-promotion, dashboard-report, import-export, lineoa-chatbot, restaurant-pos, auto-packing/, procurement/ |

## Auto-Update Rule

**Update clone-skills whenever you learn something new from Jead:**
- Jead says "remember" / "set rule" → write to `rules/` or `identity.md` immediately
- Jead corrects AI → update the relevant rule
- New pitfall discovered → add to `identity.md` Critical Pitfalls or relevant rule
- New workflow → create skill in `skills/`
- New code pattern → add to `references/{domain}/`
- PR/RFQ/PO code changed → update `references/modules/procurement/references/feature-matrix.md`
- Packing/BOM code changed → update `references/modules/auto-packing/references/feature-status.md`

Notify Jead after every update.

## How to Link from a Project CLAUDE.md
```markdown
## Jead Skill Library
Read and follow `D:\bcdev\clone-skills\`:
- `identity.md` — identity + pitfalls (always read)
- `CLAUDE.md` — global rules (always read)
- `rules/*.md` — load when relevant
- `references/**/*.md` — load when relevant
- `skills/*/SKILL.md` — slash commands (auto-triggered)
```
