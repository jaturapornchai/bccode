part of 'table_bloc.dart';

abstract class TableState extends Equatable {
  const TableState();

  @override
  List<Object> get props => [];
}

class TableInitial extends TableState {}

class TableInProgress extends TableState {}

class TableLoadSuccess extends TableState {
  final List<TableModel> tables;

  const TableLoadSuccess({required this.tables});

  TableLoadSuccess copyWith({
    List<TableModel>? tables,
  }) =>
      TableLoadSuccess(tables: tables ?? this.tables);

  @override
  List<Object> get props => [tables];
}

class TableLoadFailed extends TableState {
  final String message;

  const TableLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TableSaveInitial extends TableState {}

class TableSaveInProgress extends TableState {}

class TableSaveSuccess extends TableState {}

class TableSaveFailed extends TableState {
  final String message;

  const TableSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TableDeleteInProgress extends TableState {}

class TableDeleteSuccess extends TableState {}

class TableDeleteFailed extends TableState {}

class TableDeleteManyInProgress extends TableState {}

class TableDeleteManySuccess extends TableState {}

class TableDeleteManyFailed extends TableState {}

class TableGetInProgress extends TableState {}

class TableGetSuccess extends TableState {
  final TableModel table;

  const TableGetSuccess({required this.table});

  TableGetSuccess copyWith({
    TableModel? table,
  }) =>
      TableGetSuccess(table: table ?? this.table);

  @override
  List<Object> get props => [table];
}

class TableGetFailed extends TableState {
  final String message;

  const TableGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TableUpdateInitial extends TableState {}

class TableUpdateInProgress extends TableState {}

class TableUpdateSuccess extends TableState {}

class TableUpdateFailed extends TableState {
  final String message;

  const TableUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
