part of 'debtor_group_bloc.dart';

abstract class DebtorGroupState extends Equatable {
  const DebtorGroupState();

  @override
  List<Object> get props => [];
}

class DebtorGroupInitial extends DebtorGroupState {}

class DebtorGroupInProgress extends DebtorGroupState {}

class DebtorGroupLoadSuccess extends DebtorGroupState {
  final List<DebtorGroupModel> debtorGroups;

  const DebtorGroupLoadSuccess({required this.debtorGroups});

  DebtorGroupLoadSuccess copyWith({
    List<DebtorGroupModel>? debtorGroups,
  }) =>
      DebtorGroupLoadSuccess(debtorGroups: debtorGroups ?? this.debtorGroups);

  @override
  List<Object> get props => [debtorGroups];
}

class DebtorGroupLoadFailed extends DebtorGroupState {
  final String message;

  const DebtorGroupLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DebtorGroupSaveInitial extends DebtorGroupState {}

class DebtorGroupSaveInProgress extends DebtorGroupState {}

class DebtorGroupSaveSuccess extends DebtorGroupState {}

class DebtorGroupSaveFailed extends DebtorGroupState {
  final String message;

  const DebtorGroupSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DebtorGroupDeleteInProgress extends DebtorGroupState {}

class DebtorGroupDeleteSuccess extends DebtorGroupState {}

class DebtorGroupDeleteFailed extends DebtorGroupState {}

class DebtorGroupDeleteManyInProgress extends DebtorGroupState {}

class DebtorGroupDeleteManySuccess extends DebtorGroupState {}

class DebtorGroupDeleteManyFailed extends DebtorGroupState {}

class DebtorGroupGetInProgress extends DebtorGroupState {}

class DebtorGroupGetSuccess extends DebtorGroupState {
  final DebtorGroupModel debtorGroups;

  const DebtorGroupGetSuccess({required this.debtorGroups});

  DebtorGroupGetSuccess copyWith({
    DebtorGroupModel? debtorGroups,
  }) =>
      DebtorGroupGetSuccess(debtorGroups: debtorGroups ?? this.debtorGroups);

  @override
  List<Object> get props => [debtorGroups];
}

class DebtorGroupGetFailed extends DebtorGroupState {
  final String message;

  const DebtorGroupGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DebtorGroupUpdateInitial extends DebtorGroupState {}

class DebtorGroupUpdateInProgress extends DebtorGroupState {}

class DebtorGroupUpdateSuccess extends DebtorGroupState {}

class DebtorGroupUpdateFailed extends DebtorGroupState {
  final String message;

  const DebtorGroupUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
