import 'package:flutter/material.dart';

/// Responsive Action Button สำหรับ AppBar
/// รองรับหน้าจอตั้งแต่เล็ก (mobile) ถึงใหญ่ (desktop)
///
/// Breakpoints:
/// - < 600: Mobile - แสดงแค่ icon + tooltip
/// - 600-1200: Tablet - แสดง icon + label สั้น
/// - > 1200: Desktop - แสดง icon + label เต็ม
class ResponsiveActionButton extends StatelessWidget {
  /// Icon ของปุ่ม
  final IconData icon;

  /// ข้อความ label เต็ม (สำหรับหน้าจอใหญ่)
  final String label;

  /// ข้อความ label สั้น (สำหรับหน้าจอกลาง) - ถ้าไม่ระบุจะใช้ label
  final String? shortLabel;

  /// Tooltip (สำหรับหน้าจอเล็ก)
  final String tooltip;

  /// Callback เมื่อกดปุ่ม
  final VoidCallback? onPressed;

  /// จำนวนแสดงใน badge (ถ้ามี)
  final int? badgeCount;

  /// สีพื้นหลังของปุ่ม
  final Color? backgroundColor;

  /// สีของ icon และ label
  final Color? foregroundColor;

  /// ขนาดของ icon
  final double iconSize;

  /// แสดงปุ่มหรือไม่
  final bool visible;

  /// สไตล์ของปุ่ม
  final ActionButtonStyle style;

  const ResponsiveActionButton({
    super.key,
    required this.icon,
    required this.label,
    required this.tooltip,
    this.shortLabel,
    this.onPressed,
    this.badgeCount,
    this.backgroundColor,
    this.foregroundColor,
    this.iconSize = 20.0,
    this.visible = true,
    this.style = ActionButtonStyle.elevated,
  });

