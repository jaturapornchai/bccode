import 'dart:async';
import 'package:flutter/material.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/global.dart' as global;

/// Dialog แสดงประวัติการอนุมัติ PO (รวมประวัติการแจ้งเตือน)
class POApprovalHistoryDialog extends StatefulWidget {
  final String docNo;

  const POApprovalHistoryDialog({
    super.key,
    required this.docNo,
  });

  /// แสดง Dialog ประวัติการอนุมัติ
  static Future<void> show(BuildContext context, String docNo) {
    return showDialog(
      context: context,
      builder: (context) => POApprovalHistoryDialog(docNo: docNo),
    );
  }

  @override
  State<POApprovalHistoryDialog> createState() => _POApprovalHistoryDialogState();
}

class _POApprovalHistoryDialogState extends State<POApprovalHistoryDialog> {
  bool _isLoading = true;
  POApprovalStatusModel? _status;
  List<ApprovalTimelineItem> _timeline = [];
  Map<String, dynamic>? _timelineStatus; // status จาก timeline API
  String? _errorMessage;

  // Timer สำหรับ polling ทุก 5 วินาที
  Timer? _pollingTimer;
  static const Duration _pollingInterval = Duration(seconds: 5);

  @override
  void initState() {
    super.initState();
    _loadApprovalTimeline();
    // เริ่ม polling หลังโหลดครั้งแรก
    _startPolling();
  }

  @override
  void dispose() {
    // หยุด polling เมื่อปิด Dialog
    _stopPolling();
    super.dispose();
  }

  /// เริ่ม polling timer
  void _startPolling() {
    _pollingTimer = Timer.periodic(_pollingInterval, (_) {
      _refreshDataSilently();
    });
  }

  /// หยุด polling timer
  void _stopPolling() {
    _pollingTimer?.cancel();
    _pollingTimer = null;
  }

  /// รีเฟรชข้อมูลแบบเงียบ (ไม่แสดง loading indicator)
  Future<void> _refreshDataSilently() async {
    try {
      final results = await Future.wait([
        ApprovalApiService.getApprovalTimeline(widget.docNo),
        ApprovalApiService.getPOApprovalStatus(widget.docNo),
      ]);

      final timelineResult = results[0] as ApprovalTimelineResult;
      final statusResult = results[1] as POApprovalStatusResult;

      if (!mounted) return;

      if (timelineResult.isSuccess) {
        setState(() {
          _timeline = timelineResult.timeline;
          _timelineStatus = timelineResult.status;
          if (statusResult.isSuccess && statusResult.isFound && statusResult.data != null) {
            _status = statusResult.data;
          }
          _errorMessage = null;
        });
      }
    } catch (_) {
      // ignore errors during silent refresh
    }
  }

  /// โหลด Timeline แบบละเอียด (รวมประวัติการแจ้งเตือน)
  Future<void> _loadApprovalTimeline() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // โหลด timeline และ status พร้อมกัน
      final results = await Future.wait([
        ApprovalApiService.getApprovalTimeline(widget.docNo),
        ApprovalApiService.getPOApprovalStatus(widget.docNo),
      ]);

      final timelineResult = results[0] as ApprovalTimelineResult;
      final statusResult = results[1] as POApprovalStatusResult;

