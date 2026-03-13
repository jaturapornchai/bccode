---
name: formdesign
description: >
  Form Designer (WYSIWYG document template editor) skill for creating, editing,
  and maintaining form templates used in the ERP system. Use when: (1) modifying
  formdesign Flutter code (models, widgets, canvas, panels), (2) creating or editing
  JSON template files in assets/form_templates/, (3) adding new element types or
  data bindings, (4) fixing rendering/layout issues, (5) working on form styles,
  (6) any task involving the formdesign/ directory or form template JSON files.
  Triggers on: "form design", "form template", "formdesign", "template JSON",
  "ใบกำกับภาษี template", "แบบฟอร์ม", "form editor", "form preview".
---

# Form Designer Skill

## Codebase Location
- **Flutter code**: `frontend/bcaiaccount/lib/screens/formdesign/`
- **Template assets**: `frontend/bcaiaccount/assets/form_templates/*.json`

## Architecture Overview

### Dual Layout System
1. **Row-based (primary)**: `FormLayoutRow` > `FormLayoutCell` with flex/fixedWidth (like Flutter Row/Expanded)
2. **Fixed-position (legacy)**: `FormElement` with absolute x, y, width, height

Row-based is the standard for all templates. `FormLayoutResolver.resolve(rows, sectionWidth)` converts rows into positioned `FormElement` list for rendering.

### Model Hierarchy
```
FormTemplate
  ├── header: FormSection (rows: List<FormLayoutRow>)
  ├── detail: FormSection (items table)
  ├── footer: FormSection (totals + signatures)
  ├── headerMiddle/headerLast (optional multi-page)
  └── footerFirst/footerMiddle (optional multi-page)

FormLayoutRow
  └── cells: List<FormLayoutCell>  (flex or fixedWidth sizing)

FormLayoutCell → resolves to → FormElement (for rendering)
```

### Key Files
| File | Purpose |
|------|---------|
| `models/form_template.dart` | Template + paper + margins + multi-page |
| `models/form_section.dart` | Section container (header/detail/footer) |
| `models/form_element.dart` | Element + ElementType enum (14 types) |
| `models/form_layout.dart` | Row/Cell + flex resolver |
| `models/form_style.dart` | 4 visual styles (Nopparat/SenSai/Kram/Arun) |
| `models/data_binding.dart` | 64 binding fields across 8 categories |
| `models/editor_command.dart` | Undo/redo command pattern |
| `form_design.dart` | Main screen state + keyboard shortcuts |
| `widgets/full_form_editor.dart` | WYSIWYG canvas (drag/resize/select) |
| `widgets/form_preview.dart` | Read-only scaled preview |
| `utils/form_template_factory.dart` | Load/save/cache templates from JSON |

## Template JSON Rules (CRITICAL)

### Row Layout Rules
1. **Max 4 cells per row** - prevents text overflow
2. Labels use `fixedWidth` (80-120px), data fields use `flex`
3. Standard pattern: `[label fw] [data flex] [label fw] [data flex]`
4. Use `gap` property on row for spacing between signature columns (not spacer cells)

### Text Alignment
- Text/labels: `textAlign: 0` (left)
- Numbers/money: `textAlign: 2` (right)
- Table: use `columnAligns` array (0=left, 1=center, 2=right per column)
- Column 0 (No.): center, Description: left, Qty/Price/Amount: right

### Section Positioning
- **Header**: starts at Y=0
- **Detail**: starts after header height
- **Footer**: anchored at `paperHeight - footerHeight` (bottom-up)
- Detail fills remaining space between header and footer

### Element Types
`text`, `dataField`, `image`, `separator`, `line`, `table`, `barcode`, `qrCode`, `rectangle`, `pageNumber`, `dateTime`, `signature`

### Template Structure
```json
{
  "id": "tax_invoice",
  "name": "ใบกำกับภาษี",
  "paperSize": "A4",
  "orientation": "portrait",
  "marginTop": 20, "marginBottom": 15, "marginLeft": 20, "marginRight": 20,
  "multiPageEnabled": true,
  "detailRowsPerPage": 20,
  "header": { "id": "header", "type": "header", "height": 160, "rows": [...] },
  "detail": { "id": "detail", "type": "detail", "height": 400, "rows": [...] },
  "footer": { "id": "footer", "type": "footer", "height": 180, "rows": [...] }
}
```

### Cell Structure
```json
{
  "id": "unique_id",
  "type": "text|dataField|separator|table|image",
  "flex": 1.0,
  "fixedWidth": 80,
  "text": "label text",
  "dataBindingKey": "field_name",
  "fontFamily": "Sarabun",
  "fontSize": 7,
  "fontBold": false,
  "textAlign": 0,
  "textColor": 4278190080,
  "borderColor": 4278190080,
  "borderWidth": 0.5
}
```

## FormStyle System

4 predefined styles that overlay colors/borders on any template:
- **นพรัตน์ (nopparat)**: Classic formal, black, thick borders (0.8), solid
- **เส้นสาย (sensai)**: Modern minimal, gray/blue, thin borders (0.3), dotted
- **คราม (kram)**: Premium indigo, medium borders (0.5), dashed
- **อรุณ (arun)**: Warm teal/amber, borders (0.4), solid

Apply via: `FormTemplateFactory.createTemplate('tax_invoice', styleId: 'kram')`

Style rules in `_applyToCell()`:
- Title (bold + fontSize >= 10) gets `titleColor`
- Labels get `labelColor`
- DataFields get `textColor`
- Separators/tables get `borderColor` + `lineStyle`

## References
- **Data binding fields**: See [references/data-bindings.md](references/data-bindings.md) for all 64 fields
- **Template examples**: See [references/template-patterns.md](references/template-patterns.md) for row patterns

## Common Tasks

### Add new template
1. Create JSON in `assets/form_templates/{doctype}.json`
2. Add to `availableTypes` in `form_template_factory.dart`
3. Add to `templates_index.json`
4. Follow row rules: max 4 cells, label+data pairs

### Add new ElementType
1. Add to `ElementType` enum in `form_element.dart`
2. Add rendering in `_buildElement()` in `full_form_editor.dart`
3. Add rendering in `form_preview.dart` and `section_editor.dart`
4. Add palette entry in `element_palette_panel.dart`
5. Add properties in `properties_inspector.dart`

### Add new data binding field
1. Add `DataBindingField` to appropriate category in `data_binding.dart`
2. Field auto-appears in binding picker UI

### Fix layout issue
1. Check `FormLayoutResolver.resolve()` in `form_layout.dart`
2. Footer positioning: `_sectionOffsetY()` in `full_form_editor.dart`
3. Detail height: `_detailDisplayHeight()` fills between header and footer
