part of 'master_grade_bloc.dart';

abstract class MasterGradeState extends Equatable {
  const MasterGradeState();

  @override
  List<Object> get props => [];
}

class MasterGradeInitial extends MasterGradeState {}

class MasterGradeInProgress extends MasterGradeState {}

class MasterGradeLoadSuccess extends MasterGradeState {
  final List<MasterGradeModel> grades;

  const MasterGradeLoadSuccess({required this.grades});

  MasterGradeLoadSuccess copyWith({
    List<MasterGradeModel>? grades,
  }) =>
      MasterGradeLoadSuccess(grades: grades ?? this.grades);

  @override
  List<Object> get props => [grades];
}

class MasterGradeLoadFailed extends MasterGradeState {
  final String message;

  const MasterGradeLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGradeSaveInitial extends MasterGradeState {}

class MasterGradeSaveInProgress extends MasterGradeState {}

class MasterGradeSaveSuccess extends MasterGradeState {}

class MasterGradeSaveFailed extends MasterGradeState {
  final String message;

  const MasterGradeSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGradeDeleteInProgress extends MasterGradeState {}

class MasterGradeDeleteSuccess extends MasterGradeState {}

class MasterGradeDeleteFailed extends MasterGradeState {}

class MasterGradeDeleteManyInProgress extends MasterGradeState {}

class MasterGradeDeleteManySuccess extends MasterGradeState {}

class MasterGradeDeleteManyFailed extends MasterGradeState {}

class MasterGradeGetInProgress extends MasterGradeState {}

class MasterGradeGetSuccess extends MasterGradeState {
  final MasterGradeModel grade;

  const MasterGradeGetSuccess({required this.grade});

  MasterGradeGetSuccess copyWith({
    MasterGradeModel? grade,
  }) =>
      MasterGradeGetSuccess(grade: grade ?? this.grade);

  @override
  List<Object> get props => [grade];
}

class MasterGradeGetFailed extends MasterGradeState {
  final String message;

  const MasterGradeGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGradeUpdateInitial extends MasterGradeState {}

class MasterGradeUpdateInProgress extends MasterGradeState {}

class MasterGradeUpdateSuccess extends MasterGradeState {}

class MasterGradeUpdateFailed extends MasterGradeState {
  final String message;

  const MasterGradeUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
