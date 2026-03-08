import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'dart:async';
import 'dart:math';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/alert_agent_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class AlertAgentScreen extends StatefulWidget {
  const AlertAgentScreen({super.key});

  @override
  State<AlertAgentScreen> createState() => _AlertAgentScreenState();
}

class _AlertAgentScreenState extends State<AlertAgentScreen> {
  List<AlertAgentModel> _alerts = [];
  bool _isLoading = false;
  AlertAgentModel? _selectedAlert; // Alert ที่เลือก
  bool _isEditing = false; // กำลังแก้ไขอยู่หรือไม่

  final Color primaryColor = const Color(0xFF2E7D32);
  static const String mongoCollection = 'email';

  String get apiBaseUrl {
    String apiPath = global.myAppConfig.serviceApi;
    if (!apiPath.startsWith('http')) {
      apiPath = 'https://$apiPath';
    }
    if (apiPath.endsWith('/')) {
      apiPath = apiPath.substring(0, apiPath.length - 1);
    }
    if (global.myAppConfig.servicePort.isNotEmpty) {
      return '$apiPath:${global.myAppConfig.servicePort}';
    }
    return apiPath;
  }

  @override
  void initState() {
    super.initState();
    _loadAlerts();
  }

  Future<void> _loadAlerts() async {
    setState(() => _isLoading = true);

    try {
      final shopId = global.getShopId();
      final url = Uri.parse('$apiBaseUrl/atlas/get');
      final requestBody = {'collection': mongoCollection, 'shopid': shopId};

      if (kDebugMode) {
        AppLogger.debug('Loading alerts from: $url');
        AppLogger.debug('Request body: ${jsonEncode(requestBody)}');
      }

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (kDebugMode) {
        AppLogger.debug('Response status: ${response.statusCode}');
        AppLogger.debug('Response body: ${response.body}');
      }

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List<dynamic> items = data['data'] ?? data['results'] ?? [];
        setState(() {
          _alerts = items.map((e) => AlertAgentModel.fromJson(e)).toList();
          // อัปเดต selected alert ถ้ามี
          if (_selectedAlert != null) {
            final found = _alerts.where((a) => a.id == _selectedAlert!.id).firstOrNull;
            _selectedAlert = found;
          }
        });
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error loading alerts: $e');
      }
      _showSnackBar('${global.language("error_with_detail")}: $e', isError: true);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  bool _hasValidChannel(AlertAgentModel alert) {
    final channels = alert.channels;
    final hasEmail = channels.email.isEnabled && channels.email.addresses.isNotEmpty;
    final hasLine = channels.line.isEnabled && channels.line.tokens.isNotEmpty;
    return hasEmail || hasLine;
  }

  // เพิ่ม Alert ใหม่
  void _addNewAlert() {
    setState(() {
      _selectedAlert = null;
      _isEditing = true;
    });
  }

  // เลือก Alert
  void _selectAlert(AlertAgentModel alert) {
    setState(() {
      _selectedAlert = alert;
      _isEditing = false;
    });
  }

  // แก้ไข Alert
  void _editSelectedAlert() {
    setState(() {
      _isEditing = true;
    });
  }

  // ยกเลิกการแก้ไข
  void _cancelEdit() {
    setState(() {
      if (_selectedAlert == null) {
        // ถ้ากำลังเพิ่มใหม่ ให้กลับไปหน้าว่าง
        _isEditing = false;
      } else {
        _isEditing = false;
      }
    });
  }

  // บันทึก Alert
  Future<void> _saveAlert(AlertAgentModel alert, {required bool isNew}) async {
    if (!_hasValidChannel(alert)) {
      _showSnackBar(global.language("alert_select_channel_required"), isError: true);
      return;
    }

    try {
      final url = Uri.parse('$apiBaseUrl/atlas/update');
      final taskName = isNew ? alert.taskName : (_selectedAlert?.taskName ?? alert.taskName);

      final requestBody = {'collection': mongoCollection, 'shopid': alert.shopId, 'email': taskName, 'cartid': taskName, 'data': alert.toJson(), if (isNew) 'upsert': true};

      if (kDebugMode) {
        AppLogger.debug('Saving alert: ${jsonEncode(requestBody)}');
      }

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (response.statusCode == 200) {
        _showSnackBar(isNew ? global.language("alert_create_success") : global.language("alert_save_success"));
        setState(() {
          _isEditing = false;
          _selectedAlert = alert;
        });
        _loadAlerts();
      } else {
        _showSnackBar('${global.language("error_with_detail")}: ${response.body}', isError: true);
      }
    } catch (e) {
      _showSnackBar('${global.language("error_with_detail")}: $e', isError: true);
    }
  }

  Future<void> _deleteAlert(AlertAgentModel alert) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.warning, color: Colors.orange),
            SizedBox(width: 8),
            Text(global.language("alert_confirm_delete")),
          ],
        ),
        content: Text(global.language("alert_confirm_delete_msg").replaceAll("{name}", alert.taskName)),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: Text(global.language("cancel"))),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      try {
        final url = Uri.parse('$apiBaseUrl/atlas/delete');
        final requestBody = {'collection': mongoCollection, 'shopid': alert.shopId, 'email': alert.taskName, 'cartid': alert.taskName};

        final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

        if (response.statusCode == 200) {
          _showSnackBar(global.language("alert_delete_success"));
          setState(() {
            _selectedAlert = null;
            _isEditing = false;
          });
          _loadAlerts();
        } else {
          _showSnackBar('${global.language("error_with_detail")}: ${response.body}', isError: true);
        }
      } catch (e) {
        _showSnackBar('${global.language("error_with_detail")}: $e', isError: true);
      }
    }
  }

  Future<void> _toggleActive(AlertAgentModel alert) async {
    try {
      final url = Uri.parse('$apiBaseUrl/atlas/update');
      final requestBody = {
        'collection': mongoCollection,
        'shopid': alert.shopId,
        'email': alert.taskName,
        'cartid': alert.taskName,
        'data': {'is_active': !alert.isActive},
      };

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (response.statusCode == 200) {
        _showSnackBar(alert.isActive ? global.language("alert_disabled_notification") : global.language("alert_enabled_notification"));
        _loadAlerts();
      }
    } catch (e) {
      _showSnackBar('${global.language("error_with_detail")}: $e', isError: true);
    }
  }

  void _showSnackBar(String message, {bool isError = false}) {
    if (isError) {
      global.showErrorSnackBar(context, message);
    } else {
      global.showSuccessSnackBar(context, message);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[100],
      appBar: AppBar(
        backgroundColor: primaryColor,
        title: Row(
          children: [
            const Icon(Icons.notifications_active, size: 24),
            const SizedBox(width: 8),
            Text('BC Alert Agent', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
          ],
        ),
        actions: [IconButton(icon: Icon(Icons.refresh), onPressed: _loadAlerts, tooltip: global.language("refresh"))],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Row(
              children: [
                // ===== ซ้าย: รายการ Alert =====
                SizedBox(width: 350, child: _buildLeftPanel()),
                // ===== เส้นแบ่ง =====
                Container(width: 1, color: Colors.grey[300]),
                // ===== ขวา: รายละเอียด/แก้ไข =====
                Expanded(child: _buildRightPanel()),
              ],
            ),
    );
  }

  // ===== Panel ซ้าย: รายการ Alert =====
  Widget _buildLeftPanel() {
    return Column(
      children: [
        // Header + ปุ่มเพิ่ม
        Container(
          padding: const EdgeInsets.all(12),
          color: Colors.white,
          child: Row(
            children: [
              Icon(Icons.list, color: primaryColor),
              SizedBox(width: 8),
              Text(
                global.language("alert_notification_list"),
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: primaryColor),
              ),
              Spacer(),
              ElevatedButton.icon(
                onPressed: _addNewAlert,
                icon: Icon(Icons.add, size: 18),
                label: Text(global.language("add")),
                style: ElevatedButton.styleFrom(backgroundColor: primaryColor, foregroundColor: Colors.white, padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8)),
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        // รายการ
        Expanded(
          child: _alerts.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.notifications_off, size: 48, color: Colors.grey[400]),
                      SizedBox(height: 8),
                      Text(global.language("alert_no_notifications"), style: TextStyle(color: Colors.grey[600])),
                    ],
                  ),
                )
              : ListView.builder(
                  itemCount: _alerts.length,
                  itemBuilder: (context, index) {
                    final alert = _alerts[index];
                    final isSelected = _selectedAlert?.id == alert.id;
                    return _buildAlertListItem(alert, isSelected);
                  },
                ),
        ),
      ],
    );
  }

  // รายการ Alert แต่ละตัว
  Widget _buildAlertListItem(AlertAgentModel alert, bool isSelected) {
    return InkWell(
      onTap: () => _selectAlert(alert),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: isSelected ? primaryColor.withValues(alpha: 0.1) : Colors.white,
          border: Border(
            left: BorderSide(color: isSelected ? primaryColor : Colors.transparent, width: 4),
            bottom: BorderSide(color: Colors.grey[200]!),
          ),
        ),
        child: Row(
          children: [
            // Status Icon
            Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(color: alert.isActive ? Colors.green.withValues(alpha: 0.1) : Colors.grey.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
              child: Icon(Icons.notifications_active, color: alert.isActive ? Colors.green : Colors.grey, size: 18),
            ),
            const SizedBox(width: 10),
            // Info
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    alert.taskName,
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: alert.isActive ? Colors.black87 : Colors.grey),
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 2),
                  Text(_getReportTypeLabel(alert.reportConfig.type), style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                ],
              ),
            ),
            // Status Badge
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(color: alert.isActive ? Colors.green : Colors.grey, borderRadius: BorderRadius.circular(4)),
              child: Text(
                alert.isActive ? 'ON' : 'OFF',
                style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Panel ขวา: รายละเอียด/แก้ไข =====
  Widget _buildRightPanel() {
    if (_isEditing) {
      // แสดงฟอร์มแก้ไข
      return AlertAgentEditForm(
        alert: _selectedAlert,
        shopId: global.getShopId(),
        onSave: (alert) => _saveAlert(alert, isNew: _selectedAlert == null),
        onCancel: _cancelEdit,
        primaryColor: primaryColor,
      );
    } else if (_selectedAlert != null) {
      // แสดงรายละเอียด
      return _buildDetailView(_selectedAlert!);
    } else {
      // ยังไม่เลือก
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.touch_app, size: 64, color: Colors.grey[300]),
            SizedBox(height: 16),
            Text(global.language("alert_select_from_left"), style: TextStyle(fontSize: 16, color: Colors.grey[500])),
            SizedBox(height: 8),
            Text(global.language("alert_or_add_new"), style: TextStyle(fontSize: 14, color: Colors.grey[400])),
          ],
        ),
      );
    }
  }

  // แสดงรายละเอียด Alert
  Widget _buildDetailView(AlertAgentModel alert) {
    return Column(
      children: [
        // Header + ปุ่ม
        Container(
          padding: const EdgeInsets.all(16),
          color: Colors.white,
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(color: alert.isActive ? Colors.green.withValues(alpha: 0.1) : Colors.grey.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)),
                child: Icon(Icons.notifications_active, color: alert.isActive ? Colors.green : Colors.grey, size: 24),
              ),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(alert.taskName, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    SizedBox(height: 4),
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                          decoration: BoxDecoration(color: alert.isActive ? Colors.green : Colors.grey, borderRadius: BorderRadius.circular(4)),
                          child: Text(alert.isActive ? global.language("alert_enabled") : global.language("alert_disabled"), style: TextStyle(color: Colors.white, fontSize: 11)),
                        ),
                        const SizedBox(width: 8),
                        Text(_getReportTypeLabel(alert.reportConfig.type), style: TextStyle(fontSize: 12, color: Colors.blue[700])),
                      ],
                    ),
                  ],
                ),
              ),
              // ปุ่ม Toggle
              IconButton(
                onPressed: () => _toggleActive(alert),
                icon: Icon(alert.isActive ? Icons.pause_circle : Icons.play_circle, size: 28),
                color: alert.isActive ? Colors.orange : Colors.green,
                tooltip: alert.isActive ? global.language("alert_disabled") : global.language("alert_enabled"),
              ),
              const SizedBox(width: 8),
              // ปุ่มแก้ไข
              ElevatedButton.icon(
                onPressed: _editSelectedAlert,
                icon: Icon(Icons.edit, size: 18),
                label: Text(global.language("edit")),
                style: ElevatedButton.styleFrom(backgroundColor: primaryColor, foregroundColor: Colors.white),
              ),
              const SizedBox(width: 8),
              // ปุ่มลบ
              ElevatedButton.icon(
                onPressed: () => _deleteAlert(alert),
                icon: Icon(Icons.delete, size: 18),
                label: Text(global.language("delete")),
                style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        // Content
        Expanded(
          child: SingleChildScrollView(
            padding: EdgeInsets.all(16),
            child: Column(
              children: [
                _buildDetailCard(global.language("alert_sending_settings"), Icons.schedule, [
                  _buildDetailRow(global.language("alert_schedule_format"), _formatScheduleMode(alert.schedule)),
                  if (alert.metadata.lastRunAt != null) _buildDetailRow(global.language("alert_last_run"), _formatDateTime(alert.metadata.lastRunAt)),
                ]),
                SizedBox(height: 16),
                _buildDetailCard(global.language("alert_notification_channel"), Icons.send, [
                  if (alert.channels.email.isEnabled) _buildChannelDetail('Email', Icons.email, Colors.blue, alert.channels.email.addresses),
                  if (alert.channels.line.isEnabled) _buildChannelDetail('LINE Notify', Icons.chat, Colors.green, alert.channels.line.tokens),
                  if (!alert.channels.email.isEnabled && !alert.channels.line.isEnabled) Text(global.language("alert_no_channel_set"), style: TextStyle(color: Colors.grey[500])),
                ]),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDetailCard(String title, IconData icon, List<Widget> children) {
    return Card(
      elevation: 1,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, color: primaryColor, size: 20),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: primaryColor),
                ),
              ],
            ),
            const Divider(height: 24),
            ...children,
          ],
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              '$label:',
              style: TextStyle(color: Colors.grey[600], fontWeight: FontWeight.w500),
            ),
          ),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }

  Widget _buildChannelDetail(String title, IconData icon, Color color, List<String> items) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: color, size: 18),
              const SizedBox(width: 8),
              Text(title, style: const TextStyle(fontWeight: FontWeight.bold)),
            ],
          ),
          const SizedBox(height: 4),
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: items
                .map(
                  (item) => Chip(
                    label: Text(item.length > 30 ? '${item.substring(0, 30)}...' : item, style: const TextStyle(fontSize: 12)),
                    backgroundColor: color.withValues(alpha: 0.1),
                    padding: EdgeInsets.zero,
                    visualDensity: VisualDensity.compact,
                  ),
                )
                .toList(),
          ),
        ],
      ),
    );
  }

  String _getReportTypeLabel(String type) {
    switch (type) {
      case 'sales':
        return global.language("alert_sales_summary");
      case 'gross_profit':
        return global.language("alert_gross_profit_report");
      default:
        return type;
    }
  }

  String _formatScheduleMode(ScheduleModel schedule) {
    if (schedule.scheduleMode == 'interval') {
      switch (schedule.intervalMinutes) {
        case 1:
          return global.language("alert_interval_1_min_test");
        case 5:
          return global.language("alert_interval_5_min_test");
        case 15:
          return global.language("alert_interval_15_min");
        case 30:
          return global.language("alert_interval_30_min");
        case 60:
          return global.language("alert_interval_60_min");
        default:
          return global.language("alert_every_x_min").replaceAll("{minutes}", schedule.intervalMinutes.toString());
      }
    } else if (schedule.scheduleMode == 'weekly') {
      return _formatWeeklySchedule(schedule.weeklySchedule);
    } else if (schedule.scheduleMode == 'monthly') {
      return _formatMonthlySchedule(schedule.monthlySchedule);
    } else {
      return global.language("alert_not_specified");
    }
  }

  /// จัดรูปแบบการแสดงผล monthly schedule
  String _formatMonthlySchedule(Map<int, List<String>> monthlySchedule) {
    if (monthlySchedule.isEmpty) return global.language("alert_no_date_defined");

    List<String> parts = [];

    // เรียงวันที่ (-1 = วันสิ้นเดือน อยู่ท้ายสุด)
    final sortedDates = monthlySchedule.keys.toList()
      ..sort((a, b) {
        if (a == -1) return 1;
        if (b == -1) return -1;
        return a.compareTo(b);
      });

    for (final date in sortedDates) {
      final times = monthlySchedule[date];
      if (times != null && times.isNotEmpty) {
        final dateLabel = date == ScheduleModel.lastDayOfMonth ? global.language("alert_end_of_month") : global.language("alert_date_label").replaceAll("{date}", date.toString());
        parts.add('$dateLabel (${times.join(", ")})');
      }
    }

    return parts.join(' | ');
  }

  String _formatWeeklySchedule(Map<int, List<String>> weeklySchedule) {
    if (weeklySchedule.isEmpty) return global.language("alert_no_time_defined");

    List<String> parts = [];
    for (int i = 0; i < 7; i++) {
      if (weeklySchedule.containsKey(i) && weeklySchedule[i]!.isNotEmpty) {
        final dayName = ScheduleModel.dayNamesShort[i];
        final times = weeklySchedule[i]!.join(', ');
        parts.add('$dayName $times');
      }
    }
    return parts.join(' | ');
  }

  String _formatDateTime(String? dateTimeStr) {
    if (dateTimeStr == null || dateTimeStr.isEmpty) return '-';
    try {
      final dt = DateTime.parse(dateTimeStr).toLocal();
      return global.formatThaiDateTime(dateTime: dt, showTime: true);
    } catch (e) {
      return dateTimeStr;
    }
  }
}

