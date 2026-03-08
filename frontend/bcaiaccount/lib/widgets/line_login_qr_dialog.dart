import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:qr_flutter/qr_flutter.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

/// LINE Login Result Data
class LineLoginResult {
  final String userId;
  final String displayName;
  final String? pictureUrl;
  final String? email;

  LineLoginResult({
    required this.userId,
    required this.displayName,
    this.pictureUrl,
    this.email,
  });
}

/// LINE Login QR Code Dialog
/// แสดง QR Code สำหรับเข้าสู่ระบบด้วย LINE
class LineLoginQRDialog extends StatefulWidget {
  final String liffId;
  final String liffBaseUrl;
  final void Function(LineLoginResult result)? onLogin;

  const LineLoginQRDialog({
    super.key,
    required this.liffId,
    this.liffBaseUrl = 'https://dev-api.bcaicloud.com/liff',
    this.onLogin,
  });

  @override
  State<LineLoginQRDialog> createState() => _LineLoginQRDialogState();
}

class _LineLoginQRDialogState extends State<LineLoginQRDialog> {
  String _loginCode = '';
  String _liffUrl = '';
  Timer? _pollingTimer;
  Timer? _countdownTimer;
  bool _isConfirmed = false;
  bool _isPolling = false;
  String? _confirmedUserName;
  DateTime? _expiresAt;
  bool _isExpired = false;
  bool _isLoading = true;
  String? _errorMessage;
  int _countdownSeconds = 5;

  @override
  void initState() {
    super.initState();
    _createLoginCode();
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    _countdownTimer?.cancel();
    super.dispose();
  }

