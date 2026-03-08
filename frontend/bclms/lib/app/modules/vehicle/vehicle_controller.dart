import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/modules/vehicle/vehicle_model.dart';
import 'package:bclms/app/providers/service_providers.dart';

class VehicleState {
  final List<VehicleModel> vehicles;
  final bool isLoading;
  final String searchQuery;
  final int? statusFilter;
  final ScreenMode screenMode;
  final VehicleModel? selectedVehicle;
  final Set<String> selectedIds;
  final bool isSaving;
  final String vehicleTypeValue;
  final int statusValue;
  final String fuelTypeValue;
  final DateTime? insuranceExpiry;
  final DateTime? registrationExpiry;
  final String? successMessage;
  final String? errorMessage;

  const VehicleState({
    this.vehicles = const [],
    this.isLoading = false,
    this.searchQuery = '',
    this.statusFilter,
    this.screenMode = ScreenMode.list,
    this.selectedVehicle,
    this.selectedIds = const {},
    this.isSaving = false,
    this.vehicleTypeValue = '',
    this.statusValue = 1,
    this.fuelTypeValue = '',
    this.insuranceExpiry,
    this.registrationExpiry,
    this.successMessage,
    this.errorMessage,
  });

  VehicleState copyWith({
    List<VehicleModel>? vehicles,
    bool? isLoading,
    String? searchQuery,
    int? Function()? statusFilter,
    ScreenMode? screenMode,
    VehicleModel? Function()? selectedVehicle,
    Set<String>? selectedIds,
    bool? isSaving,
    String? vehicleTypeValue,
    int? statusValue,
    String? fuelTypeValue,
    DateTime? Function()? insuranceExpiry,
    DateTime? Function()? registrationExpiry,
    String? successMessage,
    String? errorMessage,
  }) {
    return VehicleState(
      vehicles: vehicles ?? this.vehicles,
      isLoading: isLoading ?? this.isLoading,
      searchQuery: searchQuery ?? this.searchQuery,
      statusFilter: statusFilter != null ? statusFilter() : this.statusFilter,
      screenMode: screenMode ?? this.screenMode,
      selectedVehicle: selectedVehicle != null ? selectedVehicle() : this.selectedVehicle,
      selectedIds: selectedIds ?? this.selectedIds,
      isSaving: isSaving ?? this.isSaving,
      vehicleTypeValue: vehicleTypeValue ?? this.vehicleTypeValue,
      statusValue: statusValue ?? this.statusValue,
      fuelTypeValue: fuelTypeValue ?? this.fuelTypeValue,
      insuranceExpiry: insuranceExpiry != null ? insuranceExpiry() : this.insuranceExpiry,
      registrationExpiry: registrationExpiry != null ? registrationExpiry() : this.registrationExpiry,
      successMessage: successMessage,
      errorMessage: errorMessage,
    );
  }

  List<VehicleModel> get filteredVehicles {
    if (searchQuery.isEmpty) return vehicles;
    final q = searchQuery.toLowerCase();
    return vehicles.where((v) {
      return v.code.toLowerCase().contains(q) ||
          v.licensePlate.toLowerCase().contains(q) ||
          v.displayName.toLowerCase().contains(q) ||
          v.brand.toLowerCase().contains(q);
    }).toList();
  }
}

/// Controller สำหรับหน้าจัดการยานพาหนะ — ตาม Backend API schema
class VehicleNotifier extends Notifier<VehicleState> {
  final formKey = GlobalKey<FormState>();

  // Form controllers — ข้อมูลทั่วไป
  late final TextEditingController codeCtrl;
  late final TextEditingController nameThCtrl;
  late final TextEditingController licensePlateCtrl;
  late final TextEditingController provinceCtrl;
  late final TextEditingController brandCtrl;
  late final TextEditingController modelCtrl;
  late final TextEditingController yearCtrl;
  late final TextEditingController colorCtrl;

  // Form controllers — ข้อมูลตัวรถ
  late final TextEditingController capacityWeightCtrl;
  late final TextEditingController capacityVolumeCtrl;

  // Form controllers — คนขับ
  late final TextEditingController driverCodeCtrl;
  late final TextEditingController driverNameCtrl;

  // Form controllers — อื่นๆ
  late final TextEditingController notesCtrl;

