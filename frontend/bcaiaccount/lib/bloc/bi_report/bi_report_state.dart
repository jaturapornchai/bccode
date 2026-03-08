part of 'bi_report_bloc.dart';

sealed class BiReportState extends Equatable {
  const BiReportState();

  @override
  List<Object> get props => [];
}

final class BiReportInitial extends BiReportState {}

// Complete Report Generation States
final class BiReportGenerating extends BiReportState {
  final BiReportType reportType;

  const BiReportGenerating({required this.reportType});

  @override
  List<Object> get props => [reportType];
}

final class BiReportGenerateProgress extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final int progress;
  final String statusMessage;

  const BiReportGenerateProgress({
    required this.reportType,
    required this.jobId,
    required this.progress,
    required this.statusMessage,
  });

  @override
  List<Object> get props => [reportType, jobId, progress, statusMessage];
}

final class BiReportGenerateSuccess extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final List<dynamic> data;
  final BiReportMeta meta;

  const BiReportGenerateSuccess({
    required this.reportType,
    required this.jobId,
    required this.data,
    required this.meta,
  });

  @override
  List<Object> get props => [reportType, jobId, data, meta];
}

final class BiReportGenerateFailure extends BiReportState {
  final String message;
  final int? errorCode;

  const BiReportGenerateFailure({required this.message, this.errorCode});

  @override
  List<Object> get props => [message, errorCode ?? 0];
}

final class BiReportDetailLoading extends BiReportState {
  final BiReportType reportType;
  final String jobId;

  const BiReportDetailLoading({required this.reportType, required this.jobId});

  @override
  List<Object> get props => [reportType, jobId];
}

final class BiReportDetailSuccess extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final List<dynamic> data;
  final BiReportMeta meta; // เปลี่ยนกลับเป็น required

  const BiReportDetailSuccess({
    required this.reportType,
    required this.jobId,
    required this.data,
    required this.meta, // เปลี่ยนกลับเป็น required
  });

  @override
  List<Object> get props => [reportType, jobId, data, meta];
}

final class BiReportDetailFailure extends BiReportState {
  final String jobId;
  final String message;
  final int? errorCode;

  const BiReportDetailFailure({
    required this.jobId,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [jobId, message, errorCode ?? 0];
}

// Get Report Detail by DocNo States
final class BiReportDetailByDocNoLoading extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final String docNo;

  const BiReportDetailByDocNoLoading({
    required this.reportType,
    required this.jobId,
    required this.docNo,
  });

  @override
  List<Object> get props => [reportType, jobId, docNo];
}

final class BiReportDetailByDocNoSuccess extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final String docNo;
  final List<dynamic> data;

  const BiReportDetailByDocNoSuccess({
    required this.reportType,
    required this.jobId,
    required this.docNo,
    required this.data,
  });

  @override
  List<Object> get props => [reportType, jobId, docNo, data];
}

final class BiReportDetailByDocNoFailure extends BiReportState {
  final String jobId;
  final String docNo;
  final String message;
  final int? errorCode;

  const BiReportDetailByDocNoFailure({
    required this.jobId,
    required this.docNo,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [jobId, docNo, message, errorCode ?? 0];
}

// Get Report Detail by DocDate States

final class BiReportDetailByDocDateLoading extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final String docDate;

  const BiReportDetailByDocDateLoading({
    required this.reportType,
    required this.jobId,
    required this.docDate,
  });

  @override
  List<Object> get props => [reportType, jobId, docDate];
}

final class BiReportDetailByDocDateSuccess extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final String docDate;
  final List<dynamic> data;

  const BiReportDetailByDocDateSuccess({
    required this.reportType,
    required this.jobId,
    required this.docDate,
    required this.data,
  });

  @override
  List<Object> get props => [reportType, jobId, docDate, data];
}

final class BiReportDetailByDocDateFailure extends BiReportState {
  final String jobId;
  final String docDate;
  final String message;
  final int? errorCode;

  const BiReportDetailByDocDateFailure({
    required this.jobId,
    required this.docDate,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [jobId, docDate, message, errorCode ?? 0];
}

// Get Report Summary States
final class BiReportSummaryLoading extends BiReportState {
  final BiReportType reportType;
  final String jobId;

  const BiReportSummaryLoading({required this.reportType, required this.jobId});

  @override
  List<Object> get props => [reportType, jobId];
}

final class BiReportSummarySuccess extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final SaleReportSummary? summaryData;
  final SaleDailyReportSummary? dailySummaryData;
  final StockMovmentSummaryModel? stockMovementSummaryData;
  final PaymentDailySummaryModel? paymentDailySummaryData;
  final SaleReturnSummaryModel? saleReturnSummaryData;
  final StockBalanceSummaryModel? stockBalanceSummaryData;
  final PurchasePartialSummaryModel? purchasePartialSummaryData;
  final GrossProfitByDocumentSummary? grossProfitSummaryData;
  final GrossProfitByProductSummary? grossProfitByProductSummaryData;
  final VatSaleSummary? vatSaleSummaryData;
  final VatBuySummary? vatBuySummaryData;

  const BiReportSummarySuccess({
    required this.reportType,
    required this.jobId,
    this.summaryData,
    this.dailySummaryData,
    this.stockMovementSummaryData,
    this.paymentDailySummaryData,
    this.saleReturnSummaryData,
    this.stockBalanceSummaryData,
    this.purchasePartialSummaryData,
    this.grossProfitSummaryData,
    this.grossProfitByProductSummaryData,
    this.vatSaleSummaryData,
    this.vatBuySummaryData,
  });

  @override
  List<Object> get props => [
    reportType,
    jobId,
    summaryData ?? '',
    dailySummaryData ?? '',
    stockMovementSummaryData ?? '',
    paymentDailySummaryData ?? '',
    saleReturnSummaryData ?? '',
    stockBalanceSummaryData ?? '',
    purchasePartialSummaryData ?? '',
    grossProfitSummaryData ?? '',
    grossProfitByProductSummaryData ?? '',
    vatSaleSummaryData ?? '',
    vatBuySummaryData ?? '',
  ];
}

final class BiReportSummaryFailure extends BiReportState {
  final String jobId;
  final String message;
  final int? errorCode;

