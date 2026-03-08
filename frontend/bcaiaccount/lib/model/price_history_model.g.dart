// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'price_history_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PriceHistoryModel _$PriceHistoryModelFromJson(Map<String, dynamic> json) =>
    PriceHistoryModel(
      guidfixed: json['guidfixed'] as String?,
      shopid: json['shopid'] as String?,
      id: json['id'] as String?,
      productbarcodeguid: json['productbarcodeguid'] as String?,
      barcode: json['barcode'] as String?,
      productname: json['productname'] as String?,
      pricetype: json['pricetype'] as String?,
      keynumber: (json['keynumber'] as num?)?.toInt(),
      oldprice: (json['oldprice'] as num?)?.toDouble(),
      newprice: (json['newprice'] as num?)?.toDouble(),
      pricedifference: (json['pricedifference'] as num?)?.toDouble(),
      action: json['action'] as String?,
      createdby: json['createdby'] as String?,
      createdat: json['createdat'] as String?,
      remark: json['remark'] as String?,
    );

Map<String, dynamic> _$PriceHistoryModelToJson(PriceHistoryModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'shopid': instance.shopid,
      'id': instance.id,
      'productbarcodeguid': instance.productbarcodeguid,
      'barcode': instance.barcode,
      'productname': instance.productname,
      'pricetype': instance.pricetype,
      'keynumber': instance.keynumber,
      'oldprice': instance.oldprice,
      'newprice': instance.newprice,
      'pricedifference': instance.pricedifference,
      'action': instance.action,
      'createdby': instance.createdby,
      'createdat': instance.createdat,
      'remark': instance.remark,
    };

PaginationModel _$PaginationModelFromJson(Map<String, dynamic> json) =>
    PaginationModel(
      total: (json['total'] as num?)?.toInt(),
      page: (json['page'] as num?)?.toInt(),
      perPage: (json['perPage'] as num?)?.toInt(),
      prev: (json['prev'] as num?)?.toInt(),
      next: (json['next'] as num?)?.toInt(),
      totalPage: (json['totalPage'] as num?)?.toInt(),
    );

Map<String, dynamic> _$PaginationModelToJson(PaginationModel instance) =>
    <String, dynamic>{
      'total': instance.total,
      'page': instance.page,
      'perPage': instance.perPage,
      'prev': instance.prev,
      'next': instance.next,
      'totalPage': instance.totalPage,
    };

PriceHistoryResponseModel _$PriceHistoryResponseModelFromJson(
  Map<String, dynamic> json,
) => PriceHistoryResponseModel(
  success: json['success'] as bool?,
  data: (json['data'] as List<dynamic>?)
      ?.map((e) => PriceHistoryModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  pagination: json['pagination'] == null
      ? null
      : PaginationModel.fromJson(json['pagination'] as Map<String, dynamic>),
);

Map<String, dynamic> _$PriceHistoryResponseModelToJson(
  PriceHistoryResponseModel instance,
) => <String, dynamic>{
  'success': instance.success,
  'data': instance.data,
  'pagination': instance.pagination,
};
