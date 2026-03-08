part of 'creditor_group_bloc.dart';

abstract class CreditorGroupState extends Equatable {
  const CreditorGroupState();

  @override
  List<Object> get props => [];
}

class CreditorGroupInitial extends CreditorGroupState {}

class CreditorGroupInProgress extends CreditorGroupState {}

class CreditorGroupLoadSuccess extends CreditorGroupState {
  final List<CreditorGroupModel> creditorGroups;

  const CreditorGroupLoadSuccess({required this.creditorGroups});

  CreditorGroupLoadSuccess copyWith({
    List<CreditorGroupModel>? creditorGroups,
  }) =>
      CreditorGroupLoadSuccess(
          creditorGroups: creditorGroups ?? this.creditorGroups);

  @override
  List<Object> get props => [creditorGroups];
}

class CreditorGroupLoadFailed extends CreditorGroupState {
  final String message;

  const CreditorGroupLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorGroupSaveInitial extends CreditorGroupState {}

class CreditorGroupSaveInProgress extends CreditorGroupState {}

class CreditorGroupSaveSuccess extends CreditorGroupState {}

class CreditorGroupSaveFailed extends CreditorGroupState {
  final String message;

  const CreditorGroupSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorGroupDeleteInProgress extends CreditorGroupState {}

class CreditorGroupDeleteSuccess extends CreditorGroupState {}

class CreditorGroupDeleteFailed extends CreditorGroupState {}

class CreditorGroupDeleteManyInProgress extends CreditorGroupState {}

class CreditorGroupDeleteManySuccess extends CreditorGroupState {}

class CreditorGroupDeleteManyFailed extends CreditorGroupState {}

class CreditorGroupGetInProgress extends CreditorGroupState {}

class CreditorGroupGetSuccess extends CreditorGroupState {
  final CreditorGroupModel creditorGroup;

  const CreditorGroupGetSuccess({required this.creditorGroup});

  CreditorGroupGetSuccess copyWith({
    CreditorGroupModel? creditorGroup,
  }) =>
      CreditorGroupGetSuccess(
          creditorGroup: creditorGroup ?? this.creditorGroup);

  @override
  List<Object> get props => [creditorGroup];
}

class CreditorGroupGetFailed extends CreditorGroupState {
  final String message;

  const CreditorGroupGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorGroupUpdateInitial extends CreditorGroupState {}

class CreditorGroupUpdateInProgress extends CreditorGroupState {}

class CreditorGroupUpdateSuccess extends CreditorGroupState {}

class CreditorGroupUpdateFailed extends CreditorGroupState {
  final String message;

  const CreditorGroupUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
