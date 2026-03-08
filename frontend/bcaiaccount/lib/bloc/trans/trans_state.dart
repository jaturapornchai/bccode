part of 'trans_bloc.dart';

abstract class TransState extends Equatable {
  const TransState();

  @override
  List<Object> get props => [];
}

class TransInitial extends TransState {}

class TransInProgress extends TransState {}

class TransLoadSuccess extends TransState {
  final List<TransactionModel> trans;
  final int totalRecord;

  const TransLoadSuccess({required this.totalRecord, required this.trans});

  @override
  List<Object> get props => [totalRecord, trans];
}

class TransLoadFailed extends TransState {
  final String message;

  const TransLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransSaveInitial extends TransState {}

class TransSaveInProgress extends TransState {}

class TransSaveSuccess extends TransState {
  final String docno;
  const TransSaveSuccess({
    required this.docno,
  });

  @override
  List<Object> get props => [docno];
}

class TransSaveFailed extends TransState {
  final String message;

  const TransSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransDeleteFailed extends TransState {
  final String message;

  const TransDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransDeleteInProgress extends TransState {}

class TransDeleteSuccess extends TransState {}

class TransDeleteManyInProgress extends TransState {}

class TransDeleteManySuccess extends TransState {}

class TransDeleteManyFailed extends TransState {}

class TransGetInProgress extends TransState {}

class TransGetSuccess extends TransState {
  final TransactionModel trans;

  const TransGetSuccess({required this.trans});

  TransGetSuccess copyWith({
    TransactionModel? trans,
  }) =>
      TransGetSuccess(trans: trans ?? this.trans);

  @override
  List<Object> get props => [trans];
}

class TransGetFailed extends TransState {
  final String message;

  const TransGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransUpdateInitial extends TransState {}

class TransUpdateInProgress extends TransState {}

class TransUpdateSuccess extends TransState {}

class TransUpdateFailed extends TransState {
  final String message;

  const TransUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransFullInvoiceInProgress extends TransState {}

class TransFullInvoiceSuccess extends TransState {
  final String docno;
  const TransFullInvoiceSuccess({
    required this.docno,
  });

  @override
  List<Object> get props => [docno];
}

class TransFullInvoiceFailed extends TransState {
  final String message;

  const TransFullInvoiceFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