  @override
  Widget build(BuildContext context) {
    if (!visible) return const SizedBox.shrink();

    return LayoutBuilder(
      builder: (context, constraints) {
        final screenWidth = MediaQuery.of(context).size.width;

        // กำหนดโหมดการแสดงผลตามขนาดหน้าจอ
        final displayMode = _getDisplayMode(screenWidth);

        return Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4.0),
          child: _buildButton(context, displayMode),
        );
      },
    );
  }

  /// กำหนดโหมดการแสดงผลตามความกว้างหน้าจอ
  ButtonDisplayMode _getDisplayMode(double width) {
    if (width < 600) {
      return ButtonDisplayMode.iconOnly;
    } else if (width < 1200) {
      return ButtonDisplayMode.iconWithShortLabel;
    } else {
      return ButtonDisplayMode.iconWithFullLabel;
    }
  }

  /// สร้างปุ่มตามโหมดที่เลือก
  Widget _buildButton(BuildContext context, ButtonDisplayMode mode) {
    Widget iconWidget = Icon(icon, size: iconSize);

    // เพิ่ม badge ถ้ามี
    if (badgeCount != null && badgeCount! > 0) {
      iconWidget = Badge(
        label: Text(
          badgeCount! > 99 ? '99+' : badgeCount.toString(),
          style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
        ),
        backgroundColor: badgeCount! > 0 ? Colors.red : Colors.grey,
        child: iconWidget,
      );
    }

    switch (mode) {
      case ButtonDisplayMode.iconOnly:
        return _buildIconOnlyButton(context, iconWidget);

      case ButtonDisplayMode.iconWithShortLabel:
        return _buildIconWithLabelButton(
          context,
          iconWidget,
          shortLabel ?? label,
        );

      case ButtonDisplayMode.iconWithFullLabel:
        return _buildIconWithLabelButton(context, iconWidget, label);
    }
  }

  /// สร้างปุ่มแบบ icon อย่างเดียว (สำหรับหน้าจอเล็ก)
  Widget _buildIconOnlyButton(BuildContext context, Widget iconWidget) {
    switch (style) {
      case ActionButtonStyle.elevated:
        return Tooltip(
          message: tooltip,
          child: ElevatedButton(
            onPressed: onPressed,
            style: ElevatedButton.styleFrom(
              backgroundColor: backgroundColor,
              foregroundColor: foregroundColor,
              padding: const EdgeInsets.all(12),
              minimumSize: const Size(48, 48),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: iconWidget,
          ),
        );

      case ActionButtonStyle.outlined:
        return Tooltip(
          message: tooltip,
          child: OutlinedButton(
            onPressed: onPressed,
            style: OutlinedButton.styleFrom(
              foregroundColor: foregroundColor,
              padding: const EdgeInsets.all(12),
              minimumSize: const Size(48, 48),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: iconWidget,
          ),
        );

      case ActionButtonStyle.text:
        return Tooltip(
          message: tooltip,
          child: TextButton(
            onPressed: onPressed,
            style: TextButton.styleFrom(
              foregroundColor: foregroundColor,
              padding: const EdgeInsets.all(12),
              minimumSize: const Size(48, 48),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: iconWidget,
          ),
        );

      case ActionButtonStyle.icon:
        return Tooltip(
          message: tooltip,
          child: IconButton(
            onPressed: onPressed,
            icon: iconWidget,
            color: foregroundColor,
            iconSize: iconSize,
            style: IconButton.styleFrom(
              backgroundColor: backgroundColor,
              padding: const EdgeInsets.all(12),
            ),
          ),
        );
    }
  }

  /// สร้างปุ่มแบบมี label (สำหรับหน้าจอกลาง-ใหญ่)
  Widget _buildIconWithLabelButton(
    BuildContext context,
    Widget iconWidget,
    String labelText,
  ) {
    switch (style) {
      case ActionButtonStyle.elevated:
        return ElevatedButton.icon(
          onPressed: onPressed,
          icon: iconWidget,
          label: Text(
            labelText,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            overflow: TextOverflow.ellipsis,
          ),
          style: ElevatedButton.styleFrom(
            backgroundColor: backgroundColor,
            foregroundColor: foregroundColor,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        );

      case ActionButtonStyle.outlined:
        return OutlinedButton.icon(
          onPressed: onPressed,
          icon: iconWidget,
          label: Text(
            labelText,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            overflow: TextOverflow.ellipsis,
          ),
          style: OutlinedButton.styleFrom(
            foregroundColor: foregroundColor,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        );

      case ActionButtonStyle.text:
        return TextButton.icon(
          onPressed: onPressed,
          icon: iconWidget,
          label: Text(
            labelText,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            overflow: TextOverflow.ellipsis,
          ),
          style: TextButton.styleFrom(
            foregroundColor: foregroundColor,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        );

      case ActionButtonStyle.icon:
        // กรณี icon style แต่มี label ให้แสดงเป็น row
        return Tooltip(
          message: tooltip,
          child: Material(
            color: backgroundColor ?? Colors.transparent,
            borderRadius: BorderRadius.circular(8),
            child: InkWell(
              onTap: onPressed,
              borderRadius: BorderRadius.circular(8),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    iconWidget,
                    const SizedBox(width: 8),
                    Text(
                      labelText,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                        color: foregroundColor,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
    }
  }
}

/// โหมดการแสดงผลของปุ่ม
enum ButtonDisplayMode {
  /// แสดงแค่ icon (mobile)
  iconOnly,

  /// แสดง icon + label สั้น (tablet)
  iconWithShortLabel,

  /// แสดง icon + label เต็ม (desktop)
  iconWithFullLabel,
}

/// สไตล์ของปุ่ม
enum ActionButtonStyle {
  /// ElevatedButton (มีพื้นหลัง)
  elevated,

  /// OutlinedButton (มีกรอบ)
  outlined,

  /// TextButton (ไม่มีพื้นหลัง ไม่มีกรอบ)
  text,

  /// IconButton (แบบเดิม)
  icon,
}

/// Divider สำหรับแยกกลุ่มปุ่มใน AppBar
class ActionButtonDivider extends StatelessWidget {
  final double height;
  final Color? color;

  const ActionButtonDivider({
    super.key,
    this.height = 32,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      width: 1,
      margin: const EdgeInsets.symmetric(horizontal: 8),
      color: color ?? Colors.white.withValues(alpha: 0.3),
    );
  }
}
