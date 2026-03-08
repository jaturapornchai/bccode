import 'package:json_annotation/json_annotation.dart';

part 'user_model.g.dart';

/// ข้อมูลการอนุมัติแยกตามประเภทเอกสาร
@JsonSerializable()
class DocumentApprovalModel {
  @JsonKey(name: 'approval_role')
  int approvalRole;

  @JsonKey(name: 'max_approval_amount')
  double maxApprovalAmount;

  DocumentApprovalModel({
    this.approvalRole = 0,
    this.maxApprovalAmount = 0,
  });

  factory DocumentApprovalModel.fromJson(Map<String, dynamic> json) =>
      _$DocumentApprovalModelFromJson(json);
  Map<String, dynamic> toJson() => _$DocumentApprovalModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class UserModel {
  int role;
  String shopid;
  String username;
  String? editusername;
  String? email;
  @JsonKey(name: 'line_user_id')
  String? lineUserId;
  @JsonKey(name: 'line_display_name')
  String? lineDisplayName;
  @JsonKey(name: 'line_picture_url')
  String? linePictureUrl;

  // === ฟิลด์สำหรับระบบอนุมัติ ===
  /// ตำแหน่งงาน เช่น "หัวหน้าแผนก", "ผู้จัดการฝ่าย", "ผู้อำนวยการ", "กรรมการผู้จัดการ"
  String? position;

  /// แผนก เช่น "IT", "บัญชี", "จัดซื้อ", "การตลาด"
  String? department;

  /// บทบาทการอนุมัติ: 0=ไม่มีสิทธิ์, 1=ผู้อนุมัติระดับ1, 2=ผู้อนุมัติระดับ2, 3=ผู้อนุมัติระดับ3, 4=ผู้อนุมัติสูงสุด
  @JsonKey(name: 'approval_role')
  int? approvalRole;

  /// วงเงินอนุมัติสูงสุด (บาท) - 0 = ไม่จำกัด
  @JsonKey(name: 'max_approval_amount')
  double? maxApprovalAmount;

  /// ข้อมูลการอนุมัติใบสั่งซื้อ (Purchase Order)
  @JsonKey(name: 'po_approval')
  DocumentApprovalModel? poApproval;

  /// ข้อมูลการอนุมัติใบเสนอราคา (Quotation)
  @JsonKey(name: 'quotation_approval')
  DocumentApprovalModel? quotationApproval;

  UserModel({
    required this.role,
    required this.shopid,
    required this.username,
    String? editusername,
    String? email,
    this.lineUserId,
    this.lineDisplayName,
    this.linePictureUrl,
    this.position,
    this.department,
    this.approvalRole,
    this.maxApprovalAmount,
    this.poApproval,
    this.quotationApproval,
  })  : editusername = editusername ?? '',
        email = email ?? '';

  factory UserModel.fromJson(Map<String, dynamic> json) =>
      _$UserModelFromJson(json);
  Map<String, dynamic> toJson() => _$UserModelToJson(this);
}
