---
name: Chatbot Answer Format
description: "Nong Kung" chatbot answer must contain Plan + Result + Detailed Analysis. NO recommendations, NO next-steps sections.
---

# Chatbot Answer Format Rule

**Set by Jead on 2026-04-08** (revised same day to add Plan section)

## The Rule

Every chatbot answer must contain exactly **3 sections** in this order:

1. **แผนการหาคำตอบ (Plan)** — show how the AI plans to answer (understanding, strategy, expected result) — 2-5 short bullets. Mandatory for ALL questions.
2. **ผลลัพธ์ (Result)** — what was found / not found (Markdown table for multiple rows). For non-DB questions (recipes, how-to), this is the actual content.
3. **บทวิเคราะห์ละเอียด (Detailed analysis)** — go DEEP, not shallow

**Why Plan is mandatory:** Jead's feedback (2026-04-08) — users want to see HOW the AI decided to approach their question. The Plan section makes the AI's reasoning transparent and helps users trust the result. Even simple questions get a Plan.

**FORBIDDEN sections:**
- ❌ "Recommendations" (Thai: "คำแนะนำ")
- ❌ "Next steps" (Thai: "ขั้นตอนถัดไป")
- ❌ "Suggested actions"
- ❌ Any bullet list asking "Would you like to...?" (Thai: "ต้องการ...หรือไม่คะ?")

The user wants **pure data + analysis only**. Follow-up questions belong in the separate `suggested_questions` field (parsed by frontend separately, displayed as clickable chips), NOT in the answer body.

## Why?

Jead's feedback (2026-04-08): the recommendation section was redundant — users already know how to navigate the app, and follow-up questions are already shown via `suggested_questions`. Putting them in the answer body wastes space and adds noise.

## What "detailed analysis" means

Go beyond simple data dump. Examples by domain:

| Domain | Shallow (BAD) | Deep (GOOD) |
|--------|---------------|-------------|
| Customer record | "Found ABC Co." | Code, taxid, full name, address, contact, credit terms, last activity, business context (B2B vs B2C), credit risk indicators |
| Sales data | "Total: 10,000 baht" | WoW/MoM trends, top contributors, anomalies vs average, seasonality |
| Stock data | "100 units left" | Turnover rate, days of supply, dead stock risk, reorder implications |
| Not found | "Not found" | List every table searched, explain WHY each was searched, infer input pattern (taxid? barcode? name?), possible reasons (typo, not registered, archived) |

## File affected

- `backend/internal/goapi/handlers/aichat/agent_loop_v2.go` — `kungSystemPrompt()` "Answer Quality" section
- (ReAct path uses different system prompt — apply same rule there if Jead reports issue with Ollama path)
