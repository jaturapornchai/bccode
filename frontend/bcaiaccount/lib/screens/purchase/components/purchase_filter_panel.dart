import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'purchase_search_dialog.dart';
import '../../../global.dart' as global;

/// Panel แสดงเงื่อนไขการค้นหาใบสั่งซื้อ
class PurchaseFilterPanel extends StatelessWidget {
  final PurchaseSearchCondition condition;
  final VoidCallback onEditPressed;
  final VoidCallback onRefreshPressed;
  final VoidCallback? onClearPressed;

  const PurchaseFilterPanel({
    super.key,
    required this.condition,
    required this.onEditPressed,
    required this.onRefreshPressed,
    this.onClearPressed,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 1,
      margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.all(6),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header row with title and buttons
            _buildHeaderRow(),

            // แสดงเงื่อนไขที่มีการตั้งค่า
            if (condition.hasConditions) ...[
              const Divider(height: 8),
              _buildConditionsDisplay(),
            ],
          ],
        ),
      ),
    );
  }

  /// สร้าง header row
  Widget _buildHeaderRow() {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: Colors.blue.shade50,
            borderRadius: BorderRadius.circular(6),
          ),
          child: Icon(Icons.filter_list, size: 16, color: Colors.blue.shade600),
        ),
        SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('search_condition'),
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
              ),
              if (!condition.hasConditions)
                Text(
                  global.language('no_condition_show_all'),
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.grey.shade600,
                  ),
                ),
            ],
          ),
        ),
        // ปุ่มล้างเงื่อนไข
        if (condition.hasConditions && onClearPressed != null)
          IconButton(
            icon: Icon(Icons.clear_all, size: 18, color: Colors.red.shade400),
            tooltip: global.language('clear_conditions'),
            onPressed: onClearPressed,
            padding: const EdgeInsets.all(4),
            constraints: const BoxConstraints(),
          ),
        const SizedBox(width: 4),
        // ปุ่มรีเฟรช
        IconButton(
          icon: Icon(Icons.refresh, size: 18, color: Colors.green.shade600),
          tooltip: global.language('database_master_info.refresh'),
          onPressed: onRefreshPressed,
          padding: const EdgeInsets.all(4),
          constraints: const BoxConstraints(),
        ),
        const SizedBox(width: 4),
        // ปุ่มแก้ไข
        ElevatedButton.icon(
          onPressed: onEditPressed,
          icon: Icon(Icons.tune, size: 14),
          label: Text(global.language('edit'), style: TextStyle(fontSize: 12)),
          style: ElevatedButton.styleFrom(
            backgroundColor: Colors.blue.shade600,
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(6),
            ),
          ),
        ),
      ],
    );
  }

  /// แสดงเงื่อนไขที่ตั้งค่าไว้
  Widget _buildConditionsDisplay() {
    return Wrap(
      spacing: 4,
      runSpacing: 4,
      children: [
        // ช่วงวันที่
        if (condition.fromDate != null || condition.toDate != null)
          _buildConditionChip(
            icon: Icons.date_range,
            label: _formatDateRange(),
            color: Colors.blue,
          ),

        // ช่วงจำนวนเงิน
        if (condition.minAmount != null || condition.maxAmount != null)
          _buildConditionChip(
            icon: Icons.attach_money,
            label: _formatAmountRange(),
            color: Colors.green,
          ),

        // เลขที่เอกสาร
        if (condition.docNo.isNotEmpty)
          _buildConditionChip(
            icon: Icons.receipt_long,
            label: '${global.language("docno")}: ${condition.docNo}',
            color: Colors.purple,
          ),

        // รหัสผู้ขาย
        if (condition.creditorCode.isNotEmpty)
          _buildConditionChip(
            icon: Icons.badge,
            label: '${global.language("code")}: ${condition.creditorCode}',
            color: Colors.orange,
          ),

        // ชื่อผู้ขาย
        if (condition.creditorName.isNotEmpty)
          _buildConditionChip(
            icon: Icons.business,
            label: '${global.language("seller")}: ${condition.creditorName}',
            color: Colors.teal,
          ),

        // คำค้นหา
        if (condition.searchText.isNotEmpty)
          _buildConditionChip(
            icon: Icons.search,
            label: '"${condition.searchText}"',
            color: Colors.indigo,
          ),
      ],
    );
  }

  /// สร้าง chip แสดงเงื่อนไข
  Widget _buildConditionChip({
    required IconData icon,
    required String label,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.w500,
              color: color.shade700,
            ),
          ),
        ],
      ),
    );
  }

  /// จัดรูปแบบช่วงวันที่
  String _formatDateRange() {
    final dateFormat = DateFormat('dd/MM/yyyy');
    final from = condition.fromDate != null
        ? dateFormat.format(condition.fromDate!)
        : '-';
    final to = condition.toDate != null
        ? dateFormat.format(condition.toDate!)
        : '-';
    return '$from ${global.language("to")} $to';
  }

  /// จัดรูปแบบช่วงจำนวนเงิน
  String _formatAmountRange() {
    final numberFormat = NumberFormat('#,##0.00', 'th');
    final min = condition.minAmount != null
        ? numberFormat.format(condition.minAmount)
        : '-';
    final max = condition.maxAmount != null
        ? numberFormat.format(condition.maxAmount)
        : '-';
    return '$min - $max ${global.language("baht")}';
  }
}

/// Extension สำหรับ Color shade
extension ColorShade on Color {
  Color get shade700 {
    final hsl = HSLColor.fromColor(this);
    return hsl.withLightness((hsl.lightness - 0.1).clamp(0.0, 1.0)).toColor();
  }
}
