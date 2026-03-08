part of 'department_bloc.dart';

abstract class DepartmentState extends Equatable {
  const DepartmentState();

  @override
  List<Object> get props => [];
}

class DepartmentInitial extends DepartmentState {}

class DepartmentInProgress extends DepartmentState {}

class DepartmentLoadSuccess extends DepartmentState {
  final List<DepartmentModel> department;

  const DepartmentLoadSuccess({required this.department});

  DepartmentLoadSuccess copyWith({
    List<DepartmentModel>? department,
  }) =>
      DepartmentLoadSuccess(department: department ?? this.department);

  @override
  List<Object> get props => [department];
}

class DepartmentLoadFailed extends DepartmentState {
  final String message;

  const DepartmentLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DepartmentSaveInitial extends DepartmentState {}

class DepartmentSaveInProgress extends DepartmentState {}

class DepartmentSaveSuccess extends DepartmentState {
  final String responsesID;

  const DepartmentSaveSuccess({
    required this.responsesID,
  });

  @override
  List<Object> get props => [responsesID];
}

class DepartmentSaveFailed extends DepartmentState {
  final String message;

  const DepartmentSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DepartmentDeleteInProgress extends DepartmentState {}

class DepartmentDeleteSuccess extends DepartmentState {}

class DepartmentDeleteFailed extends DepartmentState {}

class DepartmentDeleteManyInProgress extends DepartmentState {}

class DepartmentDeleteManySuccess extends DepartmentState {}

class DepartmentDeleteManyFailed extends DepartmentState {}

class DepartmentGetInProgress extends DepartmentState {}

class DepartmentGetSuccess extends DepartmentState {
  final DepartmentModel department;

  const DepartmentGetSuccess({required this.department});

  DepartmentGetSuccess copyWith({
    DepartmentModel? department,
  }) =>
      DepartmentGetSuccess(department: department ?? this.department);

  @override
  List<Object> get props => [department];
}

class DepartmentGetFailed extends DepartmentState {
  final String message;

  const DepartmentGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DepartmentUpdateInitial extends DepartmentState {}

class DepartmentUpdateInProgress extends DepartmentState {}

class DepartmentUpdateSuccess extends DepartmentState {}

class DepartmentUpdateFailed extends DepartmentState {
  final String message;

  const DepartmentUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
