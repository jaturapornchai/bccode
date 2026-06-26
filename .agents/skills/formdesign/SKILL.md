---
name: formdesign
description: Use when applying BC Ai Account colors and UX/UI visual treatment from the Ban Chiang / Google Stitch design.
---

## 1. Color Palette (Stitch Ban Chiang Reference)
Map colors to existing CSS variables/tokens. Avoid raw hex values in components.
- `primary`: `#812920` (terracotta primary accent)
- `primary-container`: `#a04035`
- `on-primary`: `#ffffff`
- `on-primary-container`: `#ffcec7`
- `tertiary`: `#83280e`
- `tertiary-container`: `#a33f23`
- `surface`: `#fbf9f8` (warm cream/paper surface)
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

*Note: Derive Dark Theme through CSS variables, not by hardcoding separate palettes in components.*

## 2. Surfaces & Spacing
- **Aesthetics**: Warm earth-tone, professional, dense, data-first.
- **Borders & Shadows**: Use subtle 1px borders. Use soft terracotta-tinted shadows:
  - Rest: `0 4px 12px rgba(160, 64, 53, 0.08)`
  - Hover/popover: `0 8px 24px rgba(160, 64, 53, 0.12)`
- **Corner Radius (tight — half scale, set 2026-06-26)**: Keep corners crisp, not pill-soft. Use only the radius tokens in `globals.css` — `--radius-xs:2 / sm:3 / md:5 / lg:7 / xl:10 / 2xl:14`px and `--radius-full:9999` (circular avatars / pills ONLY). Inputs, buttons, selects, cards, dialogs → `rounded-*` utilities (which map to these tokens); default small: inputs/buttons/selects `rounded-md` (5px), cards/dialogs `rounded-lg`/`rounded-xl`.
  - **Never hardcode** `border-radius:Npx` in CSS or `rounded-[Npx]` / inline `borderRadius:N` in TSX — it bypasses the scale (this was the bug: 113 hardcoded px in globals.css didn't shrink when only the vars were halved). Always go through the token/utility.
  - To change global roundness, edit the 6 `--radius-xs..2xl` tokens once (keep `full` as the pill). Use a byte-preserving Node patch on `globals.css` (mixed EOL — never the Edit tool).
- **Glassmorphism**: Use frosted glass, radial gradients, scale transitions, and monospace terminal chips for workspace/login screen portals to WOW users.
- **Display Integrity**: No truncation of text using `...` (such as codes, IDs, usernames, or emails). Enable wrapping: `word-break: break-word` and `white-space: normal`.
- **Baseline Alignment**: Inline text-only metadata groups such as chips, badges, labels plus code/name, and compact context rows should align text by baseline (`align-items: baseline`) with normalized line-height. Use vertical centering only for icon/control groups where the icon is the primary alignment target.
- **Card Grids**: Minimum column width of 480px for company/branch/shop cards to prevent layout squishing. Use horizontal sub-details layouts.

## 3. Spacing & Density Contract
- **Rhythm**: Spacing based on 8px rhythm. Margins, padding, input height, row height, and card padding must stay dense.
- **Workbench UI**: Topbar, menu bar, open-tab strip, screen header, and search toolbar must be compact (height around `h-8`/32px where practical).
- **Popups**: Menus, dropdowns, and pickers must calculate their size and position dynamically and clamp to the viewport (no overflowing off-screen).
- **Responsive stack**: Mobile/tablet views must collapse navigation, wrap toolbars, and stack forms cleanly without causing horizontal scrolling.

## 4. UI Components Styling
- **Buttons**: Terracotta solid for primary actions, outline for secondary, ghost for utility.
- **Inputs**: Compact heights, top labels, clear focus outline.
- **Tables**: Dense row heights, sticky headers, subtle odd/even row colors, terracotta accent focus state.
- **Chips**: Muted backgrounds with high-contrast text.
- **Dialogs**: Clean, centered, responsive, and not full-width on desktop.

## 5. Verification
Verify light/dark theme, responsive widths, language switches, and scroll behavior.
