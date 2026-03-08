part of 'point_transaction_bloc.dart';

abstract class PointTransactionEvent extends Equatable {
  const PointTransactionEvent();

  @override
  List<Object> get props => [];
}

class PointTransactionLoadByDebtorCode extends PointTransactionEvent {
  final String debtorCode;

  const PointTransactionLoadByDebtorCode({required this.debtorCode});

  @override
  List<Object> get props => [debtorCode];
}

class PointTransactionDeleteManual extends PointTransactionEvent {
  final String pointscode;
  final String docno;

  const PointTransactionDeleteManual({
    required this.pointscode,
    required this.docno,
  });

  @override
  List<Object> get props => [pointscode, docno];
}
