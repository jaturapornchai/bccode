// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sale_channel_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SaleChannelModel _$SaleChannelModelFromJson(Map<String, dynamic> json) =>
    SaleChannelModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String?,
      gp: (json['gp'] as num?)?.toDouble(),
      gptype: (json['gptype'] as num?)?.toInt(),
      imageuri: json['imageuri'] as String?,
      name: json['name'] as String?,
      price: (json['price'] as num?)?.toInt(),
    );

Map<String, dynamic> _$SaleChannelModelToJson(SaleChannelModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'gp': instance.gp,
      'gptype': instance.gptype,
      'imageuri': instance.imageuri,
      'name': instance.name,
      'price': instance.price,
    };
