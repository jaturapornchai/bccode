---
name: datacrud
description: Use when creating, reviewing, or modifying BC Account Next.js data CRUD screens with a left record list/table and a right inline view/add/edit/detail panel. This replaces the old "datalist" wording.
---

# BC Account DataCRUD Screen Skill

Use this skill for data-heavy business screens that list records and let the user view, add, edit, delete, search, filter, or select a record.

Do not call this pattern `datalist` in new work. Use `datacrud`.

## Required Source Order

1. Read `D:\bccode\AGENTS.md` and `D:\bccode\AI_INDEX.md`.
2. Read the current Next.js screen/component under `D:\bccode\frontend`.
3. If this is a migration or behavior match, read the matching Flutter screen under `D:\bcdev\frontend\bcaiaccount`.
4. Inspect existing CRUD patterns before inventing a new layout.

## Core Layout Contract

- Use a two-pane layout on desktop:
  - Left pane: record list/table, search, filters, status counts, row actions.
  - Right pane: selected record detail, add form, or edit form.
- Do not use a modal/popup for normal add, edit, or detail flows. Keep the right pane inline.
- Default split should be left `30%`, right `70%`.
- Make the split resizable by mouse drag.
- Clamp both panes so either side can shrink to about `5%` of viewport width, without breaking the page.
- Right pane stays in place while the left pane scrolls.
- If right content exceeds the available pane height, only the right pane should scroll. Constrain both left and right panes dynamically using dynamic pixel max-heights (e.g. `listMaxHeight` and `detailMaxHeight`) calculated in realtime using a `ResizeObserver` observing the element and window.innerHeight. This eliminates any outer/viewport vertical scrollbar on desktop.
- In embedded tab/menu screens, avoid page-level vertical scroll. Calculate available height from the real viewport/container and update on resize in realtime.
- For nested tables, grids, or panels inside forms/views, let them expand to their natural content height automatically in realtime (expanded like Flutter) to avoid nested scrollbars. Avoid hardcoded static limits (such as `max-h-[52dvh]`) or inner scrollbar containers for checklist matrices or checkboxes grids. Let the outer scrollable pane container handle the scrolling, and calculate layout heights dynamically in realtime only for the main pane views.
- Avoid horizontal scrolling in the left pane under typical viewports by using column priority (responsive Tailwind classes like hidden md:inline-flex to hide less important columns on narrower screens) combined with flex-nowrap on row and header containers, ensuring columns never wrap inside rows and break alignment. For extreme pane resizing, make horizontal overflow explicit and visible instead of hiding data.
- Left and right panes must use native browser scrolling so Mac trackpads and mobile finger scrolling work without custom wheel/touch handlers.
- Scrollable panes should use `overflow-y-auto`, `overflow-x-hidden`, `overscroll-contain`, and project global scroll CSS for momentum/touch behavior.
- Do not add custom `wheel`, `touchmove`, or pointer scroll logic unless native scroll cannot meet the requirement; custom handlers often break two-finger trackpad and mobile scrolling.

## List/Table Behavior

- Data-heavy screens should use horizontal rows/columns, not tall cards, keeping row layouts aligned with the sticky headers.
- Load the first 100 records first. When the user scrolls near the bottom, load the next batch.
- Count labels:
  - `ทั้งหมด` means all records from the backend dataset for the current scope.
  - Loaded/visible count must be shown separately when useful, for example `100 / 102 รายการ`.
- Search must be backend full-text search when the user asks for real search behavior. Do not rely on client-only search for Thai search behavior.
- Do not create mock, fake, or demo business data to prove the flow.
- Rows must support selection, edit, delete, and other record actions without forcing horizontal scroll.
- Do not truncate important left-pane values with `...` unless the column is intentionally compact and a tooltip/full text is still available.
- Use alternating odd/even row backgrounds.
- Hover background must be subtle and must respect the current row state:
  - selected row remains clearly selected,
  - odd/even background remains distinguishable,
  - hover is lighter than selected background.
