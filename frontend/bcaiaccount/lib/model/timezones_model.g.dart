// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'timezones_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

TimezonesModel _$TimezonesModelFromJson(Map<String, dynamic> json) =>
    TimezonesModel(
      value: json['value'] as String,
      abbr: json['abbr'] as String,
      offset: json['offset'] as String,
      isDst: json['isdst'] as bool,
      text: json['text'] as String,
      utc: (json['utc'] as List<dynamic>).map((e) => e as String).toList(),
    );

Map<String, dynamic> _$TimezonesModelToJson(TimezonesModel instance) =>
    <String, dynamic>{
      'value': instance.value,
      'abbr': instance.abbr,
      'offset': instance.offset,
      'isdst': instance.isDst,
      'text': instance.text,
      'utc': instance.utc,
    };
