// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'choice_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ChoiceModel _$ChoiceModelFromJson(Map<String, dynamic> json) => ChoiceModel(
  barcode: json['barcode'] as String?,
  isdefault: json['default'] as bool,
  itemunit: json['itemunit'] as String?,
  name1: json['name1'] as String?,
  name2: json['name2'] as String?,
  name3: json['name3'] as String?,
  name4: json['name4'] as String?,
  name5: json['name5'] as String?,
  price: (json['price'] as num?)?.toDouble(),
  qty: (json['qty'] as num?)?.toDouble(),
  qtymax: (json['qtymax'] as num?)?.toDouble(),
  selected: json['selected'] as bool,
  suggestcode: json['suggestcode'] as String?,
);

Map<String, dynamic> _$ChoiceModelToJson(ChoiceModel instance) =>
    <String, dynamic>{
      'barcode': instance.barcode,
      'default': instance.isdefault,
      'itemunit': instance.itemunit,
      'name1': instance.name1,
      'name2': instance.name2,
      'name3': instance.name3,
      'name4': instance.name4,
      'name5': instance.name5,
      'price': instance.price,
      'qty': instance.qty,
      'qtymax': instance.qtymax,
      'selected': instance.selected,
      'suggestcode': instance.suggestcode,
    };
