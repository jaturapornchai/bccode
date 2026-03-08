import 'package:logger/logger.dart';

/// Web platform: ใช้แค่ console output (ไม่มี file logging)
LogOutput createLogOutput() => ConsoleOutput();

/// Web platform: ไม่มี log file ให้ล้าง
void clearLogFileImpl() {}

/// Web platform: ไม่มี project root (file:/// URI ไม่ทำงานบน browser)
String getProjectRoot() => '';
