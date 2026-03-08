import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Helper สำหรับ parse วันที่จากฐานข้อมูล (UTC+0) ให้ถูกต้อง
DateTime? parseUtcDateTimeNullable(String? dateStr) {
  if (dateStr == null || dateStr.isEmpty) return null;
  // ถ้าไม่มี Z ต่อท้าย และไม่มี timezone offset ให้เติม Z เพื่อบอก Dart ว่าเป็น UTC
  if (!dateStr.endsWith('Z') && !dateStr.contains('+')) {
    dateStr = '${dateStr}Z';
  }
  return DateTime.tryParse(dateStr)?.toUtc();
}

DateTime parseUtcDateTime(String? dateStr) {
  return parseUtcDateTimeNullable(dateStr) ?? DateTime.now().toUtc();
}

/// ประเภทการจัดซื้อ (Purchase Type) - รองรับหลายภาษา
class PurchaseTypeModel {
  final String? guid;
  final String code;
  final List<LanguageDataModel> names; // ชื่อหลายภาษา
  final List<LanguageDataModel> descriptions; // รายละเอียดหลายภาษา
  final bool isActive;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  PurchaseTypeModel({
    this.guid,
    required this.code,
    required this.names,
    this.descriptions = const [],
    this.isActive = true,
    this.createdAt,
    this.updatedAt,
  });

  /// ดึงชื่อตามภาษาปัจจุบัน
  String getName(String langCode) {
    if (names.isEmpty) return '';
    final found = names.where((n) => n.code == langCode).firstOrNull;
    return found?.name ?? names.first.name;
  }

  /// ดึงรายละเอียดตามภาษาปัจจุบัน
  String getDescription(String langCode) {
    if (descriptions.isEmpty) return '';
    final found = descriptions.where((d) => d.code == langCode).firstOrNull;
    return found?.name ?? descriptions.first.name;
  }

  factory PurchaseTypeModel.fromJson(Map<String, dynamic> json) {
    // รองรับทั้ง format เก่า (name: String) และ format ใหม่ (names: [])
    List<LanguageDataModel> namesList = [];
    if (json['names'] != null && json['names'] is List) {
      namesList = (json['names'] as List)
          .map((item) => LanguageDataModel.fromJson(item))
          .toList();
    } else if (json['name'] != null) {
      // รองรับ format เก่า (name: String)
      namesList = [LanguageDataModel(code: 'th', name: json['name'])];
    }

    List<LanguageDataModel> descriptionsList = [];
    if (json['descriptions'] != null && json['descriptions'] is List) {
      descriptionsList = (json['descriptions'] as List)
          .map((item) => LanguageDataModel.fromJson(item))
          .toList();
    } else if (json['description'] != null) {
      // รองรับ format เก่า (description: String)
      descriptionsList = [LanguageDataModel(code: 'th', name: json['description'])];
    }

    return PurchaseTypeModel(
      guid: json['guid'] ?? json['_id'],
      code: json['code'] ?? '',
      names: namesList,
      descriptions: descriptionsList,
      isActive: json['is_active'] ?? true,
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
      if (guid != null) 'guid': guid,
      'code': code,
      'names': names.map((n) => n.toJson()).toList(),
      'descriptions': descriptions.map((d) => d.toJson()).toList(),
      'is_active': isActive,
    };
  }
}

/// ข้อมูลผู้อนุมัติ (Approver Info) - เก็บข้อมูลผู้มีสิทธิ์อนุมัติในแต่ละกฎ
class ApproverInfoModel {
  final String userCode; // รหัสผู้ใช้
  final String userName; // ชื่อผู้ใช้
  final String? email; // อีเมล (ถ้ามี)
  final String? lineUserId; // LINE User ID (ถ้ามี)
  final String? lineDisplayName; // ชื่อ LINE (ถ้ามี)
  final String? position; // ตำแหน่ง (ถ้ามี)
  final String? department; // แผนก (ถ้ามี)

