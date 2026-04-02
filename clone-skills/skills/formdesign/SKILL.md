---
name: formdesign
description: >
  WYSIWYG form template editor for ERP document templates. Row-based layout, element types,
  data bindings, styles, JSON templates.
  Triggers: "form design", "form template", "template JSON", "form editor", "form preview".
---

# Form Designer Skill

## Codebase Location
- **Flutter code**: `frontend/bcaiaccount/lib/screens/formdesign/`
- **Template assets**: `frontend/bcaiaccount/assets/form_templates/*.json`

## Architecture

### Dual Layout System
1. **Row-based (primary)**: `FormLayoutRow` > `FormLayoutCell` with flex/fixedWidth
2. **Fixed-position (legacy)**: `FormElement` with absolute x, y, width, height

Row-based is standard. `FormLayoutResolver.resolve(rows, sectionWidth)` converts rows to positioned elements.

### Model Hierarchy
```
FormTemplate
  |- header: FormSection (rows)
  |- detail: FormSection (items table)
  |- footer: FormSection (totals + signatures)
  |- headerMiddle/headerLast (optional multi-page)
  |- footerFirst/footerMiddle (optional multi-page)

FormLayoutRow > cells: List<FormLayoutCell> > resolves to FormElement
```

### Key Files
| File | Purpose |
|------|---------|
| `models/form_template.dart` | Template + paper + margins + multi-page |
| `models/form_section.dart` | Section container (header/detail/footer) |
| `models/form_element.dart` | Element + ElementType enum (14 types) |
| `models/form_layout.dart` | Row/Cell + flex resolver |
| `models/form_style.dart` | 4 visual styles |
| `models/data_binding.dart` | 64 binding fields across 8 categories |
| `models/editor_command.dart` | Undo/redo command pattern |
| `form_design.dart` | Main screen + keyboard shortcuts |
| `widgets/full_form_editor.dart` | WYSIWYG canvas (drag/resize/select) |
| `widgets/form_preview.dart` | Read-only scaled preview |
| `utils/form_template_factory.dart` | Load/save/cache templates |

## Template JSON Rules

### Row Layout
- Max 4 cells per row (prevents text overflow)
- Labels use `fixedWidth` (80-120px), data fields use `flex`
- Standard: `[label fw] [data flex] [label fw] [data flex]`
- Use `gap` property for spacing between signature columns

### Text Alignment
- Text/labels: `textAlign: 0` (left)
- Numbers/money: `textAlign: 2` (right)
- Tables: `columnAligns` array (0=left, 1=center, 2=right)

### Section Positioning
- Header: Y=0
- Detail: after header height
- Footer: anchored at `paperHeight - footerHeight`

### Element Types
`text`, `dataField`, `image`, `separator`, `line`, `table`, `barcode`, `qrCode`, `rectangle`, `pageNumber`, `dateTime`, `signature`

## FormStyle System

4 predefined styles overlaying colors/borders on any template:
- **Nopparat**: Classic formal, black, thick borders (0.8)
- **SenSai**: Modern minimal, gray/blue, thin borders (0.3)
- **Kram**: Premium indigo, medium borders (0.5)
- **Arun**: Warm teal/amber, borders (0.4)

Apply: `FormTemplateFactory.createTemplate('tax_invoice', styleId: 'kram')`

## Common Tasks

### Add new template
1. Create JSON in `assets/form_templates/{doctype}.json`
2. Add to `availableTypes` in `form_template_factory.dart`
3. Add to `templates_index.json`
4. Follow row rules: max 4 cells, label+data pairs

### Add new ElementType
1. Add to `ElementType` enum in `form_element.dart`
2. Add rendering in `full_form_editor.dart`, `form_preview.dart`, `section_editor.dart`
3. Add palette entry in `element_palette_panel.dart`
4. Add properties in `properties_inspector.dart`

### Add new data binding field
Add `DataBindingField` to appropriate category in `data_binding.dart` — auto-appears in UI.

### Fix layout issue
1. Check `FormLayoutResolver.resolve()` in `form_layout.dart`
2. Footer: `_sectionOffsetY()` in `full_form_editor.dart`
3. Detail height: `_detailDisplayHeight()` fills between header and footer

## References
- [Data Bindings](references/data-bindings.md) — all 64 fields across 8 categories
- [Template Patterns](references/template-patterns.md) — row patterns + JSON examples
