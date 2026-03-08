import 'package:flutter/material.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/warehouse/warehouse_controller.dart';
import 'package:bclms/app/modules/warehouse/warehouse_model.dart';

/// Panel ขวา: ฟอร์มคลังสินค้า (view/edit/add) + drill-down Location/Shelf
class WarehouseFormPanel extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const WarehouseFormPanel(
      {super.key, required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final mode = state.screenMode;

    // ยังไม่ได้เลือก → แสดงข้อความเชิญ
    if (mode == ScreenMode.list || mode == ScreenMode.multiSelect) {
      return _EmptyState();
    }

    final isEditable = mode == ScreenMode.add || mode == ScreenMode.edit;
    final title = switch (mode) {
      ScreenMode.add => 'เพิ่มคลังสินค้า',
      ScreenMode.edit => 'แก้ไขคลังสินค้า',
      ScreenMode.view => 'รายละเอียดคลังสินค้า',
      _ => '',
    };

    return Scaffold(
      backgroundColor: AppTheme.cream,
      appBar: AppBar(
        title: Text(title),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: notifier.cancelForm,
        ),
        actions: _buildActions(context, mode),
      ),
      body: isEditable
          ? _WarehouseEditForm(state: state, notifier: notifier)
          : _WarehouseViewDetail(state: state, notifier: notifier),
      bottomNavigationBar:
          isEditable ? _SaveBar(state: state, notifier: notifier) : null,
    );
  }

  List<Widget> _buildActions(BuildContext context, ScreenMode mode) {
    if (mode == ScreenMode.view) {
      return [
        IconButton(
          icon: const Icon(Icons.edit_outlined),
          tooltip: 'แก้ไข',
          onPressed: notifier.startEdit,
        ),
        IconButton(
          icon: const Icon(Icons.delete_outline),
          tooltip: 'ลบ',
          onPressed: () => _confirmDelete(context),
        ),
      ];
    }
    return [];
  }

  Future<void> _confirmDelete(BuildContext context) async {
    final warehouse = state.selectedWarehouse;
    if (warehouse == null) return;

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
          'ต้องการลบ "${warehouse.displayName}" หรือไม่?\nรวมถึงที่เก็บและชั้นวางทั้งหมด\nการดำเนินการนี้ไม่สามารถย้อนกลับได้',
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
      await notifier.deleteWarehouse();
    }
  }
}

// ============================================================
// Empty State
// ============================================================

class _EmptyState extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.warehouse_outlined,
            size: 80,
            color: AppTheme.earthBrown.withValues(alpha: 0.15),
          ),
          const SizedBox(height: 16),
          Text(
            'เลือกคลังสินค้าจากรายการ',
            style: TextStyle(
              fontSize: 16,
              color: AppTheme.earthBrown.withValues(alpha: 0.4),
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'หรือกด + เพื่อเพิ่มใหม่',
            style: TextStyle(
              fontSize: 13,
              color: AppTheme.earthBrown.withValues(alpha: 0.3),
            ),
          ),
        ],
      ),
    );
  }
}

// ============================================================
// Save Bar
// ============================================================

class _SaveBar extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _SaveBar({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 8,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: OutlinedButton(
              onPressed: notifier.cancelForm,
              style: OutlinedButton.styleFrom(
                minimumSize: const Size(0, 48),
                side: const BorderSide(color: AppTheme.terracotta),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: const Text(
                'ยกเลิก',
                style: TextStyle(color: AppTheme.terracotta),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            flex: 2,
            child: ElevatedButton(
              onPressed: state.isSaving ? null : notifier.saveWarehouse,
              style: ElevatedButton.styleFrom(
                minimumSize: const Size(0, 48),
                backgroundColor: AppTheme.terracotta,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: state.isSaving
                  ? const SizedBox(
                      width: 22,
                      height: 22,
                      child: CircularProgressIndicator(
                        strokeWidth: 2.5,
                        color: Colors.white,
                      ),
                    )
                  : const Text('บันทึก'),
            ),
          ),
        ],
      ),
    );
  }
}

// ============================================================
// Edit Form (Add/Edit mode)
// ============================================================

class _WarehouseEditForm extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _WarehouseEditForm({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    return Form(
      key: notifier.formKey,
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            _SectionCard(
              title: 'ข้อมูลคลังสินค้า',
              icon: Icons.warehouse_outlined,
              initiallyExpanded: true,
              children: [
                _buildTextField('รหัสคลัง *', notifier.codeCtrl,
                    validator: _required),
                _buildTextField('ชื่อคลัง (ไทย) *', notifier.nameThCtrl,
                    validator: _required),
                _buildTextField('ชื่อคลัง (อังกฤษ)', notifier.nameEnCtrl),
              ],
            ),
            const SizedBox(height: 80),
          ],
        ),
      ),
    );
  }
}