- **Unique Row Identification**: For identifying rows in active/selection, hover, and editing states, always match against a guaranteed unique business key (e.g. `code` or `username`) instead of database IDs (`guid_fixed`) if database IDs could potentially be empty, null, or duplicated in legacy records, preventing multiple rows from highlighting simultaneously.


## Right Pane Behavior

- The right pane is the single place for display, add, and edit.
- Add/edit buttons should switch the right pane state inline.
- If a form is dirty and the user tries to select another row, switch route/tab, close the form, or start another add/edit flow, show the shared Next.js confirm dialog first.
- Never use `window.alert`, `window.confirm`, or browser-native dialogs for business warnings. Use the app confirm dialog so the message can explain the impact in normal Thai.
- The confirm dialog should be compact, centered, responsive, and not full-width on desktop.
- Save/cancel actions should stay visible or reachable without forcing page-level scroll.
- In read-only detail panels, avoid native HTML controls with the `disabled` attribute (such as `<input type="checkbox" disabled />`) for displaying status, as browsers force-fade them to an illegible light gray. Instead, implement a **custom high-contrast indicator** (e.g., CSS/SVG styled boxes using brand colors like `bg-primary` for checked and distinct borders for unchecked states) to maintain visibility and readability.
- **Checkbox Grid Column Layout**: When displaying checklist matrices or grid-like selectors (e.g. Permission Definition grid) inside forms or detail views, arrange checkboxes horizontally in columns (e.g., `grid grid-cols-2 sm:grid-cols-5`) to optimize vertical space. Add `inline-flex w-auto max-w-none` to labels to override global full-width styles.
- **Read-Before-Modify & No-Reversion**: Always inspect the latest state of target source files using `view_file` before replacing text. Ensure you do not revert premium design upgrades (such as column layouts, pottery theme styling, compact headers, or tab shapes) to generic layouts.

## Toolbar And Density

- Place search first.
- Place compact counts immediately after the search field when there is space.
- Then place refresh/add/export/other actions.
- Keep margins, padding, input height, row height, card padding, and gaps dense through shared tokens/classes first.
- Use full-width and wrap-first layout. Do not add fixed desktop-only widths except icon-only or intentionally fixed controls.

## Popups And Pickers

- Combobox/datalist-like pickers must open below the field when possible.
- Before showing any picker/menu, calculate trigger position and viewport/container bounds.
- Clamp width and left position so it never overflows the right edge.
- Any popup, picker, dropdown, or menu column that can exceed the viewport must have a max height based on `dvh` and native vertical scrolling with `overscroll-contain`.
- Keep picker rows compact and sorted by the user-facing name when requested.
- **Expandable Preset/Symbol Selection Grids**: When displaying a matrix or grid of preset options (e.g. currency symbol presets) that is reasonably small, render them naturally in a grid without forcing `max-h-*` and `overflow-y-auto` scrollbars, letting the container scale to avoid nested inner-scrollbars.

## Language And Business Rules

- Visible labels must follow the project language rules.
- Multi-language fields must respect the company's configured default language and enabled languages.
- Removing a language from active settings must not delete existing stored data for that language.
- Do not hardcode company, branch, currency, timezone, tax, warehouse, user, or business defaults.

## Verification Checklist

Run the narrowest useful checks before saying done:

- `cd D:\bccode\frontend; npm run typecheck`
- Browser check the affected screen at desktop and a narrower viewport.
- Confirm:
  - no page-level vertical overflow in embedded layout unless intentionally required,
  - left pane scrolls vertically and has no horizontal scroll,
  - right pane stays usable and scrolls independently when needed,
  - Mac trackpad/two-finger scroll and mobile finger scroll work on left pane, right pane, dropdowns, and pickers,
  - split resize works and respects the `5%` clamp,
  - dirty-form guard appears before losing unsaved edits,
  - Thai full-text search works through the backend API when implemented.
