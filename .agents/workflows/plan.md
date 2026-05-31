---
description: Force planning before multi-file change. Uses thinking_level=high.
---

# Plan First — thinking_level: HIGH

## ⚠️ This workflow forces high thinking (most expensive)
Only use for architecture / multi-file rewrite / migration.

## Output before any code

1. **ปัญหา:** <what + scope, 1 line>
2. **Root cause:** <why, 1 line>
3. **ไฟล์ที่จะแก้:** <list + line range>
4. **Diff size:** small (<20) | medium (<100) | large (>100)
5. **R level:** R0 (ask) / R1 (tell why) / R2 (just do)
6. **Test plan:** <command + expected>
7. **Token estimate:** <input k / output k>
8. **Subtask split:** if est > 16k output → break into N steps

## Decision gates
- R0/R1 → wait for confirm
- R2 + small → proceed
- R2 + large → propose split first
