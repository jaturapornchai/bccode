import 'package:flutter/material.dart';
import 'package:smlaicloud/services/attachment_api_service.dart';
import 'package:smlaicloud/widgets/attachment_dialog.dart';
import 'package:smlaicloud/global.dart' as global;

/// Modern Attachment Button - สวยงาม มี animation
class ModernAttachmentButton extends StatefulWidget {
  final String docNo;
  final String guidFixed;
  final String screenType;

  const ModernAttachmentButton({
    super.key,
    required this.docNo,
    required this.guidFixed,
    required this.screenType,
  });

  @override
  State<ModernAttachmentButton> createState() =>
      _ModernAttachmentButtonState();
}

class _ModernAttachmentButtonState extends State<ModernAttachmentButton>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  int _attachmentCount = 0;
  bool _isLoading = true;
  bool _isHovered = false;
  final AttachmentApiService _apiService = AttachmentApiService();
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _loadAttachmentCount();
    _pulseController = AnimationController(
      duration: const Duration(milliseconds: 1500),
      vsync: this,
    )..repeat(reverse: true);
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.1).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(ModernAttachmentButton oldWidget) {
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

  void _onHover(bool hovering) {
    setState(() => _isHovered = hovering);
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final isMobile = screenWidth < 600;
    final isTablet = screenWidth >= 600 && screenWidth < 1200;
    final primaryColor = global.theme.infoHighlightTextColor;

    return MouseRegion(
      onEnter: (_) => _onHover(true),
      onExit: (_) => _onHover(false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          boxShadow: _isHovered
              ? [
                  BoxShadow(
                    color: primaryColor.withValues(alpha: 0.3),
                    blurRadius: 8,
                    offset: const Offset(0, 4),
                  ),
                ]
              : [],
        ),
        child: Material(
          color: Colors.transparent,
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
            borderRadius: BorderRadius.circular(12),
            child: Container(
              padding: EdgeInsets.symmetric(
                horizontal: isMobile ? 12 : (isTablet ? 16 : 20),
                vertical: isMobile ? 10 : 12,
              ),
              decoration: BoxDecoration(
                gradient: _isHovered
                    ? LinearGradient(
                        colors: [
                          primaryColor.withValues(alpha: 0.1),
                          primaryColor.withValues(alpha: 0.05),
                        ],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      )
                    : null,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: _isHovered
                      ? primaryColor.withValues(alpha: 0.5)
                      : global.theme.dividerBorderColor.withValues(alpha: 0.3),
                  width: 1.5,
                ),
              ),
              child: _buildContent(isMobile, isTablet, primaryColor),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildContent(bool isMobile, bool isTablet, Color primaryColor) {
    final iconWidget = _buildIconWithBadge(primaryColor);

    if (isMobile) {
      return Tooltip(
        message: '${global.language("attach_document_file")} ($_attachmentCount ${global.language("files")})',
        child: iconWidget,
      );
    }

    final labelText = isTablet
        ? global.language("attach_file")
        : '${global.language("attach_document_file")}${_attachmentCount > 0 ? ' ($_attachmentCount)' : ''}';

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        iconWidget,
        const SizedBox(width: 8),
        Flexible(
          child: Text(
            labelText,
            style: TextStyle(
              fontSize: isMobile ? 12 : (isTablet ? 13 : 14),
              fontWeight: FontWeight.w600,
              color: primaryColor,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }

  Widget _buildIconWithBadge(Color primaryColor) {
    Widget icon = Icon(
      Icons.attach_file_rounded,
      size: 22,
      color: primaryColor,
    );

    if (_isLoading) {
      return SizedBox(
        width: 22,
        height: 22,
        child: CircularProgressIndicator(
          strokeWidth: 2,
          valueColor: AlwaysStoppedAnimation<Color>(primaryColor),
        ),
      );
    }

    if (_attachmentCount > 0) {
      return Stack(
        clipBehavior: Clip.none,
        children: [
          icon,
          Positioned(
            right: -8,
            top: -8,
            child: ScaleTransition(
              scale: _pulseAnimation,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      global.theme.infoHighlightTextColor,
                      global.theme.infoHighlightTextColor,
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(10),
                  boxShadow: [
                    BoxShadow(
                      color: global.theme.infoHighlightTextColor.withValues(alpha: 0.4),
                      blurRadius: 6,
                      offset: const Offset(0, 2),
                    ),
                  ],
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
          ),
        ],
      );
    }

    // No attachments - show subtle indicator
    return Stack(
      clipBehavior: Clip.none,
      children: [
        icon,
        Positioned(
          right: -8,
          top: -8,
          child: Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              color: global.theme.iconSecondaryColor,
              shape: BoxShape.circle,
            ),
          ),
        ),
      ],
    );
  }
}
