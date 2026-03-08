import 'dart:io';
import 'package:flutter/material.dart';
import '../models/form_element.dart';
import '../models/form_section.dart';
import '../models/form_template.dart';
import '../models/paper_size.dart';

class FormPreview extends StatelessWidget {
  final FormTemplate template;
  final double scale;

  const FormPreview({
    super.key,
    required this.template,
    this.scale = 1.0,
  });

  @override
  Widget build(BuildContext context) {
    final paperWidth = template.paperSize.width * scale;
    final paperHeight = template.paperSize.height * scale;

    return SingleChildScrollView(
      child: Container(
        width: paperWidth,
        height: paperHeight,
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(color: Colors.grey),
          boxShadow: [
            BoxShadow(
              color: Colors.grey.withOpacity(0.5),
              spreadRadius: 2,
              blurRadius: 5,
              offset: const Offset(0, 3),
            ),
          ],
          image: template.globalBackgroundImagePath != null
              ? DecorationImage(
                  image: FileImage(File(template.globalBackgroundImagePath!)),
                  fit: BoxFit.cover,
                  opacity: 0.1,
                )
              : null,
        ),
        child: Column(
          children: [
            _buildSection(template.header, paperWidth),
            _buildSection(template.detail, paperWidth),
            _buildSection(template.footer, paperWidth),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(FormSection section, double paperWidth) {
    return Container(
      width: paperWidth,
      height: section.height * scale,
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
            left: element.x * scale,
            top: element.y * scale,
            child: SizedBox(
              width: element.width * scale,
              height: element.height * scale,
              child: _buildElement(element),
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildElement(FormElement element) {
    switch (element.type) {
      case ElementType.text:
        return Container(
          decoration: BoxDecoration(
            color: element.backgroundColor,
            border: element.borderColor != null
                ? Border.all(
                    color: element.borderColor!,
                    width: (element.borderWidth ?? 1) * scale,
                  )
                : null,
          ),
          child: Text(
            element.text ?? '',
            style: element.textStyle?.copyWith(
              fontSize: (element.textStyle?.fontSize ?? 16) * scale,
            ),
            textAlign: element.textAlign ?? TextAlign.left,
          ),
        );
      case ElementType.image:
        if (element.imagePath != null) {
          return Image.file(
            File(element.imagePath!),
            fit: BoxFit.contain,
          );
        }
        return const Icon(Icons.image);
      case ElementType.rectangle:
        return Container(
          decoration: BoxDecoration(
            color: element.backgroundColor,
            border: element.borderColor != null
                ? Border.all(
                    color: element.borderColor!,
                    width: (element.borderWidth ?? 1) * scale,
                  )
                : null,
          ),
        );
      case ElementType.table:
        return _buildTable(element);
      case ElementType.line:
        return Container(
          color: element.borderColor ?? Colors.black,
          height: (element.borderWidth ?? 1) * scale,
        );
    }
  }

  Widget _buildTable(FormElement element) {
    if (element.tableData == null) return Container();

    return Table(
      border: TableBorder.all(
        color: element.borderColor ?? Colors.black,
        width: (element.borderWidth ?? 1) * scale,
      ),
      children: element.tableData!.map((row) {
        return TableRow(
          children: row.map((cell) {
            return Padding(
              padding: EdgeInsets.all(4 * scale),
              child: Text(
                cell,
                style: TextStyle(fontSize: 12 * scale),
              ),
            );
          }).toList(),
        );
      }).toList(),
    );
  }
}
