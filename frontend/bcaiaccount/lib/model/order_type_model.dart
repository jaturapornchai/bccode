import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'order_type_model.g.dart';

@JsonSerializable()
class OrderTypeModel {
  String? guidfixed;
  String code;
  List<LanguageDataModel> names;
  List<Price>? prices;
  List<List<LanguageDataModel>>? remarks;

  OrderTypeModel({
    String? guidfixed,
    required this.code,
    required this.names,
    List<Price>? prices,
    List<List<LanguageDataModel>>? remarks,
  })  : guidfixed = guidfixed ?? "",
        prices = prices ?? [],
        remarks = remarks ?? [];

  factory OrderTypeModel.fromJson(Map<String, dynamic> json) =>
      _$OrderTypeModelFromJson(json);
  Map<String, dynamic> toJson() => _$OrderTypeModelToJson(this);
}

@JsonSerializable()
class Price {
  double price;
  int type;

  Price({
    required this.price,
    required this.type,
  });

  factory Price.fromJson(Map<String, dynamic> json) => _$PriceFromJson(json);
  Map<String, dynamic> toJson() => _$PriceToJson(this);
}
