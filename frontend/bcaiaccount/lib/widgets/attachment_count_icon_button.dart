import 'package:flutter/material.dart';
import 'package:smlaicloud/services/attachment_api_service.dart';
import 'package:smlaicloud/widgets/attachment_dialog.dart';
import 'package:smlaicloud/global.dart' as global;

/// IconButton สำหรับแนบไฟล์ พร้อมแสดงจำนวนไฟล์แนบ (Badge)
class AttachmentCountIconButton extends StatefulWidget {
  final String docNo;
  final String guidFixed;
  final String screenType;
  final double iconSize;

  const AttachmentCountIconButton({
    super.key,
    required this.docNo,
    required this.guidFixed,
    required this.screenType,
    this.iconSize = 26.0,
  });

  @override
  State<AttachmentCountIconButton> createState() => _AttachmentCountIconButtonState();
}

class _AttachmentCountIconButtonState extends State<AttachmentCountIconButton> {
  int _attachmentCount = 0;
  bool _isLoading = true;
  final AttachmentApiService _apiService = AttachmentApiService();

  @override
  void initState() {
    super.initState();
    _loadAttachmentCount();
  }

  @override
  void didUpdateWidget(AttachmentCountIconButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    // โหลดใหม่ถ้า docNo หรือ guidFixed เปลี่ยน
    if (oldWidget.docNo != widget.docNo ||
        oldWidget.guidFixed != widget.guidFixed ||
        oldWidget.screenType != widget.screenType) {
      _loadAttachmentCount();
    }
  }

  /// โหลดจำนวนไฟล์แนบจาก API
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
    return IconButton(
      focusNode: FocusNode(skipTraversal: true),
      onPressed: () async {
        await showAttachmentDialog(
          context,
          docNo: widget.docNo,
          guidFixed: widget.guidFixed,
          screenType: widget.screenType,
        );
        // โหลดใหม่หลังปิด dialog (เผื่อมีการเพิ่มหรือลบไฟล์)
        _loadAttachmentCount();
      },
      icon: Stack(
        clipBehavior: Clip.none,
        children: [
          Icon(Icons.attach_file, size: widget.iconSize),
          // แสดง badge เสมอ (รวมถึง 0) และไม่กำลังโหลด
          if (!_isLoading)
            Positioned(
              right: -6,
              top: -6,
              child: Container(
                padding: const EdgeInsets.all(4),
                decoration: BoxDecoration(
                  color: _attachmentCount > 0 ? global.theme.infoHighlightTextColor : global.theme.iconSecondaryColor,
                  shape: BoxShape.circle,
                  border: Border.all(color: global.theme.cardColor, width: 1.5),
                ),
                constraints: const BoxConstraints(minWidth: 18, minHeight: 18),
                child: Text(
                  _attachmentCount > 99 ? '99+' : _attachmentCount.toString(),
                  style: TextStyle(
                    color: global.theme.onPrimaryColor,
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
            ),
        ],
      ),
      tooltip: global.language("attach_document_file"),
    );
  }
}
