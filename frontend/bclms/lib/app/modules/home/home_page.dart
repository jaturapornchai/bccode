import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/home/home_controller.dart';
import 'package:bclms/app/modules/home/widgets/menu_grid_tab.dart';

class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(homeProvider);
    final notifier = ref.read(homeProvider.notifier);

    return Scaffold(
      body: Column(
        children: [
          _ShopStatusBar(state: state),
          Expanded(
            child: IndexedStack(
              index: state.currentIndex,
              children: [
                MenuGridTab(
                  items: HomeNotifier.operationMenus,
                  headerTitle: 'ธุรกรรม',
                ),
                MenuGridTab(
                  items: HomeNotifier.reportMenus,
                  headerTitle: 'รายงาน',
                ),
                MenuGridTab(
                  items: HomeNotifier.masterMenus,
                  headerTitle: 'ข้อมูลหลัก',
                ),
                MenuGridTab(
                  items: HomeNotifier.settingMenus,
                  headerTitle: 'ตั้งค่า',
                ),
              ],
            ),
          ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: state.currentIndex,
        onDestinationSelected: notifier.changeTab,
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.play_circle_outline),
            selectedIcon: Icon(Icons.play_circle),
            label: 'ปฏิบัติการ',
          ),
          NavigationDestination(
            icon: Icon(Icons.bar_chart_outlined),
            selectedIcon: Icon(Icons.bar_chart),
            label: 'รายงาน',
          ),
          NavigationDestination(
            icon: Icon(Icons.inventory_2_outlined),
            selectedIcon: Icon(Icons.inventory_2),
            label: 'ข้อมูลหลัก',
          ),
          NavigationDestination(
            icon: Icon(Icons.settings_outlined),
            selectedIcon: Icon(Icons.settings),
            label: 'ตั้งค่า',
          ),
        ],
      ),
    );
  }
}

/// Status Bar แสดงข้อมูลร้านค้า/ผู้ใช้ — ดูได้ตลอด
class _ShopStatusBar extends StatelessWidget {
  final HomeState state;

  const _ShopStatusBar({required this.state});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: AppTheme.earthBrown.withValues(alpha: 0.08),
        border: Border(
          bottom: BorderSide(color: AppTheme.terracotta.withValues(alpha: 0.15)),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        child: Wrap(
          spacing: 8,
          runSpacing: 4,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            if (state.shopId.isNotEmpty)
              _StatusChip(
                icon: Icons.store,
                label: 'Shop ID: ${_shortId(state.shopId)}',
                color: AppTheme.terracotta,
                tooltip: state.shopId,
              ),
            _StatusChip(
              icon: Icons.cloud,
              label: state.apiUrl,
              color: AppTheme.cloudBlue,
            ),
            if (state.userName.isNotEmpty)
              _StatusChip(
                icon: Icons.person,
                label: state.userName,
                color: AppTheme.earthBrown,
              ),
            if (state.roleName.isNotEmpty)
              _RoleBadge(roleName: state.roleName),
            if (state.branchGuidFixed.isNotEmpty)
              _StatusChip(
                icon: Icons.business,
                label: 'สาขา: ${_shortId(state.branchGuidFixed)}',
                color: const Color(0xFF2E7D32),
                tooltip: state.branchGuidFixed,
              ),
          ],
        ),
      ),
    );
  }

  String _shortId(String id) {
    if (id.length <= 12) return id;
    return '${id.substring(0, 8)}...';
  }
}

class _StatusChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final String? tooltip;

  const _StatusChip({
    required this.icon,
    required this.label,
    required this.color,
    this.tooltip,
  });

  @override
  Widget build(BuildContext context) {
    final chip = Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.2)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 13, color: color),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w500,
              color: color,
            ),
          ),
        ],
      ),
    );

    if (tooltip != null) {
      return Tooltip(message: tooltip!, child: chip);
    }
    return chip;
  }
}

class _RoleBadge extends StatelessWidget {
  final String roleName;

  const _RoleBadge({required this.roleName});

  Color get _badgeColor {
    switch (roleName) {
      case 'Super Admin':
        return const Color(0xFFD32F2F);
      case 'Admin':
        return AppTheme.terracotta;
      default:
        return Colors.grey.shade600;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: _badgeColor,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        roleName,
        style: const TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: Colors.white,
        ),
      ),
    );
  }
}
