part of 'master_group_sub1_bloc.dart';

abstract class MasterGroupSub1State extends Equatable {
  const MasterGroupSub1State();

  @override
  List<Object> get props => [];
}

class MasterGroupSub1Initial extends MasterGroupSub1State {}

class MasterGroupSub1InProgress extends MasterGroupSub1State {}

class MasterGroupSub1LoadSuccess extends MasterGroupSub1State {
  final List<MasterGroupSub1Model> groupSub1s;

  const MasterGroupSub1LoadSuccess({required this.groupSub1s});

  MasterGroupSub1LoadSuccess copyWith({
    List<MasterGroupSub1Model>? groupSub1s,
  }) =>
      MasterGroupSub1LoadSuccess(groupSub1s: groupSub1s ?? this.groupSub1s);

  @override
  List<Object> get props => [groupSub1s];
}

class MasterGroupSub1LoadFailed extends MasterGroupSub1State {
  final String message;

  const MasterGroupSub1LoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSub1SaveInitial extends MasterGroupSub1State {}

class MasterGroupSub1SaveInProgress extends MasterGroupSub1State {}

class MasterGroupSub1SaveSuccess extends MasterGroupSub1State {}

class MasterGroupSub1SaveFailed extends MasterGroupSub1State {
  final String message;

  const MasterGroupSub1SaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSub1DeleteInProgress extends MasterGroupSub1State {}

class MasterGroupSub1DeleteSuccess extends MasterGroupSub1State {}

class MasterGroupSub1DeleteFailed extends MasterGroupSub1State {}

class MasterGroupSub1DeleteManyInProgress extends MasterGroupSub1State {}

class MasterGroupSub1DeleteManySuccess extends MasterGroupSub1State {}

class MasterGroupSub1DeleteManyFailed extends MasterGroupSub1State {}

class MasterGroupSub1GetInProgress extends MasterGroupSub1State {}

class MasterGroupSub1GetSuccess extends MasterGroupSub1State {
  final MasterGroupSub1Model groupSub1;

  const MasterGroupSub1GetSuccess({required this.groupSub1});

  MasterGroupSub1GetSuccess copyWith({
    MasterGroupSub1Model? groupSub1,
  }) =>
      MasterGroupSub1GetSuccess(groupSub1: groupSub1 ?? this.groupSub1);

  @override
  List<Object> get props => [groupSub1];
}

class MasterGroupSub1GetFailed extends MasterGroupSub1State {
  final String message;

  const MasterGroupSub1GetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSub1UpdateInitial extends MasterGroupSub1State {}

class MasterGroupSub1UpdateInProgress extends MasterGroupSub1State {}

class MasterGroupSub1UpdateSuccess extends MasterGroupSub1State {}

class MasterGroupSub1UpdateFailed extends MasterGroupSub1State {
  final String message;

  const MasterGroupSub1UpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
