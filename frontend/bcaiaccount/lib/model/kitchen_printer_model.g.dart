// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'kitchen_printer_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

KitchenPrinterModel _$KitchenPrinterModelFromJson(Map<String, dynamic> json) =>
    KitchenPrinterModel(
      guidfixed: json['guidfixed'] as String,
      guidkitchen: json['guidkitchen'] as String,
      kitchencode: json['kitchencode'] as String,
      kitchenname: (json['kitchenname'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      primary: json['primary'] as String,
      spare: json['spare'] as String,
    );

Map<String, dynamic> _$KitchenPrinterModelToJson(
  KitchenPrinterModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'guidkitchen': instance.guidkitchen,
  'kitchencode': instance.kitchencode,
  'kitchenname': instance.kitchenname,
  'primary': instance.primary,
  'spare': instance.spare,
};

KitchenPrinterSaveModel _$KitchenPrinterSaveModelFromJson(
  Map<String, dynamic> json,
) => KitchenPrinterSaveModel(
  guidfixed: json['guidfixed'] as String,
  guidkitchen: json['guidkitchen'] as String,
  kitchencode: json['kitchencode'] as String,
  kitchenname: (json['kitchenname'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  printers: (json['printers'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
);

Map<String, dynamic> _$KitchenPrinterSaveModelToJson(
  KitchenPrinterSaveModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'guidkitchen': instance.guidkitchen,
  'kitchencode': instance.kitchencode,
  'kitchenname': instance.kitchenname,
  'printers': instance.printers,
};
