import '../../../global.dart' as global;

/// Model สำหรับสินค้าจาก ClickHouse หรือ PostgreSQL
class ProductSearchModel {
  final String shopId;
  final String itemCode;
  final String barcode;
  final String name0; // ชื่อสินค้า
  final String unitCode;
  final String unitName;
  final double price1; // ราคาขายส่ง (keynumber=2)
  final double priceRetail; // ราคาขายปลีก (keynumber=1)
  final double unitStand; // อัตราส่วนหน่วย (Float64)
  final double unitDive; // ตัวหาร (Float64)
  final String? imageuri; // URL รูปภาพสินค้า

  ProductSearchModel({
    required this.shopId,
    required this.itemCode,
    required this.barcode,
    required this.name0,
    required this.unitCode,
    required this.unitName,
    required this.price1,
    this.priceRetail = 0.0,
    this.unitStand = 1.0,
    this.unitDive = 1.0,
    this.imageuri,
  });

  // สร้างจาก JSON response ของ ClickHouse หรือ PostgreSQL
  factory ProductSearchModel.fromJson(Map<String, dynamic> json) {
    // Helper function to parse numeric values that could be String or num
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return ProductSearchModel(
      shopId: json['shopid'] as String? ?? '',
      itemCode: json['itemcode'] as String? ?? '',
      barcode: json['barcode'] as String? ?? '',
      name0: json['name0'] as String? ?? '',
      unitCode: json['unitcode'] as String? ?? '',
      unitName: json['unitname'] as String? ?? '',
      price1: parseDouble(json['price1'], 0.0),
      priceRetail: parseDouble(json['price_retail'], 0.0),
      unitStand: parseDouble(json['unitstand'], 1.0),
      unitDive: parseDouble(json['unitdivide'], 1.0),
      imageuri: json['imageuri'] as String?,
    );
  }

  // แสดงข้อมูลราคาพร้อมหน่วย (ไม่มี symbol - ให้ UI จัดการ)
  String get priceDisplay {
    final price = priceRetail > 0 ? priceRetail : price1;
    return '${price.toStringAsFixed(2)}/$unitName';
  }

  // แสดงชื่อพร้อมหน่วย
  String get fullName {
    if (unitStand > 1) {
      return '$name0 ($unitStand $unitName)';
    }
    return '$name0 ($unitName)';
  }

  @override
  String toString() {
    return 'ProductSearchModel(itemCode: $itemCode, name: $name0, price: $price1, unit: $unitName)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is ProductSearchModel &&
        other.shopId == shopId &&
        other.itemCode == itemCode &&
        other.barcode == barcode;
  }

  @override
  int get hashCode => Object.hash(shopId, itemCode, barcode);
}

/// Result wrapper สำหรับ response จาก ClickHouse API
class ProductSearchResult {
  final String status;
  final String? message;
  final int count;
  final List<ProductSearchModel> products;

  ProductSearchResult({
    required this.status,
    this.message,
    required this.count,
    required this.products,
  });

  bool get isSuccess => status == 'success';
  bool get hasProducts => products.isNotEmpty;
}

/// Model สำหรับยอดคงเหลือแยกตาม Location
class LocationBalanceModel {
  final String locationCode;
  final double balanceQty;
  final String balanceWord;

  LocationBalanceModel({
    required this.locationCode,
    required this.balanceQty,
    required this.balanceWord,
  });

  factory LocationBalanceModel.fromJson(Map<String, dynamic> json) {
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return LocationBalanceModel(
      locationCode: json['location_code'] as String? ?? '',
      balanceQty: parseDouble(json['balance_qty'], 0.0),
      balanceWord: json['balance_word'] as String? ?? '',
    );
  }
}

/// Model สำหรับยอดคงเหลือแยกตาม Warehouse พร้อม Locations
class WarehouseBalanceModel {
  final String warehouseCode;
  final double balanceQty;
  final String balanceWord;
  final List<LocationBalanceModel> locations;

  WarehouseBalanceModel({
    required this.warehouseCode,
    required this.balanceQty,
    required this.balanceWord,
    required this.locations,
  });

  /// มี locations หรือไม่
  bool get hasLocations => locations.isNotEmpty;

