import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/vehicle/vehicle_controller.dart';
import 'package:bclms/app/modules/vehicle/vehicle_model.dart';
import 'package:bclms/app/modules/vehicle/widgets/vehicle_form_panel.dart';
import 'package:bclms/app/modules/vehicle/widgets/vehicle_list_panel.dart';

/// หน้าจัดการยานพาหนะ — responsive split view / single view
class VehicleListPage extends ConsumerWidget {
  const VehicleListPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(vehicleProvider);
    final notifier = ref.read(vehicleProvider.notifier);

    // Listen for success/error messages → show snackbar
    ref.listen(vehicleProvider, (prev, next) {
      if (next.successMessage != null && next.successMessage != prev?.successMessage) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.successMessage!),
            backgroundColor: const Color(0xFF2E7D32),
            behavior: SnackBarBehavior.floating,
            margin: const EdgeInsets.all(16),
          ),
        );
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

    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth > 800) {
          return _DesktopLayout(state: state, notifier: notifier);
        }
        return _MobileLayout(state: state, notifier: notifier);
      },
    );
  }
}

/// Desktop: Split View — list ซ้าย 40% | form ขวา 60%
class _DesktopLayout extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

  const _DesktopLayout({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        // Panel ซ้าย — รายการ
        SizedBox(
          width: MediaQuery.of(context).size.width > 1200
              ? 420
              : MediaQuery.of(context).size.width * 0.4,
          child: VehicleListPanel(state: state, notifier: notifier),
        ),
        // เส้นแบ่ง
        Container(
          width: 1,
          color: AppTheme.warmBeige,
        ),
        // Panel ขวา — ฟอร์ม
        Expanded(
          child: VehicleFormPanel(state: state, notifier: notifier),
        ),
      ],
    );
  }
}

/// Mobile: แสดงอย่างใดอย่างหนึ่ง
class _MobileLayout extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

  const _MobileLayout({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final mode = state.screenMode;

    // แสดง list ในโหมด list หรือ multiSelect
    if (mode == ScreenMode.list || mode == ScreenMode.multiSelect) {
      return VehicleListPanel(state: state, notifier: notifier);
    }

    // แสดง form ในโหมดอื่น (view, add, edit)
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) {
          notifier.backToList();
        }
      },
      child: VehicleFormPanel(state: state, notifier: notifier),
    );
  }
}
