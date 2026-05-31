---
name: datacrud
description: Use when creating, reviewing, or modifying Next.js data CRUD screens with a left list and a right inline detail/add/edit form.
---

## 1. Setup & Precedence
- **Location**: `D:\bccode\frontend`.
- **References**: Read [AGENTS.md](file:///D:/bccode/AGENTS.md) and [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md) first.
- **Goal**: Build data-heavy screens using a left list/table and a right inline detail/add/edit form. No popups for editing.

## 2. Layout & Scrolling Contract
- **Pane Split**: Desktop split default 30% left / 70% right. Must be resizable (mouse drag, clamp both sides to 5% min width).
- **Height Locking**: Calculate `listMaxHeight` and `detailMaxHeight` using `ResizeObserver` and `window.innerHeight`. Right pane scrolls independently when content overflows. No outer/viewport scrollbars on desktop.
- **Nested Expandable**: Nested checkbox grids or tables inside details must auto-expand vertically to content height (no inner scrollbars, no static `max-h-*` limits).
- **Responsive Columns**: Hide secondary columns on narrow viewports (`hidden md:inline-flex` etc.) combined with `flex-nowrap` on row containers to prevent wrapping and alignment breakage.

## 3. List/Table Behavior
- **Global CSS Classes (MANDATORY)**: Every data list/table MUST use the project-wide compact CSS classes defined in `globals.css`. Do NOT use ad-hoc inline Tailwind for list chrome:
  - `.bc-list-toolbar` — summary bar (title + count). `font-size: 0.75rem`, minimal padding.
  - `.bc-list-header` — column header row. Sticky, uppercase, `font-size: 0.7rem`, `font-weight: 800`, primary-tinted gradient background, 2px primary bottom border.
  - `.bc-list-row` — each data row. `font-size: 0.75rem`, `padding: 3px 8px`, `display: flex` by default. For grid-based screens, add Tailwind `grid` + `lg:grid-cols-[...]` classes alongside `bc-list-row`.
- **Font Size Uniformity**: All columns in both header and rows must share the same inherited font-size from the CSS class. Do NOT use `font-bold` or `text-sm` / `text-base` on individual cells — use `font-medium` at most for the primary identifier column.
- **Data Load**: Load first 100 records, paginating dynamically as the user scrolls.
- **Count Labels**: Show total dataset count (`ทั้งหมด`) and loaded counts clearly (e.g. `100 / 102 รายการ`).
- **Search**: Must call backend full-text search. Client-only search is not allowed for Thai text matching. No mock data.
- **Row Interactions**: Support selection, editing, and deletion inline. Odd/even alternating backgrounds. When a row is actively being edited, it must change its background to a distinct amber/orange highlight style (`bg-amber-100/70 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-amber-200/50`) so the user can easily identify the row being modified.
- **Actions Column**: Use explicit, visible, colored icon buttons matching the product screen style (e.g. blue/primary outline Pencil `size-7 rounded-lg bg-background text-primary hover:bg-primary/10 border-border`, red/destructive Trash2 `size-7 rounded-lg bg-background text-destructive hover:bg-destructive/10 border-border`). Stop event propagation so clicking an action does not trigger row selection.
- **Row Highlight Key**: Use a unique business key (e.g., `code` or `username`), not database `guid_fixed`, to avoid empty/null/duplicate highlight bugs.

## 4. Right Pane & Form Behavior
- **Actions**: Save/cancel buttons must stay visible/reachable.
- **Dirty Form Warning**: Show a compact, responsive, centered Next.js confirmation dialog (no `window.confirm` or native alerts) if the user has unsaved edits and tries to change rows or navigate away.
- **Multilingual Fields**: Render the company's active languages from `settings.languageconfigs` in order. Normal CRUD forms must not show add-language buttons; language management belongs to Active Languages. Removing an active language must hide the input without deleting stored hidden-language values, and values must not be auto-filled from another language. **Always use the shared `<NamesEditor>` component (imported from `@/components/product-barcode/names-editor`) to render localized name inputs in standard 2-column grids with country flags.**
- **Checkbox Grid Layout**: Arrange checklist matrices horizontally using responsive columns (e.g., `grid-cols-2 sm:grid-cols-5`). Use `inline-flex w-auto max-w-none` on labels to prevent vertical stacking.
- **Read-Only High Contrast**: Do not use native `<input type="checkbox" disabled />` for displaying status, as browsers fade them to illegible gray. Render custom themed CSS/SVG high-contrast indicators.
- **Nested & Tree-Structured Data**: For hierarchical structures (e.g., Warehouse → Location/Zone → Shelf), the frontend should manage the entire tree structure in a single local state and send the complete nested payload via a single `PUT` request (Cascading Save) to ensure atomic state updates. The corresponding backend transaction must fully sync (delete-and-reinsert or diff-update) all nested child records accordingly.

## 5. Density & Density Optimization
- **Workbench Layout**: Minimize margin, padding, section spacing, toolbar gaps, and row heights (use compact heights e.g., `h-8` where practical). Gaps between fields on forms should not stretch.
- **List Density**: Always use `.bc-list-*` CSS classes. Never override font-size per-column. Header and row font must match (`0.75rem` / 12px). Padding is `3-4px 8px`. Do NOT use `text-sm` (14px) on list rows.
- **Popups & Pickers**: By default, open below the trigger. However, if viewport space below is limited (e.g., less than 300px) and there is more space above, dynamically flip the picker to open above the trigger (Flipping Placement) by setting `bottom` relative to the trigger's top, clearing `top` style, and calculating appropriate `maxHeight`. Clamp width/left so it never overflows off-screen. **To prevent CSS transform or overflow-hidden on parent elements from breaking position: fixed positioning, always render popups, dropdowns, and pickers using React Portal (`createPortal`) targeted directly to `document.body`.**

## 6. Pre-Commit Verification
- Run `npm run typecheck` in `frontend/` only before commit/summary (fast-iteration dev mode relies on `next dev` HMR while iterating).
- Verify scrolling behavior: left pane, right pane, and dropdowns scroll independently using native scrolling. Verify trackpad and mobile scrolling works.
- Verify dirty-form guard appears correctly.
