part of 'bi_report_bloc.dart';

sealed class BiReportEvent extends Equatable {
  const BiReportEvent();

  @override
  List<Object> get props => [];
}

// Complete report generation (Submit → Poll → Get Details)
class GenerateBiReportRequested extends BiReportEvent {
  final BiReportType reportType;
  final dynamic
  conditions; // Can be ReportConditions, ProductReportConditions, etc.
  final String token;
  final Duration? pollInterval;
  final Duration? timeout;
  final int? page;
  final int? size;

  const GenerateBiReportRequested({
    required this.reportType,
    required this.conditions,
    required this.token,
    this.pollInterval,
    this.timeout,
    this.page,
    this.size,
  });

  @override
  List<Object> get props => [reportType, conditions, token];
}

class GetBiReportDetailRequested extends BiReportEvent {
  final BiReportType reportType;
  final String jobId;
  final String token;
  final int? page;
  final int? size;

  const GetBiReportDetailRequested({
    required this.reportType,
    required this.jobId,
    required this.token,
    this.page,
    this.size,
  });

  @override
  List<Object> get props => [reportType, jobId, token];
}

// Get report summary
class GetBiReportSummaryRequested extends BiReportEvent {
  final BiReportType reportType;
  final String jobId;
  final String token;

  const GetBiReportSummaryRequested({
    required this.reportType,
    required this.jobId,
    required this.token,
  });

  @override
  List<Object> get props => [reportType, jobId, token];
}

// Get report detail by docno
class GetBiReportDetailByDocNoRequested extends BiReportEvent {
  final BiReportType reportType;
  final String jobId;
  final String docNo;
  final String token;

  const GetBiReportDetailByDocNoRequested({
    required this.reportType,
    required this.jobId,
    required this.docNo,
    required this.token,
  });

  @override
  List<Object> get props => [reportType, jobId, docNo, token];
}

// Get report detail by docDate
class GetBiReportDetailByDocDateRequested extends BiReportEvent {
  final BiReportType reportType;
  final String jobId;
  final String docDate;
  final String token;

  const GetBiReportDetailByDocDateRequested({
    required this.reportType,
    required this.jobId,
    required this.docDate,
    required this.token,
  });

  @override
  List<Object> get props => [reportType, jobId, docDate, token];
}

// Cancel ongoing report generation
class CancelBiReportRequested extends BiReportEvent {
  final String jobId;

  const CancelBiReportRequested({required this.jobId});

  @override
  List<Object> get props => [jobId];
}

// Reset state
class ResetBiReportState extends BiReportEvent {
  const ResetBiReportState();
}

// ==================== EXPORT EVENTS ====================

// Submit export job
class SubmitBiReportExportRequested extends BiReportEvent {
  final BiReportType reportType;
  final String jobId;
  final ExportFormat format;
  final String token;

  const SubmitBiReportExportRequested({
    required this.reportType,
    required this.jobId,
    required this.format,
    required this.token,
  });

  @override
  List<Object> get props => [reportType, jobId, format, token];
}

// Check export status
class CheckBiReportExportStatusRequested extends BiReportEvent {
  final BiReportType reportType;
  final String exportJobId;
  final String token;

  const CheckBiReportExportStatusRequested({
    required this.reportType,
    required this.exportJobId,
    required this.token,
  });

  @override
  List<Object> get props => [reportType, exportJobId, token];
}

// Download export file
class DownloadBiReportExportRequested extends BiReportEvent {
  final BiReportType reportType;
  final String exportJobId;
  final String fileName;
  final String token;

  const DownloadBiReportExportRequested({
    required this.reportType,
    required this.exportJobId,
    required this.fileName,
    required this.token,
  });

  @override
  List<Object> get props => [reportType, exportJobId, fileName, token];
}

// Cancel export process
class CancelBiReportExportRequested extends BiReportEvent {
  const CancelBiReportExportRequested();
}

// Reset export state
class ResetBiReportExportState extends BiReportEvent {
  const ResetBiReportExportState();
}

// Download export file directly
class DownloadBiReportExportFile extends BiReportEvent {
  final BiReportType reportType;
  final String exportJobId;
  final String token;
  final String fileName;

  const DownloadBiReportExportFile({
    required this.reportType,
    required this.exportJobId,
    required this.token,
    required this.fileName,
  });

  @override
  List<Object> get props => [reportType, exportJobId, token, fileName];
}
