import 'dart:convert';

import 'package:smlaicloud/model/master_group_sub2_model.dart';
import 'package:smlaicloud/repositories/master_group_sub2_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_group_sub2_event.dart';
part 'master_group_sub2_state.dart';

class MasterGroupSub2Bloc
    extends Bloc<MasterGroupSub2Event, MasterGroupSub2State> {
  final MasterGroupSub2Repository _masterGroupSub2Repository;

  MasterGroupSub2Bloc(
      {required MasterGroupSub2Repository masterGroupSub2Repository})
      : _masterGroupSub2Repository = masterGroupSub2Repository,
        super(MasterGroupSub2Initial()) {
    on<MasterGroupSub2LoadList>(onMasterGroupSub2Load);
    on<MasterGroupSub2Save>(onMasterGroupSub2Save);
    on<MasterGroupSub2Update>(onMasterGroupSub2Update);
    on<MasterGroupSub2Delete>(onMasterGroupSub2Delete);
    on<MasterGroupSub2DeleteMany>(onMasterGroupSub2DeleteMany);
    on<MasterGroupSub2Get>(onMasterGroupSub2Get);
    on<MasterGroupSub2GetByCode>(onMasterGroupSub2GetByCode);
  }

  void onMasterGroupSub2Load(
      MasterGroupSub2LoadList event, Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2InProgress());

    try {
      final results = await _masterGroupSub2Repository.getGroupSub2List(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterGroupSub2Model> groupSub2s = (results.data as List)
            .map((groupSub2) => MasterGroupSub2Model.fromJson(groupSub2))
            .toList();
        emit(MasterGroupSub2LoadSuccess(groupSub2s: groupSub2s));
      } else {
        emit(const MasterGroupSub2LoadFailed(message: 'Group Sub2 Not Found'));
      }
    } catch (e) {
      emit(MasterGroupSub2LoadFailed(message: e.toString()));
    }
  }

  void onMasterGroupSub2Delete(
      MasterGroupSub2Delete event, Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2DeleteInProgress());
    try {
      await _masterGroupSub2Repository.deleteGroupSub2(event.guid);
      emit(MasterGroupSub2DeleteSuccess());
    } catch (e) {
      emit(MasterGroupSub2DeleteFailed());
    }
  }

  void onMasterGroupSub2DeleteMany(MasterGroupSub2DeleteMany event,
      Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2DeleteManyInProgress());
    try {
      await _masterGroupSub2Repository.deleteGroupSub2Many(event.guid);
      emit(MasterGroupSub2DeleteManySuccess());
    } catch (e) {
      emit(MasterGroupSub2DeleteManyFailed());
    }
  }

  void onMasterGroupSub2Save(
      MasterGroupSub2Save event, Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2SaveInProgress());
    try {
      await _masterGroupSub2Repository.saveGroupSub2(event.groupSub2Model);
      emit(MasterGroupSub2SaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterGroupSub2SaveFailed(message: error['message']));
    }
  }

  void onMasterGroupSub2Update(
      MasterGroupSub2Update event, Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2UpdateInProgress());
    try {
      await _masterGroupSub2Repository.updateGroupSub2(
          event.guid, event.groupSub2Model);
      emit(MasterGroupSub2UpdateSuccess());
    } catch (e) {
      emit(MasterGroupSub2UpdateFailed(message: e.toString()));
    }
  }

  void onMasterGroupSub2Get(
      MasterGroupSub2Get event, Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2GetInProgress());
    try {
      final result = await _masterGroupSub2Repository.getGroupSub2(event.guid);
      if (result.success) {
        MasterGroupSub2Model groupSub2 =
            MasterGroupSub2Model.fromJson(result.data);
        emit(MasterGroupSub2GetSuccess(groupSub2: groupSub2));
      } else {
        emit(const MasterGroupSub2GetFailed(message: 'Group Sub2 Not Found'));
      }
    } catch (e) {
      emit(MasterGroupSub2GetFailed(message: e.toString()));
    }
  }

  void onMasterGroupSub2GetByCode(MasterGroupSub2GetByCode event,
      Emitter<MasterGroupSub2State> emit) async {
    emit(MasterGroupSub2GetInProgress());
    try {
      final result =
          await _masterGroupSub2Repository.getGroupSub2ByCode(event.code);
      if (result.success) {
        MasterGroupSub2Model groupSub2 =
            MasterGroupSub2Model.fromJson(result.data);
        emit(MasterGroupSub2GetSuccess(groupSub2: groupSub2));
      } else {
        emit(const MasterGroupSub2GetFailed(message: 'Group Sub2 Not Found'));
      }
    } catch (e) {
      emit(MasterGroupSub2GetFailed(message: e.toString()));
    }
  }
}
