import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// Model สำหรับรายการซื้อ/ขายของสินค้า
class ProductTransactionModel {
  final String docCode; // เลขที่เอกสาร
  final DateTime docDate; // วันที่เอกสาร
  final String itemCode; // รหัสสินค้า
  final String itemName; // ชื่อสินค้า
  final double qty; // จำนวน
  final String unitName; // หน่วย
  final double price; // ราคาต่อหน่วย
  final double amount; // จำนวนเงิน
  final String docType; // ประเภทเอกสาร (PO, SO, IV, etc.)
  final String docTypeName; // ชื่อประเภทเอกสาร (ซื้อ, ขาย, ฯลฯ)
  final String? refCode; // เลขที่อ้างอิง
  final String? customerName; // ชื่อลูกค้า/ผู้จำหน่าย
  final String? remark; // หมายเหตุ

  ProductTransactionModel({
    required this.docCode,
    required this.docDate,
    required this.itemCode,
    required this.itemName,
    required this.qty,
    required this.unitName,
    required this.price,
    required this.amount,
    required this.docType,
    required this.docTypeName,
    this.refCode,
    this.customerName,
    this.remark,
  });

  factory ProductTransactionModel.fromJson(Map<String, dynamic> json) {
    // Parse date (รองรับทั้ง String และ DateTime)
    DateTime parseDate(dynamic dateValue) {
      if (dateValue is DateTime) {
        return dateValue;
      } else if (dateValue is String) {
        try {
          return DateTime.parse(dateValue);
        } catch (e) {
          return DateTime.now();
        }
      }
      return DateTime.now();
    }

    return ProductTransactionModel(
      docCode: json['docno'] as String? ?? json['doccode'] as String? ?? '',
      docDate: parseDate(json['docdatetime'] ?? json['docdate']),
      itemCode: json['itemcode'] as String? ?? '',
      itemName: json['itemname'] as String? ?? json['name0'] as String? ?? '',
      qty: (json['qty'] as num?)?.toDouble() ?? 0.0,
      unitName:
          json['unitname'] as String? ?? json['unitcode'] as String? ?? '',
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      amount: (json['amount'] as num?)?.toDouble() ?? 0.0,
      docType: json['doctype'] as String? ?? '',
      docTypeName: json['doctypename'] as String? ?? '',
      refCode: json['refcode'] as String?,
      customerName: json['customername'] as String?,
      remark: json['remark'] as String?,
    );
  }

  /// แสดงวันที่แบบสั้น (dd/MM/yyyy)
  String get formattedDate {
    return '${docDate.day.toString().padLeft(2, '0')}/${docDate.month.toString().padLeft(2, '0')}/${docDate.year}';
  }

  /// แสดงวันที่แบบเต็ม (dd/MM/yyyy HH:mm)
  String get formattedDateTime {
    return '$formattedDate ${docDate.hour.toString().padLeft(2, '0')}:${docDate.minute.toString().padLeft(2, '0')}';
  }

  /// แสดงวันที่แบบไทย (ใช้ global.formatThaiDateTime)
  /// ต้อง import '../../../global.dart' as global; ก่อนใช้งาน
  String getThaiDateTime() {
    // ใช้ global.formatThaiDateTime จากไฟล์ที่เรียกใช้
    // หรือจะ return formattedDateTime ไปก่อนแล้วให้ UI จัดการเอง
    return formattedDateTime;
  }

  @override
  String toString() {
    return 'ProductTransactionModel(docCode: $docCode, date: $formattedDate, qty: $qty $unitName, price: $price)';
  }
}

/// Model สำหรับยอดคงเหลือตาม Warehouse และ Location
class ProductWarehouseBalanceModel {
  final String warehouseId; // รหัสคลัง
  final String warehouseName; // ชื่อคลัง
  final double totalBalance; // ยอดรวมในคลังนี้
  final Map<String, ProductLocationBalanceModel>
  locationBalances; // ยอดแยกตาม location

  ProductWarehouseBalanceModel({
    required this.warehouseId,
    required this.warehouseName,
    required this.totalBalance,
    required this.locationBalances,
  });

