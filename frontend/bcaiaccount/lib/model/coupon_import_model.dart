import 'package:json_annotation/json_annotation.dart';

part 'coupon_import_model.g.dart';

/// Response model for /coupon/upload-excel
@JsonSerializable()
class CouponImportUploadResponse {
  final bool success;
  final CouponImportUploadData data;

  CouponImportUploadResponse({required this.success, required this.data});

  factory CouponImportUploadResponse.fromJson(Map<String, dynamic> json) =>
      _$CouponImportUploadResponseFromJson(json);

  Map<String, dynamic> toJson() => _$CouponImportUploadResponseToJson(this);
}

@JsonSerializable()
class CouponImportUploadData {
  final bool success;
  @JsonKey(name: 'batch_id')
  final String batchId;
  @JsonKey(name: 'imported_count')
  final int importedCount;
  @JsonKey(name: 'error_count')
  final int errorCount;
  @JsonKey(name: 'total_rows')
  final int totalRows;
  final String message;
  final List<CouponImportError>? errors;

  CouponImportUploadData({
    required this.success,
    required this.batchId,
    required this.importedCount,
    required this.errorCount,
    required this.totalRows,
    required this.message,
    this.errors,
  });

  factory CouponImportUploadData.fromJson(Map<String, dynamic> json) =>
      _$CouponImportUploadDataFromJson(json);

  Map<String, dynamic> toJson() => _$CouponImportUploadDataToJson(this);
}

@JsonSerializable()
class CouponImportError {
  final int? row;
  final String? field;
  final String? message;
  final String? error;
  @JsonKey(name: 'coupon_code')
  final String? couponCode;

  CouponImportError({
    this.row,
    this.field,
    this.message,
    this.error,
    this.couponCode,
  });

  factory CouponImportError.fromJson(Map<String, dynamic> json) =>
      _$CouponImportErrorFromJson(json);

  Map<String, dynamic> toJson() => _$CouponImportErrorToJson(this);

  // Helper to get error message (support both 'message' and 'error' fields)
  String get errorMessage => error ?? message ?? 'Unknown error';
}

/// Response model for /coupon/import-status/{batch_id}
@JsonSerializable()
class CouponImportStatusResponse {
  final bool success;
  final CouponImportStatusData data;

  CouponImportStatusResponse({required this.success, required this.data});

  factory CouponImportStatusResponse.fromJson(Map<String, dynamic> json) =>
      _$CouponImportStatusResponseFromJson(json);

  Map<String, dynamic> toJson() => _$CouponImportStatusResponseToJson(this);
}

@JsonSerializable()
class CouponImportStatusData {
  @JsonKey(name: 'batch_id')
  final String batchId;
  final String status; // "processing", "completed", "failed"
  final int progress; // 0-100
  @JsonKey(name: 'total_rows')
  final int totalRows;
  @JsonKey(name: 'processed_rows')
  final int processedRows;
  @JsonKey(name: 'success_count')
  final int successCount;
  @JsonKey(name: 'error_count')
  final int errorCount;
  final String message;
  @JsonKey(name: 'updated_at')
  final String updatedAt;
  final List<CouponImportError>? errors;

  CouponImportStatusData({
    required this.batchId,
    required this.status,
    required this.progress,
    required this.totalRows,
    required this.processedRows,
    required this.successCount,
    required this.errorCount,
    required this.message,
    required this.updatedAt,
    this.errors,
  });

  factory CouponImportStatusData.fromJson(Map<String, dynamic> json) =>
      _$CouponImportStatusDataFromJson(json);

  Map<String, dynamic> toJson() => _$CouponImportStatusDataToJson(this);

  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isProcessing => status == 'processing';
}
