// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'option_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

OptionModel _$OptionModelFromJson(Map<String, dynamic> json) => OptionModel(
  guidfixed: json['guidfixed'] as String?,
  choicetype: (json['choicetype'] as num).toInt(),
  code: json['code'] as String,
  maxselect: (json['maxselect'] as num).toInt(),
  name1: json['name1'] as String,
  name2: json['name2'] as String?,
  name3: json['name3'] as String?,
  name4: json['name4'] as String?,
  name5: json['name5'] as String?,
  isrequired: json['required'] as bool,
  choices: (json['choices'] as List<dynamic>?)
      ?.map((e) => ChoiceModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$OptionModelToJson(OptionModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'choicetype': instance.choicetype,
      'code': instance.code,
      'maxselect': instance.maxselect,
      'name1': instance.name1,
      'name2': instance.name2,
      'name3': instance.name3,
      'name4': instance.name4,
      'name5': instance.name5,
      'required': instance.isrequired,
      'choices': instance.choices,
    };
