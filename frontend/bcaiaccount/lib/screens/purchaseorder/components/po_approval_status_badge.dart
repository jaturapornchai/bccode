import 'package:flutter/material.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_helper.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/widgets/po_approval_history_dialog.dart';
import '../../../global.dart' as global;

/// Widget แสดงสถานะการอนุมัติ PO - แสดงเป็น banner ด้านบนของเอกสาร
class POApprovalStatusBadge extends StatefulWidget {
  final TransactionModel screenData;
  final POApprovalStatusModel? approvalStatus;
  final bool isLoading;
  final Future<bool> Function()? onWithdrawApproval;
  final Future<bool> Function(String comment)? onApprove;
  final Future<bool> Function(String comment)? onReject;

  const POApprovalStatusBadge({
    super.key,
    required this.screenData,
    required this.approvalStatus,
    required this.isLoading,
    this.onWithdrawApproval,
    this.onApprove,
    this.onReject,
  });

  @override
  State<POApprovalStatusBadge> createState() => _POApprovalStatusBadgeState();
}

class _POApprovalStatusBadgeState extends State<POApprovalStatusBadge> with global.ThemeRefreshMixin {
  bool _isWithdrawing = false;
  bool _isApproving = false;
  bool _isRejecting = false;

  @override
  Widget build(BuildContext context) {
    // ตรวจสอบสถานะยกเลิกก่อน - แสดงทันทีไม่ต้องรอ approval status
    if (widget.screenData.iscancel == true) {
      return _buildCancelledBadge();
    }

    if (widget.isLoading) {
      return _buildLoadingBadge();
    }

    if (widget.approvalStatus == null) {
      return const SizedBox.shrink();
    }

    return _buildApprovalBadge(context);
  }

