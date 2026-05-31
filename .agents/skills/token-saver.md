---
name: token-saver
description: Active when context large or user mentions speed/cost. Enforces prompt-caching strategy + output discipline. Model-agnostic.
---

# Token Conservation (all models)

## 💾 Caching first (big input savings)
- System prompt + project context = keep at TOP, unchanged across turns so it stays cached
- Don't paraphrase cached content in answer
- Reference by anchor, not full repeat

## Output discipline (output costs more than input on most models)
- Code: diff only
- Explanation: ≤3 sentences
- JSON output: use schema constraint, not free prose
- No "let me", "I'll now", "first I'll"
- No closing recap

## Reasoning-depth control
- Multi-turn context grows — keep prompts lean
- Simple Q&A → minimal depth, clear reasoning to save tokens
- Long agent loop → keep depth as needed (improves quality)

## Tool call discipline
- Reduce reasoning depth FIRST if over-calling
- Constrain tools in system instruction
- Batch tool calls when possible (parallel MCP)

## Context placement (cache-friendly)
- TOP: stable (system, project context, docs)
- MIDDLE: relevant code excerpts
- BOTTOM: user question

## Multi-step budget warning
- 5-step agent = 2-3x base token usage
- Budget: estimate before /plan execution
- Hit limit → drop reasoning depth + split task

## Read budget
- Max 3 files/turn (unless multi-file task)
- Max 200 lines/file (offset+limit)
- Skip files in "Do NOT touch"
