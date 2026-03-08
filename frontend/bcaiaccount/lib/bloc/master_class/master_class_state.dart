part of 'master_class_bloc.dart';

abstract class MasterClassState extends Equatable {
  const MasterClassState();

  @override
  List<Object> get props => [];
}

class MasterClassInitial extends MasterClassState {}

class MasterClassInProgress extends MasterClassState {}

class MasterClassLoadSuccess extends MasterClassState {
  final List<MasterClassModel> classes;

  const MasterClassLoadSuccess({required this.classes});

  MasterClassLoadSuccess copyWith({
    List<MasterClassModel>? classes,
  }) =>
      MasterClassLoadSuccess(classes: classes ?? this.classes);

  @override
  List<Object> get props => [classes];
}

class MasterClassLoadFailed extends MasterClassState {
  final String message;

  const MasterClassLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterClassSaveInitial extends MasterClassState {}

class MasterClassSaveInProgress extends MasterClassState {}

class MasterClassSaveSuccess extends MasterClassState {}

class MasterClassSaveFailed extends MasterClassState {
  final String message;

  const MasterClassSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterClassDeleteInProgress extends MasterClassState {}

class MasterClassDeleteSuccess extends MasterClassState {}

class MasterClassDeleteFailed extends MasterClassState {}

class MasterClassDeleteManyInProgress extends MasterClassState {}

class MasterClassDeleteManySuccess extends MasterClassState {}

class MasterClassDeleteManyFailed extends MasterClassState {}

class MasterClassGetInProgress extends MasterClassState {}

class MasterClassGetSuccess extends MasterClassState {
  final MasterClassModel classData;

  const MasterClassGetSuccess({required this.classData});

  MasterClassGetSuccess copyWith({
    MasterClassModel? classData,
  }) =>
      MasterClassGetSuccess(classData: classData ?? this.classData);

  @override
  List<Object> get props => [classData];
}

class MasterClassGetFailed extends MasterClassState {
  final String message;

  const MasterClassGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterClassUpdateInitial extends MasterClassState {}

class MasterClassUpdateInProgress extends MasterClassState {}

class MasterClassUpdateSuccess extends MasterClassState {}

class MasterClassUpdateFailed extends MasterClassState {
  final String message;

  const MasterClassUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
