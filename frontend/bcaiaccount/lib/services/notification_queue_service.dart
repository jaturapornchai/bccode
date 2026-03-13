import 'package:smlaicloud/global.dart' as global;
import 'dart:async';
import 'package:flutter/material.dart';

/// ประเภทของ Notification
enum NotificationType {
  info,
  success,
  warning,
  error,
  approval, // สำหรับการอนุมัติ PO
}

/// Model สำหรับ Notification
class NotificationItem {
  final String id;
  final String title;
  final String message;
  final NotificationType type;
  final DateTime createdAt;
  final VoidCallback? onTap;
  final Duration duration;
  final Map<String, dynamic>? data; // ข้อมูลเพิ่มเติม เช่น docno, guidfixed

  NotificationItem({
    String? id,
    required this.title,
    required this.message,
    this.type = NotificationType.info,
    this.onTap,
    this.duration = const Duration(seconds: 5),
    this.data,
  })  : id = id ?? DateTime.now().millisecondsSinceEpoch.toString(),
        createdAt = DateTime.now();

  /// สร้าง Notification สำหรับการอนุมัติ PO
  factory NotificationItem.poApproved({
    required String docNo,
    required String approverName,
    String? guidFixed,
    VoidCallback? onTap,
  }) {
    return NotificationItem(
      title: 'PO ได้รับการอนุมัติ',
      message: 'ใบสั่งซื้อ $docNo ได้รับการอนุมัติโดย $approverName',
      type: NotificationType.approval,
      duration: const Duration(seconds: 8),
      onTap: onTap,
      data: {
        'docno': docNo,
        'guidfixed': guidFixed,
        'approver_name': approverName,
        'action': 'approved',
      },
    );
  }

  /// สร้าง Notification สำหรับการปฏิเสธ PO
  factory NotificationItem.poRejected({
    required String docNo,
    required String rejecterName,
    String? reason,
    String? guidFixed,
    VoidCallback? onTap,
  }) {
    return NotificationItem(
      title: 'PO ถูกปฏิเสธ',
      message: 'ใบสั่งซื้อ $docNo ถูกปฏิเสธโดย $rejecterName${reason != null ? '\nเหตุผล: $reason' : ''}',
      type: NotificationType.warning,
      duration: const Duration(seconds: 10),
      onTap: onTap,
      data: {
        'docno': docNo,
        'guidfixed': guidFixed,
        'rejecter_name': rejecterName,
        'reason': reason,
        'action': 'rejected',
      },
    );
  }
}

/// Global Notification Queue Service (Singleton)
/// ใช้สำหรับแสดง notification ทั่วทั้ง app
class NotificationQueueService {
  // Singleton instance
  static final NotificationQueueService _instance = NotificationQueueService._internal();
  factory NotificationQueueService() => _instance;
  NotificationQueueService._internal();

  // Queue ของ notifications ที่รอแสดง
  final List<NotificationItem> _queue = [];

  // Notification ที่กำลังแสดงอยู่
  NotificationItem? _currentNotification;

  // Stream controller สำหรับ broadcast การเปลี่ยนแปลง
  final _notificationController = StreamController<NotificationItem?>.broadcast();

  // Stream สำหรับ listen การเปลี่ยนแปลง
  Stream<NotificationItem?> get notificationStream => _notificationController.stream;

  // Getter สำหรับ notification ปัจจุบัน
  NotificationItem? get currentNotification => _currentNotification;

  // Timer สำหรับ auto-dismiss
  Timer? _dismissTimer;

  // GlobalKey สำหรับ overlay
  final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();
  OverlayEntry? _overlayEntry;

  /// เพิ่ม notification เข้า queue
  void add(NotificationItem notification) {
    _queue.add(notification);
    // ถ้าไม่มี notification ที่กำลังแสดงอยู่ ให้แสดงเลย
    if (_currentNotification == null) {
      _showNext();
    }
  }

  /// เพิ่ม notification แบบง่าย
  void addSimple({
    required String title,
    required String message,
    NotificationType type = NotificationType.info,
    Duration duration = const Duration(seconds: 5),
    VoidCallback? onTap,
  }) {
    add(NotificationItem(
      title: title,
      message: message,
      type: type,
      duration: duration,
      onTap: onTap,
    ));
  }

  /// แสดง notification ถัดไปใน queue
  void _showNext() {
    if (_queue.isEmpty) {
      _currentNotification = null;
      _notificationController.add(null);
      _hideOverlay();
      return;
    }

    _currentNotification = _queue.removeAt(0);
    _notificationController.add(_currentNotification);

    // ตั้ง timer สำหรับ auto-dismiss
    _dismissTimer?.cancel();
    _dismissTimer = Timer(_currentNotification!.duration, () {
      dismiss();
    });
  }

  /// ปิด notification ปัจจุบัน
  void dismiss() {
    _dismissTimer?.cancel();
    _showNext();
  }

