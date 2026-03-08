// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'order_type_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

OrderTypeModel _$OrderTypeModelFromJson(Map<String, dynamic> json) =>
    OrderTypeModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      prices: (json['prices'] as List<dynamic>?)
          ?.map((e) => Price.fromJson(e as Map<String, dynamic>))
          .toList(),
      remarks: (json['remarks'] as List<dynamic>?)
          ?.map(
            (e) => (e as List<dynamic>)
                .map(
                  (e) => LanguageDataModel.fromJson(e as Map<String, dynamic>),
                )
                .toList(),
          )
          .toList(),
    );

Map<String, dynamic> _$OrderTypeModelToJson(OrderTypeModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names,
      'prices': instance.prices,
      'remarks': instance.remarks,
    };

Price _$PriceFromJson(Map<String, dynamic> json) => Price(
  price: (json['price'] as num).toDouble(),
  type: (json['type'] as num).toInt(),
);

Map<String, dynamic> _$PriceToJson(Price instance) => <String, dynamic>{
  'price': instance.price,
  'type': instance.type,
};
