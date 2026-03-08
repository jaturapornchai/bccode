import 'package:flutter/foundation.dart';
import 'dart:convert';
// import 'dart:io'; // Not used

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/unit_repository.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'unit_event.dart';
part 'unit_state.dart';

class UnitBloc extends Bloc<UnitEvent, UnitState> {
  final UnitRepository _unitRepository;

  UnitBloc({required UnitRepository unitRepository})
    : _unitRepository = unitRepository,
      super(UnitInitial()) {
    on<UnitLoadList>(onUnitLoad);
    on<UnitSave>(onUnitSave);
    on<UnitUpdate>(onUnitUpdate);
    on<UnitDelete>(unitDelete);
    on<UnitDeleteMany>(unitDeleteMany);
    on<UnitGet>(onUnitGet);
    on<UnitSaveBulk>(onUnitSaveBulk);
    on<UnitUploadExcel>(onUnitUploadExcel);
    on<UnitLoadFromMainShop>(onUnitLoadFromMainShop);
  }

  void onUnitLoad(UnitLoadList event, Emitter<UnitState> emit) async {
    emit(UnitInProgress());

    try {
      final results = await _unitRepository.getUnitList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
      );

      if (results.success) {
        List<UnitModel> units = (results.data as List)
            .map((unit) => UnitModel.fromJson(unit))
            .toList();
        // print(units.length);
        emit(UnitLoadSuccess(units: units));
      } else {
        emit(const UnitLoadFailed(message: 'Unit Not Found'));
      }
    } catch (e) {
      emit(UnitLoadFailed(message: e.toString()));
    }
  }

  void unitDelete(UnitDelete event, Emitter<UnitState> emit) async {
    emit(UnitDeleteInProgress());
    try {
      await _unitRepository.deleteUnit(event.guid);

      emit(UnitDeleteSuccess());
    } catch (e) {
      // emit(UnitDeleteFailure(message: e.toString()));
    }
  }

  void unitDeleteMany(UnitDeleteMany event, Emitter<UnitState> emit) async {
    emit(UnitDeleteManyInProgress());
    try {
      await _unitRepository.deleteUnitMany(event.guid);

      emit(UnitDeleteManySuccess());
    } catch (e) {
      // emit(UnitDeleteFailure(message: e.toString()));
    }
  }

  void onUnitSave(UnitSave event, Emitter<UnitState> emit) async {
    emit(UnitSaveInProgress());
    try {
      await _unitRepository.saveUnit(event.unitModel);
      emit(UnitSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(UnitSaveFailed(message: error['message']));
    }
  }

  void onUnitSaveBulk(UnitSaveBulk event, Emitter<UnitState> emit) async {
    emit(UnitSaveInProgress());
    try {
      await _unitRepository.saveUnitBulk(event.units);
      emit(UnitSaveSuccess());
    } catch (e) {
      emit(UnitSaveFailed(message: e.toString()));
    }
  }

  void onUnitUpdate(UnitUpdate event, Emitter<UnitState> emit) async {
    emit(UnitUpdateInProgress());
    try {
      await _unitRepository.updateUnit(event.guid, event.unitModel);
      emit(UnitUpdateSuccess());
    } catch (e) {
      emit(UnitUpdateFailed(message: e.toString()));
    }
  }

  void onUnitGet(UnitGet event, Emitter<UnitState> emit) async {
    emit(UnitGetInProgress());
    try {
      final result = await _unitRepository.getUnit(event.guid);
      if (result.success) {
        UnitModel unit = UnitModel.fromJson(result.data);
        emit(UnitGetSuccess(unit: unit));
      } else {
        emit(const UnitGetFailed(message: 'Unit Not Found'));
      }
    } catch (e) {
      // emit(UnitDeleteFailure(message: e.toString()));
    }
  }

  void onUnitLoadFromMainShop(
    UnitLoadFromMainShop event,
    Emitter<UnitState> emit,
  ) async {
    emit(UnitLoadFromMainShopInProgress());

    try {
      final results = await _unitRepository.getUnitListFromMainShop(
        event.mainShopId,
      );

      if (results.success) {
        List<UnitModel> units = (results.data as List)
            .map((unit) => UnitModel.fromJson(unit))
            .toList();
        emit(UnitLoadFromMainShopSuccess(units: units));
      } else {
        emit(
          const UnitLoadFromMainShopFailed(
            message: 'Unit from Main Shop Not Found',
          ),
        );
      }
    } catch (e) {
      emit(UnitLoadFromMainShopFailed(message: e.toString()));
    }
  }

  void onUnitUploadExcel(UnitUploadExcel event, Emitter<UnitState> emit) async {
    emit(UnitUploadExcelInProgress());
    try {
      final result = await _unitRepository.uploadExcelFile(
        event.file,
        event.filename,
      );
      if (result.success) {
        emit(UnitUploadExcelSuccess(uploadResult: result.data.toString()));
      } else {
        emit(const UnitUploadExcelFailed(message: 'Upload Failed'));
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("Upload error: $e");
      }
      emit(UnitUploadExcelFailed(message: e.toString()));
    }
  }
}
