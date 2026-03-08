import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget สำหรับแสดง Loading Overlay
/// ใช้สำหรับแสดงสถานะ Loading ทับหน้าจอหรือ Dialog
class LoadingOverlay extends StatelessWidget {
  final String message;
  final String? subtitle;

  const LoadingOverlay({super.key, required this.message, this.subtitle});

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.black.withValues(alpha: 0.5),
      child: Center(
        child: Container(
          padding: const EdgeInsets.all(24),
          margin: const EdgeInsets.symmetric(horizontal: 32),
          decoration: BoxDecoration(
            color: global.theme.backgroundColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: global.theme.appBarColor.withValues(alpha: 0.2),
              width: 1,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.15),
                blurRadius: 20,
                offset: const Offset(0, 8),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Loading indicator
              const SizedBox(
                width: 48,
                height: 48,
                child: CircularProgressIndicator(
                  strokeWidth: 3,
                  valueColor: AlwaysStoppedAnimation<Color>(Colors.blue),
                  backgroundColor: Color(0x1A2196F3),
                ),
              ),

              const SizedBox(height: 20),

              // Loading text
              Text(
                message,
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w500,
                ),
                textAlign: TextAlign.center,
              ),

              const SizedBox(height: 8),

              // Subtitle
              Text(
                subtitle ?? global.language('please_wait'),
                style: TextStyle(
                  color: Colors.black87.withValues(alpha: 0.6),
                  fontSize: 13,
                  fontWeight: FontWeight.w400,
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
