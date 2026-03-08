part of 'device_print_bloc.dart';

abstract class DevicePrintState extends Equatable {
  const DevicePrintState();

  @override
  List<Object> get props => [];
}

class DevicePrintInitial extends DevicePrintState {}

class DevicePrintInProgress extends DevicePrintState {}

class DevicePrintLoadSuccess extends DevicePrintState {
  final List<DevicePrinterSaveModel> devicePrints;

  const DevicePrintLoadSuccess({required this.devicePrints});

  DevicePrintLoadSuccess copyWith({
    String guid = '',
    List<DevicePrinterSaveModel>? devicePrints,
  }) =>
      DevicePrintLoadSuccess(devicePrints: devicePrints ?? this.devicePrints);

  @override
  List<Object> get props => [devicePrints];
}

class DevicePrintLoadFailed extends DevicePrintState {
  final String message;

  const DevicePrintLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DevicePrintSaveInitial extends DevicePrintState {}

class DevicePrintSaveInProgress extends DevicePrintState {}

class DevicePrintSaveSuccess extends DevicePrintState {}

class DevicePrintSaveFailed extends DevicePrintState {
  final String message;

  const DevicePrintSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DevicePrintDeleteInProgress extends DevicePrintState {}

class DevicePrintDeleteSuccess extends DevicePrintState {}

class DevicePrintDeleteFailed extends DevicePrintState {}

class DevicePrintDeleteManyInProgress extends DevicePrintState {}

class DevicePrintDeleteManySuccess extends DevicePrintState {}

class DevicePrintDeleteManyFailed extends DevicePrintState {}

class DevicePrintGetInProgress extends DevicePrintState {}

class DevicePrintGetSuccess extends DevicePrintState {
  final DevicePrinterModel devicePrints;

  const DevicePrintGetSuccess({required this.devicePrints});

  DevicePrintGetSuccess copyWith({
    DevicePrinterModel? devicePrints,
  }) =>
      DevicePrintGetSuccess(devicePrints: devicePrints ?? this.devicePrints);

  @override
  List<Object> get props => [devicePrints];
}

class DevicePrintGetFailed extends DevicePrintState {
  final String message;

  const DevicePrintGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DevicePrintUpdateInitial extends DevicePrintState {}

class DevicePrintUpdateInProgress extends DevicePrintState {}

class DevicePrintUpdateSuccess extends DevicePrintState {}

class DevicePrintUpdateFailed extends DevicePrintState {
  final String message;

  const DevicePrintUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
