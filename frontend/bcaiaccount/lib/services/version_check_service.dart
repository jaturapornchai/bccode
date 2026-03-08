import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Service that checks for new application versions
/// by polling build-info.json every 5 minutes
class VersionCheckService {
  Timer? _timer;
  int? _currentBuildTimestamp;
  final Function()? onNewVersionAvailable;
  final Duration checkInterval;
  static const String buildInfoUrl = '/build-info.json';

  VersionCheckService({
    this.onNewVersionAvailable,
    this.checkInterval = const Duration(minutes: 5),
  });

  /// Start the version checking service
  Future<void> start() async {
    // Load current version first
    await _loadCurrentVersion();

    // Start periodic checks
    _timer = Timer.periodic(checkInterval, (_) async {
      await _checkForUpdates();
    });

    if (kDebugMode) {
      AppLogger.debug(
        '[VersionCheck] Service started - checking every ${checkInterval.inMinutes} minutes',
      );
    }
  }

  /// Stop the version checking service
  void stop() {
    _timer?.cancel();
    _timer = null;
    if (kDebugMode) {
      AppLogger.info('[VersionCheck] Service stopped');
    }
  }

  /// Load the current build timestamp
  Future<void> _loadCurrentVersion() async {
    try {
      final response = await http.get(
        Uri.parse('$buildInfoUrl?t=${DateTime.now().millisecondsSinceEpoch}'),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        _currentBuildTimestamp = data['buildTimestamp'] as int?;
        if (kDebugMode) {
          AppLogger.debug(
            '[VersionCheck] Current version loaded: $_currentBuildTimestamp',
          );
          AppLogger.debug('[VersionCheck] Version: ${data['version']}');
          AppLogger.debug('[VersionCheck] Build date: ${data['buildDate']}');
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('[VersionCheck] Error loading current version: $e');
      }
    }
  }

  /// Check if a new version is available
  Future<void> _checkForUpdates() async {
    if (_currentBuildTimestamp == null) {
      if (kDebugMode) {
        AppLogger.debug('[VersionCheck] Current version unknown, skipping check');
      }
      return;
    }

    try {
      // Add cache-busting parameter
      final response = await http.get(
        Uri.parse('$buildInfoUrl?t=${DateTime.now().millisecondsSinceEpoch}'),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final latestBuildTimestamp = data['buildTimestamp'] as int?;

        if (latestBuildTimestamp != null &&
            latestBuildTimestamp > _currentBuildTimestamp!) {
          if (kDebugMode) {
            AppLogger.debug('[VersionCheck] New version detected!');
            AppLogger.debug('[VersionCheck] Current: $_currentBuildTimestamp');
            AppLogger.debug('[VersionCheck] Latest: $latestBuildTimestamp');
            AppLogger.debug('[VersionCheck] New version: ${data['version']}');
          }

          // Notify about new version
          onNewVersionAvailable?.call();
        } else {
          if (kDebugMode) {
            AppLogger.debug('[VersionCheck] No new version available');
          }
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('[VersionCheck] Error checking for updates: $e');
      }
    }
  }

  /// Get the current version info as a map
  Future<Map<String, dynamic>?> getCurrentVersionInfo() async {
    try {
      final response = await http.get(
        Uri.parse('$buildInfoUrl?t=${DateTime.now().millisecondsSinceEpoch}'),
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body) as Map<String, dynamic>;
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('[VersionCheck] Error getting version info: $e');
      }
    }
    return null;
  }
}
