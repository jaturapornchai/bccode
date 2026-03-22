import 'dart:convert';

import 'package:smlaicloud/model/cost_center_model.dart';
import 'package:smlaicloud/repositories/cost_center_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'cost_center_event.dart';
part 'cost_center_state.dart';

class CostCenterBloc extends Bloc<CostCenterEvent, CostCenterState> {
  final CostCenterRepository _costCenterRepository;

  CostCenterBloc({required CostCenterRepository costCenterRepository})
      : _costCenterRepository = costCenterRepository,
        super(CostCenterInitial()) {
    on<CostCenterLoadList>(onCostCenterLoad);
    on<CostCenterSave>(onCostCenterSave);
    on<CostCenterUpdate>(onCostCenterUpdate);
    on<CostCenterDelete>(onCostCenterDelete);
    on<CostCenterDeleteMany>(onCostCenterDeleteMany);
    on<CostCenterGet>(onCostCenterGet);
  }

  void onCostCenterLoad(
      CostCenterLoadList event, Emitter<CostCenterState> emit) async {
    emit(CostCenterInProgress());

    try {
      final results = await _costCenterRepository.getCostCenterList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<CostCenterModel> costCenter = (results.data as List)
            .map((costCenter) => CostCenterModel.fromJson(costCenter))
            .toList();
        emit(CostCenterLoadSuccess(costCenter: costCenter));
      } else {
        emit(const CostCenterLoadFailed(message: 'CostCenter Not Found'));
      }
    } catch (e) {
      emit(CostCenterLoadFailed(message: e.toString()));
    }
  }

  void onCostCenterDelete(
      CostCenterDelete event, Emitter<CostCenterState> emit) async {
    emit(CostCenterDeleteInProgress());
    try {
      await _costCenterRepository.deleteCostCenter(event.guid);

      emit(CostCenterDeleteSuccess());
    } catch (e) {
      emit(CostCenterDeleteFailed());
    }
  }

  void onCostCenterDeleteMany(
      CostCenterDeleteMany event, Emitter<CostCenterState> emit) async {
    emit(CostCenterDeleteManyInProgress());
    try {
      await _costCenterRepository.deleteCostCenterMany(event.guid);

      emit(CostCenterDeleteManySuccess());
    } catch (e) {
      emit(CostCenterDeleteFailed());
    }
  }

  void onCostCenterSave(
      CostCenterSave event, Emitter<CostCenterState> emit) async {
    emit(CostCenterSaveInProgress());
    try {
      final result =
          await _costCenterRepository.saveCostCenter(event.costCenter);
      if (result.success) {
        emit(CostCenterSaveSuccess(responsesID: result.id));
      }
    } catch (e) {
      try {
        final error = jsonDecode(e.toString());
        emit(CostCenterSaveFailed(message: error['message']));
      } catch (_) {
        emit(CostCenterSaveFailed(message: e.toString()));
      }
    }
  }

  void onCostCenterUpdate(
      CostCenterUpdate event, Emitter<CostCenterState> emit) async {
    emit(CostCenterUpdateInProgress());
    try {
      await _costCenterRepository.updateCostCenter(
          event.guid, event.costCenter);
      emit(CostCenterUpdateSuccess());
    } catch (e) {
      try {
        final error = jsonDecode(e.toString());
        emit(CostCenterUpdateFailed(message: error['message']));
      } catch (_) {
        emit(CostCenterUpdateFailed(message: e.toString()));
      }
    }
  }

  void onCostCenterGet(
      CostCenterGet event, Emitter<CostCenterState> emit) async {
    emit(CostCenterGetInProgress());
    try {
      final result = await _costCenterRepository.getCostCenter(event.guid);
      if (result.success) {
        CostCenterModel costCenter = CostCenterModel.fromJson(result.data);
        emit(CostCenterGetSuccess(costCenter: costCenter));
      } else {
        emit(const CostCenterGetFailed(message: 'CostCenter Not Found'));
      }
    } catch (e) {
      emit(CostCenterGetFailed(message: e.toString()));
    }
  }
}
