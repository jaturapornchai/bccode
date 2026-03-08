import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/warehouse/warehouse_controller.dart';
import 'package:bclms/app/modules/warehouse/warehouse_model.dart';
import 'package:bclms/app/modules/warehouse/widgets/warehouse_form_panel.dart';
import 'package:bclms/app/modules/warehouse/widgets/warehouse_list_panel.dart';

/// หน้าจัดการคลังสินค้า — responsive split view / single view
class WarehouseListPage extends ConsumerWidget {
  const WarehouseListPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(warehouseProvider);
    final notifier = ref.read(warehouseProvider.notifier);

    // Listen for success/error messages → show snackbar
    ref.listen(warehouseProvider, (prev, next) {
      if (next.successMessage != null &&
          next.successMessage != prev?.successMessage) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.successMessage!),
            backgroundColor: const Color(0xFF2E7D32),
            behavior: SnackBarBehavior.floating,
            margin: const EdgeInsets.all(16),
          ),
        );
      }
      if (next.errorMessage != null &&
          next.errorMessage != prev?.errorMessage) {
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
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _DesktopLayout({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        SizedBox(
          width: MediaQuery.of(context).size.width > 1200
              ? 420
              : MediaQuery.of(context).size.width * 0.4,
          child: WarehouseListPanel(state: state, notifier: notifier),
        ),
        Container(width: 1, color: AppTheme.warmBeige),
        Expanded(
          child: WarehouseFormPanel(state: state, notifier: notifier),
        ),
      ],
    );
  }
}

/// Mobile: แสดงอย่างใดอย่างหนึ่ง
class _MobileLayout extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _MobileLayout({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final mode = state.screenMode;

    if (mode == ScreenMode.list || mode == ScreenMode.multiSelect) {
      return WarehouseListPanel(state: state, notifier: notifier);
    }

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) {
          notifier.backToList();
        }
      },
      child: WarehouseFormPanel(state: state, notifier: notifier),
    );
  }
}