  const BiReportSummaryFailure({
    required this.jobId,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [jobId, message, errorCode ?? 0];
}

// ==================== EXPORT STATES ====================

// Export Initial State
final class BiReportExportInitial extends BiReportState {}

// Export Submit States
final class BiReportExportSubmitting extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final ExportFormat format;

  const BiReportExportSubmitting({
    required this.reportType,
    required this.jobId,
    required this.format,
  });

  @override
  List<Object> get props => [reportType, jobId, format];
}

final class BiReportExportSubmitSuccess extends BiReportState {
  final BiReportType reportType;
  final String originalJobId;
  final String exportJobId;
  final ExportFormat format;
  final String message;

  const BiReportExportSubmitSuccess({
    required this.reportType,
    required this.originalJobId,
    required this.exportJobId,
    required this.format,
    required this.message,
  });

  @override
  List<Object> get props => [
    reportType,
    originalJobId,
    exportJobId,
    format,
    message,
  ];
}

final class BiReportExportSubmitFailure extends BiReportState {
  final BiReportType reportType;
  final String jobId;
  final ExportFormat format;
  final String message;
  final int? errorCode;

  const BiReportExportSubmitFailure({
    required this.reportType,
    required this.jobId,
    required this.format,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [
    reportType,
    jobId,
    format,
    message,
    errorCode ?? 0,
  ];
}

// Export Status States
final class BiReportExportStatusChecking extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;

  const BiReportExportStatusChecking({
    required this.reportType,
    required this.exportJobId,
  });

  @override
  List<Object> get props => [reportType, exportJobId];
}

final class BiReportExportStatusProgress extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;
  final ExportState state;
  final int progress;
  final String statusMessage;

  const BiReportExportStatusProgress({
    required this.reportType,
    required this.exportJobId,
    required this.state,
    required this.progress,
    required this.statusMessage,
  });

  @override
  List<Object> get props => [
    reportType,
    exportJobId,
    state,
    progress,
    statusMessage,
  ];
}

final class BiReportExportReady extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;
  final String fileName;
  final String downloadUrl;
  final ExportFormat format;

  const BiReportExportReady({
    required this.reportType,
    required this.exportJobId,
    required this.fileName,
    required this.downloadUrl,
    required this.format,
  });

  @override
  List<Object> get props => [
    reportType,
    exportJobId,
    fileName,
    downloadUrl,
    format,
  ];
}

final class BiReportExportStatusFailure extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;
  final String message;
  final int? errorCode;

  const BiReportExportStatusFailure({
    required this.reportType,
    required this.exportJobId,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [reportType, exportJobId, message, errorCode ?? 0];
}

// Export Download States
final class BiReportExportDownloading extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;
  final String fileName;

  const BiReportExportDownloading({
    required this.reportType,
    required this.exportJobId,
    required this.fileName,
  });

  @override
  List<Object> get props => [reportType, exportJobId, fileName];
}

final class BiReportExportDownloadSuccess extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;
  final String fileName;
  final String filePath;

  const BiReportExportDownloadSuccess({
    required this.reportType,
    required this.exportJobId,
    required this.fileName,
    required this.filePath,
  });

  @override
  List<Object> get props => [reportType, exportJobId, fileName, filePath];
}

final class BiReportExportDownloadFailure extends BiReportState {
  final BiReportType reportType;
  final String exportJobId;
  final String fileName;
  final String message;
  final int? errorCode;

  const BiReportExportDownloadFailure({
    required this.reportType,
    required this.exportJobId,
    required this.fileName,
    required this.message,
    this.errorCode,
  });

  @override
  List<Object> get props => [
    reportType,
    exportJobId,
    fileName,
    message,
    errorCode ?? 0,
  ];
}

// Export Cancelled State
final class BiReportExportCancelled extends BiReportState {
  final String message;

  const BiReportExportCancelled({this.message = 'การ Export ถูกยกเลิก'});

  @override
  List<Object> get props => [message];
}