  ApproverInfoModel({
    required this.userCode,
    required this.userName,
    this.email,
    this.lineUserId,
    this.lineDisplayName,
    this.position,
    this.department,
  });

  factory ApproverInfoModel.fromJson(Map<String, dynamic> json) {
    return ApproverInfoModel(
      userCode: json['user_code'] ?? json['username'] ?? '',
      userName: json['user_name'] ?? json['editusername'] ?? '',
      email: json['email'],
      lineUserId: json['line_user_id'],
      lineDisplayName: json['line_display_name'],
      position: json['position'],
      department: json['department'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user_code': userCode,
      'user_name': userName,
      if (email != null && email!.isNotEmpty) 'email': email,
      if (lineUserId != null && lineUserId!.isNotEmpty) 'line_user_id': lineUserId,
      if (lineDisplayName != null && lineDisplayName!.isNotEmpty) 'line_display_name': lineDisplayName,
      if (position != null && position!.isNotEmpty) 'position': position,
      if (department != null && department!.isNotEmpty) 'department': department,
    };
  }

  /// ตรวจสอบว่ามี LINE หรือไม่
  bool get hasLine => lineUserId != null && lineUserId!.isNotEmpty;

  /// ตรวจสอบว่ามี email หรือไม่
  bool get hasEmail => email != null && email!.isNotEmpty;

  /// สร้าง ApproverInfoModel จาก UserModel
  static ApproverInfoModel fromUserModel(dynamic user) {
    return ApproverInfoModel(
      userCode: user.username ?? '',
      userName: user.editusername ?? user.username ?? '',
      email: user.email,
      lineUserId: user.lineUserId,
      lineDisplayName: user.lineDisplayName,
      position: user.position,
      department: user.department,
    );
  }
}

/// กฎการอนุมัติ (Approval Rule) - ตามระดับราคา
class ApprovalRuleModel {
  final double minAmount; // วงเงินขั้นต่ำ (บาท)
  final double maxAmount; // วงเงินสูงสุด (บาท) - 0 = ไม่จำกัด
  final int approvalLevel; // ระดับผู้อนุมัติที่ต้องการ (0=ไม่ต้องอนุมัติ, 1-4=ระดับ)
  final String approvalLevelName; // ชื่อระดับ เช่น หัวหน้าแผนก, ผู้จัดการ
  final List<ApproverInfoModel> approvers; // รายการผู้มีสิทธิ์อนุมัติ (ใครก็ได้ในรายการ)

  ApprovalRuleModel({
    required this.minAmount,
    required this.maxAmount,
    required this.approvalLevel,
    required this.approvalLevelName,
    this.approvers = const [],
  });

  factory ApprovalRuleModel.fromJson(Map<String, dynamic> json) {
    final approversList = (json['approvers'] as List?)
            ?.map((item) => ApproverInfoModel.fromJson(item))
            .toList() ??
        [];

    return ApprovalRuleModel(
      minAmount: global.safeToDouble(json['min_amount']),
      maxAmount: global.safeToDouble(json['max_amount']),
      approvalLevel: json['approval_level'] ?? 0,
      approvalLevelName: json['approval_level_name'] ?? '',
      approvers: approversList,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'min_amount': minAmount,
      'max_amount': maxAmount,
      'approval_level': approvalLevel,
      'approval_level_name': approvalLevelName,
      'approvers': approvers.map((a) => a.toJson()).toList(),
    };
  }

