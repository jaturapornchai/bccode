import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// Modern Action Button with beautiful design
/// - สวยงาม modern
/// - มี hover effects
/// - Animation smooth
/// - Responsive design
class ModernActionButton extends StatefulWidget {
  final IconData icon;
  final String label;
  final String? shortLabel;
  final String? tooltip;
  final VoidCallback? onPressed;
  final int? badgeCount;
  final Color? primaryColor;
  final bool visible;
  final bool compact;

  const ModernActionButton({
    super.key,
    required this.icon,
    required this.label,
    this.shortLabel,
    this.tooltip,
    this.onPressed,
    this.badgeCount,
    this.primaryColor,
    this.visible = true,
    this.compact = false,
  });

  @override
  State<ModernActionButton> createState() => _ModernActionButtonState();
}

class _ModernActionButtonState extends State<ModernActionButton>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  bool _isHovered = false;
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 150),
      vsync: this,
    );
    _scaleAnimation = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _onHover(bool hovering) {
    setState(() => _isHovered = hovering);
    if (hovering) {
      _controller.forward();
    } else {
      _controller.reverse();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!widget.visible) return const SizedBox.shrink();

    final screenWidth = MediaQuery.of(context).size.width;
    final isMobile = screenWidth < 600;
    final isTablet = screenWidth >= 600 && screenWidth < 1200;
    final color = widget.primaryColor ?? Theme.of(context).primaryColor;

    return ScaleTransition(
      scale: _scaleAnimation,
      child: MouseRegion(
        onEnter: (_) => _onHover(true),
        onExit: (_) => _onHover(false),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          margin: EdgeInsets.symmetric(
            horizontal: widget.compact ? 2 : 4,
            vertical: 4,
          ),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            boxShadow: _isHovered
                ? [
                    BoxShadow(
                      color: color.withValues(alpha: 0.3),
                      blurRadius: 8,
                      offset: const Offset(0, 4),
                    ),
                  ]
                : [],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: widget.onPressed,
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
                            color.withValues(alpha: 0.1),
                            color.withValues(alpha: 0.05),
                          ],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        )
                      : null,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: _isHovered
                        ? color.withValues(alpha: 0.5)
                        : global.theme.dividerBorderColor.withValues(alpha: 0.3),
                    width: 1.5,
                  ),
                ),
                child: _buildContent(isMobile, isTablet, color),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildContent(bool isMobile, bool isTablet, Color color) {
    final iconWidget = _buildIcon(color);

    if (isMobile) {
      // Mobile: Icon only with tooltip
      return Tooltip(
        message: widget.tooltip ?? widget.label,
        child: iconWidget,
      );
    }

    // Tablet/Desktop: Icon + Label
    final labelText = isTablet
        ? (widget.shortLabel ?? widget.label)
        : widget.label;

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
              color: color,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }

  Widget _buildIcon(Color color) {
    Widget icon = Icon(
      widget.icon,
      size: 22,
      color: color,
    );

    if (widget.badgeCount != null && widget.badgeCount! > 0) {
      return Stack(
        clipBehavior: Clip.none,
        children: [
          icon,
          Positioned(
            right: -8,
            top: -8,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [global.theme.negativeHighlightTextColor, global.theme.negativeHighlightTextColor],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(10),
                boxShadow: [
                  BoxShadow(
                    color: global.theme.negativeHighlightTextColor.withValues(alpha: 0.4),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              constraints: const BoxConstraints(minWidth: 18, minHeight: 18),
              child: Text(
                widget.badgeCount! > 99 ? '99+' : widget.badgeCount.toString(),
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
      );
    }

    return icon;
  }
}

/// Action Button Group - จัดกลุ่มปุ่มที่เกี่ยวข้องกัน
class ActionButtonGroup extends StatelessWidget {
  final String? title;
  final List<Widget> buttons;
  final bool wrap;

  const ActionButtonGroup({
    super.key,
    this.title,
    required this.buttons,
    this.wrap = false,
  });

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final isMobile = screenWidth < 600;

    Widget buttonRow = wrap
        ? Wrap(
            spacing: 4,
            runSpacing: 4,
            children: buttons,
          )
        : Row(
            mainAxisSize: MainAxisSize.min,
            children: buttons,
          );

    if (title == null) return buttonRow;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (!isMobile)
          Padding(
            padding: const EdgeInsets.only(left: 8, bottom: 4),
            child: Text(
              title!,
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: global.theme.textSecondaryColor,
                letterSpacing: 0.5,
              ),
            ),
          ),
        buttonRow,
      ],
    );
  }
}

/// Divider สำหรับแยกกลุ่ม - สวยกว่าเดิม
class ModernButtonDivider extends StatelessWidget {
  final double height;
  final double width;
  final Color? color;

  const ModernButtonDivider({
    super.key,
    this.height = 32,
    this.width = 1,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      width: width,
      margin: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            Colors.transparent,
            (color ?? global.theme.dividerBorderColor).withValues(alpha: 0.5),
            Colors.transparent,
          ],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
        ),
      ),
    );
  }
}

/// AppBar Container with modern design
class ModernAppBarActions extends StatelessWidget {
  final List<Widget> children;
  final Color? backgroundColor;

  const ModernAppBarActions({
    super.key,
    required this.children,
    this.backgroundColor,
  });

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final isMobile = screenWidth < 600;

    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: isMobile ? 4 : 8,
        vertical: 4,
      ),
      decoration: BoxDecoration(
        color: backgroundColor ?? Colors.white.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: global.theme.cardColor.withValues(alpha: 0.1),
          width: 1,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: children,
      ),
    );
  }
}
