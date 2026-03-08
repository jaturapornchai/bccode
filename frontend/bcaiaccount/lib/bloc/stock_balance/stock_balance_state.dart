part of 'stock_balance_bloc.dart';

abstract class StockBalanceState extends Equatable {
  const StockBalanceState();

  @override
  List<Object> get props => [];
}

class UploadFileExcelInProgress extends StockBalanceState {}

class StockBalanceInitial extends StockBalanceState {}

class UploadFileExcelSuccess extends StockBalanceState {
  final UploadSuccessModel response;

  const UploadFileExcelSuccess({
    required this.response,
  });

  @override
  List<Object> get props => [response];
}

class UploadFileExcelFailed extends StockBalanceState {
  final String message;

  const UploadFileExcelFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LoadStockBalanceImportByTaskidSuccess extends StockBalanceState {
  final List<StockBalanceImportModel> data;
  final Pagination pagination;

  const LoadStockBalanceImportByTaskidSuccess({
    required this.data,
    required this.pagination,
  });

  @override
  List<Object> get props => [data, pagination];
}

class LoadStockBalanceImportByTaskidFailed extends StockBalanceState {
  final String message;

  const LoadStockBalanceImportByTaskidFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DeleteDetailByGuidInProgress extends StockBalanceState {}

class DeleteDetailByGuidSuccess extends StockBalanceState {}

class DeleteDetailByGuidFailed extends StockBalanceState {
  final String message;

  const DeleteDetailByGuidFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DeleteTaskidInProgress extends StockBalanceState {}

class DeleteTaskidSuccess extends StockBalanceState {}

class DeleteTaskidFailed extends StockBalanceState {
  final String message;

  const DeleteTaskidFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UpdateDetailInProgress extends StockBalanceState {}

class UpdateDetailSuccess extends StockBalanceState {}

class UpdateDetailFailed extends StockBalanceState {
  final String message;

  const UpdateDetailFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class AddDetailInProgress extends StockBalanceState {}

class AddDetailSuccess extends StockBalanceState {}

class AddDetailFailed extends StockBalanceState {
  final String message;

  const AddDetailFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LoadTotalInProgress extends StockBalanceState {}

class LoadTotalSuccess extends StockBalanceState {
  final TotalModel total;

  const LoadTotalSuccess({
    required this.total,
  });

  @override
  List<Object> get props => [total];
}

class LoadTotalFailed extends StockBalanceState {
  final String message;

  const LoadTotalFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class SaveTransStockBalanceInProgress extends StockBalanceState {}

class SaveTransStockBalanceSuccess extends StockBalanceState {}

class SaveTransStockBalanceFailed extends StockBalanceState {
  final String message;

  const SaveTransStockBalanceFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LoadTransStockBalanceDetailByDocnoInProgress extends StockBalanceState {}

class LoadTransStockBalanceDetailByDocnoSuccess extends StockBalanceState {
  final List<TransactionDetailModel> data;
  final Pagination pagination;

  const LoadTransStockBalanceDetailByDocnoSuccess({
    required this.data,
    required this.pagination,
  });

  @override
  List<Object> get props => [data, pagination];
}

class LoadTransStockBalanceDetailByDocnoFailed extends StockBalanceState {
  final String message;

  const LoadTransStockBalanceDetailByDocnoFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
