import 'package:flutter/material.dart';
import 'form_element.dart';
import 'form_layout.dart';
import 'form_section.dart';
import 'form_template.dart';

/// Visual style ที่ apply ทับ template ได้ทุก doc type
/// เปลี่ยนสี, เส้นขอบ, separator style โดยไม่แตะโครงสร้าง
class FormStyle {
  final String id;
  final String name;
  final String nameEn;
  final String description;
  final IconData icon;

  // Colors
  final Color titleColor; // document title, bold headers
  final Color textColor; // regular text, data fields
  final Color labelColor; // field labels, small text
  final Color borderColor; // borders, separators, tables
  final Color? accentColor; // optional accent highlight

  // Line styling
  final double borderWidth;
  final double tableBorderWidth;
  final String separatorStyle; // solid, dashed, dotted

  const FormStyle({
    required this.id,
    required this.name,
    required this.nameEn,
    required this.description,
    required this.icon,
    required this.titleColor,
    required this.textColor,
    required this.labelColor,
    required this.borderColor,
    this.accentColor,
    this.borderWidth = 0.5,
    this.tableBorderWidth = 0.5,
    this.separatorStyle = 'solid',
  });

  /// Apply style colors to template (returns new template)
  FormTemplate apply(FormTemplate template) {
    return template.copyWith(
      header: _applyToSection(template.header),
      detail: _applyToSection(template.detail),
      footer: _applyToSection(template.footer),
    );
  }

  FormSection _applyToSection(FormSection section) {
    if (!section.hasRows) return section;
    return section.copyWith(
      rows: section.rows.map(_applyToRow).toList(),
    );
  }

  FormLayoutRow _applyToRow(FormLayoutRow row) {
    return row.copyWith(
      cells: row.cells.map(_applyToCell).toList(),
    );
  }

  FormLayoutCell _applyToCell(FormLayoutCell cell) {
    switch (cell.type) {
      case ElementType.separator:
      case ElementType.line:
        return cell.copyWith(
          borderColor: borderColor,
          borderWidth: borderWidth,
          lineStyle: separatorStyle,
        );
      case ElementType.table:
        return cell.copyWith(
          textColor: textColor,
          borderColor: borderColor,
          borderWidth: tableBorderWidth,
        );
      case ElementType.text:
        // Title = bold + fontSize >= 10
        final isTitle = cell.fontBold && cell.fontSize >= 10;
        return cell.copyWith(
          textColor: isTitle ? titleColor : labelColor,
          borderColor: cell.borderColor != null ? borderColor : null,
          borderWidth: cell.borderColor != null ? borderWidth : null,
        );
      case ElementType.dataField:
        return cell.copyWith(
          textColor: textColor,
          borderColor: cell.borderColor != null ? borderColor : null,
          borderWidth: cell.borderColor != null ? borderWidth : null,
        );
      case ElementType.signature:
        return cell.copyWith(textColor: textColor);
      case ElementType.pageNumber:
      case ElementType.dateTime:
        return cell.copyWith(textColor: labelColor);
      default:
        return cell.copyWith(textColor: textColor);
    }
  }
}

/// Predefined styles — 4 สไตล์ชื่อเก๋
const List<FormStyle> formStyles = [
  // ── 1. นพรัตน์ — คลาสสิกทางการ ──
  FormStyle(
    id: 'nopparat',
    name: 'นพรัตน์',
    nameEn: 'Nopparat',
    description: 'คลาสสิกเข้มข้น เส้นหนาคมชัด สง่างาม',
    icon: Icons.shield_outlined,
    titleColor: Color(0xFF000000),
    textColor: Color(0xFF000000),
    labelColor: Color(0xFF212121),
    borderColor: Color(0xFF000000),
    borderWidth: 0.8,
    tableBorderWidth: 0.8,
    separatorStyle: 'solid',
  ),

  // ── 2. เส้นสาย — โมเดิร์นมินิมอล ──
  FormStyle(
    id: 'sensai',
    name: 'เส้นสาย',
    nameEn: 'Sen Sai',
    description: 'สะอาด เรียบง่าย โมเดิร์น สายมินิมอล',
    icon: Icons.auto_awesome_outlined,
    titleColor: Color(0xFF3B5998),
    textColor: Color(0xFF333333),
    labelColor: Color(0xFF546E7A),
    borderColor: Color(0xFFB0BEC5),
    borderWidth: 0.3,
    tableBorderWidth: 0.3,
    separatorStyle: 'dotted',
  ),

  // ── 3. คราม — หรูหราพรีเมียม ──
  FormStyle(
    id: 'kram',
    name: 'คราม',
    nameEn: 'Kram',
    description: 'หรูหราลึกซึ้ง โทนครามเข้ม พรีเมียม',
    icon: Icons.diamond_outlined,
    titleColor: Color(0xFF3F51B5),
    textColor: Color(0xFF1A237E),
    labelColor: Color(0xFF3F51B5),
    borderColor: Color(0xFF7986CB),
    borderWidth: 0.5,
    tableBorderWidth: 0.5,
    separatorStyle: 'dashed',
  ),

  // ── 4. อรุณ — อบอุ่นสดใส ──
  FormStyle(
    id: 'arun',
    name: 'อรุณ',
    nameEn: 'Arun',
    description: 'อบอุ่นสดใส เขียวอำพัน สไตล์สร้างสรรค์',
    icon: Icons.wb_sunny_outlined,
    titleColor: Color(0xFFFF6F00),
    textColor: Color(0xFF004D40),
    labelColor: Color(0xFF00796B),
    borderColor: Color(0xFF80CBC4),
    borderWidth: 0.4,
    tableBorderWidth: 0.4,
    separatorStyle: 'solid',
  ),
];

/// หา style ตาม id
FormStyle? findFormStyle(String id) {
  for (final s in formStyles) {
    if (s.id == id) return s;
  }
  return null;
}
