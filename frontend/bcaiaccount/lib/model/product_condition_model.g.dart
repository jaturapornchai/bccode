// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_condition_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductConditionModel _$ProductConditionModelFromJson(
  Map<String, dynamic> json,
) => ProductConditionModel(
  productCodes: (json['product_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  groupCodes: (json['group_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  groupSuboneCodes: (json['group_subone_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  groupSubtwoCodes: (json['group_subtwo_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  brandCodes: (json['brand_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  designCodes: (json['design_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  modelCodes: (json['model_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  patternCodes: (json['pattern_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  gradeCodes: (json['grade_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  categoryCodes: (json['category_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  classCodes: (json['class_codes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  minimumAmount: (json['minimum_amount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ProductConditionModelToJson(
  ProductConditionModel instance,
) => <String, dynamic>{
  'product_codes': instance.productCodes,
  'group_codes': instance.groupCodes,
  'group_subone_codes': instance.groupSuboneCodes,
  'group_subtwo_codes': instance.groupSubtwoCodes,
  'brand_codes': instance.brandCodes,
  'design_codes': instance.designCodes,
  'model_codes': instance.modelCodes,
  'pattern_codes': instance.patternCodes,
  'grade_codes': instance.gradeCodes,
  'category_codes': instance.categoryCodes,
  'class_codes': instance.classCodes,
  'minimum_amount': instance.minimumAmount,
};
