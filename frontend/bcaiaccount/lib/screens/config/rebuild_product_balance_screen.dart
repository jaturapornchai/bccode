import 'dart:async';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// หน้าจอสำหรับ Rebuild Product Balance
/// เรียก API /api/process/product-balance เพื่อประมวลผลยอดคงเหลือ + ค้างรับ + ค้างส่ง ทั้งหมดใหม่
class RebuildProductBalanceScreen extends StatefulWidget {
  const RebuildProductBalanceScreen({super.key});

  @override
  State<RebuildProductBalanceScreen> createState() => _RebuildProductBalanceScreenState();
}

class _RebuildProductBalanceScreenState extends State<RebuildProductBalanceScreen> {
  bool _isProcessing = false;
  String _statusMessage = '';
  final List<String> _logs = [];

  double _progress = 0.0;
  String _currentStep = '';
  Timer? _progressTimer;

  Future<void> _rebuildBalance() async {
    final confirmed = await _showConfirmDialog();
    if (confirmed != true) return;

    setState(() {
      _isProcessing = true;
      _statusMessage = '${global.language("processing")}...';
      _logs.clear();
      _progress = 0.0;
      _currentStep = '${global.language("starting")}...';
    });

    _startProgressSimulation();

    try {
      final shopId = global.getShopId();

      _updateProgress(0.05, '${global.language("preparing_data")}...');
      _addLog('Shop ID: $shopId');
      await Future.delayed(const Duration(milliseconds: 300));

      _updateProgress(0.15, '${global.language("sending_request_to_server")}...');

      final url = global.goApiUrlPath('api/process/product-balance');
      _addLog('URL: $url');

      final result = await global.goApiPost(url, {"shopid": shopId});

      _updateProgress(0.50, '${global.language("received_response_from_server")}...');
      _addLog('Response: $result');

      if (result is Map && result['success'] == true) {
        _updateProgress(0.80, '${global.language("backend_processing")}...');
        await Future.delayed(const Duration(milliseconds: 500));

        _updateProgress(1.0, '${global.language("completed")}!');

        setState(() {
          _statusMessage = '${global.language("rebuild_product_balance")} ${global.language("success")}!';
        });

        if (mounted) {
          global.showSnackBar(
            context,
            const Icon(Icons.check_circle, color: Colors.white),
            '${global.language("rebuild_product_balance")} ${global.language("success")}',
            Colors.green,
          );
        }

        AppLogger.success('Rebuild Product Balance สำเร็จ');
      } else {
        final msg = (result is Map ? result['message'] : result.toString()) ?? 'Unknown error';
        _updateProgress(0.0, global.language('failed'));

        setState(() {
          _statusMessage = '${global.language("rebuild_product_balance")} ${global.language("failed")}: $msg';
        });

        if (mounted) {
          global.showSnackBar(
            context,
            const Icon(Icons.error, color: Colors.white),
            '${global.language("rebuild_product_balance")} ${global.language("failed")}: $msg',
            Colors.red,
          );
        }

        AppLogger.err('Rebuild Product Balance ล้มเหลว: $msg');
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
          const Icon(Icons.error, color: Colors.white),
          '${global.language("error_occurred")}: ${e.toString()}',
          Colors.red,
        );
      }

      AppLogger.err('Rebuild Product Balance error', err: e);
    } finally {
      _stopProgressSimulation();
      if (mounted) {
        setState(() {
          _isProcessing = false;
        });
      }
    }
  }

  void _updateProgress(double progress, String step) {
    if (mounted) {
      setState(() {
        _progress = progress;
        _currentStep = step;
      });
      _addLog(step);
    }
  }

  void _startProgressSimulation() {
    _progressTimer = Timer.periodic(const Duration(milliseconds: 200), (timer) {
      if (!_isProcessing || _progress >= 0.95) {
        timer.cancel();
        return;
      }
      if (mounted && _progress < 0.95) {
        setState(() {
          _progress = (_progress + 0.003).clamp(0.0, 0.95);
        });
      }
    });
  }

  void _stopProgressSimulation() {
    _progressTimer?.cancel();
    _progressTimer = null;
  }

  @override
  void dispose() {
    _stopProgressSimulation();
    super.dispose();
  }

  void _addLog(String message) {
    setState(() {
      _logs.add('[${DateTime.now().toString().substring(11, 19)}] $message');
    });
    AppLogger.info(message);
  }

  Future<bool?> _showConfirmDialog() {
    return showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('confirm_rebuild_product_balance')),
        content: Text(
          global.language('rebuild_product_balance_warning'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.orange),
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
        title: Text(global.language('rebuild_product_balance')),
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
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.info, color: Colors.orange[800]),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            global.language('rebuild_product_balance'),
                            style: TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: Colors.orange[800],
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Text(
                      global.language('rebuild_product_balance_description'),
                      style: const TextStyle(height: 1.5),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // ปุ่ม Rebuild
            ElevatedButton.icon(
              onPressed: _isProcessing ? null : _rebuildBalance,
              icon: _isProcessing
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                      ),
                    )
                  : const Icon(Icons.calculate),
              label: Text(
                _isProcessing ? '${global.language("processing")}...' : global.language('rebuild_product_balance'),
                style: const TextStyle(fontSize: 16),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.deepOrange,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
            const SizedBox(height: 24),

            // Progress Bar
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
                      Row(
                        children: [
                          Icon(Icons.info_outline, color: Colors.blue[700], size: 20),
                          const SizedBox(width: 8),
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
                      const SizedBox(height: 16),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: LinearProgressIndicator(
                          value: _progress,
                          minHeight: 12,
                          backgroundColor: Colors.grey[200],
                          valueColor: AlwaysStoppedAnimation<Color>(
                            _progress >= 1.0 ? Colors.green : Colors.blue.shade600,
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '${(_progress * 100).toStringAsFixed(0)}%',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: _progress >= 1.0 ? Colors.green[800] : Colors.blue[800],
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
              const SizedBox(height: 16),
            ],

            // Status message
            if (_statusMessage.isNotEmpty)
              Card(
                color: _statusMessage.contains(global.language('success'))
                    ? Colors.green[50]
                    : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('error_occurred'))
                        ? Colors.red[50]
                        : Colors.blue[50],
                child: Padding(
                  padding: const EdgeInsets.all(12.0),
                  child: Row(
                    children: [
                      Icon(
                        _statusMessage.contains(global.language('success'))
                            ? Icons.check_circle
                            : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('error_occurred'))
                                ? Icons.error
                                : Icons.info,
                        color: _statusMessage.contains(global.language('success'))
                            ? Colors.green[800]
                            : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('error_occurred'))
                                ? Colors.red[800]
                                : Colors.blue[800],
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _statusMessage,
                          style: TextStyle(
                            color: _statusMessage.contains(global.language('success'))
                                ? Colors.green[800]
                                : _statusMessage.contains(global.language('failed')) || _statusMessage.contains(global.language('error_occurred'))
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
            const SizedBox(height: 16),

            // Logs
            if (_logs.isNotEmpty) ...[
              const Text(
                'Logs:',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
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
                          style: const TextStyle(
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
