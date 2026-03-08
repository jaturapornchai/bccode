import 'dart:async';
import 'dart:math';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:qr_flutter/qr_flutter.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

/// LINE Link Result Data
class LineLinkResult {
  final String userId;
  final String displayName;
  final String? pictureUrl;
  final String linkCode;

  LineLinkResult({
    required this.userId,
    required this.displayName,
    this.pictureUrl,
    required this.linkCode,
  });
}

/// LINE Link QR Code Dialog
/// แสดง QR Code สำหรับเชื่อมต่อบัญชี LINE
class LineLinkQRDialog extends StatefulWidget {
  final String employeeCode;
  final String employeeName;
  final String? shopId;
  final String liffBaseUrl;
  final String liffId;
  final String? apiBaseUrl;
  final void Function(LineLinkResult result)? onLinked;

  const LineLinkQRDialog({
    super.key,
    required this.employeeCode,
    required this.employeeName,
    this.shopId,
    required this.liffBaseUrl,
    required this.liffId,
    this.apiBaseUrl,
    this.onLinked,
  });

  @override
  State<LineLinkQRDialog> createState() => _LineLinkQRDialogState();
}

class _LineLinkQRDialogState extends State<LineLinkQRDialog> {
  late String _linkCode;
  late String _liffUrl;
  Timer? _pollingTimer;
  bool _isLinked = false;
  bool _isPolling = false;
  String? _linkedUserName;
  DateTime? _expiresAt;

  @override
  void initState() {
    super.initState();
    _generateLinkCode();
    _startPolling();
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    super.dispose();
  }

  void _generateLinkCode() {
    // Generate random 6-character code
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
    final random = Random();
    _linkCode = List.generate(6, (_) => chars[random.nextInt(chars.length)]).join();

    // Build LIFF URL with employee_code
    final shopId = widget.shopId ?? 'default';
    final employeeCode = Uri.encodeComponent(widget.employeeCode);
    _liffUrl = 'https://liff.line.me/${widget.liffId}/link?code=$_linkCode&shop_id=$shopId&employee_code=$employeeCode';

    // Set expiry time (30 minutes)
    _expiresAt = DateTime.now().add(const Duration(minutes: 30));
  }

  void _startPolling() {
    _pollingTimer = Timer.periodic(const Duration(seconds: 3), (_) async {
      if (_isLinked) return;
      await _checkLinkStatus();
    });
  }

  Future<void> _checkLinkStatus() async {
    if (_isPolling) return;

    setState(() => _isPolling = true);

    try {
      final apiUrl = widget.apiBaseUrl ?? widget.liffBaseUrl;
      final response = await http.get(
        Uri.parse('$apiUrl/api/link?code=$_linkCode'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true && data['linked'] == true) {
          final linkedData = data['data'];
          setState(() {
            _isLinked = true;
            _linkedUserName = linkedData?['displayName'] ?? 'Unknown';
          });
          _pollingTimer?.cancel();

          // Call callback with linked data
          if (widget.onLinked != null && linkedData != null) {
            widget.onLinked!(LineLinkResult(
              userId: linkedData['userId'] ?? '',
              displayName: linkedData['displayName'] ?? '',
              pictureUrl: linkedData['pictureUrl'],
              linkCode: _linkCode,
            ));
          }
        }
      }
    } catch (e) {
      // Ignore polling errors
      debugPrint('Polling error: $e');
    } finally {
      if (mounted) {
        setState(() => _isPolling = false);
      }
    }
  }

