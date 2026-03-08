part of 'stock_balance_bloc.dart';

abstract class StockBalanceEvent extends Equatable {
  const StockBalanceEvent();

  @override
  List<Object> get props => [];
}

class UploadFileExcel extends StockBalanceEvent {
  final String filename;
  final Uint8List file;

  const UploadFileExcel({
    required this.file,
    required this.filename,
  });

  @override
  List<Object> get props => [file, filename];
}

class LoadStockBalanceImportByTaskid extends StockBalanceEvent {
  final String taskid;
  final int limit;
  final int page;
  final String q;

  const LoadStockBalanceImportByTaskid({
    required this.taskid,
    required this.limit,
    required this.page,
    required this.q,
  });

  @override
  List<Object> get props => [taskid];
}

class DeleteTaskid extends StockBalanceEvent {
  final String taskid;

  const DeleteTaskid({
    required this.taskid,
  });

  @override
  List<Object> get props => [taskid];
}

class UpdateDetail extends StockBalanceEvent {
  final String guid;
  final StockBalanceImportModel stockBalanceImportModel;

  const UpdateDetail({
    required this.guid,
    required this.stockBalanceImportModel,
  });

  @override
  List<Object> get props => [guid, stockBalanceImportModel];
}

class AddDetail extends StockBalanceEvent {
  final StockBalanceImportModel stockBalanceImportModel;

  const AddDetail({
    required this.stockBalanceImportModel,
  });

  @override
  List<Object> get props => [stockBalanceImportModel];
}

class DeleteDetailByGuid extends StockBalanceEvent {
  final String guid;

  const DeleteDetailByGuid({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class LoadTotal extends StockBalanceEvent {
  final String taskid;

  const LoadTotal({
    required this.taskid,
  });

  @override
  List<Object> get props => [taskid];
}

class SaveTransStockBalance extends StockBalanceEvent {
  final String taskid;
  final TransactionModel transactionModel;

  const SaveTransStockBalance({
    required this.taskid,
    required this.transactionModel,
  });

  @override
  List<Object> get props => [taskid, transactionModel];
}

class LoadTransStockBalanceDetailByDocno extends StockBalanceEvent {
  final String docno;
  final int limit;
  final int page;
  final String q;

  const LoadTransStockBalanceDetailByDocno({
    required this.docno,
    required this.limit,
    required this.page,
    required this.q,
  });

  @override
  List<Object> get props => [docno, limit, page, q];
}
