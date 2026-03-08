import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:file_picker/file_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'dart:async';
import 'dart:convert';
import 'package:smlaicloud/services/chunked_upload_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/model/import/product_import_response.dart';
import 'package:smlaicloud/screens/import/import_product_detail_screen.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ImportProductFromFileScreen extends StatefulWidget {
  const ImportProductFromFileScreen({super.key});

  @override
  State<ImportProductFromFileScreen> createState() =>
      _ImportProductFromFileScreenState();
}

class _ImportProductFromFileScreenState
    extends State<ImportProductFromFileScreen>
    with SingleTickerProviderStateMixin {
  // Upload state
  bool isUploading = false;
  bool uploadComplete = false;
  double uploadProgress = 0.0;
  String? selectedFileName;
  String? uploadId;
  String? errorMessage;
  String? lastUploadedFileName; // เก็บชื่อไฟล์ล่าสุด
  String? lastUploadedFileUrl; // เก็บ presigned URL จาก upload response

  // Import result
  ProductImportResponse? importResult;

  // File info
  int? fileSize;
  int? totalChunks;
  int uploadedChunks = 0;

  // Animation
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;

  // Services
  final ChunkedUploadService _uploadService = ChunkedUploadService();

  // SharedPreferences keys
  static const String _lastFileNameKey = 'last_uploaded_file_name';
  static const String _lastFileUrlKey = 'last_uploaded_file_url';

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 500),
      vsync: this,
    );
    _scaleAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.elasticOut),
    );
    _loadLastFileName();
  }

  // โหลดชื่อไฟล์ล่าสุด
  Future<void> _loadLastFileName() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final fileName = prefs.getString(_lastFileNameKey);
      final fileUrl = prefs.getString(_lastFileUrlKey);
      if (fileName != null && mounted) {
        setState(() {
          lastUploadedFileName = fileName;
          lastUploadedFileUrl = fileUrl;
        });
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error loading last file name: $e');
      }
    }
  }

  // บันทึกชื่อไฟล์ล่าสุด
  Future<void> _saveLastFileName(String fileName, {String? fileUrl}) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_lastFileNameKey, fileName);
      if (fileUrl != null && fileUrl.isNotEmpty) {
        await prefs.setString(_lastFileUrlKey, fileUrl);
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error saving last file name: $e');
      }
    }
  }

  @override
  void dispose() {
    _animationController.dispose();
    _uploadService.dispose();
    super.dispose();
  }

  Future<void> _pickAndUploadFile() async {
    try {
      // Pick file - Allow XLSX only
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx'],
        withData: kIsWeb,
        withReadStream: !kIsWeb,
      );

      if (result == null) return;

      PlatformFile file = result.files.first;

      setState(() {
        selectedFileName = file.name;
        fileSize = file.size;
        errorMessage = null;
      });

      // Validate file extension
      if (!file.name.toLowerCase().endsWith('.xlsx')) {
        setState(() {
          errorMessage = global.language('import_product_file_excel_only');
        });
        return;
      }

      // Validate file size
      if (file.size <= 0) {
        setState(() {
          errorMessage =
              global.language('cannot_read_file_size');
        });
        return;
      }

      if (file.size > 500 * 1024 * 1024) {
        setState(() {
          errorMessage =
              'ไฟล์มีขนาดใหญ่เกินไป (${global.language('import_product_file_max_size')})';
        });
        return;
      }

      // Start upload
      await _uploadFile(file);
    } catch (e) {
      setState(() {
        errorMessage = '${global.language("error_occurred")}: $e';
        isUploading = false;
      });

      if (kDebugMode) {
        AppLogger.error('Error picking file: $e');
      }
    }
  }

  Future<void> _uploadFile(PlatformFile file) async {
    setState(() {
      isUploading = true;
      uploadProgress = 0.0;
      uploadedChunks = 0;
      errorMessage = null;
    });

    try {
      ({String fileName, String fileUrl})? uploadResult;

      if (kDebugMode) {
        AppLogger.debug('Starting upload: ${file.name}, size: ${file.size}');
      }

      if (kIsWeb) {
        // Web: use bytes
        if (file.bytes == null) {
          throw Exception(global.language('cannot_read_file_data'));
        }

        if (kDebugMode) {
          AppLogger.debug('Uploading from bytes: ${file.bytes!.length} bytes');
        }

        uploadResult = await _uploadService.uploadFromBytes(
          bytes: file.bytes!,
          fileName: file.name,
          shopId: global.getShopId(),
          onProgress: (progress, chunks, total) {
            setState(() {
              uploadProgress = progress;
              uploadedChunks = chunks;
              totalChunks = total;
            });
          },
        );
      } else {
        // Mobile/Desktop: use stream
        if (file.readStream == null) {
          throw Exception(global.language('cannot_read_file'));
        }

        uploadResult = await _uploadService.uploadFromStream(
          stream: file.readStream!,
          fileSize: file.size,
          fileName: file.name,
          shopId: global.getShopId(),
          onProgress: (progress, chunks, total) {
            setState(() {
              uploadProgress = progress;
              uploadedChunks = chunks;
              totalChunks = total;
            });
          },
        );
      }

      // Upload successful
      if (mounted) {
        setState(() {
          isUploading = false;
          uploadProgress = 1.0;
          uploadComplete = true;
        });

        // บันทึกชื่อไฟล์และ fileUrl ล่าสุด
        await _saveLastFileName(uploadResult.fileName, fileUrl: uploadResult.fileUrl);
        setState(() {
          lastUploadedFileName = uploadResult!.fileName;
          lastUploadedFileUrl = uploadResult.fileUrl;
        });

        // Animate checkmark
        _animationController.forward();

        // แสดง snackbar หลังจาก animation เสร็จ
        await Future.delayed(const Duration(milliseconds: 800));

        if (mounted) {
          // แสดง snackbar แทน dialog (ไม่รีเซ็ต animation)
          global.showSnackBar(
            context,
            Icon(Icons.check_circle, color: Colors.white),
            '${global.language("upload_file_success")}: ${uploadResult.fileName}',
            Colors.green,
          );
        }
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          isUploading = false;
          errorMessage = '${global.language("upload_failed")}: $e';
        });
      }

      if (kDebugMode) {
        AppLogger.error('Upload error: $e');
      }
    }
  }

  Future<void> _cancelUpload() async {
    if (uploadId != null) {
      try {
        await _uploadService.cancelUpload(uploadId!);
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error cancelling upload: $e');
        }
      }
    }

    setState(() {
      isUploading = false;
      uploadComplete = false;
      uploadProgress = 0.0;
      uploadedChunks = 0;
      totalChunks = null;
      uploadId = null;
    });

    _animationController.reset();
  }

  // นำเข้าสินค้า/Barcode
  Future<void> _importProducts() async {
    if (lastUploadedFileName == null) return;

    // Confirm dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('import_product_file_confirm_import')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('${global.language("file")}: $lastUploadedFileName'),
            SizedBox(height: 10),
            Text('${global.language('import_product_file_confirm_import')}?'),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('import_product_file_cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.green[700],
              foregroundColor: Colors.white,
            ),
            child: Text(global.language('import_product_file_confirm')),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      // Show loading
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => AlertDialog(
          content: Row(
            children: [
              const CircularProgressIndicator(),
              SizedBox(width: 20),
              Text(global.language('import_product_file_importing')),
            ],
          ),
        ),
      );

      // Call Go API to start product import
      final url = global.goApiUrlPath('xlsx/product/start');

      if (kDebugMode) {
        AppLogger.debug('Sending request to: $url');
        AppLogger.debug('shopId: ${global.getShopId()}');
        AppLogger.info('fileName: $lastUploadedFileName');
      }

      // Try sending as JSON first
      var httpClient = http.Client();
      try {
        final uri = Uri.parse(url);

        // Send as JSON body instead of multipart
        final response = await httpClient.post(
          uri,
          headers: {
            'Content-Type': 'application/json',
            if (global.apiToken.isNotEmpty)
              'Authorization': 'Bearer ${global.apiToken}',
          },
          body: jsonEncode({
            'shopId': global.getShopId(),
            'fileName': lastUploadedFileName, // ส่งชื่อไฟล์ที่อัปโหลดแล้ว
            if (lastUploadedFileUrl != null && lastUploadedFileUrl!.isNotEmpty)
              'fileUrl': lastUploadedFileUrl, // presigned URL จาก upload response (SeaweedFS)
          }),
        );

        if (mounted) {
          Navigator.pop(context); // Close loading

          if (response.statusCode == 200) {
            final responseData = json.decode(response.body);
            var result = ProductImportResponse.fromJson(responseData);

            // Parse duplicateBarcodes from errorMessage if status = 3
            if (result.status == 3 &&
                result.errorMessage != null &&
                result.duplicateBarcodes == null) {
              final duplicates = _parseDuplicateBarcodes(result.errorMessage!);
              if (duplicates.isNotEmpty) {
                // Create new instance with parsed duplicates
                result = ProductImportResponse(
                  comparison: result.comparison,
                  duration: result.duration,
                  errorCount: result.errorCount,
                  fileName: result.fileName,
                  shopId: result.shopId,
                  status: result.status,
                  success: result.success,
                  successCount: result.successCount,
                  totalRows: result.totalRows,
                  errorMessage: result.errorMessage,
                  duplicateBarcodes: duplicates,
                );
              }
            }

            setState(() {
              importResult = result;
            });

            // แสดง SnackBar ตามสถานะ
            if (result.status == 3) {
              // Status 3 = มี duplicate barcodes
              global.showSnackBar(
                context,
                Icon(Icons.warning, color: Colors.white),
                '${global.language("found_duplicate_barcode")} ${result.duplicateBarcodes?.length ?? 0} ${global.language("items")}',
                Colors.orange,
              );
            } else if (result.status == 2) {
              // Status 2 = สำเร็จปกติ
              global.showSnackBar(
                context,
                Icon(Icons.check_circle, color: Colors.white),
                '${global.language("import_data_success")}: ${result.successCount} ${global.language("items")}',
                Colors.green,
              );
            } else {
              // Status อื่นๆ
              global.showSnackBar(
                context,
                Icon(Icons.warning, color: Colors.white),
                result.errorMessage ?? '${global.language("unknown_status")}: ${result.status}',
                Colors.orange,
              );
            }
          } else {
            // API error
            String errorMsg = 'Unknown error';
            try {
              final errorData = json.decode(response.body);
              errorMsg = errorData['error'] ?? errorData['message'] ?? errorMsg;
            } catch (_) {
              errorMsg = response.body.isNotEmpty
                  ? response.body
                  : 'HTTP ${response.statusCode}';
            }

            global.showSnackBar(
              context,
              Icon(Icons.error, color: Colors.white),
              '${global.language("error_occurred")}: $errorMsg',
              Colors.red,
            );
          }
        }
      } finally {
        httpClient.close();
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // Close loading
        global.showSnackBar(
          context,
          Icon(Icons.error, color: Colors.white),
          '${global.language("error_occurred")}: $e',
          Colors.red,
        );
      }

      if (kDebugMode) {
        AppLogger.error('Import error: $e');
      }
    }
  }

  // Parse duplicate barcodes from error message
  List<DuplicateBarcode> _parseDuplicateBarcodes(String errorMessage) {
    final List<DuplicateBarcode> duplicates = [];

    // Format: "พบ barcode ซ้ำกัน 9 รายการ: 05712526 (ซ้ำ 2 ครั้ง ที่แถว: 604, 604); ..."
    try {
      // Extract the part after ": "
      final parts = errorMessage.split(': ');
      if (parts.length < 2) return duplicates;

      final barcodesPart = parts
          .sublist(1)
          .join(': '); // Join in case there are more ":"

      // Split by ";" to get each barcode entry
      final entries = barcodesPart.split(';');

      for (final entry in entries) {
        final trimmed = entry.trim();
        if (trimmed.isEmpty) continue;

        // Parse: "05712526 (ซ้ำ 2 ครั้ง ที่แถว: 604, 604)"
        final match = RegExp(
          r'^(\S+)\s+\(ซ้ำ\s+(\d+)\s+ครั้ง\s+ที่แถว:\s+([0-9,\s]+)\)',
        ).firstMatch(trimmed);

        if (match != null) {
          final barcode = match.group(1)!;
          final count = int.tryParse(match.group(2)!) ?? 0;
          final rowsStr = match.group(3)!;
          final rowNumbers = rowsStr
              .split(',')
              .map((r) => int.tryParse(r.trim()) ?? 0)
              .where((r) => r > 0)
              .toList();

          duplicates.add(
            DuplicateBarcode(
              barcode: barcode,
              count: count,
              rowNumbers: rowNumbers,
            ),
          );
        }
      }
    } catch (e) {
      AppLogger.error('Error parsing duplicate barcodes: $e');
    }

    return duplicates;
  }

  void _showDuplicateBarcodeDialog(ProductImportResponse result) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.warning, color: Colors.orange[700]),
            SizedBox(width: 10),
            Text(
              global.language('import_product_file_duplicate_barcode_found'),
            ),
          ],
        ),
        content: SizedBox(
          width: double.maxFinite,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (result.errorMessage != null) ...[
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.orange[50],
                    border: Border.all(color: Colors.orange[200]!),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.info_outline, color: Colors.orange[700]),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          result.errorMessage!,
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.orange[900],
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 20),
              ],
              if (result.duplicateBarcodes != null &&
                  result.duplicateBarcodes!.isNotEmpty) ...[
                Text(
                  global.language(
                    'import_product_file_duplicate_barcode_detail',
                  ),
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 10),

                // DataTable for duplicate barcodes
                Expanded(
                  child: SingleChildScrollView(
                    child: DataTable(
                      headingRowColor: WidgetStateProperty.all(
                        Colors.orange[50],
                      ),
                      border: TableBorder.all(
                        color: Colors.grey[300]!,
                        width: 1,
                      ),
                      columnSpacing: 20,
                      columns: [
                        const DataColumn(
                          label: Text(
                            '#',
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                        ),
                        const DataColumn(
                          label: Text(
                            'Barcode',
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                        ),
                        DataColumn(
                          label: Text(
                            global.language(
                              'import_product_file_duplicate_count',
                            ),
                            style: const TextStyle(fontWeight: FontWeight.bold),
                          ),
                        ),
                        DataColumn(
                          label: Text(
                            global.language('import_product_file_row_found'),
                            style: const TextStyle(fontWeight: FontWeight.bold),
                          ),
                        ),
                      ],
                      rows: result.duplicateBarcodes!.asMap().entries.map((
                        entry,
                      ) {
                        final index = entry.key;
                        final dup = entry.value;
                        return DataRow(
                          color: WidgetStateProperty.all(
                            index % 2 == 0 ? Colors.white : Colors.orange[25],
                          ),
                          cells: [
                            DataCell(
                              Text(
                                '${index + 1}',
                                style: TextStyle(color: Colors.grey[600]),
                              ),
                            ),
                            DataCell(
                              Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(
                                    Icons.qr_code,
                                    color: Colors.orange[700],
                                    size: 16,
                                  ),
                                  const SizedBox(width: 6),
                                  Text(
                                    dup.barcode,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            DataCell(
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.orange[100],
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Text(
                                  '${dup.count}',
                                  style: TextStyle(
                                    color: Colors.orange[900],
                                    fontWeight: FontWeight.bold,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                            ),
                            DataCell(
                              Text(
                                dup.rowNumbers.join(', '),
                                style: TextStyle(
                                  color: Colors.grey[700],
                                  fontSize: 13,
                                ),
                              ),
                            ),
                          ],
                        );
                      }).toList(),
                    ),
                  ),
                ),
                const SizedBox(height: 10),
              ],
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.blue[50],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    Icon(Icons.info_outline, color: Colors.blue[700]),
                    SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        global.language(
                          'import_product_file_fix_duplicates_msg',
                        ),
                        style: TextStyle(fontSize: 13, color: Colors.blue[900]),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('import_product_file_close')),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryRow(String label, String value, {Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontSize: 14)),
          Text(
            value,
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(2)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(2)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('import_product_file_title')),
        backgroundColor: const Color(0xFF0A3880),
        actions: [
          IconButton(
            icon: Icon(Icons.exit_to_app),
            tooltip: global.language('import_product_file_back_to_main'),
            onPressed: () {
              Navigator.pop(context);
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Select File Button (ย้ายมาด้านบน)
            if (!isUploading) ...[
              ElevatedButton.icon(
                onPressed: _pickAndUploadFile,
                icon: Icon(Icons.upload_file),
                label: Text(global.language('import_product_file_select_file')),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.all(20),
                  backgroundColor: const Color(0xFF0A3880),
                  foregroundColor: Colors.white,
                ),
              ),
              const SizedBox(height: 20),
            ],

            // Info Card
            Card(
              color: Colors.blue[50],
              child: Padding(
                padding: EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.info_outline, color: Colors.blue[700]),
                        SizedBox(width: 10),
                        Text(
                          global.language('import_product_file_excel_only'),
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Colors.blue[900],
                          ),
                        ),
                      ],
                    ),
                    SizedBox(height: 10),
                    Text(
                      global.language('import_product_file_max_size'),
                      style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                    ),
                  ],
                ),
              ),
            ),

            const SizedBox(height: 20),

            // ไฟล์ล่าสุด
            if (lastUploadedFileName != null && !isUploading) ...[
              Card(
                color: Colors.green[50],
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Icon(Icons.history, color: Colors.green[700]),
                          SizedBox(width: 10),
                          Text(
                            global.language('import_product_file_latest_file'),
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Colors.green[900],
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      Row(
                        children: [
                          Icon(
                            Icons.insert_drive_file,
                            size: 16,
                            color: Colors.green[700],
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              lastUploadedFileName!,
                              style: TextStyle(
                                fontSize: 14,
                                color: Colors.green[900],
                                fontWeight: FontWeight.w500,
                              ),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),
                      SizedBox(height: 8),
                      Text(
                        global.language('import_product_file_upload_complete'),
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.green[700],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 20),
            ],

            // Error Message
            if (errorMessage != null && !isUploading) ...[
              Card(
                color: Colors.red[50],
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      Icon(Icons.error_outline, color: Colors.red[700]),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          errorMessage!,
                          style: TextStyle(color: Colors.red[900]),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 20),
            ],

            // Upload Progress Section (แสดงทั้งตอน uploading และ complete)
            if (isUploading || uploadComplete) ...[
              // Upload Progress
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Icon(
                            uploadComplete
                                ? Icons.check_circle
                                : Icons.cloud_upload,
                            color: uploadComplete ? Colors.green : Colors.blue,
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Text(
                              selectedFileName ?? 'Uploading...',
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),

                      if (fileSize != null) ...[
                        SizedBox(height: 5),
                        Text(
                          '${global.language('import_product_file_size')}: ${_formatFileSize(fileSize!)}',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],

                      const SizedBox(height: 20),

                      // Animated Circular Progress or Checkmark
                      Center(
                        child: SizedBox(
                          width: 120,
                          height: 120,
                          child: Stack(
                            alignment: Alignment.center,
                            children: [
                              // Circular Progress Indicator
                              SizedBox(
                                width: 120,
                                height: 120,
                                child: CircularProgressIndicator(
                                  value: uploadProgress,
                                  strokeWidth: 8,
                                  backgroundColor: Colors.grey[200],
                                  valueColor: AlwaysStoppedAnimation<Color>(
                                    uploadComplete ? Colors.green : Colors.blue,
                                  ),
                                ),
                              ),
                              // Percentage or Checkmark
                              if (!uploadComplete)
                                Text(
                                  '${(uploadProgress * 100).toStringAsFixed(0)}%',
                                  style: TextStyle(
                                    fontSize: 28,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.blue[700],
                                  ),
                                )
                              else
                                AnimatedBuilder(
                                  animation: _animationController,
                                  builder: (context, child) {
                                    return Transform.scale(
                                      scale: _scaleAnimation.value,
                                      child: Icon(
                                        Icons.check_circle,
                                        size: 80,
                                        color: Colors.green[600],
                                      ),
                                    );
                                  },
                                ),
                            ],
                          ),
                        ),
                      ),

                      const SizedBox(height: 20),

                      // Linear Progress Bar (secondary indicator)
                      LinearProgressIndicator(
                        value: uploadProgress,
                        minHeight: 6,
                        backgroundColor: Colors.grey[200],
                        valueColor: AlwaysStoppedAnimation<Color>(
                          uploadComplete ? Colors.green : Colors.blue[700]!,
                        ),
                      ),

                      const SizedBox(height: 10),

                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            uploadComplete
                                ? global.language(
                                    'import_product_file_upload_success',
                                  )
                                : global.language(
                                    'import_product_file_uploading',
                                  ),
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              color: uploadComplete
                                  ? Colors.green[700]
                                  : Colors.blue[700],
                            ),
                          ),
                          if (totalChunks != null && !uploadComplete)
                            Text(
                              'Chunk $uploadedChunks/$totalChunks',
                              style: TextStyle(
                                fontSize: 12,
                                color: Colors.grey[600],
                              ),
                            ),
                        ],
                      ),

                      if (!uploadComplete) ...[
                        const SizedBox(height: 20),
                        // Cancel Button
                        Center(
                          child: TextButton.icon(
                            onPressed: _cancelUpload,
                            icon: Icon(Icons.cancel),
                            label: Text(
                              global.language('import_product_file_cancel'),
                            ),
                            style: TextButton.styleFrom(
                              foregroundColor: Colors.red,
                            ),
                          ),
                        ),
                      ] else ...[
                        const SizedBox(height: 20),
                        // Close Button (when complete)
                        Center(
                          child: ElevatedButton.icon(
                            onPressed: () {
                              setState(() {
                                uploadComplete = false;
                                uploadProgress = 0.0;
                                selectedFileName = null;
                              });
                              _animationController.reset();
                            },
                            icon: Icon(Icons.close),
                            label: Text(
                              global.language('import_product_file_close'),
                            ),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.grey[700],
                              foregroundColor: Colors.white,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ],
            // Action Buttons (enabled when file is uploaded)
            if (lastUploadedFileName != null && !isUploading) ...[
              // ปุ่มตรวจสอบข้อมูล
              Row(
                children: [
                  // ปุ่มนำเข้าสินค้า/Barcode
                  ElevatedButton.icon(
                    onPressed: _importProducts,
                    icon: Icon(Icons.inventory),
                    label: Text(
                      global.language(
                        'import_product_file_import_product_barcode',
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.all(20),
                      backgroundColor: Colors.green[700],
                      foregroundColor: Colors.white,
                    ),
                  ),
                ],
              ),
            ],

            // Display Import Results Table
            if (importResult != null) ...[
              const SizedBox(height: 30),

              // Summary Section
              Card(
                elevation: 2,
                child: Padding(
                  padding: EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Icon(
                            importResult!.status == 3
                                ? Icons.warning
                                : Icons.check_circle,
                            color: importResult!.status == 3
                                ? Colors.orange[700]
                                : Colors.green[700],
                          ),
                          SizedBox(width: 10),
                          Text(
                            importResult!.status == 3
                                ? global.language(
                                    'import_product_file_error_found',
                                  )
                                : global.language(
                                    'import_product_file_import_result',
                                  ),
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 15),

                      // Summary rows
                      _buildSummaryRow(
                        global.language('import_product_file_file'),
                        importResult!.fileName,
                      ),
                      _buildSummaryRow(
                        global.language('import_product_file_time_used'),
                        importResult!.duration,
                      ),

                      // Show error message if status = 3
                      if (importResult!.status == 3) ...[
                        const Divider(height: 20),
                        // Show error message if available
                        if (importResult!.errorMessage != null) ...[
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: Colors.orange[50],
                              border: Border.all(color: Colors.orange[200]!),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Icon(
                                  Icons.error_outline,
                                  color: Colors.orange[700],
                                ),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: Text(
                                    importResult!.errorMessage!,
                                    style: TextStyle(
                                      fontSize: 13,
                                      color: Colors.orange[900],
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 10),
                        ],
                        // Show duplicate barcodes section
                        if (importResult!.duplicateBarcodes != null &&
                            importResult!.duplicateBarcodes!.isNotEmpty) ...[
                          Container(
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: Colors.orange[50],
                              border: Border.all(color: Colors.orange[200]!),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                Row(
                                  children: [
                                    Icon(
                                      Icons.warning_amber,
                                      color: Colors.orange[700],
                                    ),
                                    SizedBox(width: 10),
                                    Text(
                                      global.language(
                                        'import_product_file_duplicate_barcode_found',
                                      ),
                                      style: TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                                SizedBox(height: 12),
                                Text(
                                  '${global.language("found_duplicate_barcode")} ${importResult!.duplicateBarcodes!.length} ${global.language("items")}',
                                  style: TextStyle(
                                    fontSize: 14,
                                    color: Colors.grey[700],
                                  ),
                                ),
                                SizedBox(height: 16),
                                ElevatedButton.icon(
                                  onPressed: () {
                                    _showDuplicateBarcodeDialog(importResult!);
                                  },
                                  icon: Icon(Icons.list),
                                  label: Text(
                                    global.language(
                                      'import_product_file_view_duplicate_list',
                                    ),
                                  ),
                                  style: ElevatedButton.styleFrom(
                                    padding: const EdgeInsets.all(16),
                                    backgroundColor: Colors.orange[700],
                                    foregroundColor: Colors.white,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ],

                      const Divider(height: 20),

                      if (importResult!.comparison != null) ...[
                        _buildSummaryRow(
                          global.language('import_product_file_excel_rows'),
                          '${importResult!.comparison!.totalExcelRows} ${global.language("rows")}',
                        ),
                        _buildSummaryRow(
                          global.language(
                            'import_product_file_products_in_system',
                          ),
                          '${importResult!.comparison!.totalMongoProducts} ${global.language("items")}',
                        ),
                        Divider(height: 20),
                        _buildSummaryRow(
                          global.language('import_product_file_new_products'),
                          '${importResult!.comparison!.newProductCount} ${global.language("items")}',
                          color: Colors.green,
                        ),
                        _buildSummaryRow(
                          global.language('import_product_file_updated'),
                          '${importResult!.comparison!.updatedProductCount} ${global.language("items")}',
                          color: Colors.orange,
                        ),
                        _buildSummaryRow(
                          global.language('import_product_file_unchanged'),
                          '${importResult!.comparison!.unchangedCount} ${global.language("items")}',
                          color: Colors.grey,
                        ),
                        const Divider(height: 20),
                      ],

                      _buildSummaryRow(
                        global.language('import_product_file_success'),
                        '${importResult!.successCount}/${importResult!.totalRows} ${global.language("items")}',
                        color: Colors.green,
                      ),
                      _buildSummaryRow(
                        global.language('import_product_file_error'),
                        '${importResult!.errorCount} ${global.language("items")}',
                        color: importResult!.errorCount > 0
                            ? Colors.red
                            : Colors.grey,
                      ),
                    ],
                  ),
                ),
              ),

              // Buttons Section
              const SizedBox(height: 20),

              // Button to view details in new screen
              if (importResult!.comparison != null &&
                  importResult!.comparison!.products.isNotEmpty) ...[
                Card(
                  elevation: 2,
                  child: Padding(
                    padding: EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          children: [
                            Icon(Icons.list_alt, color: Colors.blue[700]),
                            SizedBox(width: 10),
                            Text(
                              global.language(
                                'import_product_file_import_detail',
                              ),
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                        SizedBox(height: 16),
                        Text(
                          '${global.language("total_items")}: ${importResult!.comparison!.products.length} ${global.language("items")}',
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.grey[700],
                          ),
                        ),
                        SizedBox(height: 20),
                        ElevatedButton.icon(
                          onPressed: () {
                            Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (context) => ImportProductDetailScreen(
                                  importResult: importResult!,
                                ),
                              ),
                            );
                          },
                          icon: Icon(Icons.visibility),
                          label: Text(
                            global.language(
                              'import_product_file_view_all_details',
                            ),
                          ),
                          style: ElevatedButton.styleFrom(
                            padding: const EdgeInsets.all(16),
                            backgroundColor: Colors.blue[700],
                            foregroundColor: Colors.white,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ],
          ],
        ),
      ),
    );
  }
}
