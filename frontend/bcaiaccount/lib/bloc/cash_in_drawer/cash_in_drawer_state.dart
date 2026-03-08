import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/cash_in_drawer_model.dart';
import 'package:smlaicloud/model/pagination.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/shift_detail_model.dart';

abstract class CashInDrawerState extends Equatable {
  const CashInDrawerState();

  @override
  List<Object?> get props => [];
}

class CashInDrawerInitial extends CashInDrawerState {
  const CashInDrawerInitial();
}

class CashInDrawerInProgress extends CashInDrawerState {
  const CashInDrawerInProgress();
}

class CashInDrawerLoadListSuccess extends CashInDrawerState {
  final List<CashInDrawerModel> data;
  final Pagination pagination;
  final int currentLimit;

  const CashInDrawerLoadListSuccess({
    required this.data,
    required this.pagination,
    this.currentLimit = 10,
  });

  @override
  List<Object> get props => [data, pagination, currentLimit];
}

class CashInDrawerLoadFailed extends CashInDrawerState {
  final String message;

  const CashInDrawerLoadFailed(this.message);

  @override
  List<Object> get props => [message];
}

class CashInDrawerShiftReportDetailsInProgress extends CashInDrawerState {
  final String docno;

  const CashInDrawerShiftReportDetailsInProgress(this.docno);

  @override
  List<Object> get props => [docno];
}

class CashInDrawerShiftReportDetailsSuccess extends CashInDrawerState {
  final String docno;
  final List<TransactionModel> billDetails;
  final List<ShiftDetailModel> shifts;

  const CashInDrawerShiftReportDetailsSuccess({
    required this.docno,
    required this.billDetails,
    required this.shifts,
  });

  @override
  List<Object> get props => [docno, billDetails, shifts];
}

class CashInDrawerShiftReportDetailsFailed extends CashInDrawerState {
  final String docno;
  final String message;

  const CashInDrawerShiftReportDetailsFailed({
    required this.docno,
    required this.message,
  });

  @override
  List<Object> get props => [docno, message];
}
