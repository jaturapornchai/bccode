part of 'order_template_setting_bloc.dart';

abstract class OrderTemplateSettingState extends Equatable {
  const OrderTemplateSettingState();

  @override
  List<Object> get props => [];
}

class OrderTemplateSettingInitial extends OrderTemplateSettingState {}

class OrderTemplateSettingInProgress extends OrderTemplateSettingState {}

class OrderTemplateSettingLoadSuccess extends OrderTemplateSettingState {
  final List<OrderTemplateSettingModel> orderTemplateSettings;

  const OrderTemplateSettingLoadSuccess({required this.orderTemplateSettings});

  OrderTemplateSettingLoadSuccess copyWith({
    List<OrderTemplateSettingModel>? orderTemplateSettings,
  }) =>
      OrderTemplateSettingLoadSuccess(
          orderTemplateSettings:
              orderTemplateSettings ?? this.orderTemplateSettings);

  @override
  List<Object> get props => [orderTemplateSettings];
}

class OrderTemplateSettingLoadFailed extends OrderTemplateSettingState {
  final String message;

  const OrderTemplateSettingLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTemplateSettingSaveInitial extends OrderTemplateSettingState {}

class OrderTemplateSettingSaveInProgress extends OrderTemplateSettingState {}

class OrderTemplateSettingSaveSuccess extends OrderTemplateSettingState {}

class OrderTemplateSettingSaveFailed extends OrderTemplateSettingState {
  final String message;

  const OrderTemplateSettingSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTemplateSettingDeleteInProgress extends OrderTemplateSettingState {}

class OrderTemplateSettingDeleteSuccess extends OrderTemplateSettingState {}

class OrderTemplateSettingDeleteFailed extends OrderTemplateSettingState {}

class OrderTemplateSettingDeleteManyInProgress
    extends OrderTemplateSettingState {}

class OrderTemplateSettingDeleteManySuccess extends OrderTemplateSettingState {}

class OrderTemplateSettingDeleteManyFailed extends OrderTemplateSettingState {}

class OrderTemplateSettingGetInProgress extends OrderTemplateSettingState {}

class OrderTemplateSettingGetSuccess extends OrderTemplateSettingState {
  final OrderTemplateSettingModel orderTemplateSettings;

  const OrderTemplateSettingGetSuccess({required this.orderTemplateSettings});

  OrderTemplateSettingGetSuccess copyWith({
    OrderTemplateSettingModel? orderTemplateSettings,
  }) =>
      OrderTemplateSettingGetSuccess(
          orderTemplateSettings:
              orderTemplateSettings ?? this.orderTemplateSettings);

  @override
  List<Object> get props => [orderTemplateSettings];
}

class OrderTemplateSettingGetFailed extends OrderTemplateSettingState {
  final String message;

  const OrderTemplateSettingGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class OrderTemplateSettingUpdateInitial extends OrderTemplateSettingState {}

class OrderTemplateSettingUpdateInProgress extends OrderTemplateSettingState {}

class OrderTemplateSettingUpdateSuccess extends OrderTemplateSettingState {}

class OrderTemplateSettingUpdateFailed extends OrderTemplateSettingState {
  final String message;

  const OrderTemplateSettingUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
