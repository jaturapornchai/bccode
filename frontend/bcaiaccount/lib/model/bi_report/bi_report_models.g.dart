// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'bi_report_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReportConditionsModel _$ReportConditionsModelFromJson(
  Map<String, dynamic> json,
) => ReportConditionsModel(
  fromdate: json['fromdate'] as String? ?? '',
  todate: json['todate'] as String? ?? '',
  branchcode: json['branchcode'] as String? ?? '',
  showdetail: json['showdetail'] as bool? ?? true,
  iscancel: json['iscancel'] as String? ?? '',
  inquirytype: json['inquirytype'] as String? ?? '',
  ispos: json['ispos'] as bool? ?? false,
  creditorcode: json['creditorcode'] as String? ?? '',
  salecode: json['salecode'] as String? ?? '',
  debtorcode: json['debtorcode'] as String? ?? '',
  barcode: json['barcode'] as String? ?? '',
);

Map<String, dynamic> _$ReportConditionsModelToJson(
  ReportConditionsModel instance,
) => <String, dynamic>{
  'fromdate': instance.fromdate,
  'todate': instance.todate,
  'branchcode': instance.branchcode,
  'showdetail': instance.showdetail,
  'iscancel': instance.iscancel,
  'inquirytype': instance.inquirytype,
  'ispos': instance.ispos,
  'creditorcode': instance.creditorcode,
  'salecode': instance.salecode,
  'debtorcode': instance.debtorcode,
  'barcode': instance.barcode,
};

BiReportJobResponse _$BiReportJobResponseFromJson(Map<String, dynamic> json) =>
    BiReportJobResponse(
      status: json['status'] as String,
      jobId: json['job_id'] as String,
      message: json['message'] as String,
    );

Map<String, dynamic> _$BiReportJobResponseToJson(
  BiReportJobResponse instance,
) => <String, dynamic>{
  'status': instance.status,
  'job_id': instance.jobId,
  'message': instance.message,
};

BiReportStatusResponse _$BiReportStatusResponseFromJson(
  Map<String, dynamic> json,
) => BiReportStatusResponse(
  status: json['status'] as String,
  data: BiReportStatusData.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$BiReportStatusResponseToJson(
  BiReportStatusResponse instance,
) => <String, dynamic>{'status': instance.status, 'data': instance.data};

BiReportStatusData _$BiReportStatusDataFromJson(Map<String, dynamic> json) =>
    BiReportStatusData(
      success: json['success'] as bool,
      jobId: json['job_id'] as String,
      state: json['state'] as String,
      progress: (json['progress'] as num).toInt(),
      createdAt: json['createdAt'] as String,
      processedOn: json['processedOn'] as String?,
      finishedOn: json['finishedOn'] as String?,
      failedReason: json['failedReason'] as String?,
    );

Map<String, dynamic> _$BiReportStatusDataToJson(BiReportStatusData instance) =>
    <String, dynamic>{
      'success': instance.success,
      'job_id': instance.jobId,
      'state': instance.state,
      'progress': instance.progress,
      'createdAt': instance.createdAt,
      'processedOn': instance.processedOn,
      'finishedOn': instance.finishedOn,
      'failedReason': instance.failedReason,
    };

BiReportDetailResponse<T> _$BiReportDetailResponseFromJson<T>(
  Map<String, dynamic> json,
  T Function(Object? json) fromJsonT,
) => BiReportDetailResponse<T>(
  status: json['status'] as String,
  data: (json['data'] as List<dynamic>).map(fromJsonT).toList(),
  meta: BiReportMeta.fromJson(json['meta'] as Map<String, dynamic>),
);

Map<String, dynamic> _$BiReportDetailResponseToJson<T>(
  BiReportDetailResponse<T> instance,
  Object? Function(T value) toJsonT,
) => <String, dynamic>{
  'status': instance.status,
  'data': instance.data.map(toJsonT).toList(),
  'meta': instance.meta,
};

BiReportMeta _$BiReportMetaFromJson(Map<String, dynamic> json) => BiReportMeta(
  page: (json['page'] as num).toInt(),
  size: (json['size'] as num).toInt(),
  total: (json['total'] as num).toInt(),
  totalPage: (json['total_page'] as num).toInt(),
);

Map<String, dynamic> _$BiReportMetaToJson(BiReportMeta instance) =>
    <String, dynamic>{
      'page': instance.page,
      'size': instance.size,
      'total': instance.total,
      'total_page': instance.totalPage,
    };

BiReportErrorResponse _$BiReportErrorResponseFromJson(
  Map<String, dynamic> json,
) => BiReportErrorResponse(
  code: (json['code'] as num).toInt(),
  message: json['message'] as String,
);

Map<String, dynamic> _$BiReportErrorResponseToJson(
  BiReportErrorResponse instance,
) => <String, dynamic>{'code': instance.code, 'message': instance.message};
