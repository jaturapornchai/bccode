import 'package:flutter/foundation.dart' show kIsWeb, kDebugMode;
import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/utils/backend_url_manager.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class AppConfigClass {
  String serviceClickhouse = '';
  String serviceApi = '';
  String servicePort = '';
  String reportApiPath = '';
  String reportApiPort = "";

  // GoAPI Backend (รวมเข้า MainAPI แล้ว — ใช้ port เดียวกัน path /goapi)
  String goApiUrl = '';
  String goApiPort = '';
}

AppConfigClass appConfigInit() {
  Environment env = Environment();
  AppConfigClass appConfig = AppConfigClass();

  // Get values from current environment
  appConfig.serviceClickhouse = env.config.serviceClickhouse;
  appConfig.serviceApi = env.config.reportApiPath;
  appConfig.reportApiPath = env.config.reportApiPath;
  appConfig.reportApiPort = env.config.reportApiPort;

  // Make sure URLs are correctly formatted for web
  if (kIsWeb) {
    // Force absolute URLs for web environment
    if (!appConfig.serviceClickhouse.startsWith('http')) {
      appConfig.serviceClickhouse = 'https://${appConfig.serviceClickhouse}';
    }
    if (!appConfig.serviceApi.startsWith('http')) {
      appConfig.serviceApi = 'https://${appConfig.serviceApi}';
    }
    if (!appConfig.reportApiPath.startsWith('http')) {
      appConfig.reportApiPath = 'https://${appConfig.reportApiPath}';
    }

    // Web release mode: ใช้ goApiUrl เดียวกับ reportApiPath (goapi ผ่าน Caddy proxy)
    if (!kDebugMode && appConfig.goApiUrl.isEmpty) {
      appConfig.goApiUrl = appConfig.reportApiPath;
      appConfig.goApiPort = '';
    }

    if (kDebugMode) {
      AppLogger.debug("Web environment detected, using absolute URLs");
      AppLogger.debug("serviceClickhouse: ${appConfig.serviceClickhouse}");
      AppLogger.debug("serviceApi: ${appConfig.serviceApi}");
      AppLogger.debug("reportApiPath: ${appConfig.reportApiPath}");
      AppLogger.debug("goApiUrl: ${appConfig.goApiUrl}");
    }
  }

  // goApiUrl — ตั้งจาก reportApiPath (Environment) ถ้ายังว่าง
  // Language API, Report, PDF, Approval อยู่ที่ goapi ไม่ใช่ mainapi
  if (appConfig.goApiUrl.isEmpty) {
    appConfig.goApiUrl = appConfig.reportApiPath;
    appConfig.goApiPort = appConfig.reportApiPort;
  }

  // ใช้ URL จาก BackendUrlManager ทุก platform
  // Web: มาจาก /config.json (server-side), Native: มาจาก user input หรือ config.json ข้างๆ exe
  final effectiveUrl = BackendUrlManager.getEffectiveUrl();
  if (effectiveUrl != null) {
    // เก็บ URL เต็ม (รวม path prefix เช่น /goapi) เพื่อให้ goApiUrlPath() สร้าง URL ถูกต้อง
    // ตัวอย่าง VPS:   https://dev-api.bcaicloud.com/goapi → goApiUrl = https://dev-api.bcaicloud.com/goapi
    // ตัวอย่าง Local: http://localhost:9090/goapi        → goApiUrl = http://localhost:9090/goapi
    final cleanUrl = effectiveUrl.endsWith('/')
        ? effectiveUrl.substring(0, effectiveUrl.length - 1)
        : effectiveUrl;
    appConfig.goApiUrl = cleanUrl;
    appConfig.goApiPort = ''; // port อยู่ใน URL แล้ว

    if (kDebugMode) {
      AppLogger.debug('=== appConfigInit (BackendUrlManager) ===');
      AppLogger.debug('goApiUrl: ${appConfig.goApiUrl}');
      AppLogger.debug('mainApiUrl: ${BackendUrlManager.getMainApiUrl()}');
    }
  } else {
    AppLogger.error('[appConfigInit] Backend URL ยังไม่ได้ตั้งค่า — กรุณาตั้งค่าในหน้า Login');
  }

  return appConfig;
}
