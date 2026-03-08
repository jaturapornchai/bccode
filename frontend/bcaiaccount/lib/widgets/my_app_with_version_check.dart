import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/services/version_check_service.dart';
import 'package:smlaicloud/widgets/update_available_dialog.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Wrapper widget that adds version checking capability to the app
/// Only active on web platform (kIsWeb)
///
/// Usage:
/// ```dart
/// void main() {
///   runApp(const MyAppWithVersionCheck(
///     child: MyApp(),
///   ));
/// }
/// ```
class MyAppWithVersionCheck extends StatefulWidget {
  final Widget child;

  const MyAppWithVersionCheck({super.key, required this.child});

  @override
  State<MyAppWithVersionCheck> createState() => _MyAppWithVersionCheckState();
}

class _MyAppWithVersionCheckState extends State<MyAppWithVersionCheck> {
  VersionCheckService? _versionCheckService;

  @override
  void initState() {
    super.initState();
    // Only enable version checking on web platform
    if (kIsWeb) {
      _startVersionCheck();
    }
  }

  void _startVersionCheck() {
    _versionCheckService = VersionCheckService(
      onNewVersionAvailable: _showUpdateDialog,
      checkInterval: const Duration(minutes: 5),
    );
    _versionCheckService!.start();

    if (kDebugMode) {
      AppLogger.info('[MyApp] Version check service started');
    }
  }

  Future<void> _showUpdateDialog() async {
    // Get latest version info
    final versionInfo = await _versionCheckService?.getCurrentVersionInfo();

    // Check if widget is still mounted before showing dialog
    if (!mounted) return;

    // Show non-blocking update notification
    showDialog(
      context: context,
      barrierDismissible: true, // Allow dismissing by tapping outside
      builder: (context) => UpdateAvailableDialog(versionInfo: versionInfo),
    );
  }

  @override
  void dispose() {
    _versionCheckService?.stop();
    if (kDebugMode) {
      AppLogger.info('[MyApp] Version check service stopped');
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return widget.child;
  }
}
