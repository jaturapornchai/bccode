import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Card widget สำหรับแสดง PR (ใบขอซื้อ) ในรายการ
/// รองรับ Responsive Layout - แสดงข้อมูลแบบ compact และปรับตามความกว้างหน้าจอ
class PRListCard extends StatelessWidget {
  const PRListCard({
    super.key,
    required this.pr,
    required this.isSelected,
    required this.fontSize,
    this.approvalStatus,
    this.onTap,
  });

  final TransactionModel pr;
  final bool isSelected;
  final double fontSize;
  final POApprovalStatusModel? approvalStatus;
  final VoidCallback? onTap;

  static const double _wideScreenThreshold = 600;

  bool get _isInactive => pr.isdelete == true || pr.iscancel;

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
                  Icon(Icons.assignment_outlined, size: 16, color: global.theme.infoHighlightTextColor),
                  const SizedBox(width: 6),
                  Text(
                    _formatDate(pr.docdatetime),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.w500,
                      color: global.theme.infoHighlightTextColor,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    pr.docno,
                    style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.bold),
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

        // แถวที่ 2: Supplier + Meta info
        Row(
          children: [
            Expanded(
              flex: 3,
              child: Text(
                '${global.activeLangName(pr.custnames ?? [])} (${pr.custcode})',
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

        // แถวที่ 3: Badges
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
                Icon(Icons.assignment_outlined, size: 16, color: global.theme.infoHighlightTextColor),
                const SizedBox(width: 6),
                Text(
                  _formatDate(pr.docdatetime),
                  style: TextStyle(
                    fontSize: fontSize,
                    fontWeight: FontWeight.w500,
                    color: global.theme.infoHighlightTextColor,
                  ),
                ),
              ],
            ),
            Text(
              pr.docno,
              style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.bold),
            ),
          ],
        ),

        const SizedBox(height: 3),

        // แถว 2: Supplier
        Text(
          '${global.activeLangName(pr.custnames ?? [])} (${pr.custcode})',
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
                '${pr.detailcount ?? 0} ${global.language("items")}',
                style: TextStyle(fontSize: fontSize - 1),
              ),
            ),
            const SizedBox(width: 6),
            Icon(Icons.person_outline, size: 12, color: global.theme.iconSecondaryColor),
            const SizedBox(width: 3),
            Expanded(
              child: Text(
                pr.creatorname ?? pr.creatorcode ?? 'System',
                style: TextStyle(fontSize: fontSize - 1, color: global.theme.iconSecondaryColor),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            _buildAmountText(fontSize + 1),
          ],
        ),

        // แถว 4: Badges
        if (_hasBadges) ...[
          const SizedBox(height: 4),
          _buildBadgesRow(fontSize - 1),
        ],
      ],
    );
  }

  bool get _hasBadges {
    return approvalStatus != null ||
        (!pr.iscancel && (pr.isdelete != true) && approvalStatus == null) ||
        pr.isref == true ||
        pr.iscancel ||
        pr.isdelete == true;
  }

  Widget _buildAmountText(double fontSize) {
    final baseAmount = global.safeToDouble(pr.totalamount);
    return Text(
      global.formatNumber(baseAmount),
      style: TextStyle(
        fontSize: fontSize,
        fontWeight: FontWeight.bold,
        color: global.theme.positiveHighlightTextColor,
      ),
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
            '${pr.detailcount ?? 0} ${global.language("items")}',
            style: TextStyle(fontSize: fontSize),
          ),
        ),
        const SizedBox(width: 6),
        Icon(Icons.person_outline, size: 12, color: global.theme.iconSecondaryColor),
        const SizedBox(width: 3),
        Flexible(
          child: Text(
            pr.creatorname ?? pr.creatorcode ?? 'System',
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
        if (pr.isdelete == true)
          _buildStatusChip(
            icon: Icons.delete_forever,
            text: global.language('deleted'),
            color: global.theme.negativeHighlightTextColor,
            fontSize: fontSize,
          ),
        if (pr.iscancel)
          _buildStatusChip(
            icon: Icons.cancel,
            text: global.language('cancelled'),
            color: global.theme.negativeHighlightTextColor,
            fontSize: fontSize,
          ),
        if (pr.isdelete != true && !pr.iscancel) ...[
          if (approvalStatus != null)
            _PRApprovalBadge(status: approvalStatus!, fontSize: fontSize)
          else
            _PRPendingBadge(fontSize: fontSize),
        ],
        if (pr.isref == true)
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
            style: TextStyle(color: color, fontWeight: FontWeight.w600, fontSize: fontSize),
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

/// Badge แสดงสถานะการอนุมัติ PR
class _PRApprovalBadge extends StatelessWidget {
  const _PRApprovalBadge({required this.status, required this.fontSize});

  final POApprovalStatusModel status;
  final double fontSize;

  @override
  Widget build(BuildContext context) {
    Color bgColor;
    Color textColor;
    IconData icon;
    String text;

    switch (status.status) {
      case POApprovalStatus.autoApproved:
        bgColor = global.theme.infoHighlightColor;
        textColor = global.theme.infoHighlightTextColor;
        icon = Icons.verified;
        text = global.language('approval_auto_approved');
        break;
      case POApprovalStatus.approved:
        bgColor = global.theme.positiveHighlightColor;
        textColor = global.theme.positiveHighlightTextColor;
        icon = Icons.check_circle;
        text = global.language('approval_approved');
        break;
      case POApprovalStatus.rejected:
        bgColor = global.theme.negativeHighlightColor;
        textColor = global.theme.negativeHighlightTextColor;
        icon = Icons.cancel;
        text = global.language('reject');
        break;
      case POApprovalStatus.pending:
        bgColor = global.theme.warningHighlightColor;
        textColor = global.theme.warningHighlightTextColor;
        icon = Icons.hourglass_empty;
        text = global.language('pending_approval');
        break;
      case POApprovalStatus.draft:
        bgColor = global.theme.cardColor;
        textColor = global.theme.textSecondaryColor;
        icon = Icons.edit_note;
        text = global.language('draft');
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: textColor.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: textColor),
          const SizedBox(width: 3),
          Text(
            text,
            style: TextStyle(color: textColor, fontWeight: FontWeight.w600, fontSize: fontSize),
          ),
        ],
      ),
    );
  }
}

/// Badge แสดงสถานะรอส่งอนุมัติ
class _PRPendingBadge extends StatelessWidget {
  const _PRPendingBadge({required this.fontSize});

  final double fontSize;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: global.theme.iconSecondaryColor),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.schedule, size: 12, color: global.theme.iconSecondaryColor),
          const SizedBox(width: 3),
          Text(
            global.language('not_yet_submitted'),
            style: TextStyle(
              color: global.theme.iconSecondaryColor,
              fontWeight: FontWeight.w600,
              fontSize: fontSize,
            ),
          ),
        ],
      ),
    );
  }
}
