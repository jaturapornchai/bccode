import 'package:flutter/material.dart';
import 'package:bclms/app/core/theme/app_theme.dart';

/// หน้า placeholder สำหรับเมนูที่ยังไม่ได้พัฒนา — ใช้ร่วมกันทุกเมนู
class PlaceholderPage extends StatelessWidget {
  const PlaceholderPage({super.key});

  @override
  Widget build(BuildContext context) {
    final args = ModalRoute.of(context)?.settings.arguments as Map<String, dynamic>?;
    final title = args?['title'] ?? 'กำลังพัฒนา';

    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 100,
              height: 100,
              decoration: BoxDecoration(
                color: AppTheme.terracotta.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(
                Icons.construction_rounded,
                size: 48,
                color: AppTheme.terracotta.withValues(alpha: 0.5),
              ),
            ),
            const SizedBox(height: 24),
            Text(
              title,
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.w600,
                color: AppTheme.earthBrown,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'กำลังพัฒนา',
              style: TextStyle(
                fontSize: 14,
                color: AppTheme.earthBrown.withValues(alpha: 0.5),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