// ========================================
// ฟอร์มแก้ไข Alert (แสดงใน Panel ขวา)
// ========================================
class AlertAgentEditForm extends StatefulWidget {
  final AlertAgentModel? alert;
  final String shopId;
  final Function(AlertAgentModel) onSave;
  final VoidCallback onCancel;
  final Color primaryColor;

  const AlertAgentEditForm({super.key, this.alert, required this.shopId, required this.onSave, required this.onCancel, required this.primaryColor});

  @override
  State<AlertAgentEditForm> createState() => _AlertAgentEditFormState();
}

class _AlertAgentEditFormState extends State<AlertAgentEditForm> {
  final _formKey = GlobalKey<FormState>();
  final _taskNameController = TextEditingController();

  String _reportType = 'sales';
  String _reportPeriod = 'today';

  // Schedule settings - เลือกใช้ได้แค่ 1 โหมด แต่เก็บข้อมูลไว้ได้ทุกโหมด
  String _scheduleMode = 'interval'; // 'interval', 'weekly', 'monthly'
  int _intervalMinutes = 60; // ใช้เมื่อ mode = 'interval'
  Map<int, List<String>> _weeklySchedule = {}; // ใช้เมื่อ mode = 'weekly' (key: 0-6 = อา.-ส.)
  Map<int, List<String>> _monthlySchedule = {}; // ใช้เมื่อ mode = 'monthly' (key: 1-31 หรือ -1 = วันสิ้นเดือน)

