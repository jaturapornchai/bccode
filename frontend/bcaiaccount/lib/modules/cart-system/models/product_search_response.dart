/// Models สำหรับ response จาก Backend Unified Product Search API
/// POST /api/product/search/unified
library;

/// Response หลักจาก API
class ProductSearchResponse {
  final String status;
  final int count;
  final int total; // จำนวนทั้งหมดที่พบ (ก่อน pagination)
  final bool hasMore; // มีข้อมูลเพิ่มหรือไม่ (สำหรับ infinite scroll)
  final List<ProductItem> products;
  final List<String> tokens;

  ProductSearchResponse({
    required this.status,
    required this.count,
    required this.total,
    required this.hasMore,
    required this.products,
    required this.tokens,
  });

  bool get isSuccess => status == 'success';
  bool get hasProducts => products.isNotEmpty;

  factory ProductSearchResponse.fromJson(Map<String, dynamic> json) {
    return ProductSearchResponse(
      status: json['status'] as String? ?? 'error',
      count: json['count'] as int? ?? 0,
      total: json['total'] as int? ?? 0,
      hasMore: json['hasMore'] as bool? ?? false,
      products: (json['products'] as List<dynamic>?)
              ?.map((e) => ProductItem.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      tokens:
          (json['tokens'] as List<dynamic>?)?.map((e) => e as String).toList() ??
              [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'status': status,
      'count': count,
      'total': total,
      'hasMore': hasMore,
      'products': products.map((e) => e.toJson()).toList(),
      'tokens': tokens,
    };
  }
}

/// สินค้าที่ค้นพบ พร้อมหน่วยทั้งหมดและยอดคงเหลือ
class ProductItem {
  final String itemCode;
  final String name0;
  final List<ProductUnit> units;
  final ProductBalance? balance;
  final int score;

  ProductItem({
    required this.itemCode,
    required this.name0,
    required this.units,
    this.balance,
    required this.score,
  });

  /// หน่วยแรก (หน่วยใหญ่สุด) สำหรับแสดงผล
  ProductUnit? get primaryUnit => units.isNotEmpty ? units.first : null;

  /// หน่วยเล็กสุด (ratio = 1)
  ProductUnit? get baseUnit {
    for (final unit in units) {
      if (unit.ratio == 1.0) return unit;
    }
    return units.isNotEmpty ? units.last : null;
  }

  /// ราคาขายปลีก (จากหน่วยแรก)
  double get priceRetail => primaryUnit?.priceRetail ?? 0.0;

  /// ราคาขายส่ง (จากหน่วยแรก)
  double get price1 => primaryUnit?.price1 ?? 0.0;

  factory ProductItem.fromJson(Map<String, dynamic> json) {
    return ProductItem(
      itemCode: json['itemcode'] as String? ?? '',
      name0: json['name0'] as String? ?? '',
      units: (json['units'] as List<dynamic>?)
              ?.map((e) => ProductUnit.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      balance: json['balance'] != null
          ? ProductBalance.fromJson(json['balance'] as Map<String, dynamic>)
          : null,
      score: json['score'] as int? ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'itemcode': itemCode,
      'name0': name0,
      'units': units.map((e) => e.toJson()).toList(),
      'balance': balance?.toJson(),
      'score': score,
    };
  }

  @override
  String toString() {
    return 'ProductItem(itemCode: $itemCode, name: $name0, units: ${units.length}, score: $score)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is ProductItem && other.itemCode == itemCode;
  }

  @override
  int get hashCode => itemCode.hashCode;
}

/// หน่วยสินค้าพร้อมราคา
class ProductUnit {
  final String barcode;
  final String unitCode;
  final String unitName;
  final double unitStand;
  final double unitDivide;
  final double price1;
  final double priceRetail;

  ProductUnit({
    required this.barcode,
    required this.unitCode,
    required this.unitName,
    required this.unitStand,
    required this.unitDivide,
    required this.price1,
    required this.priceRetail,
  });

  /// อัตราส่วนหน่วย (unitStand / unitDivide)
  double get ratio => unitStand / unitDivide;

  /// ชื่อหน่วยสำหรับแสดงผล
  String get displayName => unitName.isNotEmpty ? unitName : unitCode;

  /// ราคาสำหรับแสดงผล (ใช้ราคาขายปลีก ถ้าไม่มีใช้ price1)
  double get displayPrice => priceRetail > 0 ? priceRetail : price1;

  /// แสดงราคาพร้อมหน่วย (ไม่มี symbol - ให้ UI จัดการ)
  String get priceDisplay => '${displayPrice.toStringAsFixed(2)}/$displayName';

  factory ProductUnit.fromJson(Map<String, dynamic> json) {
    // Helper function to parse numeric values
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return ProductUnit(
      barcode: json['barcode'] as String? ?? '',
      unitCode: json['unitcode'] as String? ?? '',
      unitName: json['unitname'] as String? ?? '',
      unitStand: parseDouble(json['unitstand'], 1.0),
      unitDivide: parseDouble(json['unitdivide'], 1.0),
      price1: parseDouble(json['price1'], 0.0),
      priceRetail: parseDouble(json['price_retail'], 0.0),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'barcode': barcode,
      'unitcode': unitCode,
      'unitname': unitName,
      'unitstand': unitStand,
      'unitdivide': unitDivide,
      'price1': price1,
      'price_retail': priceRetail,
    };
  }

  @override
  String toString() {
    return 'ProductUnit(barcode: $barcode, unit: $displayName, ratio: $ratio, price: $displayPrice)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is ProductUnit && other.barcode == barcode;
  }

  @override
  int get hashCode => barcode.hashCode;
}

/// ยอดคงเหลือแยกตาม Location
class LocationBalance {
  final String locationCode;
  final double balanceQty;
  final String balanceWord;

  LocationBalance({
    required this.locationCode,
    required this.balanceQty,
    required this.balanceWord,
  });

  factory LocationBalance.fromJson(Map<String, dynamic> json) {
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return LocationBalance(
      locationCode: json['location_code'] as String? ?? '',
      balanceQty: parseDouble(json['balance_qty'], 0.0),
      balanceWord: json['balance_word'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'location_code': locationCode,
      'balance_qty': balanceQty,
      'balance_word': balanceWord,
    };
  }

  @override
  String toString() {
    return 'LocationBalance(location: $locationCode, qty: $balanceQty, word: $balanceWord)';
  }
}

/// ยอดคงเหลือแยกตาม Warehouse พร้อม Locations
class WarehouseBalance {
  final String warehouseCode;
  final double balanceQty;
  final String balanceWord;
  final List<LocationBalance> locations;

  WarehouseBalance({
    required this.warehouseCode,
    required this.balanceQty,
    required this.balanceWord,
    required this.locations,
  });

  /// มี locations หรือไม่
  bool get hasLocations => locations.isNotEmpty;

  factory WarehouseBalance.fromJson(Map<String, dynamic> json) {
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return WarehouseBalance(
      warehouseCode: json['warehouse_code'] as String? ?? '',
      balanceQty: parseDouble(json['balance_qty'], 0.0),
      balanceWord: json['balance_word'] as String? ?? '',
      locations: (json['locations'] as List<dynamic>?)
              ?.map((e) => LocationBalance.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'warehouse_code': warehouseCode,
      'balance_qty': balanceQty,
      'balance_word': balanceWord,
      'locations': locations.map((e) => e.toJson()).toList(),
    };
  }

  @override
  String toString() {
    return 'WarehouseBalance(wh: $warehouseCode, qty: $balanceQty, locations: ${locations.length})';
  }
}

/// ยอดคงเหลือพร้อม format แบบ multi-level packing และแยกตาม warehouse/location
class ProductBalance {
  final double total;
  final String formatted;
  final List<WarehouseBalance> warehouses;
  final double pendingRecvQty; // ยอดค้างรับ
  final String pendingRecvWord; // ยอดค้างรับ formatted
  final double pendingSendQty; // ยอดค้างส่ง
  final String pendingSendWord; // ยอดค้างส่ง formatted

  ProductBalance({
    required this.total,
    required this.formatted,
    required this.warehouses,
    this.pendingRecvQty = 0.0,
    this.pendingRecvWord = '',
    this.pendingSendQty = 0.0,
    this.pendingSendWord = '',
  });

  /// มียอดคงเหลือหรือไม่
  bool get hasStock => total > 0;

  /// มีข้อมูลแยกตาม warehouse หรือไม่
  bool get hasWarehouses => warehouses.isNotEmpty;

  /// จำนวน warehouses ทั้งหมด
  int get warehouseCount => warehouses.length;

  /// มียอดค้างรับหรือไม่
  bool get hasPendingRecv => pendingRecvQty > 0;

  /// มียอดค้างส่งหรือไม่
  bool get hasPendingSend => pendingSendQty > 0;

  factory ProductBalance.fromJson(Map<String, dynamic> json) {
    // Helper function to parse numeric values
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return ProductBalance(
      total: parseDouble(json['total'], 0.0),
      formatted: json['formatted'] as String? ?? '0',
      warehouses: (json['warehouses'] as List<dynamic>?)
              ?.map((e) => WarehouseBalance.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      pendingRecvQty: parseDouble(json['pending_recv_qty'], 0.0),
      pendingRecvWord: json['pending_recv_word'] as String? ?? '',
      pendingSendQty: parseDouble(json['pending_send_qty'], 0.0),
      pendingSendWord: json['pending_send_word'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'total': total,
      'formatted': formatted,
      'warehouses': warehouses.map((e) => e.toJson()).toList(),
      'pending_recv_qty': pendingRecvQty,
      'pending_recv_word': pendingRecvWord,
      'pending_send_qty': pendingSendQty,
      'pending_send_word': pendingSendWord,
    };
  }

  @override
  String toString() {
    return 'ProductBalance(total: $total, formatted: $formatted, warehouses: ${warehouses.length})';
  }
}
