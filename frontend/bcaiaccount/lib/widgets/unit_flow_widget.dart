// วิดเจ็ตแสดงแผนผังการแปลงหน่วยสินค้า
// แสดง 2 โหมดต่อกัน: ผังภาพ (flow diagram) + ตารางแปลงหน่วย (conversion table)
// เรียงตามอัตราส่วน (ตัวตั้ง/ตัวหาร) จากเล็กไปใหญ่
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/product_model.dart';

/// โมเดลภายในสำหรับจัดกลุ่ม barcode ตามหน่วย
class _UnitGroup {
  final String unitCode;
  final String unitName;

  /// อัตราส่วนเทียบหน่วยฐาน (ตัวตั้ง/ตัวหาร)
  final double ratio;
  final List<String> barcodes;

  _UnitGroup({
    required this.unitCode,
    required this.unitName,
    required this.ratio,
    required this.barcodes,
  });
}

/// แสดงแผนผังการแปลงหน่วยสินค้า — 2 โหมดต่อกัน (ผัง + ตาราง)
class UnitFlowWidget extends StatelessWidget {
  /// รายการ barcode ทั้งหมดของสินค้านี้
  final List<ProductBarcodeModel> productBarcodes;

  /// barcode ที่กำลังเลือกอยู่
  final String currentBarcode;

  const UnitFlowWidget({
    super.key,
    required this.productBarcodes,
    required this.currentBarcode,
  });

