// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'report_condition_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ConditionLocationSelectModel _$ConditionLocationSelectModelFromJson(
  Map<String, dynamic> json,
) => ConditionLocationSelectModel(
  json['code'] as String,
  json['title'] as String,
)..isSelected = json['isSelected'] as bool;

Map<String, dynamic> _$ConditionLocationSelectModelToJson(
  ConditionLocationSelectModel instance,
) => <String, dynamic>{
  'code': instance.code,
  'title': instance.title,
  'isSelected': instance.isSelected,
};

ConditionWareHouseSelectModel _$ConditionWareHouseSelectModelFromJson(
  Map<String, dynamic> json,
) =>
    ConditionWareHouseSelectModel(
        json['code'] as String,
        json['title'] as String,
      )
      ..isSelected = json['isSelected'] as bool
      ..locations = (json['locations'] as List<dynamic>)
          .map(
            (e) => ConditionLocationSelectModel.fromJson(
              e as Map<String, dynamic>,
            ),
          )
          .toList();

Map<String, dynamic> _$ConditionWareHouseSelectModelToJson(
  ConditionWareHouseSelectModel instance,
) => <String, dynamic>{
  'code': instance.code,
  'title': instance.title,
  'isSelected': instance.isSelected,
  'locations': instance.locations.map((e) => e.toJson()).toList(),
};
