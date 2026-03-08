import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/login/login_controller.dart';

class LoginPage extends ConsumerWidget {
  const LoginPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final loginState = ref.watch(loginProvider);
    final notifier = ref.read(loginProvider.notifier);

    // Listen for login success → navigate
    ref.listen(loginProvider, (prev, next) {
      if (next.loginSuccess) {
        context.go('/shop-selection');
      }
      if (next.errorMessage != null && next.errorMessage != prev?.errorMessage) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.errorMessage!),
            backgroundColor: Colors.red.shade700,
            behavior: SnackBarBehavior.floating,
            margin: const EdgeInsets.all(16),
          ),
        );
      }
    });

    return Scaffold(
      body: Stack(
        children: [
          const _BanChiangBackground(),
          Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 420),
                child: _LoginCard(
                  notifier: notifier,
                  state: loginState,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// การ์ดฟอร์ม Login
class _LoginCard extends StatelessWidget {
  final LoginNotifier notifier;
  final LoginState state;

  const _LoginCard({required this.notifier, required this.state});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 8,
      shadowColor: AppTheme.terracotta.withValues(alpha: 0.3),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(24),
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(32, 40, 32, 32),
        child: Form(
          key: notifier.formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Logo
              Container(
                width: 120,
                height: 120,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  boxShadow: [
                    BoxShadow(
                      color: AppTheme.terracotta.withValues(alpha: 0.2),
                      blurRadius: 20,
                      spreadRadius: 2,
                    ),
                  ],
                ),
                child: ClipOval(
                  child: Image.asset(
                    'assets/logos/logo_bc_ai_cloud.png',
                    fit: BoxFit.cover,
                  ),
                ),
              ),
              const SizedBox(height: 20),

              // ชื่อแอป
              const Text(
                'BC LMS',
                style: TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.bold,
                  color: AppTheme.terracotta,
                  letterSpacing: 2,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Logistics Management System',
                style: TextStyle(
                  fontSize: 14,
                  color: AppTheme.earthBrown.withValues(alpha: 0.7),
                  letterSpacing: 1,
                ),
              ),

              const SizedBox(height: 8),
              _BanChiangDivider(),
              const SizedBox(height: 12),

              // Backend URL Selector
              _BackendUrlSelector(state: state, notifier: notifier),
              const SizedBox(height: 16),

              // Username
              TextFormField(
                controller: notifier.usernameController,
                decoration: const InputDecoration(
                  labelText: 'ชื่อผู้ใช้',
                  hintText: 'กรอกชื่อผู้ใช้',
                  prefixIcon: Icon(Icons.person_outline, color: AppTheme.terracotta),
                ),
                textInputAction: TextInputAction.next,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'กรุณากรอกชื่อผู้ใช้';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),

              // Password
              TextFormField(
                controller: notifier.passwordController,
                obscureText: state.obscurePassword,
                decoration: InputDecoration(
                  labelText: 'รหัสผ่าน',
                  hintText: 'กรอกรหัสผ่าน',
                  prefixIcon: const Icon(Icons.lock_outline, color: AppTheme.terracotta),
                  suffixIcon: IconButton(
                    icon: Icon(
                      state.obscurePassword
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined,
                      color: AppTheme.earthBrown.withValues(alpha: 0.5),
                    ),
                    onPressed: notifier.togglePasswordVisibility,
                  ),
                ),
                textInputAction: TextInputAction.done,
                onFieldSubmitted: (_) => notifier.login(),
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'กรุณากรอกรหัสผ่าน';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 4),

              // บันทึกรหัสผ่าน
              CheckboxListTile(
                value: state.rememberLogin,
                onChanged: notifier.toggleRememberLogin,
                title: Text(
                  'บันทึกรหัสผ่าน',
                  style: TextStyle(
                    fontSize: 14,
                    color: AppTheme.earthBrown.withValues(alpha: 0.8),
                  ),
                ),
                controlAffinity: ListTileControlAffinity.leading,
                contentPadding: EdgeInsets.zero,
                dense: true,
                visualDensity: VisualDensity.compact,
              ),
              const SizedBox(height: 20),

              // Login Button
              ElevatedButton(
                onPressed: state.isLoading ? null : notifier.login,
                child: state.isLoading
                    ? const SizedBox(
                        height: 22,
                        width: 22,
                        child: CircularProgressIndicator(
                          strokeWidth: 2.5,
                          color: Colors.white,
                        ),
                      )
                    : const Text('เข้าสู่ระบบ'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// เลือก Backend URL
class _BackendUrlSelector extends StatelessWidget {
  final LoginState state;
  final LoginNotifier notifier;

  const _BackendUrlSelector({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: AppTheme.earthBrown.withValues(alpha: 0.06),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppTheme.earthBrown.withValues(alpha: 0.12)),
      ),
      child: Row(
        children: [
          Icon(Icons.dns_outlined, size: 16, color: AppTheme.earthBrown.withValues(alpha: 0.5)),
          const SizedBox(width: 6),
          // Dropdown
          Expanded(
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                value: state.urlHistory.contains(state.backendUrl) ? state.backendUrl : null,
                hint: Text(
                  state.backendUrl.isNotEmpty ? state.backendUrl : 'เลือกเซิร์ฟเวอร์',
                  style: TextStyle(
                    fontSize: 12,
                    color: AppTheme.earthBrown.withValues(alpha: 0.7),
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
                isExpanded: true,
                isDense: true,
                iconSize: 18,
                iconEnabledColor: AppTheme.earthBrown.withValues(alpha: 0.5),
                style: TextStyle(fontSize: 12, color: AppTheme.earthBrown),
                items: state.urlHistory.map((url) {
                  return DropdownMenuItem<String>(
                    value: url,
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(url, overflow: TextOverflow.ellipsis),
                        ),
                        InkWell(
                          onTap: () {
                            Navigator.of(context).pop();
                            notifier.removeUrlFromHistory(url);
                          },
                          child: Padding(
                            padding: const EdgeInsets.all(4),
                            child: Icon(Icons.close, size: 14,
                                color: AppTheme.earthBrown.withValues(alpha: 0.3)),
                          ),
                        ),
                      ],
                    ),
                  );
                }).toList(),
                onChanged: (url) {
                  if (url != null) notifier.setBackendUrl(url);
                },
              ),
            ),
          ),
          // Copy
          _UrlActionButton(
            icon: Icons.copy,
            tooltip: 'คัดลอก URL',
            onTap: () {
              if (state.backendUrl.isNotEmpty) {
                Clipboard.setData(ClipboardData(text: state.backendUrl));
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('คัดลอก URL แล้ว'),
                    duration: Duration(seconds: 1),
                    behavior: SnackBarBehavior.floating,
                  ),
                );
              }
            },
          ),
          const SizedBox(width: 4),
          // Add
          _UrlActionButton(
            icon: Icons.add,
            tooltip: 'เพิ่มเซิร์ฟเวอร์',
            onTap: () => _showAddUrlDialog(context),
          ),
          const SizedBox(width: 4),
          // Test connection
          _UrlTestButton(
            status: state.connectionTestStatus,
            onTap: notifier.testConnection,
          ),
        ],
      ),
    );
  }

  void _showAddUrlDialog(BuildContext context) {
    final addController = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Row(
          children: [
            Icon(Icons.dns, color: AppTheme.terracotta, size: 22),
            SizedBox(width: 8),
            Text('เพิ่มเซิร์ฟเวอร์'),
          ],
        ),
        content: TextField(
          controller: addController,
          autofocus: true,
          decoration: const InputDecoration(
            hintText: 'http://192.168.0.5:9090',
            labelText: 'Server URL',
          ),
          onSubmitted: (_) {
            final url = addController.text.trim();
            if (url.isNotEmpty) {
              Navigator.pop(ctx);
              notifier.addBackendUrl(url);
            }
          },
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('ยกเลิก'),
          ),
          FilledButton(
            onPressed: () {
              final url = addController.text.trim();
              if (url.isNotEmpty) {
                Navigator.pop(ctx);
                notifier.addBackendUrl(url);
              }
            },
            child: const Text('บันทึก'),
          ),
        ],
      ),
    );
  }
}

