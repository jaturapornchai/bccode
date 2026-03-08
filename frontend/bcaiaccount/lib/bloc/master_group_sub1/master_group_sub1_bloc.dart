import 'dart:convert';

import 'package:smlaicloud/model/master_group_sub1_model.dart';
import 'package:smlaicloud/repositories/master_group_sub1_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_group_sub1_event.dart';
part 'master_group_sub1_state.dart';

class MasterGroupSub1Bloc
    extends Bloc<MasterGroupSub1Event, MasterGroupSub1State> {
  final MasterGroupSub1Repository _masterGroupSub1Repository;

  MasterGroupSub1Bloc(
      {required MasterGroupSub1Repository masterGroupSub1Repository})
      : _masterGroupSub1Repository = masterGroupSub1Repository,
        super(MasterGroupSub1Initial()) {
    on<MasterGroupSub1LoadList>(onMasterGroupSub1Load);
    on<MasterGroupSub1Save>(onMasterGroupSub1Save);
    on<MasterGroupSub1Update>(onMasterGroupSub1Update);
    on<MasterGroupSub1Delete>(onMasterGroupSub1Delete);
    on<MasterGroupSub1DeleteMany>(onMasterGroupSub1DeleteMany);
    on<MasterGroupSub1Get>(onMasterGroupSub1Get);
    on<MasterGroupSub1GetByCode>(onMasterGroupSub1GetByCode);
  }

  void onMasterGroupSub1Load(
      MasterGroupSub1LoadList event, Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1InProgress());

    try {
      final results = await _masterGroupSub1Repository.getGroupSub1List(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterGroupSub1Model> groupSub1s = (results.data as List)
            .map((groupSub1) => MasterGroupSub1Model.fromJson(groupSub1))
            .toList();
        emit(MasterGroupSub1LoadSuccess(groupSub1s: groupSub1s));
      } else {
        emit(const MasterGroupSub1LoadFailed(message: 'Group Sub1 Not Found'));
      }
    } catch (e) {
      emit(MasterGroupSub1LoadFailed(message: e.toString()));
    }
  }

  void onMasterGroupSub1Delete(
      MasterGroupSub1Delete event, Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1DeleteInProgress());
    try {
      await _masterGroupSub1Repository.deleteGroupSub1(event.guid);
      emit(MasterGroupSub1DeleteSuccess());
    } catch (e) {
      emit(MasterGroupSub1DeleteFailed());
    }
  }

  void onMasterGroupSub1DeleteMany(MasterGroupSub1DeleteMany event,
      Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1DeleteManyInProgress());
    try {
      await _masterGroupSub1Repository.deleteGroupSub1Many(event.guid);
      emit(MasterGroupSub1DeleteManySuccess());
    } catch (e) {
      emit(MasterGroupSub1DeleteManyFailed());
    }
  }

  void onMasterGroupSub1Save(
      MasterGroupSub1Save event, Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1SaveInProgress());
    try {
      await _masterGroupSub1Repository.saveGroupSub1(event.groupSub1Model);
      emit(MasterGroupSub1SaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterGroupSub1SaveFailed(message: error['message']));
    }
  }

  void onMasterGroupSub1Update(
      MasterGroupSub1Update event, Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1UpdateInProgress());
    try {
      await _masterGroupSub1Repository.updateGroupSub1(
          event.guid, event.groupSub1Model);
      emit(MasterGroupSub1UpdateSuccess());
    } catch (e) {
      emit(MasterGroupSub1UpdateFailed(message: e.toString()));
    }
  }

  void onMasterGroupSub1Get(
      MasterGroupSub1Get event, Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1GetInProgress());
    try {
      final result = await _masterGroupSub1Repository.getGroupSub1(event.guid);
      if (result.success) {
        MasterGroupSub1Model groupSub1 =
            MasterGroupSub1Model.fromJson(result.data);
        emit(MasterGroupSub1GetSuccess(groupSub1: groupSub1));
      } else {
        emit(const MasterGroupSub1GetFailed(message: 'Group Sub1 Not Found'));
      }
    } catch (e) {
      emit(MasterGroupSub1GetFailed(message: e.toString()));
    }
  }

  void onMasterGroupSub1GetByCode(MasterGroupSub1GetByCode event,
      Emitter<MasterGroupSub1State> emit) async {
    emit(MasterGroupSub1GetInProgress());
    try {
      final result =
          await _masterGroupSub1Repository.getGroupSub1ByCode(event.code);
      if (result.success) {
        MasterGroupSub1Model groupSub1 =
            MasterGroupSub1Model.fromJson(result.data);
        emit(MasterGroupSub1GetSuccess(groupSub1: groupSub1));
      } else {
        emit(const MasterGroupSub1GetFailed(message: 'Group Sub1 Not Found'));
      }
    } catch (e) {
      emit(MasterGroupSub1GetFailed(message: e.toString()));
    }
  }
}
