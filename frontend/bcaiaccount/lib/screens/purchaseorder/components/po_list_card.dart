import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Card widget สำหรับแสดง PO ในรายการ
/// รองรับ Responsive Layout - แสดงข้อมูลแบบ compact และปรับตามความกว้างหน้าจอ
class POListCard extends StatelessWidget {
  const POListCard({
    super.key,
    required this.po,
    required this.isSelected,
    required this.fontSize,
    this.approvalStatus,
    this.onTap,
  });

  final TransactionModel po;
  final bool isSelected;
  final double fontSize;
  final POApprovalStatusModel? approvalStatus;
  final VoidCallback? onTap;

  /// กำหนด threshold สำหรับ wide screen
  static const double _wideScreenThreshold = 600;

  /// ตรวจสอบว่าเอกสารถูกลบหรือยกเลิก (ใช้สำหรับลด opacity)
  bool get _isInactive => po.isdelete == true || po.iscancel;

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
          side: isSelected ? const BorderSide(color: Colors.blue, width: 2) : BorderSide.none,
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

  /// Layout สำหรับหน้าจอกว้าง - จัดข้อมูลแบบ multi-column
  Widget _buildWideLayout(double fontSize) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // แถวที่ 1: Date + DocNo + Amount (อยู่บรรทัดเดียวกัน)
        Row(
          children: [
            // ซ้าย: Date + DocNo
            Expanded(
              flex: 3,
              child: Row(
                children: [
                  Icon(Icons.description_outlined, size: 16, color: Colors.blue[700]),
                  const SizedBox(width: 6),
                  Text(
                    _formatDate(po.docdatetime),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.w500,
                      color: Colors.blue[700],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    po.docno,
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            // ขวา: Amount
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
            // ซ้าย: Supplier name
            Expanded(
              flex: 3,
              child: Text(
                '${global.activeLangName(po.custnames ?? [])} (${po.custcode})',
                style: TextStyle(fontSize: fontSize),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            // ขวา: Detail count + Creator
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

  /// Layout สำหรับหน้าจอปกติ - compact แต่ยังคงโครงสร้างแถว
  Widget _buildCompactLayout(double fontSize) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        // แถว 1: Date + DocNo (บรรทัดเดียว)
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              children: [
                Icon(Icons.description_outlined, size: 16, color: Colors.blue[700]),
                const SizedBox(width: 6),
                Text(
                  _formatDate(po.docdatetime),
                  style: TextStyle(
                    fontSize: fontSize,
                    fontWeight: FontWeight.w500,
                    color: Colors.blue[700],
                  ),
                ),
              ],
            ),
            Text(
              po.docno,
              style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.bold),
            ),
          ],
        ),

        const SizedBox(height: 3),

        // แถว 2: Supplier (บรรทัดเดียว)
        Text(
          '${global.activeLangName(po.custnames ?? [])} (${po.custcode})',
          style: TextStyle(fontSize: fontSize),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),

        const SizedBox(height: 2),

        // แถว 3: Meta info + Amount (อยู่บรรทัดเดียวกัน)
        Row(
          children: [
            // Meta info
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
              decoration: BoxDecoration(
                color: Colors.grey[200],
                borderRadius: BorderRadius.circular(3),
              ),
              child: Text(
                '${po.detailcount ?? 0} ${global.language("items")}',
                style: TextStyle(fontSize: fontSize - 1),
              ),
            ),
            const SizedBox(width: 6),
            Icon(Icons.person_outline, size: 12, color: Colors.grey[600]),
            const SizedBox(width: 3),
            Expanded(
              child: Text(
                po.creatorname ?? po.creatorcode ?? 'System',
                style: TextStyle(fontSize: fontSize - 1, color: Colors.grey[600]),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            // Amount (ชิดขวา)
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

  /// ตรวจสอบว่ามี badges ที่ต้องแสดงหรือไม่
  bool get _hasBadges {
    return approvalStatus != null ||
        (!po.iscancel && (po.isdelete != true) && approvalStatus == null) ||
        po.isref == true ||
        (po.iscomparedsuccess != null && po.iscomparedsuccess! > 0) ||
        po.isclosedmanual == true ||
        po.iscancel ||
        po.isdelete == true;
  }

  /// สร้าง widget แสดงจำนวนเงิน
  /// - ถ้าสกุลเงินเดียว → แสดงยอดเงินอย่างเดียว
  /// - ถ้า multi-currency → แสดงยอดสกุลเงินเอกสาร + ยอด base currency พร้อม currency code
  Widget _buildAmountText(double fontSize) {
    final baseAmount = global.safeToDouble(po.totalamount);
    final baseCurrency = (po.currency ?? global.getBaseCurrency()).toUpperCase();
    final docCurrency = (po.docCurrency ?? '').toUpperCase();
    final docSymbol = po.docCurrencySymbol ?? '';
    final baseSymbol = global.getBaseCurrencySymbol();
    double safeRate = po.exchangerate ?? 1.0;
    if (safeRate <= 0 || safeRate.isNaN || safeRate.isInfinite) safeRate = 1.0;

    // ตรวจสอบ multi-currency: สกุลเงินเอกสารต่างจาก base และ exchangerate ไม่ใช่ 1
    final isMultiCurrency = docCurrency.isNotEmpty &&
                            docCurrency != baseCurrency &&
                            safeRate != 1.0;

    if (!isMultiCurrency) {
      return Text(
        global.formatNumber(baseAmount),
        style: TextStyle(
          fontSize: fontSize,
          fontWeight: FontWeight.bold,
          color: Colors.green[700],
        ),
      );
    }

    // คำนวณยอดสกุลเงินเอกสาร — ใช้ค่าที่บันทึกไว้ หรือคำนวณจาก base / rate
    double docAmount = global.safeToDouble(po.totalamountDoc);
    if (docAmount == 0 && baseAmount > 0) {
      final dp = global.getDecimalDocument();
      docAmount = double.parse((baseAmount / safeRate).toStringAsFixed(dp));
    }

    // กรณี multi-currency → แสดงทั้งสองสกุลเงินพร้อม currency code
    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      mainAxisSize: MainAxisSize.min,
      children: [
        // ยอดสกุลเงินเอกสาร (เช่น $63.00 USD)
        Text(
          '$docSymbol${global.formatNumber(docAmount)} $docCurrency',
          style: TextStyle(
            fontSize: fontSize,
            fontWeight: FontWeight.bold,
            color: Colors.blue[700],
          ),
        ),
        // ยอด base currency (เช่น ฿2,203.13 THB)
        Text(
          '$baseSymbol${global.formatNumber(baseAmount)} $baseCurrency',
          style: TextStyle(
            fontSize: fontSize - 2,
            fontWeight: FontWeight.w500,
            color: Colors.green[700],
          ),
        ),
      ],
    );
  }

  /// สร้าง meta info แบบ inline (สำหรับ wide screen)
  Widget _buildMetaInfoInline(double fontSize) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
          decoration: BoxDecoration(
            color: Colors.grey[200],
            borderRadius: BorderRadius.circular(3),
          ),
          child: Text(
            '${po.detailcount ?? 0} ${global.language("items")}',
            style: TextStyle(fontSize: fontSize),
          ),
        ),
        const SizedBox(width: 6),
        Icon(Icons.person_outline, size: 12, color: Colors.grey[600]),
        const SizedBox(width: 3),
        Flexible(
          child: Text(
            po.creatorname ?? po.creatorcode ?? 'System',
            style: TextStyle(fontSize: fontSize, color: Colors.grey[600]),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }

  /// สร้างแถว badges ทั้งหมด
  Widget _buildBadgesRow(double fontSize) {
    return Wrap(
      spacing: 6,
      runSpacing: 4,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        // สถานะลบแล้ว (แสดงก่อนเลย เพราะสำคัญที่สุด)
        if (po.isdelete == true)
          _buildStatusChip(
            icon: Icons.delete_forever,
            text: global.language('deleted'),
            color: Colors.red,
            fontSize: fontSize,
          ),

        // สถานะยกเลิก
        if (po.iscancel)
          _buildStatusChip(
            icon: Icons.cancel,
            text: global.language('cancelled'),
            color: Colors.red,
            fontSize: fontSize,
          ),

        // Approval status badge (ไม่แสดงถ้าลบหรือยกเลิก)
        if (po.isdelete != true && !po.iscancel) ...[
          if (approvalStatus != null)
            POApprovalBadge(status: approvalStatus!, fontSize: fontSize, isCompact: true)
          else
            POPendingBadge(fontSize: fontSize, isCompact: true),
        ],

        // สถานะอ้างอิง
        if (po.isref == true)
          _buildStatusChip(
            icon: Icons.link,
            text: global.language('referenced'),
            color: Colors.blue,
            fontSize: fontSize,
          ),

        // สถานะการรับสินค้า (iscomparedsuccess)
        if (po.iscomparedsuccess == 1)
          _buildStatusChip(
            icon: Icons.check_circle,
            text: global.language('fully_received'),
            color: Colors.green,
            fontSize: fontSize,
          ),
        if (po.iscomparedsuccess == 2)
          _buildStatusChip(
            icon: Icons.pending,
            text: global.language('partially_received'),
            color: Colors.orange,
            fontSize: fontSize,
          ),
        if (po.iscomparedsuccess == 3)
          _buildStatusChip(
            icon: Icons.warning,
            text: global.language('over_received'),
            color: Colors.orange,
            fontSize: fontSize,
          ),

        // สถานะปิดเอกสารด้วยมือ
        if (po.isclosedmanual == true)
          _buildStatusChip(
            icon: Icons.lock,
            text: '${global.language("manual_closed")} (${po.closedmanualByName ?? ""})',
            color: Colors.purple,
            fontSize: fontSize,
          ),
      ],
    );
  }

  /// สร้าง status chip แบบ compact
  Widget _buildStatusChip({
    required IconData icon,
    required String text,
    required MaterialColor color,
    required double fontSize,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color[200]!),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color[700]),
          const SizedBox(width: 3),
          Text(
            text,
            style: TextStyle(
              color: color[700],
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

/// Badge แสดงสถานะการอนุมัติ
class POApprovalBadge extends StatelessWidget {
  const POApprovalBadge({
    super.key,
    required this.status,
    required this.fontSize,
    this.isCompact = false,
  });

  final POApprovalStatusModel status;
  final double fontSize;
  final bool isCompact;

  @override
  Widget build(BuildContext context) {
    Color bgColor;
    Color textColor;
    IconData icon;
    String text;

    switch (status.status) {
      case POApprovalStatus.autoApproved:
        bgColor = Colors.blue[50]!;
        textColor = Colors.blue[700]!;
        icon = Icons.verified;
        text = global.language('approval_auto_approved');
        break;
      case POApprovalStatus.approved:
        bgColor = Colors.green[50]!;
        textColor = Colors.green[700]!;
        icon = Icons.check_circle;
        text = global.language('approval_approved');
        break;
      case POApprovalStatus.rejected:
        bgColor = Colors.red[50]!;
        textColor = Colors.red[700]!;
        icon = Icons.cancel;
        text = global.language('reject');
        break;
      case POApprovalStatus.pending:
        bgColor = Colors.orange[50]!;
        textColor = Colors.orange[700]!;
        icon = Icons.hourglass_empty;
        text = global.language('pending_approval');
        break;
      case POApprovalStatus.draft:
        bgColor = Colors.grey[100]!;
        textColor = Colors.grey[600]!;
        icon = Icons.edit_note;
        text = global.language('draft');
        break;
    }

    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: isCompact ? 6 : 8,
        vertical: isCompact ? 2 : 4,
      ),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(isCompact ? 4 : 4),
        border: Border.all(color: textColor.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: isCompact ? 12 : 14, color: textColor),
          const SizedBox(width: 3),
          Text(
            text,
            style: TextStyle(
              color: textColor,
              fontWeight: FontWeight.w600,
              fontSize: fontSize,
            ),
          ),
        ],
      ),
    );
  }
}

/// Badge แสดงสถานะรอส่งอนุมัติ
class POPendingBadge extends StatelessWidget {
  const POPendingBadge({
    super.key,
    required this.fontSize,
    this.isCompact = false,
  });

  final double fontSize;
  final bool isCompact;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: isCompact ? 6 : 8,
        vertical: isCompact ? 2 : 4,
      ),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        borderRadius: BorderRadius.circular(isCompact ? 4 : 4),
        border: Border.all(color: Colors.grey[400]!),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.schedule, size: isCompact ? 12 : 14, color: Colors.grey[600]),
          SizedBox(width: 3),
          Text(
            global.language('not_yet_submitted'),
            style: TextStyle(
              color: Colors.grey[600],
              fontWeight: FontWeight.w600,
              fontSize: fontSize,
            ),
          ),
        ],
      ),
    );
  }
}