  @override
  VehicleState build() {
    codeCtrl = TextEditingController();
    nameThCtrl = TextEditingController();
    licensePlateCtrl = TextEditingController();
    provinceCtrl = TextEditingController();
    brandCtrl = TextEditingController();
    modelCtrl = TextEditingController();
    yearCtrl = TextEditingController();
    colorCtrl = TextEditingController();
    capacityWeightCtrl = TextEditingController();
    capacityVolumeCtrl = TextEditingController();
    driverCodeCtrl = TextEditingController();
    driverNameCtrl = TextEditingController();
    notesCtrl = TextEditingController();

    ref.onDispose(() {
      codeCtrl.dispose();
      nameThCtrl.dispose();
      licensePlateCtrl.dispose();
      provinceCtrl.dispose();
      brandCtrl.dispose();
      modelCtrl.dispose();
      yearCtrl.dispose();
      colorCtrl.dispose();
      capacityWeightCtrl.dispose();
      capacityVolumeCtrl.dispose();
      driverCodeCtrl.dispose();
      driverNameCtrl.dispose();
      notesCtrl.dispose();
    });

    loadVehicles();
    return const VehicleState();
  }

  // === โหลดข้อมูล ===

  Future<void> loadVehicles() async {
    state = state.copyWith(isLoading: true);
    try {
      final service = ref.read(vehicleServiceProvider);
      final result = await service.list(
        search: state.searchQuery.isEmpty ? null : state.searchQuery,
        status: state.statusFilter,
      );
      state = state.copyWith(vehicles: result, isLoading: false);
      AppLogger.info('[Vehicle] โหลดรายการยานพาหนะ: ${result.length} คัน');
    } catch (e) {
      AppLogger.error('[Vehicle] โหลดรายการยานพาหนะล้มเหลว', error: e);
      state = state.copyWith(isLoading: false);
    }
  }

  Future<void> refreshList() async {
    await loadVehicles();
  }

  void setStatusFilter(int? status) {
    state = state.copyWith(statusFilter: () => status);
    loadVehicles();
  }

  void setSearchQuery(String query) {
    state = state.copyWith(searchQuery: query);
    loadVehicles();
  }

  // === Navigation / Mode ===

  void selectVehicle(VehicleModel vehicle) {
    state = state.copyWith(
      selectedVehicle: () => vehicle,
      screenMode: ScreenMode.view,
    );
    AppLogger.debug('[Vehicle] เลือกยานพาหนะ: ${vehicle.code}');
  }

  void startAdd() {
    _clearForm();
    state = state.copyWith(
      statusValue: 1,
      screenMode: ScreenMode.add,
    );
    AppLogger.debug('[Vehicle] เปิดโหมดเพิ่มยานพาหนะ');
  }

  void startEdit() {
    final v = state.selectedVehicle;
    if (v == null) return;
    _populateForm(v);
    state = state.copyWith(screenMode: ScreenMode.edit);
    AppLogger.debug('[Vehicle] เปิดโหมดแก้ไขยานพาหนะ: ${v.code}');
  }

  void cancelForm() {
    if (state.screenMode == ScreenMode.edit) {
      state = state.copyWith(screenMode: ScreenMode.view);
    } else {
      state = state.copyWith(
        screenMode: ScreenMode.list,
        selectedVehicle: () => null,
      );
    }
  }

  void backToList() {
    state = state.copyWith(
      screenMode: ScreenMode.list,
      selectedVehicle: () => null,
    );
  }

  // === CRUD ===

  Future<bool> saveVehicle() async {
    if (!formKey.currentState!.validate()) return false;

    state = state.copyWith(isSaving: true);
    try {
      final service = ref.read(vehicleServiceProvider);
      final vehicle = _buildFromForm();
      final isAdd = state.screenMode == ScreenMode.add;

      bool success;
      if (isAdd) {
        success = await service.create(vehicle);
      } else {
        success = await service.update(
          state.selectedVehicle!.guidFixed,
          vehicle,
        );
      }

      if (success) {
        AppLogger.success('[Vehicle] บันทึกยานพาหนะสำเร็จ: ${vehicle.code}');
        await loadVehicles();
        state = state.copyWith(
          isSaving: false,
          screenMode: ScreenMode.list,
          selectedVehicle: () => null,
          successMessage: isAdd ? 'เพิ่มยานพาหนะสำเร็จ' : 'อัพเดทยานพาหนะสำเร็จ',
        );
        return true;
      } else {
        state = state.copyWith(
          isSaving: false,
          errorMessage: 'ไม่สามารถบันทึกข้อมูลได้',
        );
        return false;
      }
    } catch (e) {
      AppLogger.error('[Vehicle] บันทึกยานพาหนะล้มเหลว', error: e);
      state = state.copyWith(
        isSaving: false,
        errorMessage: 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้',
      );
      return false;
    }
  }

  Future<bool> deleteVehicle() async {
    final v = state.selectedVehicle;
    if (v == null) return false;

    final service = ref.read(vehicleServiceProvider);
    final success = await service.delete(v.guidFixed);
    if (success) {
      AppLogger.success('[Vehicle] ลบยานพาหนะสำเร็จ: ${v.code}');
      await loadVehicles();
      state = state.copyWith(
        screenMode: ScreenMode.list,
        selectedVehicle: () => null,
      );
      return true;
    }
    return false;
  }

