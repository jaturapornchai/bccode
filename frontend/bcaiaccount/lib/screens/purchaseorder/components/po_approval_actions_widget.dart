import 'package:flutter/material.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget แสดงปุ่ม actions สำหรับ PO Approval
/// - ส่งอนุมัติ (เมื่อยังไม่ส่ง)
/// - ส่งแจ้งเตือนซ้ำ (เมื่อรออนุมัติ)
class POApprovalActionsWidget extends StatefulWidget {
  const POApprovalActionsWidget({
    super.key,
    required this.docNo,
    required this.approvalStatus,
    required this.onSubmitApproval,
    this.isEnabled = true,
    this.compact = false,
  });

  final String docNo;
  final POApprovalStatusModel? approvalStatus;
  final Future<void> Function() onSubmitApproval;
  final bool isEnabled;
  final bool compact;

  @override
  State<POApprovalActionsWidget> createState() =>
      _POApprovalActionsWidgetState();
}

class _POApprovalActionsWidgetState extends State<POApprovalActionsWidget> {
  bool _isResending = false;

  @override
  Widget build(BuildContext context) {
    if (!widget.isEnabled) {
      return const SizedBox.shrink();
    }

    final status = widget.approvalStatus?.status;

    // ถ้ารออนุมัติ แสดงปุ่มส่งแจ้งเตือนซ้ำ
    if (status == POApprovalStatus.pending) {
      return _buildResendButton(context);
    }

    // ถ้ายังไม่ส่งหรือถูก reject แสดงปุ่มส่งอนุมัติ
    if (status == null ||
        status == POApprovalStatus.draft ||
        status == POApprovalStatus.rejected) {
      return _buildSubmitButton(context);
    }

    return const SizedBox.shrink();
  }

  Widget _buildSubmitButton(BuildContext context) {
    if (widget.compact) {
      return IconButton(
        icon: Icon(Icons.send),
        tooltip: global.language('submit_for_approval'),
        onPressed: widget.onSubmitApproval,
      );
    }

    return ElevatedButton.icon(
      icon: Icon(Icons.send, size: 18),
      label: Text(global.language('submit_for_approval')),
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.blue,
        foregroundColor: Colors.white,
      ),
      onPressed: widget.onSubmitApproval,
    );
  }

  Widget _buildResendButton(BuildContext context) {
    if (widget.compact) {
      return IconButton(
        icon: _isResending
            ? const SizedBox(
                width: 18,
                height: 18,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : const Icon(Icons.notifications_active),
        tooltip: global.language('resend_notification'),
        onPressed: _isResending ? null : () => _resendNotification(context),
      );
    }

    return ElevatedButton.icon(
      icon: _isResending
          ? const SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                color: Colors.white,
              ),
            )
          : const Icon(Icons.notifications_active, size: 18),
      label: Text(_isResending ? global.language('sending') : global.language('resend_notification')),
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.orange,
        foregroundColor: Colors.white,
      ),
      onPressed: _isResending ? null : () => _resendNotification(context),
    );
  }

  Future<void> _resendNotification(BuildContext ctx) async {
    if (widget.docNo.isEmpty) return;

    setState(() => _isResending = true);

    try {
      final result = await ApprovalApiService.resendApprovalNotification(
        docNo: widget.docNo,
      );

      if (!mounted) return;

      if (result.isSuccess) {
        global.showSuccessSnackBar(ctx, result.message ?? global.language('resend_notification_success'));
      } else {
        global.showErrorSnackBar(ctx, result.errorMessage ?? global.language('resend_notification_failed'));
      }
    } finally {
      if (mounted) {
        setState(() => _isResending = false);
      }
    }
  }
}

/// ปุ่ม Quick Actions สำหรับ PO ใน list (compact)
class POQuickActionsBar extends StatelessWidget {
  const POQuickActionsBar({
    super.key,
    required this.approvalStatus,
    this.onEdit,
    this.onPrint,
    this.onCancel,
    this.canEdit = true,
    this.canPrint = false,
    this.canCancel = true,
  });

  final POApprovalStatusModel? approvalStatus;
  final VoidCallback? onEdit;
  final VoidCallback? onPrint;
  final VoidCallback? onCancel;
  final bool canEdit;
  final bool canPrint;
  final bool canCancel;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (canEdit && onEdit != null)
          IconButton(
            icon: Icon(Icons.edit, size: 20),
            tooltip: global.language('edit'),
            onPressed: onEdit,
            color: Colors.blue,
          ),
        if (canPrint && onPrint != null)
          IconButton(
            icon: Icon(Icons.print, size: 20),
            tooltip: global.language('print'),
            onPressed: onPrint,
            color: Colors.green,
          ),
        if (canCancel && onCancel != null)
          IconButton(
            icon: Icon(Icons.cancel, size: 20),
            tooltip: global.language('cancel'),
            onPressed: onCancel,
            color: Colors.red,
          ),
      ],
    );
  }
}
