import 'package:flutter/material.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/vehicle/vehicle_controller.dart';
import 'package:bclms/app/modules/vehicle/vehicle_model.dart';
import 'package:bclms/app/modules/vehicle/widgets/vehicle_status_badge.dart';

/// Panel ซ้าย: รายการยานพาหนะ + search + filter + FAB
class VehicleListPanel extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

  const VehicleListPanel({super.key, required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.cream,
      appBar: _buildAppBar(context),
      body: Column(
        children: [
          _SearchBar(notifier: notifier),
          _StatusFilterChips(state: state, notifier: notifier),
          Expanded(child: _VehicleList(state: state, notifier: notifier)),
        ],
      ),
      floatingActionButton: state.screenMode == ScreenMode.multiSelect
          ? null
          : FloatingActionButton(
              onPressed: notifier.startAdd,
              backgroundColor: AppTheme.terracotta,
              child: const Icon(Icons.add, color: Colors.white),
            ),
    );
  }

  PreferredSizeWidget _buildAppBar(BuildContext context) {
    final isMultiSelect = state.screenMode == ScreenMode.multiSelect;

    if (isMultiSelect) {
      return AppBar(
        backgroundColor: AppTheme.earthBrown,
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: notifier.toggleMultiSelect,
        ),
        title: Text('เลือก ${state.selectedIds.length} รายการ'),
        actions: [
          if (state.selectedIds.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'ลบที่เลือก',
              onPressed: () => _confirmDeleteSelected(context),
            ),
        ],
      );
    }

    return AppBar(
      title: const Text('ยานพาหนะ'),
      actions: [
        IconButton(
          icon: const Icon(Icons.checklist_rounded),
          tooltip: 'เลือกหลายรายการ',
          onPressed: notifier.toggleMultiSelect,
        ),
        IconButton(
          icon: const Icon(Icons.refresh),
          tooltip: 'รีเฟรช',
          onPressed: notifier.refreshList,
        ),
      ],
    );
  }

  Future<void> _confirmDeleteSelected(BuildContext context) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: Colors.orange),
            SizedBox(width: 8),
            Text('ลบยานพาหนะ'),
          ],
        ),
        content: Text(
          'ต้องการลบยานพาหนะ ${state.selectedIds.length} รายการหรือไม่?\nการดำเนินการนี้ไม่สามารถย้อนกลับได้',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('ยกเลิก'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('ลบ'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await notifier.deleteSelected();
    }
  }
}

/// Search bar
class _SearchBar extends StatelessWidget {
  final VehicleNotifier notifier;

  const _SearchBar({required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
      child: TextField(
        onChanged: notifier.setSearchQuery,
        decoration: InputDecoration(
          hintText: 'ค้นหา รหัส, ทะเบียน, ชื่อ...',
          hintStyle: TextStyle(
            color: AppTheme.earthBrown.withValues(alpha: 0.4),
            fontSize: 14,
          ),
          prefixIcon: Icon(
            Icons.search,
            color: AppTheme.earthBrown.withValues(alpha: 0.4),
          ),
          filled: true,
          fillColor: Colors.white,
          contentPadding: const EdgeInsets.symmetric(vertical: 10),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide(color: AppTheme.warmBeige),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide(color: AppTheme.warmBeige),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: const BorderSide(color: AppTheme.terracotta, width: 2),
          ),
        ),
      ),
    );
  }
}

/// Chips กรองสถานะ (null=ทั้งหมด, 1=ใช้งาน, 2=ซ่อมบำรุง, 0=ไม่ใช้งาน)
class _StatusFilterChips extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

  const _StatusFilterChips({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final selected = state.statusFilter;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            _FilterChip(
              label: 'ทั้งหมด',
              isSelected: selected == null,
              onTap: () => notifier.setStatusFilter(null),
            ),
            const SizedBox(width: 8),
            _FilterChip(
              label: 'ใช้งาน',
              isSelected: selected == 1,
              color: const Color(0xFF2E7D32),
              onTap: () => notifier.setStatusFilter(1),
            ),
            const SizedBox(width: 8),
            _FilterChip(
              label: 'ซ่อมบำรุง',
              isSelected: selected == 2,
              color: const Color(0xFFE65100),
              onTap: () => notifier.setStatusFilter(2),
            ),
            const SizedBox(width: 8),
            _FilterChip(
              label: 'ไม่ใช้งาน',
              isSelected: selected == 0,
              color: const Color(0xFFD32F2F),
              onTap: () => notifier.setStatusFilter(0),
            ),
          ],
        ),
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool isSelected;
  final Color? color;
  final VoidCallback onTap;

  const _FilterChip({
    required this.label,
    required this.isSelected,
    this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final chipColor = color ?? AppTheme.terracotta;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected
              ? chipColor.withValues(alpha: 0.15)
              : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected ? chipColor : AppTheme.warmBeige,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 12,
            fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
            color: isSelected ? chipColor : AppTheme.earthBrown,
          ),
        ),
      ),
    );
  }
}

