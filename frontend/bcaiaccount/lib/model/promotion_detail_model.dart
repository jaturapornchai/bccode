import 'package:smlaicloud/model/product_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'promotion_detail_model.g.dart';

@JsonSerializable()
class PromotionDetailModel {
  int detailtype;
  double minimum;
  double discount;
  ProductBarcodeModel productbarcode;

  PromotionDetailModel({
    required this.detailtype,
    required this.minimum,
    required this.discount,
    required this.productbarcode,
  });

  factory PromotionDetailModel.fromJson(Map<String, dynamic> json) =>
      _$PromotionDetailModelFromJson(json);
  Map<String, dynamic> toJson() => _$PromotionDetailModelToJson(this);
}
