import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/model/pos_setting_model.dart';
import 'package:smlaicloud/repositories/pos_setting_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'pos_setting_event.dart';
part 'pos_setting_state.dart';

class PosSettingBloc extends Bloc<PosSettingEvent, PosSettingState> {
  final PosSettingRepository _posSettingRepository;

  PosSettingBloc({required PosSettingRepository posSettingRepository})
      : _posSettingRepository = posSettingRepository,
        super(PosSettingInitial()) {
    on<PosSettingLoadList>(onPosSettingLoad);
    on<PosSettingSave>(onPosSettingSave);
    on<PosSettingUpdate>(onPosSettingUpdate);
    on<PosSettingDelete>(onPosSettingDelete);
    on<PosSettingDeleteMany>(onPosSettingDeleteMany);
    on<PosSettingGet>(onPosSettingGet);
    on<PosSettingWithImageSave>(onPosSettingWithImageSave);
    on<PosSettingWithImageUpdate>(onPosSettingWithImageUpdate);
    on<GetApiKey>(onGetApikey);
    on<DeleteApikey>(onDeleteApikey);
  }

  void onPosSettingLoad(
      PosSettingLoadList event, Emitter<PosSettingState> emit) async {
    emit(PosSettingInProgress());

    try {
      final results = await _posSettingRepository.getPosList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<PosSettingModel> posSettings = (results.data as List)
            .map((posSettings) => PosSettingModel.fromJson(posSettings))
            .toList();
        emit(PosSettingLoadSuccess(posSettings: posSettings));
      } else {
        emit(const PosSettingLoadFailed(message: 'Pos Settings Not Found'));
      }
    } catch (e) {
      emit(PosSettingLoadFailed(message: e.toString()));
    }
  }

  void onPosSettingDelete(
      PosSettingDelete event, Emitter<PosSettingState> emit) async {
    emit(PosSettingDeleteInProgress());
    try {
      await _posSettingRepository.deletePos(event.guid);

      emit(PosSettingDeleteSuccess());
    } catch (e) {
      emit(PosSettingDeleteFailed());
    }
  }

  void onPosSettingDeleteMany(
      PosSettingDeleteMany event, Emitter<PosSettingState> emit) async {
    emit(PosSettingDeleteManyInProgress());
    try {
      await _posSettingRepository.deletePosMany(event.guid);

      emit(PosSettingDeleteManySuccess());
    } catch (e) {
      emit(PosSettingDeleteFailed());
    }
  }

  void onPosSettingSave(
      PosSettingSave event, Emitter<PosSettingState> emit) async {
    emit(PosSettingSaveInProgress());
    try {
      await _posSettingRepository.savePos(event.posSetting);
      emit(PosSettingSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(PosSettingSaveFailed(message: error['message']));
    }
  }

  void onPosSettingUpdate(
      PosSettingUpdate event, Emitter<PosSettingState> emit) async {
    emit(PosSettingUpdateInProgress());
    try {
      await _posSettingRepository.updatePos(event.guid, event.posSetting);
      emit(PosSettingUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(PosSettingUpdateFailed(message: error['message']));
    }
  }

  void onPosSettingGet(
      PosSettingGet event, Emitter<PosSettingState> emit) async {
    emit(PosSettingGetInProgress());
    try {
      final result = await _posSettingRepository.getPos(event.guid);
      if (result.success) {
        PosSettingModel posSetting = PosSettingModel.fromJson(result.data);
        emit(PosSettingGetSuccess(posSetting: posSetting));
      } else {
        emit(const PosSettingGetFailed(message: 'PosSetting Not Found'));
      }
    } catch (e) {
      emit(PosSettingGetFailed(message: e.toString()));
    }
  }

  void onPosSettingWithImageSave(
      PosSettingWithImageSave event, Emitter<PosSettingState> emit) async {
    emit(PosSettingSaveInProgress());
    try {
      ApiResponse result = await _posSettingRepository.uploadImage(
          event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        PosSettingModel posSettingModel = event.posSetting;
        posSettingModel.logourl = uploadImage.uri;
        await _posSettingRepository.savePos(posSettingModel);
        emit(PosSettingSaveSuccess());
      } else {
        emit(PosSettingSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(PosSettingSaveFailed(message: e.toString()));
    }
  }

  void onPosSettingWithImageUpdate(
      PosSettingWithImageUpdate event, Emitter<PosSettingState> emit) async {
    emit(PosSettingUpdateInProgress());
    try {
      ApiResponse result = await _posSettingRepository.uploadImage(
          event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        PosSettingModel posSettingModel = event.posSetting;
        posSettingModel.logourl = uploadImage.uri;
        await _posSettingRepository.updatePos(event.guid, posSettingModel);
        emit(PosSettingUpdateSuccess());
      } else {
        emit(PosSettingUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(PosSettingUpdateFailed(message: e.toString()));
    }
  }

  void onGetApikey(GetApiKey event, Emitter<PosSettingState> emit) async {
    emit(GetApiKeyInProgress());
    try {
      final result = await _posSettingRepository.apiKeyService();
      if (result.success) {
        emit(GetApiKeySuccess(success: true, token: result.token!));
      } else {
        emit(const GetApiKeyFailed(message: 'Api key Not Found'));
      }
    } catch (e) {
      emit(GetApiKeyFailed(message: e.toString()));
    }
  }

  void onDeleteApikey(DeleteApikey event, Emitter<PosSettingState> emit) async {
    emit(DeleteApikeyInProgress());
    try {
      await _posSettingRepository.deleteApiKeyService(event.apikey);
      emit(DeleteApikeySuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DeleteApikeyFailed(message: error['message']));
    }
  }
}
