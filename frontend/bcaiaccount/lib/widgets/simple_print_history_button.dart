import 'package:flutter/material.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// Simple Print History Button - มี Icon + Text เสมอ พร้อม Badge จำนวนครั้งที่พิมพ์
class SimplePrintHistoryButton extends StatefulWidget {
  final String collection;
  final String docNo;
  final String title;
  final Color? color;
  final bool visible;
  final double iconSize;
  final double fontSize;

  const SimplePrintHistoryButton({
    super.key,
    required this.collection,
    required this.docNo,
    required this.title,
    this.color,
    this.visible = true,
    this.iconSize = 20.0,
    this.fontSize = 13.0,
  });

  @override
  State<SimplePrintHistoryButton> createState() =>
      _SimplePrintHistoryButtonState();
}

class _SimplePrintHistoryButtonState extends State<SimplePrintHistoryButton> {
  int _printCount = 0;
  bool _isHovered = false;
  final PdfService _pdfService = PdfService();

  @override
  void initState() {
    super.initState();
    _loadPrintCount();
  }

  @override
  void didUpdateWidget(SimplePrintHistoryButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.docNo != widget.docNo ||
        oldWidget.collection != widget.collection) {
      _loadPrintCount();
    }
  }

  Future<void> _loadPrintCount() async {
    if (widget.docNo.isEmpty) {
      if (mounted) {
        setState(() {
          _printCount = 0;
        });
      }
      return;
    }

    final count = await _pdfService.getPrintHistoryCount(
      collection: widget.collection,
      docNo: widget.docNo,
    );

    if (mounted) {
      setState(() {
        _printCount = count;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!widget.visible) return const SizedBox.shrink();

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
            await _pdfService.showPrintHistoryDialog(
              context: context,
              collection: widget.collection,
              docNo: widget.docNo,
              title: widget.title,
            );
            _loadPrintCount();
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
                    '${global.language("print_history")}${_printCount > 0 ? ' ($_printCount)' : ''}',
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
      Icons.history,
      size: widget.iconSize,
      color: color,
    );

    if (_printCount > 0) {
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
                color: Colors.red.shade500,
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
                _printCount > 99 ? '99+' : _printCount.toString(),
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
