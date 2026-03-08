part of 'devices_bloc.dart';

abstract class DevicesState extends Equatable {
  const DevicesState();

  @override
  List<Object> get props => [];
}

class DevicesInitial extends DevicesState {}

class DevicesInProgress extends DevicesState {}

class DevicesLoadSuccess extends DevicesState {
  final List<DevicesModel> devices;

  const DevicesLoadSuccess({required this.devices});

  DevicesLoadSuccess copyWith({
    List<DevicesModel>? devices,
  }) =>
      DevicesLoadSuccess(devices: devices ?? this.devices);

  @override
  List<Object> get props => [devices];
}

class DevicesLoadFailed extends DevicesState {
  final String message;

  const DevicesLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DevicesSaveInitial extends DevicesState {}

class DevicesSaveInProgress extends DevicesState {}

class DevicesSaveSuccess extends DevicesState {}

class DevicesSaveFailed extends DevicesState {
  final String message;

  const DevicesSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DevicesDeleteInProgress extends DevicesState {}

class DevicesDeleteSuccess extends DevicesState {}

class DevicesDeleteFailed extends DevicesState {}

class DevicesDeleteManyInProgress extends DevicesState {}

class DevicesDeleteManySuccess extends DevicesState {}

class DevicesDeleteManyFailed extends DevicesState {}

class DevicesGetInProgress extends DevicesState {}

class DevicesGetSuccess extends DevicesState {
  final DevicesModel devices;

  const DevicesGetSuccess({required this.devices});

  DevicesGetSuccess copyWith({
    DevicesModel? devices,
  }) =>
      DevicesGetSuccess(devices: devices ?? this.devices);

  @override
  List<Object> get props => [devices];
}

class DevicesGetFailed extends DevicesState {
  final String message;

  const DevicesGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DevicesUpdateInitial extends DevicesState {}

class DevicesUpdateInProgress extends DevicesState {}

class DevicesUpdateSuccess extends DevicesState {}

class DevicesUpdateFailed extends DevicesState {
  final String message;

  const DevicesUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
