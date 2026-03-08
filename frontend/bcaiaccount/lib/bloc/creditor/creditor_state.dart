part of 'creditor_bloc.dart';

abstract class CreditorState extends Equatable {
  const CreditorState();

  @override
  List<Object> get props => [];
}

class CreditorInitial extends CreditorState {}

class CreditorInProgress extends CreditorState {}

class CreditorLoadSuccess extends CreditorState {
  final List<CreditorModel> creditors;

  const CreditorLoadSuccess({required this.creditors});

  CreditorLoadSuccess copyWith({
    List<CreditorModel>? creditors,
  }) =>
      CreditorLoadSuccess(creditors: creditors ?? this.creditors);

  @override
  List<Object> get props => [creditors];
}

class CreditorLoadFailed extends CreditorState {
  final String message;

  const CreditorLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorSaveInitial extends CreditorState {}

class CreditorSaveInProgress extends CreditorState {}

class CreditorSaveSuccess extends CreditorState {}

class CreditorSaveFailed extends CreditorState {
  final String message;

  const CreditorSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorDeleteInProgress extends CreditorState {}

class CreditorDeleteSuccess extends CreditorState {}

class CreditorDeleteFailed extends CreditorState {}

class CreditorDeleteManyInProgress extends CreditorState {}

class CreditorDeleteManySuccess extends CreditorState {}

class CreditorDeleteManyFailed extends CreditorState {}

class CreditorGetInProgress extends CreditorState {}

class CreditorGetSuccess extends CreditorState {
  final CreditorModel creditors;

  const CreditorGetSuccess({required this.creditors});

  CreditorGetSuccess copyWith({
    CreditorModel? creditors,
  }) =>
      CreditorGetSuccess(creditors: creditors ?? this.creditors);

  @override
  List<Object> get props => [creditors];
}

class CreditorGetFailed extends CreditorState {
  final String message;

  const CreditorGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorUpdateInitial extends CreditorState {}

class CreditorUpdateInProgress extends CreditorState {}

class CreditorUpdateSuccess extends CreditorState {}

class CreditorUpdateFailed extends CreditorState {
  final String message;

  const CreditorUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorGetBycodeInProgress extends CreditorState {}

class CreditorGetBycodeSuccess extends CreditorState {
  final CreditorModel creditors;

  const CreditorGetBycodeSuccess({required this.creditors});

  CreditorGetBycodeSuccess copyWith({
    CreditorModel? creditors,
  }) =>
      CreditorGetBycodeSuccess(creditors: creditors ?? this.creditors);

  @override
  List<Object> get props => [creditors];
}

class CreditorGetBycodeFailed extends CreditorState {
  final String message;

  const CreditorGetBycodeFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CreditorBulkSaveInProgress extends CreditorState {}

class CreditorBulkSaveSuccess extends CreditorState {
  final int savedCount;

  const CreditorBulkSaveSuccess({required this.savedCount});

  @override
  List<Object> get props => [savedCount];
}

class CreditorBulkSaveFailed extends CreditorState {
  final String message;

  const CreditorBulkSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
