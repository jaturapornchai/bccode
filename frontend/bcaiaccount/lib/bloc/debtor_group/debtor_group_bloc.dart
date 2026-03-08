import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/debtor_group_model.dart';
import 'package:smlaicloud/repositories/debtor_group_repository.dart';

part 'debtor_group_event.dart';
part 'debtor_group_state.dart';

class DebtorGroupBloc extends Bloc<DebtorGroupEvent, DebtorGroupState> {
  final DebtorGroupRepository _debtorGroupRepository;

  DebtorGroupBloc({required DebtorGroupRepository debtorGroupRepository})
      : _debtorGroupRepository = debtorGroupRepository,
        super(DebtorGroupInitial()) {
    on<DebtorGroupLoadList>(onDebtorGroupLoad);
    on<DebtorGroupSave>(onDebtorGroupSave);
    on<DebtorGroupUpdate>(onDebtorGroupUpdate);
    on<DebtorGroupDelete>(onDebtorGroupDelete);
    on<DebtorGroupDeleteMany>(onDebtorGroupDeleteMany);
    on<DebtorGroupGet>(onDebtorGroupGet);
  }

  void onDebtorGroupLoad(
      DebtorGroupLoadList event, Emitter<DebtorGroupState> emit) async {
    emit(DebtorGroupInProgress());

    try {
      final results = await _debtorGroupRepository.getDebtorGroupList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<DebtorGroupModel> debtorGroups = (results.data as List)
            .map((debtorGroups) => DebtorGroupModel.fromJson(debtorGroups))
            .toList();
        emit(DebtorGroupLoadSuccess(debtorGroups: debtorGroups));
      } else {
        emit(const DebtorGroupLoadFailed(message: 'Debtor Group Not Found'));
      }
    } catch (e) {
      emit(DebtorGroupLoadFailed(message: e.toString()));
    }
  }

  void onDebtorGroupDelete(
      DebtorGroupDelete event, Emitter<DebtorGroupState> emit) async {
    emit(DebtorGroupDeleteInProgress());
    try {
      await _debtorGroupRepository.deleteDebtorGroup(event.guid);

      emit(DebtorGroupDeleteSuccess());
    } catch (e) {
      emit(DebtorGroupDeleteFailed());
    }
  }

  void onDebtorGroupDeleteMany(
      DebtorGroupDeleteMany event, Emitter<DebtorGroupState> emit) async {
    emit(DebtorGroupDeleteManyInProgress());
    try {
      await _debtorGroupRepository.deleteDebtorGroupMany(event.guid);

      emit(DebtorGroupDeleteManySuccess());
    } catch (e) {
      emit(DebtorGroupDeleteFailed());
    }
  }

  void onDebtorGroupSave(
      DebtorGroupSave event, Emitter<DebtorGroupState> emit) async {
    emit(DebtorGroupSaveInProgress());
    try {
      await _debtorGroupRepository.saveDebtorGroup(event.debtorGroups);
      emit(DebtorGroupSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DebtorGroupSaveFailed(message: error['message']));
    }
  }

  void onDebtorGroupUpdate(
      DebtorGroupUpdate event, Emitter<DebtorGroupState> emit) async {
    emit(DebtorGroupUpdateInProgress());
    try {
      await _debtorGroupRepository.updateDebtorGroup(
          event.guid, event.debtorGroups);
      emit(DebtorGroupUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DebtorGroupUpdateFailed(message: error['message']));
    }
  }

  void onDebtorGroupGet(
      DebtorGroupGet event, Emitter<DebtorGroupState> emit) async {
    emit(DebtorGroupGetInProgress());
    try {
      final result = await _debtorGroupRepository.getDebtorGroup(event.guid);
      if (result.success) {
        DebtorGroupModel debtorGroups = DebtorGroupModel.fromJson(result.data);
        emit(DebtorGroupGetSuccess(debtorGroups: debtorGroups));
      } else {
        emit(const DebtorGroupGetFailed(message: 'DebtorGroup Not Found'));
      }
    } catch (e) {
      // emit(DebtorGroupDeleteFailure(message: e.toString()));
    }
  }
}
