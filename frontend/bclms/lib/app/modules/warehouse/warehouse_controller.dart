import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/modules/warehouse/warehouse_model.dart';
import 'package:bclms/app/providers/service_providers.dart';

class WarehouseState {
  final List<WarehouseModel> warehouses;
  final bool isLoading;
  final String searchQuery;
  final ScreenMode screenMode;
  final WarehouseModel? selectedWarehouse;
  final LocationModel? selectedLocation;
  final ShelfModel? selectedShelf;
  final Set<String> selectedIds;
  final bool isSaving;
  final String? successMessage;
  final String? errorMessage;

  const WarehouseState({
    this.warehouses = const [],
    this.isLoading = false,
    this.searchQuery = '',
    this.screenMode = ScreenMode.list,
    this.selectedWarehouse,
    this.selectedLocation,
    this.selectedShelf,
    this.selectedIds = const {},
    this.isSaving = false,
    this.successMessage,
    this.errorMessage,
  });

  WarehouseState copyWith({
    List<WarehouseModel>? warehouses,
    bool? isLoading,
    String? searchQuery,
    ScreenMode? screenMode,
    WarehouseModel? Function()? selectedWarehouse,
    LocationModel? Function()? selectedLocation,
    ShelfModel? Function()? selectedShelf,
    Set<String>? selectedIds,
    bool? isSaving,
    String? successMessage,
    String? errorMessage,
  }) {
    return WarehouseState(
      warehouses: warehouses ?? this.warehouses,
      isLoading: isLoading ?? this.isLoading,
      searchQuery: searchQuery ?? this.searchQuery,
      screenMode: screenMode ?? this.screenMode,
      selectedWarehouse: selectedWarehouse != null
          ? selectedWarehouse()
          : this.selectedWarehouse,
      selectedLocation: selectedLocation != null
          ? selectedLocation()
          : this.selectedLocation,
      selectedShelf:
          selectedShelf != null ? selectedShelf() : this.selectedShelf,
      selectedIds: selectedIds ?? this.selectedIds,
      isSaving: isSaving ?? this.isSaving,
      successMessage: successMessage,
      errorMessage: errorMessage,
    );
  }

  /// กรองรายการคลังตาม searchQuery
  List<WarehouseModel> get filteredWarehouses {
    if (searchQuery.isEmpty) return warehouses;
    final q = searchQuery.toLowerCase();
    return warehouses.where((w) {
      return w.code.toLowerCase().contains(q) ||
          w.displayName.toLowerCase().contains(q);
    }).toList();
  }
}

/// Controller สำหรับหน้าจัดการคลังสินค้า
class WarehouseNotifier extends Notifier<WarehouseState> {
  final formKey = GlobalKey<FormState>();

  // Form controllers — Warehouse
  late final TextEditingController codeCtrl;
  late final TextEditingController nameThCtrl;
  late final TextEditingController nameEnCtrl;

  // Form controllers — Location
  late final TextEditingController locCodeCtrl;
  late final TextEditingController locNameThCtrl;
  late final TextEditingController locNameEnCtrl;

  // Form controllers — Shelf
  late final TextEditingController shelfCodeCtrl;
  late final TextEditingController shelfNameCtrl;
  late final TextEditingController shelfMinCtrl;
  late final TextEditingController shelfMaxCtrl;

  @override
  WarehouseState build() {
    codeCtrl = TextEditingController();
    nameThCtrl = TextEditingController();
    nameEnCtrl = TextEditingController();
    locCodeCtrl = TextEditingController();
    locNameThCtrl = TextEditingController();
    locNameEnCtrl = TextEditingController();
    shelfCodeCtrl = TextEditingController();
    shelfNameCtrl = TextEditingController();
    shelfMinCtrl = TextEditingController();
    shelfMaxCtrl = TextEditingController();

    ref.onDispose(() {
      codeCtrl.dispose();
      nameThCtrl.dispose();
      nameEnCtrl.dispose();
      locCodeCtrl.dispose();
      locNameThCtrl.dispose();
      locNameEnCtrl.dispose();
      shelfCodeCtrl.dispose();
      shelfNameCtrl.dispose();
      shelfMinCtrl.dispose();
      shelfMaxCtrl.dispose();
    });

    Future.microtask(() => loadWarehouses());
    return const WarehouseState();
  }

  // ============================================================
  // โหลดข้อมูล
  // ============================================================

