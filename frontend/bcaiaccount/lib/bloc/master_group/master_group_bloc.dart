import 'dart:convert';

import 'package:smlaicloud/model/master_group_model.dart';
import 'package:smlaicloud/repositories/master_group_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_group_event.dart';
part 'master_group_state.dart';

class MasterGroupBloc extends Bloc<MasterGroupEvent, MasterGroupState> {
  final MasterGroupRepository _masterGroupRepository;

  MasterGroupBloc({required MasterGroupRepository masterGroupRepository})
      : _masterGroupRepository = masterGroupRepository,
        super(MasterGroupInitial()) {
    on<MasterGroupLoadList>(onMasterGroupLoad);
    on<MasterGroupSave>(onMasterGroupSave);
    on<MasterGroupUpdate>(onMasterGroupUpdate);
    on<MasterGroupDelete>(onMasterGroupDelete);
    on<MasterGroupDeleteMany>(onMasterGroupDeleteMany);
    on<MasterGroupGet>(onMasterGroupGet);
    on<MasterGroupGetByCode>(onMasterGroupGetByCode);
  }

  void onMasterGroupLoad(
      MasterGroupLoadList event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupInProgress());

    try {
      final results = await _masterGroupRepository.getGroupList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterGroupModel> groups = (results.data as List)
            .map((group) => MasterGroupModel.fromJson(group))
            .toList();
        emit(MasterGroupLoadSuccess(groups: groups));
      } else {
        emit(const MasterGroupLoadFailed(message: 'Group Not Found'));
      }
    } catch (e) {
      emit(MasterGroupLoadFailed(message: e.toString()));
    }
  }

  void onMasterGroupDelete(
      MasterGroupDelete event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupDeleteInProgress());
    try {
      await _masterGroupRepository.deleteGroup(event.guid);
      emit(MasterGroupDeleteSuccess());
    } catch (e) {
      emit(MasterGroupDeleteFailed());
    }
  }

  void onMasterGroupDeleteMany(
      MasterGroupDeleteMany event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupDeleteManyInProgress());
    try {
      await _masterGroupRepository.deleteGroupMany(event.guid);
      emit(MasterGroupDeleteManySuccess());
    } catch (e) {
      emit(MasterGroupDeleteManyFailed());
    }
  }

  void onMasterGroupSave(
      MasterGroupSave event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupSaveInProgress());
    try {
      await _masterGroupRepository.saveGroup(event.groupModel);
      emit(MasterGroupSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterGroupSaveFailed(message: error['message']));
    }
  }

  void onMasterGroupUpdate(
      MasterGroupUpdate event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupUpdateInProgress());
    try {
      await _masterGroupRepository.updateGroup(event.guid, event.groupModel);
      emit(MasterGroupUpdateSuccess());
    } catch (e) {
      emit(MasterGroupUpdateFailed(message: e.toString()));
    }
  }

  void onMasterGroupGet(
      MasterGroupGet event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupGetInProgress());
    try {
      final result = await _masterGroupRepository.getGroup(event.guid);
      if (result.success) {
        MasterGroupModel group = MasterGroupModel.fromJson(result.data);
        emit(MasterGroupGetSuccess(group: group));
      } else {
        emit(const MasterGroupGetFailed(message: 'Group Not Found'));
      }
    } catch (e) {
      emit(MasterGroupGetFailed(message: e.toString()));
    }
  }

  void onMasterGroupGetByCode(
      MasterGroupGetByCode event, Emitter<MasterGroupState> emit) async {
    emit(MasterGroupGetInProgress());
    try {
      final result = await _masterGroupRepository.getGroupByCode(event.code);
      if (result.success) {
        MasterGroupModel group = MasterGroupModel.fromJson(result.data);
        emit(MasterGroupGetSuccess(group: group));
      } else {
        emit(const MasterGroupGetFailed(message: 'Group Not Found'));
      }
    } catch (e) {
      emit(MasterGroupGetFailed(message: e.toString()));
    }
  }
}
