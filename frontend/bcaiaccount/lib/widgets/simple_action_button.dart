import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// Simple Action Button - Icon + Text เสมอทุกขนาดหน้าจอ
/// Design: เรียบง่าย สวยงาม ไม่มี border ไม่มี background
class SimpleActionButton extends StatefulWidget {
  final IconData icon;
  final String label;
  final VoidCallback? onPressed;
  final int? badgeCount;
  final Color? color;
  final bool visible;
  final double iconSize;
  final double fontSize;

  const SimpleActionButton({
    super.key,
    required this.icon,
    required this.label,
    this.onPressed,
    this.badgeCount,
    this.color,
    this.visible = true,
    this.iconSize = 20.0,
    this.fontSize = 13.0,
  });

  @override
  State<SimpleActionButton> createState() => _SimpleActionButtonState();
}

class _SimpleActionButtonState extends State<SimpleActionButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    if (!widget.visible) return const SizedBox.shrink();

    final color = widget.color ?? global.theme.onPrimaryColor.withValues(alpha: 0.9);
    final hoverColor = widget.color ?? global.theme.onPrimaryColor;

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
          onTap: widget.onPressed,
          borderRadius: BorderRadius.circular(8),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Icon with badge
                _buildIconWithBadge(
                  _isHovered ? hoverColor : color,
                ),
                const SizedBox(width: 8),
                // Label
                AnimatedDefaultTextStyle(
                  duration: const Duration(milliseconds: 200),
                  style: TextStyle(
                    fontSize: widget.fontSize,
                    fontWeight: FontWeight.w500,
                    color: _isHovered ? hoverColor : color,
                    letterSpacing: 0.3,
                  ),
                  child: Text(
                    widget.label,
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
      widget.icon,
      size: widget.iconSize,
      color: color,
    );

    if (widget.badgeCount != null && widget.badgeCount! > 0) {
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
                color: global.theme.negativeHighlightTextColor,
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
                widget.badgeCount! > 99 ? '99+' : widget.badgeCount.toString(),
                style: TextStyle(
                  color: global.theme.onPrimaryColor,
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

/// Vertical Divider - เส้นแบ่งแนวตั้งสวยๆ
class SimpleButtonDivider extends StatelessWidget {
  final double height;
  final Color? color;

  const SimpleButtonDivider({
    super.key,
    this.height = 24,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      width: 1,
      margin: const EdgeInsets.symmetric(horizontal: 8),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            Colors.transparent,
            (color ?? global.theme.onPrimaryColor.withValues(alpha: 0.3)),
            Colors.transparent,
          ],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
        ),
      ),
    );
  }
}

/// Button Group Label - หัวข้อกลุ่มปุ่ม
class ButtonGroupLabel extends StatelessWidget {
  final String label;
  final Color? color;

  const ButtonGroupLabel({
    super.key,
    required this.label,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 12, right: 12, bottom: 4),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w600,
          color: (color ?? global.theme.onPrimaryColor).withValues(alpha: 0.6),
          letterSpacing: 1.0,
        ),
      ),
    );
  }
}
