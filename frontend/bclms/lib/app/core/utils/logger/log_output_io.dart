import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';

/// Native platform: ใช้ console + file output
LogOutput createLogOutput() {
  if (kDebugMode) {
    return DevFileLogOutput();
  }
  return ConsoleOutput();
}

/// Native platform: ล้าง log file
void clearLogFileImpl() {
  if (kDebugMode) {
    DevFileLogOutput.clearFile();
  }
}

/// Native platform: ดึง project root จาก Directory.current
String getProjectRoot() {
  try {
    final path = Directory.current.path.replaceAll('\\', '/');
    // Windows: แปลง drive letter เป็น lowercase สำหรับ file:/// URI
    if (path.length >= 2 && path[1] == ':') {
      return '${path[0].toLowerCase()}${path.substring(1)}';
    }
    return path;
  } catch (_) {
    return '';
  }
}

/// เขียน log ลง console + file สำหรับ dev mode
/// ไฟล์อยู่ที่ logs/flutter_debug.log ใน working directory
class DevFileLogOutput extends LogOutput {
  static const _logDir = 'logs';
  static const _logFileName = 'flutter_debug.log';
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
    } catch (_) {}
  }

  @override
  void output(OutputEvent event) {
    for (var line in event.lines) {
      // Console output
      // ignore: avoid_print
      print(line);

      // File output
      if (_fileSink != null && _initialized) {
        try {
          _fileSink!.writeln(line);
        } catch (_) {}
      }
    }

    // Flush หลังเขียนเสร็จ
    if (_fileSink != null && _initialized) {
      try {
        _fileSink!.flush();
      } catch (_) {}
    }
  }

  /// ล้าง log file (เรียกตอน hot reload/restart)
  static void clearFile() {
    try {
      if (_fileSink != null) {
        try { _fileSink!.flush(); } catch (_) {}
        try { _fileSink!.close(); } catch (_) {}
      }
      _fileSink = null;
      _initialized = false;

      final file = File('$_logDir/$_logFileName');
      if (file.existsSync()) {
        try {
          file.writeAsStringSync('--- Log cleared: ${DateTime.now()} ---\n');
        } catch (_) {}
      }

      _initFile();
    } catch (_) {}
  }

  @override
  Future<void> destroy() async {
    if (_fileSink != null) {
      try { await _fileSink!.flush(); } catch (_) {}
      try { await _fileSink!.close(); } catch (_) {}
      _fileSink = null;
    }
    _initialized = false;
  }
}
