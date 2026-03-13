import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import '../models/form_element.dart';
import '../models/form_section.dart';
import 'ruler_widget.dart';

class SectionEditor extends StatefulWidget {
  final FormSection section;
  final Function(FormSection) onSectionChanged;
  final double paperWidth;
  final double zoomLevel;
  final bool showRulers;

  const SectionEditor({
    super.key,
    required this.section,
    required this.onSectionChanged,
    required this.paperWidth,
    this.zoomLevel = 1.0,
    this.showRulers = true,
  });

  @override
  State<SectionEditor> createState() => _SectionEditorState();
}

class _SectionEditorState extends State<SectionEditor> with global.ThemeRefreshMixin {
  FormElement? selectedElement;
  final ImagePicker _picker = ImagePicker();

  void _addTextElement() {
    final newElement = FormElement(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: ElementType.text,
      x: 50,
      y: 50,
      width: 200,
      height: 40,
      text: global.language('full_form_editor_new_text'),
      textStyle: TextStyle(fontSize: 16, color: global.theme.textColor),
      textAlign: TextAlign.left,
    );

    final updatedSection = widget.section.copyWith(
      elements: [...widget.section.elements, newElement],
    );
    widget.onSectionChanged(updatedSection);
  }

  void _addRectangleElement() {
    final newElement = FormElement(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: ElementType.rectangle,
      x: 50,
      y: 50,
      width: 200,
      height: 100,
      borderColor: global.theme.textColor,
      borderWidth: 2,
    );

    final updatedSection = widget.section.copyWith(
      elements: [...widget.section.elements, newElement],
    );
    widget.onSectionChanged(updatedSection);
  }

  void _addTableElement() {
    final newElement = FormElement(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: ElementType.table,
      x: 50,
      y: 50,
      width: 400,
      height: 200,
      rows: 3,
      columns: 3,
      tableData: List.generate(
        3,
        (i) => List.generate(3, (j) => 'Cell ${i + 1},${j + 1}'),
      ),
      borderColor: global.theme.textColor,
      borderWidth: 1,
    );

    final updatedSection = widget.section.copyWith(
      elements: [...widget.section.elements, newElement],
    );
    widget.onSectionChanged(updatedSection);
  }

  Future<void> _addImageElement() async {
    final XFile? image = await _picker.pickImage(source: ImageSource.gallery);
    if (image != null) {
      final newElement = FormElement(
        id: DateTime.now().millisecondsSinceEpoch.toString(),
        type: ElementType.image,
        x: 50,
        y: 50,
        width: 200,
        height: 200,
        imagePath: image.path,
      );

      final updatedSection = widget.section.copyWith(
        elements: [...widget.section.elements, newElement],
      );
      widget.onSectionChanged(updatedSection);
    }
  }

  Future<void> _setBackgroundImage() async {
    final XFile? image = await _picker.pickImage(source: ImageSource.gallery);
    if (image != null) {
      final updatedSection = widget.section.copyWith(
        backgroundImagePath: image.path,
      );
      widget.onSectionChanged(updatedSection);
    }
  }

  void _removeBackgroundImage() {
    final updatedSection = widget.section.copyWith(
      backgroundImagePath: null,
    );
    widget.onSectionChanged(updatedSection);
  }

  void _deleteElement(FormElement element) {
    final updatedElements =
        widget.section.elements.where((e) => e.id != element.id).toList();
    final updatedSection = widget.section.copyWith(elements: updatedElements);
    widget.onSectionChanged(updatedSection);
    setState(() {
      selectedElement = null;
    });
  }

  void _updateElement(FormElement element) {
    final updatedElements = widget.section.elements.map((e) {
      return e.id == element.id ? element : e;
    }).toList();
    final updatedSection = widget.section.copyWith(elements: updatedElements);
    widget.onSectionChanged(updatedSection);
  }

