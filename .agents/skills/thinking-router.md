---
name: thinking-router
description: Auto-select reasoning depth based on task complexity. Trigger on every new task to set the right reasoning depth before execution. Model-agnostic — each model maps depth to its own knob (ZCode native reasoning / Codex model_reasoning_effort / Gemini thinking_level).
---

# Reasoning Depth Router

## Auto-classify before any action

### → minimal (fastest, cheapest)
- "What does X do?"
- "Format this", "rename"
- Reading file to answer Q
- Listing/searching

### → low
- Single-file CRUD endpoint
- Simple bug fix (known root cause)
- Boilerplate generation
- Test scaffolding
- /qcrud workflow

### → medium (default — most coding)
- Multi-file edit (2-5 files)
- Debug with unknown cause
- Refactor within module
- UI component generation
- /qui workflow

### → high (expensive — use sparingly)
- Architecture decision
- Multi-file rewrite (>5 files)
- Novel algorithm design
- Performance optimization with constraints
- /plan workflow
- Migration strategy

## Output before executing
🧠 reasoning depth: <minimal|low|medium|high>
reason: <1 line>
(map to your model: ZCode native reasoning / Codex model_reasoning_effort / Gemini thinking_level)

## Override rules
- User says "think harder" → bump up one level
- User says "quick" → drop to minimal/low
- Hit output cap → split task + drop one level
