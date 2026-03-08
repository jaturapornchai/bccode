import 'package:flutter/material.dart';
import 'package:smlaicloud/services/attachment_api_service.dart';
import 'package:smlaicloud/widgets/attachment_dialog.dart';
import 'package:smlaicloud/global.dart' as global;

/// Simple Attachment Button - มี Icon + Text เสมอ พร้อม Badge
class SimpleAttachmentButton extends StatefulWidget {
  final String docNo;
  final String guidFixed;
  final String screenType;
  final Color? color;
  final double iconSize;
  final double fontSize;

  const SimpleAttachmentButton({
    super.key,
    required this.docNo,
    required this.guidFixed,
    required this.screenType,
    this.color,
    this.iconSize = 20.0,
    this.fontSize = 13.0,
  });

  @override
  State<SimpleAttachmentButton> createState() =>
      _SimpleAttachmentButtonState();
}

class _SimpleAttachmentButtonState extends State<SimpleAttachmentButton> {
  int _attachmentCount = 0;
  bool _isHovered = false;
  final AttachmentApiService _apiService = AttachmentApiService();

  @override
  void initState() {
    super.initState();
    _loadAttachmentCount();
  }

  @override
  void didUpdateWidget(SimpleAttachmentButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.docNo != widget.docNo ||
        oldWidget.guidFixed != widget.guidFixed ||
        oldWidget.screenType != widget.screenType) {
      _loadAttachmentCount();
    }
  }

  Future<void> _loadAttachmentCount() async {
    if (widget.docNo.isEmpty || widget.guidFixed.isEmpty) {
      if (mounted) {
        setState(() {
          _attachmentCount = 0;
        });
      }
      return;
    }

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
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _attachmentCount = 0;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final color = widget.color ?? Colors.white.withValues(alpha: 0.9);
    final hoverColor = widget.color ?? Colors.white;

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        decoration: BoxDecoration(
          color: _isHovered ? Colors.white.withValues(alpha: 0.15) : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
        ),
        child: InkWell(
          onTap: () async {
            await showAttachmentDialog(
              context,
              docNo: widget.docNo,
              guidFixed: widget.guidFixed,
              screenType: widget.screenType,
            );
            _loadAttachmentCount();
          },
          borderRadius: BorderRadius.circular(8),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: 8, vertical: 6),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Icon with badge/loading
                _buildIconWithBadge(
                  _isHovered ? hoverColor : color,
                ),
                const SizedBox(width: 8),
                // Label with count
                AnimatedDefaultTextStyle(
                  duration: const Duration(milliseconds: 200),
                  style: TextStyle(
                    fontSize: widget.fontSize,
                    fontWeight: FontWeight.w500,
                    color: _isHovered ? hoverColor : color,
                    letterSpacing: 0.3,
                  ),
                  child: Text(
                    '${global.language("attach_file")}${_attachmentCount > 0 ? ' ($_attachmentCount)' : ''}',
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildIconWithBadge(Color color) {
    Widget iconWidget = Icon(
      Icons.attach_file_rounded,
      size: widget.iconSize,
      color: color,
    );

    if (_attachmentCount > 0) {
      return Stack(
        clipBehavior: Clip.none,
        children: [
          iconWidget,
          Positioned(
            right: -6,
            top: -6,
            child: Container(
              padding: const EdgeInsets.all(4),
              decoration: BoxDecoration(
                color: Colors.blue.shade500,
                shape: BoxShape.circle,
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.3),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              constraints: const BoxConstraints(minWidth: 16, minHeight: 16),
              child: Text(
                _attachmentCount > 99 ? '99+' : _attachmentCount.toString(),
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 9,
                  fontWeight: FontWeight.bold,
                  height: 1,
                ),
                textAlign: TextAlign.center,
              ),
            ),
          ),
        ],
      );
    }

    return iconWidget;
  }
}