  bool _emailEnabled = false;
  List<String> _emailAddresses = [];
  final _emailController = TextEditingController();

  bool _lineEnabled = false;
  List<String> _lineTokens = [];
  final _lineTokenController = TextEditingController();
  final Map<String, String> _lineUserNames = {}; // เก็บ mapping userId -> displayName

  bool get isNew => widget.alert == null;

  @override
  void initState() {
    super.initState();
    if (widget.alert != null) {
      final alert = widget.alert!;
      _taskNameController.text = alert.taskName;
      _reportType = alert.reportConfig.type;
      _reportPeriod = alert.reportConfig.period;
      _scheduleMode = alert.schedule.scheduleMode;
      _intervalMinutes = alert.schedule.intervalMinutes;
      // Copy weeklySchedule
      _weeklySchedule = {};
      alert.schedule.weeklySchedule.forEach((key, value) {
        _weeklySchedule[key] = List<String>.from(value);
      });
      // Copy monthlySchedule
      _monthlySchedule = {};
      alert.schedule.monthlySchedule.forEach((key, value) {
        _monthlySchedule[key] = List<String>.from(value);
      });
      _emailEnabled = alert.channels.email.isEnabled;
      _emailAddresses = List.from(alert.channels.email.addresses);
      _lineEnabled = alert.channels.line.isEnabled;
      _lineTokens = List.from(alert.channels.line.tokens);
    }
    // Fetch ชื่อ LINE users จาก MongoDB
    _fetchLineUserNames();
  }

