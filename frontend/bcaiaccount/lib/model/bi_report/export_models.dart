// Export Models for BiReport Export functionality

import 'package:json_annotation/json_annotation.dart';
import '../../global.dart' as global;

part 'export_models.g.dart';

/// Enum for export formats
@JsonEnum()
enum ExportFormat {
  @JsonValue('pdf')
  pdf,
  @JsonValue('excel')
  excel,
}

extension ExportFormatExtension on ExportFormat {
  String get displayName {
    switch (this) {
      case ExportFormat.pdf:
        return 'PDF';
      case ExportFormat.excel:
        return 'Excel';
    }
  }

  String get value {
    switch (this) {
      case ExportFormat.pdf:
        return 'pdf';
      case ExportFormat.excel:
        return 'excel';
    }
  }
}

/// Export request model
@JsonSerializable()
class BiReportExportRequest {
  final ExportFormat format;

  const BiReportExportRequest({required this.format});

  factory BiReportExportRequest.fromJson(Map<String, dynamic> json) =>
      _$BiReportExportRequestFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportExportRequestToJson(this);
}

/// Export response model (when submitting export job)
@JsonSerializable()
class BiReportExportResponse {
  final bool success;
  @JsonKey(name: 'job_id')
  final String jobId;
  final String message;

  const BiReportExportResponse({
    required this.success,
    required this.jobId,
    required this.message,
  });

  factory BiReportExportResponse.fromJson(Map<String, dynamic> json) =>
      _$BiReportExportResponseFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportExportResponseToJson(this);
}

/// Export status response model
@JsonSerializable()
class BiReportExportStatusResponse {
  final bool success;
  @JsonKey(name: 'job_id')
  final String jobId;
  final String state;
  final int progress;
  final String createdAt;
  final String? processedOn;
  final String? finishedOn;
  final String? fileName;
  final String? failedReason;

  const BiReportExportStatusResponse({
    required this.success,
    required this.jobId,
    required this.state,
    required this.progress,
    required this.createdAt,
    this.processedOn,
    this.finishedOn,
    this.fileName,
    this.failedReason,
  });

  factory BiReportExportStatusResponse.fromJson(Map<String, dynamic> json) =>
      _$BiReportExportStatusResponseFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportExportStatusResponseToJson(this);
}

/// Export download info model
@JsonSerializable()
class BiReportExportDownloadInfo {
  final String jobId;
  final String fileName;
  final String downloadUrl;
  final ExportFormat format;

  const BiReportExportDownloadInfo({
    required this.jobId,
    required this.fileName,
    required this.downloadUrl,
    required this.format,
  });

  factory BiReportExportDownloadInfo.fromJson(Map<String, dynamic> json) =>
      _$BiReportExportDownloadInfoFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportExportDownloadInfoToJson(this);
}

/// Export state enum
enum ExportState { waiting, progress, completed, failed }

extension ExportStateExtension on ExportState {
  String get displayName {
    switch (this) {
      case ExportState.waiting:
        return 'รอการดำเนินการ';
      case ExportState.progress:
        return 'กำลังสร้างไฟล์';
      case ExportState.completed:
        return global.language('done');
      case ExportState.failed:
        return global.language('failed');
    }
  }
}

/// Helper function to convert string to ExportState
ExportState exportStateFromString(String state) {
  switch (state.toLowerCase()) {
    case 'waiting':
      return ExportState.waiting;
    case 'progress':
      return ExportState.progress;
    case 'completed':
      return ExportState.completed;
    case 'failed':
      return ExportState.failed;
    default:
      return ExportState.waiting;
  }
}
