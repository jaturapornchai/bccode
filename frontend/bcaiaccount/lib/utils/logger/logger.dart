import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';

abstract class Log {
  void trace(dynamic message, {dynamic error, StackTrace? stackTrace});
  void debug(dynamic message, {dynamic error, StackTrace? stackTrace});
  void info(dynamic message, {dynamic error, StackTrace? stackTrace});
  void warn(dynamic message, {dynamic error, StackTrace? stackTrace});
  void error(dynamic message, {dynamic error, StackTrace? stackTrace});
  void dangerFailure(dynamic message, {dynamic error, StackTrace? stackTrace});
}

/// Custom Logger Printer ที่แสดง filename:line
class CustomLogPrinter extends LogPrinter {
  final String? className;
  final bool printTime;
  final bool printEmojis;
  final bool colors;

  CustomLogPrinter({
    this.className,
    this.printTime = true,
    this.printEmojis = true,
    this.colors = false,
  });

  @override
  List<String> log(LogEvent event) {
    final emoji = PrettyPrinter.defaultLevelEmojis[event.level];
    final message = event.message;

    // ดึง StackTrace เพื่อหาชื่อไฟล์และบรรทัด
    String fileInfo = _getFileInfo();

    // สร้าง timestamp
    String timeStr = '';
    if (printTime) {
      final now = DateTime.now();
      timeStr =
          '${now.hour.toString().padLeft(2, '0')}:${now.minute.toString().padLeft(2, '0')}:${now.second.toString().padLeft(2, '0')}.${now.millisecond.toString().padLeft(3, '0')}';
    }

    // สร้าง level emoji/text
    String levelStr = '';
    if (printEmojis && emoji != null) {
      levelStr = '$emoji ';
    }
    levelStr += event.level
        .toString()
        .split('.')
        .last
        .toUpperCase()
        .padRight(5);

    // Format: [HH:MM:SS.mmm] [LEVEL] (file:line:column) message
    // ใช้ format (path:line:column) ให้ VSCode รู้จักและคลิกได้
    final output = StringBuffer();

    if (printTime) {
      output.write('[$timeStr] ');
    }

    output.write('[$levelStr] ');
    output.write('($fileInfo) '); // ใช้ (path:line:column) format
    output.write(message);

    // เพิ่ม error ถ้ามี
    if (event.error != null) {
      output.write('\n    Error: ${event.error}');
    }

    // เพิ่ม stackTrace ถ้ามี
    if (event.stackTrace != null) {
      output.write('\n${event.stackTrace}');
    }

    return [output.toString()];
  }

  /// ดึงชื่อไฟล์และบรรทัดจาก StackTrace
  /// Return full path ในรูปแบบที่ VSCode Debug Console คลิกได้
  /// Format: file:///c:/full/path/to/file.dart:line:column
  String _getFileInfo() {
    try {
      final stackTrace = StackTrace.current.toString();
      final lines = stackTrace.split('\n');

      // หาบรรทัดที่ไม่ใช่ logger.dart, app_logger.dart
      for (var line in lines) {
        if (line.contains('package:smlaicloud/') &&
            !line.contains('logger.dart') &&
            !line.contains('app_logger.dart') &&
            !line.contains('/utils/logger/')) {
          // Extract file path and line number
          // Regex จับ path หลัง package:smlaicloud/
          final match = RegExp(
            r'package:smlaicloud/(.+\.dart):(\d+):(\d+)',
          ).firstMatch(line);
          if (match != null) {
            var filePath = match.group(
              1,
            ); // e.g., "bloc/shop/list_shop_bloc.dart"
            final lineNumber = match.group(2);
            final columnNumber = match.group(3);

            // Convert to absolute Windows path with file:// URI scheme
            // จาก "bloc/shop/list_shop_bloc.dart" -> "file:///c:/bcdev/bcaiaccount/lib/bloc/shop/list_shop_bloc.dart"
            // VSCode Debug Console รู้จัก file:// URI และคลิกเปิดไฟล์ได้
            final absolutePath = 'file:///c:/bcdev/bcaiaccount/lib/$filePath';

            // Return format: file:///path:line:column
            // VSCode จะรู้จัก format นี้และทำให้คลิกได้เปิดไฟล์และไปบรรทัดนั้นเลย
            return '$absolutePath:$lineNumber:$columnNumber';
          }
        }
      }

      // ถ้าหาไม่เจอ ให้แสดง className (ถ้ามี)
      if (className != null) {
        return className!;
      }

      return 'unknown';
    } catch (e) {
      return 'unknown';
    }
  }
}

