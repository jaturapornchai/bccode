import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
// ignore: avoid_web_libraries_in_flutter
import 'dart:html' as html show window;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Dialog shown when a new application version is available
class UpdateAvailableDialog extends StatelessWidget {
  final Map<String, dynamic>? versionInfo;

  const UpdateAvailableDialog({super.key, this.versionInfo});

  @override
  Widget build(BuildContext context) {
    final version = versionInfo?['version'] ?? 'Unknown';
    final buildDate = versionInfo?['buildDate'] ?? 'Unknown';
    final gitBranch = versionInfo?['gitBranch'] ?? '';

    return AlertDialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      title: Row(
        children: [
          Icon(
            Icons.system_update,
            color: Theme.of(context).colorScheme.primary,
            size: 28,
          ),
          const SizedBox(width: 12),
          const Text('Update Available'),
        ],
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'A new version of the application is available.',
            style: TextStyle(fontSize: 16),
          ),
          const SizedBox(height: 16),
          _buildInfoRow('Version', version),
          _buildInfoRow('Build Date', buildDate),
          if (gitBranch.isNotEmpty) _buildInfoRow('Branch', gitBranch),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.blue.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.blue.withOpacity(0.3)),
            ),
            child: Row(
              children: [
                Icon(Icons.info_outline, color: Colors.blue.shade700, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Please refresh to get the latest features and fixes.',
                    style: TextStyle(fontSize: 13, color: Colors.blue.shade700),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Later'),
        ),
        ElevatedButton.icon(
          onPressed: () {
            Navigator.of(context).pop();
            _refreshPage();
          },
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh Now'),
          style: ElevatedButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
          ),
        ),
      ],
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 90,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
          ),
          Expanded(child: Text(value, style: const TextStyle(fontSize: 14))),
        ],
      ),
    );
  }

  void _refreshPage() {
    if (kIsWeb) {
      // For web, reload the page
      html.window.location.reload();
    } else {
      if (kDebugMode) {
        AppLogger.debug('[UpdateDialog] Refresh not implemented for non-web platforms');
      }
    }
  }
}
