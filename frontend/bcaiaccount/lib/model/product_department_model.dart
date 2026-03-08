import 'package:json_annotation/json_annotation.dart';

part 'product_department_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ProductDepartmentModel {
  String branchcode;
  String departmentcode;
  List<String> productcodes = <String>[];

  ProductDepartmentModel({
    required this.branchcode,
    required this.departmentcode,
    List<String>? productcodes,
  }) : productcodes = productcodes ?? <String>[];

  factory ProductDepartmentModel.fromJson(Map<String, dynamic> json) =>
      _$ProductDepartmentModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductDepartmentModelToJson(this);
}
