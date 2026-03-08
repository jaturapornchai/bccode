import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';

import 'package:bclms/app/core/utils/logger/log_output.dart';

/// Log interface
abstract class Log {
  void trace(dynamic message, {dynamic error, StackTrace? stackTrace});
  void debug(dynamic message, {dynamic error, StackTrace? stackTrace});
  void info(dynamic message, {dynamic error, StackTrace? stackTrace});
  void warn(dynamic message, {dynamic error, StackTrace? stackTrace});
  void error(dynamic message, {dynamic error, StackTrace? stackTrace});
  void dangerFailure(dynamic message, {dynamic error, StackTrace? stackTrace});
}

/// Custom Logger Printer ที่แสดง file:/// URI ให้คลิกได้ใน VS Code Debug Console
class CustomLogPrinter extends LogPrinter {
  final bool printTime;
  final bool printEmojis;

  // Cache project root path (ดึงจาก platform-specific implementation)
  static String? _projectRoot;

  static String get projectRoot {
    _projectRoot ??= getProjectRoot();
    return _projectRoot!;
  }

  CustomLogPrinter({
    this.printTime = true,
    this.printEmojis = true,
  });

  static const _levelLabels = {
    Level.trace: '🔍 TRACE',
    Level.debug: '🐛 DEBUG',
    Level.info: 'ℹ️  INFO ',
    Level.warning: '⚠️  WARN ',
    Level.error: '❌ ERROR',
    Level.fatal: '💀 FATAL',
  };

  @override
  List<String> log(LogEvent event) {
    final message = event.message;
    final fileInfo = _getFileInfo();

    final output = StringBuffer();

    if (printTime) {
      final now = DateTime.now();
      final time =
          '${_pad(now.hour)}:${_pad(now.minute)}:${_pad(now.second)}.${_pad3(now.millisecond)}';
      output.write('[$time] ');
    }

    final label = _levelLabels[event.level] ?? 'UNKNOWN';
    output.write('[$label] ');
    output.write('($fileInfo) ');
    output.write(message);

    if (event.error != null) {
      output.write('\n    Error: ${event.error}');
    }
    if (event.stackTrace != null) {
      output.write('\n${event.stackTrace}');
    }

    return [output.toString()];
  }

  String _pad(int n) => n.toString().padLeft(2, '0');
  String _pad3(int n) => n.toString().padLeft(3, '0');

  /// ดึง file:/// URI จาก StackTrace ให้คลิกได้ใน VS Code Debug Console
  /// Format: file:///c:/git/bclms/lib/path/file.dart:line:column
  /// บน Web จะแสดงเป็น package path แทน (file:/// ไม่ทำงานบน browser)
  String _getFileInfo() {
    try {
      final stackTrace = StackTrace.current.toString();
      final lines = stackTrace.split('\n');

      for (var line in lines) {
        // ข้าม logger files
        if (line.contains('logger.dart') ||
            line.contains('app_logger.dart') ||
            line.contains('log_output') ||
            line.contains('<asynchronous suspension>')) {
          continue;
        }

        // ลอง match file:/// URI ก่อน (บาง runtime ใช้ file URI โดยตรง)
        final fileUriMatch = RegExp(r'\((file:///[^)]+)\)').firstMatch(line);
        if (fileUriMatch != null) {
          return fileUriMatch.group(1)!;
        }

        // Match package:bclms/... → แปลงเป็น file:/// URI (ถ้ามี project root)
        if (line.contains('package:bclms/')) {
          final match = RegExp(
            r'package:bclms/(.+\.dart):(\d+):(\d+)',
          ).firstMatch(line);
          if (match != null) {
            final filePath = match.group(1);
            final lineNum = match.group(2);
            final colNum = match.group(3);

            // Native: ใช้ file:/// URI ให้คลิกได้
            if (projectRoot.isNotEmpty) {
              return 'file:///$projectRoot/lib/$filePath:$lineNum:$colNum';
            }
            // Web: แสดง package path
            return 'lib/$filePath:$lineNum:$colNum';
          }

          // match แบบไม่มี column
          final match2 = RegExp(
            r'package:bclms/(.+\.dart):(\d+)',
          ).firstMatch(line);
          if (match2 != null) {
            final filePath = match2.group(1);
            final lineNum = match2.group(2);
            if (projectRoot.isNotEmpty) {
              return 'file:///$projectRoot/lib/$filePath:$lineNum';
            }
            return 'lib/$filePath:$lineNum';
          }
        }
      }
    } catch (_) {}
    return 'unknown';
  }
}

/// Log implementation ใช้ logger package
/// - Native (Windows/Android/iOS): console + file logging
/// - Web: console only
class LogImpl implements Log {
  late final Logger logger;

  LogImpl() {
    logger = Logger(
      printer: CustomLogPrinter(printTime: true, printEmojis: true),
      output: createLogOutput(),
      level: kDebugMode ? Level.trace : Level.info,
    );
  }

  /// ล้าง log file (Native only, no-op บน Web)
  static void clearLogFile() {
    clearLogFileImpl();
  }

  @override
  void trace(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) logger.t(message, error: error, stackTrace: stackTrace);
  }

  @override
  void debug(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) logger.d(message, error: error, stackTrace: stackTrace);
  }

  @override
  void info(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    logger.i(message, error: error, stackTrace: stackTrace);
  }

  @override
  void warn(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    logger.w(message, error: error, stackTrace: stackTrace);
  }

  @override
  void error(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    logger.e(message, error: error, stackTrace: stackTrace);
  }

  @override
  void dangerFailure(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    logger.f(message, error: error, stackTrace: stackTrace);
  }
}
