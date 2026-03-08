import 'dart:convert';

import 'package:smlaicloud/model/work_day_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/json_repository.dart';

part 'work_day_event.dart';
part 'work_day_state.dart';

class WorkDayBloc extends Bloc<WorkDayEvent, WorkDayState> {
  final JsonRepository _jsonRepository;

  WorkDayBloc({required JsonRepository jsonRepository})
      : _jsonRepository = jsonRepository,
        super(WorkDayInitial()) {
    on<WorkDayLoad>(onWorkDayLoad);
    on<WorkDaySave>(onWorkDaySave);
    on<WorkDayUpdate>(onWorkDayUpdate);
  }

  void onWorkDayLoad(WorkDayLoad event, Emitter<WorkDayState> emit) async {
    emit(WorkDayInProgress());

    try {
      final results = await _jsonRepository.getSetting("workDay", "");

      if (results.success) {
        if (results.data.length > 0) {
          WorkDayListModel workDay =
              WorkDayListModel.fromJson(json.decode(results.data[0]['body']));

          emit(WorkDayLoadSuccess(
              guidfixed: results.data[0]['guidfixed'],
              workDay: workDay.workdays));
        } else {
          emit(const WorkDayLoadFailed(message: 'WorkDay Group Not Found'));
        }
      } else {
        emit(const WorkDayLoadFailed(message: 'WorkDay Not Found'));
      }
    } catch (e) {
      emit(WorkDayLoadFailed(message: e.toString()));
    }
  }

  void onWorkDaySave(WorkDaySave event, Emitter<WorkDayState> emit) async {
    emit(WorkDaySaveInProgress());
    try {
      final data = event.workDays.toJson();

      final postData = {"code": 'workDay', "body": jsonEncode(data)};
      await _jsonRepository.saveSetting(postData);
      emit(WorkDaySaveSuccess());
    } catch (e) {
      emit(WorkDaySaveFailed(message: e.toString()));
    }
  }

  void onWorkDayUpdate(WorkDayUpdate event, Emitter<WorkDayState> emit) async {
    emit(WorkDayUpdateInProgress());
    try {
      final data = event.workDays.toJson();

      final postData = {"code": 'workDay', "body": jsonEncode(data)};

      await _jsonRepository.updateSetting(event.guid, postData);
      emit(WorkDayUpdateSuccess());
    } catch (e) {
      emit(WorkDayUpdateFailed(message: e.toString()));
    }
  }
}
