import 'dart:convert';
import 'dart:typed_data';

import 'package:smlaicloud/model/pagination.dart';
import 'package:smlaicloud/model/stock_balance_import_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/stock_balance_import_repository.dart';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'stock_balance_event.dart';
part 'stock_balance_state.dart';

class StockBalanceBloc extends Bloc<StockBalanceEvent, StockBalanceState> {
  final StockBalanceImportRepository _stockBalanceImportRepository;

  StockBalanceBloc(
      {required StockBalanceImportRepository stockBalanceRepository})
      : _stockBalanceImportRepository = StockBalanceImportRepository(),
        super(StockBalanceInitial()) {
    on<UploadFileExcel>(onUploadFileExcel);
    on<LoadStockBalanceImportByTaskid>(onLoadStockBalanceImportByTaskid);
    on<DeleteDetailByGuid>(onDeleteDetailByGuid);
    on<DeleteTaskid>(onDeleteTaskid);
    on<UpdateDetail>(onUpdateDetail);
    on<AddDetail>(onAddDetail);
    on<LoadTotal>(onLoadTotal);
    on<SaveTransStockBalance>(onSaveTransStockBalance);
    on<LoadTransStockBalanceDetailByDocno>(
        onLoadTransStockBalanceDetailByDocno);
  }

  /// upload file excel
  void onUploadFileExcel(
      UploadFileExcel event, Emitter<StockBalanceState> emit) async {
    emit(UploadFileExcelInProgress());
    try {
      final result = await _stockBalanceImportRepository.uploadFileExcel(
          event.file, event.filename);

      if (result.success) {
        UploadSuccessModel uploadSuccessModel =
            UploadSuccessModel(success: result.success, id: result.id);
        emit(UploadFileExcelSuccess(response: uploadSuccessModel));
      } else {
        emit(const UploadFileExcelFailed(message: 'Upload FileExcel Failure'));
      }
    } catch (e) {
      emit(UploadFileExcelFailed(message: e.toString()));
    }
  }

  /// load stock balance import by taskid
  void onLoadStockBalanceImportByTaskid(LoadStockBalanceImportByTaskid event,
      Emitter<StockBalanceState> emit) async {
    emit(UploadFileExcelInProgress());
    try {
      final result = await _stockBalanceImportRepository.getStockBalanceImport(
          event.taskid, event.q, event.limit, event.page);

      if (result.success) {
        List<StockBalanceImportModel> stockBalanceImportModel =
            (result.data as List)
                .map((tables) => StockBalanceImportModel.fromJson(tables))
                .toList();
        Page page = result.page ?? Page.empty;

        Pagination pagination = Pagination(
          page: page.page,
          perPage: page.perPage,
          total: page.total,
          totalPage: page.totalPage,
          next: 0,
          prev: 0,
        );

        emit(LoadStockBalanceImportByTaskidSuccess(
          data: stockBalanceImportModel,
          pagination: pagination,
        ));
      } else {
        emit(const LoadStockBalanceImportByTaskidFailed(
            message: 'Load Stock Balance Import By Taskid Failure'));
      }
    } catch (e) {
      emit(LoadStockBalanceImportByTaskidFailed(message: e.toString()));
    }
  }

  /// delete  detail stock balance import by guid
  void onDeleteDetailByGuid(
      DeleteDetailByGuid event, Emitter<StockBalanceState> emit) async {
    emit(DeleteDetailByGuidInProgress());
    try {
      final result =
          await _stockBalanceImportRepository.deleteDetailByGuid(event.guid);

      if (result.success) {
        emit(DeleteDetailByGuidSuccess());
      } else {
        emit(const DeleteDetailByGuidFailed(
            message: 'Delete Stock Balance Import By Guid Failure'));
      }
    } catch (e) {
      emit(DeleteDetailByGuidFailed(message: e.toString()));
    }
  }

