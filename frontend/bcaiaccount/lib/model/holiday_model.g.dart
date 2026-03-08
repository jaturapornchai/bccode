// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'holiday_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

HolidayModel _$HolidayModelFromJson(Map<String, dynamic> json) => HolidayModel(
  guidfixed: json['guidfixed'] as String,
  date: json['date'] as String,
  desc: (json['desc'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$HolidayModelToJson(HolidayModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'date': instance.date,
      'desc': instance.desc,
    };
