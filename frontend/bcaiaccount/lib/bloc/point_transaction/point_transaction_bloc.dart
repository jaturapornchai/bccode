import 'package:smlaicloud/model/point_transaction_model.dart';
import 'package:smlaicloud/repositories/point_transaction_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'point_transaction_event.dart';
part 'point_transaction_state.dart';

class PointTransactionBloc
    extends Bloc<PointTransactionEvent, PointTransactionState> {
  final PointTransactionRepository _pointTransactionRepository;

  PointTransactionBloc({
    required PointTransactionRepository pointTransactionRepository,
  }) : _pointTransactionRepository = pointTransactionRepository,
       super(PointTransactionInitial()) {
    on<PointTransactionLoadByDebtorCode>(onPointTransactionLoadByDebtorCode);
    on<PointTransactionDeleteManual>(onPointTransactionDeleteManual);
  }

  void onPointTransactionLoadByDebtorCode(
    PointTransactionLoadByDebtorCode event,
    Emitter<PointTransactionState> emit,
  ) async {
    emit(PointTransactionInProgress());

    try {
      final results = await _pointTransactionRepository
          .getPointTransactionsByDebtorCode(event.debtorCode);

      if (results.success) {
        List<PointTransactionModel> transactions = (results.data as List)
            .map((transaction) => PointTransactionModel.fromJson(transaction))
            .toList();
        emit(PointTransactionLoadSuccess(transactions: transactions));
      } else {
        emit(
          const PointTransactionLoadFailed(
            message: 'Point Transactions Not Found',
          ),
        );
      }
    } catch (e) {
      emit(PointTransactionLoadFailed(message: e.toString()));
    }
  }

  void onPointTransactionDeleteManual(
    PointTransactionDeleteManual event,
    Emitter<PointTransactionState> emit,
  ) async {
    emit(PointTransactionDeleteInProgress());

    try {
      final result = await _pointTransactionRepository
          .deleteManualPointTransaction(event.pointscode, event.docno);

      if (result.success) {
        emit(
          PointTransactionDeleteSuccess(
            message: result.message.isNotEmpty
                ? result.message
                : 'Manual point transaction deleted successfully',
          ),
        );
      } else {
        emit(
          PointTransactionDeleteFailed(
            message: result.message.isNotEmpty
                ? result.message
                : 'Failed to delete manual point transaction',
          ),
        );
      }
    } catch (e) {
      emit(PointTransactionDeleteFailed(message: e.toString()));
    }
  }
}
