import 'package:smlaicloud/model/global_model.dart';

/// Data class สำหรับเก็บ state ของคลังสินค้าใน PO Screen
///
/// รวม 8 ตัวแปร warehouse/location ให้เป็น object เดียว
/// เพื่อลดความซับซ้อนใน State ของหน้าจอหลัก
class POWarehouseState {
  String defaultWarehouse;
  List<LanguageDataModel> defaultWarehouseNames;
  String defaultLocation;
  List<LanguageDataModel> defaultLocationNames;
  String defaultToWarehouse;
  List<LanguageDataModel> defaultToWarehouseNames;
  String defaultToLocation;
  List<LanguageDataModel> defaultToLocationNames;

  POWarehouseState({
    this.defaultWarehouse = '',
    this.defaultWarehouseNames = const [],
    this.defaultLocation = '',
    this.defaultLocationNames = const [],
    this.defaultToWarehouse = '',
    this.defaultToWarehouseNames = const [],
    this.defaultToLocation = '',
    this.defaultToLocationNames = const [],
  });

  /// สร้าง state ใหม่จาก warehouse แรกใน list
  factory POWarehouseState.fromFirstWarehouse({
    required String warehouseCode,
    required List<LanguageDataModel> warehouseNames,
    required String locationCode,
    required List<LanguageDataModel> locationNames,
  }) {
    return POWarehouseState(
      defaultWarehouse: warehouseCode,
      defaultWarehouseNames: warehouseNames,
      defaultLocation: locationCode,
      defaultLocationNames: locationNames,
      defaultToWarehouse: warehouseCode,
      defaultToWarehouseNames: warehouseNames,
      defaultToLocation: locationCode,
      defaultToLocationNames: locationNames,
    );
  }

  /// รีเซ็ตเป็นค่าว่าง
  void reset() {
    defaultWarehouse = '';
    defaultWarehouseNames = [];
    defaultLocation = '';
    defaultLocationNames = [];
    defaultToWarehouse = '';
    defaultToWarehouseNames = [];
    defaultToLocation = '';
    defaultToLocationNames = [];
  }

  /// อัพเดทค่าทั้งหมด
  void update({
    required String warehouse,
    required List<LanguageDataModel> warehouseNames,
    required String location,
    required List<LanguageDataModel> locationNames,
    required String toWarehouse,
    required List<LanguageDataModel> toWarehouseNames,
    required String toLocation,
    required List<LanguageDataModel> toLocationNames,
  }) {
    defaultWarehouse = warehouse;
    defaultWarehouseNames = warehouseNames;
    defaultLocation = location;
    defaultLocationNames = locationNames;
    defaultToWarehouse = toWarehouse;
    defaultToWarehouseNames = toWarehouseNames;
    defaultToLocation = toLocation;
    defaultToLocationNames = toLocationNames;
  }

  /// Copy with
  POWarehouseState copyWith({
    String? defaultWarehouse,
    List<LanguageDataModel>? defaultWarehouseNames,
    String? defaultLocation,
    List<LanguageDataModel>? defaultLocationNames,
    String? defaultToWarehouse,
    List<LanguageDataModel>? defaultToWarehouseNames,
    String? defaultToLocation,
    List<LanguageDataModel>? defaultToLocationNames,
  }) {
    return POWarehouseState(
      defaultWarehouse: defaultWarehouse ?? this.defaultWarehouse,
      defaultWarehouseNames: defaultWarehouseNames ?? this.defaultWarehouseNames,
      defaultLocation: defaultLocation ?? this.defaultLocation,
      defaultLocationNames: defaultLocationNames ?? this.defaultLocationNames,
      defaultToWarehouse: defaultToWarehouse ?? this.defaultToWarehouse,
      defaultToWarehouseNames: defaultToWarehouseNames ?? this.defaultToWarehouseNames,
      defaultToLocation: defaultToLocation ?? this.defaultToLocation,
      defaultToLocationNames: defaultToLocationNames ?? this.defaultToLocationNames,
    );
  }
}