      if (timelineResult.isSuccess) {
        setState(() {
          _timeline = timelineResult.timeline;
          _timelineStatus = timelineResult.status; // เก็บ status จาก timeline API
          if (statusResult.isSuccess && statusResult.isFound && statusResult.data != null) {
            _status = statusResult.data;
          }
          _isLoading = false;
        });
      } else if (statusResult.isSuccess && statusResult.isFound && statusResult.data != null) {
        // ถ้า timeline ล้มเหลวแต่ status สำเร็จ ใช้ history จาก status แทน
        setState(() {
          _status = statusResult.data;
          _isLoading = false;
        });
      } else {
        setState(() {
          _errorMessage = global.language('no_approval_data_for_document');
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = 'เกิดข้อผิดพลาด: $e';
        _isLoading = false;
      });
    }
  }

  /// แปลงเวลา UTC จากฐานข้อมูล เป็นเวลา local ของ client
  String _formatDateTime(DateTime dateTime) {
    // เวลาจากฐานข้อมูลเป็น UTC ต้องแปลงเป็น local timezone ของ client
    final localDateTime = dateTime.toLocal();
    final formatter = DateFormat('dd/MM/yyyy HH:mm', 'th');
    return formatter.format(localDateTime);
  }

  Color _getStatusColor(POApprovalStatus status) {
    switch (status) {
      case POApprovalStatus.pending:
        return Colors.orange;
      case POApprovalStatus.approved:
        return Colors.green;
      case POApprovalStatus.rejected:
        return Colors.red;
      case POApprovalStatus.autoApproved:
        return Colors.blue;
      case POApprovalStatus.draft:
        return Colors.grey;
    }
  }

  String _getStatusText(POApprovalStatus status) {
    switch (status) {
      case POApprovalStatus.pending:
        return global.language("approval_pending");
      case POApprovalStatus.approved:
        return global.language("approval_approved");
      case POApprovalStatus.rejected:
        return global.language("approval_rejected");
      case POApprovalStatus.autoApproved:
        return global.language("approval_auto_approved");
      case POApprovalStatus.draft:
        return global.language("draft");
    }
  }

  IconData _getActionIcon(String action) {
    switch (action) {
      case 'submit':
        return Icons.send;
      case 'approve':
        return Icons.check_circle;
      case 'reject':
        return Icons.cancel;
      default:
        return Icons.history;
    }
  }

  Color _getActionColor(String action) {
    switch (action) {
      case 'submit':
        return Colors.blue;
      case 'approve':
        return Colors.green;
      case 'reject':
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  String _getActionText(String action) {
    switch (action) {
      case 'submit':
        return global.language("submit_for_approval");
      case 'approve':
        return global.language("approve");
      case 'reject':
        return global.language("approval_rejected");
      default:
        return action;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        constraints: BoxConstraints(maxWidth: 500, maxHeight: 600),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Header
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.blue.shade600,
                borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(16),
                  topRight: Radius.circular(16),
                ),
              ),
              child: Row(
                children: [
                  const Icon(Icons.history, color: Colors.white),
                  SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language("approval_history"),
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Text(
                          'เลขที่: ${widget.docNo}',
                          style: TextStyle(
                            color: Colors.white.withValues(alpha: 0.9),
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.white),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
            ),

            // Content
            Flexible(
              child: _isLoading
                  ? const Center(
                      child: Padding(
                        padding: EdgeInsets.all(32),
                        child: CircularProgressIndicator(),
                      ),
                    )
                  : _errorMessage != null
                      ? Center(
                          child: Padding(
                            padding: const EdgeInsets.all(32),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(Icons.info_outline, size: 48, color: Colors.grey.shade400),
                                const SizedBox(height: 16),
                                Text(
                                  _errorMessage!,
                                  textAlign: TextAlign.center,
                                  style: TextStyle(color: Colors.grey.shade600),
                                ),
                              ],
                            ),
                          ),
                        )
                      : _buildContent(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContent() {
    // ถ้าไม่มี status และไม่มี timeline ให้แสดง empty
    if (_status == null && _timeline.isEmpty && _timelineStatus == null) {
      return const SizedBox.shrink();
    }

    return SingleChildScrollView(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // สถานะปัจจุบัน (ใช้ _status ถ้ามี ไม่งั้นใช้ _timelineStatus)
          if (_status != null) ...[
            _buildStatusCard(),
            const SizedBox(height: 16),
            _buildInfoCard(),
            const SizedBox(height: 16),
          ] else if (_timelineStatus != null) ...[
            _buildTimelineStatusCard(),
            const SizedBox(height: 16),
            _buildTimelineInfoCard(),
            const SizedBox(height: 16),
          ],

          // ประวัติ Timeline (ใช้ _timeline ถ้ามี ไม่งั้นใช้ _status.history)
          if (_timeline.isNotEmpty || (_status != null && _status!.history.isNotEmpty)) ...[
            Text(
              global.language("action_history"),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            _timeline.isNotEmpty ? _buildDetailedTimeline() : _buildTimeline(),
          ],

          // หมายเหตุล่าสุด (จาก _status หรือ _timelineStatus)
          ..._buildLastCommentSection(),
        ],
      ),
    );
  }

  /// สร้างส่วนหมายเหตุล่าสุด
  List<Widget> _buildLastCommentSection() {
    String? lastComment;
    if (_status != null && _status!.lastComment != null && _status!.lastComment!.isNotEmpty) {
      lastComment = _status!.lastComment;
    } else if (_timelineStatus != null && _timelineStatus!['last_comment'] != null) {
      final comment = _timelineStatus!['last_comment']?.toString();
      if (comment != null && comment.isNotEmpty) {
        lastComment = comment;
      }
    }

    if (lastComment == null) return [];
    return [
      const SizedBox(height: 16),
      _buildCommentCardWithText(lastComment),
    ];
  }

  /// สร้างการ์ดหมายเหตุพร้อมข้อความ
  Widget _buildCommentCardWithText(String comment) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.orange.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.orange.shade200),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(Icons.comment, color: Colors.orange.shade600, size: 20),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language("latest_remark"),
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.orange.shade700,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  comment,
                  style: TextStyle(
                    color: Colors.grey.shade800,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// สร้างการ์ดสถานะจาก timeline API
  Widget _buildTimelineStatusCard() {
    final statusStr = _timelineStatus?['status']?.toString() ?? 'pending';
    final color = _getStatusColorFromString(statusStr);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: color,
              shape: BoxShape.circle,
            ),
            child: Icon(
              _getStatusIconFromString(statusStr),
              color: Colors.white,
              size: 24,
            ),
          ),
          SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language("current_status"),
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey.shade600,
                  ),
                ),
                Text(
                  _getStatusTextFromString(statusStr),
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: color,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// สร้างการ์ดข้อมูลจาก timeline API
  Widget _buildTimelineInfoCard() {
    final totalAmount = _timelineStatus?['total_amount'];
    final purchaseTypeName = _timelineStatus?['purchase_type_name']?.toString() ?? '-';
    final createdByName = _timelineStatus?['created_by_name']?.toString() ?? '-';
    final requiredLevel = _timelineStatus?['required_level'] ?? 0;
    final requiredLevelName = _timelineStatus?['required_level_name']?.toString() ?? '';
    final currentApprovedLevel = _timelineStatus?['current_approved_level'] ?? 0;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        children: [
          _buildInfoRow(global.language("purchase_type"), purchaseTypeName),
          Divider(height: 16),
          _buildInfoRow(global.language("document_creator"), createdByName),
          Divider(height: 16),
          _buildInfoRow(global.language("total_amount"), totalAmount != null
              ? '฿${NumberFormat('#,##0.00').format(totalAmount)}'
              : '-'),
          Divider(height: 16),
          _buildInfoRow(global.language("required_approval_level"), requiredLevelName.isNotEmpty
              ? requiredLevelName
              : '${global.language("level")} $requiredLevel'),
          if (currentApprovedLevel > 0) ...[
            Divider(height: 16),
            _buildInfoRow(global.language("approved_up_to_level"), '${global.language("level")} $currentApprovedLevel'),
          ],
        ],
      ),
    );
  }

  /// ดึงสีจาก status string
  Color _getStatusColorFromString(String status) {
    switch (status) {
      case 'pending':
        return Colors.orange;
      case 'approved':
        return Colors.green;
      case 'rejected':
        return Colors.red;
      case 'auto_approved':
        return Colors.blue;
      default:
        return Colors.grey;
    }
  }

  /// ดึง icon จาก status string
  IconData _getStatusIconFromString(String status) {
    switch (status) {
      case 'approved':
      case 'auto_approved':
        return Icons.check;
      case 'rejected':
        return Icons.close;
      case 'pending':
        return Icons.hourglass_empty;
      default:
        return Icons.help_outline;
    }
  }

  /// ดึงข้อความจาก status string
  String _getStatusTextFromString(String status) {
    switch (status) {
      case 'pending':
        return global.language("approval_pending");
      case 'approved':
        return global.language("approval_approved");
      case 'rejected':
        return global.language("approval_rejected");
      case 'auto_approved':
        return global.language("approval_auto_approved");
      default:
        return status;
    }
  }

  Widget _buildStatusCard() {
    final status = _status!.status;
    final color = _getStatusColor(status);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: color,
              shape: BoxShape.circle,
            ),
            child: Icon(
              status == POApprovalStatus.approved || status == POApprovalStatus.autoApproved
                  ? Icons.check
                  : status == POApprovalStatus.rejected
                      ? Icons.close
                      : Icons.hourglass_empty,
              color: Colors.white,
              size: 24,
            ),
          ),
          SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language("current_status"),
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey.shade600,
                  ),
                ),
                Text(
                  _getStatusText(status),
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: color,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        children: [
          _buildInfoRow(global.language("purchase_type"), _status!.purchaseTypeName),
          Divider(height: 16),
          _buildInfoRow(global.language("document_creator"), _status!.createdByName),
          Divider(height: 16),
          _buildInfoRow(global.language("total_amount"), '฿${NumberFormat('#,##0.00').format(_status!.totalAmount)}'),
          Divider(height: 16),
          _buildInfoRow(global.language("required_approval_level"), _status!.requiredLevelName.isNotEmpty ? _status!.requiredLevelName : '${global.language("level")} ${_status!.requiredLevel}'),
        ],
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: TextStyle(
            color: Colors.grey.shade600,
            fontSize: 14,
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            fontWeight: FontWeight.w500,
            fontSize: 14,
          ),
        ),
      ],
    );
  }

  Widget _buildTimeline() {
    final history = _status!.history.reversed.toList(); // แสดงล่าสุดก่อน

    return Column(
      children: history.asMap().entries.map((entry) {
        final index = entry.key;
        final item = entry.value;
        final isLast = index == history.length - 1;
        final color = _getActionColor(item.action);

        return IntrinsicHeight(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Timeline indicator
              Column(
                children: [
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: color,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      _getActionIcon(item.action),
                      color: Colors.white,
                      size: 16,
                    ),
                  ),
                  if (!isLast)
                    Expanded(
                      child: Container(
                        width: 2,
                        color: Colors.grey.shade300,
                      ),
                    ),
                ],
              ),
              const SizedBox(width: 12),
              // Content
              Expanded(
                child: Container(
                  margin: EdgeInsets.only(bottom: isLast ? 0 : 16),
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.grey.shade200),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            _getActionText(item.action),
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              color: color,
                            ),
                          ),
                          Text(
                            _formatDateTime(item.actionAt),
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade500,
                            ),
                          ),
                        ],
                      ),
                      SizedBox(height: 4),
                      Text(
                        '${global.language("by")}: ${item.actionByName}',
                        style: TextStyle(
                          fontSize: 13,
                          color: Colors.grey.shade700,
                        ),
                      ),
                      if (item.comment != null && item.comment!.isNotEmpty) ...[
                        SizedBox(height: 4),
                        Text(
                          '${global.language("remark")}: ${item.comment}',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                            fontStyle: FontStyle.italic,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      }).toList(),
    );
  }

  /// สร้าง Timeline แบบ Step (แสดงเป็น step-by-step progress)
  Widget _buildDetailedTimeline() {
    // แสดงจากเก่าไปใหม่ (submit -> notification -> opened -> approve)
    final items = _timeline;

    return Container(
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        children: items.asMap().entries.map((entry) {
          final index = entry.key;
          final item = entry.value;
          final isLast = index == items.length - 1;
          final isFirst = index == 0;
          final color = _getTimelineActionColor(item.action);
          final emoji = _getTimelineActionEmoji(item.action);

          return Container(
            decoration: BoxDecoration(
              border: isLast
                  ? null
                  : Border(
                      bottom: BorderSide(color: Colors.grey.shade200, width: 1),
                    ),
            ),
            child: Padding(
              padding: EdgeInsets.only(
                left: 16,
                right: 16,
                top: isFirst ? 12 : 10,
                bottom: isLast ? 12 : 10,
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  // Step indicator (emoji)
                  Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: color.withValues(alpha: 0.15),
                      shape: BoxShape.circle,
                      border: Border.all(color: color, width: 2),
                    ),
                    child: Center(
                      child: Text(
                        emoji,
                        style: const TextStyle(fontSize: 16),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  // Step content
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Action name with checkmark
                        Row(
                          children: [
                            Text(
                              _getStepDisplayText(item),
                              style: TextStyle(
                                fontWeight: FontWeight.w600,
                                color: Colors.grey.shade800,
                                fontSize: 14,
                              ),
                            ),
                            const SizedBox(width: 4),
                            Icon(
                              Icons.check_circle,
                              color: color,
                              size: 16,
                            ),
                          ],
                        ),
                        // Sub-detail
                        Text(
                          _getStepSubText(item),
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Timestamp
                  Text(
                    _formatDateTime(item.actionAt),
                    style: TextStyle(
                      fontSize: 11,
                      color: Colors.grey.shade500,
                    ),
                  ),
                ],
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  /// ดึง emoji สำหรับ Timeline action
  String _getTimelineActionEmoji(String action) {
    switch (action) {
      case 'submit':
        return '📤';
      case 'modify':
        return '✏️';
      case 'approve':
      case 'auto_approve':
        return '✅';
      case 'reject':
        return '❌';
      case 'notification_sent':
        return '🔔';
      case 'notification_opened':
        return '👁️';
      default:
        return '📋';
    }
  }

  /// ดึงข้อความหลักสำหรับ Step
  String _getStepDisplayText(ApprovalTimelineItem item) {
    switch (item.action) {
      case 'submit':
        return global.language("document_submitted");
      case 'modify':
        return global.language("document_modified");
      case 'approve':
        return global.language("approval_approved");
      case 'auto_approve':
        return global.language("approval_auto_approved");
      case 'reject':
        return global.language("approval_rejected");
      case 'notification_sent':
        return global.language("notification_sent");
      case 'notification_opened':
        return global.language("notification_opened");
      default:
        return item.actionDisplayName;
    }
  }

  /// ดึงข้อความรองสำหรับ Step
  String _getStepSubText(ApprovalTimelineItem item) {
    switch (item.action) {
      case 'submit':
        return '${global.language("by")}: ${item.actionByName}';
      case 'modify':
        return '${global.language("by")}: ${item.actionByName}';
      case 'approve':
      case 'auto_approve':
        return '${global.language("by")}: ${item.actionByName}${item.comment != null && item.comment!.isNotEmpty ? ' - ${item.comment}' : ''}';
      case 'reject':
        return '${global.language("by")}: ${item.actionByName}${item.comment != null && item.comment!.isNotEmpty ? ' - ${item.comment}' : ''}';
      case 'notification_sent':
        final channel = item.type == 'email' ? 'Email' : 'LINE';
        return '${global.language("to")}: ${item.actionByName} ($channel)';
      case 'notification_opened':
        final source = item.type == 'email' ? 'Email' : item.type == 'line' ? 'LINE' : 'LIFF';
        return '${global.language("by")}: ${item.actionByName} (${global.language("from")} $source)';
      default:
        return '${global.language("by")}: ${item.actionByName}';
    }
  }

  /// ดึงสีสำหรับ Timeline action (รวม notification)
  Color _getTimelineActionColor(String action) {
    switch (action) {
      case 'submit':
        return Colors.blue;
      case 'modify':
        return Colors.orange;
      case 'approve':
      case 'auto_approve':
        return Colors.green;
      case 'reject':
        return Colors.red;
      case 'notification_sent':
        return Colors.purple;
      case 'notification_opened':
        return Colors.teal;
      default:
        return Colors.grey;
    }
  }

}