/// รายการยานพาหนะ
class _VehicleList extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

  const _VehicleList({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    if (state.isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: AppTheme.terracotta),
      );
    }

    final items = state.filteredVehicles;

    if (items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.local_shipping_outlined,
              size: 64,
              color: AppTheme.earthBrown.withValues(alpha: 0.3),
            ),
            const SizedBox(height: 16),
            Text(
              state.searchQuery.isNotEmpty
                  ? 'ไม่พบยานพาหนะที่ค้นหา'
                  : 'ยังไม่มีข้อมูลยานพาหนะ',
              style: TextStyle(
                fontSize: 16,
                color: AppTheme.earthBrown.withValues(alpha: 0.5),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'กดปุ่ม + เพื่อเพิ่มยานพาหนะ',
              style: TextStyle(
                fontSize: 13,
                color: AppTheme.earthBrown.withValues(alpha: 0.3),
              ),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: notifier.refreshList,
      color: AppTheme.terracotta,
      child: ListView.builder(
        padding: const EdgeInsets.only(bottom: 80),
        itemCount: items.length,
        itemBuilder: (context, index) {
          return _VehicleListItem(
            vehicle: items[index],
            state: state,
            notifier: notifier,
          );
        },
      ),
    );
  }
}

/// แต่ละ item ในรายการ
class _VehicleListItem extends StatelessWidget {
  final VehicleModel vehicle;
  final VehicleState state;
  final VehicleNotifier notifier;

  const _VehicleListItem({
    required this.vehicle,
    required this.state,
    required this.notifier,
  });

  @override
  Widget build(BuildContext context) {
    final isMultiSelect = state.screenMode == ScreenMode.multiSelect;
    final isSelected = state.selectedIds.contains(vehicle.guidFixed);
    final isViewing =
        state.selectedVehicle?.guidFixed == vehicle.guidFixed &&
            state.screenMode != ScreenMode.multiSelect;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 3),
      decoration: BoxDecoration(
        color: isViewing
            ? AppTheme.terracotta.withValues(alpha: 0.08)
            : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isViewing
              ? AppTheme.terracotta.withValues(alpha: 0.3)
              : AppTheme.warmBeige,
        ),
      ),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        leading: isMultiSelect
            ? Checkbox(
                value: isSelected,
                onChanged: (_) =>
                    notifier.toggleItemSelect(vehicle.guidFixed),
              )
            : _VehicleTypeIcon(vehicleType: vehicle.vehicleType),
        title: Text(
          vehicle.displayName.isNotEmpty
              ? vehicle.displayName
              : vehicle.licensePlate,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppTheme.earthBrown,
          ),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(
          '${vehicle.licensePlate} | ${vehicle.code}',
          style: TextStyle(
            fontSize: 12,
            color: AppTheme.earthBrown.withValues(alpha: 0.6),
          ),
        ),
        trailing: VehicleStatusBadge(status: vehicle.status),
        onTap: () {
          if (isMultiSelect) {
            notifier.toggleItemSelect(vehicle.guidFixed);
          } else {
            notifier.selectVehicle(vehicle);
          }
        },
        onLongPress: () {
          if (!isMultiSelect) {
            notifier.toggleMultiSelect();
            notifier.toggleItemSelect(vehicle.guidFixed);
          }
        },
      ),
    );
  }
}

/// Icon แสดงประเภทรถ (ตาม string vehicleType)
class _VehicleTypeIcon extends StatelessWidget {
  final String vehicleType;

  const _VehicleTypeIcon({required this.vehicleType});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 42,
      height: 42,
      decoration: BoxDecoration(
        color: AppTheme.terracotta.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Icon(
        _icon,
        size: 22,
        color: AppTheme.terracotta,
      ),
    );
  }

  IconData get _icon {
    switch (vehicleType) {
      case 'pickup':
        return Icons.local_shipping_rounded;
      case 'van':
        return Icons.airport_shuttle_rounded;
      case 'truck':
      case 'sixwheel':
      case 'tenwheel':
        return Icons.local_shipping_rounded;
      case 'trailer':
        return Icons.rv_hookup_rounded;
      case 'motorcycle':
        return Icons.two_wheeler_rounded;
      default:
        return Icons.directions_car_rounded;
    }
  }
}
