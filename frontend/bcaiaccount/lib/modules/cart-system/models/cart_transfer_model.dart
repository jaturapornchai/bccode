import 'package:flutter/material.dart';
import 'package:smlaicloud/screens/transaction/transaction_edit.dart';
import 'package:smlaicloud/global.dart' as global;
import 'cart_model.dart';
import 'cart_item_model.dart';
import 'cart_type.dart';
import 'cart_system_type.dart';

/// Model สำหรับส่งข้อมูลจาก Cart System ไปยังระบบอื่น
///
/// ใช้สำหรับส่งข้อมูลจากตะกร้าไปยัง:
/// - /transaction/sale (ขาย)
/// - /transaction/saleorder (สั่งขาย)
/// - /transaction/quotation (เสนอราคา)
/// - /transaction/purchase (ซื้อ)
/// - /transaction/purchaseorder (สั่งซื้อ)
/// - /transaction/stocktransfer (โอน)
/// - /transaction/stockpickupproduct (เบิก)
class CartTransferModel {
  // ข้อมูลตะกร้า
  final String cartId;
  final String cartName;
  final CartSystemType systemType;
  final CartType cartType;

  // ข้อมูลร้านค้าและผู้ใช้
  final String shopId;
  final String email;

  // ข้อมูลลูกหนี้ (สำหรับระบบขาย)
  final String? debtorId;
  final String? debtorCode;
  final String? debtorName;

  // ข้อมูลเจ้าหนี้ (สำหรับระบบซื้อ)
  final String? creditorId;
  final String? creditorCode;
  final String? creditorName;

  // ข้อมูลคลังและที่เก็บ
  final String? warehouseId;
  final String? warehouseName;
  final String? locationId;
  final String? locationName;

  // รายการสินค้า
  final List<CartTransferItemModel> items;

  // ยอดรวม
  final double totalAmount;
  final int itemCount;

  // วันที่สร้าง
  final DateTime createdAt;

  CartTransferModel({
    required this.cartId,
    required this.cartName,
    required this.systemType,
    required this.cartType,
    required this.shopId,
    required this.email,
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
    required this.items,
    required this.totalAmount,
    required this.itemCount,
    required this.createdAt,
  });

  /// สร้าง CartTransferModel จาก CartModel
  factory CartTransferModel.fromCartModel(CartModel cart) {
    return CartTransferModel(
      cartId: cart.cartId,
      cartName: cart.cartName,
      systemType: cart.systemType,
      cartType: cart.cartType,
      shopId: cart.shopId,
      email: cart.email,
      debtorId: cart.debtorId,
      debtorCode: cart.debtorCode,
      debtorName: cart.debtorName,
      creditorId: cart.creditorId,
      creditorCode: cart.creditorCode,
      creditorName: cart.creditorName,
      warehouseId: cart.warehouseId,
      warehouseName: cart.warehouseName,
      locationId: cart.locationId,
      locationName: cart.locationName,
      items: cart.items
          .map((item) => CartTransferItemModel.fromCartItemModel(item))
          .toList(),
      totalAmount: cart.totalAmount,
      itemCount: cart.itemCount,
      createdAt: cart.createdAt,
    );
  }

  /// ดึง route path ตามประเภทตะกร้า
  String getTargetRoute() {
    switch (cartType) {
      case CartType.sale:
        return '/transaction/sale';
      case CartType.salesOrder:
        return '/transaction/saleorder';
      case CartType.quotation:
        return '/transaction/quotation';
      case CartType.purchaseOrder:
        return '/transaction/purchaseorder';
      case CartType.withdraw:
        return '/transaction/stockpickupproduct';
      case CartType.transfer:
        return '/transaction/stocktransfer';
    }
  }

  /// ชื่อระบบปลายทาง (สำหรับแสดงผล)
  String getTargetSystemName() {
    return cartType.displayNameThai;
  }

