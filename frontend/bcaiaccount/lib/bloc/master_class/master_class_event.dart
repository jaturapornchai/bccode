part of 'master_class_bloc.dart';

abstract class MasterClassEvent extends Equatable {
  const MasterClassEvent();

  @override
  List<Object> get props => [];
}

class MasterClassGet extends MasterClassEvent {
  final String guid;

  const MasterClassGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterClassLoadList extends MasterClassEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterClassLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterClassGetByCode extends MasterClassEvent {
  final String code;

  const MasterClassGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterClassDelete extends MasterClassEvent {
  final String guid;

  const MasterClassDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterClassDeleteMany extends MasterClassEvent {
  final List<String> guid;

  const MasterClassDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterClassSave extends MasterClassEvent {
  final MasterClassModel classModel;

  const MasterClassSave({
    required this.classModel,
  });

  @override
  List<Object> get props => [classModel];
}

class MasterClassUpdate extends MasterClassEvent {
  final String guid;
  final MasterClassModel classModel;

  const MasterClassUpdate({
    required this.guid,
    required this.classModel,
  });

  @override
  List<Object> get props => [guid, classModel];
}