class _UrlActionButton extends StatelessWidget {
  final IconData icon;
  final String tooltip;
  final VoidCallback onTap;

  const _UrlActionButton({
    required this.icon,
    required this.tooltip,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: tooltip,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(6),
        child: Container(
          padding: const EdgeInsets.all(5),
          decoration: BoxDecoration(
            color: AppTheme.earthBrown.withValues(alpha: 0.08),
            borderRadius: BorderRadius.circular(6),
          ),
          child: Icon(icon, size: 16, color: AppTheme.earthBrown.withValues(alpha: 0.5)),
        ),
      ),
    );
  }
}

class _UrlTestButton extends StatelessWidget {
  final String? status;
  final VoidCallback onTap;

  const _UrlTestButton({required this.status, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final Color bgColor;
    final Widget child;

    switch (status) {
      case 'testing':
        bgColor = AppTheme.earthBrown.withValues(alpha: 0.08);
        child = SizedBox(
          width: 14, height: 14,
          child: CircularProgressIndicator(
            strokeWidth: 2,
            color: AppTheme.earthBrown.withValues(alpha: 0.5),
          ),
        );
      case 'success':
        bgColor = const Color(0xFF2E7D32).withValues(alpha: 0.15);
        child = const Icon(Icons.check_circle, size: 16, color: Color(0xFF2E7D32));
      case 'failed':
        bgColor = const Color(0xFFD32F2F).withValues(alpha: 0.15);
        child = const Icon(Icons.error, size: 16, color: Color(0xFFD32F2F));
      default:
        bgColor = AppTheme.earthBrown.withValues(alpha: 0.08);
        child = Icon(Icons.wifi_tethering, size: 16, color: AppTheme.earthBrown.withValues(alpha: 0.5));
    }

    return Tooltip(
      message: 'ทดสอบการเชื่อมต่อ',
      child: InkWell(
        onTap: status == 'testing' ? null : onTap,
        borderRadius: BorderRadius.circular(6),
        child: Container(
          padding: const EdgeInsets.all(5),
          decoration: BoxDecoration(
            color: bgColor,
            borderRadius: BorderRadius.circular(6),
          ),
          child: child,
        ),
      ),
    );
  }
}

/// เส้นคั่นลายก้นหอยบ้านเชียง
class _BanChiangDivider extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Expanded(child: _spiralLine(false)),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Icon(
              Icons.diamond_outlined,
              size: 14,
              color: AppTheme.terracotta.withValues(alpha: 0.5),
            ),
          ),
          Expanded(child: _spiralLine(true)),
        ],
      ),
    );
  }

  Widget _spiralLine(bool reverse) {
    return Container(
      height: 2,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: reverse
              ? [AppTheme.terracotta.withValues(alpha: 0.5), AppTheme.terracotta.withValues(alpha: 0.05)]
              : [AppTheme.terracotta.withValues(alpha: 0.05), AppTheme.terracotta.withValues(alpha: 0.5)],
        ),
        borderRadius: BorderRadius.circular(1),
      ),
    );
  }
}

