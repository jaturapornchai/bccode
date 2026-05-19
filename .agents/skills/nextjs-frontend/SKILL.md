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
- Do not read `languages.tsv` fully; query exact keys only.
- Use backend language/i18n for visible text. During UI iteration, stable keys with Thai/English fallback are allowed; record missing keys under `backend/prompts/language_requests/` and batch-edit `languages.tsv` later.
- Keep UI dense, mobile-first, full-width/wrap-first, and theme-aware.
- Prefer existing components and patterns; do not copy Flutter scaffolds.
- Verify with `npm run typecheck`, focused `npm run lint -- <files>`, and focused tests when present.
