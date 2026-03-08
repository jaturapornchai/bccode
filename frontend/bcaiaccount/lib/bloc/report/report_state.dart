part of 'report_bloc.dart';

abstract class ReportState extends Equatable {
  const ReportState();

  @override
  List<Object> get props => [];
}

class ReportInitial extends ReportState {}

class GetReportInProgress extends ReportState {}

class GetReportSaleByDateSuccess extends ReportState {
  final Page page;
  final List<ReportSaleByDateModel> reportSaleByDateModels;

  const GetReportSaleByDateSuccess(
      {required this.page, required this.reportSaleByDateModels});

  @override
  List<Object> get props => [page, reportSaleByDateModels];
}

class GetReportReceiveMoneySuccess extends ReportState {
  final Page page;
  final List<ReportReceiveMoneyModel> reportReceiveMoneyModels;

  const GetReportReceiveMoneySuccess(
      {required this.page, required this.reportReceiveMoneyModels});

  @override
  List<Object> get props => [page, reportReceiveMoneyModels];
}

class GetReportFailed extends ReportState {
  final String message;

  const GetReportFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class GetReportSaleInvoiceSuccess extends ReportState {
  final Page page;
  final List<TransactionModel> reportSaleInvoiceModels;

  const GetReportSaleInvoiceSuccess(
      {required this.page, required this.reportSaleInvoiceModels});

  @override
  List<Object> get props => [page, reportSaleInvoiceModels];
}

class DownloadReportInProgress extends ReportState {}

class DownloadReportSuccess extends ReportState {
  final String savePath;
  final String jobId;
  final String guid;
  final int index;
  const DownloadReportSuccess({
    required this.savePath,
    required this.jobId,
    required this.guid,
    required this.index,
  });

  @override
  List<Object> get props => [savePath, jobId, guid, index];
}

class DownloadReportFailed extends ReportState {
  final String message;

  const DownloadReportFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class FileStatusGetInProgress extends ReportState {}

class FileStatusGetSuccess extends ReportState {
  final List<LogDownloadParthModel> logDownloadParthModels;

  const FileStatusGetSuccess({required this.logDownloadParthModels});

  @override
  List<Object> get props => [logDownloadParthModels];
}

class FileStatusGetFailed extends ReportState {
  final String message;

  const FileStatusGetFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class FileStatusDeleteByIdInProgress extends ReportState {}

class FileStatusDeleteByIdDeleteSuccess extends ReportState {}

class FileStatusDeleteByIdDeleteFailed extends ReportState {
  final String message;

  const FileStatusDeleteByIdDeleteFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class FileStatusDeleteByMenuInProgress extends ReportState {}

class FileStatusDeleteByMenuDeleteSuccess extends ReportState {}

class FileStatusDeleteByMenuDeleteFailed extends ReportState {
  final String message;

  const FileStatusDeleteByMenuDeleteFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class FileStatusSaveInProgress extends ReportState {}

class FileStatusSaveSuccess extends ReportState {}

class FileStatusSaveFailed extends ReportState {
  final String message;

  const FileStatusSaveFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class FileStatusUpdateInProgress extends ReportState {}

class FileStatusUpdateSuccess extends ReportState {}

class FileStatusUpdateFailed extends ReportState {
  final String message;

  const FileStatusUpdateFailed({required this.message});

  @override
  List<Object> get props => [message];
}
