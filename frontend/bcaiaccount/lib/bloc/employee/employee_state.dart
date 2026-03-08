part of 'employee_bloc.dart';

abstract class EmployeeState extends Equatable {
  const EmployeeState();

  @override
  List<Object> get props => [];
}

class EmployeeInitial extends EmployeeState {}

class EmployeeInProgress extends EmployeeState {}

class EmployeeLoadSuccess extends EmployeeState {
  final List<EmployeeModel> employees;

  const EmployeeLoadSuccess({required this.employees});

  EmployeeLoadSuccess copyWith({
    List<EmployeeModel>? employees,
  }) =>
      EmployeeLoadSuccess(employees: employees ?? this.employees);

  @override
  List<Object> get props => [employees];
}

class EmployeeLoadFailed extends EmployeeState {
  final String message;

  const EmployeeLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class EmployeeSaveInitial extends EmployeeState {}

class EmployeeSaveInProgress extends EmployeeState {}

class EmployeeSaveSuccess extends EmployeeState {}

class EmployeeSaveFailed extends EmployeeState {
  final String message;

  const EmployeeSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class EmployeeDeleteInProgress extends EmployeeState {}

class EmployeeDeleteSuccess extends EmployeeState {}

class EmployeeDeleteFailed extends EmployeeState {}

class EmployeeDeleteManyInProgress extends EmployeeState {}

class EmployeeDeleteManySuccess extends EmployeeState {}

class EmployeeDeleteManyFailed extends EmployeeState {}

class EmployeeGetInProgress extends EmployeeState {}

class EmployeeGetSuccess extends EmployeeState {
  final EmployeeModel employee;

  const EmployeeGetSuccess({required this.employee});

  EmployeeGetSuccess copyWith({
    EmployeeModel? employee,
  }) =>
      EmployeeGetSuccess(employee: employee ?? this.employee);

  @override
  List<Object> get props => [employee];
}

class EmployeeGetFailed extends EmployeeState {
  final String message;

  const EmployeeGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class EmployeeUpdateInitial extends EmployeeState {}

class EmployeeUpdateInProgress extends EmployeeState {}

class EmployeeUpdateSuccess extends EmployeeState {}

class EmployeeUpdateFailed extends EmployeeState {
  final String message;

  const EmployeeUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