  /// Badge สำหรับเอกสารที่ถูกยกเลิก
  Widget _buildCancelledBadge() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: global.theme.dividerBorderColor,
        border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.block, size: 20, color: global.theme.textColor),
              const SizedBox(width: 8),
              Text(
                '${global.language("status")}: ${global.language("cancelled")}',
                style: TextStyle(
                  color: global.theme.textColor,
                  fontWeight: FontWeight.bold,
                  fontSize: 14,
                ),
              ),
            ],
          ),
          // แสดงเหตุผลการยกเลิก (ถ้ามี)
          if ((widget.screenData.cancelreason ?? '').isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              '${global.language("reason")}: ${widget.screenData.cancelreason}',
              style: TextStyle(
                color: global.theme.iconSecondaryColor,
                fontSize: 12,
              ),
              textAlign: TextAlign.center,
            ),
          ],
          // แสดงข้อมูลผู้ยกเลิก (ถ้ามี)
          if ((widget.screenData.cancelusercode ?? '').isNotEmpty || (widget.screenData.canceldatetime ?? '').isNotEmpty) ...[
            const SizedBox(height: 2),
            Text(
              '${global.language("by")}: ${(widget.screenData.cancelusername ?? '').isNotEmpty ? widget.screenData.cancelusername : widget.screenData.cancelusercode}',
              style: TextStyle(
                color: global.theme.iconSecondaryColor,
                fontSize: 11,
              ),
            ),
          ],
        ],
      ),
    );
  }

  /// Badge แสดง loading
  Widget _buildLoadingBadge() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      color: global.theme.surfaceColor,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(
            width: 16,
            height: 16,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          const SizedBox(width: 8),
          Text(global.language('loading_approval_status')),
        ],
      ),
    );
  }

  /// Badge แสดงสถานะการอนุมัติ
  Widget _buildApprovalBadge(BuildContext context) {
    final status = widget.approvalStatus!.status;
    final style = POApprovalHelper.getStatusStyle(status);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: style.bgColor,
        border: Border(bottom: BorderSide(color: style.borderColor, width: 1)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(style.icon, size: 20, color: style.textColor),
          const SizedBox(width: 8),
          Text(
            '${global.language("approval_status")}: ${style.statusText}',
            style: TextStyle(
              color: style.textColor,
              fontWeight: FontWeight.bold,
              fontSize: 14,
            ),
          ),
          const SizedBox(width: 12),
          // ปุ่มดูประวัติการอนุมัติ
          _buildHistoryButton(context, style.textColor),
          // ปุ่มสำหรับสถานะ pending: อนุมัติ + ปฏิเสธ + ถอน + ส่งคำเตือน
          if (status == POApprovalStatus.pending) ...[
            if (widget.onApprove != null) ...[
              const SizedBox(width: 8),
              _buildApproveButton(context),
            ],
            if (widget.onReject != null) ...[
              const SizedBox(width: 8),
              _buildRejectButton(context),
            ],
            const SizedBox(width: 8),
            _buildWithdrawButton(context),
            const SizedBox(width: 8),
            _buildReminderButton(context),
          ],
          // แสดงเหตุผลการปฏิเสธ (ถ้ามี)
          if (status == POApprovalStatus.rejected && (widget.approvalStatus?.lastComment ?? '').isNotEmpty) ...[
            const SizedBox(width: 12),
            Flexible(
              child: Text(
                '${global.language("reason")}: ${widget.approvalStatus!.lastComment}',
                style: TextStyle(
                  color: style.textColor,
                  fontSize: 12,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ],
      ),
    );
  }

  /// ปุ่มดูประวัติการอนุมัติ
  Widget _buildHistoryButton(BuildContext context, Color textColor) {
    return InkWell(
      onTap: () {
        final docNo = widget.screenData.docno;
        if (docNo.isNotEmpty) {
          POApprovalHistoryDialog.show(context, docNo);
        }
      },
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: textColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: textColor.withValues(alpha: 0.3)),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.history, size: 16, color: textColor),
            SizedBox(width: 4),
            Text(
              global.language('view_history'),
              style: TextStyle(
                color: textColor,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// ปุ่มอนุมัติ
  Widget _buildApproveButton(BuildContext context) {
    return InkWell(
      onTap: _isApproving ? null : () => _approveDocument(context),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: global.theme.positiveHighlightColor,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: global.theme.positiveHighlightTextColor),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (_isApproving)
              SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: global.theme.positiveHighlightTextColor,
                ),
              )
            else
              Icon(Icons.check_circle, size: 16, color: global.theme.positiveHighlightTextColor),
            SizedBox(width: 4),
            Text(
              _isApproving ? global.language('approving') : global.language('approve'),
              style: TextStyle(
                color: global.theme.positiveHighlightTextColor,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// ปุ่มปฏิเสธ
  Widget _buildRejectButton(BuildContext context) {
    return InkWell(
      onTap: _isRejecting ? null : () => _rejectDocument(context),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: global.theme.negativeHighlightColor,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: global.theme.negativeHighlightTextColor),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (_isRejecting)
              SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: global.theme.negativeHighlightTextColor,
                ),
              )
            else
              Icon(Icons.cancel, size: 16, color: global.theme.negativeHighlightTextColor),
            SizedBox(width: 4),
            Text(
              _isRejecting ? global.language('rejecting') : global.language('reject'),
              style: TextStyle(
                color: global.theme.negativeHighlightTextColor,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// อนุมัติเอกสาร
  Future<void> _approveDocument(BuildContext context) async {
    if (widget.onApprove == null) return;

    final confirmed = await showDialog<bool>(
      context: context,
      useRootNavigator: true,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor),
            const SizedBox(width: 8),
            Text(global.language('approve')),
          ],
        ),
        content: Text(global.language('confirm_approve_document')),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.positiveHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            child: Text(global.language('approve')),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _isApproving = true);

    try {
      await widget.onApprove!('');
    } finally {
      if (mounted) {
        setState(() => _isApproving = false);
      }
    }
  }

  /// ปฏิเสธเอกสารพร้อมเหตุผล
  Future<void> _rejectDocument(BuildContext context) async {
    if (widget.onReject == null) return;

    final commentController = TextEditingController();

    final comment = await showDialog<String>(
      context: context,
      useRootNavigator: true,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.cancel, color: global.theme.negativeHighlightTextColor),
            const SizedBox(width: 8),
            Text(global.language('reject')),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(global.language('reject_reason_description')),
            const SizedBox(height: 12),
            TextField(
              controller: commentController,
              maxLines: 3,
              decoration: InputDecoration(
                hintText: global.language('reject_reason_hint'),
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                contentPadding: const EdgeInsets.all(12),
              ),
              autofocus: true,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(null),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(ctx).pop(commentController.text),
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.negativeHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            child: Text(global.language('reject')),
          ),
        ],
      ),
    );

    if (comment == null) return;

    setState(() => _isRejecting = true);

    try {
      await widget.onReject!(comment);
    } finally {
      if (mounted) {
        setState(() => _isRejecting = false);
      }
    }
  }

  /// ปุ่มถอนการส่งอนุมัติ
  Widget _buildWithdrawButton(BuildContext context) {
    return InkWell(
      onTap: _isWithdrawing ? null : () => _withdrawApproval(context),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: global.theme.negativeHighlightColor,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: global.theme.negativeHighlightColor),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (_isWithdrawing)
              SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: global.theme.negativeHighlightTextColor,
                ),
              )
            else
              Icon(Icons.undo, size: 16, color: global.theme.negativeHighlightTextColor),
            SizedBox(width: 4),
            Text(
              _isWithdrawing ? global.language('withdrawing') : global.language('withdraw_approval'),
              style: TextStyle(
                color: global.theme.negativeHighlightTextColor,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// ปุ่มส่งคำเตือนผู้อนุมัติ
  Widget _buildReminderButton(BuildContext context) {
    return InkWell(
      onTap: () => _sendReminder(context),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: global.theme.surfaceColor,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: global.theme.dividerBorderColor),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.notifications_active, size: 16, color: global.theme.infoHighlightTextColor),
            SizedBox(width: 4),
            Text(
              global.language('send_reminder'),
              style: TextStyle(
                color: global.theme.infoHighlightTextColor,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// ถอนการส่งอนุมัติ
  Future<void> _withdrawApproval(BuildContext context) async {
    if (widget.onWithdrawApproval == null) return;

    // แสดง confirm dialog
    final confirmed = await showDialog<bool>(
      context: context,
      useRootNavigator: true,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.undo, color: global.theme.negativeHighlightTextColor),
            const SizedBox(width: 8),
            Text(global.language('withdraw_approval')),
          ],
        ),
        content: Text(global.language('confirm_withdraw_approval')),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.negativeHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            child: Text(global.language('withdraw_approval')),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _isWithdrawing = true);

    try {
      await widget.onWithdrawApproval!();
    } finally {
      if (mounted) {
        setState(() => _isWithdrawing = false);
      }
    }
  }

  /// ส่งคำเตือนผู้อนุมัติ
  Future<void> _sendReminder(BuildContext context) async {
    final docNo = widget.screenData.docno;
    if (docNo.isEmpty) return;

    // แสดง loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      useRootNavigator: true,
      builder: (dialogContext) => const Center(
        child: CircularProgressIndicator(),
      ),
    );

    try {
      // เรียก API ส่งคำเตือน
      final result = await ApprovalApiService.resendApprovalNotification(docNo: docNo);

      // ปิด loading dialog
      if (context.mounted) {
        Navigator.of(context, rootNavigator: true).pop();
      }

      // แสดง dialog ผลลัพธ์
      if (context.mounted) {
        final isSuccess = result.isSuccess && result.sent > 0;
        await showDialog(
          context: context,
          useRootNavigator: true,
          builder: (ctx) => AlertDialog(
            title: Row(
              children: [
                Icon(
                  isSuccess ? Icons.check_circle : Icons.error,
                  color: isSuccess ? global.theme.positiveHighlightTextColor : global.theme.negativeHighlightTextColor,
                ),
                SizedBox(width: 8),
                Text(isSuccess ? global.language('send_success') : global.language('failed')),
              ],
            ),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(result.message ?? result.errorMessage ?? (isSuccess ? global.language('send_notification_success') : global.language('send_notification_failed'))),
                if (isSuccess) ...[
                  const SizedBox(height: 8),
                  Text('Email: ${result.emailSent} ${global.language("times")}', style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor)),
                  Text('LINE: ${result.lineSent} ${global.language("times")}', style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor)),
                ],
              ],
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: Text(global.language('ok')),
              ),
            ],
          ),
        );
      }
    } catch (e) {
      // ปิด loading dialog
      if (context.mounted) {
        Navigator.of(context, rootNavigator: true).pop();
      }
      // แสดง error dialog
      if (context.mounted) {
        await showDialog(
          context: context,
          useRootNavigator: true,
          builder: (ctx) => AlertDialog(
            title: Row(
              children: [
                Icon(Icons.error, color: global.theme.negativeHighlightTextColor),
                SizedBox(width: 8),
                Text(global.language('error_occurred')),
              ],
            ),
            content: Text('$e'),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: Text(global.language('ok')),
              ),
            ],
          ),
        );
      }
    }
  }
}
