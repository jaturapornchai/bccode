import 'package:flutter/foundation.dart';
import 'dart:convert';

import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/transaction_paidpay_repository.dart';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'transaction_paidpay_event.dart';
part 'transaction_paidpay_state.dart';

class TransactionPaidPayBloc
    extends Bloc<TransactionPaidPayEvent, TransactionPaidPayState> {
  final TransactionPaidPayRepository _transactionPaidPayRepository;

  TransactionPaidPayBloc({
    required TransactionPaidPayRepository transactionPaidPayRepository,
  }) : _transactionPaidPayRepository = transactionPaidPayRepository,
       super(TransactionPaidPayInitial()) {
    on<TransactionPaidPayLoad>(onTransactionPaidPayLoad);
    on<TransactionPaidPaySave>(onTransactionPaidPaySave);
    on<TransactionPaidPayUpdate>(onTransactionPaidPayUpdate);
    on<TransactionPaidPayDelete>(onTransactionPaidPayDelete);
    on<GetCustcodeTransaction>(onGetCustcodeTransaction);
  }

  void onTransactionPaidPayLoad(
    TransactionPaidPayLoad event,
    Emitter<TransactionPaidPayState> emit,
  ) async {
    emit(TransactionPaidPayInitial());

    try {
      late ApiResponse<dynamic> results;
      if (event.type == TransactionTypeEnum.paid) {
        results = await _transactionPaidPayRepository.getPaid(
          limit: event.limit,
          offset: event.offset,
          search: event.search,
        );
      } else if (event.type == TransactionTypeEnum.pay) {
        results = await _transactionPaidPayRepository.getPay(
          limit: event.limit,
          offset: event.offset,
          search: event.search,
        );
      } else {
        emit(
          const TransactionPaidPayLoadFailed(
            message: 'Transaction Paid/Pay Not Found',
          ),
        );
      }

      if (results.success) {
        List<TransactionPaidPayModel> transactionPaidPay =
            (results.data as List)
                .map(
                  (transactionPaidPay) =>
                      TransactionPaidPayModel.fromJson(transactionPaidPay),
                )
                .toList();
        for (var element in transactionPaidPay) {
          if (element.paymentdetailraw != "") {
            try {
              final List<dynamic> jsonStr = jsonDecode(
                element.paymentdetailraw,
              );
              element.paymentStructs = (jsonStr)
                  .map((e) => BillPayStruct.fromJson(e))
                  .toList();
            } catch (e) {
              if (kDebugMode) {
                AppLogger.error('Error parsing JSON: $e');
              }
            }
          }
        }
        emit(
          TransactionPaidPayLoadSuccess(transactionPaidPay: transactionPaidPay),
        );
      } else {
        emit(
          const TransactionPaidPayLoadFailed(
            message: 'Transaction Paid/Pay  Not Found',
          ),
        );
      }
    } catch (e) {
      emit(TransactionPaidPayLoadFailed(message: e.toString()));
    }
  }

  void onTransactionPaidPaySave(
    TransactionPaidPaySave event,
    Emitter<TransactionPaidPayState> emit,
  ) async {
    late ApiResponse<dynamic> results;
    emit(TransactionPaidPaySaveInProgress());
    try {
      if (event.type == TransactionTypeEnum.paid) {
        results = await _transactionPaidPayRepository.savePaid(
          event.transactionPaidPay,
        );
      } else if (event.type == TransactionTypeEnum.pay) {
        results = await _transactionPaidPayRepository.savePay(
          event.transactionPaidPay,
        );
      } else {
        emit(const TransactionPaidPaySaveFailed(message: "Failed to save"));
      }
      emit(TransactionPaidPaySaveSuccess(docno: results.data));
    } catch (e) {
      emit(TransactionPaidPaySaveFailed(message: e.toString()));
    }
  }

  void onTransactionPaidPayUpdate(
    TransactionPaidPayUpdate event,
    Emitter<TransactionPaidPayState> emit,
  ) async {
    emit(TransactionPaidPayUpdateInProgress());
    try {
      if (event.type == TransactionTypeEnum.paid) {
        await _transactionPaidPayRepository.updateTransPaid(
          event.guid,
          event.transactionPaidPay,
        );
        emit(TransactionPaidPayUpdateSuccess());
      } else if (event.type == TransactionTypeEnum.pay) {
        await _transactionPaidPayRepository.updateTransPay(
          event.guid,
          event.transactionPaidPay,
        );
        emit(TransactionPaidPayUpdateSuccess());
      } else {
        emit(const TransactionPaidPayUpdateFailed(message: "Failed to update"));
      }
    } catch (e) {
      emit(TransactionPaidPayUpdateFailed(message: e.toString()));
    }
  }

  void onTransactionPaidPayDelete(
    TransactionPaidPayDelete event,
    Emitter<TransactionPaidPayState> emit,
  ) async {
    emit(TransactionPaidPayDeleteInProgress());
    try {
      if (event.type == TransactionTypeEnum.paid) {
        await _transactionPaidPayRepository.deletePaid(event.guid);
        emit(TransactionPaidPayDeleteSuccess());
      } else if (event.type == TransactionTypeEnum.pay) {
        await _transactionPaidPayRepository.deletePay(event.guid);
        emit(TransactionPaidPayDeleteSuccess());
      } else {
        emit(const TransactionPaidPayDeleteFailed(message: "No type"));
      }
    } catch (e) {
      emit(TransactionPaidPayDeleteFailed(message: e.toString()));
    }
  }

  void onGetCustcodeTransaction(
    GetCustcodeTransaction event,
    Emitter<TransactionPaidPayState> emit,
  ) async {
    emit(GetCustcodeTransactionInProgress());
    try {
      late ApiResponse<dynamic> results;
      results = await _transactionPaidPayRepository.getCustcodeTransaction(
        event.type,
        event.custcode,
      );
      if (results.success) {
        List<GetCustcodeTransationModel> getCustcodeTransationModel =
            (results.data as List)
                .map(
                  (getCustcodeTransationModel) =>
                      GetCustcodeTransationModel.fromJson(
                        getCustcodeTransationModel,
                      ),
                )
                .toList();

        emit(
          GetCustcodeTransactionSuccess(
            getCustcodeTransationModel: getCustcodeTransationModel,
          ),
        );
      } else {
        emit(
          const GetCustcodeTransactionFailed(
            message: 'Transaction Paid/Pay  Not Found',
          ),
        );
      }
    } catch (e) {
      emit(GetCustcodeTransactionFailed(message: e.toString()));
    }
  }
}
