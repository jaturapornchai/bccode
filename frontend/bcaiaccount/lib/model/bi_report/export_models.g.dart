// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'export_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

BiReportExportRequest _$BiReportExportRequestFromJson(
  Map<String, dynamic> json,
) => BiReportExportRequest(
  format: $enumDecode(_$ExportFormatEnumMap, json['format']),
);

Map<String, dynamic> _$BiReportExportRequestToJson(
  BiReportExportRequest instance,
) => <String, dynamic>{'format': _$ExportFormatEnumMap[instance.format]!};

const _$ExportFormatEnumMap = {
  ExportFormat.pdf: 'pdf',
  ExportFormat.excel: 'excel',
};

BiReportExportResponse _$BiReportExportResponseFromJson(
  Map<String, dynamic> json,
) => BiReportExportResponse(
  success: json['success'] as bool,
  jobId: json['job_id'] as String,
  message: json['message'] as String,
);

Map<String, dynamic> _$BiReportExportResponseToJson(
  BiReportExportResponse instance,
) => <String, dynamic>{
  'success': instance.success,
  'job_id': instance.jobId,
  'message': instance.message,
};

BiReportExportStatusResponse _$BiReportExportStatusResponseFromJson(
  Map<String, dynamic> json,
) => BiReportExportStatusResponse(
  success: json['success'] as bool,
  jobId: json['job_id'] as String,
  state: json['state'] as String,
  progress: (json['progress'] as num).toInt(),
  createdAt: json['createdAt'] as String,
  processedOn: json['processedOn'] as String?,
  finishedOn: json['finishedOn'] as String?,
  fileName: json['fileName'] as String?,
  failedReason: json['failedReason'] as String?,
);

Map<String, dynamic> _$BiReportExportStatusResponseToJson(
  BiReportExportStatusResponse instance,
) => <String, dynamic>{
  'success': instance.success,
  'job_id': instance.jobId,
  'state': instance.state,
  'progress': instance.progress,
  'createdAt': instance.createdAt,
  'processedOn': instance.processedOn,
  'finishedOn': instance.finishedOn,
  'fileName': instance.fileName,
  'failedReason': instance.failedReason,
};

BiReportExportDownloadInfo _$BiReportExportDownloadInfoFromJson(
  Map<String, dynamic> json,
) => BiReportExportDownloadInfo(
  jobId: json['jobId'] as String,
  fileName: json['fileName'] as String,
  downloadUrl: json['downloadUrl'] as String,
  format: $enumDecode(_$ExportFormatEnumMap, json['format']),
);

Map<String, dynamic> _$BiReportExportDownloadInfoToJson(
  BiReportExportDownloadInfo instance,
) => <String, dynamic>{
  'jobId': instance.jobId,
  'fileName': instance.fileName,
  'downloadUrl': instance.downloadUrl,
  'format': _$ExportFormatEnumMap[instance.format]!,
};
