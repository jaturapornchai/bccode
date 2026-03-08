// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'coupon_import_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CouponImportUploadResponse _$CouponImportUploadResponseFromJson(
  Map<String, dynamic> json,
) => CouponImportUploadResponse(
  success: json['success'] as bool,
  data: CouponImportUploadData.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$CouponImportUploadResponseToJson(
  CouponImportUploadResponse instance,
) => <String, dynamic>{'success': instance.success, 'data': instance.data};

CouponImportUploadData _$CouponImportUploadDataFromJson(
  Map<String, dynamic> json,
) => CouponImportUploadData(
  success: json['success'] as bool,
  batchId: json['batch_id'] as String,
  importedCount: (json['imported_count'] as num).toInt(),
  errorCount: (json['error_count'] as num).toInt(),
  totalRows: (json['total_rows'] as num).toInt(),
  message: json['message'] as String,
  errors: (json['errors'] as List<dynamic>?)
      ?.map((e) => CouponImportError.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$CouponImportUploadDataToJson(
  CouponImportUploadData instance,
) => <String, dynamic>{
  'success': instance.success,
  'batch_id': instance.batchId,
  'imported_count': instance.importedCount,
  'error_count': instance.errorCount,
  'total_rows': instance.totalRows,
  'message': instance.message,
  'errors': instance.errors,
};

CouponImportError _$CouponImportErrorFromJson(Map<String, dynamic> json) =>
    CouponImportError(
      row: (json['row'] as num?)?.toInt(),
      field: json['field'] as String?,
      message: json['message'] as String?,
      error: json['error'] as String?,
      couponCode: json['coupon_code'] as String?,
    );

Map<String, dynamic> _$CouponImportErrorToJson(CouponImportError instance) =>
    <String, dynamic>{
      'row': instance.row,
      'field': instance.field,
      'message': instance.message,
      'error': instance.error,
      'coupon_code': instance.couponCode,
    };

CouponImportStatusResponse _$CouponImportStatusResponseFromJson(
  Map<String, dynamic> json,
) => CouponImportStatusResponse(
  success: json['success'] as bool,
  data: CouponImportStatusData.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$CouponImportStatusResponseToJson(
  CouponImportStatusResponse instance,
) => <String, dynamic>{'success': instance.success, 'data': instance.data};

CouponImportStatusData _$CouponImportStatusDataFromJson(
  Map<String, dynamic> json,
) => CouponImportStatusData(
  batchId: json['batch_id'] as String,
  status: json['status'] as String,
  progress: (json['progress'] as num).toInt(),
  totalRows: (json['total_rows'] as num).toInt(),
  processedRows: (json['processed_rows'] as num).toInt(),
  successCount: (json['success_count'] as num).toInt(),
  errorCount: (json['error_count'] as num).toInt(),
  message: json['message'] as String,
  updatedAt: json['updated_at'] as String,
  errors: (json['errors'] as List<dynamic>?)
      ?.map((e) => CouponImportError.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$CouponImportStatusDataToJson(
  CouponImportStatusData instance,
) => <String, dynamic>{
  'batch_id': instance.batchId,
  'status': instance.status,
  'progress': instance.progress,
  'total_rows': instance.totalRows,
  'processed_rows': instance.processedRows,
  'success_count': instance.successCount,
  'error_count': instance.errorCount,
  'message': instance.message,
  'updated_at': instance.updatedAt,
  'errors': instance.errors,
};