  /// สร้าง login code โดยเรียก API เพื่อบันทึกลง MongoDB
  Future<void> _createLoginCode() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // เรียก API สร้าง login code
      final response = await http.get(
        Uri.parse('${widget.liffBaseUrl}/api/login?action=create'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          setState(() {
            _loginCode = data['data']['code'];
            _liffUrl = data['data']['loginUrl'];
            _expiresAt = DateTime.tryParse(data['data']['expiresAt'] ?? '') ??
                DateTime.now().add(const Duration(minutes: 5));
            _isExpired = false;
            _isLoading = false;
          });
          _startPolling();
        } else {
          throw Exception(data['message'] ?? 'Failed to create login code');
        }
      } else {
        throw Exception('Server error: ${response.statusCode}');
      }
    } catch (e) {
      debugPrint('Error creating login code: $e');
      setState(() {
        _isLoading = false;
        _errorMessage = 'ไม่สามารถสร้างรหัสได้: $e';
      });
    }
  }

  void _startPolling() {
    _pollingTimer = Timer.periodic(const Duration(seconds: 2), (_) async {
      if (_isConfirmed) return;

      // Check if expired
      if (_expiresAt != null && DateTime.now().isAfter(_expiresAt!)) {
        setState(() => _isExpired = true);
        _pollingTimer?.cancel();
        return;
      }

      await _checkLoginStatus();
    });
  }

  Future<void> _checkLoginStatus() async {
    if (_isPolling) return;

    setState(() => _isPolling = true);

    try {
      final response = await http.get(
        Uri.parse('${widget.liffBaseUrl}/api/login?code=$_loginCode'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true && data['confirmed'] == true) {
          final confirmedData = data['data'];
          setState(() {
            _isConfirmed = true;
            _confirmedUserName = confirmedData?['displayName'] ?? 'Unknown';
          });
          _pollingTimer?.cancel();

          // Start countdown and call callback after countdown
          _startCountdown(confirmedData);
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
    _pollingTimer?.cancel();
    setState(() {
      _isConfirmed = false;
      _confirmedUserName = null;
      _isExpired = false;
    });
    _createLoginCode();
  }

  void _startCountdown(Map<String, dynamic>? confirmedData) {
    _countdownSeconds = 5;
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }

      setState(() {
        _countdownSeconds--;
      });

      if (_countdownSeconds <= 0) {
        timer.cancel();

        // Call callback with login data
        if (widget.onLogin != null && confirmedData != null) {
          widget.onLogin!(LineLoginResult(
            userId: confirmedData['userId'] ?? '',
            displayName: confirmedData['displayName'] ?? '',
            pictureUrl: confirmedData['pictureUrl'],
            email: confirmedData['email'],
          ));
        }

        // Auto close dialog
        if (mounted) {
          Navigator.pop(context);
        }
      }
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
              Icons.login,
              color: Color(0xFF06C755),
            ),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Text(
              global.language('login_with_line'),
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
      content: SizedBox(
        width: 350,
        child: _isLoading
            ? _buildLoadingContent()
            : _errorMessage != null
                ? _buildErrorContent()
                : _isConfirmed
                    ? _buildSuccessContent()
                    : _isExpired
                        ? _buildExpiredContent()
                        : _buildQRContent(),
      ),
      actions: _isConfirmed
          ? [] // No buttons - auto close after countdown
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

  Widget _buildLoadingContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const SizedBox(height: 40),
        const CircularProgressIndicator(
          color: Color(0xFF06C755),
        ),
        const SizedBox(height: 24),
        Text(
          global.language("creating_code"),
          style: const TextStyle(fontSize: 16),
        ),
        const SizedBox(height: 40),
      ],
    );
  }

  Widget _buildErrorContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.red[50],
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.red[400],
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.error_outline,
                  color: Colors.white,
                  size: 48,
                ),
              ),
              SizedBox(height: 16),
              Text(
                global.language('error_occurred'),
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Colors.red[700],
                ),
              ),
              const SizedBox(height: 8),
              Text(
                _errorMessage ?? global.language("cannot_create_code"),
                style: const TextStyle(fontSize: 14, color: Colors.grey),
                textAlign: TextAlign.center,
              ),
              SizedBox(height: 16),
              ElevatedButton(
                onPressed: _regenerateCode,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF06C755),
                  foregroundColor: Colors.white,
                ),
                child: Text(global.language('export_report_retry')),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildExpiredContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.orange[50],
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.orange[400],
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.timer_off,
                  color: Colors.white,
                  size: 48,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                global.language("code_expired"),
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Colors.orange[700],
                ),
              ),
              const SizedBox(height: 8),
              Text(
                global.language("please_create_new_code"),
                style: const TextStyle(fontSize: 14, color: Colors.grey),
              ),
            ],
          ),
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
              SizedBox(height: 16),
              Text(
                global.language('login_success'),
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF06C755),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                _confirmedUserName ?? '',
                style: const TextStyle(fontSize: 16),
              ),
              const SizedBox(height: 12),
              // Countdown indicator
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: const Color(0xFF06C755).withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        value: _countdownSeconds / 5,
                        color: const Color(0xFF06C755),
                        backgroundColor: Colors.grey[300],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'กำลังเข้าสู่ระบบใน $_countdownSeconds วินาที...',
                      style: TextStyle(fontSize: 14, color: Colors.grey[700]),
                    ),
                  ],
                ),
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
            border: Border.all(color: const Color(0xFF06C755), width: 2),
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

        // Login Code Display
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          decoration: BoxDecoration(
            color: const Color(0xFF06C755).withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: const Color(0xFF06C755).withValues(alpha: 0.3)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                global.language("code_label"),
                style: const TextStyle(fontSize: 14),
              ),
              Text(
                _loginCode,
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 4,
                  fontFamily: 'monospace',
                  color: Color(0xFF06C755),
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
              SizedBox(height: 8),
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
                  SizedBox(width: 8),
                  Text(global.language('confirm_login')),
                ],
              ),
              SizedBox(height: 8),
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
                  SizedBox(width: 8),
                  Text(global.language('auto_login')),
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

        // Polling indicator
        Padding(
          padding: EdgeInsets.only(top: 12),
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
              SizedBox(width: 8),
              Text(
                global.language('waiting_confirmation'),
                style: TextStyle(fontSize: 12, color: Colors.grey[600]),
              ),
            ],
          ),
        ),

        // Expiry info
        if (_expiresAt != null)
          Padding(
            padding: EdgeInsets.only(top: 8),
            child: Text(
              global.language('code_expires_in_5_min'),
              style: TextStyle(fontSize: 11, color: Colors.grey[500]),
            ),
          ),
      ],
    );
  }
}

/// Show LINE Login QR Dialog
Future<LineLoginResult?> showLineLoginQRDialog({
  required BuildContext context,
  required String liffId,
  String liffBaseUrl = 'https://dev-api.bcaicloud.com/liff',
}) async {
  LineLoginResult? result;

  await showDialog(
    context: context,
    barrierDismissible: false,
    builder: (context) => LineLoginQRDialog(
      liffId: liffId,
      liffBaseUrl: liffBaseUrl,
      onLogin: (loginResult) {
        result = loginResult;
      },
    ),
  );

  return result;
}
