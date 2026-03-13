import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../../models/form_element.dart';
import '../../models/form_section.dart';

class ElementPalettePanel extends StatefulWidget {
  const ElementPalettePanel({
    super.key,
  });

  @override
  State<ElementPalettePanel> createState() => _ElementPalettePanelState();
}

class _ElementPalettePanelState extends State<ElementPalettePanel>
    with global.ThemeRefreshMixin {
  String _searchText = '';
  SectionType _selectedSection = SectionType.header;

  static const _categories = <_CategoryDef>[
    _CategoryDef(
      key: 'static',
      label: 'Static',
      icon: Icons.widgets,
      types: [
        ElementType.text,
        ElementType.image,
        ElementType.line,
        ElementType.rectangle,
        ElementType.separator,
      ],
    ),
    _CategoryDef(
      key: 'data',
      label: 'Data',
      icon: Icons.data_object,
      types: [ElementType.dataField],
    ),
    _CategoryDef(
      key: 'table',
      label: 'Table',
      icon: Icons.table_chart,
      types: [ElementType.table],
    ),
    _CategoryDef(
      key: 'special',
      label: 'Special',
      icon: Icons.star,
      types: [
        ElementType.barcode,
        ElementType.qrCode,
        ElementType.pageNumber,
        ElementType.dateTime,
        ElementType.signature,
      ],
    ),
  ];

  List<ElementType> _filterTypes(List<ElementType> types) {
    if (_searchText.isEmpty) return types;
    final q = _searchText.toLowerCase();
    return types.where((t) => t.label.toLowerCase().contains(q)).toList();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 240,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(
          right: BorderSide(color: global.theme.dividerBorderColor, width: 1),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _buildHeader(),
          _buildSectionSelector(),
          _buildSearchField(),
          Expanded(child: _buildCategoryList()),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border(
          bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1),
        ),
      ),
      child: Text(
        global.language('fd_elements'),
        style: TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w600,
          color: global.theme.textColor,
        ),
      ),
    );
  }

  Widget _buildSectionSelector() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 8, 8, 0),
      child: Container(
        height: 32,
        padding: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          color: global.theme.inputFillColor,
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: global.theme.formBorderColor),
        ),
        child: DropdownButtonHideUnderline(
          child: DropdownButton<SectionType>(
            value: _selectedSection,
            isExpanded: true,
            isDense: true,
            icon: Icon(Icons.arrow_drop_down,
                size: 18, color: global.theme.iconSecondaryColor),
            dropdownColor: global.theme.dialogColor,
            style: TextStyle(fontSize: 12, color: global.theme.textColor),
            items: SectionType.values.map((s) {
              return DropdownMenuItem(
                value: s,
                child: Text(_sectionLabel(s)),
              );
            }).toList(),
            onChanged: (v) {
              if (v != null) setState(() => _selectedSection = v);
            },
          ),
        ),
      ),
    );
  }

  Widget _buildSearchField() {
    return Padding(
      padding: const EdgeInsets.all(8),
      child: SizedBox(
        height: 32,
        child: TextField(
          style: TextStyle(fontSize: 12, color: global.theme.formTextColor),
          decoration: InputDecoration(
            hintText: global.language('fd_search_elements'),
            hintStyle:
                TextStyle(fontSize: 12, color: global.theme.formHintColor),
            prefixIcon: Icon(Icons.search,
                size: 16, color: global.theme.iconSecondaryColor),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 8, vertical: 0),
            filled: true,
            fillColor: global.theme.inputFillColor,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(6),
              borderSide: BorderSide(color: global.theme.formBorderColor),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(6),
              borderSide: BorderSide(color: global.theme.formBorderColor),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(6),
              borderSide: BorderSide(color: global.theme.formFocusBorderColor),
            ),
          ),
          onChanged: (v) => setState(() => _searchText = v),
        ),
      ),
    );
  }

  Widget _buildCategoryList() {
    return ListView(
      padding: EdgeInsets.zero,
      children: _categories.map((cat) {
        final filtered = _filterTypes(cat.types);
        if (filtered.isEmpty) return const SizedBox.shrink();
        return _buildCategory(cat, filtered);
      }).toList(),
    );
  }

  Widget _buildCategory(_CategoryDef cat, List<ElementType> types) {
    return ExpansionTile(
      initiallyExpanded: true,
      dense: true,
      tilePadding: const EdgeInsets.symmetric(horizontal: 12),
      childrenPadding: const EdgeInsets.only(bottom: 4),
      iconColor: global.theme.iconSecondaryColor,
      collapsedIconColor: global.theme.iconSecondaryColor,
      title: Row(
        children: [
          Icon(cat.icon, size: 16, color: global.theme.iconColor),
          const SizedBox(width: 6),
          Text(
            global.language('fd_${cat.key}'),
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: global.theme.textColor,
            ),
          ),
        ],
      ),
      children: types.map((t) => _buildDraggableItem(t)).toList(),
    );
  }

  Widget _buildDraggableItem(ElementType type) {
    final data = <String, dynamic>{
      'type': type,
      'sectionType': _selectedSection,
    };

    return Draggable<Map<String, dynamic>>(
      data: data,
      feedback: Material(
        elevation: 4,
        borderRadius: BorderRadius.circular(6),
        color: global.theme.primaryColor.withAlpha(230),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(type.icon, size: 14, color: global.theme.onPrimaryColor),
              const SizedBox(width: 6),
              Text(
                type.label,
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.onPrimaryColor,
                ),
              ),
            ],
          ),
        ),
      ),
      childWhenDragging: Opacity(
        opacity: 0.4,
        child: _buildItemRow(type),
      ),
      child: _buildItemRow(type),
    );
  }

  Widget _buildItemRow(ElementType type) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 1),
      child: Container(
        height: 30,
        padding: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(4),
        ),
        child: Row(
          children: [
            Icon(type.icon, size: 14, color: global.theme.iconColor),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                type.label,
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.textColor,
                ),
              ),
            ),
            Icon(Icons.drag_indicator,
                size: 14, color: global.theme.iconSecondaryColor),
          ],
        ),
      ),
    );
  }

  String _sectionLabel(SectionType s) {
    switch (s) {
      case SectionType.header:
        return global.language('fd_header');
      case SectionType.detail:
        return global.language('fd_detail');
      case SectionType.footer:
        return global.language('fd_footer');
    }
  }
}

class _CategoryDef {
  final String key;
  final String label;
  final IconData icon;
  final List<ElementType> types;

  const _CategoryDef({
    required this.key,
    required this.label,
    required this.icon,
    required this.types,
  });
}
