part of 'pos_media_bloc.dart';

abstract class PosMediaState extends Equatable {
  const PosMediaState();

  @override
  List<Object> get props => [];
}

class PosMediaInitial extends PosMediaState {}

class PosMediaInProgress extends PosMediaState {}

class PosMediaLoadSuccess extends PosMediaState {
  final List<PosMediaModel> posMedias;

  const PosMediaLoadSuccess({required this.posMedias});

  PosMediaLoadSuccess copyWith({
    List<PosMediaModel>? posMedias,
  }) =>
      PosMediaLoadSuccess(posMedias: posMedias ?? this.posMedias);

  @override
  List<Object> get props => [posMedias];
}

class PosMediaLoadFailed extends PosMediaState {
  final String message;

  const PosMediaLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PosMediaSaveInitial extends PosMediaState {}

class PosMediaSaveInProgress extends PosMediaState {}

class PosMediaSaveSuccess extends PosMediaState {}

class PosMediaSaveFailed extends PosMediaState {
  final String message;

  const PosMediaSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PosMediaDeleteInProgress extends PosMediaState {}

class PosMediaDeleteSuccess extends PosMediaState {}

class PosMediaDeleteFailed extends PosMediaState {}

class PosMediaDeleteManyInProgress extends PosMediaState {}

class PosMediaDeleteManySuccess extends PosMediaState {}

class PosMediaDeleteManyFailed extends PosMediaState {}

class PosMediaGetInProgress extends PosMediaState {}

class PosMediaGetSuccess extends PosMediaState {
  final PosMediaModel posMedia;

  const PosMediaGetSuccess({required this.posMedia});

  PosMediaGetSuccess copyWith({
    PosMediaModel? posMedia,
  }) =>
      PosMediaGetSuccess(posMedia: posMedia ?? this.posMedia);

  @override
  List<Object> get props => [posMedia];
}

class PosMediaGetFailed extends PosMediaState {
  final String message;

  const PosMediaGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PosMediaUpdateInitial extends PosMediaState {}

class PosMediaUpdateInProgress extends PosMediaState {}

class PosMediaUpdateSuccess extends PosMediaState {}

class PosMediaUpdateFailed extends PosMediaState {
  final String message;

  const PosMediaUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
