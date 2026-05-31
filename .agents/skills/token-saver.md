---
name: token-saver
description: Active when context large or user mentions speed/cost. Enforces Gemini 3.5 Flash caching strategy + output discipline.
---

# Token Conservation (Gemini 3.5 Flash specific)

## 💾 Caching first (90% input savings)
- System prompt + project-context = ALWAYS cached
- Don't paraphrase cached content in answer
- Reference by anchor, not full repeat

## Output discipline (output 6x cost of input — Gemini pricing)
- Code: diff only
- Explanation: ≤3 sentences
- JSON output: use schema constraint, not free prose
- No "let me", "I'll now", "first I'll"
- No closing recap

## Thought preservation control
- Default ON (3.5 Flash) → input grows multi-turn
- Simple Q&A → clear thoughts to save tokens
- Long agent loop → keep on (improves quality)

## Tool call discipline
- Reduce thinking_level FIRST if over-calling
- Constrain tools in system instruction
- Batch tool calls when possible (parallel MCP)

## Context placement (cache-friendly)
- TOP: stable (system, project-context, docs)
- MIDDLE: relevant code excerpts
- BOTTOM: user question

## Multi-step budget warning
- 5-step agent = 2-3x base token usage
- Budget: estimate before /plan execution
- Hit limit → drop thinking_level + split task

## Read budget
- Max 3 files/turn (unless multi-file task)
- Max 200 lines/file (offset+limit)
- Skip files in "Do NOT touch"
