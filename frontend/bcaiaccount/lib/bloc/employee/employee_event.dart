part of 'employee_bloc.dart';

abstract class EmployeeEvent extends Equatable {
  const EmployeeEvent();

  @override
  List<Object> get props => [];
}

class EmployeeGet extends EmployeeEvent {
  final String guid;

  const EmployeeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class EmployeeLoadList extends EmployeeEvent {
  final int limit;
  final int offset;
  final String search;

  const EmployeeLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class EmployeeDelete extends EmployeeEvent {
  final String guid;

  const EmployeeDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class EmployeeDeleteMany extends EmployeeEvent {
  final List<String> guid;

  const EmployeeDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class EmployeeSave extends EmployeeEvent {
  final EmployeeModel employeeModel;

  const EmployeeSave({
    required this.employeeModel,
  });

  @override
  List<Object> get props => [employeeModel];
}

class EmployeeUpdate extends EmployeeEvent {
  final String guid;
  final EmployeeModel employeeModel;

  const EmployeeUpdate({
    required this.guid,
    required this.employeeModel,
  });

  @override
  List<Object> get props => [employeeModel];
}

class EmployeeWithImageSave extends EmployeeEvent {
  final File imageFile;
  final EmployeeModel employee;
  final Uint8List? imageWeb;
  const EmployeeWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.employee,
  });

  @override
  List<Object> get props => [employee, imageFile];
}

class EmployeeWithImageUpdate extends EmployeeEvent {
  final String guid;
  final EmployeeModel employee;
  final File imageFile;
  final Uint8List imageWeb;
  const EmployeeWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.employee,
  });

  @override
  List<Object> get props => [employee, imageWeb, employee];
}
