import 'dart:convert';

import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/repositories/warehouse_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'warehose_event.dart';
part 'warehose_state.dart';

class WarehouseBloc extends Bloc<WarehouseEvent, WarehouseState> {
  final WarehouseRepository _warehouseRepository;

  WarehouseBloc({required WarehouseRepository warehouseRepository})
      : _warehouseRepository = warehouseRepository,
        super(WarehouseInitial()) {
    on<WarehouseLoadList>(onWarehouseLoad);
    on<WarehouseSave>(onWarehouseSave);
    on<WarehouseUpdate>(onWarehouseUpdate);
    on<WarehouseDelete>(onWarehouseDelete);
    on<WarehouseDeleteMany>(onWarehouseDeleteMany);
    on<WarehouseGet>(onWarehouseGet);
    on<WarehouseGetByCode>(onWarehouseGetByCode);
  }

  void onWarehouseLoad(
      WarehouseLoadList event, Emitter<WarehouseState> emit) async {
    emit(WarehouseInProgress());

    try {
      final results = await _warehouseRepository.getWarehouseList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<WarehouseModel> warehouses = (results.data as List)
            .map((warehouse) => WarehouseModel.fromJson(warehouse))
            .toList();
        // print(warehouses.length);
        emit(WarehouseLoadSuccess(warehouses: warehouses));
      } else {
        emit(const WarehouseLoadFailed(message: 'Warehouse Not Found'));
      }
    } catch (e) {
      emit(WarehouseLoadFailed(message: e.toString()));
    }
  }

  void onWarehouseDelete(
      WarehouseDelete event, Emitter<WarehouseState> emit) async {
    emit(WarehouseDeleteInProgress());
    try {
      await _warehouseRepository.deleteWarehouse(event.guid);

      emit(WarehouseDeleteSuccess());
    } catch (e) {
      // emit(WarehouseDeleteFailure(message: e.toString()));
    }
  }

  void onWarehouseDeleteMany(
      WarehouseDeleteMany event, Emitter<WarehouseState> emit) async {
    emit(WarehouseDeleteManyInProgress());
    try {
      await _warehouseRepository.deleteWarehouseMany(event.guid);

      emit(WarehouseDeleteManySuccess());
    } catch (e) {
      // emit(WarehouseDeleteFailure(message: e.toString()));
    }
  }

  void onWarehouseSave(
      WarehouseSave event, Emitter<WarehouseState> emit) async {
    emit(WarehouseSaveInProgress());
    try {
      await _warehouseRepository.saveWarehouse(event.warehouseModel);
      emit(WarehouseSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(WarehouseSaveFailed(message: error['message']));
    }
  }

  void onWarehouseUpdate(
      WarehouseUpdate event, Emitter<WarehouseState> emit) async {
    emit(WarehouseUpdateInProgress());
    try {
      await _warehouseRepository.updateWarehouse(
          event.guid, event.warehouseModel);
      emit(WarehouseUpdateSuccess());
    } catch (e) {
      emit(WarehouseUpdateFailed(message: e.toString()));
    }
  }

  void onWarehouseGet(WarehouseGet event, Emitter<WarehouseState> emit) async {
    emit(WarehouseGetInProgress());
    try {
      final result = await _warehouseRepository.getWarehouse(event.guid);
      if (result.success) {
        WarehouseModel warehouse = WarehouseModel.fromJson(result.data);
        emit(WarehouseGetSuccess(warehouse: warehouse));
      } else {
        emit(const WarehouseGetFailed(message: 'Warehouse Not Found'));
      }
    } catch (e) {
      // emit(WarehouseDeleteFailure(message: e.toString()));
    }
  }

  void onWarehouseGetByCode(
      WarehouseGetByCode event, Emitter<WarehouseState> emit) async {
    emit(WarehouseGetInProgress());
    try {
      final result = await _warehouseRepository.getWarehouseByCode(event.code);
      if (result.success) {
        WarehouseModel warehouse = WarehouseModel.fromJson(result.data);
        emit(WarehouseGetSuccess(warehouse: warehouse));
      } else {
        emit(const WarehouseGetFailed(message: 'Warehouse Not Found'));
      }
    } catch (e) {
      // emit(WarehouseDeleteFailure(message: e.toString()));
    }
  }
}