/// Dev File Log Output - เขียน log ลง console + file สำหรับ dev mode
/// ไฟล์อยู่ที่ logs/flutter_debug.log ใน working directory
/// ให้ AI (Claude Code) อ่าน log ได้ง่าย
class _DevFileLogOutput extends LogOutput {
  static const String _logDir = 'logs';
  static const String _logFileName = 'flutter_debug.log';
  static IOSink? _fileSink;
  static bool _initialized = false;

  @override
  Future<void> init() async {
    if (!_initialized) _initFile();
  }

  static void _initFile() {
    if (_initialized) return;
    try {
      final dir = Directory(_logDir);
      if (!dir.existsSync()) {
        dir.createSync(recursive: true);
      }
      final file = File('$_logDir/$_logFileName');
      _fileSink = file.openWrite(mode: FileMode.append);
      _initialized = true;
    } catch (_) {
      // ถ้าเปิดไฟล์ไม่ได้ ก็ใช้แค่ console
    }
  }

  @override
  void output(OutputEvent event) {
    for (var line in event.lines) {
      // Console output (เหมือน ConsoleOutput เดิม)
      // ignore: avoid_print
      print(line);

      // File output with error handling
      if (_fileSink != null && _initialized) {
        try {
          _fileSink!.writeln(line);
        } catch (_) {
          // Ignore errors (sink might be closed during clearFile)
        }
      }
    }

    // Flush after all lines (with error handling)
    if (_fileSink != null && _initialized) {
      try {
        _fileSink!.flush();
      } catch (_) {
        // Ignore flush errors
      }
    }
  }

  /// ล้าง log file (เรียกตอน hot reload/restart)
  static void clearFile() {
    try {
      // ปิด sink ปัจจุบัน (ถ้ามี)
      if (_fileSink != null) {
        try {
          _fileSink!.flush();
        } catch (_) {
          // Ignore flush errors
        }
        try {
          _fileSink!.close();
        } catch (_) {
          // Ignore close errors
        }
      }
      _fileSink = null;
      _initialized = false;

      // เขียนข้อความล้าง log
      final file = File('$_logDir/$_logFileName');
      if (file.existsSync()) {
        try {
          file.writeAsStringSync(
            '--- Log cleared: ${DateTime.now()} ---\n',
          );
        } catch (_) {
          // Ignore write errors
        }
      }

      // เปิด sink ใหม่
      _initFile();
    } catch (_) {
      // Ignore any other errors
    }
  }

  @override
  Future<void> destroy() async {
    if (_fileSink != null) {
      try {
        await _fileSink!.flush();
      } catch (_) {
        // Ignore flush errors
      }
      try {
        await _fileSink!.close();
      } catch (_) {
        // Ignore close errors
      }
      _fileSink = null;
    }
    _initialized = false;
  }
}

class LogImpl implements Log {
  late final Logger logger;

  LogImpl() {
    // Dev mode (non-web): เขียน log ทั้ง console และ file
    LogOutput output;
    // TEMPORARY: ปิด file logging ชั่วคราวเพราะ StreamSink error
    // if (kDebugMode && !kIsWeb) {
    //   output = _DevFileLogOutput();
    // } else {
    //   output = ConsoleOutput();
    // }
    output = ConsoleOutput(); // ใช้แค่ console ก่อน

    logger = Logger(
      printer: CustomLogPrinter(
        printTime: true,
        printEmojis: true,
        colors: false,
      ),
      output: output,
      level: kDebugMode ? Level.trace : Level.info,
    );
  }

  /// ล้าง log file สำหรับ dev mode
  static void clearLogFile() {
    if (kDebugMode && !kIsWeb) {
      _DevFileLogOutput.clearFile();
    }
  }

  @override
  void trace(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      logger.t(message, error: error, stackTrace: stackTrace);
    }
  }

  @override
  void debug(dynamic message, {dynamic error, StackTrace? stackTrace}) {
    if (kDebugMode) {
      logger.d(message, error: error, stackTrace: stackTrace);
    }
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
