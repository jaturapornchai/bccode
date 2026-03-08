part of 'config_bloc.dart';

abstract class ConfigSystemState extends Equatable {
  const ConfigSystemState();

  @override
  List<Object> get props => [];
}

class ConfigSystemInitial extends ConfigSystemState {}

class ConfigSystemInProgress extends ConfigSystemState {}

class ConfigSystemLoadSuccess extends ConfigSystemState {
  final String guidFixed;
  final ConfigSystemModel data;

  const ConfigSystemLoadSuccess({required this.guidFixed, required this.data});

  @override
  List<Object> get props => [guidFixed, data];
}

class ConfigSystemLoadFailed extends ConfigSystemState {
  final String message;

  const ConfigSystemLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ConfigSystemSaveInitial extends ConfigSystemState {}

class ConfigSystemSaveInProgress extends ConfigSystemState {}

class ConfigSystemSaveSuccess extends ConfigSystemState {}

class ConfigSystemSaveFailed extends ConfigSystemState {
  final String message;

  const ConfigSystemSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ConfigSystemDeleteInProgress extends ConfigSystemState {}

class ConfigSystemDeleteSuccess extends ConfigSystemState {}

class ConfigSystemDeleteFailed extends ConfigSystemState {}

class ConfigSystemDeleteManyInProgress extends ConfigSystemState {}

class ConfigSystemDeleteManySuccess extends ConfigSystemState {}

class ConfigSystemDeleteManyFailed extends ConfigSystemState {}

class ConfigSystemGetInProgress extends ConfigSystemState {}

class ConfigSystemGetSuccess extends ConfigSystemState {
  final ConfigSystemModel data;

  const ConfigSystemGetSuccess({required this.data});

  ConfigSystemGetSuccess copyWith({
    ConfigSystemModel? data,
  }) =>
      ConfigSystemGetSuccess(data: data ?? this.data);

  @override
  List<Object> get props => [data];
}

class ConfigSystemGetFailed extends ConfigSystemState {
  final String message;

  const ConfigSystemGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ConfigSystemUpdateInitial extends ConfigSystemState {}

class ConfigSystemUpdateInProgress extends ConfigSystemState {}

class ConfigSystemUpdateSuccess extends ConfigSystemState {}

class ConfigSystemUpdateFailed extends ConfigSystemState {
  final String message;

  const ConfigSystemUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