  @override
  Widget build(BuildContext context) {
    // ไม่แสดงถ้าไม่มี barcode หรือมีแค่ 1 รายการ
    if (productBarcodes.isEmpty || productBarcodes.length <= 1) {
      return const SizedBox.shrink();
    }

    // จัดกลุ่ม barcode ตาม unit code + คำนวณ ratio
    final unitGroups = _buildUnitGroups();

    // ไม่แสดงถ้ามีแค่ 1 กลุ่มหน่วย
    if (unitGroups.length <= 1) {
      return const SizedBox.shrink();
    }

    final accentColor = global.theme.appBarColor;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: accentColor.withValues(alpha: 0.15),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // หัวข้อ
            Row(
              children: [
                Icon(Icons.swap_horiz_rounded, size: 18, color: accentColor),
                const SizedBox(width: 8),
                Text(
                  global.language('unit_conversion'),
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: accentColor,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            // โหมด 1: ผังภาพ (flow diagram)
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: _buildFlowDiagram(unitGroups, accentColor),
            ),
            const SizedBox(height: 12),
            Divider(color: accentColor.withValues(alpha: 0.1), height: 1),
            const SizedBox(height: 12),
            // โหมด 2: ตารางแปลงหน่วย
            _buildConversionTable(unitGroups, accentColor),
          ],
        ),
      ),
    );
  }

  /// คำนวณ ratio จาก ProductBarcodeModel
  /// ลำดับ: refbarcodes → top-level standvalue/dividevalue → 1
  double _computeRatio(ProductBarcodeModel item) {
    // ลำดับ 1: จาก refbarcodes (ข้อมูลจาก MongoDB detail)
    if (item.refbarcodes != null && item.refbarcodes!.isNotEmpty) {
      final ref = item.refbarcodes!.first;
      final dv = ref.dividevalue > 0 ? ref.dividevalue : 1.0;
      if (ref.standvalue > 0) return ref.standvalue / dv;
    }
    // ลำดับ 2: จาก top-level standvalue/dividevalue (PG list API)
    final sv = item.standvalue ?? 0;
    final dv = (item.dividevalue ?? 0) > 0 ? item.dividevalue! : 1.0;
    if (sv > 0) return sv / dv;
    // ลำดับ 3: ถือเป็นหน่วยฐาน
    return 1;
  }

  /// จัดกลุ่ม barcode ตามหน่วย + คำนวณ ratio
  /// เรียงตาม ratio น้อยไปมาก (หน่วยเล็กสุดก่อน)
  List<_UnitGroup> _buildUnitGroups() {
    final Map<String, _UnitGroup> groupMap = {};

    for (final item in productBarcodes) {
      final unitCode = item.itemunitcode;
      final ratio = _computeRatio(item);

      if (groupMap.containsKey(unitCode)) {
        groupMap[unitCode]!.barcodes.add(item.barcode ?? '');
      } else {
        // ดึงชื่อหน่วยจากภาษาที่ใช้งานอยู่
        final unitName =
            (item.itemunitnames != null && item.itemunitnames!.isNotEmpty)
                ? global.activeLangName(item.itemunitnames!)
                : unitCode;

        groupMap[unitCode] = _UnitGroup(
          unitCode: unitCode,
          unitName: unitName.isNotEmpty ? unitName : unitCode,
          ratio: ratio,
          barcodes: [item.barcode ?? ''],
        );
      }
    }

    // เรียงจาก ratio น้อยไปมาก (หน่วยเล็กสุดก่อน)
    final groups = groupMap.values.toList()
      ..sort((a, b) => a.ratio.compareTo(b.ratio));

    return groups;
  }

  /// หา unit code ของ barcode ที่กำลังเลือกอยู่
  String _findCurrentUnitCode() {
    for (final item in productBarcodes) {
      if (item.barcode == currentBarcode) {
        return item.itemunitcode;
      }
    }
    return '';
  }

  // ─── โหมด 1: ผังภาพ (Flow Diagram) ─────────────────────────

  /// สร้าง flow diagram แนวนอน
  Widget _buildFlowDiagram(List<_UnitGroup> groups, Color accentColor) {
    final currentUnitCode = _findCurrentUnitCode();
    final List<Widget> children = [];

    for (int i = 0; i < groups.length; i++) {
      final group = groups[i];
      final isSelected = group.unitCode == currentUnitCode;

      children.add(_buildUnitNode(group, isSelected, accentColor));

      // ลูกศรพร้อมตัวคูณ
      if (i < groups.length - 1) {
        final nextGroup = groups[i + 1];
        final multiplier =
            group.ratio > 0 ? nextGroup.ratio / group.ratio : nextGroup.ratio;
        children.add(_buildArrow(multiplier, accentColor));
      }
    }

    return Row(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: children,
    );
  }

  /// node แสดงหน่วย + barcode
  Widget _buildUnitNode(
      _UnitGroup group, bool isSelected, Color accentColor) {
    final bgColor = isSelected
        ? accentColor.withValues(alpha: 0.1)
        : Colors.grey.withValues(alpha: 0.05);
    final borderColor =
        isSelected ? accentColor : Colors.grey.withValues(alpha: 0.3);
    final textColor = isSelected ? accentColor : Colors.grey[700]!;

    return Container(
      constraints: const BoxConstraints(minWidth: 80),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: borderColor, width: isSelected ? 2 : 1),
        boxShadow: isSelected
            ? [
                BoxShadow(
                  color: accentColor.withValues(alpha: 0.15),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ]
            : null,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // ชื่อหน่วย
          Text(
            group.unitName,
            style: TextStyle(
              fontSize: 14,
              fontWeight: isSelected ? FontWeight.w700 : FontWeight.w600,
              color: textColor,
            ),
            textAlign: TextAlign.center,
          ),
          // รหัสหน่วย
          Text(
            group.unitCode,
            style: TextStyle(
              fontSize: 11,
              color: textColor.withValues(alpha: 0.6),
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 4),
          // barcode ในกลุ่มนี้
          ...group.barcodes.map((bc) {
            final isCurrent = bc == currentBarcode;
            return Container(
              margin: const EdgeInsets.only(top: 2),
              padding:
                  const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: isCurrent
                    ? accentColor.withValues(alpha: 0.15)
                    : Colors.white.withValues(alpha: 0.8),
                borderRadius: BorderRadius.circular(4),
                border: isCurrent
                    ? Border.all(color: accentColor, width: 1)
                    : null,
              ),
              child: Text(
                bc,
                style: TextStyle(
                  fontSize: 9,
                  fontWeight: isCurrent ? FontWeight.w600 : FontWeight.w400,
                  color: isCurrent ? accentColor : Colors.grey[600],
                  fontFamily: 'monospace',
                ),
                textAlign: TextAlign.center,
              ),
            );
          }),
        ],
      ),
    );
  }

  /// ลูกศรพร้อมตัวคูณ
  Widget _buildArrow(double multiplier, Color accentColor) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: accentColor.withValues(alpha: 0.08),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(
              'x${_formatNumber(multiplier)}',
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: accentColor,
              ),
            ),
          ),
          const SizedBox(height: 2),
          SizedBox(
            width: 32,
            child: Row(
              children: [
                Expanded(
                  child: Container(
                    height: 1.5,
                    color: accentColor.withValues(alpha: 0.4),
                  ),
                ),
                Icon(
                  Icons.arrow_forward_ios_rounded,
                  size: 10,
                  color: accentColor.withValues(alpha: 0.6),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ─── โหมด 2: ตารางแปลงหน่วย ─────────────────────────────

  /// ตารางแสดงรายละเอียดหน่วย + อัตราแปลง
  Widget _buildConversionTable(
      List<_UnitGroup> groups, Color accentColor) {
    final currentUnitCode = _findCurrentUnitCode();
    final baseGroup = groups.first;
    final List<Widget> rows = [];

    // แสดงทุกหน่วยพร้อมรายละเอียด
    for (int i = 0; i < groups.length; i++) {
      final group = groups[i];
      final isSelected = group.unitCode == currentUnitCode;
      final isBase = i == 0;

      // คำนวณอัตราแปลงกับหน่วยก่อนหน้า
      double? multiplierPrev;
      String? prevUnitName;
      String? prevUnitCode;
      if (i > 0) {
        final prev = groups[i - 1];
        multiplierPrev =
            prev.ratio > 0 ? group.ratio / prev.ratio : group.ratio;
        prevUnitName = prev.unitName;
        prevUnitCode = prev.unitCode;
      }

      // อัตราแปลงกับหน่วยฐาน
      final ratioToBase =
          baseGroup.ratio > 0 ? group.ratio / baseGroup.ratio : group.ratio;

      rows.add(
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          margin: const EdgeInsets.only(bottom: 4),
          decoration: BoxDecoration(
            color: isSelected
                ? accentColor.withValues(alpha: 0.06)
                : Colors.transparent,
            borderRadius: BorderRadius.circular(8),
            border: isSelected
                ? Border.all(color: accentColor.withValues(alpha: 0.3))
                : null,
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ไอคอน
              Padding(
                padding: const EdgeInsets.only(top: 2),
                child: Icon(
                  isBase
                      ? Icons.star_rounded
                      : Icons.compare_arrows_rounded,
                  size: 16,
                  color: isBase
                      ? Colors.amber[700]
                      : accentColor.withValues(alpha: 0.6),
                ),
              ),
              const SizedBox(width: 8),
              // ข้อมูลหน่วย
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // ชื่อหน่วย + รหัส
                    Row(
                      children: [
                        Text(
                          group.unitName,
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                            color: isSelected
                                ? accentColor
                                : Colors.grey[800],
                          ),
                        ),
                        const SizedBox(width: 6),
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 5, vertical: 1),
                          decoration: BoxDecoration(
                            color: Colors.grey.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            group.unitCode,
                            style: TextStyle(
                              fontSize: 10,
                              color: Colors.grey[600],
                              fontFamily: 'monospace',
                            ),
                          ),
                        ),
                        if (isBase) ...[
                          const SizedBox(width: 6),
                          Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 5, vertical: 1),
                            decoration: BoxDecoration(
                              color: Colors.amber.withValues(alpha: 0.15),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              global.language('base_unit'),
                              style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.w500,
                                color: Colors.amber[800],
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                    // อัตราแปลงหน่วย
                    if (!isBase && multiplierPrev != null) ...[
                      const SizedBox(height: 4),
                      Text.rich(
                        TextSpan(children: [
                          TextSpan(
                            text: '1 ${group.unitName}',
                            style: TextStyle(
                              fontWeight: FontWeight.w600,
                              color: isSelected
                                  ? accentColor
                                  : Colors.grey[700],
                            ),
                          ),
                          TextSpan(
                            text: '  =  ',
                            style: TextStyle(color: Colors.grey[400]),
                          ),
                          TextSpan(
                            text:
                                '${_formatNumber(multiplierPrev)} ${prevUnitName ?? ''} (${prevUnitCode ?? ''})',
                            style: TextStyle(
                              fontWeight: FontWeight.w600,
                              color: accentColor,
                            ),
                          ),
                        ]),
                        style: const TextStyle(fontSize: 12),
                      ),
                      // แสดงเทียบหน่วยฐานถ้าห่างมากกว่า 1 ระดับ
                      if (i > 1)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(
                            '= ${_formatNumber(ratioToBase)} ${baseGroup.unitName} (${baseGroup.unitCode})',
                            style: TextStyle(
                              fontSize: 11,
                              color: Colors.grey[500],
                            ),
                          ),
                        ),
                    ],
                    // barcode ในกลุ่มนี้
                    const SizedBox(height: 4),
                    Wrap(
                      spacing: 6,
                      runSpacing: 4,
                      children: group.barcodes.map((bc) {
                        final isCurrent = bc == currentBarcode;
                        return Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: isCurrent
                                ? accentColor.withValues(alpha: 0.12)
                                : Colors.grey.withValues(alpha: 0.08),
                            borderRadius: BorderRadius.circular(4),
                            border: isCurrent
                                ? Border.all(
                                    color:
                                        accentColor.withValues(alpha: 0.5))
                                : null,
                          ),
                          child: Text(
                            bc,
                            style: TextStyle(
                              fontSize: 10,
                              fontFamily: 'monospace',
                              color: isCurrent
                                  ? accentColor
                                  : Colors.grey[600],
                            ),
                          ),
                        );
                      }).toList(),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: rows,
    );
  }

  /// จัดรูปแบบตัวเลข — ถ้าเป็นจำนวนเต็มไม่ต้องแสดงทศนิยม
  String _formatNumber(double value) {
    if (value == value.roundToDouble()) {
      return value.toInt().toString();
    }
    return value.toStringAsFixed(2);
  }
}
