part of 'transaction_paidpay_bloc.dart';

abstract class TransactionPaidPayEvent extends Equatable {
  const TransactionPaidPayEvent();

  @override
  List<Object> get props => [];
}

class TransactionPaidPayGet extends TransactionPaidPayEvent {
  final String guid;

  const TransactionPaidPayGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class TransactionPaidPayLoad extends TransactionPaidPayEvent {
  final int limit;
  final int offset;
  final String search;
  final TransactionTypeEnum type;
  const TransactionPaidPayLoad(
      {required this.offset,
      required this.limit,
      required this.search,
      required this.type});

  @override
  List<Object> get props => [];
}

class TransactionPaidPayDelete extends TransactionPaidPayEvent {
  final String guid;
  final TransactionTypeEnum type;
  const TransactionPaidPayDelete({
    required this.guid,
    required this.type,
  });

  @override
  List<Object> get props => [guid, type];
}

class TransactionPaidPayDeleteMany extends TransactionPaidPayEvent {
  final List<String> guid;

  const TransactionPaidPayDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class TransactionPaidPaySave extends TransactionPaidPayEvent {
  final TransactionPaidPayModel transactionPaidPay;
  final TransactionTypeEnum type;
  const TransactionPaidPaySave({
    required this.transactionPaidPay,
    required this.type,
  });

  @override
  List<Object> get props => [transactionPaidPay, type];
}

class TransactionPaidPayUpdate extends TransactionPaidPayEvent {
  final String guid;
  final TransactionPaidPayModel transactionPaidPay;
  final TransactionTypeEnum type;
  const TransactionPaidPayUpdate({
    required this.guid,
    required this.transactionPaidPay,
    required this.type,
  });

  @override
  List<Object> get props => [guid, transactionPaidPay, type];
}

class GetCustcodeTransaction extends TransactionPaidPayEvent {
  final String custcode;
  final TransactionTypeEnum type;
  const GetCustcodeTransaction({
    required this.custcode,
    required this.type,
  });

  @override
  List<Object> get props => [custcode, type];
}
