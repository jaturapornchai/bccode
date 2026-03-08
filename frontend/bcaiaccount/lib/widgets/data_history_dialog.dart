import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/services/datahistory_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// Dialog สำหรับแสดงประวัติการแก้ไขข้อมูล
class DataHistoryDialog extends StatefulWidget {
  final String docNo;
  final String title;
  // ข้อมูล creator/modifier จาก TransactionModel (ใช้แสดงเมื่อไม่มี datahistory)
  final String? creatorCode;
  final String? creatorName;
  final String? createdAt;
  final String? modifierCode;
  final String? modifierName;
  final String? modifiedAt;

  const DataHistoryDialog({
    super.key,
    required this.docNo,
    this.title = '',
    this.creatorCode,
    this.creatorName,
    this.createdAt,
    this.modifierCode,
    this.modifierName,
    this.modifiedAt,
  });

  @override
  State<DataHistoryDialog> createState() => _DataHistoryDialogState();
}

class _DataHistoryDialogState extends State<DataHistoryDialog> {
  List<DataHistoryModel> _historyList = [];
  bool _isLoading = true;
  String? _errorMessage;
  int? _expandedIndex;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final history = await DataHistoryApiService.getPOHistory(widget.docNo);
      setState(() {
        _historyList = history;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = 'เกิดข้อผิดพลาด: $e';
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        height: MediaQuery.of(context).size.height * 0.8,
        constraints: const BoxConstraints(maxWidth: 900, maxHeight: 700),
        child: Column(
          children: [
            // Header
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).primaryColor,
                borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(4),
                  topRight: Radius.circular(4),
                ),
              ),
              child: Row(
                children: [
                  const Icon(Icons.history, color: Colors.white),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '${widget.title.isNotEmpty ? widget.title : global.language("edit_history")} - ${widget.docNo}',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.refresh, color: Colors.white),
                    onPressed: _loadHistory,
                    tooltip: global.language("refresh"),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.white),
                    onPressed: () => Navigator.of(context).pop(),
                    tooltip: global.language("close"),
                  ),
                ],
              ),
            ),
            // Content
            Expanded(
              child: _buildContent(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContent() {
    if (_isLoading) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const CircularProgressIndicator(),
            SizedBox(height: 16),
            Text(global.language("loading_data")),
          ],
        ),
      );
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 48, color: Colors.red),
            const SizedBox(height: 16),
            Text(_errorMessage!, style: const TextStyle(color: Colors.red)),
            SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadHistory,
              child: Text(global.language("try_again")),
            ),
          ],
        ),
      );
    }

    if (_historyList.isEmpty) {
      // ถ้ามีข้อมูล creator จาก pgsql ให้แสดง (modifier อยู่ใน datahistory)
      final hasCreatorInfo = widget.creatorCode?.isNotEmpty == true ||
                             widget.creatorName?.isNotEmpty == true;

      if (hasCreatorInfo) {
        return SingleChildScrollView(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ข้อความแจ้ง
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.blue[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.blue[200]!),
                ),
                child: Row(
                  children: [
                    Icon(Icons.info_outline, color: Colors.blue[700]),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        global.language("no_detailed_history_show_creator"),
                        style: const TextStyle(fontSize: 13),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              // ข้อมูลผู้สร้าง
              _buildCreatorModifierCard(
                icon: Icons.person_add,
                iconColor: Colors.green,
                title: global.language("document_creator"),
                code: widget.creatorCode ?? '-',
                name: widget.creatorName ?? '-',
                dateTime: widget.createdAt,
              ),
            ],
          ),
        );
      }

      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.inbox, size: 48, color: Colors.grey),
            SizedBox(height: 16),
            Text(global.language("no_edit_history_found"), style: TextStyle(color: Colors.grey)),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(8),
      itemCount: _historyList.length,
      itemBuilder: (context, index) {
        final item = _historyList[index];
        final isExpanded = _expandedIndex == index;

        return Card(
          margin: const EdgeInsets.symmetric(vertical: 4),
          child: Column(
            children: [
              ListTile(
                leading: _buildActionBadge(item.action),
                title: Text(
                  item.actionText,
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${global.language("by_user")} ${item.userName} (${item.userCode})',
                      style: const TextStyle(fontSize: 12),
                    ),
                    Text(
                      '${global.language("time_label")} ${_formatDateTime(item.timestamp)}',
                      style: const TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                  ],
                ),
                trailing: IconButton(
                  icon: Icon(isExpanded ? Icons.expand_less : Icons.expand_more),
                  onPressed: () {
                    setState(() {
                      _expandedIndex = isExpanded ? null : index;
                    });
                  },
                ),
                onTap: () {
                  setState(() {
                    _expandedIndex = isExpanded ? null : index;
                  });
                },
              ),
              if (isExpanded) _buildExpandedContent(item),
            ],
          ),
        );
      },
    );
  }

  Widget _buildActionBadge(String action) {
    Color color;
    IconData icon;

    switch (action) {
      case 'create':
        color = Colors.green;
        icon = Icons.add_circle;
        break;
      case 'update':
        color = Colors.blue;
        icon = Icons.edit;
        break;
      case 'delete':
        color = Colors.red;
        icon = Icons.delete;
        break;
      default:
        color = Colors.grey;
        icon = Icons.info;
    }

    return CircleAvatar(
      backgroundColor: color,
      radius: 20,
      child: Icon(icon, color: Colors.white, size: 20),
    );
  }

  Widget _buildExpandedContent(DataHistoryModel item) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(4),
          bottomRight: Radius.circular(4),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // แสดง Changes
          if (item.changes.isNotEmpty) ...[
            Text(
              '${global.language("changes")}:',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            ...item.changes.map((change) => _buildChangeRow(change)),
            const Divider(),
          ],
          // ปุ่มดู JSON เต็ม
          Row(
            children: [
              if (item.dataBefore != null)
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _showJsonDialog(global.language("data_before_edit"), item.dataBefore!),
                    icon: Icon(Icons.arrow_back, size: 16),
                    label: Text(global.language("data_before_edit")),
                  ),
                ),
              if (item.dataBefore != null && item.dataAfter != null)
                const SizedBox(width: 8),
              if (item.dataAfter != null)
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _showJsonDialog(global.language("data_after_edit"), item.dataAfter!),
                    icon: Icon(Icons.arrow_forward, size: 16),
                    label: Text(global.language("data_after_edit")),
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildChangeRow(FieldChange change) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              change.field,
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          const SizedBox(width: 8),
          if (change.oldValue != null)
            Expanded(
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.red[50],
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  '${change.oldValue}',
                  style: TextStyle(color: Colors.red[700], fontSize: 12),
                ),
              ),
            ),
          if (change.oldValue != null && change.newValue != null)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 8),
              child: Icon(Icons.arrow_forward, size: 16),
            ),
          if (change.newValue != null)
            Expanded(
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.green[50],
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  '${change.newValue}',
                  style: TextStyle(color: Colors.green[700], fontSize: 12),
                ),
              ),
            ),
        ],
      ),
    );
  }

  void _showJsonDialog(String title, Map<String, dynamic> data) {
    final jsonString = const JsonEncoder.withIndent('  ').convert(data);

    showDialog(
      context: context,
      builder: (context) => Dialog(
        child: Container(
          width: MediaQuery.of(context).size.width * 0.8,
          height: MediaQuery.of(context).size.height * 0.7,
          constraints: const BoxConstraints(maxWidth: 800, maxHeight: 600),
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                color: Colors.grey[800],
                child: Row(
                  children: [
                    const Icon(Icons.code, color: Colors.white),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        title,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.white),
                      onPressed: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: Container(
                  color: Colors.grey[900],
                  padding: const EdgeInsets.all(16),
                  child: SingleChildScrollView(
                    child: SelectableText(
                      jsonString,
                      style: const TextStyle(
                        fontFamily: 'monospace',
                        fontSize: 12,
                        color: Colors.lightGreenAccent,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDateTime(DateTime dateTime) {
    // แปลงเป็น local time
    final localDateTime = dateTime.toLocal();
    return DateFormat('dd/MM/yyyy HH:mm:ss').format(localDateTime);
  }

  /// Widget สำหรับแสดง card ผู้สร้าง/ผู้แก้ไข
  Widget _buildCreatorModifierCard({
    required IconData icon,
    required Color iconColor,
    required String title,
    required String code,
    required String name,
    String? dateTime,
  }) {
    // แปลง dateTime string เป็น DateTime แล้ว format
    String formattedDateTime = '-';
    if (dateTime != null && dateTime.isNotEmpty) {
      try {
        final dt = DateTime.parse(dateTime);
        formattedDateTime = _formatDateTime(dt);
      } catch (_) {
        formattedDateTime = dateTime;
      }
    }

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            CircleAvatar(
              backgroundColor: iconColor.withValues(alpha: 0.2),
              radius: 24,
              child: Icon(icon, color: iconColor, size: 24),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: iconColor,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      const Icon(Icons.person, size: 16, color: Colors.grey),
                      const SizedBox(width: 4),
                      Text(
                        name,
                        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
                      ),
                      if (code.isNotEmpty && code != '-') ...[
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: Colors.grey[200],
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            code,
                            style: const TextStyle(fontSize: 11, color: Colors.grey),
                          ),
                        ),
                      ],
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(Icons.access_time, size: 16, color: Colors.grey),
                      const SizedBox(width: 4),
                      Text(
                        formattedDateTime,
                        style: const TextStyle(fontSize: 13, color: Colors.grey),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// ฟังก์ชันสำหรับแสดง dialog
Future<void> showDataHistoryDialog(
  BuildContext context, {
  required String docNo,
  String? title,
  // ข้อมูล creator/modifier จาก TransactionModel
  String? creatorCode,
  String? creatorName,
  String? createdAt,
  String? modifierCode,
  String? modifierName,
  String? modifiedAt,
}) async {
  await showDialog(
    context: context,
    builder: (context) => DataHistoryDialog(
      docNo: docNo,
      title: title ?? global.language("edit_history"),
      creatorCode: creatorCode,
      creatorName: creatorName,
      createdAt: createdAt,
      modifierCode: modifierCode,
      modifierName: modifierName,
      modifiedAt: modifiedAt,
    ),
  );
}
