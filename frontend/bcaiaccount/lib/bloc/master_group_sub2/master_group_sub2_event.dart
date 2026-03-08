part of 'master_group_sub2_bloc.dart';

abstract class MasterGroupSub2Event extends Equatable {
  const MasterGroupSub2Event();

  @override
  List<Object> get props => [];
}

class MasterGroupSub2Get extends MasterGroupSub2Event {
  final String guid;

  const MasterGroupSub2Get({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterGroupSub2LoadList extends MasterGroupSub2Event {
  final int limit;
  final int offset;
  final String search;

  const MasterGroupSub2LoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterGroupSub2GetByCode extends MasterGroupSub2Event {
  final String code;

  const MasterGroupSub2GetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterGroupSub2Delete extends MasterGroupSub2Event {
  final String guid;

  const MasterGroupSub2Delete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGroupSub2DeleteMany extends MasterGroupSub2Event {
  final List<String> guid;

  const MasterGroupSub2DeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGroupSub2Save extends MasterGroupSub2Event {
  final MasterGroupSub2Model groupSub2Model;

  const MasterGroupSub2Save({
    required this.groupSub2Model,
  });

  @override
  List<Object> get props => [groupSub2Model];
}

class MasterGroupSub2Update extends MasterGroupSub2Event {
  final String guid;
  final MasterGroupSub2Model groupSub2Model;

  const MasterGroupSub2Update({
    required this.guid,
    required this.groupSub2Model,
  });

  @override
  List<Object> get props => [guid, groupSub2Model];
}
