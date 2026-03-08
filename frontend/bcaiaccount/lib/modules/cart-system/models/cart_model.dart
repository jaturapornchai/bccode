import 'package:uuid/uuid.dart';
import 'cart_item_model.dart';
import 'cart_system_type.dart';
import 'cart_type.dart';
import '../../../utils/logger/app_logger.dart';

/// Model สำหรับตะกร้าสินค้า (MongoDB Document) - รองรับ 3 ระบบ
class CartModel {
  final String shopId; // identifier 1
  final String email; // identifier 2 (ผู้ใช้)
  final String cartId; // identifier 3 (รหัสตะกร้า - GUID)
  final String cartName; // ชื่อตะกร้า
  final CartSystemType systemType; // ประเภทระบบ (sales, purchase, inventory)
  final CartType
  cartType; // ประเภทตะกร้า (sale, salesOrder, quotation, purchaseOrder, withdraw, transfer)
  final String status; // สถานะ (active, ordered, completed, cancelled)
  final List<CartItemModel> items; // รายการสินค้า
  final double totalAmount; // ยอดรวม
  final DateTime createdAt;
  final DateTime updatedAt;

  // ข้อมูลลูกหนี้ (สำหรับระบบขาย - required)
  final String? debtorId; // GUID ของลูกหนี้
  final String? debtorCode; // รหัสลูกหนี้
  final String? debtorName; // ชื่อลูกหนี้

  // ข้อมูลเจ้าหนี้ (สำหรับระบบซื้อ - required)
  final String? creditorId; // GUID ของเจ้าหนี้
  final String? creditorCode; // รหัสเจ้าหนี้
  final String? creditorName; // ชื่อเจ้าหนี้

  // ข้อมูลคลังและที่เก็บ (optional - เป็นข้อมูลเสริม)
  final String? warehouseId; // รหัสคลังสินค้า
  final String? warehouseName; // ชื่อคลังสินค้า
  final String? locationId; // รหัสที่เก็บ (location) - optional
  final String? locationName; // ชื่อที่เก็บ - optional

  final Map<String, dynamic>? customFields; // ฟิลด์เพิ่มเติม (ยืดหยุ่น)

  CartModel({
    required this.shopId,
    required this.email,
    String? cartId,
    required this.cartName,
    required this.systemType,
    CartType? cartType,
    this.status = 'active',
    List<CartItemModel>? items,
    DateTime? createdAt,
    DateTime? updatedAt,
    this.debtorId,
    this.debtorCode,
    this.debtorName,
    this.creditorId,
    this.creditorCode,
    this.creditorName,
    this.warehouseId,
    this.warehouseName,
    this.locationId,
    this.locationName,
    this.customFields,
  }) : cartId = cartId ?? const Uuid().v4(),
       cartType = cartType ?? getDefaultCartType(systemType),
       items = items ?? [],
       totalAmount = _calculateTotalAmount(items ?? []),
       createdAt = createdAt ?? DateTime.now(),
       updatedAt = updatedAt ?? DateTime.now() {
    // Validate ตาม systemType
    _validateCart();

    // Log เมื่อสร้าง cart ใหม่
    if (cartId == null) {
      AppLogger.info(
        '🛒 CartModel: Generated new GUID cartId: ${this.cartId} (${systemType.displayNameThai} - ${this.cartType.displayNameThai})',
      );
    } else {
      AppLogger.debug('🛒 CartModel: Using existing cartId: ${this.cartId}');
    }
  }

  /// ตรวจสอบความถูกต้องของข้อมูลตาม systemType
  void _validateCart() {
    // ตรวจสอบว่า cartType ตรงกับ systemType หรือไม่
    if (!cartType.isValidForSystemType(systemType)) {
      throw ArgumentError(
        'CartType "${cartType.displayNameThai}" ไม่สามารถใช้กับระบบ "${systemType.displayNameThai}" ได้',
      );
    }
  }

  // คำนวณยอดรวมทั้งหมด
  static double _calculateTotalAmount(List<CartItemModel> items) {
    return items.fold(0.0, (sum, item) => sum + item.totalPrice);
  }

  // คำนวณยอดรวมใหม่
  double get recalculatedTotal => _calculateTotalAmount(items);

  // จำนวนรายการสินค้าทั้งหมด
  int get itemCount => items.length;

  // จำนวนสินค้าทั้งหมด (รวม quantity)
  double get totalQuantity =>
      items.fold(0.0, (sum, item) => sum + item.quantity);