  /// ดึงชื่อ LINE users จาก MongoDB
  Future<void> _fetchLineUserNames() async {
    if (_lineTokens.isEmpty) return;

    try {
      final apiBaseUrl = _getApiBaseUrl();
      final url = Uri.parse('$apiBaseUrl/atlas/query');
      final requestBody = {
        'database': 'bc_line_regis',
        'collection': 'line_users',
        'query': {
          'line_user_id': {'\$in': _lineTokens},
        },
      };

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List<dynamic> items = data['data'] ?? data['results'] ?? [];

        setState(() {
          for (var item in items) {
            final userId = item['line_user_id']?.toString() ?? '';
            final displayName = item['display_name']?.toString() ?? '';
            if (userId.isNotEmpty && displayName.isNotEmpty) {
              _lineUserNames[userId] = displayName;
            }
          }
        });
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Fetch LINE user names error: $e');
      }
    }
  }

  @override
  void dispose() {
    _taskNameController.dispose();
    _emailController.dispose();
    _lineTokenController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Header
        Container(
          padding: const EdgeInsets.all(16),
          color: Colors.white,
          child: Row(
            children: [
              Icon(isNew ? Icons.add_circle : Icons.edit, color: widget.primaryColor),
              SizedBox(width: 8),
              Text(
                isNew ? global.language("alert_create_new") : global.language("alert_edit_notification"),
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: widget.primaryColor),
              ),
              Spacer(),
              TextButton(onPressed: widget.onCancel, child: Text(global.language("cancel"))),
              SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: _submit,
                icon: Icon(Icons.save, size: 18),
                label: Text(isNew ? global.language("create") : global.language("save")),
                style: ElevatedButton.styleFrom(backgroundColor: widget.primaryColor, foregroundColor: Colors.white),
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        // Form
        Expanded(
          child: Form(
            key: _formKey,
            child: SingleChildScrollView(
              padding: EdgeInsets.all(16),
              child: Column(
                children: [
                  // ข้อมูลทั่วไป
                  _buildSectionCard(global.language("alert_general_info"), Icons.info_outline, [
                    TextFormField(
                      controller: _taskNameController,
                      decoration: InputDecoration(labelText: global.language("alert_task_name"), hintText: global.language("alert_task_name_hint"), border: OutlineInputBorder(), prefixIcon: Icon(Icons.label)),
                      validator: (value) => (value == null || value.isEmpty) ? global.language("alert_task_name_required") : null,
                    ),
                    SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: DropdownButtonFormField<String>(
                            initialValue: _reportType,
                            decoration: InputDecoration(labelText: global.language("alert_report_type"), border: OutlineInputBorder(), isDense: true),
                            items: [
                              DropdownMenuItem(value: 'sales', child: Text(global.language("alert_sales_summary"))),
                              DropdownMenuItem(value: 'gross_profit', child: Text(global.language("alert_gross_profit_report"))),
                            ],
                            onChanged: (value) => setState(() => _reportType = value!),
                          ),
                        ),
                        SizedBox(width: 12),
                        Expanded(
                          child: DropdownButtonFormField<String>(
                            initialValue: _reportPeriod,
                            decoration: InputDecoration(labelText: global.language("alert_data_period"), border: OutlineInputBorder(), isDense: true),
                            items: [
                              DropdownMenuItem(value: 'today', child: Text(global.language("alert_today"))),
                              DropdownMenuItem(value: 'yesterday', child: Text(global.language("alert_yesterday"))),
                              DropdownMenuItem(value: 'this_week', child: Text(global.language("alert_this_week"))),
                              DropdownMenuItem(value: 'this_month', child: Text(global.language("alert_this_month"))),
                            ],
                            onChanged: (value) => setState(() => _reportPeriod = value!),
                          ),
                        ),
                      ],
                    ),
                  ]),
                  const SizedBox(height: 16),

                  // ตั้งค่าการส่ง
                  _buildSectionCard(global.language("alert_sending_settings"), Icons.schedule, [
                    Text(global.language("alert_select_send_format"), style: TextStyle(fontWeight: FontWeight.bold)),
                    SizedBox(height: 4),
                    Text(
                      global.language("alert_send_format_note"),
                      style: TextStyle(fontSize: 11, color: Colors.grey[600], fontStyle: FontStyle.italic),
                    ),
                    const SizedBox(height: 12),
                    _buildScheduleModeSelector(),
                    const SizedBox(height: 16),
                    // แสดง UI ตามโหมดที่เลือก
                    if (_scheduleMode == 'interval') ...[
                      _buildModeDescription(global.language("alert_interval_desc"), Icons.repeat, Colors.blue),
                      const SizedBox(height: 12),
                      _buildIntervalSelector(),
                      const SizedBox(height: 8),
                      _buildIntervalNote(),
                    ] else if (_scheduleMode == 'weekly') ...[
                      _buildModeDescription(global.language("alert_weekly_desc"), Icons.calendar_view_week, Colors.green),
                      const SizedBox(height: 12),
                      _buildWeeklyScheduleSelector(),
                    ] else if (_scheduleMode == 'monthly') ...[
                      _buildModeDescription(global.language("alert_monthly_desc"), Icons.calendar_month, Colors.orange),
                      const SizedBox(height: 12),
                      _buildMonthlyScheduleSelector(),
                    ],
                  ]),
                  const SizedBox(height: 16),

                  // ช่องทางการแจ้งเตือน
                  _buildSectionCard(global.language("alert_notification_channel"), Icons.send, [
                    _buildChannelSection('Email', Icons.email, Colors.blue, _emailEnabled, (v) => setState(() => _emailEnabled = v!), _emailAddresses, _emailController, global.language("alert_enter_email"), (v) {
                      if (v.isNotEmpty && !_emailAddresses.contains(v)) {
                        setState(() {
                          _emailAddresses.add(v);
                          _emailController.clear();
                        });
                      }
                    }, (v) => setState(() => _emailAddresses.remove(v))),
                    const Divider(),
                    _buildLineChannelSection(),
                    // แสดง QR Code และขั้นตอนการลงทะเบียน LINE OA
                    if (_lineEnabled) _buildLineOaQrCode(),
                  ]),
                  const SizedBox(height: 40),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSectionCard(String title, IconData icon, List<Widget> children) {
    return Card(
      elevation: 1,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, color: widget.primaryColor, size: 20),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: widget.primaryColor),
                ),
              ],
            ),
            const Divider(height: 24),
            ...children,
          ],
        ),
      ),
    );
  }

  Widget _buildScheduleModeSelector() {
    return Column(
      children: [
        Row(
          children: [
            Expanded(child: _buildModeCard('Interval', global.language("alert_every_x_minutes"), Icons.timer, _scheduleMode == 'interval', () => setState(() => _scheduleMode = 'interval'))),
            SizedBox(width: 8),
            Expanded(child: _buildModeCard(global.language("alert_weekly"), global.language("alert_sun_sat"), Icons.calendar_view_week, _scheduleMode == 'weekly', () => setState(() => _scheduleMode = 'weekly'))),
            SizedBox(width: 8),
            Expanded(child: _buildModeCard(global.language("alert_monthly"), global.language("alert_date_1_31"), Icons.calendar_month, _scheduleMode == 'monthly', () => setState(() => _scheduleMode = 'monthly'))),
          ],
        ),
      ],
    );
  }

  /// แสดงคำอธิบายโหมดที่เลือก
  Widget _buildModeDescription(String description, IconData icon, Color color) {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(description, style: TextStyle(fontSize: 12, color: color.withValues(alpha: 0.8))),
          ),
        ],
      ),
    );
  }

  Widget _buildModeCard(String title, String subtitle, IconData icon, bool isSelected, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isSelected ? widget.primaryColor.withValues(alpha: 0.1) : Colors.grey[100],
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: isSelected ? widget.primaryColor : Colors.grey[300]!, width: isSelected ? 2 : 1),
        ),
        child: Column(
          children: [
            Icon(icon, color: isSelected ? widget.primaryColor : Colors.grey, size: 24),
            const SizedBox(height: 4),
            Text(
              title,
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: isSelected ? widget.primaryColor : Colors.grey[700]),
            ),
            Text(subtitle, style: TextStyle(fontSize: 10, color: Colors.grey[600])),
          ],
        ),
      ),
    );
  }

  /// สร้าง UI สำหรับเลือกวัน+เวลาในสัปดาห์
  Widget _buildWeeklyScheduleSelector() {
    return Column(
      children: List.generate(7, (dayIndex) {
        final dayName = ScheduleModel.dayNames[dayIndex];
        final times = _weeklySchedule[dayIndex] ?? [];
        final isEnabled = times.isNotEmpty;

        return Container(
          margin: const EdgeInsets.only(bottom: 8),
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: isEnabled ? widget.primaryColor.withValues(alpha: 0.05) : Colors.grey[50],
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: isEnabled ? widget.primaryColor.withValues(alpha: 0.3) : Colors.grey[300]!),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header: ชื่อวัน + ปุ่มเพิ่มเวลา
              Row(
                children: [
                  Icon(isEnabled ? Icons.check_circle : Icons.circle_outlined, color: isEnabled ? widget.primaryColor : Colors.grey, size: 20),
                  const SizedBox(width: 8),
                  Text(
                    dayName,
                    style: TextStyle(fontWeight: FontWeight.bold, color: isEnabled ? widget.primaryColor : Colors.grey[700]),
                  ),
                  const Spacer(),
                  // ปุ่มเพิ่มเวลา
                  TextButton.icon(
                    onPressed: () => _addTimeForDay(dayIndex),
                    icon: Icon(Icons.add, size: 16),
                    label: Text(global.language("alert_add_time"), style: TextStyle(fontSize: 12)),
                    style: TextButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4), minimumSize: Size.zero, tapTargetSize: MaterialTapTargetSize.shrinkWrap),
                  ),
                ],
              ),
              // แสดงเวลาที่เลือก
              if (times.isNotEmpty) ...[
                const SizedBox(height: 8),
                Wrap(
                  spacing: 6,
                  runSpacing: 4,
                  children: times.map((time) {
                    return Chip(
                      label: Text(time, style: const TextStyle(fontSize: 12)),
                      deleteIcon: const Icon(Icons.close, size: 14),
                      onDeleted: () => _removeTimeForDay(dayIndex, time),
                      backgroundColor: widget.primaryColor.withValues(alpha: 0.1),
                      visualDensity: VisualDensity.compact,
                      padding: EdgeInsets.zero,
                    );
                  }).toList(),
                ),
              ],
            ],
          ),
        );
      }),
    );
  }

  /// เพิ่มเวลาสำหรับวันที่ระบุ
  Future<void> _addTimeForDay(int dayIndex) async {
    final time = await showTimePicker(context: context, initialTime: TimeOfDay.now());
    if (time != null) {
      final timeStr = '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
      setState(() {
        if (_weeklySchedule[dayIndex] == null) {
          _weeklySchedule[dayIndex] = [];
        }
        if (!_weeklySchedule[dayIndex]!.contains(timeStr)) {
          _weeklySchedule[dayIndex]!.add(timeStr);
          _weeklySchedule[dayIndex]!.sort(); // เรียงตามเวลา
        }
      });
    }
  }

  /// ลบเวลาสำหรับวันที่ระบุ
  void _removeTimeForDay(int dayIndex, String time) {
    setState(() {
      _weeklySchedule[dayIndex]?.remove(time);
      // ถ้าไม่มีเวลาเหลือ ให้ลบ key ออก
      if (_weeklySchedule[dayIndex]?.isEmpty ?? false) {
        _weeklySchedule.remove(dayIndex);
      }
    });
  }

  /// สร้าง UI สำหรับเลือกวันที่ในเดือน+เวลา
  Widget _buildMonthlyScheduleSelector() {
    // สร้าง list ของวันที่ที่มีการตั้งค่า + วันสิ้นเดือน
    final configuredDates = _monthlySchedule.keys.toList()
      ..sort((a, b) {
        // -1 (วันสิ้นเดือน) ให้อยู่ท้ายสุด
        if (a == -1) return 1;
        if (b == -1) return -1;
        return a.compareTo(b);
      });

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // ปุ่มเพิ่มวันที่
        Row(
          children: [
            ElevatedButton.icon(
              onPressed: _showAddDateDialog,
              icon: Icon(Icons.add, size: 16),
              label: Text(global.language("alert_add_date"), style: TextStyle(fontSize: 12)),
              style: ElevatedButton.styleFrom(backgroundColor: widget.primaryColor, foregroundColor: Colors.white, padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8)),
            ),
            SizedBox(width: 8),
            ElevatedButton.icon(
              onPressed: () => _addLastDayOfMonth(),
              icon: Icon(Icons.last_page, size: 16),
              label: Text(global.language("alert_end_of_month"), style: TextStyle(fontSize: 12)),
              style: ElevatedButton.styleFrom(
                backgroundColor: _monthlySchedule.containsKey(-1) ? Colors.grey : Colors.orange,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        // แสดงวันที่ที่ตั้งค่าแล้ว
        if (configuredDates.isEmpty)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(color: Colors.grey[100], borderRadius: BorderRadius.circular(8)),
            child: Center(
              child: Text(global.language("alert_no_date_set"), style: TextStyle(color: Colors.grey[500])),
            ),
          )
        else
          ...configuredDates.map((date) => _buildMonthlyDateItem(date)),
      ],
    );
  }

  /// แสดง item สำหรับแต่ละวันที่ในเดือน
  Widget _buildMonthlyDateItem(int date) {
    final times = _monthlySchedule[date] ?? [];
    final isLastDay = date == ScheduleModel.lastDayOfMonth;
    final dateLabel = isLastDay ? global.language("alert_end_of_month") : global.language("alert_date_label").replaceAll("{date}", date.toString());

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: isLastDay ? Colors.orange.withValues(alpha: 0.05) : widget.primaryColor.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: isLastDay ? Colors.orange.withValues(alpha: 0.3) : widget.primaryColor.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(isLastDay ? Icons.last_page : Icons.calendar_today, color: isLastDay ? Colors.orange : widget.primaryColor, size: 18),
              const SizedBox(width: 8),
              Text(
                dateLabel,
                style: TextStyle(fontWeight: FontWeight.bold, color: isLastDay ? Colors.orange : widget.primaryColor),
              ),
              const Spacer(),
              // ปุ่มเพิ่มเวลา
              TextButton.icon(
                onPressed: () => _addTimeForMonth(date),
                icon: Icon(Icons.add, size: 16),
                label: Text(global.language("alert_add_time"), style: TextStyle(fontSize: 12)),
                style: TextButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4), minimumSize: Size.zero, tapTargetSize: MaterialTapTargetSize.shrinkWrap),
              ),
              // ปุ่มลบวันที่
              IconButton(
                onPressed: () => _removeMonthlyDate(date),
                icon: Icon(Icons.delete_outline, size: 18, color: Colors.red[400]),
                padding: EdgeInsets.zero,
                constraints: BoxConstraints(),
                tooltip: global.language("alert_delete_this_date"),
              ),
            ],
          ),
          if (times.isNotEmpty) ...[
            const SizedBox(height: 8),
            Wrap(
              spacing: 6,
              runSpacing: 4,
              children: times.map((time) {
                return Chip(
                  label: Text(time, style: const TextStyle(fontSize: 12)),
                  deleteIcon: const Icon(Icons.close, size: 14),
                  onDeleted: () => _removeTimeForMonth(date, time),
                  backgroundColor: isLastDay ? Colors.orange.withValues(alpha: 0.1) : widget.primaryColor.withValues(alpha: 0.1),
                  visualDensity: VisualDensity.compact,
                  padding: EdgeInsets.zero,
                );
              }).toList(),
            ),
          ] else
            Padding(
              padding: EdgeInsets.only(top: 4),
              child: Text(
                global.language("alert_please_add_time"),
                style: TextStyle(fontSize: 11, color: Colors.grey[500], fontStyle: FontStyle.italic),
              ),
            ),
        ],
      ),
    );
  }

  /// แสดง dialog เพิ่มวันที่
  Future<void> _showAddDateDialog() async {
    final dateController = TextEditingController();
    final result = await showDialog<int>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("alert_add_date")),
        content: TextField(
          controller: dateController,
          keyboardType: TextInputType.number,
          decoration: InputDecoration(labelText: global.language("alert_date_input_label"), border: OutlineInputBorder(), hintText: global.language("alert_date_input_hint")),
          autofocus: true,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(global.language("cancel"))),
          ElevatedButton(
            onPressed: () {
              final date = int.tryParse(dateController.text);
              if (date != null && date >= 1 && date <= 31) {
                Navigator.pop(context, date);
              }
            },
            child: Text(global.language("ok")),
          ),
        ],
      ),
    );

    if (result != null && !_monthlySchedule.containsKey(result)) {
      setState(() {
        _monthlySchedule[result] = [];
      });
    }
  }

  /// เพิ่มวันสิ้นเดือน
  void _addLastDayOfMonth() {
    if (!_monthlySchedule.containsKey(ScheduleModel.lastDayOfMonth)) {
      setState(() {
        _monthlySchedule[ScheduleModel.lastDayOfMonth] = [];
      });
    }
  }

  /// เพิ่มเวลาสำหรับวันที่ในเดือน
  Future<void> _addTimeForMonth(int date) async {
    final time = await showTimePicker(context: context, initialTime: TimeOfDay.now());
    if (time != null) {
      final timeStr = '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
      setState(() {
        if (!_monthlySchedule[date]!.contains(timeStr)) {
          _monthlySchedule[date]!.add(timeStr);
          _monthlySchedule[date]!.sort();
        }
      });
    }
  }

  /// ลบเวลาสำหรับวันที่ในเดือน
  void _removeTimeForMonth(int date, String time) {
    setState(() {
      _monthlySchedule[date]?.remove(time);
    });
  }

  /// ลบวันที่ทั้งหมด
  void _removeMonthlyDate(int date) {
    setState(() {
      _monthlySchedule.remove(date);
    });
  }

  Widget _buildIntervalSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // แถวแรก: สำหรับทดสอบ
        Row(
          children: [
            Text(
              global.language("alert_for_testing"),
              style: TextStyle(fontSize: 11, color: Colors.orange[700], fontStyle: FontStyle.italic),
            ),
          ],
        ),
        SizedBox(height: 6),
        Row(children: [_buildIntervalOption(1, global.language("alert_every_1_min"), global.language("alert_testing")), SizedBox(width: 8), _buildIntervalOption(5, global.language("alert_every_5_min"), global.language("alert_testing"))]),
        const SizedBox(height: 10),
        // แถวสอง: สำหรับใช้งานจริง
        Row(
          children: [
            Text(
              global.language("alert_for_production"),
              style: TextStyle(fontSize: 11, color: Colors.green[700], fontStyle: FontStyle.italic),
            ),
          ],
        ),
        SizedBox(height: 6),
        Row(
          children: [
            _buildIntervalOption(15, global.language("alert_every_15_min"), global.language("alert_4_times_per_hour")),
            SizedBox(width: 8),
            _buildIntervalOption(30, global.language("alert_every_30_min"), global.language("alert_2_times_per_hour")),
            SizedBox(width: 8),
            _buildIntervalOption(60, global.language("alert_every_1_hour"), global.language("alert_24_times_per_day")),
          ],
        ),
      ],
    );
  }

  Widget _buildIntervalOption(int minutes, String title, String subtitle) {
    final isSelected = _intervalMinutes == minutes;
    return Expanded(
      child: InkWell(
        onTap: () => setState(() => _intervalMinutes = minutes),
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 6),
          decoration: BoxDecoration(
            color: isSelected ? widget.primaryColor.withValues(alpha: 0.1) : Colors.grey[100],
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: isSelected ? widget.primaryColor : Colors.grey[300]!, width: isSelected ? 2 : 1),
          ),
          child: Column(
            children: [
              Text(
                title,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: isSelected ? widget.primaryColor : Colors.grey[700]),
              ),
              Text(subtitle, style: TextStyle(fontSize: 10, color: Colors.grey[600])),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildIntervalNote() {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.amber.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.amber.shade200),
      ),
      child: Row(
        children: [
          Icon(Icons.info_outline, color: Colors.amber.shade700, size: 18),
          SizedBox(width: 8),
          Expanded(
            child: Text(global.language("alert_no_resend_note"), style: TextStyle(fontSize: 12, color: Colors.amber.shade900)),
          ),
        ],
      ),
    );
  }

  /// แสดงปุ่มลงทะเบียน LINE OA
  Widget _buildLineOaQrCode() {
    return Container(
      margin: const EdgeInsets.only(top: 12, bottom: 8),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [BoxShadow(color: Colors.green.withValues(alpha: 0.2), blurRadius: 10, spreadRadius: 2, offset: const Offset(0, 4))],
        border: Border.all(color: Colors.green.withValues(alpha: 0.3)),
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(Icons.dialpad, color: Colors.green[700], size: 24),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      global.language("alert_register_line_user_id"),
                      style: TextStyle(fontWeight: FontWeight.bold, color: Colors.green[700], fontSize: 14),
                    ),
                    SizedBox(height: 4),
                    Text(global.language("alert_register_line_hint"), style: TextStyle(fontSize: 11, color: Colors.grey[600])),
                  ],
                ),
              ),
            ],
          ),
          SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: () => _showLineRegistrationDialog(),
              icon: Icon(Icons.pin, size: 20),
              label: Text(global.language("alert_create_registration_code")),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.green,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// แสดง Dialog สำหรับลงทะเบียน LINE
  Future<void> _showLineRegistrationDialog() async {
    final result = await showDialog<Map<String, String?>?>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _LineRegistrationDialog(shopId: widget.shopId, apiBaseUrl: _getApiBaseUrl()),
    );

    // ถ้าได้ข้อมูลกลับมา ให้เพิ่มเข้า list
    if (result != null && result['user_id'] != null) {
      final userId = result['user_id']!;
      final displayName = result['display_name'] ?? '';
      setState(() {
        if (!_lineTokens.contains(userId)) {
          _lineTokens.add(userId);
        }
        if (displayName.isNotEmpty) {
          _lineUserNames[userId] = displayName;
        }
      });
      if (mounted) {
        final nameText = displayName.isNotEmpty ? ' ($displayName)' : '';
        global.showSuccessSnackBar(context, '${global.language("alert_add_line_user_success")}$nameText');
      }
    }
  }

  /// ดึง API Base URL
  String _getApiBaseUrl() {
    String apiPath = global.myAppConfig.serviceApi;
    if (!apiPath.startsWith('http')) {
      apiPath = 'https://$apiPath';
    }
    if (apiPath.endsWith('/')) {
      apiPath = apiPath.substring(0, apiPath.length - 1);
    }
    if (global.myAppConfig.servicePort.isNotEmpty) {
      return '$apiPath:${global.myAppConfig.servicePort}';
    }
    return apiPath;
  }

  /// สร้าง UI สำหรับ LINE OA พร้อมแสดงชื่อ user
  Widget _buildLineChannelSection() {
    const color = Colors.green;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CheckboxListTile(
          title: Row(
            children: [
              Icon(Icons.chat, size: 18, color: color),
              const SizedBox(width: 8),
              const Text('LINE OA', style: TextStyle(fontWeight: FontWeight.bold)),
            ],
          ),
          value: _lineEnabled,
          onChanged: (v) => setState(() => _lineEnabled = v!),
          controlAffinity: ListTileControlAffinity.leading,
          contentPadding: EdgeInsets.zero,
          dense: true,
          activeColor: color,
        ),
        // คำเตือนค่าใช้จ่าย LINE OA
        if (_lineEnabled)
          Container(
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.orange.shade50,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.orange.shade200),
            ),
            child: Row(
              children: [
                Icon(Icons.info_outline, color: Colors.orange.shade700, size: 18),
                SizedBox(width: 8),
                Expanded(
                  child: Text('\u{1F4B0} ${global.language("alert_line_oa_cost")}', style: TextStyle(fontSize: 12, color: Colors.orange.shade800)),
                ),
              ],
            ),
          ),
        if (_lineEnabled) ...[
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _lineTokenController,
                  decoration: InputDecoration(hintText: global.language("alert_enter_line_user_id"), border: OutlineInputBorder(), isDense: true, contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 10)),
                  onSubmitted: (v) {
                    if (v.isNotEmpty && !_lineTokens.contains(v)) {
                      setState(() {
                        _lineTokens.add(v);
                        _lineTokenController.clear();
                      });
                      // Fetch ชื่อ user ใหม่
                      _fetchSingleLineUserName(v);
                    }
                  },
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.add_circle, color: color),
                onPressed: () {
                  final v = _lineTokenController.text.trim();
                  if (v.isNotEmpty && !_lineTokens.contains(v)) {
                    setState(() {
                      _lineTokens.add(v);
                      _lineTokenController.clear();
                    });
                    _fetchSingleLineUserName(v);
                  }
                },
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (_lineTokens.isNotEmpty)
            Wrap(
              spacing: 8,
              runSpacing: 4,
              children: _lineTokens.map((userId) {
                final displayName = _lineUserNames[userId];
                final label = displayName ?? userId;
                return Chip(
                  avatar: displayName != null ? const Icon(Icons.person, size: 16, color: Colors.green) : null,
                  label: Text(label.length > 30 ? '${label.substring(0, 30)}...' : label, style: const TextStyle(fontSize: 11)),
                  deleteIcon: const Icon(Icons.close, size: 14),
                  onDeleted: () => setState(() {
                    _lineTokens.remove(userId);
                    _lineUserNames.remove(userId);
                  }),
                  backgroundColor: color.withValues(alpha: 0.1),
                  visualDensity: VisualDensity.compact,
                );
              }).toList(),
            ),
        ],
        const SizedBox(height: 8),
      ],
    );
  }

  /// ดึงชื่อ LINE user เดี่ยวจาก MongoDB
  Future<void> _fetchSingleLineUserName(String userId) async {
    try {
      final apiBaseUrl = _getApiBaseUrl();
      final url = Uri.parse('$apiBaseUrl/atlas/query');
      final requestBody = {
        'database': 'bc_line_regis',
        'collection': 'line_users',
        'query': {'line_user_id': userId},
      };

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List<dynamic> items = data['data'] ?? data['results'] ?? [];

        if (items.isNotEmpty) {
          final item = items.first;
          final displayName = item['display_name']?.toString() ?? '';
          if (displayName.isNotEmpty) {
            setState(() {
              _lineUserNames[userId] = displayName;
            });
          }
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Fetch single LINE user name error: $e');
      }
    }
  }

  Widget _buildChannelSection(
    String title,
    IconData icon,
    Color color,
    bool isEnabled,
    ValueChanged<bool?> onEnabledChanged,
    List<String> items,
    TextEditingController controller,
    String placeholder,
    ValueChanged<String> onAdd,
    ValueChanged<String> onRemove,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CheckboxListTile(
          title: Row(
            children: [
              Icon(icon, size: 18, color: color),
              const SizedBox(width: 8),
              Text(title, style: const TextStyle(fontWeight: FontWeight.bold)),
            ],
          ),
          value: isEnabled,
          onChanged: onEnabledChanged,
          controlAffinity: ListTileControlAffinity.leading,
          contentPadding: EdgeInsets.zero,
          dense: true,
          activeColor: color,
        ),
        if (isEnabled) ...[
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: controller,
                  decoration: InputDecoration(hintText: placeholder, border: const OutlineInputBorder(), isDense: true, contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 10)),
                  onSubmitted: onAdd,
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: Icon(Icons.add_circle, color: color),
                onPressed: () => onAdd(controller.text),
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (items.isNotEmpty)
            Wrap(
              spacing: 8,
              runSpacing: 4,
              children: items
                  .map(
                    (item) => Chip(
                      label: Text(item.length > 25 ? '${item.substring(0, 25)}...' : item, style: const TextStyle(fontSize: 11)),
                      deleteIcon: const Icon(Icons.close, size: 14),
                      onDeleted: () => onRemove(item),
                      backgroundColor: color.withValues(alpha: 0.1),
                      visualDensity: VisualDensity.compact,
                    ),
                  )
                  .toList(),
            ),
        ],
        const SizedBox(height: 8),
      ],
    );
  }

  /// เพิ่มข้อมูลที่ค้างอยู่ใน TextField อัตโนมัติ (กรณีลืมกด +)
  void _addPendingChannelData() {
    // Email
    final emailText = _emailController.text.trim();
    if (_emailEnabled && emailText.isNotEmpty && !_emailAddresses.contains(emailText)) {
      _emailAddresses.add(emailText);
      _emailController.clear();
    }

    // LINE
    final lineText = _lineTokenController.text.trim();
    if (_lineEnabled && lineText.isNotEmpty && !_lineTokens.contains(lineText)) {
      _lineTokens.add(lineText);
      _lineTokenController.clear();
    }
  }

  void _submit() {
    if (_formKey.currentState!.validate()) {
      // Validation สำหรับ weekly mode - ต้องเลือกอย่างน้อย 1 วัน+เวลา
      if (_scheduleMode == 'weekly') {
        // ตรวจสอบว่ามีวันที่เลือกและมีเวลากำหนดอย่างน้อย 1 รายการ
        final hasValidWeeklySchedule = _weeklySchedule.entries.any((e) => e.value.isNotEmpty);
        if (!hasValidWeeklySchedule) {
          global.showWarningSnackBar(context, global.language("alert_select_day_time_required"));
          return;
        }
      }

      // Validation สำหรับ monthly mode - ต้องเลือกอย่างน้อย 1 วันที่+เวลา
      if (_scheduleMode == 'monthly') {
        // ตรวจสอบว่ามีวันที่ที่กำหนดเวลาไว้อย่างน้อย 1 รายการ
        final hasValidMonthlySchedule = _monthlySchedule.entries.any((e) => e.value.isNotEmpty);
        if (!hasValidMonthlySchedule) {
          global.showWarningSnackBar(context, global.language("alert_select_date_time_required"));
          return;
        }
      }

      // เพิ่มข้อมูลที่ค้างอยู่ใน TextField อัตโนมัติ (กรณีลืมกด +)
      _addPendingChannelData();

      final hasValidEmail = _emailEnabled && _emailAddresses.isNotEmpty;
      final hasValidLine = _lineEnabled && _lineTokens.isNotEmpty;

      if (!hasValidEmail && !hasValidLine) {
        String message = global.language("alert_select_channel_required");
        if (_emailEnabled && _emailAddresses.isEmpty) {
          message = global.language("alert_add_email_required");
        } else if (_lineEnabled && _lineTokens.isEmpty) {
          message = global.language("alert_add_line_required");
        }
        global.showWarningSnackBar(context, message);
        return;
      }

      final alert = AlertAgentModel(
        id: widget.alert?.id ?? '',
        shopId: widget.shopId,
        taskName: _taskNameController.text,
        isActive: widget.alert?.isActive ?? true,
        reportConfig: ReportConfigModel(type: _reportType, period: _reportPeriod),
        schedule: ScheduleModel(scheduleMode: _scheduleMode, intervalMinutes: _intervalMinutes, weeklySchedule: _weeklySchedule, monthlySchedule: _monthlySchedule),
        channels: ChannelsModel(
          email: EmailChannelModel(isEnabled: _emailEnabled, addresses: _emailAddresses),
          line: LineChannelModel(isEnabled: _lineEnabled, tokens: _lineTokens),
        ),
        metadata: widget.alert?.metadata ?? MetadataModel(),
      );

      widget.onSave(alert);
    }
  }
}

