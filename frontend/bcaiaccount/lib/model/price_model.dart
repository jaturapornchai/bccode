import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'price_model.g.dart';

class PriceModel {
  int keyNumber;
  List<LanguageModel> names;
  bool isUse;

  PriceModel(
      {required this.keyNumber, required this.isUse, required this.names});
}

@JsonSerializable(explicitToJson: true)
class PriceDataModel {
  int keynumber;
  double price;

  PriceDataModel({required this.keynumber, required this.price});

  factory PriceDataModel.fromJson(Map<String, dynamic> json) =>
      _$PriceDataModelFromJson(json);

  Map<String, dynamic> toJson() => _$PriceDataModelToJson(this);
}
