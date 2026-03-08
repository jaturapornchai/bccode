// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'bulk_points_request_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

BulkPointsRequestModel _$BulkPointsRequestModelFromJson(
  Map<String, dynamic> json,
) => BulkPointsRequestModel(
  pointscode: json['pointscode'] as String,
  pointamount: (json['pointamount'] as num).toInt(),
  description: json['description'] as String,
);

Map<String, dynamic> _$BulkPointsRequestModelToJson(
  BulkPointsRequestModel instance,
) => <String, dynamic>{
  'pointscode': instance.pointscode,
  'pointamount': instance.pointamount,
  'description': instance.description,
};

BulkPointsResponseModel _$BulkPointsResponseModelFromJson(
  Map<String, dynamic> json,
) => BulkPointsResponseModel(
  success: json['success'] as bool,
  message: json['message'] as String,
  data: json['data'] == null
      ? null
      : BulkPointsDataModel.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$BulkPointsResponseModelToJson(
  BulkPointsResponseModel instance,
) => <String, dynamic>{
  'success': instance.success,
  'message': instance.message,
  'data': instance.data,
};

BulkPointsDataModel _$BulkPointsDataModelFromJson(Map<String, dynamic> json) =>
    BulkPointsDataModel(
      success: (json['success'] as num).toInt(),
      failed: (json['failed'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      successItems: (json['success_items'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      failedItems: (json['failed_items'] as List<dynamic>?)
          ?.map((e) => e as Map<String, dynamic>)
          .toList(),
    );

Map<String, dynamic> _$BulkPointsDataModelToJson(
  BulkPointsDataModel instance,
) => <String, dynamic>{
  'success': instance.success,
  'failed': instance.failed,
  'total': instance.total,
  'success_items': instance.successItems,
  'failed_items': instance.failedItems,
};
