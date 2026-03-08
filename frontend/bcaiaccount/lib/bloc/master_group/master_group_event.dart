part of 'master_group_bloc.dart';

abstract class MasterGroupEvent extends Equatable {
  const MasterGroupEvent();

  @override
  List<Object> get props => [];
}

class MasterGroupGet extends MasterGroupEvent {
  final String guid;

  const MasterGroupGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterGroupLoadList extends MasterGroupEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterGroupLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterGroupGetByCode extends MasterGroupEvent {
  final String code;

  const MasterGroupGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterGroupDelete extends MasterGroupEvent {
  final String guid;

  const MasterGroupDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGroupDeleteMany extends MasterGroupEvent {
  final List<String> guid;

  const MasterGroupDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGroupSave extends MasterGroupEvent {
  final MasterGroupModel groupModel;

  const MasterGroupSave({
    required this.groupModel,
  });

  @override
  List<Object> get props => [groupModel];
}

class MasterGroupUpdate extends MasterGroupEvent {
  final String guid;
  final MasterGroupModel groupModel;

  const MasterGroupUpdate({
    required this.guid,
    required this.groupModel,
  });

  @override
  List<Object> get props => [guid, groupModel];
}
