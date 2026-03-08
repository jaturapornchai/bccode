// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_branch_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductBranchModel _$ProductBranchModelFromJson(Map<String, dynamic> json) =>
    ProductBranchModel(
      branchcode: json['branchcode'] as String,
      productcodes: (json['productcodes'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$ProductBranchModelToJson(ProductBranchModel instance) =>
    <String, dynamic>{
      'branchcode': instance.branchcode,
      'productcodes': instance.productcodes,
    };