// ========================================
// Dialog สำหรับลงทะเบียน LINE OA
// ========================================
class _LineRegistrationDialog extends StatefulWidget {
  final String shopId;
  final String apiBaseUrl;

  const _LineRegistrationDialog({required this.shopId, required this.apiBaseUrl});

  @override
  State<_LineRegistrationDialog> createState() => _LineRegistrationDialogState();
}

class _LineRegistrationDialogState extends State<_LineRegistrationDialog> {
  String? _registrationCode;
  String? _lineUserId;
  String? _displayName;
  bool _isLoading = true;
  bool _isPolling = false;
  String? _errorMessage;
  Timer? _pollTimer;
  Timer? _countdownTimer;
  int _secondsRemaining = 60; // นับถอยหลัง 60 วินาที

  static const String mongoDatabase = 'bc_line_regis';
  static const String mongoCollection = 'users';

  @override
  void initState() {
    super.initState();
    _startRegistration();
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    super.dispose();
  }

  /// สร้างเลข 4 หลักสำหรับลงทะเบียน
  String _generate4DigitCode() {
    final random = Random();
    return (1000 + random.nextInt(9000)).toString(); // สร้างเลข 1000-9999
  }

  /// เริ่มกระบวนการลงทะเบียน
  Future<void> _startRegistration() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _secondsRemaining = 60; // รีเซ็ต countdown
    });

    try {
      // สร้างเลข 4 หลัก
      final code = _generate4DigitCode();

      // บันทึกไป MongoDB
      final url = Uri.parse('${widget.apiBaseUrl}/atlas/update');
      final requestBody = {
        'database': mongoDatabase,
        'collection': mongoCollection,
        'shopid': widget.shopId,
        'email': code,
        'cartid': code,
        'data': {
          'registration_code': code,
          'shop': {'shopid': widget.shopId, 'shop_name': global.getShopName()},
          'status': 'pending',
          'line_user_id': null,
          'display_name': null,
          'picture_url': null,
          'created_at': DateTime.now().toIso8601String(),
          'expires_at': DateTime.now().add(const Duration(seconds: 60)).toIso8601String(),
        },
        'upsert': true, // สร้างใหม่ถ้าไม่มี
      };

      if (kDebugMode) {
        AppLogger.debug('Creating LINE registration: $code');
        AppLogger.debug('Request URL: $url');
        AppLogger.debug('Request body: ${jsonEncode(requestBody)}');
      }

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (kDebugMode) {
        AppLogger.debug('Response status: ${response.statusCode}');
        AppLogger.debug('Response body: ${response.body}');
      }

      if (response.statusCode == 200) {
        setState(() {
          _registrationCode = code;
          _isLoading = false;
        });
        // เริ่ม polling และ countdown
        _startPolling();
        _startCountdown();
      } else {
        throw Exception('Failed to create registration: ${response.body}');
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = '${global.language("error_with_detail")}: $e';
      });
    }
  }

  /// เริ่มนับถอยหลัง 60 วินาที
  void _startCountdown() {
    _countdownTimer?.cancel();
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_secondsRemaining <= 0) {
        timer.cancel();
        return;
      }
      setState(() {
        _secondsRemaining--;
      });
    });
  }

  /// เริ่ม polling เพื่อตรวจสอบ userId
  void _startPolling() {
    setState(() => _isPolling = true);

    // Poll ทุก 2 วินาที เป็นเวลา 60 วินาที
    int pollCount = 0;
    const maxPolls = 30; // 60 วินาที / 2 วินาที = 30 ครั้ง

    _pollTimer = Timer.periodic( Duration(seconds: 2), (timer) async {
      pollCount++;

      if (pollCount >= maxPolls || _secondsRemaining <= 0) {
        timer.cancel();
        _countdownTimer?.cancel();
        // ลบ registration record เมื่อหมดเวลา
        await _deleteRegistration();
        setState(() {
          _isPolling = false;
          _errorMessage = global.language("alert_registration_timeout");
        });
        return;
      }

      try {
        final result = await _checkRegistrationStatus();
        if (result != null) {
          timer.cancel();
          _countdownTimer?.cancel();
          setState(() {
            _lineUserId = result['user_id'];
            _displayName = result['display_name'];
            _isPolling = false;
          });
          // เก็บข้อมูล user ลง line_users collection
          await _saveLineUser(result['user_id']!, result['display_name'] ?? '', result['picture_url'] ?? '');
          // ลบ registration record
          await _deleteRegistration();
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Polling error: $e');
        }
      }
    });
  }

  /// ตรวจสอบสถานะการลงทะเบียน
  Future<Map<String, String>?> _checkRegistrationStatus() async {
    try {
      final url = Uri.parse('${widget.apiBaseUrl}/atlas/get');
      final requestBody = {'database': mongoDatabase, 'collection': mongoCollection, 'shopid': widget.shopId, 'email': _registrationCode, 'cartid': _registrationCode};

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List<dynamic> items = data['data'] ?? data['results'] ?? [];

        if (items.isNotEmpty) {
          final record = items.first;
          if (record['status'] == 'completed' && record['line_user_id'] != null) {
            return {'user_id': record['line_user_id'].toString(), 'display_name': record['display_name']?.toString() ?? '', 'picture_url': record['picture_url']?.toString() ?? ''};
          }
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Check status error: $e');
      }
    }
    return null;
  }

  /// ลบ registration record
  Future<void> _deleteRegistration() async {
    try {
      final url = Uri.parse('${widget.apiBaseUrl}/atlas/delete');
      final requestBody = {'database': mongoDatabase, 'collection': mongoCollection, 'shopid': widget.shopId, 'email': _registrationCode, 'cartid': _registrationCode};

      await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Delete registration error: $e');
      }
    }
  }

  /// เก็บข้อมูล LINE user ลง line_users collection
  Future<void> _saveLineUser(String userId, String displayName, String pictureUrl) async {
    try {
      final url = Uri.parse('${widget.apiBaseUrl}/atlas/update');
      final requestBody = {
        'database': mongoDatabase,
        'collection': 'line_users',
        'shopid': widget.shopId,
        'email': userId,
        'cartid': userId,
        'data': {'line_user_id': userId, 'display_name': displayName, 'picture_url': pictureUrl, 'shop_id': widget.shopId, 'updated_at': DateTime.now().toIso8601String()},
        'upsert': true,
      };

      await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(requestBody));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Save LINE user error: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          Icon(Icons.chat, color: Colors.green, size: 24),
          SizedBox(width: 8),
          Text(global.language("alert_register_line_oa")),
        ],
      ),
      content: SizedBox(width: 350, child: _buildContent()),
      actions: [
        if (_lineUserId != null)
          ElevatedButton.icon(
            onPressed: () => Navigator.pop<Map<String, String?>>(context, {'user_id': _lineUserId!, 'display_name': _displayName}),
            icon: Icon(Icons.check),
            label: Text(_displayName != null ? '${global.language("alert_use_this_user_id")} $_displayName' : global.language("alert_use_this_user_id")),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.green, foregroundColor: Colors.white),
          )
        else
          TextButton(
            onPressed: () {
              _pollTimer?.cancel();
              Navigator.pop(context);
            },
            child: Text(global.language("cancel")),
          ),
      ],
    );
  }

  Widget _buildContent() {
    if (_isLoading) {
      return Column(mainAxisSize: MainAxisSize.min, children: [ CircularProgressIndicator(), SizedBox(height: 16), Text(global.language("alert_generating_code"))]);
    }

    if (_errorMessage != null) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.error_outline, color: Colors.red, size: 48),
          const SizedBox(height: 16),
          Text(_errorMessage!, style: const TextStyle(color: Colors.red)),
          SizedBox(height: 16),
          ElevatedButton.icon(onPressed: _startRegistration, icon: Icon(Icons.refresh), label: Text(global.language("try_again"))),
        ],
      );
    }

    if (_lineUserId != null) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.green.shade50,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.green),
            ),
            child: Column(
              children: [
                const Icon(Icons.check_circle, color: Colors.green, size: 48),
                SizedBox(height: 12),
                Text(
                  global.language("alert_registration_success"),
                  style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.green),
                ),
                const SizedBox(height: 8),
                if (_displayName != null && _displayName!.isNotEmpty) ...[
                  Text(global.language("alert_line_name"), style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                  const SizedBox(height: 4),
                  Text(_displayName!, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                ],
                Text('LINE User ID:', style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                const SizedBox(height: 4),
                SelectableText(
                  _lineUserId!,
                  style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, fontFamily: 'monospace'),
                ),
              ],
            ),
          ),
        ],
      );
    }

    // แสดง QR Code เพิ่มเพื่อน (บน) และ เลข 4 หลัก (ล่าง)
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // 1. QR Code เพิ่มเพื่อน LINE OA (บนสุด)
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [BoxShadow(color: Colors.green.withValues(alpha: 0.2), blurRadius: 8, spreadRadius: 2)],
            border: Border.all(color: Colors.green.shade200),
          ),
          child: Column(
            children: [
              Text(
                global.language("alert_step1_add_friend"),
                style: TextStyle(fontSize: 12, color: Colors.green.shade700, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Image.asset(
                  'assets/img/bclineoa.png',
                  width: 140,
                  height: 140,
                  fit: BoxFit.cover,
                  errorBuilder: (context, error, stackTrace) {
                    return Container(
                      width: 140,
                      height: 140,
                      color: Colors.grey[200],
                      child: Icon(Icons.qr_code, size: 60, color: Colors.grey),
                    );
                  },
                ),
              ),
              SizedBox(height: 8),
              Text(global.language("alert_scan_qr_search"), style: TextStyle(fontSize: 11, color: Colors.grey[600])),
            ],
          ),
        ),
        const SizedBox(height: 16),
        // 2. รหัส 4 หลัก (ด้านล่าง)
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
          decoration: BoxDecoration(
            color: Colors.green.shade50,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.green, width: 3),
            boxShadow: [BoxShadow(color: Colors.green.withValues(alpha: 0.3), blurRadius: 15, spreadRadius: 3)],
          ),
          child: Column(
            children: [
              Text(
                global.language("alert_step2_send_code"),
                style: TextStyle(fontSize: 12, color: Colors.green.shade700, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                _registrationCode!,
                style: TextStyle(fontSize: 56, fontWeight: FontWeight.bold, color: Colors.green.shade800, letterSpacing: 10, fontFamily: 'monospace'),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        // แสดง Countdown Timer
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
          decoration: BoxDecoration(
            color: _secondsRemaining <= 10 ? Colors.red.shade50 : Colors.orange.shade50,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: _secondsRemaining <= 10 ? Colors.red : Colors.orange),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.timer, color: _secondsRemaining <= 10 ? Colors.red : Colors.orange, size: 20),
              SizedBox(width: 8),
              Text(
                global.language("alert_timeout_seconds").replaceAll("{seconds}", _secondsRemaining.toString()),
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: _secondsRemaining <= 10 ? Colors.red : Colors.orange.shade800),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        if (_isPolling)
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.green)),
              SizedBox(width: 8),
              Text(global.language("alert_waiting_registration"), style: TextStyle(color: Colors.grey[600], fontSize: 12)),
            ],
          ),
      ],
    );
  }
}
