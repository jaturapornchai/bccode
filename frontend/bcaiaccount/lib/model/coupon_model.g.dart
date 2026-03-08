// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'coupon_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CouponModel _$CouponModelFromJson(Map<String, dynamic> json) => CouponModel(
  guidfixed: json['guidfixed'] as String?,
  couponcode: json['couponcode'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  couponvalue: (json['couponvalue'] as num?)?.toDouble(),
  issueddate: json['issueddate'] as String?,
  expirydate: json['expirydate'] as String?,
  coupontype: (json['coupontype'] as num?)?.toInt(),
  customercodes: (json['customercodes'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  remark: json['remark'] as String?,
  status: (json['status'] as num?)?.toInt(),
  isonetimeuse: json['isonetimeuse'] as bool?,
  maxusagecount: (json['maxusagecount'] as num?)?.toInt(),
  maxusagecountpercustomer: (json['maxusagecountpercustomer'] as num?)?.toInt(),
  productCondition: json['product_condition'] == null
      ? null
      : ProductConditionModel.fromJson(
          json['product_condition'] as Map<String, dynamic>,
        ),
);

Map<String, dynamic> _$CouponModelToJson(CouponModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'couponcode': instance.couponcode,
      'names': instance.names?.map((e) => e.toJson()).toList(),
      'couponvalue': instance.couponvalue,
      'issueddate': instance.issueddate,
      'expirydate': instance.expirydate,
      'coupontype': instance.coupontype,
      'customercodes': instance.customercodes,
      'remark': instance.remark,
      'status': instance.status,
      'isonetimeuse': instance.isonetimeuse,
      'maxusagecount': instance.maxusagecount,
      'maxusagecountpercustomer': instance.maxusagecountpercustomer,
      'product_condition': instance.productCondition?.toJson(),
    };

CouponResponseModel _$CouponResponseModelFromJson(Map<String, dynamic> json) =>
    CouponResponseModel(
      success: json['success'] as bool?,
      data: (json['data'] as List<dynamic>?)
          ?.map((e) => CouponModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$CouponResponseModelToJson(
  CouponResponseModel instance,
) => <String, dynamic>{
  'success': instance.success,
  'data': instance.data?.map((e) => e.toJson()).toList(),
};
