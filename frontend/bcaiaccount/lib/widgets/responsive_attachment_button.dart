import 'package:flutter/material.dart';
import 'package:smlaicloud/services/attachment_api_service.dart';
import 'package:smlaicloud/widgets/attachment_dialog.dart';
import 'package:smlaicloud/global.dart' as global;

/// Responsive Attachment Button พร้อม Badge แสดงจำนวนไฟล์
/// รองรับหน้าจอทุกขนาด พร้อมคำอธิบาย
class ResponsiveAttachmentButton extends StatefulWidget {
  final String docNo;
  final String guidFixed;
  final String screenType;

  const ResponsiveAttachmentButton({
    super.key,
    required this.docNo,
    required this.guidFixed,
    required this.screenType,
  });

  @override
  State<ResponsiveAttachmentButton> createState() =>
      _ResponsiveAttachmentButtonState();
}

class _ResponsiveAttachmentButtonState
    extends State<ResponsiveAttachmentButton> {
  int _attachmentCount = 0;
  bool _isLoading = true;
  final AttachmentApiService _apiService = AttachmentApiService();

  @override
  void initState() {
    super.initState();
    _loadAttachmentCount();
  }

  @override
  void didUpdateWidget(ResponsiveAttachmentButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.docNo != widget.docNo ||
        oldWidget.guidFixed != widget.guidFixed ||
        oldWidget.screenType != widget.screenType) {
      _loadAttachmentCount();
    }
  }

  Future<void> _loadAttachmentCount() async {
    if (widget.docNo.isEmpty || widget.guidFixed.isEmpty) {
      setState(() {
        _attachmentCount = 0;
        _isLoading = false;
      });
      return;
    }

    setState(() => _isLoading = true);

    try {
      final response = await _apiService.listAttachments(
        shopId: global.getShopId(),
        screenType: widget.screenType,
        docNo: widget.docNo,
        guidFixed: widget.guidFixed,
      );

      if (mounted) {
        setState(() {
          _attachmentCount = response.count;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _attachmentCount = 0;
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final isMobile = screenWidth < 600;
    final isTablet = screenWidth >= 600 && screenWidth < 1200;

    Widget iconWidget = Icon(
      Icons.attach_file,
      size: isMobile ? 20 : 22,
    );

    // Badge
    if (!_isLoading) {
      iconWidget = Badge(
        label: Text(
          _attachmentCount > 99 ? '99+' : _attachmentCount.toString(),
          style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
        ),
        backgroundColor: _attachmentCount > 0 ? Colors.blue : Colors.grey,
        child: iconWidget,
      );
    }

    // Mobile: Icon only
    if (isMobile) {
      return Padding(
        padding: EdgeInsets.symmetric(horizontal: 4.0),
        child: Tooltip(
          message: '${global.language("attach_document_file")} ($_attachmentCount ${global.language("files")})',
          child: IconButton(
            onPressed: () async {
              await showAttachmentDialog(
                context,
                docNo: widget.docNo,
                guidFixed: widget.guidFixed,
                screenType: widget.screenType,
              );
              _loadAttachmentCount();
            },
            icon: iconWidget,
            style: IconButton.styleFrom(
              padding: const EdgeInsets.all(12),
            ),
          ),
        ),
      );
    }

    // Tablet: Icon + short label
    if (isTablet) {
      return Padding(
        padding: EdgeInsets.symmetric(horizontal: 4.0),
        child: ElevatedButton.icon(
          onPressed: () async {
            await showAttachmentDialog(
              context,
              docNo: widget.docNo,
              guidFixed: widget.guidFixed,
              screenType: widget.screenType,
            );
            _loadAttachmentCount();
          },
          icon: iconWidget,
          label: Text(
            global.language("attach_file"),
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
          ),
          style: ElevatedButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),
      );
    }

    // Desktop: Icon + full label
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 4.0),
      child: ElevatedButton.icon(
        onPressed: () async {
          await showAttachmentDialog(
            context,
            docNo: widget.docNo,
            guidFixed: widget.guidFixed,
            screenType: widget.screenType,
          );
          _loadAttachmentCount();
        },
        icon: iconWidget,
        label: Text(
          '${global.language("attach_document_file")} ($_attachmentCount)',
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        ),
        style: ElevatedButton.styleFrom(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
        ),
      ),
    );
  }
}
