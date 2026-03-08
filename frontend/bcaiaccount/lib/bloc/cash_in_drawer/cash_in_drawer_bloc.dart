import 'package:flutter/foundation.dart';
import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/cash_in_drawer/cash_in_drawer_event.dart';
import 'package:smlaicloud/bloc/cash_in_drawer/cash_in_drawer_state.dart';
import 'package:smlaicloud/model/cash_in_drawer_model.dart';
import 'package:smlaicloud/model/pagination.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/shift_detail_model.dart';
import 'package:smlaicloud/repositories/cash_in_drawer_repository.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CashInDrawerBloc extends Bloc<CashInDrawerEvent, CashInDrawerState> {
  final CashInDrawerRepository cashInDrawerRepository;
  CashInDrawerBloc({required this.cashInDrawerRepository})
    : super(const CashInDrawerInitial()) {
    on<LoadCashInDrawerList>(_onLoadCashInDrawerList);
    on<LoadCashInDrawerListWithFilter>(_onLoadCashInDrawerListWithFilter);
    on<RefreshCashInDrawerList>(_onRefreshCashInDrawerList);
    on<LoadShiftReportDetails>(_onLoadShiftReportDetails);
  }

  Future<void> _onLoadCashInDrawerList(
    LoadCashInDrawerList event,
    Emitter<CashInDrawerState> emit,
  ) async {
    emit(const CashInDrawerInProgress());
    try {
      final response = await cashInDrawerRepository.getShiftList(
        page: event.page,
        limit: event.limit,
        doctype: event.doctype,
        fromdate: event.fromdate,
        todate: event.todate,
        usercode: event.usercode,
        posid: event.posid,
      );

      if (response.success) {
        final allData = (response.data as List)
            .map((data) => CashInDrawerModel.fromJson(data))
            .toList();

        final page = response.page ?? Page.empty;

        final pagination = Pagination(
          page: page.page,
          perPage: page.perPage,
          total: page.total,
          totalPage: page.totalPage,
          next: page.page < page.totalPage ? page.page + 1 : 0,
          prev: page.page > 1 ? page.page - 1 : 0,
        );

        emit(
          CashInDrawerLoadListSuccess(
            data: allData,
            pagination: pagination,
            currentLimit: event.limit,
          ),
        );
      } else {
        emit(CashInDrawerLoadFailed(response.message));
      }
    } catch (e) {
      emit(CashInDrawerLoadFailed(e.toString()));
    }
  }

  Future<void> _onRefreshCashInDrawerList(
    RefreshCashInDrawerList event,
    Emitter<CashInDrawerState> emit,
  ) async {
    add(
      LoadCashInDrawerList(
        page: event.page,
        limit: event.limit,
        doctype: const [1, 2, 3, 4],
        fromdate: '',
        todate: '',
        usercode: '',
        posid: '',
      ),
    );
  }

  Future<void> _onLoadCashInDrawerListWithFilter(
    LoadCashInDrawerListWithFilter event,
    Emitter<CashInDrawerState> emit,
  ) async {
    emit(const CashInDrawerInProgress());
    try {
      // Set doctype based on filter type
      List<int> filteredDoctype;
      if (event.filterType == 'cash_in') {
        filteredDoctype = [1, 3]; // เปิดกะ, เพิ่มเงิน
      } else if (event.filterType == 'cash_out') {
        filteredDoctype = [2, 4]; // ปิดกะ, ถอนเงิน
      } else {
        filteredDoctype = event.doctype;
      }

      final response = await cashInDrawerRepository.getShiftList(
        page: event.page,
        limit: event.limit,
        doctype: filteredDoctype,
        fromdate: event.fromdate,
        todate: event.todate,
        usercode: event.usercode,
        posid: event.posid,
      );

      if (response.success) {
        final allData = (response.data as List)
            .map((data) => CashInDrawerModel.fromJson(data))
            .toList();

        final page = response.page ?? Page.empty;

        final pagination = Pagination(
          page: page.page,
          perPage: page.perPage,
          total: page.total,
          totalPage: page.totalPage,
          next: page.page < page.totalPage ? page.page + 1 : 0,
          prev: page.page > 1 ? page.page - 1 : 0,
        );

        emit(
          CashInDrawerLoadListSuccess(
            data: allData,
            pagination: pagination,
            currentLimit: event.limit,
          ),
        );
      } else {
        emit(CashInDrawerLoadFailed(response.message));
      }
    } catch (e) {
      emit(CashInDrawerLoadFailed(e.toString()));
    }
  }

  Future<void> _onLoadShiftReportDetails(
    LoadShiftReportDetails event,
    Emitter<CashInDrawerState> emit,
  ) async {
    emit(CashInDrawerShiftReportDetailsInProgress(event.docno));
    try {
      final response = await cashInDrawerRepository.getShiftReportDetails(
        docno: event.docno,
      );

      if (response.success) {
        // ดึงข้อมูลรายการบิลจากการตอบกลับ
        List<TransactionModel> billDetails = [];
        List<ShiftDetailModel> shifts = [];

        if (response.data["saleinvoices"] != null) {
          final saleInvoices = response.data["saleinvoices"];
          billDetails = (saleInvoices as List)
              .map((data) => TransactionModel.fromJson(data))
              .toList();

          for (var element in billDetails) {
            if (element.paymentdetailraw != "") {
              try {
                final List<dynamic> jsonStr = jsonDecode(
                  element.paymentdetailraw!,
                );
                element.paymentStructs = (jsonStr)
                    .map((e) => BillPayStruct.fromJson(e))
                    .toList();
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('Error paymentdetailraw JSON: $e');
                }
              }
            }
          }
        }

        // ดึงข้อมูล shifts
        if (response.data["shifts"] != null) {
          final shiftsData = response.data["shifts"];
          shifts = (shiftsData as List)
              .map((data) => ShiftDetailModel.fromJson(data))
              .toList();
        }

        emit(
          CashInDrawerShiftReportDetailsSuccess(
            docno: event.docno,
            billDetails: billDetails,
            shifts: shifts,
          ),
        );
      } else {
        emit(
          CashInDrawerShiftReportDetailsFailed(
            docno: event.docno,
            message: response.message,
          ),
        );
      }
    } catch (e) {
      emit(
        CashInDrawerShiftReportDetailsFailed(
          docno: event.docno,
          message: e.toString(),
        ),
      );
    }
  }
}
