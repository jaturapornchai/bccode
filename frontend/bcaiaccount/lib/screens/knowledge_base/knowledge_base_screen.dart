import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:file_picker/file_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/knowledge_base/knowledge_base_cubit.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/widgets/chatbot_panel.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/utils/time_picker.dart';

class KnowledgeBaseScreen extends StatelessWidget {
  const KnowledgeBaseScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => KnowledgeBaseCubit()..loadDocuments(branchId: "*"),
      child: const _KnowledgeBaseScreenContent(),
    );
  }
}

class _KnowledgeBaseScreenContent extends StatefulWidget {
  const _KnowledgeBaseScreenContent();

  @override
  State<_KnowledgeBaseScreenContent> createState() =>
      _KnowledgeBaseScreenContentState();
}

class _KnowledgeBaseScreenContentState
    extends State<_KnowledgeBaseScreenContent>
    with global.ThemeRefreshMixin {
  // Document Controllers
  final ScrollController _documentScrollController = ScrollController();

  // Branch Selection
  List<CompanyBranchModel> _branches = [];
  String? _selectedBranchId;
  bool _showAllBranches = true;

  // Split view
  double _leftPanelWidth = 0.5; // 0.0 to 1.0 (50% default)
  bool _isDragging = false;

  // Colors
  Color get primaryColor => global.theme.primaryColor;

  // API Base URLs
  final String apiBaseUrl = kDebugMode
      ? 'http://localhost:9999/api/documents'
      : 'https://bcaicallcenter.dedetouch.com/api/documents';

  @override
  void initState() {
    super.initState();
    _loadBranches();
  }

  @override
  void dispose() {
    _documentScrollController.dispose();
    super.dispose();
  }

  // Load branches
  void _loadBranches() {
    context.read<CompanyBranchBloc>().add(
      const CompanyBranchLoadList(offset: 0, limit: 100, search: ""),
    );
  }

  // Load documents from MongoDB via Cubit
  void _loadDocuments() {
    // ถ้าเลือกทุกสาขา ให้ส่ง "*"
    // ถ้าเลือกสาขาเฉพาะ ให้ส่งรหัสสาขา
    final branchId = _showAllBranches ? "*" : _selectedBranchId;

    if (kDebugMode) {
      AppLogger.debug('Loading documents for branch: $branchId');
    }

    context.read<KnowledgeBaseCubit>().loadDocuments(branchId: branchId);
  }

  // Upload document
  Future<void> _uploadDocument() async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: [
          'pdf',
          'xlsx',
          'xls',
          'csv',
          'txt',
          'md',
          'doc',
          'docx',
        ],
        withData: true,
      );

      if (result != null && result.files.single.bytes != null) {
        final file = result.files.single;
        final bytes = file.bytes!;
        final base64Content = base64Encode(bytes);

        final shopId = global.getShopId();

        // ถ้าเลือก "ทุกสาขา" ให้ใช้ "*"
        // ถ้าเลือกสาขาเฉพาะ ให้ใช้รหัสสาขานั้น
        String branchId;
        if (_showAllBranches) {
          branchId = "*";
        } else if (_selectedBranchId != null) {
          branchId = _selectedBranchId!;
        } else {
          _showSnackBar(
            global.language('kb_please_select_branch_before_upload'),
            isError: true,
          );
          return;
        }

        if (kDebugMode) {
          AppLogger.debug('═══════════════════════════════════════');
          AppLogger.debug('📤 Uploading document');
          AppLogger.debug('File: ${file.name}');
          AppLogger.debug('Branch ID: $branchId');
          AppLogger.debug('Show All Branches: $_showAllBranches');
          AppLogger.debug('═══════════════════════════════════════');
        }

        final response = await http.post(
          Uri.parse('$apiBaseUrl/upload'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({
            'shop_id': shopId,
            'branch_id': branchId,
            'filename': file.name,
            'content_type': _getContentType(file.extension ?? ''),
            'content': base64Content,
            'uploaded_by': global.appConfig.getString('user') ?? 'unknown',
            'description': '',
            'tags': [],
            'status': true, // true = เปิดใช้งาน
            'all_day': true, // true = ทำงานตลอดเวลา
          }),
        );

        if (response.statusCode == 200) {
          final data = jsonDecode(response.body);
          if (data['success'] == true) {
            _showSnackBar(global.language('kb_upload_success'));
            // โหลดข้อมูลใหม่เฉพาะตอนอัพโหลด (เพราะเป็นเอกสารใหม่)
            _loadDocuments();
          } else {
            _showSnackBar(
              '${global.language('kb_error_occurred')}: ${data['message']}',
              isError: true,
            );
          }
        } else {
          _showSnackBar(global.language('kb_upload_error'), isError: true);
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error uploading document: $e');
      }
      _showSnackBar(
        '${global.language('kb_error_occurred')}: $e',
        isError: true,
      );
    }
  }

  // Delete document
  Future<void> _deleteDocument(DocumentModel doc) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.warning, color: Colors.orange),
            SizedBox(width: 8),
            Text(global.language('kb_confirm_delete')),
          ],
        ),
        content: Text(
          '${global.language('kb_delete_document_question')} "${doc.filename}" ${global.language('kb_or_not')}',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              foregroundColor: Colors.white,
            ),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      try {
        final response = await http.post(
          Uri.parse('$apiBaseUrl/delete'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({
            'shop_id': doc.shopId,
            'branch_id': doc.branchId,
            'filename': doc.filename,
          }),
        );

        if (response.statusCode == 200) {
          final data = jsonDecode(response.body);
          if (data['success'] == true) {
            // อัปเดต state ทันที (ไม่ต้องโหลดข้อมูลใหม่เพื่อป้องกันจอกระพริบ)
            context.read<KnowledgeBaseCubit>().removeDocument(doc.filename);
            _showSnackBar(global.language('kb_delete_success'));
          } else {
            _showSnackBar(
              '${global.language('kb_error_occurred')}: ${data['message']}',
              isError: true,
            );
          }
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error deleting document: $e');
        }
        _showSnackBar(
          '${global.language('kb_error_occurred')}: $e',
          isError: true,
        );
      }
    }
  }

  // Toggle document status (เปิด/ปิดใช้งาน)
  Future<void> _toggleDocumentStatus(DocumentModel doc) async {
    final newStatus = !doc.status; // toggle status
    final actionText = newStatus
        ? global.language('kb_enable')
        : global.language('kb_disable');

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(
              newStatus ? Icons.play_circle : Icons.pause_circle,
              color: Colors.orange,
            ),
            SizedBox(width: 8),
            Text('${global.language('kb_confirm_action')}$actionText'),
          ],
        ),
        content: Text(
          '${global.language('kb_do_you_want_to')}$actionText ${global.language('kb_document')} "${doc.filename}" ${global.language('kb_or_not')}',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.orange,
              foregroundColor: Colors.white,
            ),
            child: Text(actionText),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      try {
        final response = await http.post(
          Uri.parse('$apiBaseUrl/update-status'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({
            'shop_id': doc.shopId,
            'branch_id': doc.branchId,
            'filename': doc.filename,
            'status': newStatus,
          }),
        );

        if (response.statusCode == 200) {
          final data = jsonDecode(response.body);
          if (data['success'] == true) {
            // อัปเดต state ทันที (ไม่ต้องโหลดข้อมูลใหม่เพื่อป้องกันจอกระพริบ)
            context.read<KnowledgeBaseCubit>().updateDocumentStatus(
              doc.filename,
              newStatus,
            );
            _showSnackBar(
              newStatus
                  ? global.language('kb_enable_document_success')
                  : global.language('kb_disable_document_success'),
            );
          } else {
            _showSnackBar(
              '${global.language('kb_error_occurred')}: ${data['message']}',
              isError: true,
            );
          }
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error updating document status: $e');
        }
        _showSnackBar(
          '${global.language('kb_error_occurred')}: $e',
          isError: true,
        );
      }
    }
  }

  // Toggle all_day mode (ทำงานตลอดเวลา / ใช้ช่วงเวลาที่กำหนด)
  Future<void> _toggleAllDayMode(DocumentModel doc) async {
    try {
      final newAllDay = !doc.allDay;

      if (kDebugMode) {
        AppLogger.debug('Toggling all_day for ${doc.filename}: $newAllDay');
      }

      final response = await http.post(
        Uri.parse('$apiBaseUrl/update-allday'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'shop_id': doc.shopId,
          'branch_id': doc.branchId,
          'filename': doc.filename,
          'all_day': newAllDay,
        }),
      );

      if (kDebugMode) {
        AppLogger.debug('Response status: ${response.statusCode}');
        AppLogger.debug('Response body: ${response.body}');
      }

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          // อัปเดต state ทันที
          context.read<KnowledgeBaseCubit>().updateDocumentAllDay(
            doc.filename,
            newAllDay,
          );
          _showSnackBar(
            newAllDay
                ? global.language('kb_change_to_work_all_time')
                : global.language('kb_change_to_use_time_period'),
          );
        } else {
          _showSnackBar('เกิดข้อผิดพลาด: ${data['message']}', isError: true);
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error updating all_day mode: $e');
      }
      _showSnackBar(
        '${global.language('kb_error_occurred')}: $e',
        isError: true,
      );
    }
  }

  // Format Thai date
  String _formatThaiDate(DateTime date) {
    final day = date.day;
    final month = global.getThaiMonthNameShort(date.month); // ใช้ชื่อย่อ
    final year = date.year + 543;
    return '$day $month ${year.toString().substring(2)}'; // แสดงปีสองหลัก
  }

  // Format time
  String _formatTime(DateTime dateTime) {
    return '${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}';
  }

  // Get document date range status
  String _getDocumentDateRangeStatus(DocumentModel doc) {
    // ถ้า all_day = true แสดงว่าทำงานตลอดเวลา
    if (doc.allDay) {
      return global.language('kb_work_all_time');
    }

    // ถ้า all_day = false แต่ไม่มีวันที่กำหนด
    if (doc.startDateTime == null && doc.endDateTime == null) {
      return global.language('kb_no_date_set');
    }

    // แปลงเป็น Local Time (UTC+7) และแสดงทั้งวันที่และเวลา
    String start = '-';
    if (doc.startDateTime != null) {
      final dateTime = DateTime.parse(doc.startDateTime!).toLocal();
      start = '${_formatThaiDate(dateTime)} ${_formatTime(dateTime)}';
    }

    String end = '-';
    if (doc.endDateTime != null) {
      final dateTime = DateTime.parse(doc.endDateTime!).toLocal();
      end = '${_formatThaiDate(dateTime)} ${_formatTime(dateTime)}';
    }

    return '${global.language('kb_condition')} $start ${global.language('kb_to')} $end';
  }

  // Select start date
  Future<void> _selectStartDate(DocumentModel doc) async {
    DateTime initialDate = DateTime.now();

    // ตรวจสอบว่ามีข้อมูล start date หรือไม่
    if (doc.startDateTime != null && doc.startDateTime!.isNotEmpty) {
      try {
        final parsedDate = DateTime.parse(doc.startDateTime!);
        // ถ้าวันที่มากกว่าหรือเท่ากับปี 2000 ให้ใช้วันที่นั้น ไม่เช่นนั้นใช้วันที่ปัจจุบัน
        if (parsedDate.year >= 2000) {
          initialDate = parsedDate;
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error parsing start date: $e');
        }
      }
    }

    final DateTime? picked = await showDialog<DateTime>(
      context: context,
      builder: (BuildContext context) {
        return Dialog(
          child: Container(
            width: 320,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(8),
              color: global.theme.cardColor,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: primaryColor,
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(8),
                      topRight: Radius.circular(8),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.calendar_today, color: Colors.white),
                      SizedBox(width: 8),
                      Text(
                        global.language('kb_select_start_date'),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
                // Calendar
                // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                CustomDatePickerDialog(
                  initialDate: initialDate,
                  firstDate: DateTime(2020),
                  lastDate: DateTime(2100),
                ),
              ],
            ),
          ),
        );
      },
    );

    if (picked != null) {
      // ตรวจสอบว่าวันที่เปลี่ยนแปลงหรือไม่
      bool dateChanged = true;
      if (doc.startDateTime != null && doc.startDateTime!.isNotEmpty) {
        try {
          final currentDate = DateTime.parse(doc.startDateTime!).toLocal();
          dateChanged =
              currentDate.year != picked.year ||
              currentDate.month != picked.month ||
              currentDate.day != picked.day;
        } catch (e) {
          dateChanged = true;
        }
      }

      // ถ้าวันที่เปลี่ยนแปลง ให้ตั้งเวลา default เป็น 00:00:00 (Local Time UTC+7)
      // ถ้าไม่เปลี่ยน ให้ใช้เวลาเดิม
      DateTime newDateTime;
      if (dateChanged) {
        // ตั้งเวลาเป็น 00:00:00 ในเวลาท้องถิ่น (UTC+7)
        newDateTime = DateTime(
          picked.year,
          picked.month,
          picked.day,
          0,
          0,
          0,
          0,
          0,
        );
      } else {
        // ใช้เวลาเดิม
        final currentDate = DateTime.parse(doc.startDateTime!).toLocal();
        newDateTime = DateTime(
          picked.year,
          picked.month,
          picked.day,
          currentDate.hour,
          currentDate.minute,
          currentDate.second,
          0,
          0,
        );
      }

      // แปลงเป็น UTC สำหรับเก็บใน MongoDB
      final newDateTimeUtc = newDateTime.toUtc();

      // ถ้ามี end date อยู่แล้ว ให้ส่งทั้ง start และ end
      // ถ้าไม่มี end date ให้ส่งเฉพาะ start โดยใช้วันเดียวกัน + 1 ปี
      final endDate =
          doc.endDateTime ??
          newDateTimeUtc.add(const Duration(days: 365)).toIso8601String();

      if (kDebugMode) {
        AppLogger.debug(
          'Start date selected: ${picked.year}-${picked.month}-${picked.day}',
        );
        AppLogger.debug('Date changed: $dateChanged');
        AppLogger.debug('Start datetime Local: ${newDateTime.toString()}');
        AppLogger.debug(
          'Start datetime UTC: ${newDateTimeUtc.toIso8601String()}',
        );
        AppLogger.debug('End date: $endDate');
      }

      _updateDocumentDateRange(doc, newDateTimeUtc.toIso8601String(), endDate);
    }
  }

  // Select end date
  Future<void> _selectEndDate(DocumentModel doc) async {
    DateTime initialDate = DateTime.now();

    // ตรวจสอบว่ามีข้อมูล end date หรือไม่
    if (doc.endDateTime != null && doc.endDateTime!.isNotEmpty) {
      try {
        final parsedDate = DateTime.parse(doc.endDateTime!);
        // ถ้าวันที่มากกว่าหรือเท่ากับปี 2000 ให้ใช้วันที่นั้น ไม่เช่นนั้นใช้วันที่ปัจจุบัน
        if (parsedDate.year >= 2000) {
          initialDate = parsedDate;
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error parsing end date: $e');
        }
      }
    }

    final DateTime? picked = await showDialog<DateTime>(
      context: context,
      builder: (BuildContext context) {
        return Dialog(
          child: Container(
            width: 320,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(8),
              color: global.theme.cardColor,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: primaryColor,
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(8),
                      topRight: Radius.circular(8),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.calendar_today, color: Colors.white),
                      SizedBox(width: 8),
                      Text(
                        global.language('kb_select_end_date'),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
                // Calendar
                // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                CustomDatePickerDialog(
                  initialDate: initialDate,
                  firstDate: DateTime(2020),
                  lastDate: DateTime(2100),
                ),
              ],
            ),
          ),
        );
      },
    );

    if (picked != null) {
      // ตรวจสอบว่าวันที่เปลี่ยนแปลงหรือไม่
      bool dateChanged = true;
      if (doc.endDateTime != null && doc.endDateTime!.isNotEmpty) {
        try {
          final currentDate = DateTime.parse(doc.endDateTime!).toLocal();
          dateChanged =
              currentDate.year != picked.year ||
              currentDate.month != picked.month ||
              currentDate.day != picked.day;
        } catch (e) {
          dateChanged = true;
        }
      }

      // ถ้าวันที่เปลี่ยนแปลง ให้ตั้งเวลา default เป็น 23:59:59 (Local Time UTC+7)
      // ถ้าไม่เปลี่ยน ให้ใช้เวลาเดิม
      DateTime newDateTime;
      if (dateChanged) {
        // ตั้งเวลาเป็น 23:59:59 ในเวลาท้องถิ่น (UTC+7)
        newDateTime = DateTime(
          picked.year,
          picked.month,
          picked.day,
          23,
          59,
          59,
          999,
          999,
        );
      } else {
        // ใช้เวลาเดิม
        final currentDate = DateTime.parse(doc.endDateTime!).toLocal();
        newDateTime = DateTime(
          picked.year,
          picked.month,
          picked.day,
          currentDate.hour,
          currentDate.minute,
          currentDate.second,
          999,
          999,
        );
      }

      // แปลงเป็น UTC สำหรับเก็บใน MongoDB
      final newDateTimeUtc = newDateTime.toUtc();

      // ถ้ามี start date อยู่แล้ว ให้ส่งทั้ง start และ end
      // ถ้าไม่มี start date ให้ส่งเฉพาะ end โดยใช้วันเดียวกัน - 1 ปี
      final startDate =
          doc.startDateTime ??
          newDateTimeUtc.subtract(const Duration(days: 365)).toIso8601String();

      if (kDebugMode) {
        AppLogger.debug(
          'End date selected: ${picked.year}-${picked.month}-${picked.day}',
        );
        AppLogger.debug('Date changed: $dateChanged');
        AppLogger.debug('End datetime Local: ${newDateTime.toString()}');
        AppLogger.debug(
          'End datetime UTC: ${newDateTimeUtc.toIso8601String()}',
        );
        AppLogger.debug('Start date: $startDate');
      }

      _updateDocumentDateRange(
        doc,
        startDate,
        newDateTimeUtc.toIso8601String(),
      );
    }
  }

  // Select start time
  Future<void> _selectStartTime(DocumentModel doc) async {
    // Parse current time (แปลงจาก UTC เป็น Local Time UTC+7)
    TimeOfDay initialTime = const TimeOfDay(hour: 0, minute: 0);
    if (doc.startDateTime != null && doc.startDateTime!.isNotEmpty) {
      try {
        final parsedDate = DateTime.parse(doc.startDateTime!);
        if (parsedDate.year >= 2000) {
          final localTime = parsedDate.toLocal();
          initialTime = TimeOfDay(
            hour: localTime.hour,
            minute: localTime.minute,
          );
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error parsing start time: $e');
        }
      }
    }

    final TimeOfDay? picked = await showDialog<TimeOfDay>(
      context: context,
      barrierColor: Colors.black.withValues(alpha: 0.4),
      builder: (BuildContext context) {
        return Center(
          child: Material(
            color: Colors.transparent,
            child: CustomTimePickerDialog(
              initialTime: initialTime,
              languageCode: global.getBranchLanguage(),
            ),
          ),
        );
      },
    );

    if (picked != null) {
      // ดึงวันที่จาก start date หรือใช้วันที่ปัจจุบัน
      DateTime baseDate = DateTime.now();
      if (doc.startDateTime != null && doc.startDateTime!.isNotEmpty) {
        try {
          final parsedDate = DateTime.parse(doc.startDateTime!);
          if (parsedDate.year >= 2000) {
            baseDate = parsedDate.toLocal();
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error parsing base date: $e');
          }
        }
      }

      // สร้าง DateTime ใหม่ด้วยเวลาที่เลือก (Local Time UTC+7)
      final newDateTime = DateTime(
        baseDate.year,
        baseDate.month,
        baseDate.day,
        picked.hour,
        picked.minute,
        0,
        0,
        0,
      );

      // แปลงเป็น UTC สำหรับส่งไปเก็บใน MongoDB
      final newDateTimeUtc = newDateTime.toUtc();

      final endDate =
          doc.endDateTime ??
          newDateTimeUtc.add(const Duration(days: 365)).toIso8601String();

      if (kDebugMode) {
        AppLogger.debug('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
        AppLogger.debug('📅 START TIME SELECTION');
        AppLogger.debug(
          'Selected time: ${picked.format(context)} (Local UTC+7)',
        );
        AppLogger.debug('Base date (Local): ${baseDate.toString()}');
        AppLogger.debug(
          'New DateTime (Local UTC+7): ${newDateTime.toString()}',
        );
        AppLogger.debug(
          'New DateTime (UTC+0): ${newDateTimeUtc.toIso8601String()}',
        );
        AppLogger.debug('End date: $endDate');
        AppLogger.debug('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
      }

      _updateDocumentDateRange(doc, newDateTimeUtc.toIso8601String(), endDate);
    }
  }

  // Select end time
  Future<void> _selectEndTime(DocumentModel doc) async {
    // Parse current time (แปลงจาก UTC เป็น Local Time UTC+7)
    TimeOfDay initialTime = const TimeOfDay(hour: 23, minute: 59);
    if (doc.endDateTime != null && doc.endDateTime!.isNotEmpty) {
      try {
        final parsedDate = DateTime.parse(doc.endDateTime!);
        if (parsedDate.year >= 2000) {
          final localTime = parsedDate.toLocal();
          initialTime = TimeOfDay(
            hour: localTime.hour,
            minute: localTime.minute,
          );
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error parsing end time: $e');
        }
      }
    }

    final TimeOfDay? picked = await showDialog<TimeOfDay>(
      context: context,
      barrierColor: Colors.black.withValues(alpha: 0.4),
      builder: (BuildContext context) {
        return Center(
          child: Material(
            color: Colors.transparent,
            child: CustomTimePickerDialog(
              initialTime: initialTime,
              languageCode: global.getBranchLanguage(),
            ),
          ),
        );
      },
    );

    if (picked != null) {
      // ดึงวันที่จาก end date หรือใช้วันที่ปัจจุบัน
      DateTime baseDate = DateTime.now();
      if (doc.endDateTime != null && doc.endDateTime!.isNotEmpty) {
        try {
          final parsedDate = DateTime.parse(doc.endDateTime!);
          if (parsedDate.year >= 2000) {
            baseDate = parsedDate.toLocal();
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error parsing base date: $e');
          }
        }
      }

      // สร้าง DateTime ใหม่ด้วยเวลาที่เลือก (Local Time UTC+7)
      final newDateTime = DateTime(
        baseDate.year,
        baseDate.month,
        baseDate.day,
        picked.hour,
        picked.minute,
        59,
        999,
        999,
      );

      // แปลงเป็น UTC สำหรับส่งไปเก็บใน MongoDB
      final newDateTimeUtc = newDateTime.toUtc();

      final startDate =
          doc.startDateTime ??
          newDateTimeUtc.subtract(const Duration(days: 365)).toIso8601String();

      if (kDebugMode) {
        AppLogger.debug('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
        AppLogger.debug('📅 END TIME SELECTION');
        AppLogger.debug(
          'Selected time: ${picked.format(context)} (Local UTC+7)',
        );
        AppLogger.debug('Base date (Local): ${baseDate.toString()}');
        AppLogger.debug(
          'New DateTime (Local UTC+7): ${newDateTime.toString()}',
        );
        AppLogger.debug(
          'New DateTime (UTC+0): ${newDateTimeUtc.toIso8601String()}',
        );
        AppLogger.debug('Start date: $startDate');
        AppLogger.debug('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
      }

      _updateDocumentDateRange(
        doc,
        startDate,
        newDateTimeUtc.toIso8601String(),
      );
    }
  }

  // Update document date range (ใช้ API endpoint update-schedule)
  Future<void> _updateDocumentDateRange(
    DocumentModel doc,
    String? startDate,
    String? endDate,
  ) async {
    try {
      final shopId = global.getShopId();

      final requestBody = {
        'shop_id': shopId,
        'branch_id': doc.branchId,
        'filename': doc.filename,
        if (startDate != null) 'start_date_time': startDate,
        if (endDate != null) 'end_date_time': endDate,
      };

      if (kDebugMode) {
        AppLogger.debug('═══════════════════════════════════════');
        AppLogger.debug('Updating date range for: ${doc.filename}');
        AppLogger.debug('Start DateTime (UTC+0): $startDate');
        AppLogger.debug('End DateTime (UTC+0): $endDate');
        if (startDate != null) {
          AppLogger.debug(
            'Start DateTime parsed: ${DateTime.parse(startDate)}',
          );
        }
        if (endDate != null) {
          AppLogger.debug('End DateTime parsed: ${DateTime.parse(endDate)}');
        }
        AppLogger.debug('Request body: ${jsonEncode(requestBody)}');
        AppLogger.debug('═══════════════════════════════════════');
      }

      final response = await http.post(
        Uri.parse('$apiBaseUrl/update-schedule'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode(requestBody),
      );

      if (kDebugMode) {
        AppLogger.debug('Response status: ${response.statusCode}');
        AppLogger.debug('Response body: ${response.body}');
      }

      if (response.statusCode == 200) {
        // อัปเดต state ทันที (ไม่ต้องโหลดข้อมูลใหม่เพื่อป้องกันจอกระพริบ)
        context.read<KnowledgeBaseCubit>().updateDocumentDateRange(
          doc.filename,
          startDate,
          endDate,
        );
        final message = startDate == null && endDate == null
            ? global.language('kb_date_cleared')
            : global.language('kb_date_set_success');
        _showSnackBar(message);
      } else {
        _showSnackBar(global.language('kb_date_set_error'), isError: true);
      }
    } catch (e) {
      _showSnackBar(
        '${global.language('kb_error_occurred')}: $e',
        isError: true,
      );
    }
  }

  void _showSnackBar(String message, {bool isError = false}) {
    if (isError) {
      global.showErrorSnackBar(context, message);
    } else {
      global.showSuccessSnackBar(context, message);
    }
  }

  String _getContentType(String extension) {
    switch (extension.toLowerCase()) {
      case 'pdf':
        return 'application/pdf';
      case 'xlsx':
      case 'xls':
        return 'application/vnd.ms-excel';
      case 'csv':
        return 'text/csv';
      case 'txt':
        return 'text/plain';
      case 'md':
        return 'text/markdown';
      case 'doc':
      case 'docx':
        return 'application/msword';
      default:
        return 'application/octet-stream';
    }
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

  String _formatDateTime(String? dateTimeStr) {
    if (dateTimeStr == null || dateTimeStr.isEmpty) return '';
    try {
      final dt = DateTime.parse(dateTimeStr).toLocal(); // แปลงเป็น UTC+7
      return global.formatThaiDateTime(dateTime: dt, showTime: true);
    } catch (e) {
      return dateTimeStr;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.surfaceColor,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Row(
          children: [
            const Icon(Icons.school, size: 24),
            SizedBox(width: 8),
            Text(
              global.language('kb_title'),
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
          ],
        ),
      ),
      body: BlocListener<CompanyBranchBloc, CompanyBranchState>(
        listener: (context, state) {
          if (state is CompanyBranchLoadSuccess) {
            _branches = state.companyBranch;
          }
        },
        child: LayoutBuilder(
          builder: (context, constraints) {
            final leftWidth = constraints.maxWidth * _leftPanelWidth;
            final rightWidth = constraints.maxWidth * (1 - _leftPanelWidth);

            return Row(
              children: [
                // Left Panel - Document Management
                SizedBox(
                  width: leftWidth,
                  child: Container(
                    color: global.theme.cardColor,
                    child: Column(
                      children: [
                        _buildDocumentHeader(),
                        _buildBranchSelector(),
                        _buildActionButtons(),
                        Expanded(child: _buildDocumentList()),
                      ],
                    ),
                  ),
                ),

                // Divider/Resizer
                GestureDetector(
                  onHorizontalDragStart: (_) {
                    setState(() => _isDragging = true);
                  },
                  onHorizontalDragUpdate: (details) {
                    setState(() {
                      final newWidth =
                          _leftPanelWidth +
                          (details.delta.dx / constraints.maxWidth);
                      _leftPanelWidth = newWidth.clamp(0.2, 0.8);
                    });
                  },
                  onHorizontalDragEnd: (_) {
                    setState(() => _isDragging = false);
                  },
                  child: MouseRegion(
                    cursor: SystemMouseCursors.resizeColumn,
                    child: Container(
                      width: 8,
                      color: _isDragging
                          ? primaryColor.withValues(alpha: 0.3)
                          : global.theme.dividerBorderColor,
                      child: Center(
                        child: Container(
                          width: 2,
                          color: _isDragging ? primaryColor : global.theme.iconSecondaryColor,
                        ),
                      ),
                    ),
                  ),
                ),

                // Right Panel - Chatbot
                SizedBox(
                  width: rightWidth - 8,
                  child: ChatbotPanel(
                    welcomeMessage: global.language('kb_chatbot_welcome'),
                  ),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  // Document Header
  Widget _buildDocumentHeader() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.primaryColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          const Icon(Icons.folder_open, color: Colors.white, size: 24),
          SizedBox(width: 12),
          Text(
            global.language('kb_documents_knowledge_base'),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  // Branch Selector
  Widget _buildBranchSelector() {
    return BlocBuilder<CompanyBranchBloc, CompanyBranchState>(
      builder: (context, state) {
        // แสดง loading indicator ถ้ายังโหลดข้อมูลสาขาไม่เสร็จ
        if (state is CompanyBranchInProgress) {
          return Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor)),
            ),
            child: const Center(child: CircularProgressIndicator()),
          );
        }

        return Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('kb_select_branch_description'),
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              _branches.isEmpty
                  ? Text(
                      global.language('kb_loading_branch_data'),
                      style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
                    )
                  : Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        // All Branches Button
                        ChoiceChip(
                          label: Text(global.language('kb_all_branches')),
                          selected: _showAllBranches,
                          onSelected: (selected) {
                            if (selected) {
                              setState(() {
                                _showAllBranches = true;
                                _selectedBranchId = null;
                              });
                              _loadDocuments();
                            }
                          },
                          selectedColor: primaryColor,
                          labelStyle: TextStyle(
                            color: _showAllBranches
                                ? Colors.white
                                : global.theme.textColor,
                          ),
                        ),
                        // Individual Branch Buttons
                        ..._branches.map((branch) {
                          final isSelected =
                              !_showAllBranches &&
                              _selectedBranchId == branch.guidfixed;
                          return ChoiceChip(
                            label: Text(global.packName(branch.names)),
                            selected: isSelected,
                            onSelected: (selected) {
                              if (selected) {
                                setState(() {
                                  _showAllBranches = false;
                                  _selectedBranchId = branch.guidfixed;
                                });
                                _loadDocuments();
                              }
                            },
                            selectedColor: primaryColor,
                            labelStyle: TextStyle(
                              color: isSelected ? Colors.white : global.theme.textColor,
                            ),
                          );
                        }),
                      ],
                    ),
            ],
          ),
        );
      },
    );
  }

  // Action Buttons
  Widget _buildActionButtons() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor)),
      ),
      child: Row(
        children: [
          // Upload Button
          Expanded(
            child: ElevatedButton.icon(
              onPressed: _uploadDocument,
              icon: Icon(Icons.upload_file, size: 18),
              label: Text(global.language('kb_upload_document')),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.green,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
            ),
          ),
          const SizedBox(width: 8),
          // Refresh Button
          IconButton(
            onPressed: _loadDocuments,
            icon: Icon(Icons.refresh),
            tooltip: global.language('kb_refresh'),
          ),
        ],
      ),
    );
  }

  // Document List
  Widget _buildDocumentList() {
    return BlocBuilder<KnowledgeBaseCubit, KnowledgeBaseState>(
      builder: (context, state) {
        if (state.isLoading) {
          return const Center(child: CircularProgressIndicator());
        }

        if (state.documents.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.folder_open, size: 64, color: global.theme.iconSecondaryColor),
                SizedBox(height: 16),
                Text(
                  global.language('kb_no_documents'),
                  style: TextStyle(fontSize: 16, color: global.theme.iconSecondaryColor),
                ),
                SizedBox(height: 8),
                Text(
                  global.language('kb_click_upload_to_add'),
                  style: TextStyle(fontSize: 14, color: global.theme.iconSecondaryColor),
                ),
              ],
            ),
          );
        }

        return ListView.builder(
          controller: _documentScrollController,
          padding: const EdgeInsets.all(8),
          itemCount: state.documents.length,
          itemBuilder: (context, index) {
            final doc = state.documents[index];
            return _buildDocumentCard(doc);
          },
        );
      },
    );
  }

  // Document Card
  Widget _buildDocumentCard(DocumentModel doc) {
    final isActive = doc.status; // true = เปิดใช้งาน, false = ปิดใช้งาน

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: isActive
                ? Colors.green.withValues(alpha: 0.1)
                : global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            _getFileIcon(doc.filename),
            color: isActive ? Colors.green : global.theme.iconSecondaryColor,
          ),
        ),
        title: Text(
          doc.filename,
          style: TextStyle(
            fontWeight: FontWeight.bold,
            color: isActive ? global.theme.textColor : global.theme.textSecondaryColor,
          ),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('${global.language('kb_size')} ${_formatFileSize(doc.size)}'),
            Text(
              '${global.language('kb_uploaded_at')} ${_formatDateTime(doc.uploadedAt)}',
            ),
            if (doc.uploadedBy.isNotEmpty)
              Text('${global.language('kb_uploader')} ${doc.uploadedBy}'),
            if (doc.description.isNotEmpty)
              Text('${global.language('kb_description')} ${doc.description}'),
            const SizedBox(height: 4),
            // แสดงสถานะเอกสาร
            Row(
              children: [
                // สถานะ ปิด/เปิด
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 6,
                    vertical: 2,
                  ),
                  decoration: BoxDecoration(
                    color: isActive ? Colors.green : global.theme.iconSecondaryColor,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    '${global.language('kb_working_status')} ${isActive ? global.language('kb_on') : global.language('kb_off')}',
                    style: const TextStyle(color: Colors.white, fontSize: 10),
                  ),
                ),
                const SizedBox(width: 8),
                // แสดงเงื่อนไขวันที่
                Text(
                  _getDocumentDateRangeStatus(doc),
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.blue[700],
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(width: 4),
                Text(
                  'v${doc.version}',
                  style: TextStyle(fontSize: 10, color: global.theme.iconSecondaryColor),
                ),
              ],
            ),
            const SizedBox(height: 8),
            // ปุ่มต่างๆ
            // แถวที่ 1: ปุ่ม All Day
            Row(
              children: [
                SizedBox(
                  height: 28,
                  child: OutlinedButton.icon(
                    onPressed: () => _toggleAllDayMode(doc),
                    icon: Icon(
                      doc.allDay ? Icons.all_inclusive : Icons.schedule,
                      size: 14,
                    ),
                    label: Text(
                      doc.allDay
                          ? global.language('kb_work_all_time')
                          : global.language('kb_set_time_period'),
                      style: const TextStyle(fontSize: 11),
                    ),
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                      side: BorderSide(
                        color: doc.allDay ? Colors.blue : Colors.orange,
                        width: 2,
                      ),
                    ),
                  ),
                ),
              ],
            ),
            if (!doc.allDay) ...[
              const SizedBox(height: 4),
              // แถวที่ 2: กลุ่มวันที่เริ่มต้น + เวลา
              Row(
                children: [
                  // ป้ายกำกับ "เริ่ม:"
                  SizedBox(
                    width: 40,
                    child: Text(
                      global.language('kb_start'),
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Colors.blue[700],
                      ),
                    ),
                  ),
                  // ปุ่มจากวันที่
                  SizedBox(
                    height: 28,
                    child: OutlinedButton.icon(
                      onPressed: () => _selectStartDate(doc),
                      icon: Icon(Icons.event, size: 14),
                      label: Text(
                        doc.startDateTime != null &&
                                DateTime.parse(doc.startDateTime!).year >= 2000
                            ? _formatThaiDate(
                                DateTime.parse(doc.startDateTime!).toLocal(),
                              )
                            : global.language('kb_select_date'),
                        style: const TextStyle(fontSize: 11),
                      ),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                      ),
                    ),
                  ),
                  const SizedBox(width: 4),
                  // ปุ่มเวลาเริ่มต้น
                  SizedBox(
                    height: 28,
                    child: OutlinedButton.icon(
                      onPressed: () => _selectStartTime(doc),
                      icon: const Icon(Icons.access_time, size: 14),
                      label: Text(
                        doc.startDateTime != null &&
                                DateTime.parse(doc.startDateTime!).year >= 2000
                            ? _formatTime(
                                DateTime.parse(doc.startDateTime!).toLocal(),
                              )
                            : '00:00',
                        style: const TextStyle(fontSize: 11),
                      ),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              // แถวที่ 3: กลุ่มวันที่สิ้นสุด + เวลา
              Row(
                children: [
                  // ป้ายกำกับ "ถึง:"
                  SizedBox(
                    width: 40,
                    child: Text(
                      '${global.language('kb_to')}:',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Colors.blue[700],
                      ),
                    ),
                  ),
                  // ปุ่มถึงวันที่
                  SizedBox(
                    height: 28,
                    child: OutlinedButton.icon(
                      onPressed: () => _selectEndDate(doc),
                      icon: Icon(Icons.event, size: 14),
                      label: Text(
                        doc.endDateTime != null &&
                                DateTime.parse(doc.endDateTime!).year >= 2000
                            ? _formatThaiDate(
                                DateTime.parse(doc.endDateTime!).toLocal(),
                              )
                            : global.language('kb_select_date'),
                        style: const TextStyle(fontSize: 11),
                      ),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                      ),
                    ),
                  ),
                  const SizedBox(width: 4),
                  // ปุ่มเวลาสิ้นสุด
                  SizedBox(
                    height: 28,
                    child: OutlinedButton.icon(
                      onPressed: () => _selectEndTime(doc),
                      icon: const Icon(Icons.access_time, size: 14),
                      label: Text(
                        doc.endDateTime != null &&
                                DateTime.parse(doc.endDateTime!).year >= 2000
                            ? _formatTime(
                                DateTime.parse(doc.endDateTime!).toLocal(),
                              )
                            : '23:59',
                        style: const TextStyle(fontSize: 11),
                      ),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
            ],
            // แถวที่ 4: ปุ่มจัดการเอกสาร
            Row(
              children: [
                const Spacer(),
                const SizedBox(width: 8),
                // ปุ่มปิด/เปิด
                SizedBox(
                  height: 28,
                  child: ElevatedButton.icon(
                    onPressed: () => _toggleDocumentStatus(doc),
                    icon: Icon(
                      isActive ? Icons.pause : Icons.play_arrow,
                      size: 14,
                    ),
                    label: Text(
                      isActive
                          ? global.language('kb_off')
                          : global.language('kb_on'),
                      style: const TextStyle(fontSize: 11),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: isActive ? Colors.orange : Colors.green,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                // ปุ่มลบ
                SizedBox(
                  height: 28,
                  child: IconButton(
                    icon: const Icon(Icons.delete, size: 18),
                    color: Colors.red,
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                    onPressed: () => _deleteDocument(doc),
                  ),
                ),
              ],
            ),
          ],
        ),
        trailing: null,
      ),
    );
  }

  IconData _getFileIcon(String filename) {
    final ext = filename.split('.').last.toLowerCase();
    switch (ext) {
      case 'pdf':
        return Icons.picture_as_pdf;
      case 'xlsx':
      case 'xls':
      case 'csv':
        return Icons.table_chart;
      case 'txt':
      case 'md':
        return Icons.description;
      case 'doc':
      case 'docx':
        return Icons.article;
      default:
        return Icons.insert_drive_file;
    }
  }
}
