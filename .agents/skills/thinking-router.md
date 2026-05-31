---
name: thinking-router
description: Auto-select thinking_level based on task complexity. Trigger on every new task to set right reasoning depth before execution.
---

# Thinking Level Router

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
🧠 thinking_level: <minimal|low|medium|high>
reason: <1 line>

## Override rules
- User says "think harder" → bump up one level
- User says "quick" → drop to minimal/low
- Hit output cap 16k → split + drop level
