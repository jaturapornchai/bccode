import 'dart:io';
import 'package:flutter/material.dart';
import '../models/form_element.dart';
import '../models/form_section.dart';
import '../models/form_template.dart';
import '../models/data_binding.dart';
import 'package:smlaicloud/global.dart' as global;

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
    final paperWidth = template.effectiveWidth * scale;
    final paperHeight = template.effectiveHeight * scale;
    final sectionW = template.effectiveWidth;

    final headerH = template.header.hasRows
        ? template.header.resolveHeight(sectionW)
        : template.header.height;
    final footerH = template.footer.hasRows
        ? template.footer.resolveHeight(sectionW)
        : template.footer.height;
    // Footer anchored at bottom, detail fills remaining space
    final footerOffsetY = template.effectiveHeight - footerH;

    return SingleChildScrollView(
      child: Container(
        width: paperWidth,
        height: paperHeight,
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          border: Border.all(color: global.theme.dividerBorderColor),
          boxShadow: [
            BoxShadow(
              color: global.theme.dividerBorderColor.withValues(alpha: 0.5),
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
        child: Stack(
          children: [
            // Header at top
            Positioned(
              left: 0, top: 0,
              child: _buildSection(template.header, paperWidth, headerH),
            ),
            // Detail after header
            Positioned(
              left: 0, top: headerH * scale,
              child: _buildSection(template.detail, paperWidth,
                  footerOffsetY - headerH),
            ),
            // Footer anchored at bottom
            Positioned(
              left: 0, top: footerOffsetY * scale,
              child: _buildSection(template.footer, paperWidth, footerH),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(FormSection section, double paperWidth,
      double sectionH) {
    // ใช้ row-based layout ถ้ามี rows (auto positioning)
    final sectionW = template.effectiveWidth;
    final elements = section.hasRows
        ? section.resolveElements(sectionW)
        : section.elements;

    return Container(
      width: paperWidth,
      height: sectionH * scale,
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
        children: elements.map((element) {
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
            style: element.textStyle.copyWith(
              fontSize: (element.textStyle.fontSize ?? 16) * scale,
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
        return Center(
          child: Icon(Icons.image, size: 24 * scale, color: global.theme.iconSecondaryColor),
        );
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
      case ElementType.separator:
        return Container(
          color: element.borderColor ?? global.theme.textColor,
          height: (element.borderWidth ?? 1) * scale,
        );
      case ElementType.dataField:
        final displayText = DataBindingRegistry.getSampleValue(element.dataBindingKey ?? '');
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
            displayText,
            style: element.textStyle.copyWith(
              fontSize: (element.fontSize ?? 14) * scale,
            ),
            textAlign: element.textAlign ?? TextAlign.left,
          ),
        );
      case ElementType.barcode:
        return Center(child: Icon(Icons.view_week, size: 24 * scale, color: global.theme.textColor));
      case ElementType.qrCode:
        return Center(child: Icon(Icons.qr_code, size: 32 * scale, color: global.theme.textColor));
      case ElementType.pageNumber:
        return Text(
          element.text ?? 'Page 1/1',
          style: element.textStyle.copyWith(fontSize: (element.fontSize ?? 12) * scale),
          textAlign: element.textAlign ?? TextAlign.center,
        );
      case ElementType.dateTime:
        return Text(
          element.text ?? '13/03/2569',
          style: element.textStyle.copyWith(fontSize: (element.fontSize ?? 12) * scale),
          textAlign: element.textAlign ?? TextAlign.left,
        );
      case ElementType.signature:
        return Column(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            Container(height: 1, color: global.theme.textColor),
            SizedBox(height: 2 * scale),
            Text(element.text ?? '...................................',
                style: TextStyle(fontSize: 10 * scale, color: global.theme.textSecondaryColor)),
          ],
        );
    }
  }

  Widget _buildTable(FormElement element) {
    if (element.tableData == null) return Container();
    final aligns = element.columnAligns;

    return Table(
      border: TableBorder.all(
        color: element.borderColor ?? global.theme.textColor,
        width: (element.borderWidth ?? 1) * scale,
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
              padding: EdgeInsets.all(4 * scale),
              child: Text(
                cell,
                textAlign: align,
                style: TextStyle(
                  fontSize: 12 * scale,
                  fontWeight: isHeader ? FontWeight.bold : FontWeight.normal,
                ),
              ),
            );
          }).toList(),
        );
      }).toList(),
    );
  }
}
