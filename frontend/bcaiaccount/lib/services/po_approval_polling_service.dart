import 'dart:async';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/services/notification_queue_service.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Service สำหรับ poll สถานะการอนุมัติ PO และแจ้งเตือนเมื่อมีการเปลี่ยนแปลง
class POApprovalPollingService {
  // Singleton instance
  static final POApprovalPollingService _instance = POApprovalPollingService._internal();
  factory POApprovalPollingService() => _instance;
  POApprovalPollingService._internal();

  // Cache สถานะก่อนหน้าของแต่ละ PO
  // Key: docno, Value: status string (pending, approved, rejected, etc.)
  final Map<String, String> _previousStatusCache = {};

  // รายการ PO ที่กำลังติดตาม (pending status)
  final Set<String> _watchingDocs = {};

  // Timer สำหรับ polling
  Timer? _pollingTimer;

  // Interval สำหรับ polling (วินาที)
  int _pollingIntervalSeconds = 30;

  // Callback เมื่อสถานะเปลี่ยน
  void Function(String docNo, String oldStatus, String newStatus, String? approverName)? onStatusChanged;

  /// เริ่ม polling service
  void startPolling({int intervalSeconds = 30}) {
    // ป้องกันการสร้าง timer ซ้อนกัน
    if (_pollingTimer != null) {
      AppLogger.debug('[POApprovalPolling] Already running, skipping start');
      return;
    }
    _pollingIntervalSeconds = intervalSeconds;
    _pollingTimer = Timer.periodic(
      Duration(seconds: _pollingIntervalSeconds),
      (_) => _pollWatchingDocs(),
    );
    AppLogger.info('[POApprovalPolling] Started with interval: ${_pollingIntervalSeconds}s');
  }

  /// หยุด polling service
  void stopPolling() {
    _stopPollingTimer();
    AppLogger.info('[POApprovalPolling] Stopped');
  }

  void _stopPollingTimer() {
    _pollingTimer?.cancel();
    _pollingTimer = null;
  }

  /// เพิ่ม PO เข้ารายการติดตาม
  void watchDoc(String docNo, {String? initialStatus}) {
    if (docNo.isEmpty) return;
    _watchingDocs.add(docNo);
    if (initialStatus != null) {
      _previousStatusCache[docNo] = initialStatus;
    }
    AppLogger.debug('[POApprovalPolling] Watching doc: $docNo (initial: $initialStatus)');
  }

  /// ลบ PO ออกจากรายการติดตาม
  void unwatchDoc(String docNo) {
    _watchingDocs.remove(docNo);
    _previousStatusCache.remove(docNo);
    AppLogger.debug('[POApprovalPolling] Unwatched doc: $docNo');
  }

  /// ล้างรายการติดตามทั้งหมด
  void clearAllWatching() {
    _watchingDocs.clear();
    _previousStatusCache.clear();
    AppLogger.debug('[POApprovalPolling] Cleared all watching docs');
  }

