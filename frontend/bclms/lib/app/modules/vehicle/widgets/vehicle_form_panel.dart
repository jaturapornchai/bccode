import 'package:flutter/material.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/core/utils/global.dart';
import 'package:bclms/app/modules/vehicle/vehicle_controller.dart';
import 'package:bclms/app/modules/vehicle/vehicle_model.dart';
import 'package:bclms/app/modules/vehicle/widgets/vehicle_status_badge.dart';

/// Panel ขวา: ฟอร์มยานพาหนะ (view/edit/add) — ตาม Backend API schema
class VehicleFormPanel extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

  const VehicleFormPanel({super.key, required this.state, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final mode = state.screenMode;

    // ยังไม่ได้เลือก → แสดงข้อความเชิญ
    if (mode == ScreenMode.list || mode == ScreenMode.multiSelect) {
      return _EmptyState();
    }

    final isEditable = mode == ScreenMode.add || mode == ScreenMode.edit;
    final title = switch (mode) {
      ScreenMode.add => 'เพิ่มยานพาหนะ',
      ScreenMode.edit => 'แก้ไขยานพาหนะ',
      ScreenMode.view => 'รายละเอียดยานพาหนะ',
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
      body: Form(
        key: notifier.formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              _GeneralInfoSection(
                state: state,
                notifier: notifier,
                isEditable: isEditable,
              ),
              const SizedBox(height: 8),
              _SpecsSection(
                state: state,
                notifier: notifier,
                isEditable: isEditable,
              ),
              const SizedBox(height: 8),
              _DriverSection(
                state: state,
                notifier: notifier,
                isEditable: isEditable,
              ),
              const SizedBox(height: 8),
              _RegulatorySection(
                state: state,
                notifier: notifier,
                isEditable: isEditable,
              ),
              const SizedBox(height: 8),
              _NotesSection(
                notifier: notifier,
                isEditable: isEditable,
                notes: state.selectedVehicle?.notes ?? '',
              ),
              const SizedBox(height: 80),
            ],
          ),
        ),
      ),
      bottomNavigationBar: isEditable ? _SaveBar(state: state, notifier: notifier) : null,
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
    final vehicle = state.selectedVehicle;
    if (vehicle == null) return;

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
          'ต้องการลบ "${vehicle.displayName.isNotEmpty ? vehicle.displayName : vehicle.licensePlate}" หรือไม่?\nการดำเนินการนี้ไม่สามารถย้อนกลับได้',
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
      await notifier.deleteVehicle();
    }
  }
}

