import 'dart:convert';

import 'package:smlaicloud/model/master_class_model.dart';
import 'package:smlaicloud/repositories/master_class_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_class_event.dart';
part 'master_class_state.dart';

class MasterClassBloc extends Bloc<MasterClassEvent, MasterClassState> {
  final MasterClassRepository _masterClassRepository;

  MasterClassBloc({required MasterClassRepository masterClassRepository})
      : _masterClassRepository = masterClassRepository,
        super(MasterClassInitial()) {
    on<MasterClassLoadList>(onMasterClassLoad);
    on<MasterClassSave>(onMasterClassSave);
    on<MasterClassUpdate>(onMasterClassUpdate);
    on<MasterClassDelete>(onMasterClassDelete);
    on<MasterClassDeleteMany>(onMasterClassDeleteMany);
    on<MasterClassGet>(onMasterClassGet);
    on<MasterClassGetByCode>(onMasterClassGetByCode);
  }

  void onMasterClassLoad(
      MasterClassLoadList event, Emitter<MasterClassState> emit) async {
    emit(MasterClassInProgress());

    try {
      final results = await _masterClassRepository.getClassList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterClassModel> classes = (results.data as List)
            .map((classData) => MasterClassModel.fromJson(classData))
            .toList();
        emit(MasterClassLoadSuccess(classes: classes));
      } else {
        emit(const MasterClassLoadFailed(message: 'Class Not Found'));
      }
    } catch (e) {
      emit(MasterClassLoadFailed(message: e.toString()));
    }
  }

  void onMasterClassDelete(
      MasterClassDelete event, Emitter<MasterClassState> emit) async {
    emit(MasterClassDeleteInProgress());
    try {
      await _masterClassRepository.deleteClass(event.guid);
      emit(MasterClassDeleteSuccess());
    } catch (e) {
      emit(MasterClassDeleteFailed());
    }
  }

  void onMasterClassDeleteMany(
      MasterClassDeleteMany event, Emitter<MasterClassState> emit) async {
    emit(MasterClassDeleteManyInProgress());
    try {
      await _masterClassRepository.deleteClassMany(event.guid);
      emit(MasterClassDeleteManySuccess());
    } catch (e) {
      emit(MasterClassDeleteManyFailed());
    }
  }

  void onMasterClassSave(
      MasterClassSave event, Emitter<MasterClassState> emit) async {
    emit(MasterClassSaveInProgress());
    try {
      await _masterClassRepository.saveClass(event.classModel);
      emit(MasterClassSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterClassSaveFailed(message: error['message']));
    }
  }

  void onMasterClassUpdate(
      MasterClassUpdate event, Emitter<MasterClassState> emit) async {
    emit(MasterClassUpdateInProgress());
    try {
      await _masterClassRepository.updateClass(event.guid, event.classModel);
      emit(MasterClassUpdateSuccess());
    } catch (e) {
      emit(MasterClassUpdateFailed(message: e.toString()));
    }
  }

  void onMasterClassGet(
      MasterClassGet event, Emitter<MasterClassState> emit) async {
    emit(MasterClassGetInProgress());
    try {
      final result = await _masterClassRepository.getClass(event.guid);
      if (result.success) {
        MasterClassModel classData = MasterClassModel.fromJson(result.data);
        emit(MasterClassGetSuccess(classData: classData));
      } else {
        emit(const MasterClassGetFailed(message: 'Class Not Found'));
      }
    } catch (e) {
      emit(MasterClassGetFailed(message: e.toString()));
    }
  }

  void onMasterClassGetByCode(
      MasterClassGetByCode event, Emitter<MasterClassState> emit) async {
    emit(MasterClassGetInProgress());
    try {
      final result = await _masterClassRepository.getClassByCode(event.code);
      if (result.success) {
        MasterClassModel classData = MasterClassModel.fromJson(result.data);
        emit(MasterClassGetSuccess(classData: classData));
      } else {
        emit(const MasterClassGetFailed(message: 'Class Not Found'));
      }
    } catch (e) {
      emit(MasterClassGetFailed(message: e.toString()));
    }
  }
}
