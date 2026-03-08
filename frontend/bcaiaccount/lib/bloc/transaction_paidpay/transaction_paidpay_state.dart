part of 'transaction_paidpay_bloc.dart';

abstract class TransactionPaidPayState extends Equatable {
  const TransactionPaidPayState();

  @override
  List<Object> get props => [];
}

class TransactionPaidPayInitial extends TransactionPaidPayState {}

class TransactionPaidPayInProgress extends TransactionPaidPayState {}

class TransactionPaidPayLoadSuccess extends TransactionPaidPayState {
  final List<TransactionPaidPayModel> transactionPaidPay;

  const TransactionPaidPayLoadSuccess({required this.transactionPaidPay});

  @override
  List<Object> get props => [transactionPaidPay];
}

class TransactionPaidPayLoadFailed extends TransactionPaidPayState {
  final String message;

  const TransactionPaidPayLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransactionPaidPaySaveInitial extends TransactionPaidPayState {}

class TransactionPaidPaySaveInProgress extends TransactionPaidPayState {}

class TransactionPaidPaySaveSuccess extends TransactionPaidPayState {
  final String docno;
  const TransactionPaidPaySaveSuccess({
    required this.docno,
  });

  @override
  List<Object> get props => [docno];
}

class TransactionPaidPaySaveFailed extends TransactionPaidPayState {
  final String message;

  const TransactionPaidPaySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransactionPaidPayDeleteFailed extends TransactionPaidPayState {
  final String message;

  const TransactionPaidPayDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransactionPaidPayDeleteInProgress extends TransactionPaidPayState {}

class TransactionPaidPayDeleteSuccess extends TransactionPaidPayState {}

class TransactionPaidPayDeleteManyInProgress extends TransactionPaidPayState {}

class TransactionPaidPayDeleteManySuccess extends TransactionPaidPayState {}

class TransactionPaidPayDeleteManyFailed extends TransactionPaidPayState {}

class TransactionPaidPayGetInProgress extends TransactionPaidPayState {}

class TransactionPaidPayGetSuccess extends TransactionPaidPayState {
  final TransactionPaidPayModel transactionPaidPay;

  const TransactionPaidPayGetSuccess({required this.transactionPaidPay});

  TransactionPaidPayGetSuccess copyWith({
    TransactionPaidPayModel? transactionPaidPay,
  }) =>
      TransactionPaidPayGetSuccess(
          transactionPaidPay: transactionPaidPay ?? this.transactionPaidPay);

  @override
  List<Object> get props => [transactionPaidPay];
}

class TransactionPaidPayGetFailed extends TransactionPaidPayState {
  final String message;

  const TransactionPaidPayGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransactionPaidPayUpdateInitial extends TransactionPaidPayState {}

class TransactionPaidPayUpdateInProgress extends TransactionPaidPayState {}

class TransactionPaidPayUpdateSuccess extends TransactionPaidPayState {}

class TransactionPaidPayUpdateFailed extends TransactionPaidPayState {
  final String message;

  const TransactionPaidPayUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class GetCustcodeTransactionInProgress extends TransactionPaidPayState {}

class GetCustcodeTransactionSuccess extends TransactionPaidPayState {
  final List<GetCustcodeTransationModel> getCustcodeTransationModel;

  const GetCustcodeTransactionSuccess(
      {required this.getCustcodeTransationModel});

  @override
  List<Object> get props => [getCustcodeTransationModel];
}

class GetCustcodeTransactionFailed extends TransactionPaidPayState {
  final String message;

  const GetCustcodeTransactionFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
