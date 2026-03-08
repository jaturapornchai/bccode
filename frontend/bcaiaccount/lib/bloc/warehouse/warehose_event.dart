part of 'warehose_bloc.dart';

abstract class WarehouseEvent extends Equatable {
  const WarehouseEvent();

  @override
  List<Object> get props => [];
}

class WarehouseGet extends WarehouseEvent {
  final String guid;

  const WarehouseGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class WarehouseLoadList extends WarehouseEvent {
  final int limit;
  final int offset;
  final String search;

  const WarehouseLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

@override
List<Object> get props => [];

class WarehouseGetByCode extends WarehouseEvent {
  final String code;

  const WarehouseGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class WarehouseDelete extends WarehouseEvent {
  final String guid;

  const WarehouseDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class WarehouseDeleteMany extends WarehouseEvent {
  final List<String> guid;

  const WarehouseDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class WarehouseSave extends WarehouseEvent {
  final WarehouseModel warehouseModel;

  const WarehouseSave({
    required this.warehouseModel,
  });

  @override
  List<Object> get props => [warehouseModel];
}

class WarehouseUpdate extends WarehouseEvent {
  final String guid;
  final WarehouseModel warehouseModel;

  const WarehouseUpdate({
    required this.guid,
    required this.warehouseModel,
  });

  @override
  List<Object> get props => [warehouseModel];
}
