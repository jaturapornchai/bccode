part of 'gl_process_bloc.dart';

abstract class GlProcessEvent extends Equatable {
  const GlProcessEvent();

  @override
  List<Object> get props => [];
}

/// ซื้อ
class GetPurchaseList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetPurchaseList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

/// ขาย
class GetSaleList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetSaleList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetPurchaseReturnList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetPurchaseReturnList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetSaleReturnList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetSaleReturnList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetStockReturnProductList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetStockReturnProductList(
      {required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetStockAdjustList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetStockAdjustList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetStockPickupList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetStockPickupList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetStockReceiveList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetStockReceiveList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetPayList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetPayList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

class GetPaidList extends GlProcessEvent {
  final String fromDate;
  final String toDate;

  const GetPaidList({required this.fromDate, required this.toDate});

  @override
  List<Object> get props => [fromDate, toDate];
}

/// save journal bulk
class SaveJournalBulk extends GlProcessEvent {
  final List<JournalModel> journalData;

  const SaveJournalBulk({
    required this.journalData,
  });

  @override
  List<Object> get props => [journalData];
}
