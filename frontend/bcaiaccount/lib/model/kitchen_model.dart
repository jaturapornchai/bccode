// ignore_for_file: non_constant_identifier_names

import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/kitchen_product_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'kitchen_model.g.dart';

@JsonSerializable()
class KitchenModel {
  late String? guidfixed;
  late String code;
  late List<LanguageDataModel> names;
  late List<String>? printers;
  late List<String>? products;
  late List<String>? zones;
  int? groupnumber;
  int? zonenumber;

  KitchenModel({
    String? guidfixed,
    required this.code,
    required this.names,
    List<String>? printers,
    List<String>? products,
    List<String>? zones,
    int? groupnumber,
    int? zonenumber,
  })  : printers = printers ?? [],
        products = products ?? [],
        zones = zones ?? [],
        guidfixed = guidfixed ?? "",
        groupnumber = groupnumber ?? 1,
        zonenumber = zonenumber ?? 1;

  factory KitchenModel.fromJson(Map<String, dynamic> json) =>
      _$KitchenModelFromJson(json);
  Map<String, dynamic> toJson() => _$KitchenModelToJson(this);
}

@JsonSerializable()
class GetKitchenModel {
  late String guidfixed;
  late String code;
  late List<LanguageDataModel> names;
  late List<String>? printers;
  late List<KitchenProductModel>? products;
  late List<String>? zones;

  GetKitchenModel({
    required this.guidfixed,
    required this.code,
    required this.names,
    List<String>? printers,
    List<KitchenProductModel>? products,
    List<String>? zones,
  })  : printers = printers ?? [],
        products = products ?? <KitchenProductModel>[],
        zones = zones ?? [];

  factory GetKitchenModel.fromJson(Map<String, dynamic> json) =>
      _$GetKitchenModelFromJson(json);
  Map<String, dynamic> toJson() => _$GetKitchenModelToJson(this);
}
