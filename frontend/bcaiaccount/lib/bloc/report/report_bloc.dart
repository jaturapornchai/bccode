import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/imports_repositories.dart';
import 'package:smlaicloud/model/report_list_model.dart';
import 'package:smlaicloud/model/report_main_model.dart';
import 'package:smlaicloud/model/select_colums_csv_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/file_status_repository.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/columns_csv_list.dart' as columns_csv;

part 'report_event.dart';
part 'report_state.dart';

class ReportBloc extends Bloc<ReportEvent, ReportState> {
  final ReportRepository _reportRepository;
  final FileStatusRepository _fileStatusRepository;

  ReportBloc(
      {required ReportRepository reportRepository,
      required FileStatusRepository fileStatusRepository})
      : _reportRepository = reportRepository,
        _fileStatusRepository = fileStatusRepository,
        super(ReportInitial()) {
    on<GetReport>(_onGetReport);
    on<DownloadReport>(_onDownloadReport);
    on<FileStatusGetList>(onFileStatusGetList);
    on<FileStatusDeleteById>(onFileStatusDeleteById);
    on<FileStatusDeleteByMenu>(onFileStatusDeleteByMenu);
    on<FileStatusSave>(onFileStatusSave);
    on<FileStatusUpdate>(onFileStatusUpdate);
  }
  void _onGetReport(GetReport event, Emitter<ReportState> emit) async {
    emit(GetReportInProgress());
    try {
      final results = await _reportRepository.getReport(
          event.type, event.fromdate, event.todate, event.page, event.perpage);

      if (results.success) {
        if (event.type == ReportEnum.salebydate) {
          /// รายงานขายตามวันที่
          List<ReportSaleByDateModel> reportSaleByDateModels =
              (results.data as List)
                  .map((reportSaleByDateModels) =>
                      ReportSaleByDateModel.fromJson(reportSaleByDateModels))
                  .toList();
          emit(GetReportSaleByDateSuccess(
              page: results.page!,
              reportSaleByDateModels: reportSaleByDateModels));
        } else if (event.type == ReportEnum.receivemoney) {
          /// รายงานรับเงินตามวันที่
          List<ReportReceiveMoneyModel> reportReceiveMoneyModels = (results.data
                  as List)
              .map((reportReceiveMoneyModels) =>
                  ReportReceiveMoneyModel.fromJson(reportReceiveMoneyModels))
              .toList();
          emit(GetReportReceiveMoneySuccess(
              page: results.page!,
              reportReceiveMoneyModels: reportReceiveMoneyModels));
        } else if (event.type == ReportEnum.saleinvoice) {
          /// รายงานขาย
          List<TransactionModel> reportSaleInvoiceModels =
              (results.data as List)
                  .map((reportSaleInvoiceModels) =>
                      TransactionModel.fromJson(reportSaleInvoiceModels))
                  .toList();
          emit(GetReportSaleInvoiceSuccess(
              page: results.page!,
              reportSaleInvoiceModels: reportSaleInvoiceModels));
        } else {
          emit(const GetReportFailed(message: 'Report Not Found'));
        }
      } else {
        emit(const GetReportFailed(message: 'Report Not Found'));
      }
    } catch (e) {
      emit(GetReportFailed(message: e.toString()));
    }
  }

