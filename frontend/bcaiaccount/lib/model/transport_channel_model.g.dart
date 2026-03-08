// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'transport_channel_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

TransportChannelModel _$TransportChannelModelFromJson(
  Map<String, dynamic> json,
) => TransportChannelModel(
  guidfixed: json['guidfixed'] as String?,
  code: json['code'] as String?,
  name: json['name'] as String?,
  imageuri: json['imageuri'] as String?,
);

Map<String, dynamic> _$TransportChannelModelToJson(
  TransportChannelModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'code': instance.code,
  'name': instance.name,
  'imageuri': instance.imageuri,
};
