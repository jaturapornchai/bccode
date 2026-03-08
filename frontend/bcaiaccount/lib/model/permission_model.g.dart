// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'permission_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ScreenPermissionModel _$ScreenPermissionModelFromJson(
  Map<String, dynamic> json,
) => ScreenPermissionModel(
  screenCode: json['screenCode'] as String,
  screenName: json['screenName'] as String? ?? '',
  canView: json['canView'] as bool? ?? true,
  canAdd: json['canAdd'] as bool? ?? false,
  canEditOwn: json['canEditOwn'] as bool? ?? false,
  canEditAll: json['canEditAll'] as bool? ?? false,
  canDeleteOwn: json['canDeleteOwn'] as bool? ?? false,
  canDeleteAll: json['canDeleteAll'] as bool? ?? false,
  canPrint: json['canPrint'] as bool? ?? false,
);

Map<String, dynamic> _$ScreenPermissionModelToJson(
  ScreenPermissionModel instance,
) => <String, dynamic>{
  'screenCode': instance.screenCode,
  'screenName': instance.screenName,
  'canView': instance.canView,
  'canAdd': instance.canAdd,
  'canEditOwn': instance.canEditOwn,
  'canEditAll': instance.canEditAll,
  'canDeleteOwn': instance.canDeleteOwn,
  'canDeleteAll': instance.canDeleteAll,
  'canPrint': instance.canPrint,
};

BranchPermissionModel _$BranchPermissionModelFromJson(
  Map<String, dynamic> json,
) => BranchPermissionModel(
  branchCode: json['branchCode'] as String,
  branchName: json['branchName'] as String? ?? '',
  screens: (json['screens'] as List<dynamic>?)
      ?.map((e) => ScreenPermissionModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$BranchPermissionModelToJson(
  BranchPermissionModel instance,
) => <String, dynamic>{
  'branchCode': instance.branchCode,
  'branchName': instance.branchName,
  'screens': instance.screens.map((e) => e.toJson()).toList(),
};

PermissionDefinitionModel _$PermissionDefinitionModelFromJson(
  Map<String, dynamic> json,
) => PermissionDefinitionModel(
  id: json['_id'] as String?,
  shopid: json['shopid'] as String,
  permissionCode: json['permissionCode'] as String,
  permissionName: json['permissionName'] as String? ?? '',
  description: json['description'] as String? ?? '',
  branches: (json['branches'] as List<dynamic>?)
      ?.map((e) => BranchPermissionModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  isActive: json['isActive'] as bool? ?? true,
  createdAt: json['createdAt'] as String?,
  updatedAt: json['updatedAt'] as String?,
  createdBy: json['createdBy'] as String? ?? '',
  updatedBy: json['updatedBy'] as String? ?? '',
);

Map<String, dynamic> _$PermissionDefinitionModelToJson(
  PermissionDefinitionModel instance,
) => <String, dynamic>{
  '_id': ?instance.id,
  'shopid': instance.shopid,
  'permissionCode': instance.permissionCode,
  'permissionName': instance.permissionName,
  'description': instance.description,
  'branches': instance.branches.map((e) => e.toJson()).toList(),
  'isActive': instance.isActive,
  'createdAt': instance.createdAt,
  'updatedAt': instance.updatedAt,
  'createdBy': instance.createdBy,
  'updatedBy': instance.updatedBy,
};

EmployeePermissionModel _$EmployeePermissionModelFromJson(
  Map<String, dynamic> json,
) => EmployeePermissionModel(
  id: json['_id'] as String?,
  shopid: json['shopid'] as String,
  employeeCode: json['employeeCode'] as String,
  employeeName: json['employeeName'] as String? ?? '',
  permissionCodes: (json['permissionCodes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  createdAt: json['createdAt'] as String?,
  updatedAt: json['updatedAt'] as String?,
  createdBy: json['createdBy'] as String? ?? '',
  updatedBy: json['updatedBy'] as String? ?? '',
);

Map<String, dynamic> _$EmployeePermissionModelToJson(
  EmployeePermissionModel instance,
) => <String, dynamic>{
  '_id': ?instance.id,
  'shopid': instance.shopid,
  'employeeCode': instance.employeeCode,
  'employeeName': instance.employeeName,
  'permissionCodes': instance.permissionCodes,
  'createdAt': instance.createdAt,
  'updatedAt': instance.updatedAt,
  'createdBy': instance.createdBy,
  'updatedBy': instance.updatedBy,
};