  // ตรวจสอบว่าเป็นระบบขายหรือไม่
  bool get isSalesSystem => systemType == CartSystemType.sales;

  // ตรวจสอบว่าเป็นระบบซื้อหรือไม่
  bool get isPurchaseSystem => systemType == CartSystemType.purchase;

  // ตรวจสอบว่าเป็นระบบสินค้าคงคลังหรือไม่
  bool get isInventorySystem => systemType == CartSystemType.inventory;

  // เพิ่มสินค้า
  CartModel addItem(CartItemModel item) {
    final updatedItems = List<CartItemModel>.from(items)..add(item);
    return _copyWith(items: updatedItems);
  }

  // อัพเดทจำนวนสินค้า
  CartModel updateItemQuantity(String itemId, double newQuantity) {
    final updatedItems = items.map((item) {
      if (item.itemId == itemId) {
        return item.copyWithQuantity(newQuantity);
      }
      return item;
    }).toList();
    return _copyWith(items: updatedItems);
  }

  // ลบสินค้า
  CartModel removeItem(String itemId) {
    final updatedItems = items.where((item) => item.itemId != itemId).toList();
    return _copyWith(items: updatedItems);
  }

  // เปลี่ยนสถานะ
  CartModel updateStatus(String newStatus) {
    return _copyWith(status: newStatus);
  }

  // อัพเดทลูกหนี้ (สำหรับระบบขาย)
  CartModel updateDebtor({
    required String debtorId,
    required String debtorCode,
    required String debtorName,
  }) {
    if (!isSalesSystem) {
      throw StateError('ไม่สามารถอัพเดทลูกหนี้ได้ในระบบนี้');
    }
    return _copyWith(
      debtorId: debtorId,
      debtorCode: debtorCode,
      debtorName: debtorName,
    );
  }

  // อัพเดทเจ้าหนี้ (สำหรับระบบซื้อ)
  CartModel updateCreditor({
    required String creditorId,
    required String creditorCode,
    required String creditorName,
  }) {
    if (!isPurchaseSystem) {
      throw StateError('ไม่สามารถอัพเดทเจ้าหนี้ได้ในระบบนี้');
    }
    return _copyWith(
      creditorId: creditorId,
      creditorCode: creditorCode,
      creditorName: creditorName,
    );
  }

  // อัพเดทคลังและที่เก็บ
  CartModel updateWarehouseAndLocation({
    required String warehouseId,
    required String warehouseName,
    String? locationId,
    String? locationName,
  }) {
    return _copyWith(
      warehouseId: warehouseId,
      warehouseName: warehouseName,
      locationId: locationId,
      locationName: locationName,
    );
  }

  // อัพเดทชื่อตะกร้า
  CartModel updateName(String newName) {
    return _copyWith(cartName: newName);
  }

  // อัพเดทประเภทตะกร้า
  CartModel updateCartType(CartType newCartType) {
    // ตรวจสอบว่า cartType ใหม่สามารถใช้กับระบบนี้ได้หรือไม่
    if (!newCartType.isValidForSystemType(systemType)) {
      throw ArgumentError(
        'CartType "${newCartType.displayNameThai}" ไม่สามารถใช้กับระบบ "${systemType.displayNameThai}" ได้',
      );
    }
    return _copyWith(cartType: newCartType);
  }

  // อัพเดทชื่อและประเภทตะกร้าพร้อมกัน
  CartModel updateNameAndType(String newName, CartType newCartType) {
    // ตรวจสอบว่า cartType ใหม่สามารถใช้กับระบบนี้ได้หรือไม่
    if (!newCartType.isValidForSystemType(systemType)) {
      throw ArgumentError(
        'CartType "${newCartType.displayNameThai}" ไม่สามารถใช้กับระบบ "${systemType.displayNameThai}" ได้',
      );
    }
    return _copyWith(cartName: newName, cartType: newCartType);
  }

