import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import '../models/form_element.dart';
import '../models/form_section.dart';
import '../models/form_template.dart';
import '../models/paper_size.dart';
import 'ruler_widget.dart';
import 'package:smlaicloud/global.dart' as global;

class FullFormEditor extends StatefulWidget {
  final FormTemplate template;
  final Function(FormTemplate) onTemplateChanged;
  final double zoomLevel;
  final bool showRulers;
  final Function(FormElement?)? onElementSelected;

  const FullFormEditor({
    super.key,
    required this.template,
    required this.onTemplateChanged,
    this.zoomLevel = 1.0,
    this.showRulers = true,
    this.onElementSelected,
  });

  @override
  State<FullFormEditor> createState() => _FullFormEditorState();
}

class _FullFormEditorState extends State<FullFormEditor> {
  FormElement? selectedElement;
  SectionType? selectedSection;
  final ImagePicker _picker = ImagePicker();

  void _addTextElement(SectionType sectionType) {
    final newElement = FormElement(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: ElementType.text,
      x: 50,
      y: 50,
      width: 200,
      height: 40,
      text: global.language('full_form_editor_new_text'),
      textStyle: const TextStyle(fontSize: 16, color: Colors.black),
      textAlign: TextAlign.left,
    );

    _addElementToSection(sectionType, newElement);
  }

  void _addRectangleElement(SectionType sectionType) {
    final newElement = FormElement(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: ElementType.rectangle,
      x: 50,
      y: 50,
      width: 200,
      height: 100,
      borderColor: Colors.black,
      borderWidth: 2,
    );

    _addElementToSection(sectionType, newElement);
  }

  void _addTableElement(SectionType sectionType) {
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
      borderColor: Colors.black,
      borderWidth: 1,
    );

