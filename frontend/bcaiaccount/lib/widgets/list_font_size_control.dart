import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// ปุ่มปรับขนาด font + ความห่างบรรทัด ใน data list (A- size A+ ≡)
/// ใช้ร่วมกับ global.deviceConfig (persist ค่าผ่าน SharedPreferences)
class ListFontSizeControl extends StatelessWidget {
  /// callback เมื่อเปลี่ยนค่า (ปกติคือ setState)
  final VoidCallback onChanged;

  const ListFontSizeControl({super.key, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          // A- ลดขนาด font
          InkWell(
            onTap: () {
              global.listDataFontSizeDecrease();
              onChanged();
            },
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              child: Text(
                'A-',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: global.theme.textColor,
                ),
              ),
            ),
          ),
          // ขนาดปัจจุบัน
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: Text(
              '${global.deviceConfig.listDataFontSize.toInt()}',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: global.theme.primaryColor,
              ),
            ),
          ),
          // A+ เพิ่มขนาด font
          InkWell(
            onTap: () {
              global.listDataFontSizeIncrease();
              onChanged();
            },
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              child: Text(
                'A+',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: global.theme.textColor,
                ),
              ),
            ),
          ),
          const SizedBox(width: 4),
          // ≡ ปรับความห่างบรรทัด
          InkWell(
            onTap: () {
              global.listDataLineSpaceChange();
              onChanged();
            },
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
              child: Icon(
                Icons.density_small,
                size: 18,
                color: global.theme.textColor,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
