part of 'pos_setting_bloc.dart';

abstract class PosSettingState extends Equatable {
  const PosSettingState();

  @override
  List<Object> get props => [];
}

class PosSettingInitial extends PosSettingState {}

class PosSettingInProgress extends PosSettingState {}

class PosSettingLoadSuccess extends PosSettingState {
  final List<PosSettingModel> posSettings;

  const PosSettingLoadSuccess({required this.posSettings});

  PosSettingLoadSuccess copyWith({
    List<PosSettingModel>? posSettings,
  }) =>
      PosSettingLoadSuccess(posSettings: posSettings ?? this.posSettings);

  @override
  List<Object> get props => [posSettings];
}

class PosSettingLoadFailed extends PosSettingState {
  final String message;

  const PosSettingLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PosSettingSaveInitial extends PosSettingState {}

class PosSettingSaveInProgress extends PosSettingState {}

class PosSettingSaveSuccess extends PosSettingState {}

class PosSettingSaveFailed extends PosSettingState {
  final String message;

  const PosSettingSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PosSettingDeleteInProgress extends PosSettingState {}

class PosSettingDeleteSuccess extends PosSettingState {}

class PosSettingDeleteFailed extends PosSettingState {}

class PosSettingDeleteManyInProgress extends PosSettingState {}

class PosSettingDeleteManySuccess extends PosSettingState {}

class PosSettingDeleteManyFailed extends PosSettingState {}

class PosSettingGetInProgress extends PosSettingState {}

class PosSettingGetSuccess extends PosSettingState {
  final PosSettingModel posSetting;

  const PosSettingGetSuccess({required this.posSetting});

  PosSettingGetSuccess copyWith({
    PosSettingModel? posSetting,
  }) =>
      PosSettingGetSuccess(posSetting: posSetting ?? this.posSetting);

  @override
  List<Object> get props => [posSetting];
}

class PosSettingGetFailed extends PosSettingState {
  final String message;

  const PosSettingGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PosSettingUpdateInitial extends PosSettingState {}

class PosSettingUpdateInProgress extends PosSettingState {}

class PosSettingUpdateSuccess extends PosSettingState {}

class PosSettingUpdateFailed extends PosSettingState {
  final String message;

  const PosSettingUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class GetApiKeyInitial extends PosSettingState {}

class GetApiKeyInProgress extends PosSettingState {}

class GetApiKeySuccess extends PosSettingState {
  final bool success;
  final String token;

  const GetApiKeySuccess({
    required this.success,
    required this.token,
  });

  @override
  List<Object> get props => [success, token];
}

class GetApiKeyFailed extends PosSettingState {
  final String message;

  const GetApiKeyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

/// delete apikey
class DeleteApikeyInitial extends PosSettingState {}

class DeleteApikeyInProgress extends PosSettingState {}

class DeleteApikeySuccess extends PosSettingState {}

class DeleteApikeyFailed extends PosSettingState {
  final String message;

  const DeleteApikeyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
