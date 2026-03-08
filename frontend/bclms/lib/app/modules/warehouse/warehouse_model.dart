// โมเดลข้อมูลคลังสินค้า — ตาม Backend API schema
// Warehouse → Location → Shelf → ShelfProduct (nested hierarchy)

/// ชื่อหลายภาษา (ใช้ร่วมกับ Warehouse, Location, ShelfProduct)
class NameModel {
  final String code;
  final String name;

  const NameModel({required this.code, required this.name});

  factory NameModel.fromJson(Map<String, dynamic> json) {
    return NameModel(
      code: json['code'] ?? '',
      name: json['name'] ?? '',
    );
  }

  Map<String, dynamic> toJson() => {'code': code, 'name': name};
}

/// สินค้าบนชั้นวาง
class ShelfProduct {
  final String guidFixed;
  final String barcode;
  final String unitCode;
  final List<NameModel> unitNames;
  final List<NameModel> names;

  const ShelfProduct({
    this.guidFixed = '',
    this.barcode = '',
    this.unitCode = '',
    this.unitNames = const [],
    this.names = const [],
  });

  String get displayName {
    if (names.isEmpty) return barcode;
    final th = names.where((n) => n.code == 'th').firstOrNull;
    return th?.name ?? names.first.name;
  }

  String get displayUnit {
    if (unitNames.isEmpty) return unitCode;
    final th = unitNames.where((n) => n.code == 'th').firstOrNull;
    return th?.name ?? unitNames.first.name;
  }

  factory ShelfProduct.fromJson(Map<String, dynamic> json) {
    return ShelfProduct(
      guidFixed: json['guidfixed'] ?? '',
      barcode: json['barcode'] ?? '',
      unitCode: json['unitcode'] ?? '',
      unitNames: (json['unitnames'] as List?)
              ?.map((e) => NameModel.fromJson(e))
              .toList() ??
          [],
      names: (json['names'] as List?)
              ?.map((e) => NameModel.fromJson(e))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() => {
        'guidfixed': guidFixed,
        'barcode': barcode,
        'unitcode': unitCode,
        if (unitNames.isNotEmpty)
          'unitnames': unitNames.map((n) => n.toJson()).toList(),
        if (names.isNotEmpty) 'names': names.map((n) => n.toJson()).toList(),
      };
}

/// ชั้นวาง (Shelf/Bin)
class ShelfModel {
  final String code;
  final String name;
  final int min;
  final int max;
  final List<ShelfProduct> productItems;

  const ShelfModel({
    this.code = '',
    this.name = '',
    this.min = 0,
    this.max = 0,
    this.productItems = const [],
  });

  factory ShelfModel.fromJson(Map<String, dynamic> json) {
    return ShelfModel(
      code: json['code'] ?? '',
      name: json['name'] ?? '',
      min: json['min'] ?? 0,
      max: json['max'] ?? 0,
      productItems: (json['productitems'] as List?)
              ?.map((e) => ShelfProduct.fromJson(e))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() => {
        'code': code,
        'name': name,
        'min': min,
        'max': max,
        if (productItems.isNotEmpty)
          'productitems': productItems.map((p) => p.toJson()).toList(),
      };
}

/// ที่เก็บ/โซน (Location)
class LocationModel {
  final String code;
  final List<NameModel> names;
  final List<ShelfModel> shelves;

  const LocationModel({
    this.code = '',
    this.names = const [],
    this.shelves = const [],
  });

  String get displayName {
    if (names.isEmpty) return code;
    final th = names.where((n) => n.code == 'th').firstOrNull;
    return th?.name ?? names.first.name;
  }

  factory LocationModel.fromJson(Map<String, dynamic> json) {
    return LocationModel(
      code: json['code'] ?? '',
      names: (json['names'] as List?)
              ?.map((e) => NameModel.fromJson(e))
              .toList() ??
          [],
      shelves: (json['shelf'] as List?)
              ?.map((e) => ShelfModel.fromJson(e))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() => {
        'code': code,
        'names': names.map((n) => n.toJson()).toList(),
        if (shelves.isNotEmpty)
          'shelf': shelves.map((s) => s.toJson()).toList(),
      };
}

/// คลังสินค้า (Warehouse) — top-level entity
class WarehouseModel {
  final String guidFixed;
  final String code;
  final List<NameModel> names;
  final List<LocationModel> locations;

  // Metadata
  final String? createdBy;
  final DateTime? createdAt;
  final String? updatedBy;
  final DateTime? updatedAt;

  const WarehouseModel({
    this.guidFixed = '',
    required this.code,
    this.names = const [],
    this.locations = const [],
    this.createdBy,
    this.createdAt,
    this.updatedBy,
    this.updatedAt,
  });

  /// ชื่อคลัง (ดึงภาษาไทย)
  String get displayName {
    if (names.isEmpty) return code;
    final th = names.where((n) => n.code == 'th').firstOrNull;
    return th?.name ?? names.first.name;
  }

  /// จำนวนที่เก็บ
  int get locationCount => locations.length;

  /// จำนวนชั้นวางรวม
  int get totalShelfCount =>
      locations.fold(0, (sum, loc) => sum + loc.shelves.length);

  factory WarehouseModel.fromJson(Map<String, dynamic> json) {
    return WarehouseModel(
      guidFixed: json['guidfixed'] ?? json['_id'] ?? '',
      code: json['code'] ?? '',
      names: (json['names'] as List?)
              ?.map((e) => NameModel.fromJson(e))
              .toList() ??
          [],
      locations: (json['location'] as List?)
              ?.map((e) => LocationModel.fromJson(e))
              .toList() ??
          [],
      createdBy: json['createdby'],
      createdAt: _parseDate(json['createdat']),
      updatedBy: json['updatedby'],
      updatedAt: _parseDate(json['updatedat']),
    );
  }

  Map<String, dynamic> toJson() => {
        if (guidFixed.isNotEmpty) 'guidfixed': guidFixed,
        'code': code,
        'names': names.map((n) => n.toJson()).toList(),
        if (locations.isNotEmpty)
          'location': locations.map((l) => l.toJson()).toList(),
      };

  static DateTime? _parseDate(dynamic value) {
    if (value == null) return null;
    if (value is String && value.isNotEmpty) {
      return DateTime.tryParse(value);
    }
    return null;
  }
}

/// โหมดหน้าจอ (ใช้ร่วมใน warehouse module)
enum ScreenMode { list, view, add, edit, multiSelect }
