import 'dart:convert';

import 'package:smlaicloud/model/journal_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/gl_process_repository.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

part 'gl_process_event.dart';
part 'gl_process_state.dart';

class GlProcessBloc extends Bloc<GlProcessEvent, GlProcessState> {
  final GlProcessRepository _glProcessRepository;

  GlProcessBloc({required GlProcessRepository glProcessRepository})
      : _glProcessRepository = glProcessRepository,
        super(GlProcessInitial()) {
    on<GetPurchaseList>(onPurchaseList);
    on<GetSaleList>(onSaleList);
    on<GetPurchaseReturnList>(onPurchaseReturnList);
    on<GetSaleReturnList>(onSaleReturnList);
    on<GetStockReturnProductList>(onStockReturnProductList);
    on<GetStockAdjustList>(onStockAdjustList);
    on<GetStockPickupList>(onStockPickupList);
    on<GetStockReceiveList>(onStockReceiveList);
    on<GetPayList>(onPayList);
    on<GetPaidList>(onPaidList);
    on<SaveJournalBulk>(onSaveJournalBulk);
  }

  void onPurchaseList(
      GetPurchaseList event, Emitter<GlProcessState> emit) async {
    emit(TransPurchaseInProgress());

    try {
      final results = await _glProcessRepository.getTransactionPurchaseList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransPurchaseLoadSuccess(trans: trans));
      } else {
        emit(
            const TransPurchaseLoadFailed(message: 'Trans Purchase Not Found'));
      }
    } catch (e) {
      emit(TransPurchaseLoadFailed(message: e.toString()));
    }
  }

  void onSaleList(GetSaleList event, Emitter<GlProcessState> emit) async {
    emit(TransSaleInProgress());

    try {
      final results = await _glProcessRepository.getTransactionSaleList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransSaleLoadSuccess(trans: trans));
      } else {
        emit(const TransSaleLoadFailed(message: 'Trans Sale Not Found'));
      }
    } catch (e) {
      emit(TransSaleLoadFailed(message: e.toString()));
    }
  }

  void onPurchaseReturnList(
      GetPurchaseReturnList event, Emitter<GlProcessState> emit) async {
    emit(TransPurchaseReturnInProgress());

    try {
      final results =
          await _glProcessRepository.getTransactionPurchaseReturnList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransPurchaseReturnLoadSuccess(trans: trans));
      } else {
        emit(const TransPurchaseReturnLoadFailed(
            message: 'Trans Purchase Return Not Found'));
      }
    } catch (e) {
      emit(TransPurchaseReturnLoadFailed(message: e.toString()));
    }
  }

  void onSaleReturnList(
      GetSaleReturnList event, Emitter<GlProcessState> emit) async {
    emit(TransSaleReturnInProgress());

    try {
      final results = await _glProcessRepository.getTransactionSaleReturnList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransSaleReturnLoadSuccess(trans: trans));
      } else {
        emit(const TransSaleReturnLoadFailed(
            message: 'Trans Sale Return Not Found'));
      }
    } catch (e) {
      emit(TransSaleReturnLoadFailed(message: e.toString()));
    }
  }

  void onStockReturnProductList(
      GetStockReturnProductList event, Emitter<GlProcessState> emit) async {
    emit(TransStockReturnProductInProgress());

    try {
      final results =
          await _glProcessRepository.getTransactionStockReturnProductList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransStockReturnProductLoadSuccess(trans: trans));
      } else {
        emit(const TransStockReturnProductLoadFailed(
            message: 'Trans SaStock Return Product Not Found'));
      }
    } catch (e) {
      emit(TransStockReturnProductLoadFailed(message: e.toString()));
    }
  }

  void onStockAdjustList(
      GetStockAdjustList event, Emitter<GlProcessState> emit) async {
    emit(TransStockAdjustInProgress());

    try {
      final results = await _glProcessRepository.getTransactionStockAdjustList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransStockAdjustLoadSuccess(trans: trans));
      } else {
        emit(const TransStockAdjustLoadFailed(
            message: 'Trans Stock Adjust Not Found'));
      }
    } catch (e) {
      emit(TransStockAdjustLoadFailed(message: e.toString()));
    }
  }

  void onStockPickupList(
      GetStockPickupList event, Emitter<GlProcessState> emit) async {
    emit(TransStockPickupInProgress());

    try {
      final results = await _glProcessRepository.getTransactionStockPickupList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransStockPickupLoadSuccess(trans: trans));
      } else {
        emit(const TransStockPickupLoadFailed(
            message: 'Trans Stock Pickup Not Found'));
      }
    } catch (e) {
      emit(TransStockPickupLoadFailed(message: e.toString()));
    }
  }

  void onStockReceiveList(
      GetStockReceiveList event, Emitter<GlProcessState> emit) async {
    emit(TransStockReceiveInProgress());

    try {
      final results = await _glProcessRepository.getTransactionStockReceiveList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionModel> trans = (results.data as List)
            .map((trans) => TransactionModel.fromJson(trans))
            .toList();

        emit(TransStockReceiveLoadSuccess(trans: trans));
      } else {
        emit(const TransStockReceiveLoadFailed(
            message: 'Trans Stock Receive Not Found'));
      }
    } catch (e) {
      emit(TransStockReceiveLoadFailed(message: e.toString()));
    }
  }

  void onPayList(GetPayList event, Emitter<GlProcessState> emit) async {
    emit(TransStockReceiveInProgress());

    try {
      final results = await _glProcessRepository.getTransactionPayList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionPaidPayModel> trans = (results.data as List)
            .map((trans) => TransactionPaidPayModel.fromJson(trans))
            .toList();

        emit(TransPayLoadSuccess(trans: trans));
      } else {
        emit(const TransPayLoadFailed(message: 'Trans Pay Not Found'));
      }
    } catch (e) {
      emit(TransPayLoadFailed(message: e.toString()));
    }
  }

  void onPaidList(GetPaidList event, Emitter<GlProcessState> emit) async {
    emit(TransPaidInProgress());

    try {
      final results = await _glProcessRepository.getTransactionPaidList(
        fromDate: event.fromDate,
        toDate: event.toDate,
      );

      // // print(results.data);
      if (results.success) {
        List<TransactionPaidPayModel> trans = (results.data as List)
            .map((trans) => TransactionPaidPayModel.fromJson(trans))
            .toList();

        emit(TransPaidLoadSuccess(trans: trans));
      } else {
        emit(const TransPaidLoadFailed(message: 'Trans Paid Not Found'));
      }
    } catch (e) {
      emit(TransPaidLoadFailed(message: e.toString()));
    }
  }

  void onSaveJournalBulk(
      SaveJournalBulk event, Emitter<GlProcessState> emit) async {
    emit(SaveJournalBulkInProgress());

    try {
      await _glProcessRepository.saveJournalBulk(event.journalData);
      emit(SaveJournalBulkSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(SaveJournalBulkFailed(message: error['message']));
    }
  }
}
