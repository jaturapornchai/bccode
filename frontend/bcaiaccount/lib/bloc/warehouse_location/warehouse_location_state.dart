part of 'warehouse_location_bloc.dart';

abstract class WarehouseLocationState extends Equatable {
  const WarehouseLocationState();

  @override
  List<Object> get props => [];
}

class WarehouseLocationInitial extends WarehouseLocationState {}

class WarehouseLocationInProgress extends WarehouseLocationState {}

class WarehouseLocationLoadSuccess extends WarehouseLocationState {
  final List<WarehouseLocationModel> warehouses;
  const WarehouseLocationLoadSuccess({required this.warehouses});

  WarehouseLocationLoadSuccess copyWith({
    List<WarehouseLocationModel>? warehouses,
  }) =>
      WarehouseLocationLoadSuccess(warehouses: warehouses ?? this.warehouses);

  @override
  List<Object> get props => [warehouses];
}

class WarehouseLocationLoadFailed extends WarehouseLocationState {
  final String message;

  const WarehouseLocationLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WarehouseLocationGetInProgress extends WarehouseLocationState {}

class WarehouseLocationGetSuccess extends WarehouseLocationState {
  final WarehouseLocationModel warehouselocation;

  const WarehouseLocationGetSuccess({required this.warehouselocation});

  WarehouseLocationGetSuccess copyWith({
    WarehouseLocationModel? warehouselocation,
  }) =>
      WarehouseLocationGetSuccess(
          warehouselocation: warehouselocation ?? this.warehouselocation);

  @override
  List<Object> get props => [warehouselocation];
}

class WarehouseLocationGetFailed extends WarehouseLocationState {
  final String message;

  const WarehouseLocationGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WarehouseLocationUpdateInitial extends WarehouseLocationState {}

class WarehouseLocationUpdateInProgress extends WarehouseLocationState {}

class WarehouseLocationUpdateSuccess extends WarehouseLocationState {}

class WarehouseLocationUpdateFailed extends WarehouseLocationState {
  final String message;

  const WarehouseLocationUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WarehouseLocationDeleteManyInProgress extends WarehouseLocationState {}

class WarehouseLocationDeleteManySuccess extends WarehouseLocationState {}

class WarehouseLocationDeleteManyFailed extends WarehouseLocationState {
  final String message;

  const WarehouseLocationDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