  /// สร้าง copy พร้อมค่าใหม่
  ApprovalRuleModel copyWith({
    double? minAmount,
    double? maxAmount,
    int? approvalLevel,
    String? approvalLevelName,
    List<ApproverInfoModel>? approvers,
  }) {
    return ApprovalRuleModel(
      minAmount: minAmount ?? this.minAmount,
      maxAmount: maxAmount ?? this.maxAmount,
      approvalLevel: approvalLevel ?? this.approvalLevel,
      approvalLevelName: approvalLevelName ?? this.approvalLevelName,
      approvers: approvers ?? this.approvers,
    );
  }
}

// =====================================================
// PO Approval Status (สถานะการอนุมัติ PO)
// =====================================================

/// สถานะการอนุมัติ
enum POApprovalStatus {
  draft, // ร่าง - ยังไม่ส่งอนุมัติ
  autoApproved, // อนุมัติอัตโนมัติ (ไม่มีกฎอนุมัติ/ผู้บันทึกมีสิทธิ์เพียงพอ)
  pending, // รออนุมัติ
  approved, // อนุมัติแล้ว
  rejected, // ไม่อนุมัติ
}

/// ประวัติการอนุมัติ (Approval History)
class ApprovalHistoryModel {
  final String action; // submit, approve, reject
  final String actionBy; // รหัสผู้ดำเนินการ
  final String actionByName; // ชื่อผู้ดำเนินการ
  final int level; // ระดับการอนุมัติ
  final String? comment; // หมายเหตุ
  final DateTime actionAt; // เวลาดำเนินการ

  ApprovalHistoryModel({
    required this.action,
    required this.actionBy,
    required this.actionByName,
    required this.level,
    this.comment,
    required this.actionAt,
  });

