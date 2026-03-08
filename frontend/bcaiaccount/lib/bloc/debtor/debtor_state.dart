part of 'debtor_bloc.dart';

abstract class DebtorState extends Equatable {
  const DebtorState();

  @override
  List<Object> get props => [];
}

class DebtorInitial extends DebtorState {}

class DebtorInProgress extends DebtorState {}

class DebtorLoadSuccess extends DebtorState {
  final List<DebtorModel> debtors;

  const DebtorLoadSuccess({required this.debtors});

  DebtorLoadSuccess copyWith({List<DebtorModel>? debtors}) =>
      DebtorLoadSuccess(debtors: debtors ?? this.debtors);

  @override
  List<Object> get props => [debtors];
}

class DebtorLoadFailed extends DebtorState {
  final String message;

  const DebtorLoadFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorSaveInitial extends DebtorState {}

class DebtorSaveInProgress extends DebtorState {}

class DebtorSaveSuccess extends DebtorState {}

class DebtorSaveFailed extends DebtorState {
  final String message;

  const DebtorSaveFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorDeleteInProgress extends DebtorState {}

class DebtorDeleteSuccess extends DebtorState {}

class DebtorDeleteFailed extends DebtorState {}

class DebtorDeleteManyInProgress extends DebtorState {}

class DebtorDeleteManySuccess extends DebtorState {}

class DebtorDeleteManyFailed extends DebtorState {}

class DebtorGetInProgress extends DebtorState {}

class DebtorGetSuccess extends DebtorState {
  final DebtorModel debtors;

  const DebtorGetSuccess({required this.debtors});

  DebtorGetSuccess copyWith({DebtorModel? debtors}) =>
      DebtorGetSuccess(debtors: debtors ?? this.debtors);

  @override
  List<Object> get props => [debtors];
}

class DebtorGetFailed extends DebtorState {
  final String message;

  const DebtorGetFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorUpdateInitial extends DebtorState {}

class DebtorUpdateInProgress extends DebtorState {}

class DebtorUpdateSuccess extends DebtorState {}

class DebtorUpdateFailed extends DebtorState {
  final String message;

  const DebtorUpdateFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorGetBycodeInProgress extends DebtorState {}

class DebtorGetBycodeSuccess extends DebtorState {
  final DebtorModel debtors;

  const DebtorGetBycodeSuccess({required this.debtors});

  DebtorGetBycodeSuccess copyWith({DebtorModel? debtors}) =>
      DebtorGetBycodeSuccess(debtors: debtors ?? this.debtors);

  @override
  List<Object> get props => [debtors];
}

class DebtorGetBycodeFailed extends DebtorState {
  final String message;

  const DebtorGetBycodeFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorBulkSaveInProgress extends DebtorState {}

class DebtorBulkSaveSuccess extends DebtorState {
  final int savedCount;

  const DebtorBulkSaveSuccess({required this.savedCount});

  @override
  List<Object> get props => [savedCount];
}

class DebtorBulkSaveFailed extends DebtorState {
  final String message;

  const DebtorBulkSaveFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorAddPointInProgress extends DebtorState {}

class DebtorAddPointSuccess extends DebtorState {
  final String message;

  DebtorAddPointSuccess({String? message}) : message = message ?? global.language('add_point_success');

  @override
  List<Object> get props => [message];
}

class DebtorAddPointFailed extends DebtorState {
  final String message;

  const DebtorAddPointFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DebtorBulkAddPointsInProgress extends DebtorState {}

class DebtorBulkAddPointsSuccess extends DebtorState {
  final int successCount;
  final int failedCount;
  final int totalCount;
  final List<String>? successItems;
  final List<Map<String, dynamic>>? failedItems;

  const DebtorBulkAddPointsSuccess({
    required this.successCount,
    required this.failedCount,
    required this.totalCount,
    this.successItems,
    this.failedItems,
  });

  @override
  List<Object> get props => [successCount, failedCount, totalCount];
}

class DebtorBulkAddPointsFailed extends DebtorState {
  final String message;

  const DebtorBulkAddPointsFailed({required this.message});

  @override
  List<Object> get props => [message];
}