  /// สร้าง TransactionEditScreen Widget สำหรับเปิดใน Tab ใหม่
  Widget createTransactionScreen() {
    // แปลง CartType เป็น TransactionTypeEnum
    global.TransactionTypeEnum transactionType;

    switch (cartType) {
      case CartType.sale:
        transactionType = global.TransactionTypeEnum.sale;
        break;
      case CartType.salesOrder:
        transactionType = global.TransactionTypeEnum.saleorder;
        break;
      case CartType.quotation:
        transactionType = global.TransactionTypeEnum.quotation;
        break;
      case CartType.purchaseOrder:
        transactionType = global.TransactionTypeEnum.purchaseorder;
        break;
      case CartType.withdraw:
        transactionType = global.TransactionTypeEnum.stockpickupproduct;
        break;
      case CartType.transfer:
        transactionType = global.TransactionTypeEnum.stocktransfer;
        break;
    }

    // สร้าง Wrapper Widget ที่มี Navigator พร้อม arguments
    return _TransactionScreenWrapper(
      transactionType: transactionType,
      cartTransferModel: this,
    );
  }
}

/// Wrapper Widget สำหรับส่ง CartTransferModel ไปยัง TransactionEditScreen
class _TransactionScreenWrapper extends StatelessWidget {
  final global.TransactionTypeEnum transactionType;
  final CartTransferModel cartTransferModel;

  const _TransactionScreenWrapper({
    required this.transactionType,
    required this.cartTransferModel,
  });

  @override
  Widget build(BuildContext context) {
    // ใช้ InheritedWidget เพื่อส่ง CartTransferModel ลงไปให้ TransactionEditScreen
    return CartTransferProvider(
      cartTransferModel: cartTransferModel,
      child: TransactionEditScreen(type: transactionType),
    );
  }
}

/// InheritedWidget สำหรับส่ง CartTransferModel
class CartTransferProvider extends InheritedWidget {
  final CartTransferModel cartTransferModel;

  const CartTransferProvider({
    super.key,
    required this.cartTransferModel,
    required super.child,
  });

  static CartTransferModel? of(BuildContext context) {
    return context
        .dependOnInheritedWidgetOfExactType<CartTransferProvider>()
        ?.cartTransferModel;
  }

  @override
  bool updateShouldNotify(CartTransferProvider oldWidget) {
    return cartTransferModel != oldWidget.cartTransferModel;
  }
}

/// Model รายการสินค้าสำหรับส่งข้อมูล
class CartTransferItemModel {
  final String itemId;
  final String itemCode; // รหัสสินค้า
  final String productName;
  final String barcode;
  final String unitCode;
  final String unitName;
  final double quantity;
  final double price; // ราคาต่อหน่วย
  final double totalPrice;
  final double discount; // ส่วนลด (%)
  final double discountAmount;
  final String? remark;

  // ข้อมูลคลัง (ถ้ามี)
  final String? warehouseId;
  final String? warehouseName;
  final String? locationId;
  final String? locationName;

  // ข้อมูลเพิ่มเติม
  final int unitStand;
  final int unitDive;
  final String? imageuri;

  CartTransferItemModel({
    required this.itemId,
    required this.itemCode,
    required this.productName,
    required this.barcode,
    required this.unitCode,
    required this.unitName,
    required this.quantity,
    required this.price,
    required this.totalPrice,
    required this.discount,
    required this.discountAmount,
    this.remark,
    this.warehouseId,
    this.warehouseName,
    this.locationId,
    this.locationName,
    required this.unitStand,
    required this.unitDive,
    this.imageuri,
  });

  /// สร้าง CartTransferItemModel จาก CartItemModel
  factory CartTransferItemModel.fromCartItemModel(CartItemModel item) {
    return CartTransferItemModel(
      itemId: item.itemId,
      itemCode: item.itemCode,
      productName: item.productName,
      barcode: item.barcode,
      unitCode: item.unitCode,
      unitName: item.unitName,
      quantity: item.quantity,
      price: item.price,
      totalPrice: item.totalPrice,
      discount: item.discount,
      discountAmount: item.discountAmount,
      remark: item.remark,
      warehouseId: item.warehouseId,
      warehouseName: item.warehouseName,
      locationId: item.locationId,
      locationName: item.locationName,
      unitStand: item.unitStand,
      unitDive: item.unitDive,
      imageuri: item.imageuri,
    );
  }
}
