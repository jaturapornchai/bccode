import 'dart:convert';

import 'package:smlaicloud/model/master_pattern_model.dart';
import 'package:smlaicloud/repositories/master_pattern_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_pattern_event.dart';
part 'master_pattern_state.dart';

class MasterPatternBloc extends Bloc<MasterPatternEvent, MasterPatternState> {
  final MasterPatternRepository _masterPatternRepository;

  MasterPatternBloc({required MasterPatternRepository masterPatternRepository})
      : _masterPatternRepository = masterPatternRepository,
        super(MasterPatternInitial()) {
    on<MasterPatternLoadList>(onMasterPatternLoad);
    on<MasterPatternSave>(onMasterPatternSave);
    on<MasterPatternUpdate>(onMasterPatternUpdate);
    on<MasterPatternDelete>(onMasterPatternDelete);
    on<MasterPatternDeleteMany>(onMasterPatternDeleteMany);
    on<MasterPatternGet>(onMasterPatternGet);
    on<MasterPatternGetByCode>(onMasterPatternGetByCode);
  }

  void onMasterPatternLoad(
      MasterPatternLoadList event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternInProgress());

    try {
      final results = await _masterPatternRepository.getPatternList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterPatternModel> patterns = (results.data as List)
            .map((pattern) => MasterPatternModel.fromJson(pattern))
            .toList();
        emit(MasterPatternLoadSuccess(patterns: patterns));
      } else {
        emit(const MasterPatternLoadFailed(message: 'Pattern Not Found'));
      }
    } catch (e) {
      emit(MasterPatternLoadFailed(message: e.toString()));
    }
  }

  void onMasterPatternDelete(
      MasterPatternDelete event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternDeleteInProgress());
    try {
      await _masterPatternRepository.deletePattern(event.guid);
      emit(MasterPatternDeleteSuccess());
    } catch (e) {
      emit(MasterPatternDeleteFailed());
    }
  }

  void onMasterPatternDeleteMany(
      MasterPatternDeleteMany event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternDeleteManyInProgress());
    try {
      await _masterPatternRepository.deletePatternMany(event.guid);
      emit(MasterPatternDeleteManySuccess());
    } catch (e) {
      emit(MasterPatternDeleteManyFailed());
    }
  }

  void onMasterPatternSave(
      MasterPatternSave event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternSaveInProgress());
    try {
      await _masterPatternRepository.savePattern(event.patternModel);
      emit(MasterPatternSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterPatternSaveFailed(message: error['message']));
    }
  }

  void onMasterPatternUpdate(
      MasterPatternUpdate event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternUpdateInProgress());
    try {
      await _masterPatternRepository.updatePattern(
          event.guid, event.patternModel);
      emit(MasterPatternUpdateSuccess());
    } catch (e) {
      emit(MasterPatternUpdateFailed(message: e.toString()));
    }
  }

  void onMasterPatternGet(
      MasterPatternGet event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternGetInProgress());
    try {
      final result = await _masterPatternRepository.getPattern(event.guid);
      if (result.success) {
        MasterPatternModel pattern = MasterPatternModel.fromJson(result.data);
        emit(MasterPatternGetSuccess(pattern: pattern));
      } else {
        emit(const MasterPatternGetFailed(message: 'Pattern Not Found'));
      }
    } catch (e) {
      emit(MasterPatternGetFailed(message: e.toString()));
    }
  }

  void onMasterPatternGetByCode(
      MasterPatternGetByCode event, Emitter<MasterPatternState> emit) async {
    emit(MasterPatternGetInProgress());
    try {
      final result =
          await _masterPatternRepository.getPatternByCode(event.code);
      if (result.success) {
        MasterPatternModel pattern = MasterPatternModel.fromJson(result.data);
        emit(MasterPatternGetSuccess(pattern: pattern));
      } else {
        emit(const MasterPatternGetFailed(message: 'Pattern Not Found'));
      }
    } catch (e) {
      emit(MasterPatternGetFailed(message: e.toString()));
    }
  }
}
