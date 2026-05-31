---
description: Generate beautiful UI single pass. Uses medium reasoning depth + multimodal input.
---

# Quick UI — reasoning depth: MEDIUM

Usage:
- `/qui <component> [hint]`
- `/qui <component> [paste image]` ← preferred when the model supports image input

Examples:
- `/qui dashboard analytics with charts + KPI`
- `/qui landing SaaS dark hero + pricing` + screenshot
- `/qui form multi-step onboarding 4 steps`

## Process
1. Load `ui-beautiful` skill
2. Read brand from project-context.md
3. If image → extract palette/spacing/type as primary spec
4. Detect stack (shadcn/Tailwind/raw)
5. Generate ONE variant
6. Output: code + 3-line rationale + reference (Linear/Vercel/etc)

## Anti-slop check before output
- [ ] No 3-card grid center hero
- [ ] One accent color only
- [ ] Specific copy (numbers, names)
- [ ] Spacing on 4-multiple
- [ ] Type scale ≤5 sizes
- [ ] No purple→pink gradient

## Token budget: ~5-8k output (medium reasoning depth + UI code)

## Escalation
- >5 connected components needing consistent design system
  → escalate reasoning depth to high, or split into smaller passes with a shared token set
