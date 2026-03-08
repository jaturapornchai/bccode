part of 'table_bloc.dart';

abstract class TableEvent extends Equatable {
  const TableEvent();

  @override
  List<Object> get props => [];
}

class TableGet extends TableEvent {
  final String guid;

  const TableGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class TableLoadList extends TableEvent {
  final int limit;
  final int offset;
  final String search;
  final int groupNumber;

  const TableLoadList(
      {required this.offset,
      required this.limit,
      required this.search,
      required this.groupNumber});

  @override
  List<Object> get props => [];
}

class TableDelete extends TableEvent {
  final String guid;

  const TableDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class TableDeleteMany extends TableEvent {
  final List<String> guid;

  const TableDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class TableSave extends TableEvent {
  final TableModel tableModel;

  const TableSave({
    required this.tableModel,
  });

  @override
  List<Object> get props => [tableModel];
}

class TableUpdate extends TableEvent {
  final String guid;
  final TableModel tableModel;

  const TableUpdate({
    required this.guid,
    required this.tableModel,
  });

  @override
  List<Object> get props => [tableModel];
}

class TableUpdateXorder extends TableEvent {
  final List<TableXorderModel> tableModel;

  const TableUpdateXorder({
    required this.tableModel,
  });

  @override
  List<Object> get props => [tableModel];
}
