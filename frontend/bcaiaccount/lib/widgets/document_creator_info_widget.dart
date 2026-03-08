import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/services/datahistory_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget แสดงข้อมูลผู้สร้างและผู้แก้ไขล่าสุดของเอกสาร
/// โดยดึงข้อมูลจาก Data History API (MongoDB)
/// รองรับ fallback จากข้อมูลเอกสาร — กรณี Data History เป็น "system" หรือไม่มีข้อมูล
class DocumentCreatorInfoWidget extends StatefulWidget {
  final String docNo;

  /// fallback: รหัสผู้สร้างจากเอกสาร (ใช้เมื่อ Data History เป็น "system" หรือไม่มี)
  final String? fallbackCreatorCode;

  /// fallback: ชื่อผู้สร้างจากเอกสาร
  final String? fallbackCreatorName;

  /// fallback: วันที่สร้างเอกสาร
  final String? fallbackCreatedAt;

  const DocumentCreatorInfoWidget({
    super.key,
    required this.docNo,
    this.fallbackCreatorCode,
    this.fallbackCreatorName,
    this.fallbackCreatedAt,
  });

  @override
  State<DocumentCreatorInfoWidget> createState() =>
      _DocumentCreatorInfoWidgetState();
}

class _DocumentCreatorInfoWidgetState extends State<DocumentCreatorInfoWidget> {
  bool _isLoading = true;
  DataHistoryModel? _creatorInfo; // ข้อมูลการสร้าง (action=create)
  DataHistoryModel? _lastModifierInfo; // ข้อมูลการแก้ไขล่าสุด (action=update)

  @override
  void initState() {
    super.initState();
    _loadCreatorInfo();
  }

