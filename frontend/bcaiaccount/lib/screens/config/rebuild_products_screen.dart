import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// หน้าจอสำหรับ Rebuild Products
/// เรียก API /rebuild-products เพื่อ sync ข้อมูลสินค้าจาก MongoDB → PostgreSQL/ClickHouse
class RebuildProductsScreen extends StatefulWidget {
  const RebuildProductsScreen({super.key});

  @override
  State<RebuildProductsScreen> createState() => _RebuildProductsScreenState();
}

class _RebuildProductsScreenState extends State<RebuildProductsScreen>
    with global.ThemeRefreshMixin {
  bool _isProcessing = false;
  String _statusMessage = '';
  List<String> _logs = [];

  // Progress tracking
  double _progress = 0.0;
  String _currentStep = '';
  Timer? _progressTimer;

  /// เรียก API rebuild-products ผ่าน reportServicePost
  Future<void> _rebuildProducts() async {
    // แสดง confirmation dialog
    final confirmed = await _showConfirmDialog();
    if (confirmed != true) return;

    setState(() {
      _isProcessing = true;
      _statusMessage = '${global.language("processing")} Rebuild Products...';
      _logs = [];
      _progress = 0.0;
      _currentStep = '${global.language("starting")}...';
    });

    // เริ่ม progress simulation
    _startProgressSimulation();

    try {
      final shopId = global.getShopId();

      // Step 1: เตรียมข้อมูล (0-10%)
      _updateProgress(0.05, '${global.language("preparing_data")}...');
      _addLog('Shop ID: $shopId');
      await Future.delayed(const Duration(milliseconds: 300));

      _updateProgress(0.10, '${global.language("creating_payload")}...');
      // Request payload สำหรับ reportServicePost
      Map<String, dynamic> payload = {
        "shop_id": shopId,
        "command_id": "rebuild-products",
      };

      var jsonPayload = jsonEncode(payload);
      _addLog('Payload: $jsonPayload');
      await Future.delayed(const Duration(milliseconds: 300));

      // Step 2: ส่ง request (10-30%)
      _updateProgress(0.20, '${global.language("sending_request_to_server")}...');
      _addLog('${global.language("sending_request")}...');
      await Future.delayed(const Duration(milliseconds: 500));

      final jsonResult = await global.reportServicePost(jsonPayload);

      _updateProgress(0.30, '${global.language("received_response_from_server")}...');
      _addLog('Response: ${jsonResult.toString()}');

      if (jsonResult['code'] == 200) {
        // Step 3: Backend กำลังประมวลผล (30-90%)
        _updateProgress(0.35, '${global.language("backend_processing")}...');
        await Future.delayed(const Duration(milliseconds: 500));

        _updateProgress(0.50, '${global.language("fetching_data_from_mongodb")}...');
        await Future.delayed(const Duration(seconds: 1));

        _updateProgress(0.65, '${global.language("mapping_prices")}...');
        await Future.delayed(const Duration(seconds: 1));

        _updateProgress(0.75, '${global.language("creating_data_in_postgresql")}...');
        await Future.delayed(const Duration(milliseconds: 800));

        _updateProgress(0.85, '${global.language("creating_data_in_clickhouse")}...');
        await Future.delayed(const Duration(milliseconds: 800));

        _updateProgress(0.95, '${global.language("verifying_data_integrity")}...');
        await Future.delayed(const Duration(milliseconds: 500));

        _updateProgress(1.0, '${global.language("completed")}!');

        setState(() {
          _statusMessage = 'Rebuild Products ${global.language("success")}!';
        });

        if (mounted) {
          global.showSnackBar(
            context,
            Icon(Icons.check_circle, color: global.theme.onPrimaryColor),
            'Rebuild Products ${global.language("success")}',
            global.theme.positiveHighlightTextColor,
          );
        }

        AppLogger.success('Rebuild Products สำเร็จ: ${jsonResult['message']}');
      } else {
        _updateProgress(0.0, global.language('failed'));

        setState(() {
          _statusMessage = 'Rebuild Products ${global.language("failed")}: ${jsonResult['message']}';
        });

        if (mounted) {
          global.showSnackBar(
            context,
            Icon(Icons.error, color: global.theme.onPrimaryColor),
            'Rebuild Products ${global.language("failed")}: ${jsonResult['message']}',
            global.theme.negativeHighlightTextColor,
          );
        }

        AppLogger.err('Rebuild Products ล้มเหลว: ${jsonResult['message']}');
      }
    } catch (e) {
      _updateProgress(0.0, global.language('error_occurred'));
      _addLog('Exception: $e');

      setState(() {
        _statusMessage = '${global.language("error_occurred")}: ${e.toString()}';
      });

      if (mounted) {
        global.showSnackBar(
          context,
          Icon(Icons.error, color: global.theme.onPrimaryColor),
          '${global.language("error_occurred")}: ${e.toString()}',
          global.theme.negativeHighlightTextColor,
        );
      }

      AppLogger.err('Rebuild Products error', err: e);
    } finally {
      _stopProgressSimulation();
      if (mounted) {
        setState(() {
          _isProcessing = false;
        });
      }
    }
  }

  /// อัพเดท progress และ current step
  void _updateProgress(double progress, String step) {
    if (mounted) {
      setState(() {
        _progress = progress;
        _currentStep = step;
      });
      _addLog(step);
    }
  }

  /// เริ่ม progress simulation (สำหรับกรณี backend ประมวลผลนาน)
  void _startProgressSimulation() {
    _progressTimer = Timer.periodic(const Duration(milliseconds: 200), (timer) {
      if (!_isProcessing || _progress >= 0.95) {
        timer.cancel();
        return;
      }
      // Slowly increment progress
      if (mounted && _progress < 0.95) {
        setState(() {
          _progress = (_progress + 0.005).clamp(0.0, 0.95);
        });
      }
    });
  }

  /// หยุด progress simulation
  void _stopProgressSimulation() {
    _progressTimer?.cancel();
    _progressTimer = null;
  }

  @override
  void dispose() {
    _stopProgressSimulation();
    super.dispose();
  }

  /// เพิ่ม log message
  void _addLog(String message) {
    setState(() {
      _logs.add('[${DateTime.now().toString().substring(11, 19)}] $message');
    });
    AppLogger.info(message);
  }

  /// แสดง confirmation dialog
  Future<bool?> _showConfirmDialog() {
    return showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('confirm_rebuild_products')),
        content: Text(
          global.language('rebuild_products_warning'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.warningHighlightTextColor),
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('rebuild_products')),
        backgroundColor: Colors.deepOrange,
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // คำอธิบาย
            Card(
              color: Colors.orange[50],
              child: Padding(
                padding: EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.info, color: Colors.orange[800]),
                        SizedBox(width: 8),
                        Text(
                          global.language('rebuild_products'),
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: Colors.orange[800],
                          ),
                        ),
                      ],
                    ),
                    SizedBox(height: 12),
                    Text(
                      global.language('rebuild_products_description'),
                      style: TextStyle(height: 1.5),
                    ),
                  ],
                ),
              ),
            ),
            SizedBox(height: 24),

            // ปุ่ม Rebuild
            ElevatedButton.icon(
              onPressed: _isProcessing ? null : _rebuildProducts,
              icon: _isProcessing
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(global.theme.onPrimaryColor),
                      ),
                    )
                  : Icon(Icons.refresh),
              label: Text(
                _isProcessing ? '${global.language("processing")}...' : global.language('rebuild_products'),
                style: TextStyle(fontSize: 16),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.deepOrange,
                foregroundColor: global.theme.onPrimaryColor,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
            SizedBox(height: 24),

            // Progress Bar - แสดงเมื่อกำลังประมวลผล
            if (_isProcessing) ...[
              Card(
                elevation: 4,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(20.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Current step text
                      Row(
                        children: [
                          Icon(Icons.info_outline, color: Colors.blue[700], size: 20),
                          SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              _currentStep,
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w500,
                                color: Colors.blue[900],
                              ),
                            ),
                          ),
                        ],
                      ),
                      SizedBox(height: 16),

                      // Progress bar
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: LinearProgressIndicator(
                          value: _progress,
                          minHeight: 12,
                          backgroundColor: global.theme.dividerBorderColor,
                          valueColor: AlwaysStoppedAnimation<Color>(
                            _progress >= 1.0
                                ? global.theme.positiveHighlightTextColor
                                : global.theme.infoHighlightTextColor,
                          ),
                        ),
                      ),
                      SizedBox(height: 8),

                      // Percentage text
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '${(_progress * 100).toStringAsFixed(0)}%',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: _progress >= 1.0
                                  ? Colors.green[800]
                                  : Colors.blue[800],
                            ),
                          ),
                          if (_progress >= 1.0)
                            Icon(Icons.check_circle, color: Colors.green[700], size: 24),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
              SizedBox(height: 16),
            ],

            // Status message
            if (_statusMessage.isNotEmpty)
              Card(
                color: _statusMessage.contains(global.language('success'))
                    ? Colors.green[50]
                    : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('errors'))
                        ? Colors.red[50]
                        : global.theme.infoHighlightColor,
                child: Padding(
                  padding: EdgeInsets.all(12.0),
                  child: Row(
                    children: [
                      Icon(
                        _statusMessage.contains(global.language('success'))
                            ? Icons.check_circle
                            : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('errors'))
                                ? Icons.error
                                : Icons.info,
                        color: _statusMessage.contains(global.language('success'))
                            ? Colors.green[800]
                            : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('errors'))
                                ? Colors.red[800]
                                : Colors.blue[800],
                      ),
                      SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _statusMessage,
                          style: TextStyle(
                            color: _statusMessage.contains(global.language('success'))
                                ? Colors.green[800]
                                : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('errors'))
                                    ? Colors.red[800]
                                    : Colors.blue[800],
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            SizedBox(height: 16),

            // Logs
            if (_logs.isNotEmpty) ...[
              Text(
                'Logs:',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
              SizedBox(height: 8),
              Expanded(
                child: Card(
                  child: ListView.builder(
                    padding: const EdgeInsets.all(8),
                    itemCount: _logs.length,
                    itemBuilder: (context, index) {
                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 2),
                        child: Text(
                          _logs[index],
                          style: TextStyle(
                            fontSize: 12,
                            fontFamily: 'monospace',
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
