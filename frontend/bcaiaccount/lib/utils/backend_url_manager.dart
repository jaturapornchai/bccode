import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// จัดการ Backend URL — เก็บ URL ปัจจุบัน + ประวัติใน SharedPreferences
class BackendUrlManager {
  static const String _currentUrlKey = 'backend_url';
  static const String _historyKey = 'backend_url_history';
  static const int _maxHistory = 10;

  /// goapi URL ที่โหลดจาก config.json
  /// ตั้งแต่รวม GoAPI เข้า MainAPI แล้ว — URL จะเป็น http://localhost:9090/goapi
  static String? _configFileUrl;

  /// mainapi URL ที่โหลดจาก config.json (optional)
  /// ถ้าไม่มี → derive จาก goapi URL (strip /goapi path)
  static String? _mainApiConfigUrl;

  /// โหลด config จาก config.json (เรียกครั้งเดียวตอน app startup)
  /// รูปแบบ:
  ///   VPS:   { "goapi_url": "https://dev-api.bcaicloud.com/goapi" }
  ///   Local: { "goapi_url": "http://localhost:9090/goapi" }
  static Future<void> loadConfig() async {
    if (kIsWeb) {
      await _loadWebConfig();
    } else {
      await _loadNativeConfig();
    }
  }

  static Future<void> _loadWebConfig() async {
    try {
      final response = await http.get(Uri.parse('/config.json'));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final goapiUrl = data['goapi_url'] as String?;
        if (goapiUrl != null && goapiUrl.isNotEmpty) {
          _configFileUrl = goapiUrl;
          AppLogger.info('[BackendUrlManager] goapi_url: $goapiUrl');
        }
        final mainApiUrl = data['mainapi_url'] as String?;
        if (mainApiUrl != null && mainApiUrl.isNotEmpty) {
          _mainApiConfigUrl = mainApiUrl;
          AppLogger.info('[BackendUrlManager] mainapi_url: $mainApiUrl');
        }
      }
    } catch (e) {
      AppLogger.warning('[BackendUrlManager] ไม่พบ web config.json: $e');
    }
  }

  /// Native: อ่าน config.json ในตำแหน่งเดียวกับ exe
  /// ตัวอย่าง Windows: C:\Program Files\BcMerchant\config.json
  /// - mainapi_url: โหลดเสมอ (ไม่ขึ้นกับ saved URL)
  /// - goapi_url: โหลดเฉพาะเมื่อยังไม่มี URL ที่ผู้ใช้ตั้งไว้
  static Future<void> _loadNativeConfig() async {
    try {
      final exeDir = File(Platform.resolvedExecutable).parent;
      final configFile = File('${exeDir.path}/config.json');
      if (!configFile.existsSync()) {
        AppLogger.info('[BackendUrlManager] ไม่พบ config.json ที่ ${exeDir.path}');
        return;
      }
      final raw = await configFile.readAsString();
      final data = jsonDecode(raw) as Map<String, dynamic>;

      // mainapi_url: โหลดเสมอ (ถ้ามี — optional เพราะ derive จาก goapi_url ได้)
      final mainApiUrl = data['mainapi_url'] as String?;
      if (mainApiUrl != null && mainApiUrl.isNotEmpty) {
        _mainApiConfigUrl = mainApiUrl;
        AppLogger.info('[BackendUrlManager] mainapi_url: $mainApiUrl');
      }

      // goapi_url: โหลดเฉพาะเมื่อผู้ใช้ยังไม่ได้ตั้งค่าเอง
      if (getCurrentUrl() == null) {
        final goapiUrl = data['goapi_url'] as String?;
        if (goapiUrl != null && goapiUrl.isNotEmpty) {
          _configFileUrl = goapiUrl;
          AppLogger.info('[BackendUrlManager] goapi_url: $goapiUrl');
        }
      }
    } catch (e) {
      AppLogger.error('[BackendUrlManager] อ่าน config.json ไม่ได้: $e');
    }
  }

  /// อ่าน URL ปัจจุบันที่ผู้ใช้ตั้งไว้ (null = ยังไม่เคยตั้ง)
  static String? getCurrentUrl() {
    return global.appConfig.getString(_currentUrlKey);
  }

  /// อ่าน goapi URL ที่ใช้งาน — null ถ้ายังไม่ได้ตั้งค่า
  /// Web: config.json จาก server (บังคับ)
  /// Native: URL ที่ผู้ใช้ตั้งเอง → config.json ข้างๆ exe → null
  static String? getEffectiveUrl() {
    if (kIsWeb) return _configFileUrl;
    final saved = getCurrentUrl();
    if (saved != null && saved.isNotEmpty) return saved;
    return _configFileUrl;
  }

  /// บันทึก URL ปัจจุบัน + เพิ่มเข้าประวัติ
  static Future<void> saveUrl(String url) async {
    final normalized = _normalizeUrl(url);
    if (normalized.isEmpty) return;

    await global.appConfig.setString(_currentUrlKey, normalized);

    final history = getHistory();
    history.remove(normalized);
    history.insert(0, normalized);

    if (history.length > _maxHistory) {
      history.removeRange(_maxHistory, history.length);
    }

    await global.appConfig.setString(_historyKey, jsonEncode(history));
  }

  /// อ่านประวัติ URL ทั้งหมด (เรียงจากล่าสุดก่อน)
  static List<String> getHistory() {
    final raw = global.appConfig.getString(_historyKey);
    if (raw == null || raw.isEmpty) return [];
    final List<dynamic> decoded = jsonDecode(raw);
    return decoded.cast<String>();
  }

  /// ลบ URL ออกจากประวัติ
  static Future<void> removeFromHistory(String url) async {
    final history = getHistory();
    history.remove(url);
    await global.appConfig.setString(_historyKey, jsonEncode(history));
  }

  /// ลบ URL ปัจจุบันออก (reset)
  static Future<void> clearCurrentUrl() async {
    await global.appConfig.remove(_currentUrlKey);
  }

  /// ดึง mainapi URL
  /// GoAPI รวมเข้า MainAPI แล้ว — port เดียวกัน
  /// 1. ใช้ mainapi_url จาก config.json ถ้ามี (explicit override)
  /// 2. Derive จาก goapi URL — strip /goapi path
  ///    https://dev-api.bcaicloud.com/goapi → https://dev-api.bcaicloud.com
  ///    http://localhost:9090/goapi → http://localhost:9090
  /// 3. null ถ้ายังไม่ได้ตั้งค่า
  static String? getMainApiUrl() {
    if (_mainApiConfigUrl != null) return _mainApiConfigUrl;
    final backendUrl = getEffectiveUrl();
    if (backendUrl == null) return null;
    final uri = Uri.tryParse(backendUrl);
    if (uri != null && uri.hasScheme) {
      final port = uri.hasPort ? ':${uri.port}' : '';
      return '${uri.scheme}://${uri.host}$port';
    }
    return null;
  }

  /// Normalize URL — เพิ่ม http:// ถ้าไม่มี, ลบ trailing slash
  static String _normalizeUrl(String url) {
    var trimmed = url.trim();
    if (trimmed.isEmpty) return '';

    if (!trimmed.startsWith('http://') && !trimmed.startsWith('https://')) {
      trimmed = 'http://$trimmed';
    }
    if (trimmed.endsWith('/')) {
      trimmed = trimmed.substring(0, trimmed.length - 1);
    }
    return trimmed;
  }
}
