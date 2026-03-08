part of 'gl_process_bloc.dart';

abstract class GlProcessState extends Equatable {
  const GlProcessState();

  @override
  List<Object> get props => [];
}

class GlProcessInitial extends GlProcessState {}

class TransPurchaseInProgress extends GlProcessState {}

class TransPurchaseLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransPurchaseLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransPurchaseLoadFailed extends GlProcessState {
  final String message;

  const TransPurchaseLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransSaleInProgress extends GlProcessState {}

class TransSaleLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransSaleLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransSaleLoadFailed extends GlProcessState {
  final String message;

  const TransSaleLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class SaveJournalBulkInitial extends GlProcessState {}

class SaveJournalBulkInProgress extends GlProcessState {}

class SaveJournalBulkSuccess extends GlProcessState {}

class SaveJournalBulkFailed extends GlProcessState {
  final String message;

  const SaveJournalBulkFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

////
///

class TransPurchaseReturnInProgress extends GlProcessState {}

class TransPurchaseReturnLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransPurchaseReturnLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransPurchaseReturnLoadFailed extends GlProcessState {
  final String message;

  const TransPurchaseReturnLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransSaleReturnInProgress extends GlProcessState {}

class TransSaleReturnLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransSaleReturnLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransSaleReturnLoadFailed extends GlProcessState {
  final String message;

  const TransSaleReturnLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransStockReturnProductInProgress extends GlProcessState {}

class TransStockReturnProductLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransStockReturnProductLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransStockReturnProductLoadFailed extends GlProcessState {
  final String message;

  const TransStockReturnProductLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransStockAdjustInProgress extends GlProcessState {}

class TransStockAdjustLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransStockAdjustLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransStockAdjustLoadFailed extends GlProcessState {
  final String message;

  const TransStockAdjustLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransStockPickupInProgress extends GlProcessState {}

class TransStockPickupLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransStockPickupLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransStockPickupLoadFailed extends GlProcessState {
  final String message;

  const TransStockPickupLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransStockReceiveInProgress extends GlProcessState {}

class TransStockReceiveLoadSuccess extends GlProcessState {
  final List<TransactionModel> trans;

  const TransStockReceiveLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransStockReceiveLoadFailed extends GlProcessState {
  final String message;

  const TransStockReceiveLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransPayInProgress extends GlProcessState {}

class TransPayLoadSuccess extends GlProcessState {
  final List<TransactionPaidPayModel> trans;

  const TransPayLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransPayLoadFailed extends GlProcessState {
  final String message;

  const TransPayLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransPaidInProgress extends GlProcessState {}

class TransPaidLoadSuccess extends GlProcessState {
  final List<TransactionPaidPayModel> trans;

  const TransPaidLoadSuccess({required this.trans});

  @override
  List<Object> get props => [trans];
}

class TransPaidLoadFailed extends GlProcessState {
  final String message;

  const TransPaidLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
