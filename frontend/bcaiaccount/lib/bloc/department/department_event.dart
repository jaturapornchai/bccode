part of 'department_bloc.dart';

abstract class DepartmentEvent extends Equatable {
  const DepartmentEvent();

  @override
  List<Object> get props => [];
}

class DepartmentGet extends DepartmentEvent {
  final String guid;

  const DepartmentGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DepartmentLoadList extends DepartmentEvent {
  final int limit;
  final int offset;
  final String search;

  const DepartmentLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class DepartmentDelete extends DepartmentEvent {
  final String guid;

  const DepartmentDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DepartmentDeleteMany extends DepartmentEvent {
  final List<String> guid;

  const DepartmentDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DepartmentSave extends DepartmentEvent {
  final DepartmentModel department;

  const DepartmentSave({
    required this.department,
  });

  @override
  List<Object> get props => [department];
}

class DepartmentUpdate extends DepartmentEvent {
  final String guid;
  final DepartmentModel department;

  const DepartmentUpdate({
    required this.guid,
    required this.department,
  });

  @override
  List<Object> get props => [department];
}
