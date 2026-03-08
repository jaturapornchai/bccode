import 'package:smlaicloud/model/warehouse_location_model.dart';
import 'package:smlaicloud/model/warehouse_location_update_model.dart';
import 'package:smlaicloud/repositories/warehouse_location_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'warehouse_location_event.dart';
part 'warehouse_location_state.dart';

class WarehouseLocationBloc
    extends Bloc<WarehouseLocationEvent, WarehouseLocationState> {
  final WarehouseLocationRepository _warehouseLocationRepository;

  WarehouseLocationBloc(
      {required WarehouseLocationRepository warehouseLocationRepository})
      : _warehouseLocationRepository = warehouseLocationRepository,
        super(WarehouseLocationInitial()) {
    on<WarehouseLoadLocationList>(onWarehouseLoadLocation);
    on<WarehouseLocationGetByCode>(onWarehouseGetByCode);
    on<WarehouseLocationUpdate>(onWarehouseLocationUpdate);
    on<WarehouseLocationDeleteMany>(onWarehouseLocationDeleteMany);
  }

  void onWarehouseLoadLocation(WarehouseLoadLocationList event,
      Emitter<WarehouseLocationState> emit) async {
    emit(WarehouseLocationInProgress());

    try {
      final results =
          await _warehouseLocationRepository.getWarehouseLocationList(
              offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<WarehouseLocationModel> warehouses = (results.data as List)
            .map((warehouse) => WarehouseLocationModel.fromJson(warehouse))
            .toList();
        // print(warehouses.length);
        emit(WarehouseLocationLoadSuccess(warehouses: warehouses));
      } else {
        emit(const WarehouseLocationLoadFailed(message: 'Warehouse Not Found'));
      }
    } catch (e) {
      emit(WarehouseLocationLoadFailed(message: e.toString()));
    }
  }

  /// **Load data by warehousecode and locationcode
  void onWarehouseGetByCode(WarehouseLocationGetByCode event,
      Emitter<WarehouseLocationState> emit) async {
    emit(WarehouseLocationGetInProgress());
    try {
      final result = await _warehouseLocationRepository
          .getWarehouseLocationByCode(event.warehousecode, event.locationcode);
      if (result.success) {
        WarehouseLocationModel warehouselocation =
            WarehouseLocationModel.fromJson(result.data);
        emit(WarehouseLocationGetSuccess(warehouselocation: warehouselocation));
      } else {
        emit(const WarehouseLocationGetFailed(message: 'Location Not Found'));
      }
    } catch (e) {
      emit(WarehouseLocationGetFailed(message: e.toString()));
    }
  }

  /// *** Update Location
  void onWarehouseLocationUpdate(WarehouseLocationUpdate event,
      Emitter<WarehouseLocationState> emit) async {
    emit(WarehouseLocationUpdateInProgress());
    try {
      await _warehouseLocationRepository.updateWarehouseLocation(
          event.warehousecode,
          event.locationcode,
          event.warehouseLocationUpdateModel);
      emit(WarehouseLocationUpdateSuccess());
    } catch (e) {
      emit(WarehouseLocationUpdateFailed(message: e.toString()));
    }
  }

  /// **Delete Location By Many
  void onWarehouseLocationDeleteMany(WarehouseLocationDeleteMany event,
      Emitter<WarehouseLocationState> emit) async {
    emit(WarehouseLocationDeleteManyInProgress());
    try {
      await _warehouseLocationRepository.deleteWarehouseLocationMany(
          event.warehousecode, event.locationcode);

      emit(WarehouseLocationDeleteManySuccess());
    } catch (e) {
      emit(WarehouseLocationDeleteManyFailed(message: e.toString()));
    }
  }
}
