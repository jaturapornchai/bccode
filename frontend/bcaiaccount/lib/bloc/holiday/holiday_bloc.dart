import 'dart:convert';

import 'package:smlaicloud/model/holiday_model.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'holiday_event.dart';
part 'holiday_state.dart';

class HolidayBloc extends Bloc<HolidayEvent, HolidayState> {
  final JsonRepository _jsonRepository;

  HolidayBloc({required JsonRepository jsonRepository})
      : _jsonRepository = jsonRepository,
        super(HolidayInitial()) {
    on<HolidayLoadList>(onHolidayLoad);
    on<HolidaySave>(onHolidaySave);
    on<HolidayUpdate>(onHolidayUpdate);
    on<HolidayDelete>(onHolidayDelete);
    on<HolidayDeleteMany>(onHolidayDeleteMany);
  }

  void onHolidayLoad(HolidayLoadList event, Emitter<HolidayState> emit) async {
    emit(HolidayInProgress());

    try {
      final results = await _jsonRepository.getSetting('Holiday', event.search);

      if (results.success) {
        if (results.data.length > 0) {
          List<HolidayModel> holidayList = [];

          for (int i = 0; i < results.data.length; i++) {
            HolidayModel holidayModel =
                HolidayModel.fromJson(json.decode(results.data[i]['body']));
            holidayModel.guidfixed = results.data[i]['guidfixed'];
            holidayList.add(holidayModel);
          }

          emit(HolidayLoadSuccess(holidays: holidayList));
        } else {
          emit(const HolidayLoadFailed(message: 'Holiday No Data Found'));
        }
      } else {
        emit(const HolidayLoadFailed(message: 'Holiday Not Found'));
      }
    } catch (e) {
      emit(HolidayLoadFailed(message: e.toString()));
    }
  }

  void onHolidaySave(HolidaySave event, Emitter<HolidayState> emit) async {
    emit(HolidaySaveInProgress());
    try {
      final postData = {
        "code": 'Holiday',
        "body": jsonEncode(event.holidayModel)
      };
      await _jsonRepository.saveSetting(postData);
      emit(HolidaySaveSuccess());
    } catch (e) {
      emit(HolidaySaveFailed(message: e.toString()));
    }
  }

  void onHolidayUpdate(HolidayUpdate event, Emitter<HolidayState> emit) async {
    emit(HolidayUpdateInProgress());
    try {
      final postData = {
        "code": 'Holiday',
        "body": jsonEncode(event.holidayModel)
      };

      await _jsonRepository.updateSetting(event.guid, postData);
      emit(HolidayUpdateSuccess());
    } catch (e) {
      emit(HolidayUpdateFailed(message: e.toString()));
    }
  }

  void onHolidayDelete(HolidayDelete event, Emitter<HolidayState> emit) async {
    emit(HolidayDeleteInProgress());
    try {
      await _jsonRepository.deleteSetting(event.guid);

      emit(HolidayDeleteSuccess());
    } catch (e) {
      emit(HolidayDeleteFailed());
    }
  }

  void onHolidayDeleteMany(
      HolidayDeleteMany event, Emitter<HolidayState> emit) async {
    emit(HolidayDeleteManyInProgress());
    try {
      await _jsonRepository.deleteManySetting(event.guid);

      emit(HolidayDeleteManySuccess());
    } catch (e) {
      emit(HolidayDeleteManyFailed());
    }
  }
}
