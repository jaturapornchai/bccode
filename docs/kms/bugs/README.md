# Bug Diary

บันทึก bug + root cause — กัน regression

## Format

```markdown
---
date: 2026-04-13
severity: high  # low | medium | high | critical
component: [backend, frontend, db]
tags: [bc-account, postgres]
fixed: true
---

# Symptom

อาการที่เจอ (sharp, reproducible)

## Root Cause
สาเหตุจริงๆ (ไม่ใช่ symptom)

## Fix
แก้ยังไง — link commit: `abc1234`

## Regression Test
ทำยังไงให้มันไม่เกิดซ้ำ
```