  /// ล้าง queue ทั้งหมด
  void clearAll() {
    _queue.clear();
    _dismissTimer?.cancel();
    _currentNotification = null;
    _notificationController.add(null);
    _hideOverlay();
  }

  /// แสดง Overlay
  void showOverlay(BuildContext context) {
    _hideOverlay();
    _overlayEntry = OverlayEntry(
      builder: (context) => _NotificationOverlay(service: this),
    );
    Overlay.of(context).insert(_overlayEntry!);
  }

  /// ซ่อน Overlay
  void _hideOverlay() {
    _overlayEntry?.remove();
    _overlayEntry = null;
  }

  /// Dispose
  void dispose() {
    _dismissTimer?.cancel();
    _notificationController.close();
    _hideOverlay();
  }
}

/// Widget สำหรับแสดง Notification Overlay
class _NotificationOverlay extends StatefulWidget {
  final NotificationQueueService service;

  const _NotificationOverlay({required this.service});

  @override
  State<_NotificationOverlay> createState() => _NotificationOverlayState();
}

class _NotificationOverlayState extends State<_NotificationOverlay>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<Offset> _slideAnimation;
  late Animation<double> _fadeAnimation;
  StreamSubscription<NotificationItem?>? _subscription;

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );
    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, -1),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOutCubic,
    ));
    _fadeAnimation = Tween<double>(
      begin: 0,
      end: 1,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOut,
    ));

    _subscription = widget.service.notificationStream.listen((notification) {
      if (notification != null) {
        _animationController.forward(from: 0);
      } else {
        _animationController.reverse();
      }
      if (mounted) setState(() {});
    });

    // แสดง animation ถ้ามี notification อยู่แล้ว
    if (widget.service.currentNotification != null) {
      _animationController.forward();
    }
  }

  @override
  void dispose() {
    _subscription?.cancel();
    _animationController.dispose();
    super.dispose();
  }

  Color _getBackgroundColor(NotificationType type) {
    switch (type) {
      case NotificationType.success:
        return Colors.green.shade600;
      case NotificationType.warning:
        return Colors.orange.shade600;
      case NotificationType.error:
        return Colors.red.shade600;
      case NotificationType.approval:
        return Colors.blue.shade600;
      case NotificationType.info:
        return Colors.blueGrey.shade700;
    }
  }

  IconData _getIcon(NotificationType type) {
    switch (type) {
      case NotificationType.success:
        return Icons.check_circle;
      case NotificationType.warning:
        return Icons.warning;
      case NotificationType.error:
        return Icons.error;
      case NotificationType.approval:
        return Icons.approval;
      case NotificationType.info:
        return Icons.info;
    }
  }

  @override
  Widget build(BuildContext context) {
    final notification = widget.service.currentNotification;
    if (notification == null) return const SizedBox.shrink();

    return Positioned(
      top: MediaQuery.of(context).padding.top + 8,
      left: 8,
      right: 8,
      child: SlideTransition(
        position: _slideAnimation,
        child: FadeTransition(
          opacity: _fadeAnimation,
          child: Material(
            elevation: 8,
            borderRadius: BorderRadius.circular(12),
            color: _getBackgroundColor(notification.type),
            child: InkWell(
              onTap: () {
                notification.onTap?.call();
                widget.service.dismiss();
              },
              borderRadius: BorderRadius.circular(12),
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Row(
                  children: [
                    Icon(
                      _getIcon(notification.type),
                      color: global.theme.cardColor,
                      size: 28,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            notification.title,
                            style: TextStyle(
                              color: global.theme.onPrimaryColor,
                              fontWeight: FontWeight.bold,
                              fontSize: 14,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            notification.message,
                            style: TextStyle(
                              color: global.theme.onPrimaryColor.withValues(alpha: 0.9),
                              fontSize: 12,
                            ),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: Icon(Icons.close, color: global.theme.onPrimaryColor, size: 20),
                      onPressed: () => widget.service.dismiss(),
                      padding: EdgeInsets.zero,
                      constraints: const BoxConstraints(
                        minWidth: 32,
                        minHeight: 32,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// Widget สำหรับ wrap app และแสดง notification
/// ใช้: NotificationWrapper(child: YourApp())
class NotificationWrapper extends StatefulWidget {
  final Widget child;

  const NotificationWrapper({super.key, required this.child});

  @override
  State<NotificationWrapper> createState() => _NotificationWrapperState();
}

class _NotificationWrapperState extends State<NotificationWrapper> {
  @override
  void initState() {
    super.initState();
    // แสดง overlay หลังจาก build เสร็จ
    WidgetsBinding.instance.addPostFrameCallback((_) {
      NotificationQueueService().showOverlay(context);
    });
  }

  @override
  Widget build(BuildContext context) {
    return widget.child;
  }
}
