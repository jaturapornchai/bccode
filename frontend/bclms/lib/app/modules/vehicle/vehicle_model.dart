/// โมเดลข้อมูลยานพาหนะ — ตาม Backend API schema
class VehicleModel {
  final String guidFixed;
  final String code;
  final List<VehicleName> names;
  final String vehicleType;
  final String licensePlate;
  final String province;
  final String brand;
  final String model;
  final String color;
  final int year;
  final double? capacityWeight;
  final double? capacityVolume;
  final String fuelType;
  final int status;
  final String driverCode;
  final String driverName;
  final DateTime? insuranceExpiry;
  final DateTime? registrationExpiry;
  final String notes;
  final String imageUri;

  // Metadata
  final String? createdBy;
  final DateTime? createdAt;
  final String? updatedBy;
  final DateTime? updatedAt;

  const VehicleModel({
    required this.guidFixed,
    required this.code,
    this.names = const [],
    this.vehicleType = '',
    this.licensePlate = '',
    this.province = '',
    this.brand = '',
    this.model = '',
    this.color = '',
    this.year = 0,
    this.capacityWeight,
    this.capacityVolume,
    this.fuelType = '',
    this.status = 1,
    this.driverCode = '',
    this.driverName = '',
    this.insuranceExpiry,
    this.registrationExpiry,
    this.notes = '',
    this.imageUri = '',
    this.createdBy,
    this.createdAt,
    this.updatedBy,
    this.updatedAt,
  });

  /// ชื่อยานพาหนะ (ดึงจาก names ภาษาไทย)
  String get displayName {
    if (names.isEmpty) return '';
    final th = names.where((n) => n.code == 'th').firstOrNull;
    return th?.name ?? names.first.name;
  }

  factory VehicleModel.fromJson(Map<String, dynamic> json) {
    return VehicleModel(
      guidFixed: json['guidfixed'] ?? json['_id'] ?? '',
      code: json['code'] ?? '',
      names: (json['names'] as List?)
              ?.map((e) => VehicleName.fromJson(e))
              .toList() ??
          [],
      vehicleType: json['vehicletype'] ?? '',
      licensePlate: json['licenseplate'] ?? '',
      province: json['province'] ?? '',
      brand: json['brand'] ?? '',
      model: json['model'] ?? '',
      color: json['color'] ?? '',
      year: json['year'] ?? 0,
      capacityWeight: (json['capacityweight'] as num?)?.toDouble(),
      capacityVolume: (json['capacityvolume'] as num?)?.toDouble(),
      fuelType: json['fueltype'] ?? '',
      status: json['status'] ?? 1,
      driverCode: json['drivercode'] ?? '',
      driverName: json['drivername'] ?? '',
      insuranceExpiry: _parseDate(json['insuranceexpiry']),
      registrationExpiry: _parseDate(json['registrationexpiry']),
      notes: json['notes'] ?? '',
      imageUri: json['imageuri'] ?? '',
      createdBy: json['createdby'],
      createdAt: _parseDate(json['createdat']),
      updatedBy: json['updatedby'],
      updatedAt: _parseDate(json['updatedat']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'guidfixed': guidFixed,
      'code': code,
      'names': names.map((n) => n.toJson()).toList(),
      'vehicletype': vehicleType,
      'licenseplate': licensePlate,
      'province': province,
      'brand': brand,
      'model': model,
      'color': color,
      'year': year,
      if (capacityWeight != null) 'capacityweight': capacityWeight,
      if (capacityVolume != null) 'capacityvolume': capacityVolume,
      'fueltype': fuelType,
      'status': status,
      'drivercode': driverCode,
      'drivername': driverName,
      if (insuranceExpiry != null)
        'insuranceexpiry': insuranceExpiry!.toIso8601String(),
      if (registrationExpiry != null)
        'registrationexpiry': registrationExpiry!.toIso8601String(),
      'notes': notes,
      'imageuri': imageUri,
    };
  }

  static DateTime? _parseDate(dynamic value) {
    if (value == null) return null;
    if (value is String && value.isNotEmpty) {
      return DateTime.tryParse(value);
    }
    return null;
  }

  /// ชื่อประเภทรถ (แปลง string → ภาษาไทย)
  String get vehicleTypeName {
    switch (vehicleType) {
      case 'truck':
        return 'รถบรรทุก';
      case 'van':
        return 'รถตู้';
      case 'motorcycle':
        return 'มอเตอร์ไซค์';
      case 'pickup':
        return 'รถกระบะ';
      case 'trailer':
        return 'หัวลาก';
      case 'sixwheel':
        return 'รถ 6 ล้อ';
      case 'tenwheel':
        return 'รถ 10 ล้อ';
      default:
        return vehicleType.isNotEmpty ? vehicleType : 'ไม่ระบุ';
    }
  }

  /// ชื่อสถานะ (0=ไม่ใช้งาน, 1=ใช้งาน, 2=ซ่อมบำรุง)
  String get statusName {
    switch (status) {
      case 0:
        return 'ไม่ใช้งาน';
      case 1:
        return 'ใช้งาน';
      case 2:
        return 'ซ่อมบำรุง';
      default:
        return 'ไม่ระบุ';
    }
  }

  /// ชื่อประเภทเชื้อเพลิง
  String get fuelTypeName {
    switch (fuelType) {
      case 'diesel':
        return 'ดีเซล';
      case 'gasoline':
        return 'เบนซิน';
      case 'electric':
        return 'ไฟฟ้า (EV)';
      case 'lpg':
        return 'LPG';
      case 'cng':
        return 'CNG';
      case 'hybrid':
        return 'ไฮบริด';
      default:
        return fuelType.isNotEmpty ? fuelType : 'ไม่ระบุ';
    }
  }
}

/// ชื่อยานพาหนะ (multilingual)
class VehicleName {
  final String code;
  final String name;

  const VehicleName({required this.code, required this.name});

  factory VehicleName.fromJson(Map<String, dynamic> json) {
    return VehicleName(
      code: json['code'] ?? '',
      name: json['name'] ?? '',
    );
  }

  Map<String, dynamic> toJson() => {'code': code, 'name': name};
}

/// โหมดหน้าจอ
enum ScreenMode { list, view, add, edit, multiSelect }
