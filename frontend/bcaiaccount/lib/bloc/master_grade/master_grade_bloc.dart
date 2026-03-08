import 'dart:convert';

import 'package:smlaicloud/model/master_grade_model.dart';
import 'package:smlaicloud/repositories/master_grade_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_grade_event.dart';
part 'master_grade_state.dart';

class MasterGradeBloc extends Bloc<MasterGradeEvent, MasterGradeState> {
  final MasterGradeRepository _masterGradeRepository;

  MasterGradeBloc({required MasterGradeRepository masterGradeRepository})
      : _masterGradeRepository = masterGradeRepository,
        super(MasterGradeInitial()) {
    on<MasterGradeLoadList>(onMasterGradeLoad);
    on<MasterGradeSave>(onMasterGradeSave);
    on<MasterGradeUpdate>(onMasterGradeUpdate);
    on<MasterGradeDelete>(onMasterGradeDelete);
    on<MasterGradeDeleteMany>(onMasterGradeDeleteMany);
    on<MasterGradeGet>(onMasterGradeGet);
    on<MasterGradeGetByCode>(onMasterGradeGetByCode);
  }

  void onMasterGradeLoad(
      MasterGradeLoadList event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeInProgress());

    try {
      final results = await _masterGradeRepository.getGradeList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterGradeModel> grades = (results.data as List)
            .map((grade) => MasterGradeModel.fromJson(grade))
            .toList();
        emit(MasterGradeLoadSuccess(grades: grades));
      } else {
        emit(const MasterGradeLoadFailed(message: 'Grade Not Found'));
      }
    } catch (e) {
      emit(MasterGradeLoadFailed(message: e.toString()));
    }
  }

  void onMasterGradeDelete(
      MasterGradeDelete event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeDeleteInProgress());
    try {
      await _masterGradeRepository.deleteGrade(event.guid);
      emit(MasterGradeDeleteSuccess());
    } catch (e) {
      emit(MasterGradeDeleteFailed());
    }
  }

  void onMasterGradeDeleteMany(
      MasterGradeDeleteMany event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeDeleteManyInProgress());
    try {
      await _masterGradeRepository.deleteGradeMany(event.guid);
      emit(MasterGradeDeleteManySuccess());
    } catch (e) {
      emit(MasterGradeDeleteManyFailed());
    }
  }

  void onMasterGradeSave(
      MasterGradeSave event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeSaveInProgress());
    try {
      await _masterGradeRepository.saveGrade(event.gradeModel);
      emit(MasterGradeSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterGradeSaveFailed(message: error['message']));
    }
  }

  void onMasterGradeUpdate(
      MasterGradeUpdate event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeUpdateInProgress());
    try {
      await _masterGradeRepository.updateGrade(event.guid, event.gradeModel);
      emit(MasterGradeUpdateSuccess());
    } catch (e) {
      emit(MasterGradeUpdateFailed(message: e.toString()));
    }
  }

  void onMasterGradeGet(
      MasterGradeGet event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeGetInProgress());
    try {
      final result = await _masterGradeRepository.getGrade(event.guid);
      if (result.success) {
        MasterGradeModel grade = MasterGradeModel.fromJson(result.data);
        emit(MasterGradeGetSuccess(grade: grade));
      } else {
        emit(const MasterGradeGetFailed(message: 'Grade Not Found'));
      }
    } catch (e) {
      emit(MasterGradeGetFailed(message: e.toString()));
    }
  }

  void onMasterGradeGetByCode(
      MasterGradeGetByCode event, Emitter<MasterGradeState> emit) async {
    emit(MasterGradeGetInProgress());
    try {
      final result = await _masterGradeRepository.getGradeByCode(event.code);
      if (result.success) {
        MasterGradeModel grade = MasterGradeModel.fromJson(result.data);
        emit(MasterGradeGetSuccess(grade: grade));
      } else {
        emit(const MasterGradeGetFailed(message: 'Grade Not Found'));
      }
    } catch (e) {
      emit(MasterGradeGetFailed(message: e.toString()));
    }
  }
}
