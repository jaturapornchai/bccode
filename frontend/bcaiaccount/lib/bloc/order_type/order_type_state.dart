part of 'order_type_bloc.dart';

abstract class OrderTypeState extends Equatable {
  const OrderTypeState();

  @override
  List<Object> get props => [];
}

class OrderTypeInitial extends OrderTypeState {}

class OrderTypeInProgress extends OrderTypeState {}

class OrderTypeLoadSuccess extends OrderTypeState {
  final List<OrderTypeModel> ordertype;

  const OrderTypeLoadSuccess({required this.ordertype});

  OrderTypeLoadSuccess copyWith({
    List<OrderTypeModel>? ordertype,
  }) =>
      OrderTypeLoadSuccess(ordertype: ordertype ?? this.ordertype);

  @override
  List<Object> get props => [ordertype];
}

class OrderTypeLoadFailed extends OrderTypeState {
  final String message;

  const OrderTypeLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTypeSaveInitial extends OrderTypeState {}

class OrderTypeSaveInProgress extends OrderTypeState {}

class OrderTypeSaveSuccess extends OrderTypeState {}

class OrderTypeSaveFailed extends OrderTypeState {
  final String message;

  const OrderTypeSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTypeDeleteInProgress extends OrderTypeState {}

class OrderTypeDeleteSuccess extends OrderTypeState {}

class OrderTypeDeleteFailed extends OrderTypeState {
  final String message;

  const OrderTypeDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTypeDeleteManyInProgress extends OrderTypeState {}

class OrderTypeDeleteManySuccess extends OrderTypeState {}

class OrderTypeDeleteManyFailed extends OrderTypeState {
  final String message;

  const OrderTypeDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTypeGetInProgress extends OrderTypeState {}

class OrderTypeGetSuccess extends OrderTypeState {
  final OrderTypeModel ordertype;

  const OrderTypeGetSuccess({required this.ordertype});

  OrderTypeGetSuccess copyWith({
    OrderTypeModel? ordertype,
  }) =>
      OrderTypeGetSuccess(ordertype: ordertype ?? this.ordertype);

  @override
  List<Object> get props => [ordertype];
}

class OrderTypeGetFailed extends OrderTypeState {
  final String message;

  const OrderTypeGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTypeUpdateInitial extends OrderTypeState {}

class OrderTypeUpdateInProgress extends OrderTypeState {}

class OrderTypeUpdateSuccess extends OrderTypeState {}

class OrderTypeUpdateFailed extends OrderTypeState {
  final String message;

  const OrderTypeUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