  Future<void> loadWarehouses() async {
    state = state.copyWith(isLoading: true);
    try {
      final service = ref.read(warehouseServiceProvider);
      final result = await service.list(
        search: state.searchQuery.isEmpty ? null : state.searchQuery,
      );
      state = state.copyWith(warehouses: result, isLoading: false);
      AppLogger.info('[Warehouse] โหลดรายการคลัง: ${result.length} รายการ');
    } catch (e) {
      AppLogger.error('[Warehouse] โหลดรายการคลังล้มเหลว', error: e);
      state = state.copyWith(isLoading: false);
    }
  }

  Future<void> refreshList() async {
    await loadWarehouses();
  }

  /// โหลดข้อมูลคลังเต็มรูปแบบ (รวม location/shelf)
  Future<void> _reloadSelectedWarehouse() async {
    final wh = state.selectedWarehouse;
    if (wh == null) return;
    final service = ref.read(warehouseServiceProvider);
    final full = await service.getById(wh.guidFixed);
    if (full != null) {
      state = state.copyWith(selectedWarehouse: () => full);
    }
  }

  void setSearchQuery(String query) {
    state = state.copyWith(searchQuery: query);
    loadWarehouses();
  }

  // ============================================================
  // Navigation / Mode
  // ============================================================

  void selectWarehouse(WarehouseModel warehouse) async {
    state = state.copyWith(
      selectedWarehouse: () => warehouse,
      selectedLocation: () => null,
      selectedShelf: () => null,
      screenMode: ScreenMode.view,
    );
    // โหลดข้อมูลเต็ม (รวม nested locations)
    await _reloadSelectedWarehouse();
    AppLogger.debug('[Warehouse] เลือกคลัง: ${warehouse.code}');
  }

  void selectLocation(LocationModel location) {
    state = state.copyWith(
      selectedLocation: () => location,
      selectedShelf: () => null,
    );
    AppLogger.debug('[Warehouse] เลือก Location: ${location.code}');
  }

  void selectShelf(ShelfModel shelf) {
    state = state.copyWith(selectedShelf: () => shelf);
    AppLogger.debug('[Warehouse] เลือก Shelf: ${shelf.code}');
  }

  void startAdd() {
    _clearWarehouseForm();
    state = state.copyWith(
      screenMode: ScreenMode.add,
      selectedWarehouse: () => null,
      selectedLocation: () => null,
      selectedShelf: () => null,
    );
    AppLogger.debug('[Warehouse] เปิดโหมดเพิ่มคลัง');
  }

  void startEdit() {
    final wh = state.selectedWarehouse;
    if (wh == null) return;
    _populateWarehouseForm(wh);
    state = state.copyWith(screenMode: ScreenMode.edit);
    AppLogger.debug('[Warehouse] เปิดโหมดแก้ไขคลัง: ${wh.code}');
  }

  void cancelForm() {
    if (state.screenMode == ScreenMode.edit) {
      state = state.copyWith(screenMode: ScreenMode.view);
    } else {
      state = state.copyWith(
        screenMode: ScreenMode.list,
        selectedWarehouse: () => null,
        selectedLocation: () => null,
        selectedShelf: () => null,
      );
    }
  }

  void backToList() {
    state = state.copyWith(
      screenMode: ScreenMode.list,
      selectedWarehouse: () => null,
      selectedLocation: () => null,
      selectedShelf: () => null,
    );
  }

  // ============================================================
  // Warehouse CRUD
  // ============================================================