/// Badge แสดงสถานะอ้างอิง/ยกเลิก (Legacy - ใช้สำหรับ narrow layout)
class POStatusBadges extends StatelessWidget {
  const POStatusBadges({
    super.key,
    required this.po,
    required this.fontSize,
  });

  final TransactionModel po;
  final double fontSize;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(top: 6),
      padding: const EdgeInsets.all(6),
      decoration: BoxDecoration(
        color: Colors.blue[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: Colors.blue[200]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          if (po.isref == true) ...[
            Row(
              children: [
                Icon(Icons.link, size: 14, color: Colors.blue[700]),
                SizedBox(width: 4),
                Text(
                  global.language('referenced'),
                  style: TextStyle(
                    color: Colors.blue[700],
                    fontWeight: FontWeight.bold,
                    fontSize: fontSize,
                  ),
                ),
              ],
            ),
            Row(
              children: [
                Icon(
                  po.isclosed == true ? Icons.check_circle : Icons.pending,
                  size: 14,
                  color: po.isclosed == true ? Colors.green[700] : Colors.orange[700],
                ),
                SizedBox(width: 4),
                Text(
                  po.isclosed == true ? global.language('fully_received') : global.language('partially_received'),
                  style: TextStyle(
                    color: po.isclosed == true ? Colors.green[700] : Colors.orange[700],
                    fontWeight: FontWeight.bold,
                    fontSize: fontSize,
                  ),
                ),
              ],
            ),
          ],
          if (po.iscancel)
            Row(
              children: [
                Icon(Icons.cancel, size: 14, color: Colors.red[700]),
                SizedBox(width: 4),
                Text(
                  global.language('cancelled'),
                  style: TextStyle(
                    color: Colors.red[700],
                    fontWeight: FontWeight.bold,
                    fontSize: fontSize,
                  ),
                ),
              ],
            ),
        ],
      ),
    );
  }
}
