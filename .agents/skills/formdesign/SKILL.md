---
name: formdesign
description: Use when applying BC Ai Account frontend colors and UX/UI visual treatment from the user-approved Ban Chiang / Google Stitch reference. Use only for palette, spacing, surfaces, shadows, layout density, forms, tables, menus, and interaction polish.
---

# BC Ai Account Color And UX/UI Skill

## Scope

Use this skill only for visual UX/UI decisions:

- Color palette and theme token mapping.
- Surface, border, shadow, radius, spacing, and density.
- Sidebar/menu, topbar, tabs, KPI cards, forms, dialogs, tables, filters, chips, empty/loading/error states.
- Mobile-first responsive behavior and interaction polish.

Do not use this skill as a source for backend behavior, business rules, demo data, labels, API routes, schemas, fonts, images, or HTML templates.

## Reference

User-approved source:

- `C:\Users\jatur\Downloads\stitch_ban_chiang_erp_palette.zip`
- Summary: `references/ban-chiang-google-stitch-design.md`

Extract only colors and UX/UI patterns. Do not paste the generated HTML into the product.

## Palette

Map colors to existing CSS variables/tokens first. Avoid raw one-off colors inside components.

- `primary`: `#812920`
- `primary-container`: `#a04035`
- `on-primary`: `#ffffff`
- `on-primary-container`: `#ffcec7`
- `tertiary`: `#83280e`
- `tertiary-container`: `#a33f23`
- `surface`: `#fbf9f8`
- `surface-container-lowest`: `#ffffff`
- `surface-container-low`: `#f6f3f2`
- `surface-container`: `#f0eded`
- `surface-container-high`: `#eae8e7`
- `surface-container-highest`: `#e4e2e1`
- `on-surface`: `#1b1c1c`
- `on-surface-variant`: `#56423f`
- `outline`: `#89726e`
- `outline-variant`: `#dcc0bc`
- `secondary`: `#615e53`
- `secondary-container`: `#e4dfd1`
- `error`: `#ba1a1a`
- `error-container`: `#ffdad6`

Dark theme must be derived through project CSS variables, not by hardcoding a second independent palette in components.

## UX/UI Direction

- Modern SaaS/business, warm earth-tone, professional, dense, and data-first.
- Use terracotta for primary action, active navigation, focus, selected state, and key accents.
- Use cream/paper surfaces instead of stark white where it improves comfort.
- Use subtle 1px borders for structure.
- Use soft terracotta-tinted shadows:
  - Rest: `0 4px 12px rgba(160, 64, 53, 0.08)`
  - Hover/popover: `0 8px 24px rgba(160, 64, 53, 0.12)`
- For core transaction/business screens: Keep them dense, data-first, professional, using warm earth-tone surfaces, subtle 1px borders, and soft terracotta-tinted shadows. Avoid heavy dark shadows, decorative blobs, or low-contrast text.
- For entry, portal, and workspace selection views: Use premium frosted glassmorphism, linear/radial gradient glows, micro-interactions, scale transitions on hover, and digital chip tags (e.g., monospace terminal tags) to create a striking first impression.
- **Display Integrity (No Truncation / No Omissions)**: Displayed information (such as IDs, creator emails, titles, and codes) must be fully visible. Do not truncate, clip, or omit details using ellipsis or hidden overflows. Ensure wrapping is enabled (`word-break: break-word` and `white-space: normal`) to preserve readability across all responsive breakpoints.
- **Grid & Card Layouts**: For entity cards (such as companies, shops, branches), ensure grid columns are wide enough (minimum width 480px) to prevent vertical layout clamping. Use horizontal details rows and flex alignments to display sub-details side-by-side.
- Decorative Ban Chiang/spiral motifs are allowed only as very low-opacity background accents and must never reduce readability.

## Density And Layout

- Default to compact business spacing: 8px rhythm, 8-16px gaps.
- Reduce page margins, section padding, toolbar gaps, form row spacing, table/list row height, and control height as far as practical while preserving readability and tap safety.
- Apply density globally or through shared screen tokens before adding page-specific overrides.
- Use full-width, wrap-first layouts.
- Above-the-fold UI is a compact workbench, not a banner. Keep the topbar, category/menu row, tab strip, screen header, and search/action toolbar low-height and dense; use compact controls, tight gaps, and wrapping instead of tall stacked chrome.
- Do not enlarge these zones with hero-style cards, decorative whitespace, tall tab cards, or default-size utility buttons unless the user explicitly requests a roomier layout for that screen.
- Popups, dropdowns, menus, and combobox panels must calculate width/position from the trigger and current viewport/container before showing, then clamp to the visible area instead of overflowing to the right.
- Keep controls compact but touch-safe on mobile.
- Desktop: sidebar/tree menu + topbar + tab strip + dense content canvas.
- Tablet/mobile: collapse navigation, wrap action bars, stack forms, keep tables scroll-contained or card-based.
- Avoid fixed desktop-only widths and horizontal overflow.

## Components

- Buttons: primary solid terracotta, secondary outline, ghost for utility.
- Inputs: top labels, clear focus ring/border, compact height, helper/error text when needed.
- Tables: compact rows, sticky/action header when useful, horizontal separators, row actions, filter/search/pagination states.
- Cards: subtle border, light surface, optional left accent bar for section headers.
- Chips: muted semantic background with high-contrast text.
- Dialogs: centered and responsive; use only when interaction benefits from a modal.
- Menus: tree/collapsible groups, visible active state, badges, scrollable when taller than viewport.

## Guardrails

- Do not introduce visible text from the reference. Use the project language system.
- Do not introduce fonts or remote images from the reference.
- Do not change business logic while applying color/UX/UI.
- Preserve the project rules for language, light/dark theme, responsive layout, no hardcode, and verification.

## Verification

For changed UI, verify:

- Light and dark theme.
- Mobile, tablet, and notebook widths.
- No horizontal overflow.
- Menu/sidebar scrolls when content exceeds viewport height.
- Language switch still updates visible text.