  factory ProductWarehouseBalanceModel.fromJson(Map<String, dynamic> json) {
    final locationBalances = <String, ProductLocationBalanceModel>{};

    if (json['locations'] != null) {
      final locations = json['locations'] as List<dynamic>;
      for (final loc in locations) {
        final locModel = ProductLocationBalanceModel.fromJson(loc);
        locationBalances[locModel.locationId] = locModel;
      }
    }

    return ProductWarehouseBalanceModel(
      warehouseId: json['warehouse_id'] as String? ?? '',
      warehouseName: json['warehouse_name'] as String? ?? '',
      totalBalance: (json['total_balance'] as num?)?.toDouble() ?? 0.0,
      locationBalances: locationBalances,
    );
  }

  /// จำนวน location ที่มียอดคงเหลือ
  int get locationCount => locationBalances.length;

  /// รายการ location ทั้งหมด
  List<ProductLocationBalanceModel> get locations =>
      locationBalances.values.toList();

  /// แสดงยอดคงเหลือแบบสรุป
  /// Format: "WH01 (รวม 195) - A01: 100, A02: 95" หรือ "WH01 (ทั่วไป): 100"
  String getSummary() {
    // แสดงรหัสคลัง ถ้าว่างให้แสดง "คลังหลัก"
    final warehouseDisplay = warehouseId.isNotEmpty ? warehouseId : global.language("default_warehouse");

    AppLogger.debug(
      '🔍 getSummary() - warehouseId: "$warehouseId", '
      'warehouseName: "$warehouseName", '
      'totalBalance: $totalBalance, '
      'locationCount: ${locationBalances.length}',
    );

    if (locationBalances.isEmpty) {
      final result =
          '$warehouseDisplay (${global.language("total")} ${totalBalance.toStringAsFixed(2)})';
      AppLogger.debug('📝 Result (no locations): $result');
      return result;
    }

    // ถ้ามี location เดียวและเป็น location ว่าง แสดงแบบย่อ
    if (locationBalances.length == 1) {
      final singleLoc = locations.first;
      AppLogger.debug(
        '📍 Single location - id: "${singleLoc.locationId}", '
        'name: "${singleLoc.locationName}", '
        'balance: ${singleLoc.balance}',
      );
      if (singleLoc.locationId.isEmpty) {
        final result =
            '$warehouseDisplay (${global.language("general_location")}): ${singleLoc.balance.toStringAsFixed(2)}';
        AppLogger.debug('📝 Result (single empty location): $result');
        return result;
      }
    }

    // ถ้ามีหลาย location แสดงแบบละเอียด
    final locationParts = locations
        .where((loc) => loc.balance > 0)
        .map((loc) {
          AppLogger.debug(
            '📍 Location - id: "${loc.locationId}", balance: ${loc.balance}',
          );
          // ถ้า locationId ว่าง ให้แสดง "ทั่วไป" เท่านั้น
          // ถ้ามี locationId ให้แสดงรหัส location
          return loc.locationId.isEmpty
              ? '${global.language("general_location")}: ${loc.balance.toStringAsFixed(2)}'
              : '${loc.locationId}: ${loc.balance.toStringAsFixed(2)}';
        })
        .join(', ');

    final result =
        '$warehouseDisplay (${global.language("total")} ${totalBalance.toStringAsFixed(2)}) - $locationParts';
    AppLogger.debug('📝 Result (multiple locations): $result');
    return result;
  }
}

/// Model สำหรับยอดคงเหลือตาม Location
class ProductLocationBalanceModel {
  final String locationId; // รหัสที่เก็บ
  final String locationName; // ชื่อที่เก็บ
  final double balance; // ยอดคงเหลือ

  ProductLocationBalanceModel({
    required this.locationId,
    required this.locationName,
    required this.balance,
  });

  factory ProductLocationBalanceModel.fromJson(Map<String, dynamic> json) {
    return ProductLocationBalanceModel(
      locationId: json['location_id'] as String? ?? '',
      locationName: json['location_name'] as String? ?? '',
      balance: (json['balance'] as num?)?.toDouble() ?? 0.0,
    );
  }

  @override
  String toString() {
    return 'ProductLocationBalanceModel(location: $locationName, balance: $balance)';
  }
}