  /// ตรวจสอบสถานะ PO ทั้งหมดที่กำลังติดตาม
  Future<void> _pollWatchingDocs() async {
    if (_watchingDocs.isEmpty) return;

    try {
      final docNos = _watchingDocs.toList();
      AppLogger.debug('[POApprovalPolling] Polling ${docNos.length} docs');

      // ดึงสถานะแบบ batch
      final statuses = await ApprovalApiService.getBatchPOApprovalStatus(docNos);

      // ตรวจสอบการเปลี่ยนแปลง
      for (final docNo in docNos) {
        final statusInfo = statuses[docNo];
        if (statusInfo == null) continue;

        final newStatus = statusInfo.status;
        final oldStatus = _previousStatusCache[docNo];

        // ถ้าสถานะเปลี่ยน
        if (oldStatus != null && oldStatus != newStatus) {
          AppLogger.info('[POApprovalPolling] Status changed: $docNo ($oldStatus -> $newStatus)');

          // ดึงข้อมูลละเอียดเพื่อหาชื่อผู้อนุมัติ
          String? approverName;
          if (newStatus == 'approved' || newStatus == 'rejected') {
            final detailResult = await ApprovalApiService.getPOApprovalStatus(docNo);
            if (detailResult.isSuccess && detailResult.data != null) {
              // หา history entry ล่าสุดที่เป็น approve/reject
              final history = detailResult.data!.history;
              if (history.isNotEmpty) {
                final lastAction = history.last;
                if (lastAction.action == 'approve' || lastAction.action == 'reject') {
                  approverName = lastAction.actionByName;
                }
              }
            }
          }

          // แจ้งเตือน
          _notifyStatusChange(docNo, oldStatus, newStatus, approverName);

          // Callback
          onStatusChanged?.call(docNo, oldStatus, newStatus, approverName);

          // ถ้าสถานะเป็น approved หรือ rejected ให้หยุดติดตาม
          if (newStatus == 'approved' || newStatus == 'rejected' || newStatus == 'auto_approved') {
            _watchingDocs.remove(docNo);
          }
        }

        // อัพเดท cache
        _previousStatusCache[docNo] = newStatus;
      }
    } catch (e) {
      AppLogger.error('[POApprovalPolling] Error polling: $e');
    }
  }

  /// แจ้งเตือนเมื่อสถานะเปลี่ยน
  void _notifyStatusChange(String docNo, String oldStatus, String newStatus, String? approverName) {
    if (newStatus == 'approved' || newStatus == 'auto_approved') {
      global.notificationQueue.add(NotificationItem.poApproved(
        docNo: docNo,
        approverName: approverName ?? 'ระบบ',
      ));
    } else if (newStatus == 'rejected') {
      global.notificationQueue.add(NotificationItem.poRejected(
        docNo: docNo,
        rejecterName: approverName ?? global.language('point_transaction.unknown'),
      ));
    }
  }

  /// ตรวจสอบสถานะทันที (ไม่ต้องรอ polling interval)
  Future<void> checkNow() async {
    await _pollWatchingDocs();
  }

  /// ตรวจสอบและอัพเดทสถานะของ PO เฉพาะตัว
  /// คืนค่า true ถ้าสถานะเปลี่ยนจาก pending -> approved/rejected
  Future<bool> checkAndNotify(String docNo) async {
    if (docNo.isEmpty) return false;

    try {
      final result = await ApprovalApiService.getPOApprovalStatus(docNo);
      if (!result.isSuccess || result.data == null) return false;

      final newStatus = result.data!.status.name;
      final oldStatus = _previousStatusCache[docNo];

      // อัพเดท cache
      _previousStatusCache[docNo] = newStatus;

      // ถ้าสถานะเปลี่ยน
      if (oldStatus != null && oldStatus != newStatus) {
        String? approverName;
        final history = result.data!.history;
        if (history.isNotEmpty) {
          final lastAction = history.last;
          if (lastAction.action == 'approve' || lastAction.action == 'reject') {
            approverName = lastAction.actionByName;
          }
        }

        _notifyStatusChange(docNo, oldStatus, newStatus, approverName);
        return true;
      }

      return false;
    } catch (e) {
      AppLogger.error('[POApprovalPolling] Error checking doc $docNo: $e');
      return false;
    }
  }

  /// ดึงสถานะปัจจุบันจาก cache
  String? getCachedStatus(String docNo) => _previousStatusCache[docNo];

  /// จำนวน PO ที่กำลังติดตาม
  int get watchingCount => _watchingDocs.length;

  /// รายการ PO ที่กำลังติดตาม
  List<String> get watchingDocs => _watchingDocs.toList();

  /// Dispose
  void dispose() {
    _stopPollingTimer();
    _watchingDocs.clear();
    _previousStatusCache.clear();
  }
}

/// Global instance
final poApprovalPolling = POApprovalPollingService();
