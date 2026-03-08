import 'dart:convert';

import 'package:smlaicloud/model/master_model_model.dart';
import 'package:smlaicloud/repositories/master_model_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_model_event.dart';
part 'master_model_state.dart';

class MasterModelBloc extends Bloc<MasterModelEvent, MasterModelState> {
  final MasterModelRepository _masterModelRepository;

  MasterModelBloc({required MasterModelRepository masterModelRepository})
      : _masterModelRepository = masterModelRepository,
        super(MasterModelInitial()) {
    on<MasterModelLoadList>(onMasterModelLoad);
    on<MasterModelSave>(onMasterModelSave);
    on<MasterModelUpdate>(onMasterModelUpdate);
    on<MasterModelDelete>(onMasterModelDelete);
    on<MasterModelDeleteMany>(onMasterModelDeleteMany);
    on<MasterModelGet>(onMasterModelGet);
    on<MasterModelGetByCode>(onMasterModelGetByCode);
  }

  void onMasterModelLoad(
      MasterModelLoadList event, Emitter<MasterModelState> emit) async {
    emit(MasterModelInProgress());

    try {
      final results = await _masterModelRepository.getModelList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterModelModel> models = (results.data as List)
            .map((model) => MasterModelModel.fromJson(model))
            .toList();
        emit(MasterModelLoadSuccess(models: models));
      } else {
        emit(const MasterModelLoadFailed(message: 'Model Not Found'));
      }
    } catch (e) {
      emit(MasterModelLoadFailed(message: e.toString()));
    }
  }

  void onMasterModelDelete(
      MasterModelDelete event, Emitter<MasterModelState> emit) async {
    emit(MasterModelDeleteInProgress());
    try {
      await _masterModelRepository.deleteModel(event.guid);
      emit(MasterModelDeleteSuccess());
    } catch (e) {
      emit(MasterModelDeleteFailed());
    }
  }

  void onMasterModelDeleteMany(
      MasterModelDeleteMany event, Emitter<MasterModelState> emit) async {
    emit(MasterModelDeleteManyInProgress());
    try {
      await _masterModelRepository.deleteModelMany(event.guid);
      emit(MasterModelDeleteManySuccess());
    } catch (e) {
      emit(MasterModelDeleteManyFailed());
    }
  }

  void onMasterModelSave(
      MasterModelSave event, Emitter<MasterModelState> emit) async {
    emit(MasterModelSaveInProgress());
    try {
      await _masterModelRepository.saveModel(event.modelModel);
      emit(MasterModelSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterModelSaveFailed(message: error['message']));
    }
  }

  void onMasterModelUpdate(
      MasterModelUpdate event, Emitter<MasterModelState> emit) async {
    emit(MasterModelUpdateInProgress());
    try {
      await _masterModelRepository.updateModel(event.guid, event.modelModel);
      emit(MasterModelUpdateSuccess());
    } catch (e) {
      emit(MasterModelUpdateFailed(message: e.toString()));
    }
  }

  void onMasterModelGet(
      MasterModelGet event, Emitter<MasterModelState> emit) async {
    emit(MasterModelGetInProgress());
    try {
      final result = await _masterModelRepository.getModel(event.guid);
      if (result.success) {
        MasterModelModel model = MasterModelModel.fromJson(result.data);
        emit(MasterModelGetSuccess(model: model));
      } else {
        emit(const MasterModelGetFailed(message: 'Model Not Found'));
      }
    } catch (e) {
      emit(MasterModelGetFailed(message: e.toString()));
    }
  }

  void onMasterModelGetByCode(
      MasterModelGetByCode event, Emitter<MasterModelState> emit) async {
    emit(MasterModelGetInProgress());
    try {
      final result = await _masterModelRepository.getModelByCode(event.code);
      if (result.success) {
        MasterModelModel model = MasterModelModel.fromJson(result.data);
        emit(MasterModelGetSuccess(model: model));
      } else {
        emit(const MasterModelGetFailed(message: 'Model Not Found'));
      }
    } catch (e) {
      emit(MasterModelGetFailed(message: e.toString()));
    }
  }
}
