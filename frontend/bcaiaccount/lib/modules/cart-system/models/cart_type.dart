import 'cart_system_type.dart';
import '../../../global.dart' as global;

/// ประเภทของตะกร้า - แยกตาม CartSystemType
///
/// แต่ละระบบมีประเภทตะกร้าที่แตกต่างกัน:
/// - ระบบขาย (Sales): ขาย, สั่งขาย, เสนอราคา
/// - ระบบซื้อ (Purchase): สั่งซื้อ
/// - ระบบคงคลัง (Inventory): เบิก, โอน
enum CartType {
  // ระบบขาย (Sales)
  sale, // ขาย
  salesOrder, // สั่งขาย
  quotation, // เสนอราคา

  // ระบบซื้อ (Purchase)
  purchaseOrder, // สั่งซื้อ

  // ระบบคงคลัง (Inventory)
  withdraw, // เบิก
  transfer, // โอน
}

extension CartTypeExtension on CartType {
  /// ชื่อแสดงผล (รองรับหลายภาษา)
  String get displayNameThai {
    switch (this) {
      case CartType.sale:
        return global.language("cart_type_sale");
      case CartType.salesOrder:
        return global.language("cart_type_sales_order");
      case CartType.quotation:
        return global.language("cart_type_quotation");
      case CartType.purchaseOrder:
        return global.language("cart_type_purchase_order");
      case CartType.withdraw:
        return global.language("cart_type_withdraw");
      case CartType.transfer:
        return global.language("cart_type_transfer");
    }
  }

  /// ชื่อแสดงผลภาษาอังกฤษ
  String get displayNameEnglish {
    switch (this) {
      case CartType.sale:
        return 'Sale';
      case CartType.salesOrder:
        return 'Sales Order';
      case CartType.quotation:
        return 'Quotation';
      case CartType.purchaseOrder:
        return 'Purchase Order';
      case CartType.withdraw:
        return 'Withdraw';
      case CartType.transfer:
        return 'Transfer';
    }
  }

  /// แปลงเป็น JSON string
  String toJson() {
    switch (this) {
      case CartType.sale:
        return 'sale';
      case CartType.salesOrder:
        return 'sales_order';
      case CartType.quotation:
        return 'quotation';
      case CartType.purchaseOrder:
        return 'purchase_order';
      case CartType.withdraw:
        return 'withdraw';
      case CartType.transfer:
        return 'transfer';
    }
  }

  /// ตรวจสอบว่าสามารถใช้กับระบบนี้ได้หรือไม่
  bool isValidForSystemType(CartSystemType systemType) {
    switch (systemType) {
      case CartSystemType.sales:
        return this == CartType.sale ||
            this == CartType.salesOrder ||
            this == CartType.quotation;
      case CartSystemType.purchase:
        return this == CartType.purchaseOrder;
      case CartSystemType.inventory:
        return this == CartType.withdraw || this == CartType.transfer;
    }
  }

  /// ระบบที่ cart type นี้ใช้ได้
  CartSystemType get systemType {
    switch (this) {
      case CartType.sale:
      case CartType.salesOrder:
      case CartType.quotation:
        return CartSystemType.sales;
      case CartType.purchaseOrder:
        return CartSystemType.purchase;
      case CartType.withdraw:
      case CartType.transfer:
        return CartSystemType.inventory;
    }
  }
}

/// Helper function: แปลง JSON string กลับเป็น CartType
CartType? cartTypeFromJson(String? json) {
  if (json == null) return null;

  switch (json) {
    case 'sale':
      return CartType.sale;
    case 'sales_order':
      return CartType.salesOrder;
    case 'quotation':
      return CartType.quotation;
    case 'purchase_order':
      return CartType.purchaseOrder;
    case 'withdraw':
      return CartType.withdraw;
    case 'transfer':
      return CartType.transfer;
    default:
      return null;
  }
}

/// Helper function: รับรายการ CartType ทั้งหมดที่ใช้ได้กับระบบนี้
List<CartType> getCartTypesForSystem(CartSystemType systemType) {
  switch (systemType) {
    case CartSystemType.sales:
      return [
        CartType.sale,
        CartType.salesOrder,
        CartType.quotation,
      ];
    case CartSystemType.purchase:
      return [
        CartType.purchaseOrder,
      ];
    case CartSystemType.inventory:
      return [
        CartType.withdraw,
        CartType.transfer,
      ];
  }
}

/// Helper function: ดึง CartType default สำหรับแต่ละระบบ
CartType getDefaultCartType(CartSystemType systemType) {
  switch (systemType) {
    case CartSystemType.sales:
      return CartType.sale;
    case CartSystemType.purchase:
      return CartType.purchaseOrder;
    case CartSystemType.inventory:
      return CartType.withdraw;
  }
}
