// Global Logger สำหรับใช้ทั้ง Project
//
// การใช้งาน:
// ```dart
// import 'package:dedecashier/core/logger/app_logger.dart';
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
// AppLogger.warn('This is a warning');  // Adds ⚠️ automatically
// AppLogger.err('This is an error');    // Adds ❌ automatically
// ```
//
// Output จะแสดงเป็น (คลิกที่ file path เพื่อเปิดไฟล์ได้เลย):
// ```
// [13:45:30.123] [🐛 DEBUG] lib/features/pos/presentation/screens/pos_screen.dart:456 Debug message
// [13:45:31.456] [ℹ️  INFO ] lib/features/pos/presentation/screens/pos_process.dart:789 Info message
// [13:45:32.789] [⚠️  WARN ] lib/util/pos_util.dart:123 Warning message
// [13:45:33.012] [❌ ERROR] lib/util/pos_print.dart:234 Error message
// ```
//
// 💡 Tip: คลิกที่ file path ใน Debug Console จะเปิดไฟล์และไปที่บรรทัดนั้นเลย!

import 'package:flutter/foundation.dart';
import 'package:smlaicloud/utils/logger/logger.dart';

class AppLogger {
  // ป้องกันการสร้าง instance
  AppLogger._();

  // สร้าง Logger instance โดยตรง (ไม่ต้องพึ่ง GetIt/ServiceLocator)
  static final Log _logger = LogImpl();

  /// Trace level - สำหรับ debug ละเอียดมาก
  static void trace(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.trace(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Debug level - สำหรับ debug ทั่วไป
  static void debug(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.debug(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Info level - สำหรับข้อมูลทั่วไป
  static void info(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.info(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Warning level - สำหรับคำเตือน
  static void warning(
    dynamic message, {
    dynamic error,
    StackTrace? stackTrace,
  }) {
    if (kDebugMode) {
      _logger.warn(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Error level - สำหรับ error
  static void error(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.error(message, error: error, stackTrace: stackTrace);
    }
  }

  /// Fatal level - สำหรับ critical error
  static void fatal(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      _logger.dangerFailure(message, error: error, stackTrace: stackTrace);
    }
  }

  // ===== Helper Methods สำหรับ Debug แบบกลุ่ม =====

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

  /// แสดง success message with emoji
  static void success(String message) {
    if (kDebugMode) {
      info('✅ $message');
    }
  }

  /// แสดง warning message with emoji (shorthand)
  static void warn(String message) {
    if (kDebugMode) {
      warning('⚠️ $message');
    }
  }

  /// แสดง error message with emoji (shorthand)
  static void err(String message, {dynamic err, StackTrace? stackTrace}) {
    if (kDebugMode) {
      // ใช้ _logger.error แทน AppLogger.error เพื่อหลีกเลี่ยง parameter shadowing
      _logger.error('❌ $message', error: err, stackTrace: stackTrace);
    }
  }

  /// แสดง performance timing
  static void timing(String operation, int milliseconds) {
    if (kDebugMode) {
      if (milliseconds < 100) {
        debug('⚡ $operation: ${milliseconds}ms (Fast)');
      } else if (milliseconds < 500) {
        debug('🕐 $operation: ${milliseconds}ms (Normal)');
      } else {
        warning('🐌 $operation: ${milliseconds}ms (Slow)');
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
        err('Response: $statusCode', err: body);
      }
    }
  }

  /// ล้าง log file (เรียกตอน hot reload/restart)
  /// Log file อยู่ที่ logs/flutter_debug.log ใน working directory
  static void clearLogFile() {
    LogImpl.clearLogFile();
  }
}
