import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../../models/form_element.dart';
import '../../models/form_section.dart';
import '../../models/data_binding.dart';

// ============================================================
// PropertiesInspector — Right panel for editing selected element
// ============================================================

class PropertiesInspector extends StatefulWidget {
  final FormElement? selectedElement;
  final SectionType? selectedSection;
  final Function(FormElement updatedElement) onElementChanged;
  final Function() onDeleteElement;

  const PropertiesInspector({
    super.key,
    required this.selectedElement,
    this.selectedSection,
    required this.onElementChanged,
    required this.onDeleteElement,
  });

  @override
  State<PropertiesInspector> createState() => _PropertiesInspectorState();
}

class _PropertiesInspectorState extends State<PropertiesInspector>
    with global.ThemeRefreshMixin {
  bool _lockAspectRatio = false;
  double _aspectRatio = 1.0;

  @override
  void didUpdateWidget(covariant PropertiesInspector oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.selectedElement?.id != oldWidget.selectedElement?.id) {
      _lockAspectRatio = false;
      final el = widget.selectedElement;
      if (el != null && el.height > 0) {
        _aspectRatio = el.width / el.height;
      }
    }
  }

  void _update(FormElement Function(FormElement e) updater) {
    if (widget.selectedElement == null) return;
    widget.onElementChanged(updater(widget.selectedElement!));
  }

  bool get _hasTextStyle {
    final t = widget.selectedElement?.type;
    return t == ElementType.text ||
        t == ElementType.dataField ||
        t == ElementType.pageNumber ||
        t == ElementType.dateTime;
  }

  bool get _isDataField =>
      widget.selectedElement?.type == ElementType.dataField;

  bool get _isTable => widget.selectedElement?.type == ElementType.table;

  bool get _hasContent =>
      widget.selectedElement?.type == ElementType.text || _isDataField;

  @override
  Widget build(BuildContext context) {
    final el = widget.selectedElement;

    return Container(
      width: 280,
      color: global.theme.cardColor,
      child: el == null ? _buildEmptyState() : _buildInspector(el),
    );
  }

  // ── Empty state ──────────────────────────────────────────

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.touch_app_outlined,
                size: 48, color: global.theme.textSecondaryColor),
            const SizedBox(height: 12),
            Text(
              global.language('fd_select_element_hint'),
              textAlign: TextAlign.center,
              style: TextStyle(
                color: global.theme.textSecondaryColor,
                fontSize: 13,
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ── Inspector body ───────────────────────────────────────

  Widget _buildInspector(FormElement el) {
    return Column(
      children: [
        _buildHeader(el),
        Expanded(
          child: ListView(
            padding: const EdgeInsets.only(bottom: 24),
            children: [
              _PositionSizeSection(
                element: el,
                lockAspectRatio: _lockAspectRatio,
                onLockChanged: (v) {
                  setState(() {
                    _lockAspectRatio = v;
                    if (v && el.height > 0) {
                      _aspectRatio = el.width / el.height;
                    }
                  });
                },
                onChanged: (updated) {
                  if (_lockAspectRatio) {
                    if (updated.width != el.width) {
                      updated = updated.copyWith(
                          height: updated.width / _aspectRatio);
                    } else if (updated.height != el.height) {
                      updated = updated.copyWith(
                          width: updated.height * _aspectRatio);
                    }
                  }
                  widget.onElementChanged(updated);
                },
              ),
              if (_hasTextStyle)
                _TextStyleSection(
                  element: el,
                  onChanged: widget.onElementChanged,
                ),
              _AppearanceSection(
                element: el,
                onChanged: widget.onElementChanged,
              ),
              if (_isDataField)
                _DataBindingSection(
                  element: el,
                  onChanged: widget.onElementChanged,
                ),
              if (_isTable)
                _TableSection(
                  element: el,
                  onChanged: widget.onElementChanged,
                ),
              if (_hasContent)
                _ContentSection(
                  element: el,
                  onChanged: widget.onElementChanged,
                ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildHeader(FormElement el) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border(
          bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1),
        ),
      ),
      child: Row(
        children: [
          Icon(el.type.icon, size: 18, color: global.theme.iconColor),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              el.type.label,
              style: TextStyle(
                color: global.theme.textColor,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          _buildHeaderAction(
            icon: el.isVisible ? Icons.visibility : Icons.visibility_off,
            tooltip: el.isVisible ? global.language('fd_hide') : global.language('fd_show'),
            onTap: () => _update((e) => e.copyWith(isVisible: !e.isVisible)),
          ),
          const SizedBox(width: 4),
          _buildHeaderAction(
            icon: el.isLocked ? Icons.lock : Icons.lock_open,
            tooltip: el.isLocked ? global.language('fd_unlock') : global.language('fd_lock'),
            onTap: () => _update((e) => e.copyWith(isLocked: !e.isLocked)),
          ),
          const SizedBox(width: 4),
          _buildHeaderAction(
            icon: Icons.delete_outline,
            tooltip: global.language('fd_delete'),
            onTap: widget.onDeleteElement,
            color: global.theme.buttonDangerColor,
          ),
        ],
      ),
    );
  }

  Widget _buildHeaderAction({
    required IconData icon,
    required String tooltip,
    required VoidCallback onTap,
    Color? color,
  }) {
    return Tooltip(
      message: tooltip,
      child: InkWell(
        borderRadius: BorderRadius.circular(4),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(4),
          child: Icon(icon, size: 18, color: color ?? global.theme.iconColor),
        ),
      ),
    );
  }
}

// ============================================================
// Shared helpers
// ============================================================

const _kSectionPadding = EdgeInsets.symmetric(horizontal: 12, vertical: 4);

Widget _buildNumberField({
  required String label,
  required double value,
  required ValueChanged<double> onChanged,
  double min = 0,
  double max = 9999,
  double step = 1,
}) {
  return Row(
    children: [
      SizedBox(
        width: 28,
        child: Text(
          label,
          style: TextStyle(color: global.theme.textSecondaryColor, fontSize: 11),
        ),
      ),
      _StepButton(
        icon: Icons.remove,
        onTap: () {
          final v = (value - step).clamp(min, max);
          onChanged(v);
        },
      ),
      Expanded(
        child: Container(
          height: 28,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(4),
            border: Border.all(color: global.theme.dividerBorderColor, width: 0.5),
          ),
          child: EditableText(
            controller: TextEditingController(text: value.toStringAsFixed(
                value == value.roundToDouble() ? 0 : 1)),
            focusNode: FocusNode(),
            style: TextStyle(color: global.theme.textColor, fontSize: 12),
            cursorColor: global.theme.primaryColor,
            backgroundCursorColor: global.theme.surfaceColor,
            textAlign: TextAlign.center,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            onSubmitted: (text) {
              final parsed = double.tryParse(text);
              if (parsed != null) {
                onChanged(parsed.clamp(min, max));
              }
            },
          ),
        ),
      ),
      _StepButton(
        icon: Icons.add,
        onTap: () {
          final v = (value + step).clamp(min, max);
          onChanged(v);
        },
      ),
    ],
  );
}

class _StepButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback onTap;

  const _StepButton({required this.icon, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(4),
      onTap: onTap,
      child: Container(
        width: 24,
        height: 28,
        alignment: Alignment.center,
        child: Icon(icon, size: 14, color: global.theme.iconColor),
      ),
    );
  }
}

Widget _buildSectionTile({
  required IconData icon,
  required String title,
  required List<Widget> children,
  bool initiallyExpanded = true,
}) {
  return Theme(
    data: ThemeData(dividerColor: Colors.transparent),
    child: ExpansionTile(
      leading: Icon(icon, size: 18, color: global.theme.iconColor),
      title: Text(
        title,
        style: TextStyle(
          color: global.theme.textColor,
          fontSize: 12,
          fontWeight: FontWeight.w600,
        ),
      ),
      tilePadding: const EdgeInsets.symmetric(horizontal: 12),
      childrenPadding: _kSectionPadding,
      initiallyExpanded: initiallyExpanded,
      iconColor: global.theme.iconColor,
      collapsedIconColor: global.theme.iconSecondaryColor,
      children: children,
    ),
  );
}

// ── Color picker chips ─────────────────────────────────────

const _kPresetColors = <_ColorPreset>[
  _ColorPreset('Black', Color(0xFF000000)),
  _ColorPreset('White', Color(0xFFFFFFFF)),
  _ColorPreset('Red', Color(0xFFE53935)),
  _ColorPreset('Blue', Color(0xFF1E88E5)),
  _ColorPreset('Green', Color(0xFF43A047)),
  _ColorPreset('Orange', Color(0xFFFB8C00)),
  _ColorPreset('Grey', Color(0xFF757575)),
];

class _ColorPreset {
  final String name;
  final Color color;
  const _ColorPreset(this.name, this.color);
}

Widget _buildColorChips({
  required Color? selected,
  required ValueChanged<Color?> onChanged,
  bool allowTransparent = false,
}) {
  return Wrap(
    spacing: 6,
    runSpacing: 6,
    children: [
      if (allowTransparent)
        _buildSingleColorChip(
          color: Colors.transparent,
          label: 'None',
          isSelected: selected == null || selected == Colors.transparent,
          onTap: () => onChanged(Colors.transparent),
          showBorder: true,
        ),
      for (final preset in _kPresetColors)
        _buildSingleColorChip(
          color: preset.color,
          label: preset.name,
          isSelected: selected?.toARGB32() == preset.color.toARGB32(),
          onTap: () => onChanged(preset.color),
        ),
    ],
  );
}

Widget _buildSingleColorChip({
  required Color color,
  required String label,
  required bool isSelected,
  required VoidCallback onTap,
  bool showBorder = false,
}) {
  return Tooltip(
    message: label,
    child: GestureDetector(
      onTap: onTap,
      child: Container(
        width: 24,
        height: 24,
        decoration: BoxDecoration(
          color: color,
          shape: BoxShape.circle,
          border: Border.all(
            color: isSelected
                ? global.theme.primaryColor
                : (showBorder
                    ? global.theme.dividerBorderColor
                    : color == const Color(0xFFFFFFFF)
                        ? global.theme.dividerBorderColor
                        : Colors.transparent),
            width: isSelected ? 2.5 : 1,
          ),
        ),
        child: color == Colors.transparent
            ? Icon(Icons.block, size: 14, color: global.theme.textSecondaryColor)
            : null,
      ),
    ),
  );
}

Widget _buildFieldLabel(String text) {
  return Padding(
    padding: const EdgeInsets.only(top: 8, bottom: 4),
    child: Text(
      text,
      style: TextStyle(
        color: global.theme.textSecondaryColor,
        fontSize: 11,
        fontWeight: FontWeight.w500,
      ),
    ),
  );
}

// ============================================================
// 1. Position & Size Section
// ============================================================

class _PositionSizeSection extends StatelessWidget {
  final FormElement element;
  final bool lockAspectRatio;
  final ValueChanged<bool> onLockChanged;
  final ValueChanged<FormElement> onChanged;

  const _PositionSizeSection({
    required this.element,
    required this.lockAspectRatio,
    required this.onLockChanged,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return _buildSectionTile(
      icon: Icons.open_with,
      title: global.language('fd_position_size'),
      children: [
        Row(
          children: [
            Expanded(
              child: _buildNumberField(
                label: 'X',
                value: element.x,
                onChanged: (v) => onChanged(element.copyWith(x: v)),
                min: -9999,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _buildNumberField(
                label: 'Y',
                value: element.y,
                onChanged: (v) => onChanged(element.copyWith(y: v)),
                min: -9999,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        Row(
          children: [
            Expanded(
              child: _buildNumberField(
                label: 'W',
                value: element.width,
                onChanged: (v) => onChanged(element.copyWith(width: v)),
                min: 1,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _buildNumberField(
                label: 'H',
                value: element.height,
                onChanged: (v) => onChanged(element.copyWith(height: v)),
                min: 1,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Row(
          children: [
            Icon(Icons.link, size: 14, color: global.theme.iconSecondaryColor),
            const SizedBox(width: 4),
            Text(
              global.language('fd_lock_aspect_ratio'),
              style: TextStyle(
                  color: global.theme.textSecondaryColor, fontSize: 11),
            ),
            const Spacer(),
            SizedBox(
              height: 24,
              child: Switch(
                value: lockAspectRatio,
                onChanged: onLockChanged,
                activeThumbColor: global.theme.primaryColor,
                materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        _buildFieldLabel(global.language('fd_rotation')),
        Row(
          children: [
            Expanded(
              child: SliderTheme(
                data: SliderThemeData(
                  trackHeight: 2,
                  thumbShape:
                      const RoundSliderThumbShape(enabledThumbRadius: 6),
                  activeTrackColor: global.theme.primaryColor,
                  inactiveTrackColor: global.theme.dividerBorderColor,
                  thumbColor: global.theme.primaryColor,
                  overlayShape:
                      const RoundSliderOverlayShape(overlayRadius: 12),
                ),
                child: Slider(
                  value: element.rotation,
                  min: 0,
                  max: 360,
                  divisions: 72,
                  onChanged: (v) => onChanged(element.copyWith(rotation: v)),
                ),
              ),
            ),
            SizedBox(
              width: 40,
              child: Text(
                '${element.rotation.round()}\u00B0',
                textAlign: TextAlign.right,
                style:
                    TextStyle(color: global.theme.textColor, fontSize: 11),
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
      ],
    );
  }
}

// ============================================================
// 2. Text Style Section
// ============================================================

/// Font families ตรงกับ backend availableFonts map (gen-trans-pdf/common.go)
/// key = ชื่อที่ส่งให้ backend, label = ชื่อที่แสดงใน UI
const _kFontFamilies = <Map<String, String>>[
  {'key': 'Sarabun', 'label': 'Sarabun (ซาราบัน)'},
  {'key': 'Kanit', 'label': 'Kanit (คณิต)'},
  {'key': 'Prompt', 'label': 'Prompt (พรอมต์)'},
  {'key': 'Mitr', 'label': 'Mitr (มิตร)'},
  {'key': 'NotoSansThai', 'label': 'Noto Sans Thai'},
  {'key': 'NotoSansThai-Light', 'label': 'Noto Sans Thai Light'},
  {'key': 'NotoSansThai-Condensed', 'label': 'Noto Sans Thai Condensed'},
  {'key': 'NotoSansThai-SemiCondensed', 'label': 'Noto Sans Thai Semi-Condensed'},
  {'key': 'GoNotoCurrent', 'label': 'Go Noto (Universal)'},
  {'key': 'NotoSansCJKsc', 'label': 'Noto Sans CJK (จีน)'},
  {'key': 'NotoSansCJKjp', 'label': 'Noto Sans CJK (ญี่ปุ่น)'},
  {'key': 'NotoSansCJKkr', 'label': 'Noto Sans CJK (เกาหลี)'},
  {'key': 'NotoSansKhmer', 'label': 'Noto Sans Khmer'},
  {'key': 'NotoSansLao', 'label': 'Noto Sans Lao'},
  {'key': 'NotoSansMyanmar', 'label': 'Noto Sans Myanmar'},
];

class _TextStyleSection extends StatelessWidget {
  final FormElement element;
  final ValueChanged<FormElement> onChanged;

  const _TextStyleSection({
    required this.element,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final fontFamily = element.fontFamily ?? 'Sarabun';
    final fontKeys = _kFontFamilies.map((f) => f['key']!).toList();
    final fontSize = element.fontSize ?? 14.0;
    final align = element.textAlign ?? TextAlign.left;

    return _buildSectionTile(
      icon: Icons.text_format,
      title: global.language('fd_text_style'),
      children: [
        // Font family
        _buildFieldLabel(global.language('fd_font_family')),
        Container(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(4),
            border: Border.all(
                color: global.theme.dividerBorderColor, width: 0.5),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<String>(
              value: fontKeys.contains(fontFamily) ? fontFamily : 'Sarabun',
              isExpanded: true,
              isDense: true,
              dropdownColor: global.theme.cardColor,
              style: TextStyle(color: global.theme.textColor, fontSize: 12),
              icon: Icon(Icons.arrow_drop_down,
                  size: 18, color: global.theme.iconColor),
              items: _kFontFamilies
                  .map((f) => DropdownMenuItem(
                        value: f['key']!,
                        child: Text(f['label']!,
                            style: TextStyle(
                                color: global.theme.textColor, fontSize: 12)),
                      ))
                  .toList(),
              onChanged: (v) {
                if (v != null) {
                  onChanged(element.copyWith(fontFamily: v));
                }
              },
            ),
          ),
        ),
        const SizedBox(height: 8),

        // Font size
        _buildNumberField(
          label: 'Size',
          value: fontSize,
          onChanged: (v) => onChanged(element.copyWith(fontSize: v)),
          min: 8,
          max: 72,
          step: 1,
        ),
        const SizedBox(height: 8),

        // Bold / Italic / Underline
        _buildFieldLabel(global.language('fd_text_style')),
        Row(
          children: [
            _ToggleStyleButton(
              icon: Icons.format_bold,
              isActive: element.fontBold,
              onTap: () =>
                  onChanged(element.copyWith(fontBold: !element.fontBold)),
            ),
            const SizedBox(width: 4),
            _ToggleStyleButton(
              icon: Icons.format_italic,
              isActive: element.fontItalic,
              onTap: () =>
                  onChanged(element.copyWith(fontItalic: !element.fontItalic)),
            ),
            const SizedBox(width: 4),
            _ToggleStyleButton(
              icon: Icons.format_underlined,
              isActive: element.fontUnderline,
              onTap: () => onChanged(
                  element.copyWith(fontUnderline: !element.fontUnderline)),
            ),
            const Spacer(),
          ],
        ),
        const SizedBox(height: 8),

        // Text alignment
        _buildFieldLabel(global.language('fd_alignment')),
        Row(
          children: [
            _ToggleStyleButton(
              icon: Icons.format_align_left,
              isActive: align == TextAlign.left,
              onTap: () =>
                  onChanged(element.copyWith(textAlign: TextAlign.left)),
            ),
            const SizedBox(width: 4),
            _ToggleStyleButton(
              icon: Icons.format_align_center,
              isActive: align == TextAlign.center,
              onTap: () =>
                  onChanged(element.copyWith(textAlign: TextAlign.center)),
            ),
            const SizedBox(width: 4),
            _ToggleStyleButton(
              icon: Icons.format_align_right,
              isActive: align == TextAlign.right,
              onTap: () =>
                  onChanged(element.copyWith(textAlign: TextAlign.right)),
            ),
            const SizedBox(width: 4),
            _ToggleStyleButton(
              icon: Icons.format_align_justify,
              isActive: align == TextAlign.justify,
              onTap: () =>
                  onChanged(element.copyWith(textAlign: TextAlign.justify)),
            ),
            const Spacer(),
          ],
        ),
        const SizedBox(height: 8),

        // Text color
        _buildFieldLabel(global.language('fd_text_color')),
        _buildColorChips(
          selected: element.textColor,
          onChanged: (c) =>
              onChanged(element.copyWith(textColor: c ?? const Color(0xFF000000))),
        ),
        const SizedBox(height: 8),
      ],
    );
  }
}

class _ToggleStyleButton extends StatelessWidget {
  final IconData icon;
  final bool isActive;
  final VoidCallback onTap;

  const _ToggleStyleButton({
    required this.icon,
    required this.isActive,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(4),
      onTap: onTap,
      child: Container(
        width: 32,
        height: 28,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: isActive
              ? global.theme.primaryColor.withValues(alpha: 0.15)
              : global.theme.surfaceColor,
          borderRadius: BorderRadius.circular(4),
          border: Border.all(
            color: isActive
                ? global.theme.primaryColor
                : global.theme.dividerBorderColor,
            width: isActive ? 1.5 : 0.5,
          ),
        ),
        child: Icon(
          icon,
          size: 16,
          color: isActive ? global.theme.primaryColor : global.theme.iconColor,
        ),
      ),
    );
  }
}

// ============================================================
// 3. Appearance Section
// ============================================================

class _AppearanceSection extends StatelessWidget {
  final FormElement element;
  final ValueChanged<FormElement> onChanged;

  const _AppearanceSection({
    required this.element,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return _buildSectionTile(
      icon: Icons.palette_outlined,
      title: global.language('fd_appearance'),
      initiallyExpanded: false,
      children: [
        // Background color
        _buildFieldLabel(global.language('fd_background')),
        _buildColorChips(
          selected: element.backgroundColor,
          onChanged: (c) => onChanged(element.copyWith(
              backgroundColor: c ?? Colors.transparent)),
          allowTransparent: true,
        ),
        const SizedBox(height: 8),

        // Border color
        _buildFieldLabel(global.language('fd_border_color')),
        _buildColorChips(
          selected: element.borderColor,
          onChanged: (c) => onChanged(element.copyWith(
              borderColor: c ?? Colors.transparent)),
          allowTransparent: true,
        ),
        const SizedBox(height: 8),

        // Border width
        _buildFieldLabel(global.language('fd_border_width')),
        _buildSliderRow(
          value: element.borderWidth ?? 0,
          min: 0,
          max: 5,
          divisions: 10,
          suffix: 'px',
          onChanged: (v) => onChanged(element.copyWith(borderWidth: v)),
        ),
        const SizedBox(height: 4),

        // Border radius
        _buildFieldLabel(global.language('fd_border_radius')),
        _buildSliderRow(
          value: element.borderRadius ?? 0,
          min: 0,
          max: 20,
          divisions: 20,
          suffix: 'px',
          onChanged: (v) => onChanged(element.copyWith(borderRadius: v)),
        ),
        const SizedBox(height: 4),

        // Opacity
        _buildFieldLabel(global.language('fd_opacity')),
        _buildSliderRow(
          value: element.opacity,
          min: 0,
          max: 1,
          divisions: 20,
          suffix: '${(element.opacity * 100).round()}%',
          onChanged: (v) => onChanged(element.copyWith(opacity: v)),
        ),
        const SizedBox(height: 8),
      ],
    );
  }

  Widget _buildSliderRow({
    required double value,
    required double min,
    required double max,
    required int divisions,
    required String suffix,
    required ValueChanged<double> onChanged,
  }) {
    return Row(
      children: [
        Expanded(
          child: SliderTheme(
            data: SliderThemeData(
              trackHeight: 2,
              thumbShape:
                  const RoundSliderThumbShape(enabledThumbRadius: 6),
              activeTrackColor: global.theme.primaryColor,
              inactiveTrackColor: global.theme.dividerBorderColor,
              thumbColor: global.theme.primaryColor,
              overlayShape:
                  const RoundSliderOverlayShape(overlayRadius: 12),
            ),
            child: Slider(
              value: value,
              min: min,
              max: max,
              divisions: divisions,
              onChanged: onChanged,
            ),
          ),
        ),
        SizedBox(
          width: 40,
          child: Text(
            suffix,
            textAlign: TextAlign.right,
            style: TextStyle(color: global.theme.textColor, fontSize: 11),
          ),
        ),
      ],
    );
  }
}

// ============================================================
// 4. Data Binding Section
// ============================================================

class _DataBindingSection extends StatefulWidget {
  final FormElement element;
  final ValueChanged<FormElement> onChanged;

  const _DataBindingSection({
    required this.element,
    required this.onChanged,
  });

  @override
  State<_DataBindingSection> createState() => _DataBindingSectionState();
}

class _DataBindingSectionState extends State<_DataBindingSection> {
  String? _selectedCategoryId;

  @override
  void initState() {
    super.initState();
    _initCategory();
  }

  @override
  void didUpdateWidget(covariant _DataBindingSection oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.element.dataBindingKey != oldWidget.element.dataBindingKey) {
      _initCategory();
    }
  }

  void _initCategory() {
    final key = widget.element.dataBindingKey;
    if (key != null) {
      final field = DataBindingRegistry.findField(key);
      if (field != null) {
        _selectedCategoryId = field.category;
        return;
      }
    }
    _selectedCategoryId =
        DataBindingRegistry.categories.isNotEmpty
            ? DataBindingRegistry.categories.first.id
            : null;
  }

  List<DataBindingField> get _filteredFields {
    if (_selectedCategoryId == null) return [];
    final cat = DataBindingRegistry.categories
        .where((c) => c.id == _selectedCategoryId);
    if (cat.isEmpty) return [];
    return cat.first.fields;
  }

  @override
  Widget build(BuildContext context) {
    final bindingKey = widget.element.dataBindingKey;
    final sampleValue =
        bindingKey != null ? DataBindingRegistry.getSampleValue(bindingKey) : '';

    return _buildSectionTile(
      icon: Icons.data_object,
      title: global.language('fd_data_binding'),
      children: [
        // Category
        _buildFieldLabel('Category'),
        Container(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(4),
            border: Border.all(
                color: global.theme.dividerBorderColor, width: 0.5),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<String>(
              value: _selectedCategoryId,
              isExpanded: true,
              isDense: true,
              dropdownColor: global.theme.cardColor,
              style: TextStyle(color: global.theme.textColor, fontSize: 12),
              icon: Icon(Icons.arrow_drop_down,
                  size: 18, color: global.theme.iconColor),
              items: DataBindingRegistry.categories
                  .map((c) => DropdownMenuItem(
                        value: c.id,
                        child: Text('${c.labelEn} (${c.label})',
                            style: TextStyle(
                                color: global.theme.textColor, fontSize: 12)),
                      ))
                  .toList(),
              onChanged: (v) {
                setState(() => _selectedCategoryId = v);
              },
            ),
          ),
        ),
        const SizedBox(height: 8),

        // Field
        _buildFieldLabel(global.language('fd_data_field')),
        Container(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(4),
            border: Border.all(
                color: global.theme.dividerBorderColor, width: 0.5),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<String>(
              value: _filteredFields.any((f) => f.key == bindingKey)
                  ? bindingKey
                  : null,
              isExpanded: true,
              isDense: true,
              dropdownColor: global.theme.cardColor,
              style: TextStyle(color: global.theme.textColor, fontSize: 12),
              icon: Icon(Icons.arrow_drop_down,
                  size: 18, color: global.theme.iconColor),
              hint: Text(global.language('fd_select_variable'),
                  style: TextStyle(
                      color: global.theme.textSecondaryColor, fontSize: 12)),
              items: _filteredFields
                  .map((f) => DropdownMenuItem(
                        value: f.key,
                        child: Text('${f.labelEn} (${f.label})',
                            style: TextStyle(
                                color: global.theme.textColor, fontSize: 12)),
                      ))
                  .toList(),
              onChanged: (v) {
                if (v != null) {
                  final field = DataBindingRegistry.findField(v);
                  widget.onChanged(widget.element.copyWith(
                    dataBindingKey: v,
                    dataBindingFormat: field?.defaultFormat,
                  ));
                }
              },
            ),
          ),
        ),
        const SizedBox(height: 8),

        // Format string
        _buildFieldLabel(global.language('fd_format')),
        Container(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(4),
            border: Border.all(
                color: global.theme.dividerBorderColor, width: 0.5),
          ),
          child: EditableText(
            controller: TextEditingController(
                text: widget.element.dataBindingFormat ?? ''),
            focusNode: FocusNode(),
            style: TextStyle(color: global.theme.textColor, fontSize: 12),
            cursorColor: global.theme.primaryColor,
            backgroundCursorColor: global.theme.surfaceColor,
            onSubmitted: (v) {
              widget.onChanged(widget.element.copyWith(dataBindingFormat: v));
            },
          ),
        ),
        const SizedBox(height: 8),

        // Preview
        if (sampleValue.isNotEmpty) ...[
          _buildFieldLabel('Preview'),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(
                  color: global.theme.primaryColor.withValues(alpha: 0.3),
                  width: 1),
            ),
            child: Text(
              sampleValue,
              style: TextStyle(
                color: global.theme.primaryColor,
                fontSize: 12,
                fontStyle: FontStyle.italic,
              ),
            ),
          ),
        ],
        const SizedBox(height: 8),
      ],
    );
  }
}

// ============================================================
// 5. Table Section
// ============================================================

class _TableSection extends StatelessWidget {
  final FormElement element;
  final ValueChanged<FormElement> onChanged;

  const _TableSection({
    required this.element,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final rows = element.rows ?? 3;
    final columns = element.columns ?? 3;

    return _buildSectionTile(
      icon: Icons.table_chart_outlined,
      title: global.language('fd_table'),
      children: [
        _buildNumberField(
          label: 'Rows',
          value: rows.toDouble(),
          onChanged: (v) {
            final newRows = v.round().clamp(1, 50);
            final data = _resizeTableData(
                element.tableData, newRows, columns);
            onChanged(element.copyWith(rows: newRows, tableData: data));
          },
          min: 1,
          max: 50,
          step: 1,
        ),
        const SizedBox(height: 6),
        _buildNumberField(
          label: 'Cols',
          value: columns.toDouble(),
          onChanged: (v) {
            final newCols = v.round().clamp(1, 20);
            final data = _resizeTableData(
                element.tableData, rows, newCols);
            final widths = _resizeColumnWidths(
                element.columnWidths, newCols, element.width);
            onChanged(element.copyWith(
                columns: newCols, tableData: data, columnWidths: widths));
          },
          min: 1,
          max: 20,
          step: 1,
        ),
        const SizedBox(height: 8),

        // Has header row
        Row(
          children: [
            Icon(Icons.table_rows_outlined,
                size: 14, color: global.theme.iconSecondaryColor),
            const SizedBox(width: 4),
            Text(
              'Header row',
              style: TextStyle(
                  color: global.theme.textSecondaryColor, fontSize: 11),
            ),
            const Spacer(),
            SizedBox(
              height: 24,
              child: Switch(
                value: element.tableHasHeader,
                onChanged: (v) =>
                    onChanged(element.copyWith(tableHasHeader: v)),
                activeThumbColor: global.theme.primaryColor,
                materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),

        // Quick actions
        Row(
          children: [
            Expanded(
              child: _TableActionButton(
                icon: Icons.add,
                label: '+ Row',
                onTap: () {
                  final newRows = (rows + 1).clamp(1, 50);
                  final data = _resizeTableData(
                      element.tableData, newRows, columns);
                  onChanged(element.copyWith(rows: newRows, tableData: data));
                },
              ),
            ),
            const SizedBox(width: 4),
            Expanded(
              child: _TableActionButton(
                icon: Icons.remove,
                label: '- Row',
                onTap: rows > 1
                    ? () {
                        final newRows = (rows - 1).clamp(1, 50);
                        final data = _resizeTableData(
                            element.tableData, newRows, columns);
                        onChanged(
                            element.copyWith(rows: newRows, tableData: data));
                      }
                    : null,
              ),
            ),
            const SizedBox(width: 4),
            Expanded(
              child: _TableActionButton(
                icon: Icons.add,
                label: '+ Col',
                onTap: () {
                  final newCols = (columns + 1).clamp(1, 20);
                  final data = _resizeTableData(
                      element.tableData, rows, newCols);
                  final widths = _resizeColumnWidths(
                      element.columnWidths, newCols, element.width);
                  onChanged(element.copyWith(
                      columns: newCols, tableData: data, columnWidths: widths));
                },
              ),
            ),
            const SizedBox(width: 4),
            Expanded(
              child: _TableActionButton(
                icon: Icons.remove,
                label: '- Col',
                onTap: columns > 1
                    ? () {
                        final newCols = (columns - 1).clamp(1, 20);
                        final data = _resizeTableData(
                            element.tableData, rows, newCols);
                        final widths = _resizeColumnWidths(
                            element.columnWidths, newCols, element.width);
                        onChanged(element.copyWith(
                            columns: newCols,
                            tableData: data,
                            columnWidths: widths));
                      }
                    : null,
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
      ],
    );
  }

  List<List<String>> _resizeTableData(
      List<List<String>>? existing, int rows, int cols) {
    final data = existing ?? [];
    return List.generate(rows, (r) {
      if (r < data.length) {
        final row = data[r];
        return List.generate(cols, (c) => c < row.length ? row[c] : '');
      }
      return List.generate(cols, (_) => '');
    });
  }

  List<double> _resizeColumnWidths(
      List<double>? existing, int cols, double totalWidth) {
    final defaultW = totalWidth / cols;
    if (existing == null) {
      return List.generate(cols, (_) => defaultW);
    }
    return List.generate(
        cols, (c) => c < existing.length ? existing[c] : defaultW);
  }
}

class _TableActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback? onTap;

  const _TableActionButton({
    required this.icon,
    required this.label,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final enabled = onTap != null;
    return InkWell(
      borderRadius: BorderRadius.circular(4),
      onTap: onTap,
      child: Container(
        height: 28,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: global.theme.surfaceColor,
          borderRadius: BorderRadius.circular(4),
          border: Border.all(
              color: global.theme.dividerBorderColor, width: 0.5),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: enabled
                ? global.theme.textColor
                : global.theme.textSecondaryColor,
            fontSize: 10,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }
}

// ============================================================
// 6. Content Section
// ============================================================

class _ContentSection extends StatefulWidget {
  final FormElement element;
  final ValueChanged<FormElement> onChanged;

  const _ContentSection({
    required this.element,
    required this.onChanged,
  });

  @override
  State<_ContentSection> createState() => _ContentSectionState();
}

class _ContentSectionState extends State<_ContentSection> {
  late TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.element.text ?? '');
  }

  @override
  void didUpdateWidget(covariant _ContentSection oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.element.id != oldWidget.element.id) {
      _controller.text = widget.element.text ?? '';
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return _buildSectionTile(
      icon: Icons.edit_note,
      title: 'Content',
      children: [
        Container(
          constraints: const BoxConstraints(minHeight: 80, maxHeight: 160),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(4),
            border: Border.all(
                color: global.theme.dividerBorderColor, width: 0.5),
          ),
          child: TextField(
            controller: _controller,
            maxLines: null,
            minLines: 3,
            style: TextStyle(color: global.theme.textColor, fontSize: 12),
            decoration: InputDecoration(
              contentPadding: const EdgeInsets.all(8),
              border: InputBorder.none,
              hintText: 'Enter text content...',
              hintStyle: TextStyle(
                  color: global.theme.textSecondaryColor, fontSize: 12),
            ),
            cursorColor: global.theme.primaryColor,
            onChanged: (v) {
              widget.onChanged(widget.element.copyWith(text: v));
            },
          ),
        ),
        const SizedBox(height: 8),
      ],
    );
  }
}
