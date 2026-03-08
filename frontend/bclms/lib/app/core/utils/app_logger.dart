// Global Logger สำหรับใช้ทั้ง Project
//
// การใช้งาน:
// ```dart
// import 'package:bclms/app/core/utils/app_logger.dart';
//
// // Basic logging
// AppLogger.debug('Debug message');
// AppLogger.info('Info message');
// AppLogger.warning('Warning message');
// AppLogger.error('Error message', error: exception);
// AppLogger.fatal('Critical error', error: exception, stackTrace: stackTrace);
//
// // Helper methods with emoji
// AppLogger.success('Operation completed');
// AppLogger.warn('This is a warning');
// AppLogger.err('This is an error');
// ```
//
// Output (คลิกที่ file path เพื่อเปิดไฟล์ได้เลย):
// ```
// [13:45:30.123] [ℹ️  INFO ] lib/app/services/auth_service.dart:45 เข้าสู่ระบบสำเร็จ
// [13:45:31.456] [⚠️  WARN ] lib/app/services/api_service.dart:89 ได้รับ 401
// [13:45:32.789] [❌ ERROR] lib/app/services/api_service.dart:103 Refresh ล้มเหลว
// ```

import 'package:flutter/foundation.dart';
import 'package:bclms/app/core/utils/logger/logger.dart';

class AppLogger {
  AppLogger._();

  static final Log _logger = LogImpl();

  /// Trace level — สำหรับ debug ละเอียดมาก
  static void trace(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.trace(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Debug level — สำหรับ debug ทั่วไป
  static void debug(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.debug(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Info level — สำหรับข้อมูลทั่วไป
  static void info(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.info(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Warning level — สำหรับคำเตือน
  static void warning(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.warn(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Error level — สำหรับ error
  static void error(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.error(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Fatal level — สำหรับ critical error
  static void fatal(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.dangerFailure(message, error: error, stackTrace: stackTrace);
    }
  }

  // ===== Helper Methods =====

  /// แสดง separator line
  static void separator([String title = '']) {
    if (kDebugMode) {
      if (title.isEmpty) {
        debug('═══════════════════════════════════════════════════════');
      } else {
        debug('═══════════════════════════════════════════════════════');
        debug('  $title');
        debug('═══════════════════════════════════════════════════════');
      }
    }
  }

  /// แสดง header
  static void header(String title) {
    if (kDebugMode) {
      separator();
      debug('🎯 $title');
      debug('───────────────────────────────────────────────────────');
    }
  }

  /// แสดง footer
  static void footer() {
    if (kDebugMode) {
      separator();
    }
  }

  /// แสดง section
  static void section(String title, Function() content) {
    if (kDebugMode) {
      header(title);
      content();
      footer();
    }
  }

  /// แสดง key-value pair
  static void keyValue(String key, dynamic value) {
    if (kDebugMode) {
      debug('   $key: $value');
    }
  }

  /// แสดง list items
  static void list(String title, List<dynamic> items) {
    if (kDebugMode) {
      debug('📋 $title (${items.length} items):');
      for (var i = 0; i < items.length; i++) {
        debug('   ${i + 1}. ${items[i]}');
      }
    }
  }

  /// แสดง success message
  static void success(String message) {
    if (kDebugMode) {
      info('✅ $message');
    }
  }

  /// แสดง warning message (shorthand)
  static void warn(String message) {
    if (kDebugMode) {
      warning('⚠️ $message');
    }
  }

  /// แสดง error message (shorthand)
  static void err(String message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.error('❌ $message', error: error, stackTrace: stackTrace);
    }
  }

  /// แสดง performance timing
  static void timing(String operation, int milliseconds) {
    if (kDebugMode) {
      if (milliseconds < 100) {
        debug('⚡ $operation: ${milliseconds}ms (เร็ว)');
      } else if (milliseconds < 500) {
        debug('🕐 $operation: ${milliseconds}ms (ปกติ)');
      } else {
        warning('🐌 $operation: ${milliseconds}ms (ช้า)');
      }
    }
  }

  /// แสดง object details
  static void object(String name, dynamic obj) {
    if (kDebugMode) {
      debug('📦 $name: $obj');
    }
  }

  /// แสดง network request
  static void request(String method, String url, {dynamic body}) {
    if (kDebugMode) {
      debug('🌐 [$method] $url');
      if (body != null) {
        debug('   Body: $body');
      }
    }
  }

  /// แสดง network response
  static void response(int statusCode, dynamic body) {
    if (kDebugMode) {
      if (statusCode >= 200 && statusCode < 300) {
        success('Response: $statusCode');
      } else {
        err('Response: $statusCode', error: body);
      }
    }
  }

  /// ล้าง log file
  static void clearLogFile() {
    LogImpl.clearLogFile();
  }
}
