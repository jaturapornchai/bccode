import 'package:uuid/uuid.dart';

/// Model สำหรับรายการสินค้าในตะกร้า
class CartItemModel {
  final String itemId; // รหัสรายการ (UUID)
  final String itemCode; // รหัสสินค้า
  final String barcode; // บาร์โค้ด
  final String productName; // ชื่อสินค้า
  final String unitCode; // รหัสหน่วย
  final String unitName; // ชื่อหน่วย
  final double quantity; // จำนวน
  final double price; // ราคาต่อหน่วย (price1)
  final double discount; // ส่วนลด (%)
  final String? remark; // หมายเหตุ
  final String? imageuri; // URL รูปภาพสินค้า
  final double totalPrice; // ราคารวมหลังหักส่วนลด
  final int unitStand; // อัตราส่วนหน่วย (เช่น 12 ชิ้น/แพ็ค)
  final int unitDive; // ตัวหาร (เช่น 1)
  final DateTime addedAt;

  // ข้อมูลคลังและที่เก็บ (บันทึกในแต่ละรายการสินค้า)
  final String? warehouseId; // รหัสคลังสินค้า
  final String? warehouseName; // ชื่อคลังสินค้า
  final String? locationId; // รหัสที่เก็บ
  final String? locationName; // ชื่อที่เก็บ

  CartItemModel({
    String? itemId,
    required this.itemCode,
    required this.barcode,
    required this.productName,
    required this.unitCode,
    required this.unitName,
    required this.quantity,
    required this.price,
    this.discount = 0.0,
    this.remark,
    this.imageuri,
    required this.unitStand,
    required this.unitDive,
    DateTime? addedAt,
    this.warehouseId,
    this.warehouseName,
    this.locationId,
    this.locationName,
  }) : itemId = itemId ?? const Uuid().v4(),
       totalPrice = _calculateTotalPrice(quantity, price, discount),
       addedAt = addedAt ?? DateTime.now();

  /// คำนวณราคารวมหลังหักส่วนลด
  static double _calculateTotalPrice(double quantity, double price, double discount) {
    final subtotal = quantity * price;
    final discountAmount = subtotal * (discount / 100);
    return subtotal - discountAmount;
  }

  /// ราคารวมก่อนหักส่วนลด
  double get subtotal => quantity * price;

  /// จำนวนเงินส่วนลด
  double get discountAmount => subtotal * (discount / 100);

  // คำนวณราคารวมใหม่เมื่อเปลี่ยนจำนวน
  CartItemModel copyWithQuantity(double newQuantity) {
    return CartItemModel(
      itemId: itemId,
      itemCode: itemCode,
      barcode: barcode,
      productName: productName,
      unitCode: unitCode,
      unitName: unitName,
      quantity: newQuantity,
      price: price,
      discount: discount,
      remark: remark,
      imageuri: imageuri,
      unitStand: unitStand,
      unitDive: unitDive,
      addedAt: addedAt,
      warehouseId: warehouseId,
      warehouseName: warehouseName,
      locationId: locationId,
      locationName: locationName,
    );
  }

  // Copy พร้อมแก้ไขค่าต่างๆ
  CartItemModel copyWith({
    double? quantity,
    double? price,
    double? discount,
    String? remark,
    String? imageuri,
    String? warehouseId,
    String? warehouseName,
    String? locationId,
    String? locationName,
  }) {
    return CartItemModel(
      itemId: itemId,
      itemCode: itemCode,
      barcode: barcode,
      productName: productName,
      unitCode: unitCode,
      unitName: unitName,
      quantity: quantity ?? this.quantity,
      price: price ?? this.price,
      discount: discount ?? this.discount,
      remark: remark ?? this.remark,
      imageuri: imageuri ?? this.imageuri,
      unitStand: unitStand,
      unitDive: unitDive,
      addedAt: addedAt,
      warehouseId: warehouseId ?? this.warehouseId,
      warehouseName: warehouseName ?? this.warehouseName,
      locationId: locationId ?? this.locationId,
      locationName: locationName ?? this.locationName,
    );
  }

  // แปลงเป็น JSON สำหรับบันทึกใน MongoDB
  Map<String, dynamic> toJson() {
    return {
      'itemId': itemId,
      'itemCode': itemCode,
      'barcode': barcode,
      'productName': productName,
      'unitCode': unitCode,
      'unitName': unitName,
      'quantity': quantity,
      'price': price,
      'discount': discount,
      'remark': remark,
      'imageuri': imageuri,
      'subtotal': subtotal,
      'discountAmount': discountAmount,
      'totalPrice': totalPrice,
      'unitStand': unitStand,
      'unitDive': unitDive,
      'addedAt': addedAt.toIso8601String(),
      'warehouseId': warehouseId,
      'warehouseName': warehouseName,
      'locationId': locationId,
      'locationName': locationName,
    };
  }

  // สร้างจาก JSON
  factory CartItemModel.fromJson(Map<String, dynamic> json) {
    return CartItemModel(
      itemId: json['itemId'] as String,
      itemCode: json['itemCode'] as String,
      barcode: json['barcode'] as String? ?? '',
      productName: json['productName'] as String,
      unitCode: json['unitCode'] as String,
      unitName: json['unitName'] as String,
      quantity: (json['quantity'] as num).toDouble(),
      price: (json['price'] as num).toDouble(),
      discount: (json['discount'] as num?)?.toDouble() ?? 0.0,
      remark: json['remark'] as String?,
      imageuri: json['imageuri'] as String?,
      unitStand: json['unitStand'] as int? ?? 1,
      unitDive: json['unitDive'] as int? ?? 1,
      addedAt: json['addedAt'] != null
          ? DateTime.parse(json['addedAt'] as String)
          : DateTime.now(),
      warehouseId: json['warehouseId'] as String?,
      warehouseName: json['warehouseName'] as String?,
      locationId: json['locationId'] as String?,
      locationName: json['locationName'] as String?,
    );
  }

  @override
  String toString() {
    return 'CartItemModel(itemId: $itemId, productName: $productName, quantity: $quantity, price: $price, totalPrice: $totalPrice)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is CartItemModel && other.itemId == itemId;
  }

  @override
  int get hashCode => itemId.hashCode;
}
