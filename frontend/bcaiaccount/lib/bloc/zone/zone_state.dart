part of 'zone_bloc.dart';

abstract class ZoneState extends Equatable {
  const ZoneState();

  @override
  List<Object> get props => [];
}

class ZoneInitial extends ZoneState {}

class ZoneInProgress extends ZoneState {}

class ZoneLoadSuccess extends ZoneState {
  final List<ZoneDataModel> zones;

  const ZoneLoadSuccess({required this.zones});

  ZoneLoadSuccess copyWith({
    List<ZoneDataModel>? zones,
  }) =>
      ZoneLoadSuccess(zones: zones ?? this.zones);

  @override
  List<Object> get props => [zones];
}

class ZoneLoadFailed extends ZoneState {
  final String message;

  const ZoneLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ZoneSaveInitial extends ZoneState {}

class ZoneSaveInProgress extends ZoneState {}

class ZoneSaveSuccess extends ZoneState {}

class ZoneSaveFailed extends ZoneState {
  final String message;

  const ZoneSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ZoneDeleteInProgress extends ZoneState {}

class ZoneDeleteSuccess extends ZoneState {}

class ZoneDeleteFailed extends ZoneState {
  final String message;

  const ZoneDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ZoneDeleteManyInProgress extends ZoneState {}

class ZoneDeleteManySuccess extends ZoneState {}

class ZoneDeleteManyFailed extends ZoneState {}

class ZoneGetInProgress extends ZoneState {}

class ZoneGetSuccess extends ZoneState {
  final ZoneDataModel zone;

  const ZoneGetSuccess({required this.zone});

  ZoneGetSuccess copyWith({
    ZoneDataModel? zone,
  }) =>
      ZoneGetSuccess(zone: zone ?? this.zone);

  @override
  List<Object> get props => [zone];
}

class ZoneGetFailed extends ZoneState {
  final String message;

  const ZoneGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ZoneUpdateInitial extends ZoneState {}

class ZoneUpdateInProgress extends ZoneState {}

class ZoneUpdateSuccess extends ZoneState {}

class ZoneUpdateFailed extends ZoneState {
  final String message;

  const ZoneUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
