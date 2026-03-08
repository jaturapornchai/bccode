part of 'master_model_bloc.dart';

abstract class MasterModelState extends Equatable {
  const MasterModelState();

  @override
  List<Object> get props => [];
}

class MasterModelInitial extends MasterModelState {}

class MasterModelInProgress extends MasterModelState {}

class MasterModelLoadSuccess extends MasterModelState {
  final List<MasterModelModel> models;

  const MasterModelLoadSuccess({required this.models});

  MasterModelLoadSuccess copyWith({
    List<MasterModelModel>? models,
  }) =>
      MasterModelLoadSuccess(models: models ?? this.models);

  @override
  List<Object> get props => [models];
}

class MasterModelLoadFailed extends MasterModelState {
  final String message;

  const MasterModelLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterModelSaveInitial extends MasterModelState {}

class MasterModelSaveInProgress extends MasterModelState {}

class MasterModelSaveSuccess extends MasterModelState {}

class MasterModelSaveFailed extends MasterModelState {
  final String message;

  const MasterModelSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterModelDeleteInProgress extends MasterModelState {}

class MasterModelDeleteSuccess extends MasterModelState {}

class MasterModelDeleteFailed extends MasterModelState {}

class MasterModelDeleteManyInProgress extends MasterModelState {}

class MasterModelDeleteManySuccess extends MasterModelState {}

class MasterModelDeleteManyFailed extends MasterModelState {}

class MasterModelGetInProgress extends MasterModelState {}

class MasterModelGetSuccess extends MasterModelState {
  final MasterModelModel model;

  const MasterModelGetSuccess({required this.model});

  MasterModelGetSuccess copyWith({
    MasterModelModel? model,
  }) =>
      MasterModelGetSuccess(model: model ?? this.model);

  @override
  List<Object> get props => [model];
}

class MasterModelGetFailed extends MasterModelState {
  final String message;

  const MasterModelGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterModelUpdateInitial extends MasterModelState {}

class MasterModelUpdateInProgress extends MasterModelState {}

class MasterModelUpdateSuccess extends MasterModelState {}

class MasterModelUpdateFailed extends MasterModelState {
  final String message;

  const MasterModelUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
