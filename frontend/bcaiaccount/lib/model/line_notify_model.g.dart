// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'line_notify_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

LineNotifyModel _$LineNotifyModelFromJson(Map<String, dynamic> json) =>
    LineNotifyModel(
      guidfixed: json['guidfixed'] as String?,
      branchevents: (json['branchevents'] as List<dynamic>?)
          ?.map((e) => BranchEvent.fromJson(e as Map<String, dynamic>))
          .toList(),
      name: json['name'] as String?,
      token: json['token'] as String?,
      type: json['type'] as String?,
    );

Map<String, dynamic> _$LineNotifyModelToJson(LineNotifyModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'branchevents': instance.branchevents,
      'name': instance.name,
      'token': instance.token,
      'type': instance.type,
    };

BranchEvent _$BranchEventFromJson(Map<String, dynamic> json) => BranchEvent(
  branch: json['branch'] == null
      ? null
      : Branch.fromJson(json['branch'] as Map<String, dynamic>),
  isenable: json['isenable'] as bool?,
  isnearoutofstock: json['isnearoutofstock'] as bool?,
  isoutofstock: json['isoutofstock'] as bool?,
  issavebill: json['issavebill'] as bool?,
  ispreorder: json['ispreorder'] as bool?,
);

Map<String, dynamic> _$BranchEventToJson(BranchEvent instance) =>
    <String, dynamic>{
      'branch': instance.branch,
      'isenable': instance.isenable,
      'isnearoutofstock': instance.isnearoutofstock,
      'isoutofstock': instance.isoutofstock,
      'issavebill': instance.issavebill,
      'ispreorder': instance.ispreorder,
    };

Branch _$BranchFromJson(Map<String, dynamic> json) => Branch(
  code: json['code'] as String?,
  guidfixed: json['guidfixed'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$BranchToJson(Branch instance) => <String, dynamic>{
  'code': instance.code,
  'guidfixed': instance.guidfixed,
  'names': instance.names,
};