  void _regenerateCode() {
    setState(() {
      _isLinked = false;
      _linkedUserName = null;
      _generateLinkCode();
    });
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: const Color(0xFF06C755).withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(
              Icons.qr_code,
              color: Color(0xFF06C755),
            ),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language('connect_line'),
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                Text(
                  widget.employeeName,
                  style: TextStyle(fontSize: 14, color: Colors.grey[600]),
                ),
              ],
            ),
          ),
        ],
      ),
      content: SizedBox(
        width: 350,
        child: _isLinked ? _buildSuccessContent() : _buildQRContent(),
      ),
      actions: _isLinked
          ? [
              ElevatedButton(
                onPressed: () => Navigator.pop(context),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF06C755),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                ),
                child: Text(global.language('close')),
              ),
            ]
          : [
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: Text(global.language('cancel')),
              ),
              TextButton(
                onPressed: _regenerateCode,
                child: Text(global.language('create_new_code')),
              ),
            ],
    );
  }

  Widget _buildSuccessContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.green[50],
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: const BoxDecoration(
                  color: Color(0xFF06C755),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.check,
                  color: Colors.white,
                  size: 48,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                global.language("link_success"),
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF06C755),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                _linkedUserName ?? '',
                style: const TextStyle(fontSize: 16),
              ),
              const SizedBox(height: 4),
              Text(
                'รหัส: ${widget.employeeCode}',
                style: TextStyle(fontSize: 14, color: Colors.grey[600]),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildQRContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // QR Code
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.grey[300]!),
          ),
          child: QrImageView(
            data: _liffUrl,
            version: QrVersions.auto,
            size: 200,
            backgroundColor: Colors.white,
            errorCorrectionLevel: QrErrorCorrectLevel.M,
          ),
        ),
        const SizedBox(height: 16),

        // Link Code Display
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          decoration: BoxDecoration(
            color: Colors.amber[50],
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.amber[200]!),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                global.language("code_label"),
                style: const TextStyle(fontSize: 14),
              ),
              Text(
                _linkCode,
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 4,
                  fontFamily: 'monospace',
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),

        // Instructions
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.grey[100],
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 24,
                    height: 24,
                    decoration: BoxDecoration(
                      color: const Color(0xFF06C755),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Center(
                      child: Text('1', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(global.language("scan_qr_with_line")),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Container(
                    width: 24,
                    height: 24,
                    decoration: BoxDecoration(
                      color: const Color(0xFF06C755),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Center(
                      child: Text('2', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(global.language("add_line_oa_friend")),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Container(
                    width: 24,
                    height: 24,
                    decoration: BoxDecoration(
                      color: const Color(0xFF06C755),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Center(
                      child: Text('3', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(global.language("confirm_link")),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),

        // Copy Link Button
        OutlinedButton.icon(
          onPressed: () {
            Clipboard.setData(ClipboardData(text: _liffUrl));
            global.showInfoSnackBar(context, global.language('link_copied'));
          },
          icon: Icon(Icons.copy, size: 18),
          label: Text(global.language('copy_link')),
        ),

        // Polling indicator - always show while waiting
        Padding(
          padding: const EdgeInsets.only(top: 12),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: Colors.grey[400],
                ),
              ),
              const SizedBox(width: 8),
              Text(
                global.language("waiting_for_link"),
                style: TextStyle(fontSize: 12, color: Colors.grey[600]),
              ),
            ],
          ),
        ),

        // Expiry info
        if (_expiresAt != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              global.language("code_expires_in_30_min"),
              style: TextStyle(fontSize: 11, color: Colors.grey[500]),
            ),
          ),
      ],
    );
  }
}

/// Show LINE Link QR Dialog
Future<void> showLineLinkQRDialog({
  required BuildContext context,
  required String employeeCode,
  required String employeeName,
  String? shopId,
  required String liffId,
  String liffBaseUrl = 'https://dev-api.bcaicloud.com/liff',
  String? apiBaseUrl,
  void Function(LineLinkResult result)? onLinked,
}) {
  return showDialog(
    context: context,
    barrierDismissible: false,
    builder: (context) => LineLinkQRDialog(
      employeeCode: employeeCode,
      employeeName: employeeName,
      shopId: shopId,
      liffId: liffId,
      liffBaseUrl: liffBaseUrl,
      apiBaseUrl: apiBaseUrl,
      onLinked: onLinked,
    ),
  );
}
