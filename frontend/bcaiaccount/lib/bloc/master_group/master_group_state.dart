part of 'master_group_bloc.dart';

abstract class MasterGroupState extends Equatable {
  const MasterGroupState();

  @override
  List<Object> get props => [];
}

class MasterGroupInitial extends MasterGroupState {}

class MasterGroupInProgress extends MasterGroupState {}

class MasterGroupLoadSuccess extends MasterGroupState {
  final List<MasterGroupModel> groups;

  const MasterGroupLoadSuccess({required this.groups});

  MasterGroupLoadSuccess copyWith({
    List<MasterGroupModel>? groups,
  }) =>
      MasterGroupLoadSuccess(groups: groups ?? this.groups);

  @override
  List<Object> get props => [groups];
}

class MasterGroupLoadFailed extends MasterGroupState {
  final String message;

  const MasterGroupLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupSaveInitial extends MasterGroupState {}

class MasterGroupSaveInProgress extends MasterGroupState {}

class MasterGroupSaveSuccess extends MasterGroupState {}

class MasterGroupSaveFailed extends MasterGroupState {
  final String message;

  const MasterGroupSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupDeleteInProgress extends MasterGroupState {}

class MasterGroupDeleteSuccess extends MasterGroupState {}

class MasterGroupDeleteFailed extends MasterGroupState {}

class MasterGroupDeleteManyInProgress extends MasterGroupState {}

class MasterGroupDeleteManySuccess extends MasterGroupState {}

class MasterGroupDeleteManyFailed extends MasterGroupState {}

class MasterGroupGetInProgress extends MasterGroupState {}

class MasterGroupGetSuccess extends MasterGroupState {
  final MasterGroupModel group;

  const MasterGroupGetSuccess({required this.group});

  MasterGroupGetSuccess copyWith({
    MasterGroupModel? group,
  }) =>
      MasterGroupGetSuccess(group: group ?? this.group);

  @override
  List<Object> get props => [group];
}

class MasterGroupGetFailed extends MasterGroupState {
  final String message;

  const MasterGroupGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterGroupUpdateInitial extends MasterGroupState {}

class MasterGroupUpdateInProgress extends MasterGroupState {}

class MasterGroupUpdateSuccess extends MasterGroupState {}

class MasterGroupUpdateFailed extends MasterGroupState {
  final String message;

  const MasterGroupUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
