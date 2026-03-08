part of 'order_setting_bloc.dart';

abstract class OrderSettingState extends Equatable {
  const OrderSettingState();

  @override
  List<Object> get props => [];
}

class OrderSettingInitial extends OrderSettingState {}

class OrderSettingInProgress extends OrderSettingState {}

class OrderSettingLoadSuccess extends OrderSettingState {
  final List<OrderSettingModel> orderSettings;

  const OrderSettingLoadSuccess({required this.orderSettings});

  OrderSettingLoadSuccess copyWith({
    List<OrderSettingModel>? orderSettings,
  }) =>
      OrderSettingLoadSuccess(
          orderSettings: orderSettings ?? this.orderSettings);

  @override
  List<Object> get props => [orderSettings];
}

class OrderSettingLoadFailed extends OrderSettingState {
  final String message;

  const OrderSettingLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderSettingSaveInitial extends OrderSettingState {}

class OrderSettingSaveInProgress extends OrderSettingState {}

class OrderSettingSaveSuccess extends OrderSettingState {}

class OrderSettingSaveFailed extends OrderSettingState {
  final String message;

  const OrderSettingSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderSettingDeleteInProgress extends OrderSettingState {}

class OrderSettingDeleteSuccess extends OrderSettingState {}

class OrderSettingDeleteFailed extends OrderSettingState {}

class OrderSettingDeleteManyInProgress extends OrderSettingState {}

class OrderSettingDeleteManySuccess extends OrderSettingState {}

class OrderSettingDeleteManyFailed extends OrderSettingState {}

class OrderSettingGetInProgress extends OrderSettingState {}

class OrderSettingGetSuccess extends OrderSettingState {
  final OrderSettingModel orderSettings;

  const OrderSettingGetSuccess({required this.orderSettings});

  OrderSettingGetSuccess copyWith({
    OrderSettingModel? orderSettings,
  }) =>
      OrderSettingGetSuccess(
          orderSettings: orderSettings ?? this.orderSettings);

  @override
  List<Object> get props => [orderSettings];
}

class OrderSettingGetFailed extends OrderSettingState {
  final String message;

  const OrderSettingGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderSettingUpdateInitial extends OrderSettingState {}

class OrderSettingUpdateInProgress extends OrderSettingState {}

class OrderSettingUpdateSuccess extends OrderSettingState {}

class OrderSettingUpdateFailed extends OrderSettingState {
  final String message;

  const OrderSettingUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