  Future<bool> saveWarehouse() async {
    if (!formKey.currentState!.validate()) return false;

    state = state.copyWith(isSaving: true);
    try {
      final service = ref.read(warehouseServiceProvider);
      final warehouse = _buildWarehouseFromForm();
      final isAdd = state.screenMode == ScreenMode.add;

      bool success;
      if (isAdd) {
        success = await service.create(warehouse);
      } else {
        success = await service.update(
          state.selectedWarehouse!.guidFixed,
          warehouse,
        );
      }

      if (success) {
        AppLogger.success(
            '[Warehouse] บันทึกคลังสำเร็จ: ${warehouse.code}');
        await loadWarehouses();
        state = state.copyWith(
          isSaving: false,
          screenMode: ScreenMode.list,
          selectedWarehouse: () => null,
          successMessage:
              isAdd ? 'เพิ่มคลังสินค้าสำเร็จ' : 'อัพเดทคลังสินค้าสำเร็จ',
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
      AppLogger.error('[Warehouse] บันทึกคลังล้มเหลว', error: e);
      state = state.copyWith(
        isSaving: false,
        errorMessage: 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้',
      );
      return false;
    }
  }

  Future<bool> deleteWarehouse() async {
    final wh = state.selectedWarehouse;
    if (wh == null) return false;

    final service = ref.read(warehouseServiceProvider);
    final success = await service.delete(wh.guidFixed);
    if (success) {
      AppLogger.success('[Warehouse] ลบคลังสำเร็จ: ${wh.code}');
      await loadWarehouses();
      state = state.copyWith(
        screenMode: ScreenMode.list,
        selectedWarehouse: () => null,
        successMessage: 'ลบคลังสินค้าสำเร็จ',
      );
      return true;
    }
    state = state.copyWith(errorMessage: 'ไม่สามารถลบคลังสินค้าได้');
    return false;
  }

  // ============================================================
  // Location CRUD
  // ============================================================

  Future<bool> saveLocation() async {
    final wh = state.selectedWarehouse;
    if (wh == null) return false;
    if (locCodeCtrl.text.trim().isEmpty) return false;

    state = state.copyWith(isSaving: true);
    try {
      final service = ref.read(warehouseServiceProvider);
      final location = _buildLocationFromForm();
      final isNew = state.selectedLocation == null;

      bool success;
      if (isNew) {
        success = await service.createLocation(wh.code, location);
      } else {
        success = await service.updateLocation(
            wh.code, state.selectedLocation!.code, location);
      }

      if (success) {
        await _reloadSelectedWarehouse();
        state = state.copyWith(
          isSaving: false,
          selectedLocation: () => null,
          successMessage:
              isNew ? 'เพิ่มที่เก็บสำเร็จ' : 'อัพเดทที่เก็บสำเร็จ',
        );
        return true;
      } else {
        state = state.copyWith(
          isSaving: false,
          errorMessage: 'ไม่สามารถบันทึกที่เก็บได้',
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isSaving: false,
        errorMessage: 'เกิดข้อผิดพลาด',
      );
      return false;
    }
  }

  Future<bool> deleteLocation(String locationCode) async {
    final wh = state.selectedWarehouse;
    if (wh == null) return false;

    final service = ref.read(warehouseServiceProvider);
    final success = await service.deleteLocations(wh.code, [locationCode]);
    if (success) {
      await _reloadSelectedWarehouse();
      state = state.copyWith(
        selectedLocation: () => null,
        successMessage: 'ลบที่เก็บสำเร็จ',
      );
      return true;
    }
    state = state.copyWith(errorMessage: 'ไม่สามารถลบที่เก็บได้');
    return false;
  }

  // ============================================================
  // Shelf CRUD
  // ============================================================

  Future<bool> saveShelf() async {
    final wh = state.selectedWarehouse;
    final loc = state.selectedLocation;
    if (wh == null || loc == null) return false;
    if (shelfCodeCtrl.text.trim().isEmpty) return false;

    state = state.copyWith(isSaving: true);
    try {
      final service = ref.read(warehouseServiceProvider);
      final shelf = _buildShelfFromForm();
      final isNew = state.selectedShelf == null;

      bool success;
      if (isNew) {
        success = await service.createShelf(wh.code, loc.code, shelf);
      } else {
        success = await service.updateShelf(
            wh.code, loc.code, state.selectedShelf!.code, shelf);
      }

      if (success) {
        await _reloadSelectedWarehouse();
        // อัพเดท selectedLocation ด้วย
        final updatedWh = state.selectedWarehouse;
        if (updatedWh != null) {
          final updatedLoc = updatedWh.locations
              .where((l) => l.code == loc.code)
              .firstOrNull;
          state = state.copyWith(
            isSaving: false,
            selectedLocation: () => updatedLoc,
            selectedShelf: () => null,
            successMessage:
                isNew ? 'เพิ่มชั้นวางสำเร็จ' : 'อัพเดทชั้นวางสำเร็จ',
          );
        }
        return true;
      } else {
        state = state.copyWith(
          isSaving: false,
          errorMessage: 'ไม่สามารถบันทึกชั้นวางได้',
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isSaving: false,
        errorMessage: 'เกิดข้อผิดพลาด',
      );
      return false;
    }
  }

  Future<bool> deleteShelf(String shelfCode) async {
    final wh = state.selectedWarehouse;
    final loc = state.selectedLocation;
    if (wh == null || loc == null) return false;

    final service = ref.read(warehouseServiceProvider);
    final success =
        await service.deleteShelves(wh.code, loc.code, [shelfCode]);
    if (success) {
      await _reloadSelectedWarehouse();
      final updatedWh = state.selectedWarehouse;
      if (updatedWh != null) {
        final updatedLoc = updatedWh.locations
            .where((l) => l.code == loc.code)
            .firstOrNull;
        state = state.copyWith(
          selectedLocation: () => updatedLoc,
          selectedShelf: () => null,
          successMessage: 'ลบชั้นวางสำเร็จ',
        );
      }
      return true;
    }
    state = state.copyWith(errorMessage: 'ไม่สามารถลบชั้นวางได้');
    return false;
  }

  // ============================================================
  // MultiSelect
  // ============================================================

  void toggleMultiSelect() {
    if (state.screenMode == ScreenMode.multiSelect) {
      state = state.copyWith(screenMode: ScreenMode.list, selectedIds: {});
    } else {
      state =
          state.copyWith(screenMode: ScreenMode.multiSelect, selectedIds: {});
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

    final service = ref.read(warehouseServiceProvider);
    final success =
        await service.deleteMultiple(state.selectedIds.toList());
    if (success) {
      AppLogger.success(
          '[Warehouse] ลบคลังหลายรายการสำเร็จ: ${state.selectedIds.length} รายการ');
      state = state.copyWith(
        selectedIds: {},
        successMessage: 'ลบคลังสินค้าสำเร็จ',
      );
      await loadWarehouses();
      state = state.copyWith(screenMode: ScreenMode.list);
      return true;
    }
    return false;
  }

  // ============================================================
  // Form Helpers — Warehouse
  // ============================================================

  void _clearWarehouseForm() {
    codeCtrl.clear();
    nameThCtrl.clear();
    nameEnCtrl.clear();
  }

  void _populateWarehouseForm(WarehouseModel wh) {
    codeCtrl.text = wh.code;
    nameThCtrl.text = wh.names
            .where((n) => n.code == 'th')
            .firstOrNull
            ?.name ??
        '';
    nameEnCtrl.text = wh.names
            .where((n) => n.code == 'en')
            .firstOrNull
            ?.name ??
        '';
  }

  WarehouseModel _buildWarehouseFromForm() {
    final names = <NameModel>[];
    if (nameThCtrl.text.trim().isNotEmpty) {
      names.add(NameModel(code: 'th', name: nameThCtrl.text.trim()));
    }
    if (nameEnCtrl.text.trim().isNotEmpty) {
      names.add(NameModel(code: 'en', name: nameEnCtrl.text.trim()));
    }
    return WarehouseModel(
      guidFixed: state.selectedWarehouse?.guidFixed ?? '',
      code: codeCtrl.text.trim(),
      names: names,
    );
  }

  // ============================================================
  // Form Helpers — Location
  // ============================================================

  void clearLocationForm() {
    locCodeCtrl.clear();
    locNameThCtrl.clear();
    locNameEnCtrl.clear();
    state = state.copyWith(selectedLocation: () => null);
  }

  void populateLocationForm(LocationModel loc) {
    locCodeCtrl.text = loc.code;
    locNameThCtrl.text = loc.names
            .where((n) => n.code == 'th')
            .firstOrNull
            ?.name ??
        '';
    locNameEnCtrl.text = loc.names
            .where((n) => n.code == 'en')
            .firstOrNull
            ?.name ??
        '';
  }

  LocationModel _buildLocationFromForm() {
    final names = <NameModel>[];
    if (locNameThCtrl.text.trim().isNotEmpty) {
      names.add(NameModel(code: 'th', name: locNameThCtrl.text.trim()));
    }
    if (locNameEnCtrl.text.trim().isNotEmpty) {
      names.add(NameModel(code: 'en', name: locNameEnCtrl.text.trim()));
    }
    return LocationModel(
      code: locCodeCtrl.text.trim(),
      names: names,
    );
  }

  // ============================================================
  // Form Helpers — Shelf
  // ============================================================

  void clearShelfForm() {
    shelfCodeCtrl.clear();
    shelfNameCtrl.clear();
    shelfMinCtrl.clear();
    shelfMaxCtrl.clear();
    state = state.copyWith(selectedShelf: () => null);
  }

  void populateShelfForm(ShelfModel shelf) {
    shelfCodeCtrl.text = shelf.code;
    shelfNameCtrl.text = shelf.name;
    shelfMinCtrl.text = shelf.min > 0 ? shelf.min.toString() : '';
    shelfMaxCtrl.text = shelf.max > 0 ? shelf.max.toString() : '';
  }

  ShelfModel _buildShelfFromForm() {
    return ShelfModel(
      code: shelfCodeCtrl.text.trim(),
      name: shelfNameCtrl.text.trim(),
      min: int.tryParse(shelfMinCtrl.text.trim()) ?? 0,
      max: int.tryParse(shelfMaxCtrl.text.trim()) ?? 0,
    );
  }
}

final warehouseProvider =
    NotifierProvider<WarehouseNotifier, WarehouseState>(
  WarehouseNotifier.new,
);
