---
name: ui-beautiful
description: Generate beautiful UI components — anti-slop, multimodal-aware. Trigger on "create UI", "build page", "design", "landing", "dashboard", "form". Uses thinking_level=medium + image input if provided.
---

# Anti-Slop UI (Gemini 3.5 Flash — leverages multimodal strength)

## Set thinking_level: medium

## 📸 MULTIMODAL FIRST (Gemini 3.5 Flash strength)
If user provides image/screenshot:
- Use it as primary reference
- Extract: color palette, type scale, spacing, component patterns
- Match fidelity > creative interpretation

If no image:
- Ask "มี reference image ไหม? (เพิ่ม fidelity 3-5x)"
- Skip ask if user explicitly says "from scratch"

## ⚠️ Known weakness: Gemini แพ้ Opus 4.7 ใน complex design system
- ถ้าต้อง maintain consistency >5 components → escalate thinking=high
- หรือ route ไป Claude Opus 4.7 (Antigravity multi-model routing)

## NEVER generate (generic AI tells)
- ❌ Center hero + 3-column card grid + CTA gradient
- ❌ Serif heading on sans body (unless brand specified)
- ❌ Left accent bar on every card
- ❌ Emoji icons inline (use Lucide/Phosphor)
- ❌ Purple→pink gradient
- ❌ Rounded-full pills everywhere
- ❌ "✨ AI-powered" copy

## ALWAYS apply
1. Hierarchy through size + weight + color (not boxes)
2. Asymmetry intentional
3. ONE accent color per screen
4. Spacing rhythm: 4, 8, 12, 16, 24, 32, 48, 64
5. Type scale: 1.25 or 1.333 ratio, ≤5 sizes
6. Negative space > decoration
7. Specific copy ("Track 2,847 users" not "Track users")

## Component defaults
- **Table:** sticky header, no zebra, hover bg-muted/40
- **Form:** label above, helper below, error red-500 + icon
- **Card:** border-border > shadow, p-6, no double border
- **Button:** primary solid, secondary outline, ghost tertiary — max 2/row
- **Modal:** max-w-md, scroll inside, Esc + backdrop dismiss

## Process
1. Read project-context.md → Brand section
2. If brand undefined → ask "brand ไหน หรือ minimal default?"
3. ONE variant first → ask before alternatives
4. Output: code + 1-line rationale per design decision
5. Reference (not copy): Linear / Vercel / Stripe / Notion / Raycast

## Output schema
```json
{
  "component": "<name>",
  "stack": "<detected_from_project-context>",
  "rationale": {
    "layout": "<why>",
    "color": "<why>",
    "hierarchy": "<why>"
  },
  "code": "<diff_or_full>"
}
```
