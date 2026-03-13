import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/model/attachment_model.dart';
import 'package:smlaicloud/services/attachment_api_service.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:intl/intl.dart';

/// Dialog สำหรับจัดการไฟล์แนบเอกสาร
/// - แสดงรายการไฟล์แนบ
/// - อัปโหลดไฟล์ใหม่
/// - ดู/ดาวน์โหลดไฟล์
/// - ลบไฟล์
class AttachmentDialog extends StatefulWidget {
  final String docNo;
  final String guidFixed;
  final String screenType;

  const AttachmentDialog({
    super.key,
    required this.docNo,
    required this.guidFixed,
    required this.screenType,
  });

  @override
  State<AttachmentDialog> createState() => _AttachmentDialogState();
}

class _AttachmentDialogState extends State<AttachmentDialog> {
  final AttachmentApiService _apiService = AttachmentApiService();
  List<AttachmentModel> _attachments = [];
  bool _isLoading = false;
  bool _isUploading = false;

  @override
  void initState() {
    super.initState();
    _loadAttachments();
  }

  Future<void> _loadAttachments() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final response = await _apiService.listAttachments(
        shopId: global.getShopId(),
        screenType: widget.screenType,
        docNo: widget.docNo,
        guidFixed: widget.guidFixed,
      );

      setState(() {
        _attachments = response.data;
        _isLoading = false;
      });
    } catch (e) {
      AppLogger.error('Failed to load attachments: $e');
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("cannot_load_attachments")}: $e');
      }
    }
  }

  Future<void> _pickAndUploadFile() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['pdf', 'xlsx', 'xls', 'jpg', 'jpeg', 'png'],
        allowMultiple: false,
      );

      if (result == null || result.files.isEmpty) return;

      final platformFile = result.files.first;
      if (platformFile.path == null) {
        global.showErrorSnackBar(context, global.language('cannot_read_file'));
        return;
      }

      final file = File(platformFile.path!);
      final fileSize = await file.length();

      // Check file size (max 20MB)
      const maxSize = 20 * 1024 * 1024;
      if (fileSize > maxSize) {
        if (mounted) {
          global.showWarningSnackBar(context, '${global.language("file_too_large")} 20 MB');
        }
        return;
      }

      // Show description dialog
      final description = await _showDescriptionDialog();
      if (description == null) return; // User cancelled

      setState(() {
        _isUploading = true;
      });

      final response = await _apiService.uploadAttachment(
        file: file,
        shopId: global.getShopId(),
        screenType: widget.screenType,
        docNo: widget.docNo,
        guidFixed: widget.guidFixed,
        uploadedBy: global.profileData.username ?? 'unknown',
        uploadedName: global.profileData.name,
        description: description,
      );

      setState(() {
        _isUploading = false;
      });

      if (response.status == 'success') {
        if (mounted) {
          global.showSuccessSnackBar(context, global.language('upload_file_success'));
        }
        _loadAttachments(); // Reload list
      } else {
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("upload_failed")}: ${response.message}');
        }
      }
    } catch (e) {
      AppLogger.error('Failed to upload attachment: $e');
      setState(() {
        _isUploading = false;
      });
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("upload_failed")}: $e');
      }
    }
  }

  Future<String?> _showDescriptionDialog() async {
    final controller = TextEditingController();
    return showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("file_description")),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            hintText: 'เช่น ใบเสนอราคาจากซัพพลายเออร์',
            border: OutlineInputBorder(),
          ),
          maxLines: 3,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, controller.text),
            child: Text(global.language('ok')),
          ),
        ],
      ),
    );
  }

  Future<void> _deleteAttachment(AttachmentModel attachment) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('alert_confirm_delete')),
        content: Text('${global.language("confirm_delete_file")} "${attachment.originalName}"'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
            onPressed: () => Navigator.pop(context, true),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      final success = await _apiService.deleteAttachment(
        shopId: global.getShopId(),
        attachmentId: attachment.id!,
      );

      if (success) {
        if (mounted) {
          global.showSuccessSnackBar(context, global.language("delete_file_success"));
        }
        _loadAttachments(); // Reload list
      }
    } catch (e) {
      AppLogger.error('Failed to delete attachment: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("delete_file_failed")}: $e');
      }
    }
  }

  Future<void> _openAttachment(AttachmentModel attachment) async {
    if (attachment.url == null || attachment.url!.isEmpty) {
      global.showWarningSnackBar(context, global.language("download_url_not_found"));
      return;
    }

    try {
      final uri = Uri.parse(global.resolveFileUrl(attachment.url!));
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        if (mounted) {
          global.showErrorSnackBar(context, global.language("cannot_open_file"));
        }
      }
    } catch (e) {
      AppLogger.error('Failed to open attachment: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("cannot_open_file")}: $e');
      }
    }
  }

  IconData _getFileIcon(String fileType) {
    switch (fileType.toLowerCase()) {
      case 'pdf':
        return Icons.picture_as_pdf;
      case 'xlsx':
      case 'xls':
        return Icons.table_chart;
      case 'jpg':
      case 'jpeg':
      case 'png':
        return Icons.image;
      default:
        return Icons.attach_file;
    }
  }

  Color _getFileIconColor(String fileType) {
    switch (fileType.toLowerCase()) {
      case 'pdf':
        return global.theme.negativeHighlightTextColor;
      case 'xlsx':
      case 'xls':
        return global.theme.positiveHighlightTextColor;
      case 'jpg':
      case 'jpeg':
      case 'png':
        return global.theme.infoHighlightTextColor;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

  String _formatDate(String isoDate) {
    try {
      final date = DateTime.parse(isoDate);
      return DateFormat('dd/MM/yyyy HH:mm').format(date.toLocal());
    } catch (e) {
      return isoDate;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        width: 600,
        height: 500,
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  global.language("document_attachments"),
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                IconButton(
                  icon: Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
            Text(
              'เอกสารเลขที่: ${widget.docNo}',
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 16),

            // Upload button
            ElevatedButton.icon(
              onPressed: _isUploading ? null : _pickAndUploadFile,
              icon: _isUploading
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Icon(Icons.upload_file),
              label: Text(_isUploading ? global.language('import_product_file_uploading') : global.language("add_attachment")),
            ),
            const SizedBox(height: 16),

            // Attachment list
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _attachments.isEmpty
                      ? Center(
                          child: Text(
                            global.language("no_attachments"),
                            style: TextStyle(color: global.theme.textSecondaryColor),
                          ),
                        )
                      : ListView.builder(
                          itemCount: _attachments.length,
                          itemBuilder: (context, index) {
                            final attachment = _attachments[index];
                            return Card(
                              child: ListTile(
                                leading: Icon(
                                  _getFileIcon(attachment.fileType),
                                  color: _getFileIconColor(attachment.fileType),
                                  size: 40,
                                ),
                                title: Text(attachment.originalName),
                                subtitle: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    if (attachment.description != null && attachment.description!.isNotEmpty) Text(attachment.description!),
                                    Text(
                                      '${_formatFileSize(attachment.size)} • ${_formatDate(attachment.createdAt)}',
                                      style: TextStyle(fontSize: 12),
                                    ),
                                    Text(
                                      '${global.language("uploaded_by")} ${attachment.uploadedName ?? attachment.uploadedBy}',
                                      style: TextStyle(fontSize: 12),
                                    ),
                                  ],
                                ),
                                trailing: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      icon: Icon(Icons.open_in_new),
                                      onPressed: () => _openAttachment(attachment),
                                      tooltip: global.language("view_file"),
                                    ),
                                    IconButton(
                                      icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                                      onPressed: () => _deleteAttachment(attachment),
                                      tooltip: global.language("delete_file"),
                                    ),
                                  ],
                                ),
                              ),
                            );
                          },
                        ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Show attachment dialog
Future<void> showAttachmentDialog(
  BuildContext context, {
  required String docNo,
  required String guidFixed,
  required String screenType,
}) async {
  await showDialog(
    context: context,
    builder: (context) => AttachmentDialog(
      docNo: docNo,
      guidFixed: guidFixed,
      screenType: screenType,
    ),
  );
}