  /// delete taskid
  void onDeleteTaskid(
      DeleteTaskid event, Emitter<StockBalanceState> emit) async {
    emit(DeleteTaskidInProgress());
    try {
      final result =
          await _stockBalanceImportRepository.deleteTaskid(event.taskid);

      if (result.success) {
        emit(DeleteTaskidSuccess());
      } else {
        emit(const DeleteTaskidFailed(message: 'Delete Task ID Failure'));
      }
    } catch (e) {
      emit(DeleteTaskidFailed(message: e.toString()));
    }
  }

  /// update detail
  void onUpdateDetail(
      UpdateDetail event, Emitter<StockBalanceState> emit) async {
    emit(UpdateDetailInProgress());
    try {
      final result = await _stockBalanceImportRepository.updateDetail(
          event.guid, event.stockBalanceImportModel);

      if (result.success) {
        emit(UpdateDetailSuccess());
      } else {
        emit(const UpdateDetailFailed(message: 'Update Detail Failure'));
      }
    } catch (e) {
      emit(UpdateDetailFailed(message: e.toString()));
    }
  }

  /// add detail
  void onAddDetail(AddDetail event, Emitter<StockBalanceState> emit) async {
    emit(AddDetailInProgress());
    try {
      final result = await _stockBalanceImportRepository
          .addDetail(event.stockBalanceImportModel);

      if (result.success) {
        emit(AddDetailSuccess());
      } else {
        emit(const AddDetailFailed(message: 'Add Detail Failure'));
      }
    } catch (e) {
      emit(AddDetailFailed(message: e.toString()));
    }
  }

  /// load total
  void onLoadTotal(LoadTotal event, Emitter<StockBalanceState> emit) async {
    emit(LoadTotalInProgress());
    try {
      final result = await _stockBalanceImportRepository.getTotal(event.taskid);

      if (result.success) {
        TotalModel total = TotalModel.fromJson(result.data);
        emit(LoadTotalSuccess(total: total));
      } else {
        emit(const LoadTotalFailed(message: 'Load Total Failure'));
      }
    } catch (e) {
      emit(LoadTotalFailed(message: e.toString()));
    }
  }

  /// save trans stock balance
  void onSaveTransStockBalance(
      SaveTransStockBalance event, Emitter<StockBalanceState> emit) async {
    emit(SaveTransStockBalanceInProgress());
    try {
      final result = await _stockBalanceImportRepository.saveTransStockBalance(
          event.taskid, event.transactionModel);

      if (result.success) {
        emit(SaveTransStockBalanceSuccess());
      } else {
        emit(const SaveTransStockBalanceFailed(
            message: 'Save Transaction Failure'));
      }
    } catch (e) {
      // Parsing JSON response
      Map<String, dynamic> data = jsonDecode(e.toString());

      // Extracting the message
      String message = data['message'];

      emit(SaveTransStockBalanceFailed(message: message));
    }
  }

  /// load tran detail stock balance by docno
  void onLoadTransStockBalanceDetailByDocno(
      LoadTransStockBalanceDetailByDocno event,
      Emitter<StockBalanceState> emit) async {
    emit(LoadTransStockBalanceDetailByDocnoInProgress());
    try {
      final result = await _stockBalanceImportRepository
          .loadTransStockBalanceDetailByDocno(
              event.docno, event.q, event.limit, event.page);

      if (result.success) {
        List<TransactionDetailModel> stockBalanceImportModel =
            (result.data as List)
                .map((tables) => TransactionDetailModel.fromJson(tables))
                .toList();
        Page page = result.page ?? Page.empty;

        Pagination pagination = Pagination(
          page: page.page,
          perPage: page.perPage,
          total: page.total,
          totalPage: page.totalPage,
          next: 0,
          prev: 0,
        );

        emit(LoadTransStockBalanceDetailByDocnoSuccess(
          data: stockBalanceImportModel,
          pagination: pagination,
        ));
      } else {
        emit(const LoadTransStockBalanceDetailByDocnoFailed(
            message: 'Load Stock Balance Import By Taskid Failure'));
      }
    } catch (e) {
      emit(LoadTransStockBalanceDetailByDocnoFailed(message: e.toString()));
    }
  }
}