  void _onDownloadReport(
      DownloadReport event, Emitter<ReportState> emit) async {
    emit(DownloadReportInProgress());
    try {
      List<ListColumsCsvModel> columncsv = [];

      if (event.listcolumscsv.isEmpty) {
        emit(const DownloadReportFailed(message: 'Please select columns'));
        return;
      } else {
        if (event.listcolumscsv.length == columns_csv.listColumns.length) {
          columncsv.clear();
        } else {
          columncsv = event.listcolumscsv;
        }
      }

      final results = await _reportRepository.getPDFDownload(
        event.type,
        event.fromdate,
        event.todate,
        event.showDetail,
        event.showSumByDate,
        event.search,
        event.yearnum,
        event.monthnum,
        event.fromcustcode,
        event.tocustcode,
        event.branch,
        event.iscancel,
        event.fromsalecode,
        event.tosalecode,
        event.inquirytype,
        event.ispos,
        event.frombarcode,
        event.tobarcode,
        event.fromgroup,
        event.togroup,
        event.iscost,
        event.barcode,
        event.typefile,
        columncsv,
      );

      if (results.success) {
        final logDownloadParthModel = LogDownloadParthModel(
          xorder: event.xorder,
          guidfixed: event.guid,
          menu: event.type.toString(),
          path: results.data['downloadLink'],
          status: 'processing',
          jobid: results.data['jobId'],
          filter: FilterrReportModel(
            type: event.type,
            fromdate: event.fromdate,
            todate: event.todate,
            showdetail: event.showDetail,
            showsumbydate: event.showSumByDate,
            search: event.search,
            yearnum: event.yearnum,
            monthnum: event.monthnum,
            fromcustcode: event.fromcustcode,
            tocustcode: event.tocustcode,
            branch: event.branch,
            iscancel: event.iscancel,
            iscost: event.iscost,
            fromsalecode: event.fromsalecode,
            tosalecode: event.tosalecode,
            inquirytype: event.inquirytype,
            ispos: event.ispos,
            barcode: event.barcode,
            frombarcode: event.frombarcode,
            tobarcode: event.tobarcode,
            fromgroup: event.fromgroup,
            togroup: event.togroup,
            typefile: event.typefile,
            listcolumscsv: columncsv,
          ),
        );
        ApiResponse result =
            await _fileStatusRepository.saveFileStatue(logDownloadParthModel);
        if (result.success) {
          emit(DownloadReportSuccess(
              savePath: results.data['downloadLink'],
              jobId: results.data['jobId'],
              guid: event.guid,
              index: event.xorder));
        } else {
          emit(DownloadReportFailed(message: result.message));
        }
      } else {
        emit(const DownloadReportFailed(message: 'Report Not Found'));
      }
    } catch (e) {
      emit(DownloadReportFailed(message: e.toString()));
    }
  }

  void onFileStatusGetList(
      FileStatusGetList event, Emitter<ReportState> emit) async {
    emit(FileStatusGetInProgress());

    try {
      final results = await _fileStatusRepository.getFileStatue(
          offset: event.offset, limit: event.limit, menu: event.menu);

      if (results.success) {
        List<LogDownloadParthModel> logDownloadParthModel =
            (results.data as List)
                .map((data) => LogDownloadParthModel.fromJson(data))
                .toList();
        emit(FileStatusGetSuccess(
            logDownloadParthModels: logDownloadParthModel));
      } else {
        emit(const FileStatusGetFailed(message: 'File Not Found'));
      }
    } catch (e) {
      emit(FileStatusGetFailed(message: e.toString()));
    }
  }

  void onFileStatusDeleteById(
      FileStatusDeleteById event, Emitter<ReportState> emit) async {
    emit(FileStatusDeleteByIdInProgress());

    try {
      final results =
          await _fileStatusRepository.deleteFileStatueById(guid: event.guid);

      if (results.success) {
        emit(FileStatusDeleteByIdDeleteSuccess());
      } else {
        emit(const FileStatusDeleteByIdDeleteFailed(message: 'File Not Found'));
      }
    } catch (e) {
      emit(FileStatusDeleteByIdDeleteFailed(message: e.toString()));
    }
  }

  void onFileStatusDeleteByMenu(
      FileStatusDeleteByMenu event, Emitter<ReportState> emit) async {
    emit(FileStatusDeleteByMenuInProgress());

    try {
      final results =
          await _fileStatusRepository.deleteFileStatueByMenu(menu: event.menu);

      if (results.success) {
        emit(FileStatusDeleteByMenuDeleteSuccess());
      } else {
        emit(const FileStatusDeleteByMenuDeleteFailed(
            message: 'File Not Found'));
      }
    } catch (e) {
      emit(FileStatusDeleteByMenuDeleteFailed(message: e.toString()));
    }
  }

  void onFileStatusSave(FileStatusSave event, Emitter<ReportState> emit) async {
    emit(FileStatusSaveInProgress());

    try {
      final results = await _fileStatusRepository
          .saveFileStatue(event.logDownloadParthModel);

      if (results.success) {
        emit(FileStatusSaveSuccess());
      } else {
        emit(const FileStatusSaveFailed(message: 'File Not Found'));
      }
    } catch (e) {
      emit(FileStatusSaveFailed(message: e.toString()));
    }
  }

  void onFileStatusUpdate(
      FileStatusUpdate event, Emitter<ReportState> emit) async {
    emit(FileStatusUpdateInProgress());

    try {
      final results = await _fileStatusRepository.updateFileStatue(
          event.logDownloadParthModel, event.guid);

      if (results.success) {
        emit(FileStatusUpdateSuccess());
      } else {
        emit(const FileStatusUpdateFailed(message: 'File Not Found'));
      }
    } catch (e) {
      emit(FileStatusUpdateFailed(message: e.toString()));
    }
  }
}
