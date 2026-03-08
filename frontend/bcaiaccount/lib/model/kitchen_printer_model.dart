// ignore_for_file: non_constant_identifier_names

import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'kitchen_printer_model.g.dart';

@JsonSerializable()
class KitchenPrinterModel {
  late String guidfixed;
  late String guidkitchen;
  late String kitchencode;
  late List<LanguageDataModel> kitchenname;
  late String primary;
  late String spare;

  KitchenPrinterModel({
    required this.guidfixed,
    required this.guidkitchen,
    required this.kitchencode,
    required this.kitchenname,
    required this.primary,
    required this.spare,
  });

  factory KitchenPrinterModel.fromJson(Map<String, dynamic> json) =>
      _$KitchenPrinterModelFromJson(json);
  Map<String, dynamic> toJson() => _$KitchenPrinterModelToJson(this);
}

@JsonSerializable()
class KitchenPrinterSaveModel {
  late String guidfixed;
  late String guidkitchen;
  late String kitchencode;
  late List<LanguageDataModel> kitchenname;
  late List<String> printers;

  KitchenPrinterSaveModel({
    required this.guidfixed,
    required this.guidkitchen,
    required this.kitchencode,
    required this.kitchenname,
    required this.printers,
  });

  factory KitchenPrinterSaveModel.fromJson(Map<String, dynamic> json) =>
      _$KitchenPrinterSaveModelFromJson(json);
  Map<String, dynamic> toJson() => _$KitchenPrinterSaveModelToJson(this);
}