  factory WarehouseBalanceModel.fromJson(Map<String, dynamic> json) {
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    return WarehouseBalanceModel(
      warehouseCode: json['warehouse_code'] as String? ?? '',
      balanceQty: parseDouble(json['balance_qty'], 0.0),
      balanceWord: json['balance_word'] as String? ?? '',
      locations: (json['locations'] as List<dynamic>?)
              ?.map((e) => LocationBalanceModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

/// Model สำหรับยอดคงเหลือของสินค้า (รองรับแยกตาม warehouse/location)
class ProductBalanceModel {
  final String itemCode;
  final double totalBalance; // ยอดคงเหลือทั้งหมด (หน่วยเล็กสุด)
  final Map<String, double> balanceByShop; // ยอดคงเหลือแยกตามคลัง shopId -> balance (deprecated)
  final String? formattedBalance; // ยอดคงเหลือแบบ formatted (เช่น "1 กล่อง x 2 โหล x 3 ชิ้น")
  final List<WarehouseBalanceModel> warehouses; // ยอดคงเหลือแยกตาม warehouse/location
  final double pendingRecvQty; // ยอดค้างรับ (PO ที่ยังไม่ได้รับของ)
  final String pendingRecvWord; // ยอดค้างรับ formatted
  final double pendingSendQty; // ยอดค้างส่ง (SO ที่ยังไม่ได้ส่งของ)
  final String pendingSendWord; // ยอดค้างส่ง formatted

  ProductBalanceModel({
    required this.itemCode,
    required this.totalBalance,
    required this.balanceByShop,
    this.formattedBalance,
    this.warehouses = const [],
    this.pendingRecvQty = 0.0,
    this.pendingRecvWord = '',
    this.pendingSendQty = 0.0,
    this.pendingSendWord = '',
  });

  /// มีข้อมูลแยกตาม warehouse หรือไม่
  bool get hasWarehouses => warehouses.isNotEmpty;

  /// จำนวน warehouses ทั้งหมด
  int get warehouseCount => warehouses.length;

  factory ProductBalanceModel.fromJson(Map<String, dynamic> json) {
    // Helper function to parse numeric values that could be String or num
    double parseDouble(dynamic value, double defaultValue) {
      if (value == null) return defaultValue;
      if (value is num) return value.toDouble();
      if (value is String) return double.tryParse(value) ?? defaultValue;
      return defaultValue;
    }

    final balanceByShop = <String, double>{};
    if (json['balance_by_shop'] != null) {
      final balanceData = json['balance_by_shop'] as Map<String, dynamic>;
      balanceData.forEach((key, value) {
        balanceByShop[key] = parseDouble(value, 0.0);
      });
    }

    // Parse warehouses (hierarchical balance)
    final warehouses = <WarehouseBalanceModel>[];
    if (json['warehouses'] != null) {
      final warehouseData = json['warehouses'] as List<dynamic>;
      for (final wh in warehouseData) {
        warehouses.add(WarehouseBalanceModel.fromJson(wh as Map<String, dynamic>));
      }
    }

    return ProductBalanceModel(
      itemCode: json['itemcode'] as String? ?? json['item_code'] as String? ?? '',
      totalBalance: parseDouble(json['total_balance'] ?? json['total'], 0.0),
      balanceByShop: balanceByShop,
      formattedBalance: json['formatted'] as String? ?? json['formatted_balance'] as String?,
      warehouses: warehouses,
      pendingRecvQty: parseDouble(json['pending_recv_qty'], 0.0),
      pendingRecvWord: json['pending_recv_word'] as String? ?? '',
      pendingSendQty: parseDouble(json['pending_send_qty'], 0.0),
      pendingSendWord: json['pending_send_word'] as String? ?? '',
    );
  }

  /// คำนวณ auto packing จากยอดคงเหลือและหน่วยนับ
  /// ตัวอย่าง: ถ้ามี 63 ชิ้น และ unit คือ 12:1 (โหล), baseUnit คือ "ชิ้น"
  /// จะแสดง "5 โหลx3 ชิ้น"
  ///
  /// หากชื่อหน่วยซ้ำกัน (unitName == baseUnitName) จะใช้ "อัน-ชิ้น" แทน
  /// ตัวอย่าง: ถ้ามี 25 อัน และ unit คือ 10:1 "อัน", baseUnit คือ "อัน"
  /// จะแสดง "2 อันx5 อัน-ชิ้น"
  String formatWithPacking({
    required double unitStand,
    required double unitDive,
    required String unitName,
    required String baseUnitName,
  }) {
    if (totalBalance <= 0) {
      return '${global.language("balance_remaining")} 0';
    }

    // คำนวณอัตราส่วน (เช่น 12:1 => 1 โหล = 12 ชิ้น)
    final ratio = unitStand / unitDive;

    if (ratio <= 1.0) {
      // ถ้าไม่มี packing (อัตราส่วน 1:1) ให้แสดงตามปกติ
      return '${global.language("balance_remaining")} ${totalBalance.toStringAsFixed(0)} $unitName';
    }

    // คำนวณจำนวน pack และเศษ
    final packCount = (totalBalance / ratio).floor();
    final remainder = totalBalance % ratio;

    if (remainder == 0) {
      // ไม่มีเศษ
      return '${global.language("balance_remaining")} $packCount $unitName';
    } else {
      // มีเศษ - ตรวจสอบว่าชื่อหน่วยซ้ำกันหรือไม่
      final remainderUnitName = (unitName == baseUnitName)
          ? global.language("piece")
          : baseUnitName;

      return '${global.language("balance_remaining")} $packCount ${unitName}x${remainder.toStringAsFixed(0)} $remainderUnitName';
    }
  }

  /// คำนวณ auto packing แบบหลายระดับ (Multi-Level Packing)
  /// ตัวอย่าง: ถ้ามี 171 ชิ้น และมีหน่วย [1:1 "ชิ้น", 12:1 "โหล", 144:1 "กล่อง"]
  /// จะแสดง "คงเหลือ 1 กล่อง x 2 โหล x 3 ชิ้น"
  ///
  /// Algorithm: Greedy - หารด้วยหน่วยใหญ่สุดก่อน แล้วหารเศษด้วยหน่วยถัดไป
  ///
  /// กรณีหน่วยซ้ำ: ถ้าชื่อหน่วยใหญ่ซ้ำกับหน่วยเล็ก → ใช้ "แพ็ค" แทน
  String formatWithMultiLevelPacking({
    required List<ProductSearchModel> allUnits,
  }) {
    if (totalBalance <= 0) {
      return '${global.language("balance_remaining")} 0';
    }

    if (allUnits.isEmpty) {
      return '${global.language("balance_remaining")} ${totalBalance.toStringAsFixed(0)}';
    }

    // หาหน่วยฐาน (1:1)
    final baseUnit = allUnits.firstWhere(
      (u) => u.unitStand == 1 && u.unitDive == 1,
      orElse: () => allUnits.first,
    );
    final baseUnitName = baseUnit.unitName.isNotEmpty
        ? baseUnit.unitName
        : baseUnit.unitCode;

    // กรอง units ที่มี ratio > 1 และเรียงจากมาก → น้อย
    final packingUnits =
        allUnits.where((u) {
          final ratio = u.unitStand / u.unitDive;
          return ratio > 1.0;
        }).toList()..sort((a, b) {
          final ratioA = a.unitStand / a.unitDive;
          final ratioB = b.unitStand / b.unitDive;
          return ratioB.compareTo(ratioA); // มาก → น้อย
        });

    // ถ้าไม่มีหน่วย packing ให้แสดงแบบปกติ
    if (packingUnits.isEmpty) {
      return '${global.language("balance_remaining")} ${totalBalance.toStringAsFixed(0)} $baseUnitName';
    }

    // คำนวณแบบ Greedy
    double remainingBalance = totalBalance;
    final List<String> parts = [];
    final Set<String> usedUnitNames = {}; // ตรวจสอบชื่อหน่วยซ้ำ

    for (final unit in packingUnits) {
      final ratio = unit.unitStand / unit.unitDive;
      final count = (remainingBalance / ratio).floor();

      if (count > 0) {
        // ตรวจสอบชื่อหน่วยซ้ำ
        String displayUnitName = unit.unitName.isNotEmpty
            ? unit.unitName
            : unit.unitCode;

        // ถ้าชื่อซ้ำกับที่ใช้ไปแล้ว หรือซ้ำกับ base unit → ใช้ "แพ็ค"
        if (usedUnitNames.contains(displayUnitName) ||
            displayUnitName == baseUnitName) {
          displayUnitName = global.language("pack");
        }

        usedUnitNames.add(displayUnitName);
        parts.add('$count $displayUnitName');
        remainingBalance -= count * ratio;
      }
    }

    // เศษที่เหลือ (ถ้ามี)
    if (remainingBalance > 0.001) {
      // เปลี่ยนจาก 0.01 → 0.001 เพื่อความแม่นยำ
      // แสดงทศนิยมตามจริง (สูงสุด 6 ตำแหน่ง) และตัด trailing zeros
      String remainderDisplay;

      // ตรวจสอบว่าเป็นจำนวนเต็มหรือมีทศนิยมเท่ากับ 0 (เช่น 3.00)
      final floorValue = remainingBalance.floor().toDouble();
      if ((remainingBalance - floorValue).abs() < 0.0001) {
        // เป็นจำนวนเต็มหรือใกล้เคียงจำนวนเต็มมาก (เช่น 3.00) → แสดงเป็นเลขเต็ม
        remainderDisplay = floorValue.toInt().toString();
      } else {
        // มีทศนิยมจริงๆ → แสดงทศนิยมตามจริง (สูงสุด 6 ตำแหน่ง)
        remainderDisplay = remainingBalance
            .toStringAsFixed(6)
            .replaceAll(
              RegExp(r'0+$'),
              '',
            ) // ตัด trailing zeros (เช่น 3.500000 → 3.5)
            .replaceAll(
              RegExp(r'\.$'),
              '',
            ); // ตัดจุดทศนิยมถ้าไม่มีเลขหลังจุด (เช่น 3. → 3)
      }

      // ถ้าหน่วยเศษซ้ำกับหน่วยที่ใช้ไปแล้ว → ใช้ "อัน-ชิ้น"
      final remainderUnitName = usedUnitNames.contains(baseUnitName)
          ? global.language("piece")
          : baseUnitName;

      parts.add('$remainderDisplay $remainderUnitName');
    }

    // สร้าง output string
    if (parts.isEmpty) {
      return '${global.language("balance_remaining")} ${totalBalance.toStringAsFixed(0)} $baseUnitName';
    }

    return '${global.language("balance_remaining")} ${parts.join(' x ')}';
  }

  @override
  String toString() {
    return 'ProductBalanceModel(itemCode: $itemCode, totalBalance: $totalBalance, shops: ${balanceByShop.keys.length})';
  }
}