// ============================================================
// View Detail (with drill-down Location/Shelf)
// ============================================================

class _WarehouseViewDetail extends StatelessWidget {
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _WarehouseViewDetail({required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final wh = state.selectedWarehouse;
    if (wh == null) return const SizedBox.shrink();

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          // ข้อมูลคลัง
          _SectionCard(
            title: 'ข้อมูลคลังสินค้า',
            icon: Icons.warehouse_outlined,
            initiallyExpanded: true,
            children: [
              _viewField('รหัสคลัง', wh.code),
              _viewField('ชื่อ (ไทย)', wh.displayName),
              _viewField(
                  'ชื่อ (อังกฤษ)',
                  wh.names
                          .where((n) => n.code == 'en')
                          .firstOrNull
                          ?.name ??
                      '-'),
              _viewField('จำนวนที่เก็บ', '${wh.locationCount}'),
              _viewField('จำนวนชั้นวางรวม', '${wh.totalShelfCount}'),
            ],
          ),
          const SizedBox(height: 8),

          // รายการที่เก็บ (Locations)
          _SectionCard(
            title: 'ที่เก็บ (${wh.locations.length})',
            icon: Icons.location_on_outlined,
            initiallyExpanded: true,
            trailing: IconButton(
              icon: const Icon(Icons.add_circle_outline,
                  color: AppTheme.terracotta, size: 22),
              tooltip: 'เพิ่มที่เก็บ',
              onPressed: () {
                notifier.clearLocationForm();
                // ไม่ต้อง selectLocation — ให้ selectedLocation = null เพื่อบอก saveLocation ว่าเป็นการเพิ่มใหม่
                _showLocationDialog(context, isNew: true);
              },
            ),
            children: [
              if (wh.locations.isEmpty)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  child: Text(
                    'ยังไม่มีที่เก็บ — กดปุ่ม + เพื่อเพิ่ม',
                    style: TextStyle(
                      fontSize: 13,
                      color: AppTheme.earthBrown.withValues(alpha: 0.4),
                    ),
                  ),
                )
              else
                ...wh.locations.map((loc) => _LocationTile(
                      location: loc,
                      state: state,
                      notifier: notifier,
                    )),
            ],
          ),
          const SizedBox(height: 80),
        ],
      ),
    );
  }

  void _showLocationDialog(BuildContext context, {required bool isNew}) {
    showDialog(
      context: context,
      builder: (ctx) => _LocationFormDialog(
        notifier: notifier,
        isNew: isNew,
      ),
    );
  }
}

// ============================================================
// Location Tile (expandable — shows shelves)
// ============================================================

class _LocationTile extends StatelessWidget {
  final LocationModel location;
  final WarehouseState state;
  final WarehouseNotifier notifier;