  // === MultiSelect ===

  void toggleMultiSelect() {
    if (state.screenMode == ScreenMode.multiSelect) {
      state = state.copyWith(screenMode: ScreenMode.list, selectedIds: {});
    } else {
      state = state.copyWith(screenMode: ScreenMode.multiSelect, selectedIds: {});
    }
  }

  void toggleItemSelect(String guidFixed) {
    final newIds = Set<String>.from(state.selectedIds);
    if (newIds.contains(guidFixed)) {
      newIds.remove(guidFixed);
    } else {
      newIds.add(guidFixed);
    }
    state = state.copyWith(selectedIds: newIds);
  }

  Future<bool> deleteSelected() async {
    if (state.selectedIds.isEmpty) return false;

    final service = ref.read(vehicleServiceProvider);
    final success = await service.deleteMultiple(state.selectedIds.toList());
    if (success) {
      AppLogger.success(
          '[Vehicle] ลบยานพาหนะหลายรายการสำเร็จ: ${state.selectedIds.length} คัน');
      state = state.copyWith(selectedIds: {});
      await loadVehicles();
      state = state.copyWith(screenMode: ScreenMode.list);
      return true;
    }
    return false;
  }

  // === Form Helpers ===

  void setVehicleType(String value) {
    state = state.copyWith(vehicleTypeValue: value);
  }

  void setStatus(int value) {
    state = state.copyWith(statusValue: value);
  }

  void setFuelType(String value) {
    state = state.copyWith(fuelTypeValue: value);
  }

  void setInsuranceExpiry(DateTime? value) {
    state = state.copyWith(insuranceExpiry: () => value);
  }

  void setRegistrationExpiry(DateTime? value) {
    state = state.copyWith(registrationExpiry: () => value);
  }

  void _clearForm() {
    codeCtrl.clear();
    nameThCtrl.clear();
    licensePlateCtrl.clear();
    provinceCtrl.clear();
    brandCtrl.clear();
    modelCtrl.clear();
    yearCtrl.clear();
    colorCtrl.clear();
    capacityWeightCtrl.clear();
    capacityVolumeCtrl.clear();
    driverCodeCtrl.clear();
    driverNameCtrl.clear();
    notesCtrl.clear();

    state = state.copyWith(
      vehicleTypeValue: '',
      statusValue: 1,
      fuelTypeValue: '',
      insuranceExpiry: () => null,
      registrationExpiry: () => null,
    );
  }

  void _populateForm(VehicleModel v) {
    codeCtrl.text = v.code;
    nameThCtrl.text = v.displayName;
    licensePlateCtrl.text = v.licensePlate;
    provinceCtrl.text = v.province;
    brandCtrl.text = v.brand;
    modelCtrl.text = v.model;
    yearCtrl.text = v.year > 0 ? v.year.toString() : '';
    colorCtrl.text = v.color;
    capacityWeightCtrl.text =
        v.capacityWeight != null ? v.capacityWeight.toString() : '';
    capacityVolumeCtrl.text =
        v.capacityVolume != null ? v.capacityVolume.toString() : '';
    driverCodeCtrl.text = v.driverCode;
    driverNameCtrl.text = v.driverName;
    notesCtrl.text = v.notes;

    state = state.copyWith(
      vehicleTypeValue: v.vehicleType,
      statusValue: v.status,
      fuelTypeValue: v.fuelType,
      insuranceExpiry: () => v.insuranceExpiry,
      registrationExpiry: () => v.registrationExpiry,
    );
  }

  VehicleModel _buildFromForm() {
    return VehicleModel(
      guidFixed: state.selectedVehicle?.guidFixed ?? '',
      code: codeCtrl.text.trim(),
      names: [VehicleName(code: 'th', name: nameThCtrl.text.trim())],
      vehicleType: state.vehicleTypeValue,
      licensePlate: licensePlateCtrl.text.trim(),
      province: provinceCtrl.text.trim(),
      brand: brandCtrl.text.trim(),
      model: modelCtrl.text.trim(),
      year: int.tryParse(yearCtrl.text.trim()) ?? 0,
      color: colorCtrl.text.trim(),
      status: state.statusValue,
      fuelType: state.fuelTypeValue,
      capacityWeight: double.tryParse(capacityWeightCtrl.text.trim()),
      capacityVolume: double.tryParse(capacityVolumeCtrl.text.trim()),
      driverCode: driverCodeCtrl.text.trim(),
      driverName: driverNameCtrl.text.trim(),
      insuranceExpiry: state.insuranceExpiry,
      registrationExpiry: state.registrationExpiry,
      notes: notesCtrl.text.trim(),
    );
  }
}

final vehicleProvider = NotifierProvider<VehicleNotifier, VehicleState>(
  VehicleNotifier.new,
);
