/// Model สำหรับคลังสินค้าจาก ClickHouse ({CH_DATABASE_NAME}.warehouses)
class ClickHouseWarehouseModel {
  final String shopId;
  final String guidFixed;
  final String code;
  final String name1;
  final String name2;

  ClickHouseWarehouseModel({
    required this.shopId,
    required this.guidFixed,
    required this.code,
    required this.name1,
    required this.name2,
  });

  /// ชื่อเต็ม (name1 + name2)
  String get fullName {
    if (name2.isEmpty) return name1;
    return '$name1 $name2';
  }

  /// ชื่อที่ใช้แสดง (name1 หรือ fullName ถ้า name2 ไม่ว่าง)
  String get displayName => fullName;

  /// สร้าง model จาก JSON
  factory ClickHouseWarehouseModel.fromJson(Map<String, dynamic> json) {
    return ClickHouseWarehouseModel(
      shopId: json['shopid']?.toString() ?? '',
      guidFixed: json['guidfixed']?.toString() ?? '',
      code: json['code']?.toString() ?? '',
      name1: json['name1']?.toString() ?? '',
      name2: json['name2']?.toString() ?? '',
    );
  }

  /// แปลงเป็น JSON
  Map<String, dynamic> toJson() {
    return {
      'shopid': shopId,
      'guidfixed': guidFixed,
      'code': code,
      'name1': name1,
      'name2': name2,
    };
  }

  /// แปลงเป็น JSON สำหรับบันทึกใน Cart (เก็บเฉพาะข้อมูลที่จำเป็น)
  Map<String, dynamic> toCartJson() {
    return {
      'guidfixed': guidFixed,
      'code': code,
      'name': displayName,
    };
  }

  @override
  String toString() {
    return 'ClickHouseWarehouseModel(shopId: $shopId, code: $code, name: $displayName)';
  }
}
