// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_department_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductDepartmentModel _$ProductDepartmentModelFromJson(
  Map<String, dynamic> json,
) => ProductDepartmentModel(
  branchcode: json['branchcode'] as String,
  departmentcode: json['departmentcode'] as String,
  productcodes: (json['productcodes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
);

Map<String, dynamic> _$ProductDepartmentModelToJson(
  ProductDepartmentModel instance,
) => <String, dynamic>{
  'branchcode': instance.branchcode,
  'departmentcode': instance.departmentcode,
  'productcodes': instance.productcodes,
};
