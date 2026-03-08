import 'package:smlaicloud/model/line_notify_model.dart';
import 'package:smlaicloud/repositories/line_notify_repository.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

part 'line_notify_event.dart';
part 'line_notify_state.dart';

class LineNotifyBloc extends Bloc<LineNotifyEvent, LineNotifyState> {
  final LineNotifyRepository _lineNotifyRepository;
  LineNotifyBloc({required LineNotifyRepository lineNotifyRepository})
      : _lineNotifyRepository = lineNotifyRepository,
        super(LineNotifyInitial()) {
    on<LineNotifyLoadList>(onLineNotifyLoad);
    on<LineNotifySave>(onLineNotifySave);
    on<LineNotifyUpdate>(onLineNotifyUpdate);
    on<LineNotifyDelete>(onLineNotifyDelete);
    on<LineNotifyDeleteMany>(onLineNotifyDeleteMany);
    on<LineNotifyGet>(onLineNotifyGet);
    on<LineNotifyTest>(onLineNotifyTest);
  }

  void onLineNotifyLoad(
      LineNotifyLoadList event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifyInProgress());

    try {
      final results = await _lineNotifyRepository.getLineNotifyList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<LineNotifyModel> lineNotifys = (results.data as List)
            .map((bookLineNotify) => LineNotifyModel.fromJson(bookLineNotify))
            .toList();
        emit(LineNotifyLoadSuccess(lineNotifys: lineNotifys));
      } else {
        emit(const LineNotifyLoadFailed(message: 'LineNotify Not Found'));
      }
    } catch (e) {
      emit(LineNotifyLoadFailed(message: e.toString()));
    }
  }

  void onLineNotifyDelete(
      LineNotifyDelete event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifyDeleteInProgress());
    try {
      await _lineNotifyRepository.deleteLineNotify(event.guid);

      emit(LineNotifyDeleteSuccess());
    } catch (e) {
      emit(LineNotifyDeleteFailed(message: e.toString()));
    }
  }

  void onLineNotifyDeleteMany(
      LineNotifyDeleteMany event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifyDeleteManyInProgress());
    try {
      await _lineNotifyRepository.deleteLineNotifyMany(event.guid);

      emit(LineNotifyDeleteManySuccess());
    } catch (e) {
      emit(LineNotifyDeleteFailed(message: e.toString()));
    }
  }

  void onLineNotifySave(
      LineNotifySave event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifySaveInProgress());
    try {
      await _lineNotifyRepository.saveLineNotify(event.lineNotify);
      emit(LineNotifySaveSuccess());
    } catch (e) {
      emit(LineNotifySaveFailed(message: e.toString()));
    }
  }

  void onLineNotifyUpdate(
      LineNotifyUpdate event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifyUpdateInProgress());
    try {
      await _lineNotifyRepository.updateLineNotify(
          event.guid, event.lineNotify);
      emit(LineNotifyUpdateSuccess());
    } catch (e) {
      emit(LineNotifyUpdateFailed(message: e.toString()));
    }
  }

  void onLineNotifyGet(
      LineNotifyGet event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifyGetInProgress());
    try {
      final result = await _lineNotifyRepository.getLineNotify(event.guid);
      if (result.success) {
        LineNotifyModel lineNotify = LineNotifyModel.fromJson(result.data);
        emit(LineNotifyGetSuccess(lineNotify: lineNotify));
      } else {
        emit(const LineNotifyGetFailed(message: 'Book Bank Not Found'));
      }
    } catch (e) {
      emit(LineNotifyDeleteFailed(message: e.toString()));
    }
  }

  /// test line notify
  void onLineNotifyTest(
      LineNotifyTest event, Emitter<LineNotifyState> emit) async {
    emit(LineNotifyTestInProgress());
    try {
      final result = await _lineNotifyRepository.testLineNotify(
          event.token, event.message);
      if (result.success) {
        emit(LineNotifyTestSuccess());
      } else {
        emit(const LineNotifyTestFailed(message: 'Test Line Notify Failed'));
      }
    } catch (e) {
      emit(LineNotifyTestFailed(message: e.toString()));
    }
  }
}
