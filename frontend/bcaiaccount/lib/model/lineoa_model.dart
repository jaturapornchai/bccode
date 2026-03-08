/// Line OA Configuration Model
/// ประเภทการเชื่อมต่อ Line OA สำหรับแต่ละระบบ
library;

enum LineOAType {
  sales('sales', 'ระบบขาย', 'แจ้งเตือนการขาย ใบเสร็จ โปรโมชั่น'),
  purchasing('purchasing', 'ระบบจัดซื้อ', 'แจ้งเตือนการสั่งซื้อ รับสินค้า'),
  stock('stock', 'ระบบสต๊อก', 'แจ้งเตือนสินค้าใกล้หมด สต๊อกต่ำ'),
  businessOwner('business_owner', 'ระบบเจ้าของกิจการ', 'รายงานยอดขาย กำไร สรุปธุรกิจ'),
  manager('manager', 'ระบบผู้จัดการ', 'แจ้งเตือนพนักงาน อนุมัติ ควบคุม');

  final String code;
  final String displayName;
  final String description;

  const LineOAType(this.code, this.displayName, this.description);

  static LineOAType? fromCode(String code) {
    for (var type in LineOAType.values) {
      if (type.code == code) return type;
    }
    return null;
  }
}

/// Line OA Configuration Model
class LineOAConfigModel {
  final String guid;
  final String shopId;
  final String lineOaType;
  final String channelId;
  final String channelSecret;
  final String accessToken;
  final String liffId;
  final bool isActive;
  final DateTime? lastTestAt;
  final bool lastTestSuccess;
  final String? lastTestMessage;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  LineOAConfigModel({
    required this.guid,
    required this.shopId,
    required this.lineOaType,
    required this.channelId,
    required this.channelSecret,
    required this.accessToken,
    this.liffId = '',
    this.isActive = false,
    this.lastTestAt,
    this.lastTestSuccess = false,
    this.lastTestMessage,
    this.createdAt,
    this.updatedAt,
  });

  factory LineOAConfigModel.empty(String shopId, LineOAType type) {
    return LineOAConfigModel(
      guid: '',
      shopId: shopId,
      lineOaType: type.code,
      channelId: '',
      channelSecret: '',
      accessToken: '',
      liffId: '',
      isActive: false,
    );
  }

  factory LineOAConfigModel.fromJson(Map<String, dynamic> json) {
    return LineOAConfigModel(
      guid: json['guid'] ?? '',
      shopId: json['shop_id'] ?? '',
      lineOaType: json['lineoa_type'] ?? '',
      channelId: json['channel_id'] ?? '',
      channelSecret: json['channel_secret'] ?? '',
      accessToken: json['access_token'] ?? '',
      liffId: json['liff_id'] ?? '',
      isActive: json['is_active'] ?? false,
      lastTestAt: json['last_test_at'] != null
          ? DateTime.tryParse(json['last_test_at'])
          : null,
      lastTestSuccess: json['last_test_success'] ?? false,
      lastTestMessage: json['last_test_message'],
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'])
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.tryParse(json['updated_at'])
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'guid': guid,
      'shop_id': shopId,
      'lineoa_type': lineOaType,
      'channel_id': channelId,
      'channel_secret': channelSecret,
      'access_token': accessToken,
      'liff_id': liffId,
      'is_active': isActive,
    };
  }

  LineOAConfigModel copyWith({
    String? guid,
    String? shopId,
    String? lineOaType,
    String? channelId,
    String? channelSecret,
    String? accessToken,
    String? liffId,
    bool? isActive,
    DateTime? lastTestAt,
    bool? lastTestSuccess,
    String? lastTestMessage,
  }) {
    return LineOAConfigModel(
      guid: guid ?? this.guid,
      shopId: shopId ?? this.shopId,
      lineOaType: lineOaType ?? this.lineOaType,
      channelId: channelId ?? this.channelId,
      channelSecret: channelSecret ?? this.channelSecret,
      accessToken: accessToken ?? this.accessToken,
      liffId: liffId ?? this.liffId,
      isActive: isActive ?? this.isActive,
      lastTestAt: lastTestAt ?? this.lastTestAt,
      lastTestSuccess: lastTestSuccess ?? this.lastTestSuccess,
      lastTestMessage: lastTestMessage ?? this.lastTestMessage,
      createdAt: createdAt,
      updatedAt: updatedAt,
    );
  }

  LineOAType? get type => LineOAType.fromCode(lineOaType);

  bool get isConfigured =>
      channelId.isNotEmpty &&
      channelSecret.isNotEmpty &&
      accessToken.isNotEmpty;
}

/// Line OA Employee Model - พนักงานที่เชื่อมต่อกับ Line OA
class LineOAEmployeeModel {
  final String guid;
  final String lineOaConfigGuid;
  final String employeeCode;
  final String employeeName;
  final String lineUserId;
  final String? displayName;
  final String? pictureUrl;
  final bool isActive;
  final DateTime? linkedAt;

  LineOAEmployeeModel({
    required this.guid,
    required this.lineOaConfigGuid,
    required this.employeeCode,
    required this.employeeName,
    required this.lineUserId,
    this.displayName,
    this.pictureUrl,
    this.isActive = true,
    this.linkedAt,
  });

  factory LineOAEmployeeModel.fromJson(Map<String, dynamic> json) {
    return LineOAEmployeeModel(
      guid: json['guid'] ?? '',
      lineOaConfigGuid: json['lineoa_config_guid'] ?? '',
      employeeCode: json['employee_code'] ?? '',
      employeeName: json['employee_name'] ?? '',
      lineUserId: json['line_user_id'] ?? '',
      displayName: json['display_name'],
      pictureUrl: json['picture_url'],
      isActive: json['is_active'] ?? true,
      linkedAt: json['linked_at'] != null
          ? DateTime.tryParse(json['linked_at'])
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'guid': guid,
      'lineoa_config_guid': lineOaConfigGuid,
      'employee_code': employeeCode,
      'employee_name': employeeName,
      'line_user_id': lineUserId,
      'display_name': displayName,
      'picture_url': pictureUrl,
      'is_active': isActive,
    };
  }
}

/// Test Connection Result
class LineOATestResult {
  final bool success;
  final String message;
  final Map<String, dynamic>? botInfo;

  LineOATestResult({
    required this.success,
    required this.message,
    this.botInfo,
  });

  factory LineOATestResult.fromJson(Map<String, dynamic> json) {
    return LineOATestResult(
      success: json['success'] ?? false,
      message: json['message'] ?? '',
      botInfo: json['bot_info'],
    );
  }
}
