---
name: nextjs-frontend
description: Use when creating, reviewing, or modifying the BC Account Next.js frontend in D:\bccode\frontend.
---

Target app: `D:\bccode\frontend`.

Read `D:\bccode\AGENTS.md` first; it is the full source of frontend, language,
theme, responsive, manual, security, and context-budget rules.

Fast workflow:
- Inspect only the relevant route/component/lib/test files.
- Use `rg` so `.ignore` excludes generated docs, dependencies, legacy Flutter,
  manuals, binaries, and duplicate skill bundles.
- For migration, clone, redesign, or reference-screen work, first inspect the
  matching Flutter screen under `D:\bcdev\frontend\bcaiaccount` and any
  user-provided screenshot/reference screen. Preserve the user-facing workflow,
  labels, menu hierarchy, actions, states, permissions, and API/data intent in
  the Next.js implementation.
- If the matching Flutter screen cannot be found, report the searched paths
  instead of guessing the workflow.
- Do not read `languages.tsv` fully; query exact keys only.
- Use backend language/i18n for visible text. During UI iteration, stable keys with Thai/English fallback are allowed; record missing keys under `backend/prompts/language_requests/` and batch-edit `languages.tsv` later.
- Keep UI dense, mobile-first, full-width/wrap-first, and theme-aware.
- For screens, minimize page margins, section padding, gaps, field spacing, table/list row height, and control height through shared CSS/tokens first, so the viewport shows more real business data. Keep touch targets usable, but do not add decorative whitespace.
- Prefer wrapping toolbars/tabs/forms over horizontal overflow. Before opening popups, dropdowns, or menus, calculate their size and position from the trigger and current viewport/container so they do not overflow off the right edge.
- Prefer existing components and patterns; do not copy Flutter scaffolds.
- Verify with `npm run typecheck`, focused `npm run lint -- <files>`, and focused tests when present.
