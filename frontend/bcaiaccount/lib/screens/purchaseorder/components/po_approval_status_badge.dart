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

  const POApprovalStatusBadge({
    super.key,
    required this.screenData,
    required this.approvalStatus,
    required this.isLoading,
    this.onWithdrawApproval,
  });

  @override
  State<POApprovalStatusBadge> createState() => _POApprovalStatusBadgeState();
}

class _POApprovalStatusBadgeState extends State<POApprovalStatusBadge> {
  bool _isWithdrawing = false;

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
        color: Colors.grey.shade200,
        border: Border(bottom: BorderSide(color: Colors.grey.shade400, width: 1)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.block, size: 20, color: Colors.grey.shade700),
              const SizedBox(width: 8),
              Text(
                '${global.language("status")}: ${global.language("cancelled")}',
                style: TextStyle(
                  color: Colors.grey.shade700,
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
                color: Colors.grey.shade600,
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
                color: Colors.grey.shade500,
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
      color: Colors.grey.shade100,
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
          // ปุ่มสำหรับสถานะ pending: ถอนการอนุมัติ + ส่งคำเตือน
          if (status == POApprovalStatus.pending) ...[
            const SizedBox(width: 8),
            _buildWithdrawButton(context),
            const SizedBox(width: 8),
            _buildReminderButton(context),
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

  /// ปุ่มถอนการส่งอนุมัติ
  Widget _buildWithdrawButton(BuildContext context) {
    return InkWell(
      onTap: _isWithdrawing ? null : () => _withdrawApproval(context),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: Colors.red.shade50,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.red.shade300),
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
                  color: Colors.red.shade700,
                ),
              )
            else
              Icon(Icons.undo, size: 16, color: Colors.red.shade700),
            SizedBox(width: 4),
            Text(
              _isWithdrawing ? global.language('withdrawing') : global.language('withdraw_approval'),
              style: TextStyle(
                color: Colors.red.shade700,
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
          color: Colors.blue.shade50,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.blue.shade300),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.notifications_active, size: 16, color: Colors.blue.shade700),
            SizedBox(width: 4),
            Text(
              global.language('send_reminder'),
              style: TextStyle(
                color: Colors.blue.shade700,
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
            const Icon(Icons.undo, color: Colors.red),
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
              backgroundColor: Colors.red,
              foregroundColor: Colors.white,
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
                  color: isSuccess ? Colors.green : Colors.red,
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
                  Text('Email: ${result.emailSent} ${global.language("times")}', style: const TextStyle(fontSize: 14, color: Colors.grey)),
                  Text('LINE: ${result.lineSent} ${global.language("times")}', style: const TextStyle(fontSize: 14, color: Colors.grey)),
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
                Icon(Icons.error, color: Colors.red),
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