  // Helper method สำหรับ copy with
  CartModel _copyWith({
    String? cartName,
    CartSystemType? systemType,
    CartType? cartType,
    String? status,
    List<CartItemModel>? items,
    String? debtorId,
    String? debtorCode,
    String? debtorName,
    String? creditorId,
    String? creditorCode,
    String? creditorName,
    String? warehouseId,
    String? warehouseName,
    String? locationId,
    String? locationName,
    Map<String, dynamic>? customFields,
  }) {
    final updatedItems = items ?? this.items;
    return CartModel(
      shopId: shopId,
      email: email,
      cartId: cartId,
      cartName: cartName ?? this.cartName,
      systemType: systemType ?? this.systemType,
      cartType: cartType ?? this.cartType,
      status: status ?? this.status,
      items: updatedItems,
      createdAt: createdAt,
      updatedAt: DateTime.now(),
      debtorId: debtorId ?? this.debtorId,
      debtorCode: debtorCode ?? this.debtorCode,
      debtorName: debtorName ?? this.debtorName,
      creditorId: creditorId ?? this.creditorId,
      creditorCode: creditorCode ?? this.creditorCode,
      creditorName: creditorName ?? this.creditorName,
      warehouseId: warehouseId ?? this.warehouseId,
      warehouseName: warehouseName ?? this.warehouseName,
      locationId: locationId ?? this.locationId,
      locationName: locationName ?? this.locationName,
      customFields: customFields ?? this.customFields,
    );
  }

  // แปลงเป็น JSON สำหรับบันทึกใน MongoDB
  Map<String, dynamic> toJson() {
    return {
      'shopId': shopId,
      'email': email,
      'cartId': cartId,
      'cartName': cartName,
      'systemType': systemType.toJson(),
      'cartType': cartType.toJson(),
      'status': status,
      'items': items.map((item) => item.toJson()).toList(),
      'totalAmount': recalculatedTotal,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
      // ข้อมูลลูกหนี้
      if (debtorId != null) 'debtorId': debtorId,
      if (debtorCode != null) 'debtorCode': debtorCode,
      if (debtorName != null) 'debtorName': debtorName,
      // ข้อมูลเจ้าหนี้
      if (creditorId != null) 'creditorId': creditorId,
      if (creditorCode != null) 'creditorCode': creditorCode,
      if (creditorName != null) 'creditorName': creditorName,
      // ข้อมูลคลังและที่เก็บ
      'warehouseId': warehouseId,
      'warehouseName': warehouseName,
      if (locationId != null) 'locationId': locationId,
      if (locationName != null) 'locationName': locationName,
      if (customFields != null) 'customFields': customFields,
    };
  }

  // สร้างจาก JSON
  factory CartModel.fromJson(Map<String, dynamic> json) {
    // อ่าน systemType จาก JSON หรือใช้ default ถ้าไม่มี
    final systemTypeStr = json['systemType'] as String?;
    final systemType = systemTypeStr != null
        ? cartSystemTypeFromJson(systemTypeStr)
        : CartSystemType.sales; // default เป็น sales ถ้าไม่มี

    // อ่าน cartType จาก JSON หรือใช้ default ถ้าไม่มี
    final cartType = json['cartType'] != null
        ? cartTypeFromJson(json['cartType'] as String)
        : null;

    return CartModel(
      shopId: json['shopId'] as String,
      email: json['email'] as String,
      cartId: json['cartId'] as String,
      cartName: json['cartName'] as String,
      systemType: systemType,
      cartType: cartType, // จะใช้ default ใน constructor ถ้าเป็น null
      status: json['status'] as String? ?? 'active',
      items:
          (json['items'] as List<dynamic>?)
              ?.map(
                (item) => CartItemModel.fromJson(item as Map<String, dynamic>),
              )
              .toList() ??
          [],
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'] as String)
          : DateTime.now(),
      updatedAt: json['updatedAt'] != null
          ? DateTime.parse(json['updatedAt'] as String)
          : DateTime.now(),
      // ข้อมูลลูกหนี้
      debtorId: json['debtorId'] as String?,
      debtorCode: json['debtorCode'] as String?,
      debtorName: json['debtorName'] as String?,
      // ข้อมูลเจ้าหนี้
      creditorId: json['creditorId'] as String?,
      creditorCode: json['creditorCode'] as String?,
      creditorName: json['creditorName'] as String?,
      // ข้อมูลคลังและที่เก็บ
      warehouseId: json['warehouseId'] as String?,
      warehouseName: json['warehouseName'] as String?,
      locationId: json['locationId'] as String?,
      locationName: json['locationName'] as String?,
      customFields: json['customFields'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() {
    return 'CartModel(cartId: $cartId, system: ${systemType.displayNameThai}, type: ${cartType.displayNameThai}, name: $cartName, status: $status, items: $itemCount, total: $totalAmount)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is CartModel &&
        other.shopId == shopId &&
        other.email == email &&
        other.cartId == cartId;
  }

  @override
  int get hashCode => Object.hash(shopId, email, cartId);
}
