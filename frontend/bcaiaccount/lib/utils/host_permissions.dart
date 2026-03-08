import 'dart:io' show Platform;
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

bool _canReadPlatform() {
  if (kIsWeb) {
    return false;
  }
  try {
    // Access any Platform getter to ensure availability
    // ignore: unused_local_variable
    final _ = Platform.operatingSystem;
    return true;
  } catch (e) {
    if (kDebugMode) {
      AppLogger.debug('Platform information unavailable: $e');
    }
    return false;
  }
}

bool isWindowsHost(String expectedHost) {
  if (!_canReadPlatform()) {
    return false;
  }
  try {
    if (!Platform.isWindows) {
      return false;
    }
    final host = Platform.localHostname.toLowerCase();
    return host == expectedHost.toLowerCase();
  } catch (e) {
    if (kDebugMode) {
      AppLogger.debug('Unable to read local hostname: $e');
    }
    return false;
  }
}

bool allowManualCostTestHost() {
  return isWindowsHost('jead-i9-pc') || isWindowsHost('jead-i9-ultra');
}