/// พื้นหลังลายบ้านเชียง
class _BanChiangBackground extends StatelessWidget {
  const _BanChiangBackground();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [
            AppTheme.cream,
            Color(0xFFF0E4D4),
            AppTheme.warmBeige,
          ],
        ),
      ),
      child: CustomPaint(
        size: Size.infinite,
        painter: _BanChiangPatternPainter(),
      ),
    );
  }
}

/// วาดลายก้นหอยบ้านเชียงบนพื้นหลัง
class _BanChiangPatternPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = AppTheme.terracotta.withValues(alpha: 0.04)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.0;

    _drawSpiral(canvas, paint, Offset(size.width * 0.1, size.height * 0.08), 60);
    _drawSpiral(canvas, paint, Offset(size.width * 0.9, size.height * 0.15), 45);
    _drawSpiral(canvas, paint, Offset(size.width * 0.15, size.height * 0.85), 50);
    _drawSpiral(canvas, paint, Offset(size.width * 0.85, size.height * 0.9), 55);

    final borderPaint = Paint()
      ..color = AppTheme.terracotta.withValues(alpha: 0.08)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3.0;

    _drawWavyLine(canvas, borderPaint, 0, size.width, 20);
    _drawWavyLine(canvas, borderPaint, 0, size.width, size.height - 20);
  }

  void _drawSpiral(Canvas canvas, Paint paint, Offset center, double maxRadius) {
    final path = Path();
    const turns = 3.0;
    const points = 100;

    for (int i = 0; i < points; i++) {
      final t = i / points;
      final angle = turns * 2 * 3.14159 * t;
      final radius = maxRadius * t;
      final x = center.dx + radius * _cos(angle);
      final y = center.dy + radius * _sin(angle);

      if (i == 0) {
        path.moveTo(x, y);
      } else {
        path.lineTo(x, y);
      }
    }
    canvas.drawPath(path, paint);
  }

  void _drawWavyLine(Canvas canvas, Paint paint, double startX, double endX, double y) {
    final path = Path();
    path.moveTo(startX, y);
    const waveLength = 40.0;
    const amplitude = 5.0;

    for (double x = startX; x <= endX; x += 1) {
      final waveY = y + amplitude * _sin((x / waveLength) * 2 * 3.14159);
      path.lineTo(x, waveY);
    }
    canvas.drawPath(path, paint);
  }

  double _sin(double x) => _sinApprox(x);
  double _cos(double x) => _sinApprox(x + 3.14159 / 2);

  double _sinApprox(double x) {
    x = x % (2 * 3.14159);
    if (x > 3.14159) x -= 2 * 3.14159;
    final x3 = x * x * x;
    final x5 = x3 * x * x;
    return x - x3 / 6 + x5 / 120;
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