  @override
  void didUpdateWidget(DocumentCreatorInfoWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.docNo != widget.docNo) {
      _loadCreatorInfo();
    }
  }

  Future<void> _loadCreatorInfo() async {
    if (widget.docNo.isEmpty) {
      setState(() {
        _isLoading = false;
      });
      return;
    }

    if (!mounted) return;
    setState(() {
      _isLoading = true;
    });

    try {
      final historyList = await DataHistoryApiService.getPOHistory(widget.docNo);

      if (historyList.isNotEmpty) {
        // หา record แรกที่เป็น create (เรียงจากล่าสุดก่อน ต้องหาตัวท้ายสุด)
        final creatorRecord = historyList.reversed.firstWhere(
          (h) => h.action == 'create',
          orElse: () => historyList.last,
        );

        // หา record ล่าสุดที่เป็น update (ถ้ามี)
        DataHistoryModel? lastUpdateRecord;
        try {
          lastUpdateRecord = historyList.firstWhere(
            (h) => h.action == 'update',
          );
        } catch (_) {
          // ไม่มี update record
        }

        if (!mounted) return;
        setState(() {
          _creatorInfo = creatorRecord;
          _lastModifierInfo = lastUpdateRecord;
          _isLoading = false;
        });
      } else {
        if (!mounted) return;
        setState(() {
          _isLoading = false;
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
      });
    }
  }

  /// ตรวจสอบว่า user เป็น "system" หรือไม่
  bool _isSystemUser(String userCode) {
    return userCode.isEmpty || userCode.toLowerCase() == 'system';
  }

  @override
  Widget build(BuildContext context) {
    // ไม่แสดงถ้าไม่มี docNo หรือยังไม่มีข้อมูล
    if (widget.docNo.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.06),
            blurRadius: 8,
            offset: const Offset(0, 2),
          )
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  const Color(0xFF667eea).withValues(alpha: 0.1),
                  const Color(0xFF764ba2).withValues(alpha: 0.05),
                ],
                begin: Alignment.centerLeft,
                end: Alignment.centerRight,
              ),
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(12),
                topRight: Radius.circular(12),
              ),
            ),
            child: Row(
              children: [
                const Icon(Icons.person_outline, color: Color(0xFF667eea), size: 20),
                const SizedBox(width: 8),
                Text(
                  _getLabel("document_audit_info"),
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF667eea),
                  ),
                ),
              ],
            ),
          ),
          // Content
          Padding(
            padding: const EdgeInsets.all(14),
            child: _buildContent(),
          ),
        ],
      ),
    );
  }

  Widget _buildContent() {
    if (_isLoading) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(16.0),
          child: SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
        ),
      );
    }

    // ถ้าไม่มี Data History → ใช้ fallback จากเอกสาร
    if (_creatorInfo == null) {
      if (_hasFallbackCreator()) {
        return _buildFallbackContent();
      }
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 8.0),
        child: Text(
          _getLabel("no_audit_info"),
          style: TextStyle(color: Colors.grey[600], fontSize: 12),
        ),
      );
    }

    // ถ้า Data History มีผู้สร้างเป็น "system" แต่เอกสารมีข้อมูลจริง → ใช้ fallback
    String creatorUserName = _creatorInfo!.userName.isNotEmpty
        ? _creatorInfo!.userName
        : _creatorInfo!.userCode;
    String creatorUserCode = _creatorInfo!.userCode;

    if (_isSystemUser(creatorUserCode) && _hasFallbackCreator()) {
      creatorUserName = widget.fallbackCreatorName ?? widget.fallbackCreatorCode ?? 'System';
      creatorUserCode = widget.fallbackCreatorCode ?? '';
    }

    return Column(
      children: [
        // ผู้สร้างเอกสาร
        _buildInfoRow(
          icon: Icons.person_add,
          iconColor: Colors.green,
          label: _getLabel("creator"),
          userName: creatorUserName,
          userCode: creatorUserCode,
          dateTime: _creatorInfo!.timestamp,
        ),

        // เส้นคั่น
        if (_lastModifierInfo != null) ...[
          const Divider(height: 16),
          // ผู้แก้ไขล่าสุด
          _buildInfoRow(
            icon: Icons.edit,
            iconColor: Colors.blue,
            label: _getLabel("last_modifier"),
            userName: _lastModifierInfo!.userName.isNotEmpty
                ? _lastModifierInfo!.userName
                : _lastModifierInfo!.userCode,
            userCode: _lastModifierInfo!.userCode,
            dateTime: _lastModifierInfo!.timestamp,
          ),
        ],
      ],
    );
  }

  /// ตรวจสอบว่ามี fallback creator จากเอกสารหรือไม่
  bool _hasFallbackCreator() {
    return (widget.fallbackCreatorCode != null && widget.fallbackCreatorCode!.isNotEmpty) ||
           (widget.fallbackCreatorName != null && widget.fallbackCreatorName!.isNotEmpty);
  }

  /// สร้าง content จาก fallback (กรณีไม่มี Data History เลย)
  Widget _buildFallbackContent() {
    DateTime? createdAt;
    if (widget.fallbackCreatedAt != null && widget.fallbackCreatedAt!.isNotEmpty) {
      createdAt = DateTime.tryParse(widget.fallbackCreatedAt!);
    }

    return _buildInfoRow(
      icon: Icons.person_add,
      iconColor: Colors.green,
      label: _getLabel("creator"),
      userName: widget.fallbackCreatorName ?? widget.fallbackCreatorCode ?? '',
      userCode: widget.fallbackCreatorCode ?? '',
      dateTime: createdAt ?? DateTime.now(),
    );
  }

  Widget _buildInfoRow({
    required IconData icon,
    required Color iconColor,
    required String label,
    required String userName,
    required String userCode,
    required DateTime dateTime,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Icon
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: iconColor.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: iconColor, size: 18),
        ),
        const SizedBox(width: 12),
        // Info
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Label
              Text(
                label,
                style: TextStyle(
                  fontSize: 11,
                  color: Colors.grey[600],
                  fontWeight: FontWeight.w500,
                ),
              ),
              const SizedBox(height: 2),
              // User name
              Text(
                userName,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Colors.black87,
                ),
              ),
              // User code (ถ้าต่างจาก name)
              if (userCode.isNotEmpty && userCode != userName)
                Text(
                  '($userCode)',
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.grey[500],
                  ),
                ),
            ],
          ),
        ),
        // Date/Time
        Column(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              _formatDate(dateTime),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: Colors.black54,
              ),
            ),
            Text(
              _formatTime(dateTime),
              style: TextStyle(
                fontSize: 11,
                color: Colors.grey[500],
              ),
            ),
          ],
        ),
      ],
    );
  }

  String _formatDate(DateTime dateTime) {
    final localDateTime = dateTime.toLocal();
    if (global.profileData.yeartype == "buddhist") {
      return global.dateTimeBuddhist(localDateTime, format: global.DateTimeFormatEnum.dateDay);
    }
    return DateFormat('dd/MM/yyyy').format(localDateTime);
  }

  String _formatTime(DateTime dateTime) {
    final localDateTime = dateTime.toLocal();
    return DateFormat('HH:mm').format(localDateTime);
  }

  /// Helper method สำหรับดึงข้อความ label พร้อม fallback ภาษาไทย
  String _getLabel(String key) {
    // Fallback labels ภาษาไทย
    final Map<String, String> fallbackLabels = {
      'document_audit_info': global.language('document_audit_info'),
      'creator': global.language('document_creator'),
      'last_modifier': global.language('last_modifier'),
      'no_audit_info': global.language('no_audit_info'),
    };

    // ลองดึงจาก language system ก่อน
    final translated = global.language(key);

    // ถ้าได้ค่าเดิมกลับมา (ไม่มี translation) ใช้ fallback
    if (translated == key) {
      return fallbackLabels[key] ?? key;
    }

    return translated;
  }
}
