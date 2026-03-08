// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'dimension_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DimensionModel _$DimensionModelFromJson(Map<String, dynamic> json) =>
    DimensionModel(
      guidfixed: json['guidfixed'] as String?,
      isdisabled: json['isdisabled'] as bool?,
      items: (json['items'] as List<dynamic>?)
          ?.map((e) => ItemDimension.fromJson(e as Map<String, dynamic>))
          .toList(),
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$DimensionModelToJson(DimensionModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'isdisabled': instance.isdisabled,
      'items': instance.items,
      'names': instance.names,
    };

DimensionProductModel _$DimensionProductModelFromJson(
  Map<String, dynamic> json,
) => DimensionProductModel(
  guidfixed: json['guidfixed'] as String?,
  isdisabled: json['isdisabled'] as bool?,
  item: json['item'] == null
      ? null
      : ItemDimension.fromJson(json['item'] as Map<String, dynamic>),
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$DimensionProductModelToJson(
  DimensionProductModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'isdisabled': instance.isdisabled,
  'item': instance.item,
  'names': instance.names,
};

ItemDimension _$ItemDimensionFromJson(Map<String, dynamic> json) =>
    ItemDimension(
      guidfixed: json['guidfixed'] as String?,
      isdisabled: json['isdisabled'] as bool?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ItemDimensionToJson(ItemDimension instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'isdisabled': instance.isdisabled,
      'names': instance.names,
    };