  factory ApprovalHistoryModel.fromJson(Map<String, dynamic> json) {
    return ApprovalHistoryModel(
      action: json['action'] ?? '',
      actionBy: json['action_by'] ?? '',
      actionByName: json['action_by_name'] ?? '',
      level: json['level'] ?? 0,
      comment: json['comment'],
      actionAt: parseUtcDateTime(json['action_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'action': action,
      'action_by': actionBy,
      'action_by_name': actionByName,
      'level': level,
      'comment': comment,
      'action_at': actionAt.toIso8601String(),
    };
  }
}

/// สถานะการอนุมัติ PO (PO Approval Status)
class POApprovalStatusModel {
  final String? guid;
  final String docNo; // เลขที่เอกสาร
  final String guidFixed; // GUID ของเอกสาร
  final String? sourceDocNo; // เลขที่เอกสารต้นแบบ (กรณีคัดลอกจาก PO ที่ถูกปฏิเสธ)
  final String? sourceGuidFixed; // GUID ของเอกสารต้นแบบ
  final String purchaseTypeCode; // รหัสประเภทการจัดซื้อ
  final String purchaseTypeName; // ชื่อประเภทการจัดซื้อ
  final double totalAmount; // ยอดรวม (บาท)
  final int requiredLevel; // ระดับการอนุมัติที่ต้องการ (0=ไม่ต้องอนุมัติ)
  final String requiredLevelName; // ชื่อระดับที่ต้องการ
  final int currentApprovedLevel; // ระดับที่อนุมัติแล้ว
  final POApprovalStatus status; // สถานะปัจจุบัน
  final String createdBy; // ผู้สร้างเอกสาร
  final String createdByName; // ชื่อผู้สร้าง
  final List<ApprovalHistoryModel> history; // ประวัติการอนุมัติ
  final String? lastComment; // หมายเหตุล่าสุด (กรณีปฏิเสธ)
  final DateTime? createdAt;
  final DateTime? updatedAt;

  POApprovalStatusModel({
    this.guid,
    required this.docNo,
    required this.guidFixed,
    this.sourceDocNo,
    this.sourceGuidFixed,
    required this.purchaseTypeCode,
    required this.purchaseTypeName,
    required this.totalAmount,
    required this.requiredLevel,
    required this.requiredLevelName,
    this.currentApprovedLevel = 0,
    required this.status,
    required this.createdBy,
    required this.createdByName,
    this.history = const [],
    this.lastComment,
    this.createdAt,
    this.updatedAt,
  });

  factory POApprovalStatusModel.fromJson(Map<String, dynamic> json) {
    final historyList = (json['history'] as List?)
            ?.map((item) => ApprovalHistoryModel.fromJson(item))
            .toList() ??
        [];

    return POApprovalStatusModel(
      guid: json['guid'] ?? json['_id'],
      docNo: json['docno'] ?? '',
      guidFixed: json['guidfixed'] ?? '',
      sourceDocNo: json['source_docno'],
      sourceGuidFixed: json['source_guidfixed'],
      purchaseTypeCode: json['purchase_type_code'] ?? '',
      purchaseTypeName: json['purchase_type_name'] ?? '',
      totalAmount: global.safeToDouble(json['total_amount']),
      requiredLevel: json['required_level'] ?? 0,
      requiredLevelName: json['required_level_name'] ?? '',
      currentApprovedLevel: json['current_approved_level'] ?? 0,
      status: _parseStatus(json['status']),
      createdBy: json['created_by'] ?? '',
      createdByName: json['created_by_name'] ?? '',
      history: historyList,
      lastComment: json['last_comment'],
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'])
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.tryParse(json['updated_at'])
          : null,
    );
  }

  static POApprovalStatus _parseStatus(String? status) {
    switch (status) {
      case 'draft':
        return POApprovalStatus.draft;
      case 'auto_approved':
        return POApprovalStatus.autoApproved;
      case 'pending':
        return POApprovalStatus.pending;
      case 'approved':
        return POApprovalStatus.approved;
      case 'rejected':
        return POApprovalStatus.rejected;
      default:
        return POApprovalStatus.draft;
    }
  }

  static String _statusToString(POApprovalStatus status) {
    switch (status) {
      case POApprovalStatus.draft:
        return 'draft';
      case POApprovalStatus.autoApproved:
        return 'auto_approved';
      case POApprovalStatus.pending:
        return 'pending';
      case POApprovalStatus.approved:
        return 'approved';
      case POApprovalStatus.rejected:
        return 'rejected';
    }
  }

  Map<String, dynamic> toJson() {
    return {
      if (guid != null) 'guid': guid,
      'docno': docNo,
      'guidfixed': guidFixed,
      if (sourceDocNo != null) 'source_docno': sourceDocNo,
      if (sourceGuidFixed != null) 'source_guidfixed': sourceGuidFixed,
      'purchase_type_code': purchaseTypeCode,
      'purchase_type_name': purchaseTypeName,
      'total_amount': totalAmount,
      'required_level': requiredLevel,
      'required_level_name': requiredLevelName,
      'current_approved_level': currentApprovedLevel,
      'status': _statusToString(status),
      'created_by': createdBy,
      'created_by_name': createdByName,
      'history': history.map((h) => h.toJson()).toList(),
      'last_comment': lastComment,
    };
  }

  /// ตรวจสอบว่าสามารถส่งอนุมัติได้หรือไม่
  bool get canSubmit =>
      status == POApprovalStatus.draft || status == POApprovalStatus.rejected;

  /// ตรวจสอบว่าอนุมัติเสร็จสิ้นแล้วหรือไม่
  bool get isCompleted =>
      status == POApprovalStatus.approved ||
      status == POApprovalStatus.autoApproved;

  /// ตรวจสอบว่าถูกปฏิเสธหรือไม่
  bool get isRejected => status == POApprovalStatus.rejected;

  /// รับชื่อสถานะภาษาไทย
  String get statusDisplayName {
    switch (status) {
      case POApprovalStatus.draft:
        return global.language('draft');
      case POApprovalStatus.autoApproved:
        return 'อนุมัติแล้ว (อัตโนมัติ)';
      case POApprovalStatus.pending:
        return global.language('pending_approval');
      case POApprovalStatus.approved:
        return global.language('approval_approved');
      case POApprovalStatus.rejected:
        return global.language('reject');
    }
  }
}

/// การตั้งค่าการอนุมัติใบสั่งซื้อ (PO Approval Setting)
class POApprovalSettingModel {
  final String? guid;
  final String purchaseTypeCode; // รหัสประเภทการจัดซื้อ
  final String purchaseTypeName; // ชื่อประเภทการจัดซื้อ
  final List<ApprovalRuleModel> rules; // กฎการอนุมัติตามวงเงิน
  final bool isActive;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  POApprovalSettingModel({
    this.guid,
    required this.purchaseTypeCode,
    required this.purchaseTypeName,
    required this.rules,
    this.isActive = true,
    this.createdAt,
    this.updatedAt,
  });

  factory POApprovalSettingModel.fromJson(Map<String, dynamic> json) {
    final rulesList = (json['rules'] as List?)
            ?.map((item) => ApprovalRuleModel.fromJson(item))
            .toList() ??
        [];

    return POApprovalSettingModel(
      guid: json['guid'] ?? json['_id'],
      purchaseTypeCode: json['purchase_type_code'] ?? '',
      purchaseTypeName: json['purchase_type_name'] ?? '',
      rules: rulesList,
      isActive: json['is_active'] ?? true,
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
      if (guid != null) 'guid': guid,
      'purchase_type_code': purchaseTypeCode,
      'purchase_type_name': purchaseTypeName,
      'rules': rules.map((r) => r.toJson()).toList(),
      'is_active': isActive,
    };
  }
}

// =====================================================
// Approval Timeline (ประวัติการอนุมัติรวม)
// =====================================================

/// รายการใน Timeline
class ApprovalTimelineItem {
  final String action; // submit, modify, approve, reject, notification_sent, notification_opened
  final String actionBy; // รหัสผู้ดำเนินการ
  final String actionByName; // ชื่อผู้ดำเนินการ
  final DateTime actionAt; // เวลาดำเนินการ
  final String? detail; // รายละเอียดเพิ่มเติม
  final String? comment; // หมายเหตุ
  final int? level; // ระดับการอนุมัติ
  final String? type; // ประเภท notification: email, line
  final String? recipient; // ผู้รับ notification

  ApprovalTimelineItem({
    required this.action,
    required this.actionBy,
    required this.actionByName,
    required this.actionAt,
    this.detail,
    this.comment,
    this.level,
    this.type,
    this.recipient,
  });

  factory ApprovalTimelineItem.fromJson(Map<String, dynamic> json) {
    return ApprovalTimelineItem(
      action: json['action'] ?? '',
      actionBy: json['action_by'] ?? '',
      actionByName: json['action_by_name'] ?? '',
      actionAt: parseUtcDateTime(json['action_at']),
      detail: json['detail'],
      comment: json['comment'],
      level: json['level'],
      type: json['type'],
      recipient: json['recipient'],
    );
  }

  /// ดึงชื่อ action เป็นภาษาไทย
  String get actionDisplayName {
    switch (action) {
      case 'submit':
        return 'ส่งเอกสาร';
      case 'modify':
        return global.language('edit_document');
      case 'approve':
        return global.language('approve');
      case 'reject':
        return global.language('reject');
      case 'auto_approve':
        return global.language('approval_auto_approved');
      case 'notification_sent':
        return 'ส่งแจ้งเตือน';
      case 'notification_opened':
        return global.language('open_read');
      default:
        return action;
    }
  }
}

/// ผลลัพธ์การดึง Timeline
class ApprovalTimelineResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<ApprovalTimelineItem> timeline;
  final Map<String, dynamic>? status;

  ApprovalTimelineResult({
    required this.isSuccess,
    this.errorMessage,
    this.timeline = const [],
    this.status,
  });

  factory ApprovalTimelineResult.fromJson(Map<String, dynamic> json) {
    final timelineList = (json['timeline'] as List?)
            ?.map((item) => ApprovalTimelineItem.fromJson(item))
            .toList() ??
        [];

    return ApprovalTimelineResult(
      isSuccess: true,
      timeline: timelineList,
      status: json['status'],
    );
  }

  factory ApprovalTimelineResult.error(String message) {
    return ApprovalTimelineResult(
      isSuccess: false,
      errorMessage: message,
    );
  }
}
