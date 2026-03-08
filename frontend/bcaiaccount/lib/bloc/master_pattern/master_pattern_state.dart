part of 'master_pattern_bloc.dart';

abstract class MasterPatternState extends Equatable {
  const MasterPatternState();

  @override
  List<Object> get props => [];
}

class MasterPatternInitial extends MasterPatternState {}

class MasterPatternInProgress extends MasterPatternState {}

class MasterPatternLoadSuccess extends MasterPatternState {
  final List<MasterPatternModel> patterns;

  const MasterPatternLoadSuccess({required this.patterns});

  MasterPatternLoadSuccess copyWith({
    List<MasterPatternModel>? patterns,
  }) =>
      MasterPatternLoadSuccess(patterns: patterns ?? this.patterns);

  @override
  List<Object> get props => [patterns];
}

class MasterPatternLoadFailed extends MasterPatternState {
  final String message;

  const MasterPatternLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterPatternSaveInitial extends MasterPatternState {}

class MasterPatternSaveInProgress extends MasterPatternState {}

class MasterPatternSaveSuccess extends MasterPatternState {}

class MasterPatternSaveFailed extends MasterPatternState {
  final String message;

  const MasterPatternSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterPatternDeleteInProgress extends MasterPatternState {}

class MasterPatternDeleteSuccess extends MasterPatternState {}

class MasterPatternDeleteFailed extends MasterPatternState {}

class MasterPatternDeleteManyInProgress extends MasterPatternState {}

class MasterPatternDeleteManySuccess extends MasterPatternState {}

class MasterPatternDeleteManyFailed extends MasterPatternState {}

class MasterPatternGetInProgress extends MasterPatternState {}

class MasterPatternGetSuccess extends MasterPatternState {
  final MasterPatternModel pattern;

  const MasterPatternGetSuccess({required this.pattern});

  MasterPatternGetSuccess copyWith({
    MasterPatternModel? pattern,
  }) =>
      MasterPatternGetSuccess(pattern: pattern ?? this.pattern);

  @override
  List<Object> get props => [pattern];
}

class MasterPatternGetFailed extends MasterPatternState {
  final String message;

  const MasterPatternGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterPatternUpdateInitial extends MasterPatternState {}

class MasterPatternUpdateInProgress extends MasterPatternState {}

class MasterPatternUpdateSuccess extends MasterPatternState {}

class MasterPatternUpdateFailed extends MasterPatternState {
  final String message;

  const MasterPatternUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
