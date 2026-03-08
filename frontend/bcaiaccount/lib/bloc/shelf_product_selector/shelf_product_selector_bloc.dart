import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/model/warehouse_location_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:smlaicloud/model/shelf_product_model.dart';
import 'package:smlaicloud/repositories/warehouse_repository.dart';

part 'shelf_product_selector_event.dart';
part 'shelf_product_selector_state.dart';

class ShelfProductSelectorBloc
    extends Bloc<ShelfProductSelectorEvent, ShelfProductSelectorState> {
  final WarehouseRepository _warehouseRepository;

  ShelfProductSelectorBloc({required WarehouseRepository warehouseRepository})
    : _warehouseRepository = warehouseRepository,
      super(ShelfProductSelectorInitial()) {
    on<LoadWarehouses>(_onLoadWarehouses);
    on<LoadLocations>(_onLoadLocations);
    on<SelectWarehouse>(_onSelectWarehouse);
    on<SelectLocation>(_onSelectLocation);
    on<SelectShelf>(_onSelectShelf);
    on<LoadShelfProducts>(_onLoadShelfProducts);
    on<ResetSelection>(_onResetSelection);
  }

  void _onLoadWarehouses(
    LoadWarehouses event,
    Emitter<ShelfProductSelectorState> emit,
  ) async {
    emit(
      state.copyWith(
        warehouseLoadStatus: LoadingStatus.loading,
        warehouses: [],
        // Clear selections when loading warehouses
        clearSelectedWarehouse: true,
        clearSelectedLocation: true,
        clearSelectedShelf: true,
        locations: [],
        shelves: [],
        shelfProducts: [],
      ),
    );

    try {
      final results = await _warehouseRepository.getWarehouseList(
        offset: 0,
        limit: 1000,
        search: '',
      );

      if (results.success) {
        List<WarehouseModel> warehouses = (results.data as List)
            .map((warehouse) => WarehouseModel.fromJson(warehouse))
            .toList();

        emit(
          state.copyWith(
            warehouseLoadStatus: LoadingStatus.success,
            warehouses: warehouses,
          ),
        );
      } else {
        emit(
          state.copyWith(
            warehouseLoadStatus: LoadingStatus.error,
            errorMessage: 'Warehouse Not Found',
          ),
        );
      }
    } catch (e) {
      emit(
        state.copyWith(
          warehouseLoadStatus: LoadingStatus.error,
          errorMessage: e.toString(),
        ),
      );
    }
  }

  void _onLoadLocations(
    LoadLocations event,
    Emitter<ShelfProductSelectorState> emit,
  ) async {
    if (state.selectedWarehouse == null) return;

    emit(
      state.copyWith(locationLoadStatus: LoadingStatus.loading, locations: []),
    );

    try {
      // ใช้ข้อมูล location จาก warehouse ที่เลือกแล้ว
      final selectedWarehouse = state.selectedWarehouse!;

      // แปลง LocationModel เป็น WarehouseLocationModel
      List<WarehouseLocationModel> locations = selectedWarehouse.location.map((
        location,
      ) {
        return WarehouseLocationModel(
          guidfixed:
              '${selectedWarehouse.guidfixed}_${location.code}', // สร้าง guid unique
          warehousecode: selectedWarehouse.code,
          warehousenames: selectedWarehouse.names,
          locationcode: location.code,
          locationnames: location.names,
          shelf: location.shelf,
        );
      }).toList();

      emit(
        state.copyWith(
          locationLoadStatus: LoadingStatus.success,
          locations: locations,
        ),
      );
    } catch (e) {
      emit(
        state.copyWith(
          locationLoadStatus: LoadingStatus.error,
          errorMessage: e.toString(),
        ),
      );
    }
  }

  void _onSelectWarehouse(
    SelectWarehouse event,
    Emitter<ShelfProductSelectorState> emit,
  ) {
    emit(
      state.copyWith(
        selectedWarehouse: event.warehouse,
        clearSelectedLocation: true,
        clearSelectedShelf: true,
        locations: [],
        shelves: [],
        shelfProducts: [],
      ),
    );

    // Load locations for the selected warehouse
    add(const LoadLocations());
  }

  void _onSelectLocation(
    SelectLocation event,
    Emitter<ShelfProductSelectorState> emit,
  ) {
    emit(
      state.copyWith(
        selectedLocation: event.location,
        clearSelectedShelf: true,
        shelves: event.location.shelf,
        shelfProducts: [],
      ),
    );
  }

  void _onSelectShelf(
    SelectShelf event,
    Emitter<ShelfProductSelectorState> emit,
  ) {
    emit(state.copyWith(selectedShelf: event.shelf, shelfProducts: []));

    // Load products for the selected shelf
    add(const LoadShelfProducts());
  }

  void _onLoadShelfProducts(
    LoadShelfProducts event,
    Emitter<ShelfProductSelectorState> emit,
  ) async {
    emit(
      state.copyWith(
        productLoadStatus: LoadingStatus.loading,
        shelfProducts: [],
      ),
    );

    try {
      // ใช้ข้อมูล products จาก shelf ที่เลือกแล้ว
      if (state.selectedShelf != null) {
        List<ShelfProductItemModel> products =
            state.selectedShelf!.productitems ?? [];

        emit(
          state.copyWith(
            productLoadStatus: LoadingStatus.success,
            shelfProducts: products,
          ),
        );
      } else {
        emit(
          state.copyWith(
            productLoadStatus: LoadingStatus.error,
            errorMessage: 'No shelf selected',
          ),
        );
      }
    } catch (e) {
      emit(
        state.copyWith(
          productLoadStatus: LoadingStatus.error,
          errorMessage: e.toString(),
        ),
      );
    }
  }

  void _onResetSelection(
    ResetSelection event,
    Emitter<ShelfProductSelectorState> emit,
  ) {
    // Reset selections but keep warehouses if already loaded
    final newState = state.copyWith(
      clearSelectedWarehouse: true,
      clearSelectedLocation: true,
      clearSelectedShelf: true,
      locations: [],
      shelves: [],
      shelfProducts: [],
      locationLoadStatus: LoadingStatus.initial,
      productLoadStatus: LoadingStatus.initial,
    );
    emit(newState);
  }
}
