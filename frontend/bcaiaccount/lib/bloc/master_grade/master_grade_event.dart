part of 'master_grade_bloc.dart';

abstract class MasterGradeEvent extends Equatable {
  const MasterGradeEvent();

  @override
  List<Object> get props => [];
}

class MasterGradeGet extends MasterGradeEvent {
  final String guid;

  const MasterGradeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterGradeLoadList extends MasterGradeEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterGradeLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterGradeGetByCode extends MasterGradeEvent {
  final String code;

  const MasterGradeGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterGradeDelete extends MasterGradeEvent {
  final String guid;

  const MasterGradeDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGradeDeleteMany extends MasterGradeEvent {
  final List<String> guid;

  const MasterGradeDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGradeSave extends MasterGradeEvent {
  final MasterGradeModel gradeModel;

  const MasterGradeSave({
    required this.gradeModel,
  });

  @override
  List<Object> get props => [gradeModel];
}

class MasterGradeUpdate extends MasterGradeEvent {
  final String guid;
  final MasterGradeModel gradeModel;

  const MasterGradeUpdate({
    required this.guid,
    required this.gradeModel,
  });

  @override
  List<Object> get props => [guid, gradeModel];
}