  const _LocationTile({
    required this.location,
    required this.state,
    required this.notifier,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      color: AppTheme.warmBeige.withValues(alpha: 0.5),
      margin: const EdgeInsets.only(bottom: 8),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      child: ExpansionTile(
        tilePadding: const EdgeInsets.symmetric(horizontal: 12),
        childrenPadding: const EdgeInsets.fromLTRB(12, 0, 12, 8),
        leading: Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: AppTheme.cloudBlue.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(8),
          ),
          child: const Icon(Icons.location_on_rounded,
              size: 18, color: AppTheme.cloudBlue),
        ),
        title: Text(
          location.displayName,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: AppTheme.earthBrown,
          ),
        ),
        subtitle: Text(
          '${location.code} | ${location.shelves.length} ชั้นวาง',
          style: TextStyle(
            fontSize: 11,
            color: AppTheme.earthBrown.withValues(alpha: 0.5),
          ),
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            IconButton(
              icon: Icon(Icons.edit_outlined,
                  size: 18,
                  color: AppTheme.earthBrown.withValues(alpha: 0.5)),
              tooltip: 'แก้ไขที่เก็บ',
              onPressed: () {
                notifier.populateLocationForm(location);
                notifier.selectLocation(location);
                _showLocationDialog(context, isNew: false);
              },
            ),
            IconButton(
              icon: Icon(Icons.delete_outline,
                  size: 18, color: Colors.red.withValues(alpha: 0.6)),
              tooltip: 'ลบที่เก็บ',
              onPressed: () => _confirmDeleteLocation(context),
            ),
          ],
        ),
        children: [
          // ปุ่มเพิ่มชั้นวาง
          Align(
            alignment: Alignment.centerRight,
            child: TextButton.icon(
              icon: const Icon(Icons.add, size: 16),
              label: const Text('เพิ่มชั้นวาง', style: TextStyle(fontSize: 12)),
              style: TextButton.styleFrom(
                foregroundColor: AppTheme.terracotta,
                padding:
                    const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              ),
              onPressed: () {
                notifier.selectLocation(location);
                notifier.clearShelfForm();
                _showShelfDialog(context, isNew: true);
              },
            ),
          ),
          if (location.shelves.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Text(
                'ยังไม่มีชั้นวาง',
                style: TextStyle(
                  fontSize: 12,
                  color: AppTheme.earthBrown.withValues(alpha: 0.4),
                ),
              ),
            )
          else
            ...location.shelves.map((shelf) => _ShelfTile(
                  shelf: shelf,
                  location: location,
                  notifier: notifier,
                )),
        ],
      ),
    );
  }

  void _showLocationDialog(BuildContext context, {required bool isNew}) {
    showDialog(
      context: context,
      builder: (ctx) => _LocationFormDialog(
        notifier: notifier,
        isNew: isNew,
      ),
    );
  }

  void _showShelfDialog(BuildContext context, {required bool isNew}) {
    showDialog(
      context: context,
      builder: (ctx) => _ShelfFormDialog(
        notifier: notifier,
        isNew: isNew,
      ),
    );
  }

  Future<void> _confirmDeleteLocation(BuildContext context) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: Colors.orange),
            SizedBox(width: 8),
            Text('ลบที่เก็บ'),
          ],
        ),
        content: Text(
          'ต้องการลบ "${location.displayName}" หรือไม่?\nรวมถึงชั้นวางทั้งหมดในที่เก็บนี้',
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
      await notifier.deleteLocation(location.code);
    }
  }
}

// ============================================================
// Shelf Tile
// ============================================================

class _ShelfTile extends StatelessWidget {
  final ShelfModel shelf;
  final LocationModel location;
  final WarehouseNotifier notifier;

  const _ShelfTile({
    required this.shelf,
    required this.location,
    required this.notifier,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 4),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.warmBeige),
      ),
      child: Row(
        children: [
          Container(
            width: 28,
            height: 28,
            decoration: BoxDecoration(
              color: AppTheme.terracotta.withValues(alpha: 0.08),
              borderRadius: BorderRadius.circular(6),
            ),
            child: const Icon(Icons.shelves,
                size: 14, color: AppTheme.terracotta),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  shelf.name.isNotEmpty ? shelf.name : shelf.code,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: AppTheme.earthBrown,
                  ),
                ),
                Text(
                  '${shelf.code} | สินค้า ${shelf.productItems.length} รายการ'
                  '${shelf.min > 0 || shelf.max > 0 ? ' | Min:${shelf.min} Max:${shelf.max}' : ''}',
                  style: TextStyle(
                    fontSize: 10,
                    color: AppTheme.earthBrown.withValues(alpha: 0.5),
                  ),
                ),
              ],
            ),
          ),
          IconButton(
            icon: Icon(Icons.edit_outlined,
                size: 16,
                color: AppTheme.earthBrown.withValues(alpha: 0.5)),
            tooltip: 'แก้ไขชั้นวาง',
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            onPressed: () {
              notifier.selectLocation(location);
              notifier.populateShelfForm(shelf);
              notifier.selectShelf(shelf);
              _showShelfDialog(context, isNew: false);
            },
          ),
          IconButton(
            icon: Icon(Icons.delete_outline,
                size: 16, color: Colors.red.withValues(alpha: 0.5)),
            tooltip: 'ลบชั้นวาง',
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            onPressed: () => _confirmDeleteShelf(context),
          ),
        ],
      ),
    );
  }

  void _showShelfDialog(BuildContext context, {required bool isNew}) {
    showDialog(
      context: context,
      builder: (ctx) => _ShelfFormDialog(
        notifier: notifier,
        isNew: isNew,
      ),
    );
  }

  Future<void> _confirmDeleteShelf(BuildContext context) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: Colors.orange),
            SizedBox(width: 8),
            Text('ลบชั้นวาง'),
          ],
        ),
        content: Text(
          'ต้องการลบชั้นวาง "${shelf.name.isNotEmpty ? shelf.name : shelf.code}" หรือไม่?',
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
      await notifier.deleteShelf(shelf.code);
    }
  }
}