/// แสดงเมื่อยังไม่ได้เลือกยานพาหนะ
class _EmptyState extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.local_shipping_outlined,
            size: 80,
            color: AppTheme.earthBrown.withValues(alpha: 0.15),
          ),
          const SizedBox(height: 16),
          Text(
            'เลือกยานพาหนะจากรายการ',
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

/// ปุ่มบันทึก/ยกเลิก
class _SaveBar extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;

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
              onPressed: state.isSaving ? null : notifier.saveVehicle,
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
// Section 1: ข้อมูลทั่วไป
// ============================================================

class _GeneralInfoSection extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;
  final bool isEditable;

  const _GeneralInfoSection({
    required this.state,
    required this.notifier,
    required this.isEditable,
  });

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: 'ข้อมูลทั่วไป',
      icon: Icons.info_outline_rounded,
      initiallyExpanded: true,
      children: [
        if (isEditable) ...[
          _buildTextField('รหัสยานพาหนะ *', notifier.codeCtrl,
              validator: _required),
          _buildTextField('ชื่อยานพาหนะ (ไทย)', notifier.nameThCtrl),
          _buildVehicleTypeDropdown(),
          _buildTextField('ทะเบียนรถ', notifier.licensePlateCtrl),
          _buildTextField('จังหวัด', notifier.provinceCtrl),
          _buildTextField('ยี่ห้อ', notifier.brandCtrl),
          _buildTextField('รุ่น', notifier.modelCtrl),
          _buildTextField('ปีรถ', notifier.yearCtrl,
              keyboardType: TextInputType.number),
          _buildTextField('สี', notifier.colorCtrl),
          _buildStatusDropdown(),
        ] else ...[
          _viewField('รหัสยานพาหนะ', state.selectedVehicle?.code ?? ''),
          _viewField('ชื่อยานพาหนะ', state.selectedVehicle?.displayName ?? ''),
          _viewField('ประเภทรถ', state.selectedVehicle?.vehicleTypeName ?? ''),
          _viewField('ทะเบียนรถ', state.selectedVehicle?.licensePlate ?? ''),
          _viewField('จังหวัด', state.selectedVehicle?.province ?? ''),
          _viewField('ยี่ห้อ', state.selectedVehicle?.brand ?? ''),
          _viewField('รุ่น', state.selectedVehicle?.model ?? ''),
          _viewField('ปีรถ', _yearText()),
          _viewField('สี', state.selectedVehicle?.color ?? ''),
          _viewFieldWithBadge('สถานะ', state.selectedVehicle?.status ?? 1),
        ],
      ],
    );
  }

  String _yearText() {
    final year = state.selectedVehicle?.year ?? 0;
    return year > 0 ? year.toString() : '';
  }

  Widget _buildVehicleTypeDropdown() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        initialValue: state.vehicleTypeValue.isEmpty
            ? null
            : state.vehicleTypeValue,
        decoration: const InputDecoration(labelText: 'ประเภทรถ'),
        items: const [
          DropdownMenuItem(value: 'pickup', child: Text('รถกระบะ')),
          DropdownMenuItem(value: 'van', child: Text('รถตู้')),
          DropdownMenuItem(value: 'truck', child: Text('รถบรรทุก')),
          DropdownMenuItem(value: 'sixwheel', child: Text('รถ 6 ล้อ')),
          DropdownMenuItem(value: 'tenwheel', child: Text('รถ 10 ล้อ')),
          DropdownMenuItem(value: 'trailer', child: Text('หัวลาก')),
          DropdownMenuItem(value: 'motorcycle', child: Text('มอเตอร์ไซค์')),
        ],
        onChanged: (v) => notifier.setVehicleType(v ?? ''),
      ),
    );
  }

  Widget _buildStatusDropdown() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<int>(
        initialValue: state.statusValue,
        decoration: const InputDecoration(labelText: 'สถานะ *'),
        items: const [
          DropdownMenuItem(value: 1, child: Text('ใช้งาน')),
          DropdownMenuItem(value: 2, child: Text('ซ่อมบำรุง')),
          DropdownMenuItem(value: 0, child: Text('ไม่ใช้งาน')),
        ],
        onChanged: (v) => notifier.setStatus(v ?? 1),
      ),
    );
  }

  Widget _viewFieldWithBadge(String label, int status) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
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
          VehicleStatusBadge(status: status),
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
}

// ============================================================
// Section 2: ข้อมูลตัวรถ & ความจุ
// ============================================================

class _SpecsSection extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;
  final bool isEditable;

  const _SpecsSection({
    required this.state,
    required this.notifier,
    required this.isEditable,
  });

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: 'ข้อมูลตัวรถ & ความจุ',
      icon: Icons.build_outlined,
      children: [
        if (isEditable) ...[
          _buildFuelTypeDropdown(),
          _buildTextField(
            'น้ำหนักบรรทุกสูงสุด (kg)',
            notifier.capacityWeightCtrl,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
          ),
          _buildTextField(
            'ปริมาตรสูงสุด (CBM)',
            notifier.capacityVolumeCtrl,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
          ),
        ] else ...[
          _viewField('ประเภทเชื้อเพลิง',
              state.selectedVehicle?.fuelTypeName ?? ''),
          _viewField('น้ำหนักสูงสุด', _weightText()),
          _viewField('ปริมาตรสูงสุด', _volumeText()),
        ],
      ],
    );
  }

  String _weightText() {
    final w = state.selectedVehicle?.capacityWeight;
    return w != null ? '$w kg' : '';
  }

  String _volumeText() {
    final v = state.selectedVehicle?.capacityVolume;
    return v != null ? '$v CBM' : '';
  }

  Widget _buildFuelTypeDropdown() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        initialValue: state.fuelTypeValue.isEmpty
            ? null
            : state.fuelTypeValue,
        decoration: const InputDecoration(labelText: 'ประเภทเชื้อเพลิง'),
        items: const [
          DropdownMenuItem(value: 'diesel', child: Text('ดีเซล')),
          DropdownMenuItem(value: 'gasoline', child: Text('เบนซิน')),
          DropdownMenuItem(value: 'electric', child: Text('ไฟฟ้า (EV)')),
          DropdownMenuItem(value: 'lpg', child: Text('LPG')),
          DropdownMenuItem(value: 'cng', child: Text('CNG')),
          DropdownMenuItem(value: 'hybrid', child: Text('ไฮบริด')),
        ],
        onChanged: (v) => notifier.setFuelType(v ?? ''),
      ),
    );
  }
}

// ============================================================
// Section 3: คนขับ
// ============================================================

