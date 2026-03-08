import 'package:json_annotation/json_annotation.dart';

part 'product_branch_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ProductBranchModel {
  String branchcode;
  List<String> productcodes = <String>[];

  ProductBranchModel({
    required this.branchcode,
    List<String>? productcodes,
  }) : productcodes = productcodes ?? <String>[];

  factory ProductBranchModel.fromJson(Map<String, dynamic> json) =>
      _$ProductBranchModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductBranchModelToJson(this);
}
