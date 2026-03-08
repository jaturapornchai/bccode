import '../../../global.dart' as global;

/// ประเภทระบบตะกร้า
enum CartSystemType {
  sales, // ระบบขาย - ใช้กับลูกหนี้ (Debtor)
  purchase, // ระบบซื้อ - ใช้กับเจ้าหนี้ (Creditor)
  inventory, // ระบบสินค้าคงคลัง - ไม่ต้องระบุลูกหนี้/เจ้าหนี้
}

/// Extension สำหรับแปลง enum เป็น string และกลับกัน
extension CartSystemTypeExtension on CartSystemType {
  /// แปลง enum เป็น string สำหรับบันทึกใน database
  String toJson() {
    switch (this) {
      case CartSystemType.sales:
        return 'sales';
      case CartSystemType.purchase:
        return 'purchase';
      case CartSystemType.inventory:
        return 'inventory';
    }
  }

  /// ชื่อสำหรับแสดงผล (รองรับหลายภาษา)
  String get displayNameThai {
    switch (this) {
      case CartSystemType.sales:
        return global.language("system_sales");
      case CartSystemType.purchase:
        return global.language("system_purchase");
      case CartSystemType.inventory:
        return global.language("system_inventory");
    }
  }

  /// Icon สำหรับแสดงใน tab
  String get iconName {
    switch (this) {
      case CartSystemType.sales:
        return 'shopping_cart'; // ไอคอนรถเข็น
      case CartSystemType.purchase:
        return 'add_shopping_cart'; // ไอคอนเพิ่มของในรถเข็น
      case CartSystemType.inventory:
        return 'inventory_2'; // ไอคอนคลังสินค้า
    }
  }
}

/// Helper function สำหรับแปลง string เป็น enum
CartSystemType cartSystemTypeFromJson(String value) {
  switch (value.toLowerCase()) {
    case 'sales':
      return CartSystemType.sales;
    case 'purchase':
      return CartSystemType.purchase;
    case 'inventory':
      return CartSystemType.inventory;
    default:
      throw ArgumentError('Invalid CartSystemType: $value');
  }
}