  @override
  Widget build(BuildContext context) {
    // Calculate actual dimensions with zoom
    final double actualZoom = widget.zoomLevel == -1 ? 1.0 : widget.zoomLevel;
    final double canvasWidth = widget.paperWidth * actualZoom;
    final double canvasHeight = widget.section.height * actualZoom;
    final double rulerSize = widget.showRulers ? 25 : 0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Toolbar
        Container(
          padding: const EdgeInsets.all(8),
          color: global.theme.dividerBorderColor,
          child: Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              ElevatedButton.icon(
                onPressed: _addTextElement,
                icon: Icon(Icons.text_fields, size: 18),
                label: Text(global.language('add_text')),
              ),
              ElevatedButton.icon(
                onPressed: _addRectangleElement,
                icon: Icon(Icons.rectangle_outlined, size: 18),
                label: Text(global.language('add_rectangle')),
              ),
              ElevatedButton.icon(
                onPressed: _addTableElement,
                icon: Icon(Icons.table_chart, size: 18),
                label: Text(global.language('add_table')),
              ),
              ElevatedButton.icon(
                onPressed: _addImageElement,
                icon: Icon(Icons.image, size: 18),
                label: Text(global.language('add_image')),
              ),
              SizedBox(width: 20),
              ElevatedButton.icon(
                onPressed: _setBackgroundImage,
                icon: Icon(Icons.wallpaper, size: 18),
                label: Text(global.language('set_background')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.infoHighlightColor,
                ),
              ),
              if (widget.section.backgroundImagePath != null)
                ElevatedButton.icon(
                  onPressed: _removeBackgroundImage,
                  icon: Icon(Icons.close, size: 18),
                  label: Text(global.language('remove_background')),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.negativeHighlightColor,
                  ),
                ),
            ],
          ),
        ),
        // Canvas with rulers
        Expanded(
          child: Container(
            color: global.theme.dividerBorderColor, // พื้นหลัง work area
            child: widget.zoomLevel == -1
                ? _buildFullScreenCanvas()
                : SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: SingleChildScrollView(
                      scrollDirection: Axis.vertical,
                      child: _buildCanvasWithRulers(
                          canvasWidth, canvasHeight, actualZoom, rulerSize),
                    ),
                  ),
          ),
        ),
        // Properties Panel
        if (selectedElement != null)
          Container(
            padding: const EdgeInsets.all(8),
            color: global.theme.surfaceColor,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      '${global.language("properties_label")} ${_getElementTypeName(selectedElement!.type)}',
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    const Spacer(),
                    IconButton(
                      icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                      onPressed: () => _deleteElement(selectedElement!),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                _buildPropertiesPanel(selectedElement!),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildFullScreenCanvas() {
    return LayoutBuilder(
      builder: (context, constraints) {
        final availableWidth = constraints.maxWidth;
        final availableHeight = constraints.maxHeight;

        // Calculate scale to fit
        final scaleX = (availableWidth - 50) / widget.paperWidth;
        final scaleY = (availableHeight - 50) / widget.section.height;
        final scale = scaleX < scaleY ? scaleX : scaleY;

        return Center(
          child: _buildCanvasWithRulers(
            widget.paperWidth * scale,
            widget.section.height * scale,
            scale,
            widget.showRulers ? 25 : 0,
          ),
        );
      },
    );
  }

  Widget _buildCanvasWithRulers(
      double canvasWidth, double canvasHeight, double zoom, double rulerSize) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Vertical ruler
        if (widget.showRulers)
          Column(
            children: [
              SizedBox(width: rulerSize, height: rulerSize), // Corner
              VerticalRuler(height: widget.section.height, zoom: zoom),
            ],
          ),
        // Main canvas with horizontal ruler
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Horizontal ruler
            if (widget.showRulers)
              HorizontalRuler(width: widget.paperWidth, zoom: zoom),
            // Canvas
            Container(
              width: canvasWidth,
              height: canvasHeight,
              decoration: BoxDecoration(
                color: global.theme.cardColor, // White canvas background
                border: Border.all(color: global.theme.infoHighlightTextColor, width: 2),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.2),
                    blurRadius: 8,
                    offset: const Offset(2, 2),
                  ),
                ],
                image: widget.section.backgroundImagePath != null
                    ? DecorationImage(
                        image: FileImage(
                            File(widget.section.backgroundImagePath!)),
                        fit: BoxFit.cover,
                        opacity: 0.3,
                      )
                    : null,
              ),
              child: Stack(
                children: widget.section.elements.map((element) {
                  return Positioned(
                    left: element.x * zoom,
                    top: element.y * zoom,
                    child: GestureDetector(
                      onTap: () {
                        setState(() {
                          selectedElement = element;
                        });
                      },
                      onPanUpdate: (details) {
                        final updatedElement = element.copyWith(
                          x: (element.x + details.delta.dx / zoom)
                              .clamp(0, widget.paperWidth - element.width),
                          y: (element.y + details.delta.dy / zoom)
                              .clamp(0, widget.section.height - element.height),
                        );
                        _updateElement(updatedElement);
                      },
                      child: Container(
                        width: element.width * zoom,
                        height: element.height * zoom,
                        decoration: BoxDecoration(
                          color: selectedElement?.id == element.id
                              ? global.theme.infoHighlightColor
                              : element.backgroundColor,
                          border: Border.all(
                            color: selectedElement?.id == element.id
                                ? global.theme.infoHighlightTextColor
                                : element.borderColor ?? Colors.transparent,
                            width: selectedElement?.id == element.id
                                ? 2
                                : (element.borderWidth ?? 0) * zoom,
                          ),
                        ),
                        child: Transform.scale(
                          scale: zoom,
                          alignment: Alignment.topLeft,
                          child: SizedBox(
                            width: element.width,
                            height: element.height,
                            child: _buildElementContent(element),
                          ),
                        ),
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildElementContent(FormElement element) {
    switch (element.type) {
      case ElementType.text:
        return Text(
          element.text ?? '',
          style: element.textStyle,
          textAlign: element.textAlign,
        );
      case ElementType.image:
        if (element.imagePath != null) {
          return Image.file(
            File(element.imagePath!),
            fit: BoxFit.contain,
          );
        }
        return Icon(Icons.image);
      case ElementType.table:
        return _buildTable(element);
      case ElementType.rectangle:
      case ElementType.line:
      case ElementType.separator:
        return Container();
      case ElementType.dataField:
      case ElementType.barcode:
      case ElementType.qrCode:
      case ElementType.pageNumber:
      case ElementType.dateTime:
      case ElementType.signature:
        return Text(element.displayName,
            style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor));
    }
  }

  Widget _buildTable(FormElement element) {
    if (element.tableData == null) return Container();
    final aligns = element.columnAligns;

    return Table(
      border: TableBorder.all(
        color: element.borderColor ?? global.theme.textColor,
        width: element.borderWidth ?? 1,
      ),
      children: element.tableData!.asMap().entries.map((entry) {
        final isHeader = element.tableHasHeader && entry.key == 0;
        return TableRow(
          children: entry.value.asMap().entries.map((cellEntry) {
            final colIdx = cellEntry.key;
            final cell = cellEntry.value;
            final align = aligns != null && colIdx < aligns.length
                ? TextAlign.values[aligns[colIdx].clamp(0, 5)]
                : (isHeader ? TextAlign.center : TextAlign.left);
            return Padding(
              padding: const EdgeInsets.all(4),
              child: Text(
                cell,
                textAlign: align,
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: isHeader ? FontWeight.bold : FontWeight.normal,
                ),
              ),
            );
          }).toList(),
        );
      }).toList(),
    );
  }

  String _getElementTypeName(ElementType type) {
    switch (type) {
      case ElementType.text:
        return global.language('bill_text');
      case ElementType.image:
        return global.language('bill_image');
      case ElementType.rectangle:
        return global.language('full_form_editor_frame');
      case ElementType.table:
        return global.language('full_form_editor_table');
      case ElementType.line:
        return global.language('full_form_editor_line');
      case ElementType.dataField:
      case ElementType.barcode:
      case ElementType.qrCode:
      case ElementType.separator:
      case ElementType.pageNumber:
      case ElementType.dateTime:
      case ElementType.signature:
        return type.label;
    }
  }

  Widget _buildPropertiesPanel(FormElement element) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          if (element.type == ElementType.text) ...[
            SizedBox(
              width: 200,
              child: TextField(
                decoration: InputDecoration(
                  labelText: global.language('bill_text'),
                  isDense: true,
                ),
                controller: TextEditingController(text: element.text),
                onChanged: (value) {
                  _updateElement(element.copyWith(text: value));
                },
              ),
            ),
            SizedBox(width: 8),
            SizedBox(
              width: 100,
              child: TextField(
                decoration: InputDecoration(
                  labelText: global.language('font_size'),
                  isDense: true,
                ),
                keyboardType: TextInputType.number,
                controller: TextEditingController(
                  text: element.fontSize?.toString() ?? '16',
                ),
                onChanged: (value) {
                  final fontSize = double.tryParse(value) ?? 16;
                  _updateElement(element.copyWith(
                    fontSize: fontSize,
                  ));
                },
              ),
            ),
          ],
          SizedBox(width: 8),
          SizedBox(
            width: 100,
            child: TextField(
              decoration: InputDecoration(
                labelText: global.language('full_form_editor_width'),
                isDense: true,
              ),
              keyboardType: TextInputType.number,
              controller: TextEditingController(text: element.width.toString()),
              onChanged: (value) {
                final width = double.tryParse(value) ?? element.width;
                _updateElement(element.copyWith(width: width));
              },
            ),
          ),
          SizedBox(width: 8),
          SizedBox(
            width: 100,
            child: TextField(
              decoration: InputDecoration(
                labelText: global.language('full_form_editor_height'),
                isDense: true,
              ),
              keyboardType: TextInputType.number,
              controller:
                  TextEditingController(text: element.height.toString()),
              onChanged: (value) {
                final height = double.tryParse(value) ?? element.height;
                _updateElement(element.copyWith(height: height));
              },
            ),
          ),
        ],
      ),
    );
  }
}