class _DriverSection extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;
  final bool isEditable;

  const _DriverSection({
    required this.state,
    required this.notifier,
    required this.isEditable,
  });

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: 'คนขับที่มอบหมาย',
      icon: Icons.badge_outlined,
      children: [
        if (isEditable) ...[
          _buildTextField('รหัสคนขับ', notifier.driverCodeCtrl),
          _buildTextField('ชื่อคนขับ', notifier.driverNameCtrl),
        ] else ...[
          _viewField('รหัสคนขับ',
              state.selectedVehicle?.driverCode ?? ''),
          _viewField('ชื่อคนขับ',
              state.selectedVehicle?.driverName ?? ''),
        ],
      ],
    );
  }
}

// ============================================================
// Section 4: กฎหมาย & ประกัน
// ============================================================

class _RegulatorySection extends StatelessWidget {
  final VehicleState state;
  final VehicleNotifier notifier;
  final bool isEditable;

  const _RegulatorySection({
    required this.state,
    required this.notifier,
    required this.isEditable,
  });

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: 'กฎหมาย & ประกัน',
      icon: Icons.gavel_outlined,
      children: [
        if (isEditable) ...[
          _DatePickerField(
            label: 'วันหมดอายุประกันภัย',
            value: state.insuranceExpiry,
            onChanged: notifier.setInsuranceExpiry,
            context: context,
          ),
          _DatePickerField(
            label: 'วันหมดอายุทะเบียน/ภาษี',
            value: state.registrationExpiry,
            onChanged: notifier.setRegistrationExpiry,
            context: context,
          ),
        ] else ...[
          ExpiryBadge(
            label: 'ประกันภัย',
            expiryDate: state.selectedVehicle?.insuranceExpiry,
          ),
          const SizedBox(height: 8),
          ExpiryBadge(
            label: 'ทะเบียน/ภาษี',
            expiryDate: state.selectedVehicle?.registrationExpiry,
          ),
        ],
      ],
    );
  }
}

// ============================================================
// Section 5: หมายเหตุ
// ============================================================

class _NotesSection extends StatelessWidget {
  final VehicleNotifier notifier;
  final bool isEditable;
  final String notes;

  const _NotesSection({
    required this.notifier,
    required this.isEditable,
    required this.notes,
  });

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: 'หมายเหตุ',
      icon: Icons.note_outlined,
      children: [
        if (isEditable)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: TextFormField(
              controller: notifier.notesCtrl,
              maxLines: 3,
              decoration: const InputDecoration(
                labelText: 'หมายเหตุ',
                alignLabelWithHint: true,
              ),
            ),
          )
        else
          _viewField('หมายเหตุ', notes),
      ],
    );
  }
}

// ============================================================
// DatePicker Field
// ============================================================

/// DatePicker field สำหรับวันหมดอายุ
class _DatePickerField extends StatelessWidget {
  final String label;
  final DateTime? value;
  final ValueChanged<DateTime?> onChanged;
  final BuildContext context;

  const _DatePickerField({
    required this.label,
    required this.value,
    required this.onChanged,
    required this.context,
  });

  @override
  Widget build(BuildContext context) {
    final dateText = value != null
        ? dateTimeBuddhist(value!, format: DateTimeFormatEnum.date)
        : '';
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: TextFormField(
        readOnly: true,
        decoration: InputDecoration(
          labelText: label,
          hintText: 'เลือกวันที่',
          suffixIcon: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (value != null)
                IconButton(
                  icon: Icon(
                    Icons.clear,
                    size: 18,
                    color: AppTheme.earthBrown.withValues(alpha: 0.4),
                  ),
                  onPressed: () => onChanged(null),
                ),
              IconButton(
                icon: const Icon(Icons.calendar_today, size: 18),
                onPressed: _pickDate,
              ),
            ],
          ),
        ),
        controller: TextEditingController(text: dateText),
        onTap: _pickDate,
      ),
    );
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: value ?? DateTime.now(),
      firstDate: DateTime(2020),
      lastDate: DateTime(2040),
      builder: (ctx, child) {
        return Theme(
          data: Theme.of(ctx).copyWith(
            colorScheme: Theme.of(ctx).colorScheme.copyWith(
                  primary: AppTheme.terracotta,
                ),
          ),
          child: child!,
        );
      },
    );
    if (picked != null) {
      onChanged(picked);
    }
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

  const _SectionCard({
    required this.title,
    required this.icon,
    this.initiallyExpanded = false,
    required this.children,
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
        title: Text(
          title,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppTheme.earthBrown,
          ),
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
