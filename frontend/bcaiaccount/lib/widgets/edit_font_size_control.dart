import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// ปุ่มปรับขนาด font หน้าจอ edit (A- %  A+)
/// ใช้ร่วมกับ global.deviceConfig.editFontScale (persist ค่าผ่าน SharedPreferences)
class EditFontSizeControl extends StatelessWidget {
  /// callback เมื่อเปลี่ยนค่า (ปกติคือ setState)
  final VoidCallback onChanged;

  const EditFontSizeControl({super.key, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor.withValues(alpha: 0.3),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          // A- ลดขนาด font
          InkWell(
            onTap: () {
              global.editFontScaleDecrease();
              onChanged();
            },
            child: const Padding(
              padding: EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              child: Text(
                'A-',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
            ),
          ),
          // ขนาดปัจจุบัน (%)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: Text(
              '${global.deviceConfig.editFontScale.toInt()}%',
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.white70,
              ),
            ),
          ),
          // A+ เพิ่มขนาด font
          InkWell(
            onTap: () {
              global.editFontScaleIncrease();
              onChanged();
            },
            child: const Padding(
              padding: EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              child: Text(
                'A+',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
