part of 'point_transaction_bloc.dart';

abstract class PointTransactionState extends Equatable {
  const PointTransactionState();

  @override
  List<Object> get props => [];
}

class PointTransactionInitial extends PointTransactionState {}

class PointTransactionInProgress extends PointTransactionState {}

class PointTransactionLoadSuccess extends PointTransactionState {
  final List<PointTransactionModel> transactions;

  const PointTransactionLoadSuccess({required this.transactions});

  PointTransactionLoadSuccess copyWith({
    List<PointTransactionModel>? transactions,
  }) => PointTransactionLoadSuccess(
    transactions: transactions ?? this.transactions,
  );

  @override
  List<Object> get props => [transactions];
}

class PointTransactionLoadFailed extends PointTransactionState {
  final String message;

  const PointTransactionLoadFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class PointTransactionDeleteInProgress extends PointTransactionState {}

class PointTransactionDeleteSuccess extends PointTransactionState {
  final String message;

  const PointTransactionDeleteSuccess({
    this.message = 'Manual point transaction deleted successfully',
  });

  @override
  List<Object> get props => [message];
}

class PointTransactionDeleteFailed extends PointTransactionState {
  final String message;

  const PointTransactionDeleteFailed({required this.message});

  @override
  List<Object> get props => [message];
}
