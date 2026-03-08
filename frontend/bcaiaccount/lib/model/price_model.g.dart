// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'price_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PriceDataModel _$PriceDataModelFromJson(Map<String, dynamic> json) =>
    PriceDataModel(
      keynumber: (json['keynumber'] as num).toInt(),
      price: (json['price'] as num).toDouble(),
    );

Map<String, dynamic> _$PriceDataModelToJson(PriceDataModel instance) =>
    <String, dynamic>{'keynumber': instance.keynumber, 'price': instance.price};
