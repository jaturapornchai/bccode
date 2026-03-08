import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/slip_image/slip_image_bloc.dart';
import 'package:smlaicloud/model/slip_image_model.dart';
import 'package:smlaicloud/repositories/slip_image_repository.dart'; // ใช้สำหรับ BlocProvider
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/global.dart' as global;

/// หน้าจอแสดงรายการ SLIP เงินเข้า/ออก
class SlipListScreen extends StatelessWidget {
  final SlipType slipType;

  const SlipListScreen({
    super.key,
    required this.slipType,
  });

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => SlipImageBloc(slipImageRepository: SlipImageRepository()),
      child: _SlipListContent(slipType: slipType),
    );
  }
}

/// เนื้อหาหลักของหน้าจอ SLIP
class _SlipListContent extends StatefulWidget {
  final SlipType slipType;

  const _SlipListContent({required this.slipType});

  @override
  State<_SlipListContent> createState() => _SlipListContentState();
}

class _SlipListContentState extends State<_SlipListContent> {
  final ImagePicker _imagePicker = ImagePicker();

  DateTime _fromDate = DateTime.now().subtract(const Duration(days: 30));
  DateTime _toDate = DateTime.now();

  // รูปที่เพิ่งอัพโหลด (แสดงทันทีก่อน API ตอบกลับ)
  final List<SlipImageModel> _pendingImages = [];
  bool _isUploading = false;

  // รูปที่กำลังตรวจสอบ (filename)
  String? _verifyingFilename;

  // รูปที่กำลังลบ (ป้องกันการลบซ้ำ)
  final Set<String> _deletingFilenames = {};

  // สีหลัก
  static const Color _primaryColor = Color(0xFF667eea);
  static const Color _accentColor = Color(0xFF764ba2);

