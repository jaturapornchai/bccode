import 'package:smlaicloud/model/accountbook_model.dart';
import 'package:smlaicloud/model/accountchart_model.dart';
import 'package:smlaicloud/model/accountgroup_model.dart';
import 'package:smlaicloud/repositories/chart_account_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'chart_account_event.dart';
part 'chart_account_state.dart';

class ChartAccountBloc extends Bloc<ChartAccountEvent, ChartAccountState> {
  final ChartAccountRepository _chartAccountRepository;
  ChartAccountBloc({required ChartAccountRepository chartAccountRepository})
      : _chartAccountRepository = chartAccountRepository,
        super(ChartAccountInitial()) {
    on<ChartAccountLoad>(onChartAccountLoad);
    on<AccountGroupLoad>(onAccountGroup);
    on<AccountBookLoad>(onAccountBook);
  }

  void onChartAccountLoad(
      ChartAccountLoad event, Emitter<ChartAccountState> emit) async {
    emit(ChartAccountInProgress());

    try {
      final results =
          await _chartAccountRepository.getChartAccount(search: event.search);

      if (results.success) {
        List<AccountChartModel> chartAccounts = (results.data as List)
            .map((chartAccounts) => AccountChartModel.fromJson(chartAccounts))
            .toList();
        emit(ChartAccountLoadSuccess(chartAccounts: chartAccounts));
      } else {
        emit(const ChartAccountLoadFailed(message: 'Chart Account Not Found'));
      }
    } catch (e) {
      emit(ChartAccountLoadFailed(message: e.toString()));
    }
  }

  void onAccountGroup(
      AccountGroupLoad event, Emitter<ChartAccountState> emit) async {
    emit(GroupAccountInProgress());

    try {
      final results =
          await _chartAccountRepository.getAccountGroup(search: event.search);

      if (results.success) {
        List<AccountGroupModel> groupAccounts = (results.data as List)
            .map((groupAccounts) => AccountGroupModel.fromJson(groupAccounts))
            .toList();
        emit(GroupAccountLoadSuccess(groupAccounts: groupAccounts));
      } else {
        emit(const GroupAccountLoadFailed(message: 'Group Account Not Found'));
      }
    } catch (e) {
      emit(GroupAccountLoadFailed(message: e.toString()));
    }
  }

  void onAccountBook(
      AccountBookLoad event, Emitter<ChartAccountState> emit) async {
    emit(BookAccountInProgress());

    try {
      final results =
          await _chartAccountRepository.getAccountBook(search: event.search);

      if (results.success) {
        List<AccountBookModel> bookAccounts = (results.data as List)
            .map((bookAccounts) => AccountBookModel.fromJson(bookAccounts))
            .toList();
        emit(BookAccountLoadSuccess(bookAccounts: bookAccounts));
      } else {
        emit(const BookAccountLoadFailed(message: 'Book Account Not Found'));
      }
    } catch (e) {
      emit(BookAccountLoadFailed(message: e.toString()));
    }
  }
}
