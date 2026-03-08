// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DocumentApprovalModel _$DocumentApprovalModelFromJson(
  Map<String, dynamic> json,
) => DocumentApprovalModel(
  approvalRole: (json['approval_role'] as num?)?.toInt() ?? 0,
  maxApprovalAmount: (json['max_approval_amount'] as num?)?.toDouble() ?? 0,
);

Map<String, dynamic> _$DocumentApprovalModelToJson(
  DocumentApprovalModel instance,
) => <String, dynamic>{
  'approval_role': instance.approvalRole,
  'max_approval_amount': instance.maxApprovalAmount,
};

UserModel _$UserModelFromJson(Map<String, dynamic> json) => UserModel(
  role: (json['role'] as num).toInt(),
  shopid: json['shopid'] as String,
  username: json['username'] as String,
  editusername: json['editusername'] as String?,
  email: json['email'] as String?,
  lineUserId: json['line_user_id'] as String?,
  lineDisplayName: json['line_display_name'] as String?,
  linePictureUrl: json['line_picture_url'] as String?,
  position: json['position'] as String?,
  department: json['department'] as String?,
  approvalRole: (json['approval_role'] as num?)?.toInt(),
  maxApprovalAmount: (json['max_approval_amount'] as num?)?.toDouble(),
  poApproval: json['po_approval'] == null
      ? null
      : DocumentApprovalModel.fromJson(
          json['po_approval'] as Map<String, dynamic>,
        ),
  quotationApproval: json['quotation_approval'] == null
      ? null
      : DocumentApprovalModel.fromJson(
          json['quotation_approval'] as Map<String, dynamic>,
        ),
);

Map<String, dynamic> _$UserModelToJson(UserModel instance) => <String, dynamic>{
  'role': instance.role,
  'shopid': instance.shopid,
  'username': instance.username,
  'editusername': instance.editusername,
  'email': instance.email,
  'line_user_id': instance.lineUserId,
  'line_display_name': instance.lineDisplayName,
  'line_picture_url': instance.linePictureUrl,
  'position': instance.position,
  'department': instance.department,
  'approval_role': instance.approvalRole,
  'max_approval_amount': instance.maxApprovalAmount,
  'po_approval': instance.poApproval?.toJson(),
  'quotation_approval': instance.quotationApproval?.toJson(),
};
