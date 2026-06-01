---
name: nextjs-frontend
description: Use when creating, reviewing, or modifying the Next.js frontend in D:\bccode\frontend.
---

## 1. Setup & Precedence
- **Location**: `D:\bccode\frontend`.
- **References**: Always read [AGENTS.md](file:///D:/bccode/AGENTS.md) and [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md) first.
- **Migration**: For Flutter migration, read Dart files in `D:\bcdev\frontend\bcaiaccount\lib\screens\` first. Match exact workflows, labels, and validations. Report searched paths if legacy file is missing.

## 2. Dynamic Height & Scroll Locking Layout
- **Desktop Two-Pane**: 30% left pane (list/table), 70% right pane (forms/details) split.
- **Scroll Constraints**: Viewport height must be calculated dynamically in realtime using ResizeObserver. Avoid outer page vertical scrollbars. Left list and right form scroll independently with `overflow-y-auto min-h-0 h-full`.
- **Form Layout**: Nested checkboxes/tables inside details/forms must auto-expand vertically (no inner scrollbars). Rely on the main detail pane to scroll.
- **Horizontal Scroll**: Avoid left-pane horizontal scroll. Apply responsive hide classes (`hidden md:inline-flex` etc.) on less important columns. Keep rows and headers aligned with flex-nowrap.

## 3. Product Category (`/product_category_group_select_screen`)
- **Selector Grid**: Shows 20 group cards (1-20). Minimal vertical space, flex-wrap layout. Compute card width dynamically to match parent container width. occupy full content width before selection.
- **Toolbar**: Lift search/add/back controls into the screen header beside the title once a group is selected (no toolbar above the tree).
- **Split Pane**: Left tree and right editor stretch to viewport height and scroll independently. Draggable center splitter with col-resize cursor.
- **dnd-kit Drag-n-Drop**:
  - Pointer hold-and-drag from anywhere on the row (excluding action buttons).
  - Overlay: small transparent dashed ghost row with small left chip following pointer during drag.
  - Sibling Swap: FLIP animations for real-time order previews.
  - Drop Targets: Top 28% (insert before), middle (move inside target as child), bottom 28% (insert after), root target at bottom of tree pane (clears parent GUIDs). Reject moving parent under own child.
  - Optimistic State: Apply changes in memory immediately. Pulse dropped row and scroll to it if out of view. Refresh from API only on save error.
  - Detail action: Must use optimistic values immediately after drop.
  - Indentation: Use absolute left bars for selected indicators, do not shift text layout.
- **Editor**: Compact fields (no stretching gaps). Recursive "all child categories" list under fields.
- **Product Assignment**: Integrated as "Products in category" tab using barcodes API. Explicit segmented tabs with icons.

## 4. Warehouse Location Shelf (`/product_warehouse_screen`)
- **UX Pattern**: Follow the Product Category workbench pattern: compact toolbar, left warehouse/location/shelf tree, right inline editor.
- **Split Pane**: Left tree and right editor must fill the available viewport height and scroll independently with `h-full min-h-0 overflow-y-auto`.
- **Density**: Keep warehouse selection, search, record count, tree rows, and editor fields compact. Do not stack a large tree card above a separate form card.
- **Languages**: Warehouse and location `names` fields must render exactly the active company languages in order; no extra fallback languages in normal forms.

## 5. Branch & Address Details (`/branch`)
- **Branch Page**: Single entrypoint for branch-scoped data. Embedded tabs: General, Address, Department, Work Day, Holiday, POS/Tax, Business Types.
- **Leaflet Map Picker**: Pick coordinates in full-screen map using OpenStreetMap (Bangkok default: 13.7563, 100.5018). Round to 6 decimals.
- **Thailand Address**: Select province -> district -> subdistrict -> zip code. Load data via `/goapi/api/address/thailand`. Recompute/clear postal code on changes.
- **Branch Settings**: Normal tables/forms (no raw JSON editors) for `paymentrounding` and `pointconfig`.
- **Thai Branch Code**: Legal head-office uses `00000`. Pad/normalize input to 5 digits (e.g. `1` -> `00001`).

## 6. UI Widgets & Interactions
- **CRUD Mutation Source**: Frontend create/edit/delete actions must call MongoDB-backed operational APIs only. Do not call PostgreSQL/ClickHouse projection or BI endpoints for mutations, and do not use derived stores as normal list/detail truth after saving. Refresh operational screens from MongoDB-backed APIs; Kafka/projection fan-out is backend-owned.
- **CRUD Row Click**: In data-list screens, row click selects and opens read-only detail only. Edit mode must require the explicit pencil/edit button, and selected/read-only rows must not use the amber editing highlight.
- **Company Access Editors**: `company_guids` selectors are company-level only. Render company choices, do not render branch trees, and do not send branch selections for normal CRUD/master-data access. Branch-specific screens may manage branches explicitly.
- **Tenant-Bound Master Data**: Screens that load tenant-bound master data from token-scoped APIs (for example `/product`) must ensure the backend token is selected to the active workspace `shopid` before loading records. Do not let a stale token silently show another tenant's empty list.
- **Workspace Company Cards**: Company-selection cards must use metadata already returned by `/list-shop` or the local workspace proxy. Do not depend on `/shop/:id` before `select-shop`, because shop detail endpoints require an active selected shop and will otherwise fall back to incomplete language/currency/date values.
- **Workspace Company & Branch Selection**: When a company is selected, the system must verify the branches. If no branch exists, it must automatically create the head-office branch (`00000` named "สำนักงานใหญ่"). The frontend must transition to the branch selection step (`step = "branches"`) in all cases, ensuring the user manually selects the branch, even if only a single branch is available. Automatic bypass of this selection step is prohibited.
- **Multilingual Data Fields**: `names`, localized addresses, and equivalent localized business fields must render exactly the active company languages from `settings.languageconfigs` in order. Normal data forms must not have add-language controls. Reducing active languages hides inputs only and must preserve hidden-language values on load/save. Do not auto-fill or fallback one business-data language value from another language.
- **Semantic Backgrounds**: Every major page needs a page-specific `.webp` background that communicates that screen's domain. Assets must be photorealistic, premium-camera quality, physically plausible, include presentable professional Thai people when natural for the page, and have clear depth of field. Only Codex may generate new background images; non-Codex agents must wire existing Codex-approved assets or request Codex generation.
- **Language Switch**: Flags themed grid (2 columns) for adding languages. Drag-to-reorder rows (first row is primary).
- **Custom Indicators**: High-contrast SVG/CSS styled boxes for read-only checkboxes (avoid disabled native checkboxes that are hard to read).
- **Image Previews**: Authenticated fetch for `/goapi/s3/file/...` to render browser-local object URLs (cache in session memory). Show preview next to original path.
- **Hydration Errors**: Check `mounted` state in `useEffect` before rendering window/client-side dynamic styles.
- **Root Scripts**: Use Next.js `next/script` with explicit `id` and `beforeInteractive` for bootstrap scripts. No raw `<script>` tags.

## 7. Pre-Commit Verification
- Fast-iteration dev mode: rely on `next dev` HMR while iterating. Run `npm run typecheck` in `frontend/` only before commit or when summarizing — not after every edit.
- Verify in both Light/Dark themes and check responsiveness in notebook/mobile.
