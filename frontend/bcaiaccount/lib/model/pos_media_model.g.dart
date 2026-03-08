// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'pos_media_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PosMediaModel _$PosMediaModelFromJson(Map<String, dynamic> json) =>
    PosMediaModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      description: (json['description'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      resources: (json['resources'] as List<dynamic>)
          .map((e) => ResourceModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$PosMediaModelToJson(PosMediaModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'description': instance.description,
      'resources': instance.resources,
    };

ResourceModel _$ResourceModelFromJson(Map<String, dynamic> json) =>
    ResourceModel(
      daysofweek: (json['daysofweek'] as List<dynamic>)
          .map((e) => (e as num).toInt())
          .toList(),
      description: (json['description'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      displaytime: (json['displaytime'] as num).toInt(),
      fromDate: json['fromDate'] as String,
      fromTime: json['fromTime'] as String,
      mediaType: (json['mediaType'] as num).toInt(),
      toDate: json['toDate'] as String,
      toTime: json['toTime'] as String,
      uri: json['uri'] as String,
      urilink: json['urilink'] as String?,
      uriVideo: json['uriVideo'] as String?,
    );

Map<String, dynamic> _$ResourceModelToJson(ResourceModel instance) =>
    <String, dynamic>{
      'daysofweek': instance.daysofweek,
      'description': instance.description,
      'displaytime': instance.displaytime,
      'fromDate': instance.fromDate,
      'fromTime': instance.fromTime,
      'mediaType': instance.mediaType,
      'toDate': instance.toDate,
      'toTime': instance.toTime,
      'uri': instance.uri,
      'uriVideo': instance.uriVideo,
      'urilink': instance.urilink,
    };
