import 'package:equatable/equatable.dart';

abstract class CashInDrawerEvent extends Equatable {
  const CashInDrawerEvent();

  @override
  List<Object?> get props => [];
}

class LoadCashInDrawerList extends CashInDrawerEvent {
  final int page;
  final int limit;
  final List<int> doctype;
  final String fromdate;
  final String todate;
  final String usercode;
  final String posid;

  const LoadCashInDrawerList(
      {this.page = 1,
      this.limit = 10,
      this.doctype = const [],
      this.fromdate = '',
      this.todate = '',
      this.usercode = '',
      this.posid = ''});

  @override
  List<Object?> get props =>
      [page, limit, doctype, fromdate, todate, usercode, posid];
}

class LoadShiftReportDetails extends CashInDrawerEvent {
  final String docno;

  const LoadShiftReportDetails({required this.docno});

  @override
  List<Object?> get props => [docno];
}

class RefreshCashInDrawerList extends CashInDrawerEvent {
  final int page;
  final int limit;

  const RefreshCashInDrawerList({this.page = 1, this.limit = 10});

  @override
  List<Object?> get props => [page, limit];
}

class LoadCashInDrawerListWithFilter extends CashInDrawerEvent {
  final int page;
  final int limit;
  final List<int> doctype;
  final String fromdate;
  final String todate;
  final String usercode;
  final String posid;
  final String filterType; // 'cash_in' or 'cash_out'

  const LoadCashInDrawerListWithFilter({
    this.page = 1,
    this.limit = 10,
    this.doctype = const [],
    this.fromdate = '',
    this.todate = '',
    this.usercode = '',
    this.posid = '',
    this.filterType = 'cash_in',
  });

  @override
  List<Object?> get props =>
      [page, limit, doctype, fromdate, todate, usercode, posid, filterType];
}
