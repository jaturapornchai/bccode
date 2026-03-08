import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/creditor_group_model.dart';
import 'package:smlaicloud/repositories/creditor_group_repository.dart';

part 'creditor_group_event.dart';
part 'creditor_group_state.dart';

class CreditorGroupBloc extends Bloc<CreditorGroupEvent, CreditorGroupState> {
  final CreditorGroupRepository _creditorGroupRepository;

  CreditorGroupBloc({required CreditorGroupRepository creditorGroupRepository})
      : _creditorGroupRepository = creditorGroupRepository,
        super(CreditorGroupInitial()) {
    on<CreditorGroupLoadList>(onCreditorGroupLoad);
    on<CreditorGroupSave>(onCreditorGroupSave);
    on<CreditorGroupUpdate>(onCreditorGroupUpdate);
    on<CreditorGroupDelete>(onCreditorGroupDelete);
    on<CreditorGroupDeleteMany>(onCreditorGroupDeleteMany);
    on<CreditorGroupGet>(onCreditorGroupGet);
  }

  void onCreditorGroupLoad(
      CreditorGroupLoadList event, Emitter<CreditorGroupState> emit) async {
    emit(CreditorGroupInProgress());

    try {
      final results = await _creditorGroupRepository.getCreditorGroupList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<CreditorGroupModel> creditorGroups = (results.data as List)
            .map((creditorGroup) => CreditorGroupModel.fromJson(creditorGroup))
            .toList();
        emit(CreditorGroupLoadSuccess(creditorGroups: creditorGroups));
      } else {
        emit(
            const CreditorGroupLoadFailed(message: 'Creditor Group Not Found'));
      }
    } catch (e) {
      emit(CreditorGroupLoadFailed(message: e.toString()));
    }
  }

  void onCreditorGroupDelete(
      CreditorGroupDelete event, Emitter<CreditorGroupState> emit) async {
    emit(CreditorGroupDeleteInProgress());
    try {
      await _creditorGroupRepository.deleteCreditorGroup(event.guid);

      emit(CreditorGroupDeleteSuccess());
    } catch (e) {
      emit(CreditorGroupDeleteFailed());
    }
  }

  void onCreditorGroupDeleteMany(
      CreditorGroupDeleteMany event, Emitter<CreditorGroupState> emit) async {
    emit(CreditorGroupDeleteManyInProgress());
    try {
      await _creditorGroupRepository.deleteCreditorGroupMany(event.guid);

      emit(CreditorGroupDeleteManySuccess());
    } catch (e) {
      emit(CreditorGroupDeleteFailed());
    }
  }

  void onCreditorGroupSave(
      CreditorGroupSave event, Emitter<CreditorGroupState> emit) async {
    emit(CreditorGroupSaveInProgress());
    try {
      await _creditorGroupRepository.saveCreditorGroup(event.creditorGroup);
      emit(CreditorGroupSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CreditorGroupSaveFailed(message: error['message']));
    }
  }

  void onCreditorGroupUpdate(
      CreditorGroupUpdate event, Emitter<CreditorGroupState> emit) async {
    emit(CreditorGroupUpdateInProgress());
    try {
      await _creditorGroupRepository.updateCreditorGroup(
          event.guid, event.creditorGroup);
      emit(CreditorGroupUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CreditorGroupUpdateFailed(message: error['message']));
    }
  }

  void onCreditorGroupGet(
      CreditorGroupGet event, Emitter<CreditorGroupState> emit) async {
    emit(CreditorGroupGetInProgress());
    try {
      final result =
          await _creditorGroupRepository.getCreditorGroup(event.guid);
      if (result.success) {
        CreditorGroupModel creditorGroup =
            CreditorGroupModel.fromJson(result.data);
        emit(CreditorGroupGetSuccess(creditorGroup: creditorGroup));
      } else {
        emit(const CreditorGroupGetFailed(message: 'CreditorGroup Not Found'));
      }
    } catch (e) {
      // emit(CreditorGroupDeleteFailure(message: e.toString()));
    }
  }
}