  @override
  void initState() {
    super.initState();
    // โหลดรายการหลังจาก widget ถูกสร้างเสร็จ
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadImages();
    });
  }

  /// โหลดรายการรูปภาพ
  void _loadImages() {
    context.read<SlipImageBloc>().add(LoadSlipImages(
      slipType: widget.slipType,
      fromDate: DateFormat('yyyy-MM-dd').format(_fromDate),
      toDate: DateFormat('yyyy-MM-dd').format(_toDate),
    ));
  }

  /// เลือกและอัพโหลดรูปภาพ
  Future<void> _pickAndUploadImage() async {
    debugPrint('[SLIP] _pickAndUploadImage called');
    try {
      debugPrint('[SLIP] Opening image picker...');
      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.gallery,
        maxWidth: 1920,
        maxHeight: 1920,
        imageQuality: 85,
      );

      debugPrint('[SLIP] Image picked: ${image != null ? image.name : 'null'}');

      if (image != null) {
        debugPrint('[SLIP] Reading image bytes...');
        final Uint8List imageBytes = await image.readAsBytes();
        final String fileName = image.name;
        debugPrint('[SLIP] Image size: ${imageBytes.length} bytes, filename: $fileName');

        if (mounted) {
          // สร้างรูป pending แสดงทันที
          final pendingImage = SlipImageModel(
            id: 'pending_${DateTime.now().millisecondsSinceEpoch}',
            shopId: '',
            filename: fileName,
            originalName: fileName,
            contentType: 'image/jpeg',
            size: imageBytes.length,
            category: widget.slipType.value,
            createdAt: DateTime.now(),
            imageBytes: imageBytes,
          );

          // แสดงรูปทันที + แสดง uploading state
          setState(() {
            _pendingImages.insert(0, pendingImage);
            _isUploading = true;
          });

          debugPrint('[SLIP] Dispatching UploadSlipImage event...');
          debugPrint('[SLIP] SlipType: ${widget.slipType.value}');
          context.read<SlipImageBloc>().add(UploadSlipImage(
            imageBytes: imageBytes,
            fileName: fileName,
            slipType: widget.slipType,
          ));
        }
      } else {
        debugPrint('[SLIP] No image selected');
      }
    } catch (e, stackTrace) {
      debugPrint('[SLIP] Error picking/uploading image: $e');
      debugPrint('[SLIP] StackTrace: $stackTrace');
      if (mounted) {
        global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: $e');
      }
    }
  }

  /// ยืนยันการลบรูปภาพ
  Future<void> _confirmDelete(SlipImageModel image) async {
    // ป้องกันการลบซ้ำ - ถ้ากำลังลบอยู่ให้ข้ามไป
    if (_deletingFilenames.contains(image.filename)) {
      debugPrint('[SLIP] Already deleting: ${image.filename}');
      return;
    }

    final bool? confirm = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(global.language("confirm_delete")),
        content: Text(global.language("confirm_delete_image")),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: Text(global.language("cancel")),
          ),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );

    if (confirm == true && mounted) {
      // ตรวจสอบอีกครั้งหลังจาก dialog ปิด
      if (_deletingFilenames.contains(image.filename)) {
        debugPrint('[SLIP] Already deleting (after dialog): ${image.filename}');
        return;
      }

      // เพิ่มเข้า set เพื่อป้องกันการลบซ้ำ
      setState(() {
        _deletingFilenames.add(image.filename);
      });

      debugPrint('[SLIP] Dispatching delete event for: ${image.filename}');
      context.read<SlipImageBloc>().add(DeleteSlipImage(
        filename: image.filename,
        slipType: widget.slipType,
      ));

      // ลบออกจาก set หลังจากส่ง event แล้ว (รอสักครู่เพื่อป้องกัน double-tap)
      Future.delayed(const Duration(milliseconds: 500), () {
        if (mounted) {
          setState(() {
            _deletingFilenames.remove(image.filename);
          });
        }
      });
    }
  }

  /// แสดงรูปภาพขนาดใหญ่ พร้อมรายละเอียดการตรวจสอบ
  void _showFullImage(SlipImageModel image) {
    // สร้าง image widget จาก base64 data
    final Widget imageWidget = _buildImageWidget(image, fit: BoxFit.contain);
    final bool hasVerified = image.verified == true;

    showDialog(
      context: context,
      builder: (dialogContext) => Dialog(
        backgroundColor: Colors.transparent,
        child: Stack(
          children: [
            // รูปภาพ
            Center(
              child: InteractiveViewer(
                child: imageWidget,
              ),
            ),
            // ปุ่มปิด - ElevatedButton ให้เห็นชัด
            Positioned(
              top: 16,
              right: 16,
              child: ElevatedButton.icon(
                onPressed: () => Navigator.pop(dialogContext),
                icon: Icon(Icons.close, size: 20),
                label: Text(global.language("close")),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                ),
              ),
            ),
            // รายละเอียดการตรวจสอบ (ด้านล่าง)
            if (hasVerified)
              Positioned(
                bottom: 16,
                left: 16,
                right: 16,
                child: _buildVerifyDetailCard(image),
              ),
          ],
        ),
      ),
    );
  }

  /// สร้าง Card แสดงรายละเอียดการตรวจสอบ
  Widget _buildVerifyDetailCard(SlipImageModel image) {
    final Color statusColor = _getVerifyStatusColorForImage(image);

    // กำหนดไอคอนและข้อความตามสถานะ
    IconData statusIcon;
    String statusText;

    if (image.isDuplicate == true) {
      statusIcon = Icons.warning;
      statusText = global.language("slip_duplicate");
    } else if (image.verifyStatus == 'success' || image.verified == true) {
      statusIcon = Icons.check_circle;
      statusText = global.language("slip_verified");
    } else if (image.verifyStatus == 'not_found') {
      statusIcon = Icons.search_off;
      statusText = global.language("slip_qr_not_found");
    } else {
      statusIcon = Icons.cancel;
      statusText = global.language("slip_verification_failed");
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.2),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // หัวข้อสถานะ
          Row(
            children: [
              Icon(statusIcon, color: statusColor, size: 24),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  statusText,
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: statusColor,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Divider(height: 1),
          const SizedBox(height: 12),
          // รายละเอียด
          if (image.transAmount != null)
            _buildDetailRow(global.language("amount"), '฿${image.transAmount!.toStringAsFixed(2)}', isBold: true),
          if (image.transRef != null)
            _buildDetailRow(global.language("reference_number"), image.transRef!),
          if (image.senderName != null)
            _buildDetailRow(global.language("sender"), image.senderName!),
          if (image.receiverName != null)
            _buildDetailRow(global.language("receiver"), image.receiverName!),
          if (image.createdAt != null)
            _buildDetailRow(global.language("upload_date"), DateFormat('dd/MM/yyyy HH:mm').format(image.createdAt!)),
        ],
      ),
    );
  }

  /// สร้าง row แสดงรายละเอียด
  Widget _buildDetailRow(String label, String value, {bool isBold = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              '$label:',
              style: TextStyle(
                fontSize: 13,
                color: Colors.grey.shade600,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: TextStyle(
                fontSize: 13,
                fontWeight: isBold ? FontWeight.bold : FontWeight.w500,
                color: isBold ? _primaryColor : Colors.black87,
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง Image widget จาก imageBytes, URL หรือ base64 data
  /// ลำดับความสำคัญ: imageBytes > url > base64Data
  Widget _buildImageWidget(SlipImageModel image, {BoxFit fit = BoxFit.cover}) {
    // 1. ถ้ามี imageBytes ให้ใช้เลย (รูปที่เพิ่งอัพโหลด)
    if (image.imageBytes != null && image.imageBytes!.isNotEmpty) {
      return Image.memory(
        image.imageBytes!,
        fit: fit,
        errorBuilder: (context, error, stackTrace) => _buildErrorPlaceholder(),
      );
    }

    // 2. ถ้ามี URL ให้ใช้ (รองรับทั้ง full URL และ relative path /s3/file/...)
    if (image.url != null && image.url!.isNotEmpty) {
      return Image.network(
        global.resolveFileUrl(image.url!),
        fit: fit,
        loadingBuilder: (context, child, loadingProgress) {
          if (loadingProgress == null) return child;
          return Container(
            color: Colors.grey.shade200,
            child: Center(
              child: CircularProgressIndicator(
                strokeWidth: 2,
                value: loadingProgress.expectedTotalBytes != null
                    ? loadingProgress.cumulativeBytesLoaded /
                        loadingProgress.expectedTotalBytes!
                    : null,
              ),
            ),
          );
        },
        errorBuilder: (context, error, stackTrace) {
          debugPrint('[SLIP UI] Error loading URL: $error');
          return _buildErrorPlaceholder();
        },
      );
    }

    // 3. ถ้ามี base64 data ให้ใช้ (fallback - deprecated)
    if (image.base64Data != null && image.base64Data!.isNotEmpty) {
      return _buildBase64Image(image.base64Data!, fit: fit);
    }

    // 4. ไม่มีข้อมูลรูป - แสดง placeholder
    return _buildErrorPlaceholder();
  }

  /// สร้าง Image จาก base64 string
  Widget _buildBase64Image(String base64Data, {BoxFit fit = BoxFit.cover}) {
    try {
      String base64String = base64Data;
      if (base64String.contains(',')) {
        base64String = base64String.split(',').last;
      }

      final Uint8List bytes = base64Decode(base64String);
      return Image.memory(
        bytes,
        fit: fit,
        errorBuilder: (context, error, stackTrace) => _buildErrorPlaceholder(),
      );
    } catch (e) {
      debugPrint('[SLIP UI] Error decoding base64: $e');
      return _buildErrorPlaceholder();
    }
  }

  /// สร้าง placeholder สำหรับกรณี error
  Widget _buildErrorPlaceholder() {
    return Container(
      color: Colors.grey.shade200,
      child: const Icon(
        Icons.broken_image,
        color: Colors.grey,
        size: 32,
      ),
    );
  }

  /// ตรวจสอบสลิป
  void _verifySlip(SlipImageModel image) {
    context.read<SlipImageBloc>().add(VerifySlipImage(
      imageId: image.id,
      filename: image.filename,
      slipType: widget.slipType,
      checkDuplicate: true,
      type: 'bank',
    ));
  }

  /// แสดง dialog ผลการตรวจสอบ
  void _showVerifyResultDialog(SlipVerifyResponse result, SlipImageModel image) {
    final bool isValid = result.verified == true;
    final bool isDuplicate = result.isDuplicate == true;

    // กำหนดสีและไอคอนตามสถานะ
    Color statusColor;
    IconData statusIcon;
    String titleText;

    if (isDuplicate) {
      statusColor = Colors.orange;
      statusIcon = Icons.warning;
      titleText = global.language("slip_duplicate_title");
    } else if (isValid) {
      statusColor = Colors.green;
      statusIcon = Icons.check_circle;
      titleText = global.language("slip_valid");
    } else {
      statusColor = Colors.red;
      statusIcon = Icons.cancel;
      titleText = global.language("slip_invalid");
    }

    final String statusText = _getVerifyStatusText(result.verifyStatus);

    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Row(
          children: [
            Icon(statusIcon, color: statusColor, size: 28),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                titleText,
                style: TextStyle(color: statusColor),
              ),
            ),
          ],
        ),
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              _buildResultRow(global.language("status"), statusText),
              if (isDuplicate)
                Container(
                  margin: const EdgeInsets.symmetric(vertical: 8),
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: Colors.orange.shade50,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.orange.shade200),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.warning, color: Colors.orange.shade700, size: 20),
                      SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          global.language("slip_already_used"),
                          style: TextStyle(
                            color: Colors.orange.shade700,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              if (result.transRef != null)
                _buildResultRow(global.language("reference_number"), result.transRef!),
              if (result.transAmount != null)
                _buildResultRow(global.language("amount"), '${result.transAmount!.toStringAsFixed(2)} ${global.language("baht")}'),
              if (result.senderName != null)
                _buildResultRow(global.language("sender"), result.senderName!),
              if (result.senderBank != null)
                _buildResultRow(global.language("sender_bank"), result.senderBank!),
              if (result.receiverName != null)
                _buildResultRow(global.language("receiver"), result.receiverName!),
              if (result.receiverBank != null)
                _buildResultRow(global.language("receiver_bank"), result.receiverBank!),
              if (result.transDate != null)
                _buildResultRow(global.language("transaction_date"),
                  DateFormat('dd/MM/yyyy HH:mm').format(result.transDate!)),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: Text(global.language("close")),
          ),
        ],
      ),
    );
  }

  /// สร้าง row แสดงผลลัพธ์
  Widget _buildResultRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              '$label:',
              style: TextStyle(
                color: Colors.grey.shade600,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
          ),
        ],
      ),
    );
  }

  /// แปลง verify_status เป็นข้อความ
  String _getVerifyStatusText(String? status) {
    switch (status) {
      case 'success':
        return global.language("verification_success");
      case 'valid':
        return global.language("slip_valid");
      case 'invalid':
        return global.language("slip_invalid");
      case 'duplicate':
        return global.language("slip_duplicate_short");
      case 'not_found':
        return global.language("qr_code_not_found");
      case 'error':
        return global.language("error_occurred");
      default:
        return status ?? global.language("unknown_status");
    }
  }

  /// ดึงสีสถานะการตรวจสอบ
  Color _getVerifyStatusColorForImage(SlipImageModel image) {
    if (image.isDuplicate == true) {
      return Colors.orange;
    }
    if (image.verified == true && image.verifyStatus == 'success') {
      return Colors.green;
    }
    if (image.verifyStatus == 'not_found' || image.verifyStatus == 'error') {
      return Colors.red;
    }
    return Colors.grey;
  }

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<SlipImageBloc, SlipImageState>(
      listener: (context, state) {
        // อัพโหลดสำเร็จ - ลบ pending image และ reload
        if (state is SlipImageLoaded && _isUploading) {
          setState(() {
            _pendingImages.clear();
            _isUploading = false;
          });
        }
        // อัพโหลดล้มเหลว - ลบ pending image
        if (state is SlipImageUploadFailed) {
          setState(() {
            _pendingImages.clear();
            _isUploading = false;
          });
          global.showErrorSnackBar(context, state.message);
        } else if (state is SlipImageDeleteFailed) {
          global.showErrorSnackBar(context, state.message);
        }
        // กำลังตรวจสอบสลิป
        else if (state is SlipImageVerifying) {
          setState(() {
            _verifyingFilename = state.filename;
          });
        }
        // ตรวจสอบสำเร็จ
        else if (state is SlipImageVerifySuccess) {
          setState(() {
            _verifyingFilename = null;
          });
          _showVerifyResultDialog(state.verifyResult, state.updatedImage);
        }
        // ตรวจสอบล้มเหลว
        else if (state is SlipImageVerifyFailed) {
          setState(() {
            _verifyingFilename = null;
          });
          global.showErrorSnackBar(context, state.message);
        }
      },
      builder: (context, state) {
        return Scaffold(
          appBar: AppBar(
            title: Text(widget.slipType.displayName),
            flexibleSpace: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  colors: [_primaryColor, _accentColor],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
            ),
            foregroundColor: Colors.white,
          ),
          body: Column(
            children: [
              // ส่วนเลือกช่วงวันที่
              _buildDateRangeFilter(),
              const Divider(height: 1),
              // รายการรูปภาพ
              Expanded(
                child: _buildImageList(state),
              ),
            ],
          ),
          floatingActionButton: FloatingActionButton.extended(
            onPressed: state is SlipImageUploading ? null : _pickAndUploadImage,
            backgroundColor: _primaryColor,
            foregroundColor: Colors.white,
            icon: state is SlipImageUploading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      color: Colors.white,
                      strokeWidth: 2,
                    ),
                  )
                : Icon(Icons.add_photo_alternate),
            label: Text(state is SlipImageUploading ? global.language("uploading") : global.language("add_slip")),
          ),
        );
      },
    );
  }

  /// สร้างส่วนเลือกช่วงวันที่
  Widget _buildDateRangeFilter() {
    return Container(
      padding: const EdgeInsets.all(16),
      color: Colors.grey.shade50,
      child: Row(
        children: [
          // จากวันที่
          Expanded(
            child: CustomDatePicker(
              labelText: global.language("from_date"),
              initialDate: _fromDate,
              onDateSelected: (date) {
                if (date != null) {
                  setState(() {
                    _fromDate = date;
                  });
                }
              },
              decoration: const InputDecoration(),
            ),
          ),
          const SizedBox(width: 12),
          // ถึงวันที่
          Expanded(
            child: CustomDatePicker(
              labelText: global.language("to_date"),
              initialDate: _toDate,
              onDateSelected: (date) {
                if (date != null) {
                  setState(() {
                    _toDate = date;
                  });
                }
              },
              decoration: const InputDecoration(),
            ),
          ),
          const SizedBox(width: 12),
          // ปุ่มค้นหา
          Container(
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [_primaryColor, _accentColor],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              onPressed: _loadImages,
              icon: Icon(Icons.search, color: Colors.white),
              tooltip: global.language("search"),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้างรายการรูปภาพ
  Widget _buildImageList(SlipImageState state) {
    if (state is SlipImageLoading) {
      return const Center(
        child: CircularProgressIndicator(),
      );
    }

    if (state is SlipImageLoadFailed) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 48, color: Colors.red),
            const SizedBox(height: 16),
            Text(state.message),
            SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadImages,
              child: Text(global.language("try_again")),
            ),
          ],
        ),
      );
    }

    if (state is SlipImageLoaded) {
      // รวม pending images (รูปที่กำลังอัพโหลด) กับรูปจาก API
      final allImages = [..._pendingImages, ...state.images];

      if (allImages.isEmpty) {
        return Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.image_not_supported,
                size: 64,
                color: Colors.grey.shade400,
              ),
              SizedBox(height: 16),
              Text(
                global.language("no_slip_images_found"),
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.grey.shade600,
                ),
              ),
              SizedBox(height: 8),
              Text(
                global.language("press_add_slip_to_add_image"),
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey.shade500,
                ),
              ),
            ],
          ),
        );
      }

      return GridView.builder(
        padding: const EdgeInsets.all(12),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 4,
          crossAxisSpacing: 12,
          mainAxisSpacing: 12,
          childAspectRatio: 0.85,
        ),
        itemCount: allImages.length,
        itemBuilder: (context, index) {
          final image = allImages[index];
          final isPending = image.id.startsWith('pending_');
          return _buildImageCard(image, isPending: isPending);
        },
      );
    }

    // สถานะเริ่มต้น - โหลดข้อมูล
    return const Center(
      child: CircularProgressIndicator(),
    );
  }

  /// สร้าง card แสดงรูปภาพ
  /// [isPending] - true ถ้ารูปกำลังอัพโหลด (แสดง loading overlay)
  Widget _buildImageCard(SlipImageModel image, {bool isPending = false}) {
    final bool isVerifying = _verifyingFilename == image.filename;
    final bool hasVerified = image.verified == true;
    final Color statusColor = _getVerifyStatusColorForImage(image);

    return Card(
      elevation: 3,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
      ),
      child: Stack(
        children: [
          // เนื้อหาหลัก
          InkWell(
            onTap: isPending ? null : () => _showFullImage(image),
            borderRadius: BorderRadius.circular(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // รูปภาพ
                Expanded(
                  child: ClipRRect(
                    borderRadius: const BorderRadius.vertical(
                      top: Radius.circular(12),
                    ),
                    child: _buildImageWidget(image),
                  ),
                ),
                // วันที่ + สถานะ
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                  decoration: BoxDecoration(
                    color: Colors.grey.shade100,
                    borderRadius: const BorderRadius.vertical(
                      bottom: Radius.circular(12),
                    ),
                  ),
                  child: Column(
                    children: [
                      // วันที่
                      Text(
                        isPending
                            ? global.language("uploading")
                            : (image.createdAt != null
                                ? DateFormat('dd/MM/yyyy HH:mm').format(image.createdAt!)
                                : '-'),
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 10,
                          color: isPending ? _primaryColor : Colors.grey.shade700,
                          fontWeight: isPending ? FontWeight.bold : FontWeight.normal,
                        ),
                      ),
                      const SizedBox(height: 2),
                      // สถานะการตรวจสอบ + ปุ่ม
                      if (!isPending)
                        _buildVerifyStatusRow(image, isVerifying, hasVerified, statusColor),
                    ],
                  ),
                ),
              ],
            ),
          ),
          // Loading overlay สำหรับ pending หรือ verifying
          if (isPending || isVerifying)
            Positioned.fill(
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.black.withValues(alpha: 0.3),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const CircularProgressIndicator(
                        color: Colors.white,
                        strokeWidth: 3,
                      ),
                      if (isVerifying) ...[
                        SizedBox(height: 8),
                        Text(
                          global.language("verifying"),
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          // ปุ่มลบ (มุมขวาบน) - ซ่อนถ้ากำลังอัพโหลด/ตรวจสอบ/ลบ
          if (!isPending && !isVerifying && !_deletingFilenames.contains(image.filename))
            Positioned(
              top: 4,
              right: 4,
              child: Material(
                color: Colors.red.withValues(alpha: 0.9),
                borderRadius: BorderRadius.circular(16),
                child: InkWell(
                  onTap: () => _confirmDelete(image),
                  borderRadius: BorderRadius.circular(16),
                  child: const Padding(
                    padding: EdgeInsets.all(4),
                    child: Icon(
                      Icons.close,
                      color: Colors.white,
                      size: 16,
                    ),
                  ),
                ),
              ),
            ),
          // Badge สถานะ (มุมซ้ายบน) - แสดงเฉพาะเมื่อตรวจสอบแล้ว
          if (hasVerified && !isPending && !isVerifying)
            Positioned(
              top: 4,
              left: 4,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: statusColor,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      image.isDuplicate == true
                          ? Icons.warning
                          : (image.verifyStatus == 'success' || image.verified == true)
                              ? Icons.check_circle
                              : Icons.cancel,
                      color: Colors.white,
                      size: 12,
                    ),
                    SizedBox(width: 2),
                    Text(
                      image.isDuplicate == true
                          ? global.language("duplicate_short")
                          : (image.verifyStatus == 'success' || image.verified == true)
                              ? global.language("passed")
                              : global.language("failed"),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 9,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }

  /// สร้าง row แสดงสถานะการตรวจสอบและปุ่ม
  Widget _buildVerifyStatusRow(
    SlipImageModel image,
    bool isVerifying,
    bool hasVerified,
    Color statusColor,
  ) {
    // ถ้าตรวจสอบแล้ว แสดงสถานะที่ชัดเจน (ไม่แสดงปุ่ม)
    if (hasVerified) {
      // กำหนดไอคอนและข้อความตามสถานะ
      IconData statusIcon;
      String statusText;

      if (image.isDuplicate == true) {
        statusIcon = Icons.warning;
        statusText = global.language("duplicate_short");
      } else if (image.verifyStatus == 'success' || image.verified == true) {
        statusIcon = Icons.check_circle;
        statusText = global.language("passed");
      } else if (image.verifyStatus == 'not_found') {
        statusIcon = Icons.search_off;
        statusText = global.language("qr_not_found");
      } else {
        statusIcon = Icons.cancel;
        statusText = global.language("failed");
      }

      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(
          color: statusColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(statusIcon, size: 14, color: statusColor),
            const SizedBox(width: 4),
            Text(
              statusText,
              style: TextStyle(
                fontSize: 11,
                color: statusColor,
                fontWeight: FontWeight.bold,
              ),
            ),
            if (image.transAmount != null) ...[
              const SizedBox(width: 4),
              Text(
                '฿${image.transAmount!.toStringAsFixed(0)}',
                style: TextStyle(
                  fontSize: 11,
                  color: statusColor,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ],
        ),
      );
    }

    // ยังไม่ตรวจสอบ - แสดงปุ่มตรวจสอบ (ขยายใหญ่ 100%)
    return InkWell(
      onTap: isVerifying ? null : () => _verifySlip(image),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
        decoration: BoxDecoration(
          color: _primaryColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: _primaryColor.withValues(alpha: 0.3), width: 1.5),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.qr_code_scanner,
              size: 18,
              color: _primaryColor,
            ),
            SizedBox(width: 4),
            Text(
              global.language("verify"),
              style: TextStyle(
                fontSize: 12,
                color: _primaryColor,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

