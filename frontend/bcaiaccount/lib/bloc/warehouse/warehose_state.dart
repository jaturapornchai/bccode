part of 'warehose_bloc.dart';

abstract class WarehouseState extends Equatable {
  const WarehouseState();

  @override
  List<Object> get props => [];
}

class WarehouseInitial extends WarehouseState {}

class WarehouseInProgress extends WarehouseState {}

class WarehouseLoadSuccess extends WarehouseState {
  final List<WarehouseModel> warehouses;

  const WarehouseLoadSuccess({required this.warehouses});

  WarehouseLoadSuccess copyWith({
    List<WarehouseModel>? warehouses,
  }) =>
      WarehouseLoadSuccess(warehouses: warehouses ?? this.warehouses);

  @override
  List<Object> get props => [warehouses];
}

class WarehouseLoadFailed extends WarehouseState {
  final String message;

  const WarehouseLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WarehouseSaveInitial extends WarehouseState {}

class WarehouseSaveInProgress extends WarehouseState {}

class WarehouseSaveSuccess extends WarehouseState {}

class WarehouseSaveFailed extends WarehouseState {
  final String message;

  const WarehouseSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WarehouseDeleteInProgress extends WarehouseState {}

class WarehouseDeleteSuccess extends WarehouseState {}

class WarehouseDeleteFailed extends WarehouseState {}

class WarehouseDeleteManyInProgress extends WarehouseState {}

class WarehouseDeleteManySuccess extends WarehouseState {}

class WarehouseDeleteManyFailed extends WarehouseState {}

class WarehouseGetInProgress extends WarehouseState {}

class WarehouseGetSuccess extends WarehouseState {
  final WarehouseModel warehouse;

  const WarehouseGetSuccess({required this.warehouse});

  WarehouseGetSuccess copyWith({
    WarehouseModel? warehouse,
  }) =>
      WarehouseGetSuccess(warehouse: warehouse ?? this.warehouse);

  @override
  List<Object> get props => [warehouse];
}

class WarehouseGetFailed extends WarehouseState {
  final String message;

  const WarehouseGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WarehouseUpdateInitial extends WarehouseState {}

class WarehouseUpdateInProgress extends WarehouseState {}

class WarehouseUpdateSuccess extends WarehouseState {}

class WarehouseUpdateFailed extends WarehouseState {
  final String message;

  const WarehouseUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
