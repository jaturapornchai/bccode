part of 'master_group_sub1_bloc.dart';

abstract class MasterGroupSub1Event extends Equatable {
  const MasterGroupSub1Event();

  @override
  List<Object> get props => [];
}

class MasterGroupSub1Get extends MasterGroupSub1Event {
  final String guid;

  const MasterGroupSub1Get({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterGroupSub1LoadList extends MasterGroupSub1Event {
  final int limit;
  final int offset;
  final String search;

  const MasterGroupSub1LoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterGroupSub1GetByCode extends MasterGroupSub1Event {
  final String code;

  const MasterGroupSub1GetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterGroupSub1Delete extends MasterGroupSub1Event {
  final String guid;

  const MasterGroupSub1Delete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGroupSub1DeleteMany extends MasterGroupSub1Event {
  final List<String> guid;

  const MasterGroupSub1DeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterGroupSub1Save extends MasterGroupSub1Event {
  final MasterGroupSub1Model groupSub1Model;

  const MasterGroupSub1Save({
    required this.groupSub1Model,
  });

  @override
  List<Object> get props => [groupSub1Model];
}

class MasterGroupSub1Update extends MasterGroupSub1Event {
  final String guid;
  final MasterGroupSub1Model groupSub1Model;

  const MasterGroupSub1Update({
    required this.guid,
    required this.groupSub1Model,
  });

  @override
  List<Object> get props => [guid, groupSub1Model];
}
