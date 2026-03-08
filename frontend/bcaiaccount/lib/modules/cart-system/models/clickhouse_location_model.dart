/// Model สำหรับที่เก็บสินค้าจาก ClickHouse ({CH_DATABASE_NAME}.locations)
class ClickHouseLocationModel {
  final String shopId;
  final String guidFixed;
  final String locationCode;
  final String warehouseCode;
  final String name1;
  final String name2;

  ClickHouseLocationModel({
    required this.shopId,
    required this.guidFixed,
    required this.locationCode,
    required this.warehouseCode,
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
  factory ClickHouseLocationModel.fromJson(Map<String, dynamic> json) {
    return ClickHouseLocationModel(
      shopId: json['shopid']?.toString() ?? '',
      guidFixed: json['guidfixed']?.toString() ?? '',
      locationCode: json['locationcode']?.toString() ?? '',
      warehouseCode: json['warehousecode']?.toString() ?? '',
      name1: json['name1']?.toString() ?? '',
      name2: json['name2']?.toString() ?? '',
    );
  }

  /// แปลงเป็น JSON
  Map<String, dynamic> toJson() {
    return {
      'shopid': shopId,
      'guidfixed': guidFixed,
      'locationcode': locationCode,
      'warehousecode': warehouseCode,
      'name1': name1,
      'name2': name2,
    };
  }

  /// แปลงเป็น JSON สำหรับบันทึกใน Cart (เก็บเฉพาะข้อมูลที่จำเป็น)
  Map<String, dynamic> toCartJson() {
    return {
      'guidfixed': guidFixed,
      'code': locationCode,
      'name': displayName,
    };
  }

  @override
  String toString() {
    return 'ClickHouseLocationModel(shopId: $shopId, code: $locationCode, name: $displayName, warehouse: $warehouseCode)';
  }
}
