part of 'master_group_sub2_bloc.dart';

abstract class MasterGroupSub2State extends Equatable {
  const MasterGroupSub2State();

  @override
  List<Object> get props => [];
}

class MasterGroupSub2Initial extends MasterGroupSub2State {}

class MasterGroupSub2InProgress extends MasterGroupSub2State {}

class MasterGroupSub2LoadSuccess extends MasterGroupSub2State {
  final List<MasterGroupSub2Model> groupSub2s;

  const MasterGroupSub2LoadSuccess({required this.groupSub2s});

  MasterGroupSub2LoadSuccess copyWith({
    List<MasterGroupSub2Model>? groupSub2s,
  }) =>
      MasterGroupSub2LoadSuccess(groupSub2s: groupSub2s ?? this.groupSub2s);

  @override
  List<Object> get props => [groupSub2s];
}

class MasterGroupSub2LoadFailed extends MasterGroupSub2State {
  final String message;

  const MasterGroupSub2LoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSub2SaveInitial extends MasterGroupSub2State {}

class MasterGroupSub2SaveInProgress extends MasterGroupSub2State {}

class MasterGroupSub2SaveSuccess extends MasterGroupSub2State {}

class MasterGroupSub2SaveFailed extends MasterGroupSub2State {
  final String message;

  const MasterGroupSub2SaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSub2DeleteInProgress extends MasterGroupSub2State {}

class MasterGroupSub2DeleteSuccess extends MasterGroupSub2State {}

class MasterGroupSub2DeleteFailed extends MasterGroupSub2State {}

class MasterGroupSub2DeleteManyInProgress extends MasterGroupSub2State {}

class MasterGroupSub2DeleteManySuccess extends MasterGroupSub2State {}

class MasterGroupSub2DeleteManyFailed extends MasterGroupSub2State {}

class MasterGroupSub2GetInProgress extends MasterGroupSub2State {}

class MasterGroupSub2GetSuccess extends MasterGroupSub2State {
  final MasterGroupSub2Model groupSub2;

  const MasterGroupSub2GetSuccess({required this.groupSub2});

  MasterGroupSub2GetSuccess copyWith({
    MasterGroupSub2Model? groupSub2,
  }) =>
      MasterGroupSub2GetSuccess(groupSub2: groupSub2 ?? this.groupSub2);

  @override
  List<Object> get props => [groupSub2];
}

class MasterGroupSub2GetFailed extends MasterGroupSub2State {
  final String message;

  const MasterGroupSub2GetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSub2UpdateInitial extends MasterGroupSub2State {}

class MasterGroupSub2UpdateInProgress extends MasterGroupSub2State {}

class MasterGroupSub2UpdateSuccess extends MasterGroupSub2State {}

class MasterGroupSub2UpdateFailed extends MasterGroupSub2State {
  final String message;

  const MasterGroupSub2UpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