    _addElementToSection(sectionType, newElement);
  }

  Future<void> _addImageElement(SectionType sectionType) async {
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

      _addElementToSection(sectionType, newElement);
    }
  }

  void _addElementToSection(SectionType sectionType, FormElement element) {
    FormTemplate updated;
    switch (sectionType) {
      case SectionType.header:
        updated = widget.template.copyWith(
          header: widget.template.header.copyWith(
            elements: List.from(widget.template.header.elements)..add(element),
          ),
        );
        break;
      case SectionType.detail:
        updated = widget.template.copyWith(
          detail: widget.template.detail.copyWith(
            elements: List.from(widget.template.detail.elements)..add(element),
          ),
        );
        break;
      case SectionType.footer:
        updated = widget.template.copyWith(
          footer: widget.template.footer.copyWith(
            elements: List.from(widget.template.footer.elements)..add(element),
          ),
        );
        break;
    }
    widget.onTemplateChanged(updated);
  }

  void _updateElement(SectionType sectionType, FormElement element) {
    FormSection section;
    switch (sectionType) {
      case SectionType.header:
        section = widget.template.header;
        break;
      case SectionType.detail:
        section = widget.template.detail;
        break;
      case SectionType.footer:
        section = widget.template.footer;
        break;
    }

    final updatedElements = section.elements.map((e) {
      return e.id == element.id ? element : e;
    }).toList();

    final updatedSection = section.copyWith(elements: updatedElements);

    FormTemplate updated;
    switch (sectionType) {
      case SectionType.header:
        updated = widget.template.copyWith(header: updatedSection);
        break;
      case SectionType.detail:
        updated = widget.template.copyWith(detail: updatedSection);
        break;
      case SectionType.footer:
        updated = widget.template.copyWith(footer: updatedSection);
        break;
    }

    widget.onTemplateChanged(updated);
  }

  void _deleteElement(SectionType sectionType, FormElement element) {
    FormSection section;
    switch (sectionType) {
      case SectionType.header:
        section = widget.template.header;
        break;
      case SectionType.detail:
        section = widget.template.detail;
        break;
      case SectionType.footer:
        section = widget.template.footer;
        break;
    }

    final updatedElements =
        section.elements.where((e) => e.id != element.id).toList();
    final updatedSection = section.copyWith(elements: updatedElements);

    FormTemplate updated;
    switch (sectionType) {
      case SectionType.header:
        updated = widget.template.copyWith(header: updatedSection);
        break;
      case SectionType.detail:
        updated = widget.template.copyWith(detail: updatedSection);
        break;
      case SectionType.footer:
        updated = widget.template.copyWith(footer: updatedSection);
        break;
    }

    widget.onTemplateChanged(updated);
    setState(() {
      selectedElement = null;
      selectedSection = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    final double actualZoom = widget.zoomLevel == -1 ? 1.0 : widget.zoomLevel;
    final double canvasWidth = widget.template.paperSize.width * actualZoom;
    final double canvasHeight = widget.template.paperSize.height * actualZoom;
    final double rulerSize = widget.showRulers ? 25 : 0;

    return Column(
      children: <Widget>[
        _buildToolbar(),
        Expanded(
          child: Container(
            color: Colors.grey[400],
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
        if (selectedElement != null && selectedSection != null)
          _buildPropertiesPanel(),
      ],
    );
  }

  Widget _buildToolbar() {
    return Container(
      padding: const EdgeInsets.all(8),
      color: Colors.grey[200],
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: <Widget>[
            Text(global.language('full_form_editor_add_in'),
                style: const TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(width: 8),
            DropdownButton<SectionType>(
              value: selectedSection ?? SectionType.detail,
              items: const <DropdownMenuItem<SectionType>>[
                DropdownMenuItem(
                  value: SectionType.header,
                  child: Text('Header'),
                ),
                DropdownMenuItem(
                  value: SectionType.detail,
                  child: Text('Detail'),
                ),
                DropdownMenuItem(
                  value: SectionType.footer,
                  child: Text('Footer'),
                ),
              ],
              onChanged: (value) {
                if (value != null) {
                  setState(() {
                    selectedSection = value;
                  });
                }
              },
            ),
            SizedBox(width: 16),
            ElevatedButton.icon(
              onPressed: () =>
                  _addTextElement(selectedSection ?? SectionType.detail),
              icon: Icon(Icons.text_fields, size: 18),
              label: Text(global.language('full_form_editor_text')),
            ),
            SizedBox(width: 8),
            ElevatedButton.icon(
              onPressed: () =>
                  _addRectangleElement(selectedSection ?? SectionType.detail),
              icon: Icon(Icons.rectangle_outlined, size: 18),
              label: Text(global.language('full_form_editor_frame')),
            ),
            SizedBox(width: 8),
            ElevatedButton.icon(
              onPressed: () =>
                  _addTableElement(selectedSection ?? SectionType.detail),
              icon: Icon(Icons.table_chart, size: 18),
              label: Text(global.language('full_form_editor_table')),
            ),
            SizedBox(width: 8),
            ElevatedButton.icon(
              onPressed: () =>
                  _addImageElement(selectedSection ?? SectionType.detail),
              icon: Icon(Icons.image, size: 18),
              label: Text(global.language('full_form_editor_image')),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFullScreenCanvas() {
    return LayoutBuilder(
      builder: (context, constraints) {
        final availableWidth = constraints.maxWidth;
        final availableHeight = constraints.maxHeight;

        final scaleX = (availableWidth - 50) / widget.template.paperSize.width;
        final scaleY =
            (availableHeight - 50) / widget.template.paperSize.height;
        final scale = scaleX < scaleY ? scaleX : scaleY;

        return Center(
          child: _buildCanvasWithRulers(
            widget.template.paperSize.width * scale,
            widget.template.paperSize.height * scale,
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
      children: <Widget>[
        if (widget.showRulers)
          Column(
            children: <Widget>[
              SizedBox(width: rulerSize, height: rulerSize),
              VerticalRuler(
                  height: widget.template.paperSize.height, zoom: zoom),
            ],
          ),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            if (widget.showRulers)
              HorizontalRuler(
                  width: widget.template.paperSize.width, zoom: zoom),
            _buildFullForm(canvasWidth, canvasHeight, zoom),
          ],
        ),
      ],
    );
  }

  Widget _buildFullForm(double canvasWidth, double canvasHeight, double zoom) {
    return Container(
      width: canvasWidth,
      height: canvasHeight,
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: Colors.blue, width: 2),
        boxShadow: <BoxShadow>[
          BoxShadow(
            color: Colors.black.withOpacity(0.2),
            blurRadius: 8,
            offset: const Offset(2, 2),
          ),
        ],
        image: widget.template.globalBackgroundImagePath != null
            ? DecorationImage(
                image:
                    FileImage(File(widget.template.globalBackgroundImagePath!)),
                fit: BoxFit.cover,
                opacity: 0.1,
              )
            : null,
      ),
      child: Stack(
        children: <Widget>[
          _buildSection(widget.template.header, 0, SectionType.header, zoom),
          _buildSection(widget.template.detail, widget.template.header.height,
              SectionType.detail, zoom),
          _buildSection(
              widget.template.footer,
              widget.template.header.height + widget.template.detail.height,
              SectionType.footer,
              zoom),
        ],
      ),
    );
  }

  Widget _buildSection(FormSection section, double offsetY,
      SectionType sectionType, double zoom) {
    return Positioned(
      left: 0,
      top: offsetY * zoom,
      width: widget.template.paperSize.width * zoom,
      height: section.height * zoom,
      child: Container(
        decoration: BoxDecoration(
          image: section.backgroundImagePath != null
              ? DecorationImage(
                  image: FileImage(File(section.backgroundImagePath!)),
                  fit: BoxFit.cover,
                  opacity: 0.3,
                )
              : null,
        ),
        child: Stack(
          children: section.elements.map((element) {
            return Positioned(
              left: element.x * zoom,
              top: element.y * zoom,
              child: GestureDetector(
                onTap: () {
                  setState(() {
                    selectedElement = element;
                    selectedSection = sectionType;
                  });
                  widget.onElementSelected?.call(element);
                },
                onPanUpdate: (details) {
                  final newX = element.x + details.delta.dx / zoom;
                  final newY = element.y + details.delta.dy / zoom;
                  final updatedElement = element.copyWith(
                    x: newX.clamp(
                        0.0, widget.template.paperSize.width - element.width),
                    y: newY.clamp(0.0, section.height - element.height),
                  );
                  _updateElement(sectionType, updatedElement);
                },
                child: Container(
                  width: element.width * zoom,
                  height: element.height * zoom,
                  decoration: BoxDecoration(
                    color: selectedElement?.id == element.id
                        ? Colors.blue.withOpacity(0.1)
                        : element.backgroundColor,
                    border: Border.all(
                      color: selectedElement?.id == element.id
                          ? Colors.blue
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
        return const Icon(Icons.image);
      case ElementType.table:
        return _buildTable(element);
      case ElementType.rectangle:
      case ElementType.line:
        return Container();
    }
  }

  Widget _buildTable(FormElement element) {
    if (element.tableData == null) return Container();

    return Table(
      border: TableBorder.all(
        color: element.borderColor ?? Colors.black,
        width: element.borderWidth ?? 1,
      ),
      children: element.tableData!.map((row) {
        return TableRow(
          children: row.map((cell) {
            return Padding(
              padding: const EdgeInsets.all(4),
              child: Text(
                cell,
                style: const TextStyle(fontSize: 12),
              ),
            );
          }).toList(),
        );
      }).toList(),
    );
  }

  Widget _buildPropertiesPanel() {
    return Container(
      padding: const EdgeInsets.all(8),
      color: Colors.grey[100],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Text(
                '${global.language('full_form_editor_properties')}: ${_getElementTypeName(selectedElement!.type)}',
                style: const TextStyle(fontWeight: FontWeight.bold),
              ),
              const Spacer(),
              IconButton(
                icon: const Icon(Icons.delete, color: Colors.red),
                onPressed: () =>
                    _deleteElement(selectedSection!, selectedElement!),
              ),
            ],
          ),
          SizedBox(height: 8),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: <Widget>[
                if (selectedElement!.type == ElementType.text) ...<Widget>[
                  SizedBox(
                    width: 200,
                    child: TextField(
                      decoration: InputDecoration(
                        labelText: global.language('full_form_editor_text'),
                        isDense: true,
                      ),
                      controller:
                          TextEditingController(text: selectedElement!.text),
                      onChanged: (value) {
                        _updateElement(
                          selectedSection!,
                          selectedElement!.copyWith(text: value),
                        );
                      },
                    ),
                  ),
                  SizedBox(width: 8),
                  SizedBox(
                    width: 100,
                    child: TextField(
                      decoration: InputDecoration(
                        labelText: global.language('full_form_editor_font_size'),
                        isDense: true,
                      ),
                      keyboardType: TextInputType.number,
                      controller: TextEditingController(
                        text:
                            selectedElement!.textStyle?.fontSize?.toString() ??
                                '16',
                      ),
                      onChanged: (value) {
                        final fontSize = double.tryParse(value) ?? 16;
                        _updateElement(
                          selectedSection!,
                          selectedElement!.copyWith(
                            textStyle: selectedElement!.textStyle
                                    ?.copyWith(fontSize: fontSize) ??
                                TextStyle(fontSize: fontSize),
                          ),
                        );
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
                    controller: TextEditingController(
                        text: selectedElement!.width.toString()),
                    onChanged: (value) {
                      final width =
                          double.tryParse(value) ?? selectedElement!.width;
                      _updateElement(
                        selectedSection!,
                        selectedElement!.copyWith(width: width),
                      );
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
                    controller: TextEditingController(
                        text: selectedElement!.height.toString()),
                    onChanged: (value) {
                      final height =
                          double.tryParse(value) ?? selectedElement!.height;
                      _updateElement(
                        selectedSection!,
                        selectedElement!.copyWith(height: height),
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _getElementTypeName(ElementType type) {
    switch (type) {
      case ElementType.text:
        return global.language('full_form_editor_text');
      case ElementType.image:
        return global.language('full_form_editor_image');
      case ElementType.rectangle:
        return global.language('full_form_editor_frame');
      case ElementType.table:
        return global.language('full_form_editor_table');
      case ElementType.line:
        return global.language('full_form_editor_line');
    }
  }
}
