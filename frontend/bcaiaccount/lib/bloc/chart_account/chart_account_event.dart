part of 'chart_account_bloc.dart';

abstract class ChartAccountEvent extends Equatable {
  const ChartAccountEvent();

  @override
  List<Object> get props => [];
}

class ChartAccountLoad extends ChartAccountEvent {
  final String search;

  const ChartAccountLoad({required this.search});

  @override
  List<Object> get props => [];
}

class AccountGroupLoad extends ChartAccountEvent {
  final String search;

  const AccountGroupLoad({required this.search});

  @override
  List<Object> get props => [];
}

class AccountBookLoad extends ChartAccountEvent {
  final String search;

  const AccountBookLoad({required this.search});

  @override
  List<Object> get props => [];
}

class GroupAccountInProgress extends ChartAccountState {}

class GroupAccountLoadSuccess extends ChartAccountState {
  final List<AccountGroupModel> groupAccounts;

  const GroupAccountLoadSuccess({required this.groupAccounts});

  GroupAccountLoadSuccess copyWith({
    List<AccountGroupModel>? groupAccounts,
  }) =>
      GroupAccountLoadSuccess(
          groupAccounts: groupAccounts ?? this.groupAccounts);

  @override
  List<Object> get props => [groupAccounts];
}

class GroupAccountLoadFailed extends ChartAccountState {
  final String message;

  const GroupAccountLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BookAccountInProgress extends ChartAccountState {}

class BookAccountLoadSuccess extends ChartAccountState {
  final List<AccountBookModel> bookAccounts;

  const BookAccountLoadSuccess({required this.bookAccounts});

  BookAccountLoadSuccess copyWith({
    List<AccountBookModel>? bookAccounts,
  }) =>
      BookAccountLoadSuccess(bookAccounts: bookAccounts ?? this.bookAccounts);

  @override
  List<Object> get props => [bookAccounts];
}

class BookAccountLoadFailed extends ChartAccountState {
  final String message;

  const BookAccountLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
