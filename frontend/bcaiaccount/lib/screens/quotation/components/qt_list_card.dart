import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_list_card.dart';
import 'package:smlaicloud/global.dart' as global;

/// Card widget สำหรับแสดงใบเสนอราคาในรายการ
/// รองรับ Responsive Layout - แสดงข้อมูลแบบ compact และปรับตามความกว้างหน้าจอ
/// Reuse POApprovalBadge, POPendingBadge จาก po_list_card.dart
class QTListCard extends StatelessWidget {
  const QTListCard({
    super.key,
    required this.qt,
    required this.isSelected,
    required this.fontSize,
    this.approvalStatus,
    this.onTap,
  });

  final TransactionModel qt;
  final bool isSelected;
  final double fontSize;
  final POApprovalStatusModel? approvalStatus;
  final VoidCallback? onTap;

  static const double _wideScreenThreshold = 600;

  bool get _isInactive => qt.isdelete == true || qt.iscancel;

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    final isWideScreen = size.width > _wideScreenThreshold;
    final compactFontSize = fontSize - 1;

    return Opacity(
      opacity: _isInactive ? 0.55 : 1.0,
      child: Card(
        margin: const EdgeInsets.only(bottom: 6),
        elevation: isSelected ? 3 : 1,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: isSelected ? BorderSide(color: global.theme.infoHighlightTextColor, width: 2) : BorderSide.none,
        ),
        child: InkWell(
          borderRadius: BorderRadius.circular(8),
          onTap: onTap,
          child: Padding(
            padding: isWideScreen
                ? const EdgeInsets.symmetric(horizontal: 12, vertical: 8)
                : const EdgeInsets.all(10),
            child: isWideScreen
                ? _buildWideLayout(compactFontSize)
                : _buildCompactLayout(compactFontSize),
          ),
        ),
      ),
    );
  }

  /// Layout สำหรับหน้าจอกว้าง
  Widget _buildWideLayout(double fontSize) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // แถวที่ 1: Date + DocNo + Amount
        Row(
          children: [
            Expanded(
              flex: 3,
              child: Row(
                children: [
                  Icon(Icons.request_quote_outlined, size: 16, color: Colors.teal[700]),
                  const SizedBox(width: 6),
                  Text(
                    _formatDate(qt.docdatetime),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.w500,
                      color: Colors.teal[700],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    qt.docno,
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              flex: 2,
              child: Align(
                alignment: Alignment.centerRight,
                child: _buildAmountText(fontSize + 1),
              ),
            ),
          ],
        ),

        const SizedBox(height: 4),

        // แถวที่ 2: ลูกค้า + Meta info
        Row(
          children: [
            Expanded(
              flex: 3,
              child: Text(
                '${global.activeLangName(qt.custnames ?? [])} (${qt.custcode})',
                style: TextStyle(fontSize: fontSize),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 2,
              child: Align(
                alignment: Alignment.centerRight,
                child: _buildMetaInfoInline(fontSize - 1),
              ),
            ),
          ],
        ),

        // แถวที่ 3: Badges (ถ้ามี)
        if (_hasBadges) ...[
          const SizedBox(height: 4),
          _buildBadgesRow(fontSize - 1),
        ],
      ],
    );
  }

  /// Layout สำหรับหน้าจอปกติ
  Widget _buildCompactLayout(double fontSize) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        // แถว 1: Date + DocNo
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              children: [
                Icon(Icons.request_quote_outlined, size: 16, color: Colors.teal[700]),
                const SizedBox(width: 6),
                Text(
                  _formatDate(qt.docdatetime),
                  style: TextStyle(
                    fontSize: fontSize,
                    fontWeight: FontWeight.w500,
                    color: Colors.teal[700],
                  ),
                ),
              ],
            ),
            Text(
              qt.docno,
              style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.bold),
            ),
          ],
        ),

        const SizedBox(height: 3),

        // แถว 2: ลูกค้า
        Text(
          '${global.activeLangName(qt.custnames ?? [])} (${qt.custcode})',
          style: TextStyle(fontSize: fontSize),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),

        const SizedBox(height: 2),

        // แถว 3: Meta info + Amount
        Row(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
              decoration: BoxDecoration(
                color: global.theme.dividerBorderColor,
                borderRadius: BorderRadius.circular(3),
              ),
              child: Text(
                '${qt.detailcount ?? 0} รายการ',
                style: TextStyle(fontSize: fontSize - 1),
              ),
            ),
            const SizedBox(width: 6),
            Icon(Icons.person_outline, size: 12, color: global.theme.iconSecondaryColor),
            const SizedBox(width: 3),
            Expanded(
              child: Text(
                qt.creatorname ?? qt.creatorcode ?? 'System',
                style: TextStyle(fontSize: fontSize - 1, color: global.theme.iconSecondaryColor),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            _buildAmountText(fontSize + 1),
          ],
        ),

        // แถว 4: Badges (ถ้ามี)
        if (_hasBadges) ...[
          const SizedBox(height: 4),
          _buildBadgesRow(fontSize - 1),
        ],
      ],
    );
  }

  bool get _hasBadges {
    return approvalStatus != null ||
        (!qt.iscancel && (qt.isdelete != true) && approvalStatus == null) ||
        qt.isref == true ||
        qt.iscancel ||
        qt.isdelete == true;
  }

  Widget _buildAmountText(double fontSize) {
    final baseAmount = global.safeToDouble(qt.totalamount);
    final baseCurrency = (qt.currency ?? global.getBaseCurrency()).toUpperCase();
    final docCurrency = (qt.docCurrency ?? '').toUpperCase();
    final docSymbol = qt.docCurrencySymbol ?? '';
    final baseSymbol = global.getBaseCurrencySymbol();
    double safeRate = qt.exchangerate ?? 1.0;
    if (safeRate <= 0 || safeRate.isNaN || safeRate.isInfinite) safeRate = 1.0;

    final isMultiCurrency = docCurrency.isNotEmpty &&
                            docCurrency != baseCurrency &&
                            safeRate != 1.0;

    if (!isMultiCurrency) {
      return Text(
        global.formatNumber(baseAmount),
        style: TextStyle(
          fontSize: fontSize,
          fontWeight: FontWeight.bold,
          color: global.theme.positiveHighlightTextColor,
        ),
      );
    }

    double docAmount = global.safeToDouble(qt.totalamountDoc);
    if (docAmount == 0 && baseAmount > 0) {
      final dp = global.getDecimalDocument();
      docAmount = double.parse((baseAmount / safeRate).toStringAsFixed(dp));
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          '$docSymbol${global.formatNumber(docAmount)} $docCurrency',
          style: TextStyle(
            fontSize: fontSize,
            fontWeight: FontWeight.bold,
            color: global.theme.infoHighlightTextColor,
          ),
        ),
        Text(
          '$baseSymbol${global.formatNumber(baseAmount)} $baseCurrency',
          style: TextStyle(
            fontSize: fontSize - 2,
            fontWeight: FontWeight.w500,
            color: global.theme.positiveHighlightTextColor,
          ),
        ),
      ],
    );
  }

  Widget _buildMetaInfoInline(double fontSize) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
          decoration: BoxDecoration(
            color: global.theme.dividerBorderColor,
            borderRadius: BorderRadius.circular(3),
          ),
          child: Text(
            '${qt.detailcount ?? 0} รายการ',
            style: TextStyle(fontSize: fontSize),
          ),
        ),
        const SizedBox(width: 6),
        Icon(Icons.person_outline, size: 12, color: global.theme.iconSecondaryColor),
        const SizedBox(width: 3),
        Flexible(
          child: Text(
            qt.creatorname ?? qt.creatorcode ?? 'System',
            style: TextStyle(fontSize: fontSize, color: global.theme.iconSecondaryColor),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }

  Widget _buildBadgesRow(double fontSize) {
    return Wrap(
      spacing: 6,
      runSpacing: 4,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        if (qt.isdelete == true)
          _buildStatusChip(
            icon: Icons.delete_forever,
            text: global.language('deleted'),
            color: global.theme.negativeHighlightTextColor,
            fontSize: fontSize,
          ),

        if (qt.iscancel)
          _buildStatusChip(
            icon: Icons.cancel,
            text: global.language('cancelled'),
            color: global.theme.negativeHighlightTextColor,
            fontSize: fontSize,
          ),

        // Reuse POApprovalBadge / POPendingBadge จาก po_list_card.dart
        if (qt.isdelete != true && !qt.iscancel) ...[
          if (approvalStatus != null)
            POApprovalBadge(status: approvalStatus!, fontSize: fontSize, isCompact: true)
          else
            POPendingBadge(fontSize: fontSize, isCompact: true),
        ],

        if (qt.isref == true)
          _buildStatusChip(
            icon: Icons.link,
            text: global.language('referenced'),
            color: global.theme.infoHighlightTextColor,
            fontSize: fontSize,
          ),
      ],
    );
  }

  Widget _buildStatusChip({
    required IconData icon,
    required String text,
    required Color color,
    required double fontSize,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color.withValues(alpha: 0.5)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 3),
          Text(
            text,
            style: TextStyle(
              color: color,
              fontWeight: FontWeight.w600,
              fontSize: fontSize,
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(String? dateString) {
    if (dateString == null || dateString.isEmpty) return '-';
    try {
      final date = DateTime.parse(dateString).toLocal();
      if (global.profileData.yeartype == 'buddhist') {
        return global.dateTimeBuddhist(date, format: global.DateTimeFormatEnum.dateDay);
      }
      return DateFormat('dd/MM/yyyy HH:mm').format(date);
    } catch (e) {
      return dateString;
    }
  }
}
