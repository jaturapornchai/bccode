import 'package:flutter/material.dart';

enum Flavor {
  bcaidev,
  bcaiuat,
  bcaiprod,
}

class F {
  static Flavor? appFlavor;

  static String get name => appFlavor?.name ?? '';

  /// สีหลักของ gradient (สีแรก - ด้านซ้ายบน)
  /// ธีมบ้านเชียง - สีน้ำตาลแดง (Terracotta)
  static Color get primaryGradientColor => const Color(0xFFB85C38);

  /// สีรองของ gradient (สีที่สอง - ด้านขวาล่าง)
  /// ธีมบ้านเชียง - สีทอง
  static Color get secondaryGradientColor => const Color(0xFFD4A373);

  /// สี shadow ของ gradient
  static Color get gradientShadowColor => const Color(0xFFB85C38);

  static String get title {
    switch (appFlavor) {
      case Flavor.bcaidev:
        return 'BC Ai Cloud DEV';
      case Flavor.bcaiuat:
        return 'BC Ai Cloud UAT';
      case Flavor.bcaiprod:
        return 'BC Ai Cloud';
      default:
        return 'BC Ai Cloud';
    }
  }

  /// ชื่อบริษัท
  static String get companyName => 'บริษัท บ้านเชียงซอฟท์ จำกัด';

  /// Path ของ Logo
  static String get logoPath => 'assets/logos/logo_bc_ai_cloud.png';

  /// ดึงชื่อ App (ไม่รวม environment badge)
  static String get appName => 'BC Ai Cloud';

  /// ตรวจสอบว่าเป็น DEV flavor หรือไม่
  static bool get isDev => appFlavor == Flavor.bcaidev;

  /// ตรวจสอบว่าเป็น UAT flavor หรือไม่
  static bool get isUat => appFlavor == Flavor.bcaiuat;

  /// ตรวจสอบว่าเป็น PROD flavor หรือไม่
  static bool get isProd => appFlavor == Flavor.bcaiprod;
}