// ============================================================
// Location Form Dialog
// ============================================================

class _LocationFormDialog extends StatelessWidget {
  final WarehouseNotifier notifier;
  final bool isNew;

  const _LocationFormDialog({required this.notifier, required this.isNew});

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      title: Text(isNew ? 'เพิ่มที่เก็บ' : 'แก้ไขที่เก็บ'),
      content: SizedBox(
        width: 400,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildTextField('รหัสที่เก็บ *', notifier.locCodeCtrl),
            _buildTextField('ชื่อที่เก็บ (ไทย)', notifier.locNameThCtrl),
            _buildTextField('ชื่อที่เก็บ (อังกฤษ)', notifier.locNameEnCtrl),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('ยกเลิก'),
        ),
        ElevatedButton(
          onPressed: () async {
            final success = await notifier.saveLocation();
            if (success && context.mounted) {
              Navigator.of(context).pop();
            }
          },
          style: ElevatedButton.styleFrom(
            backgroundColor: AppTheme.terracotta,
            foregroundColor: Colors.white,
          ),
          child: const Text('บันทึก'),
        ),
      ],
    );
  }
}

// ============================================================
// Shelf Form Dialog
// ============================================================

class _ShelfFormDialog extends StatelessWidget {
  final WarehouseNotifier notifier;
  final bool isNew;

  const _ShelfFormDialog({required this.notifier, required this.isNew});

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      title: Text(isNew ? 'เพิ่มชั้นวาง' : 'แก้ไขชั้นวาง'),
      content: SizedBox(
        width: 400,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildTextField('รหัสชั้นวาง *', notifier.shelfCodeCtrl),
            _buildTextField('ชื่อชั้นวาง', notifier.shelfNameCtrl),
            _buildTextField('จำนวนขั้นต่ำ (Min)', notifier.shelfMinCtrl,
                keyboardType: TextInputType.number),
            _buildTextField('จำนวนสูงสุด (Max)', notifier.shelfMaxCtrl,
                keyboardType: TextInputType.number),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('ยกเลิก'),
        ),
        ElevatedButton(
          onPressed: () async {
            final success = await notifier.saveShelf();
            if (success && context.mounted) {
              Navigator.of(context).pop();
            }
          },
          style: ElevatedButton.styleFrom(
            backgroundColor: AppTheme.terracotta,
            foregroundColor: Colors.white,
          ),
          child: const Text('บันทึก'),
        ),
      ],
    );
  }
}

// ============================================================
// Shared Widgets
// ============================================================

/// Card ครอบ section พร้อม ExpansionTile
class _SectionCard extends StatelessWidget {
  final String title;
  final IconData icon;
  final bool initiallyExpanded;
  final List<Widget> children;
  final Widget? trailing;

  const _SectionCard({
    required this.title,
    required this.icon,
    this.initiallyExpanded = false,
    required this.children,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 1,
      color: Colors.white,
      surfaceTintColor: Colors.transparent,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ExpansionTile(
        initiallyExpanded: initiallyExpanded,
        tilePadding: const EdgeInsets.symmetric(horizontal: 16),
        childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        leading: Icon(icon, size: 20, color: AppTheme.terracotta),
        title: Row(
          children: [
            Text(
              title,
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: AppTheme.earthBrown,
              ),
            ),
            if (trailing != null) ...[const Spacer(), trailing!],
          ],
        ),
        children: children,
      ),
    );
  }
}

/// TextField สำหรับฟอร์ม
Widget _buildTextField(
  String label,
  TextEditingController controller, {
  TextInputType? keyboardType,
  String? Function(String?)? validator,
}) {
  return Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: TextFormField(
      controller: controller,
      keyboardType: keyboardType,
      validator: validator,
      decoration: InputDecoration(
        labelText: label,
        isDense: true,
      ),
    ),
  );
}

/// แสดงค่าแบบ read-only (label: value)
Widget _viewField(String label, String value) {
  return Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 120,
          child: Text(
            label,
            style: TextStyle(
              fontSize: 13,
              color: AppTheme.earthBrown.withValues(alpha: 0.6),
            ),
          ),
        ),
        Expanded(
          child: Text(
            value.isNotEmpty ? value : '-',
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w500,
              color: AppTheme.earthBrown,
            ),
          ),
        ),
      ],
    ),
  );
}

String? _required(String? value) {
  if (value == null || value.trim().isEmpty) {
    return 'กรุณากรอกข้อมูล';
  }
  return null;
}
