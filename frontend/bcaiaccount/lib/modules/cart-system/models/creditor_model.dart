/// Model สำหรับเจ้าหนี้ (Creditor) - ใช้ในระบบซื้อ
class CreditorModel {
  final String shopId;
  final String guidFixed; // GUID ของเจ้าหนี้
  final String code; // รหัสเจ้าหนี้
  final String name1; // ชื่อหลัก
  final String name2; // ชื่อรอง (ถ้ามี)

  CreditorModel({
    required this.shopId,
    required this.guidFixed,
    required this.code,
    required this.name1,
    this.name2 = '',
  });

  /// ชื่อเต็ม (รวม name1 และ name2)
  String get fullName {
    if (name2.isNotEmpty) {
      return '$name1 $name2';
    }
    return name1;
  }

  /// ชื่อสำหรับการแสดงผล (รหัส - ชื่อ)
  String get displayName => '$code - $fullName';

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

  /// สร้างจาก JSON (จาก ClickHouse)
  factory CreditorModel.fromJson(Map<String, dynamic> json) {
    return CreditorModel(
      shopId: json['shopid'] as String? ?? '',
      guidFixed: json['guidfixed'] as String? ?? '',
      code: json['code'] as String? ?? '',
      name1: json['name1'] as String? ?? '',
      name2: json['name2'] as String? ?? '',
    );
  }

  /// สำหรับบันทึกใน cart (simplified version)
  Map<String, dynamic> toCartJson() {
    return {
      'creditorId': guidFixed,
      'creditorCode': code,
      'creditorName': fullName,
    };
  }

  @override
  String toString() {
    return 'CreditorModel(code: $code, name: $fullName)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is CreditorModel &&
        other.shopId == shopId &&
        other.guidFixed == guidFixed;
  }

  @override
  int get hashCode => Object.hash(shopId, guidFixed);
}
