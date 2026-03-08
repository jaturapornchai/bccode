import 'package:flutter/material.dart';
import 'package:smlaicloud/usersystem/login_password_screen.dart';

/// Widget สำหรับเลือกแสดง Login screen ตาม flavor
/// ทุก flavor ใช้ LoginPasswordScreen ที่มีครบทั้ง:
/// - Username/Password login
/// - Google Sign-In
/// - LINE Login
class FlavorLoginSelector extends StatelessWidget {
  const FlavorLoginSelector({super.key});

  @override
  Widget build(BuildContext context) {
    // ทุก flavor ใช้ LoginPasswordScreen ที่มี login ครบทุกแบบ
    return const LoginPasswordScreen();
  }
}
