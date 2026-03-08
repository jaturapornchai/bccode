import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:smlaicloud/model/config_model.dart';

part 'config_event.dart';
part 'config_state.dart';

class ConfigSystemBloc extends Bloc<ConfigSystemEvent, ConfigSystemState> {
  final JsonRepository jsonRepo;

  ConfigSystemBloc({required JsonRepository jsonRepository})
      : jsonRepo = jsonRepository,
        super(ConfigSystemInitial()) {
    on<ConfigSystemLoad>(onConfigSystemLoad);
    on<ConfigSystemSave>(onConfigSystemSave);
    on<ConfigSystemUpdate>(onConfigSystemUpdate);
  }

  void onConfigSystemLoad(
      ConfigSystemLoad event, Emitter<ConfigSystemState> emit) async {
    emit(ConfigSystemInProgress());

    try {
      final results = await jsonRepo.getSetting("ConfigSystem", "");

      if (results.success) {
        if (results.data.length > 0) {
          ConfigSystemModel configSystem =
              ConfigSystemModel.fromJson(json.decode(results.data[0]['body']));

          emit(ConfigSystemLoadSuccess(
              guidFixed: results.data[0]['guidfixed'], data: configSystem));
        } else {
          emit(const ConfigSystemLoadFailed(
              message: 'ConfigSystem Group Not Found'));
        }
      } else {
        emit(const ConfigSystemLoadFailed(message: 'ConfigSystem Not Found'));
      }
    } catch (e) {
      emit(ConfigSystemLoadFailed(message: e.toString()));
    }
  }

  void onConfigSystemSave(
      ConfigSystemSave event, Emitter<ConfigSystemState> emit) async {
    emit(ConfigSystemSaveInProgress());
    try {
      final data = event.data.toJson();

      final postData = {"code": 'ConfigSystem', "body": jsonEncode(data)};
      await jsonRepo.saveSetting(postData);
      emit(ConfigSystemSaveSuccess());
    } catch (e) {
      emit(ConfigSystemSaveFailed(message: e.toString()));
    }
  }

  void onConfigSystemUpdate(
      ConfigSystemUpdate event, Emitter<ConfigSystemState> emit) async {
    emit(ConfigSystemUpdateInProgress());
    try {
      final data = event.data.toJson();

      final postData = {"code": 'ConfigSystem', "body": jsonEncode(data)};

      await jsonRepo.updateSetting(event.guid, postData);
      emit(ConfigSystemUpdateSuccess());
    } catch (e) {
      emit(ConfigSystemUpdateFailed(message: e.toString()));
    }
  }
}
