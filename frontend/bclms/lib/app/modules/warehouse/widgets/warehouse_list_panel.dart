import 'package:flutter/material.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/warehouse/warehouse_controller.dart';
import 'package:bclms/app/modules/warehouse/warehouse_model.dart';

/// Panel ซ้าย: รายการคลังสินค้า + search + FAB
class WarehouseListPanel extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const WarehouseListPanel(
      {super.key, required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.cream,
      appBar: _buildAppBar(context),
      body: Column(
        children: [
          _SearchBar(notifier: notifier),
          Expanded(child: _WarehouseList(state: state, notifier: notifier)),
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
      title: const Text('คลังสินค้า'),
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
            Text('ลบคลังสินค้า'),
          ],
        ),
        content: Text(
          'ต้องการลบคลังสินค้า ${state.selectedIds.length} รายการหรือไม่?\nการดำเนินการนี้ไม่สามารถย้อนกลับได้',
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
  final WarehouseNotifier notifier;

  const _SearchBar({required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
      child: TextField(
        onChanged: notifier.setSearchQuery,
        decoration: InputDecoration(
          hintText: 'ค้นหา รหัส, ชื่อคลัง...',
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
            borderSide:
                const BorderSide(color: AppTheme.terracotta, width: 2),
          ),
        ),
      ),
    );
  }
}

/// รายการคลังสินค้า
class _WarehouseList extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _WarehouseList({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    if (state.isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: AppTheme.terracotta),
      );
    }

    final items = state.filteredWarehouses;

    if (items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.warehouse_outlined,
              size: 64,
              color: AppTheme.earthBrown.withValues(alpha: 0.3),
            ),
            const SizedBox(height: 16),
            Text(
              state.searchQuery.isNotEmpty
                  ? 'ไม่พบคลังสินค้าที่ค้นหา'
                  : 'ยังไม่มีข้อมูลคลังสินค้า',
              style: TextStyle(
                fontSize: 16,
                color: AppTheme.earthBrown.withValues(alpha: 0.5),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'กดปุ่ม + เพื่อเพิ่มคลังสินค้า',
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
          return _WarehouseListItem(
            warehouse: items[index],
            state: state,
            notifier: notifier,
          );
        },
      ),
    );
  }
}

/// แต่ละ item ในรายการ
class _WarehouseListItem extends StatelessWidget {
  final WarehouseModel warehouse;
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _WarehouseListItem({
    required this.warehouse,
    required this.state,
    required this.notifier,
  });

  @override
  Widget build(BuildContext context) {
    final isMultiSelect = state.screenMode == ScreenMode.multiSelect;
    final isSelected = state.selectedIds.contains(warehouse.guidFixed);
    final isViewing =
        state.selectedWarehouse?.guidFixed == warehouse.guidFixed &&
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
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        leading: isMultiSelect
            ? Checkbox(
                value: isSelected,
                onChanged: (_) =>
                    notifier.toggleItemSelect(warehouse.guidFixed),
              )
            : _WarehouseIcon(),
        title: Text(
          warehouse.displayName,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppTheme.earthBrown,
          ),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(
          '${warehouse.code} | ${warehouse.locationCount} ที่เก็บ, ${warehouse.totalShelfCount} ชั้นวาง',
          style: TextStyle(
            fontSize: 12,
            color: AppTheme.earthBrown.withValues(alpha: 0.6),
          ),
        ),
        trailing: _CountBadge(count: warehouse.locationCount),
        onTap: () {
          if (isMultiSelect) {
            notifier.toggleItemSelect(warehouse.guidFixed);
          } else {
            notifier.selectWarehouse(warehouse);
          }
        },
        onLongPress: () {
          if (!isMultiSelect) {
            notifier.toggleMultiSelect();
            notifier.toggleItemSelect(warehouse.guidFixed);
          }
        },
      ),
    );
  }
}

class _WarehouseIcon extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      width: 42,
      height: 42,
      decoration: BoxDecoration(
        color: AppTheme.terracotta.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(10),
      ),
      child: const Icon(
        Icons.warehouse_rounded,
        size: 22,
        color: AppTheme.terracotta,
      ),
    );
  }
}

class _CountBadge extends StatelessWidget {
  final int count;

  const _CountBadge({required this.count});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: AppTheme.cloudBlue.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppTheme.cloudBlue.withValues(alpha: 0.3)),
      ),
      child: Text(
        '$count',
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: AppTheme.cloudBlue,
        ),
      ),
    );
  }
}
