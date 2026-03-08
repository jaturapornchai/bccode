part of 'chart_account_bloc.dart';

abstract class ChartAccountState extends Equatable {
  const ChartAccountState();

  @override
  List<Object> get props => [];
}

class ChartAccountInitial extends ChartAccountState {}

class ChartAccountInProgress extends ChartAccountState {}

class ChartAccountLoadSuccess extends ChartAccountState {
  final List<AccountChartModel> chartAccounts;

  const ChartAccountLoadSuccess({required this.chartAccounts});

  ChartAccountLoadSuccess copyWith({
    List<AccountChartModel>? chartAccounts,
  }) =>
      ChartAccountLoadSuccess(
          chartAccounts: chartAccounts ?? this.chartAccounts);

  @override
  List<Object> get props => [chartAccounts];
}

class ChartAccountLoadFailed extends ChartAccountState {
  final String message;

  const ChartAccountLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
