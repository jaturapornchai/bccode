// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'work_day_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WorkDayModel _$WorkDayModelFromJson(Map<String, dynamic> json) => WorkDayModel(
  code: json['code'] as String,
  name: json['name'] as String,
  isactive: json['isactive'] as bool,
  fullday: json['fullday'] as bool,
  worktimes: (json['worktimes'] as List<dynamic>)
      .map((e) => WorkTimeModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$WorkDayModelToJson(WorkDayModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isactive': instance.isactive,
      'fullday': instance.fullday,
      'worktimes': instance.worktimes,
    };

WorkTimeModel _$WorkTimeModelFromJson(Map<String, dynamic> json) =>
    WorkTimeModel(
      starttime: json['starttime'] as String,
      endtime: json['endtime'] as String,
      start: json['start'] as String? ?? "",
      end: json['end'] as String? ?? "",
    );

Map<String, dynamic> _$WorkTimeModelToJson(WorkTimeModel instance) =>
    <String, dynamic>{
      'starttime': instance.starttime,
      'endtime': instance.endtime,
      'start': instance.start,
      'end': instance.end,
    };

WorkDayListModel _$WorkDayListModelFromJson(Map<String, dynamic> json) =>
    WorkDayListModel(
      workdays: (json['workdays'] as List<dynamic>)
          .map((e) => WorkDayModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$WorkDayListModelToJson(WorkDayListModel instance) =>
    <String, dynamic>{'workdays': instance.workdays};
