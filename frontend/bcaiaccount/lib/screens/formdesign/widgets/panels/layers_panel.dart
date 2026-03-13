import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../../models/form_element.dart';
import '../../models/form_section.dart';
import '../../models/form_template.dart';

class LayersPanel extends StatefulWidget {
  final FormTemplate template;
  final FormElement? selectedElement;
  final Function(FormElement element, SectionType section) onElementSelected;
  final Function(FormElement element, SectionType section)
      onElementVisibilityToggled;
  final Function(FormElement element, SectionType section)
      onElementLockToggled;
  final Function(FormTemplate template) onReorder;

  const LayersPanel({
    super.key,
    required this.template,
    this.selectedElement,
    required this.onElementSelected,
    required this.onElementVisibilityToggled,
    required this.onElementLockToggled,
    required this.onReorder,
  });

  @override
  State<LayersPanel> createState() => _LayersPanelState();
}

class _LayersPanelState extends State<LayersPanel>
    with global.ThemeRefreshMixin {
  final Map<SectionType, bool> _expanded = {
    SectionType.header: true,
    SectionType.detail: true,
    SectionType.footer: true,
  };

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 240,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(
          left: BorderSide(color: global.theme.dividerBorderColor, width: 1),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _buildHeader(),
          Expanded(child: _buildSectionList()),
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
        global.language('fd_layers'),
        style: TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w600,
          color: global.theme.textColor,
        ),
      ),
    );
  }

  Widget _buildSectionList() {
    return ListView(
      padding: EdgeInsets.zero,
      children: [
        _buildSection(SectionType.header, widget.template.header),
        _buildSection(SectionType.detail, widget.template.detail),
        _buildSection(SectionType.footer, widget.template.footer),
      ],
    );
  }

  Widget _buildSection(SectionType sectionType, FormSection section) {
    final isExpanded = _expanded[sectionType] ?? true;
    // ใช้ resolved elements ถ้า section เป็น row-based
    final elements = section.hasRows
        ? section.resolveElements(widget.template.effectiveWidth)
        : section.elements;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        InkWell(
          onTap: () {
            setState(() => _expanded[sectionType] = !isExpanded);
          },
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              border: Border(
                bottom: BorderSide(
                    color: global.theme.dividerBorderColor, width: 0.5),
              ),
            ),
            child: Row(
              children: [
                Icon(
                  isExpanded ? Icons.expand_more : Icons.chevron_right,
                  size: 16,
                  color: global.theme.iconSecondaryColor,
                ),
                const SizedBox(width: 4),
                Icon(
                  _sectionIcon(sectionType),
                  size: 14,
                  color: global.theme.iconColor,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    _sectionLabel(sectionType),
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: global.theme.textColor,
                    ),
                  ),
                ),
                Text(
                  '${elements.length}',
                  style: TextStyle(
                    fontSize: 10,
                    color: global.theme.textSecondaryColor,
                  ),
                ),
              ],
            ),
          ),
        ),
        if (isExpanded)
          _buildElementList(sectionType, elements,
              isRowBased: section.hasRows),
      ],
    );
  }

  Widget _buildElementList(
      SectionType sectionType, List<FormElement> elements,
      {bool isRowBased = false}) {
    if (elements.isEmpty) {
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Text(
          global.language('fd_no_elements'),
          style: TextStyle(
            fontSize: 11,
            fontStyle: FontStyle.italic,
            color: global.theme.textSecondaryColor,
          ),
        ),
      );
    }

    // Row-based sections: read-only list (ลำดับมาจาก rows ไม่ reorder)
    if (isRowBased) {
      return ListView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemCount: elements.length,
        itemBuilder: (context, index) {
          final element = elements[index];
          return _buildElementRow(
            key: ValueKey('${sectionType.name}_${element.id}'),
            element: element,
            sectionType: sectionType,
            index: index,
            showDragHandle: false,
          );
        },
      );
    }

    return ReorderableListView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      buildDefaultDragHandles: false,
      itemCount: elements.length,
      onReorder: (oldIndex, newIndex) =>
          _handleReorder(sectionType, elements, oldIndex, newIndex),
      itemBuilder: (context, index) {
        final element = elements[index];
        return _buildElementRow(
          key: ValueKey('${sectionType.name}_${element.id}'),
          element: element,
          sectionType: sectionType,
          index: index,
        );
      },
    );
  }

  Widget _buildElementRow({
    required Key key,
    required FormElement element,
    required SectionType sectionType,
    required int index,
    bool showDragHandle = true,
  }) {
    final isSelected = widget.selectedElement?.id == element.id;

    return InkWell(
      key: key,
      onTap: () => widget.onElementSelected(element, sectionType),
      child: Container(
        height: 30,
        padding: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          color: isSelected
              ? global.theme.rowSelectedColor
              : global.theme.cardColor,
          border: Border(
            bottom: BorderSide(
                color: global.theme.dividerBorderColor, width: 0.5),
          ),
        ),
        child: Row(
          children: [
            if (showDragHandle)
              ReorderableDragStartListener(
                index: index,
                child: Icon(Icons.drag_indicator,
                    size: 14, color: global.theme.iconSecondaryColor),
              )
            else
              const SizedBox(width: 14),
            const SizedBox(width: 4),
            Icon(
              element.type.icon,
              size: 14,
              color: isSelected
                  ? global.theme.primaryColor
                  : global.theme.iconColor,
            ),
            const SizedBox(width: 6),
            Expanded(
              child: Text(
                element.displayName,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 11,
                  color: isSelected
                      ? global.theme.primaryColor
                      : global.theme.textColor,
                ),
              ),
            ),
            _buildVisibilityButton(element, sectionType),
            _buildLockButton(element, sectionType),
          ],
        ),
      ),
    );
  }

  Widget _buildVisibilityButton(
      FormElement element, SectionType sectionType) {
    return SizedBox(
      width: 24,
      height: 24,
      child: IconButton(
        padding: EdgeInsets.zero,
        iconSize: 14,
        onPressed: () =>
            widget.onElementVisibilityToggled(element, sectionType),
        icon: Icon(
          element.isVisible ? Icons.visibility : Icons.visibility_off,
          color: element.isVisible
              ? global.theme.iconColor
              : global.theme.iconSecondaryColor,
        ),
        tooltip: element.isVisible ? global.language('fd_hide') : global.language('fd_show'),
      ),
    );
  }

  Widget _buildLockButton(FormElement element, SectionType sectionType) {
    return SizedBox(
      width: 24,
      height: 24,
      child: IconButton(
        padding: EdgeInsets.zero,
        iconSize: 14,
        onPressed: () => widget.onElementLockToggled(element, sectionType),
        icon: Icon(
          element.isLocked ? Icons.lock : Icons.lock_open,
          color: element.isLocked
              ? global.theme.warningHighlightTextColor
              : global.theme.iconSecondaryColor,
        ),
        tooltip: element.isLocked ? global.language('fd_unlock') : global.language('fd_lock'),
      ),
    );
  }

  void _handleReorder(SectionType sectionType, List<FormElement> elements,
      int oldIndex, int newIndex) {
    if (newIndex > oldIndex) newIndex--;
    final reordered = List<FormElement>.from(elements);
    final item = reordered.removeAt(oldIndex);
    reordered.insert(newIndex, item);

    // Update zIndex based on list order
    for (var i = 0; i < reordered.length; i++) {
      reordered[i].zIndex = i;
    }

    FormTemplate updated;
    switch (sectionType) {
      case SectionType.header:
        updated = widget.template.copyWith(
          header:
              widget.template.header.copyWith(elements: reordered),
        );
      case SectionType.detail:
        updated = widget.template.copyWith(
          detail:
              widget.template.detail.copyWith(elements: reordered),
        );
      case SectionType.footer:
        updated = widget.template.copyWith(
          footer:
              widget.template.footer.copyWith(elements: reordered),
        );
    }
    widget.onReorder(updated);
  }

  IconData _sectionIcon(SectionType s) {
    switch (s) {
      case SectionType.header:
        return Icons.vertical_align_top;
      case SectionType.detail:
        return Icons.view_agenda;
      case SectionType.footer:
        return Icons.vertical_align_bottom;
    }
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
