import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/inventory_repository.dart';
import 'package:smlaicloud/model/inventory_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'inventory_event.dart';
part 'inventory_state.dart';

class InventoryBloc extends Bloc<InventoryEvent, InventoryState> {
  final InventoryRepository _inventoryRepository;

  InventoryBloc({required InventoryRepository inventoryRepository})
    : _inventoryRepository = inventoryRepository,
      super(InventoryInitial()) {
    on<ListInventoryLoad>(_onListInventoryLoad);
    on<ListInventoryById>(_onGetInventoryId);
    on<InventorySaved>(_onInventorySaved);
    on<InventoryUpdate>(_onInventoryUpdate);
    on<InventoryDelete>(_onInventoryDelete);
  }
  void _onListInventoryLoad(
    ListInventoryLoad event,
    Emitter<InventoryState> emit,
  ) async {
    InventoryLoadSuccess inventoryLoadSuccess;
    List<InventoryModel> previousInventory = [];

    if (state is InventoryLoadSuccess) {
      inventoryLoadSuccess = (state as InventoryLoadSuccess).copyWith();
      previousInventory = inventoryLoadSuccess.inventory;
    }

    emit(InventoryInProgress());

    try {
      final result = await _inventoryRepository.getInventoryList(
        perPage: event.perPage,
        page: event.page,
        search: event.search,
      );

      AppLogger.debug('📋 LIST: API result success: ${result.success}');

      if (result.success) {
        AppLogger.debug('📋 LIST: Data count: ${(result.data as List).length}');

        if (event.nextPage) {
          List<InventoryModel> inventory = [];
          int index = 0;
          for (var item in (result.data as List)) {
            try {
              inventory.add(InventoryModel.fromJson(item));
            } catch (e, stackTrace) {
              AppLogger.error('❌ ERROR parsing inventory item #$index: $e');
              AppLogger.error('Item data: $item');
              AppLogger.error('Stack trace: $stackTrace');
            }
            index++;
          }

          AppLogger.debug(
            '✅ SUCCESS: Parsed ${inventory.length} inventory items',
          );
          emit(InventoryLoadSuccess(inventory: inventory, page: result.page));
        } else {
          List<InventoryModel> inventory = [];
          int index = 0;
          for (var item in (result.data as List)) {
            try {
              inventory.add(InventoryModel.fromJson(item));
            } catch (e, stackTrace) {
              AppLogger.error('❌ ERROR parsing inventory item #$index: $e');
              AppLogger.error('Item data: $item');
              AppLogger.error('Stack trace: $stackTrace');
            }
            index++;
          }

          AppLogger.debug(
            '✅ SUCCESS: Parsed ${inventory.length} more inventory items',
          );
          previousInventory.addAll(inventory);
          emit(
            InventoryLoadSuccess(
              inventory: previousInventory,
              page: result.page,
            ),
          );
        }
      } else {
        AppLogger.warning('⚠️ API returned success=false');
        emit(InventoryLoadFailed(message: 'Inventory Not Found'));
      }
    } catch (e, stackTrace) {
      AppLogger.error('❌ ERROR in _onListInventoryLoad: $e');
      AppLogger.error('Stack trace: $stackTrace');
      emit(InventoryLoadFailed(message: e.toString()));
    }
  }

  void _onGetInventoryId(
    ListInventoryById event,
    Emitter<InventoryState> emit,
  ) async {
    emit(InventorySearchInProgress());
    try {
      final result = await _inventoryRepository.getInventoryId(event.id);

      if (result.success) {
        // Debug: ตรวจสอบข้อมูลดิบจาก API
        AppLogger.debug('🔍 DEBUG: Raw data from API: ${result.data}');
        AppLogger.debug('🔍 DEBUG: options field: ${result.data['options']}');
        AppLogger.debug('🔍 DEBUG: images field: ${result.data['images']}');

        InventoryModel inventory = InventoryModel.fromJson(result.data);

        AppLogger.debug('✅ SUCCESS: Parsed inventory model');
        AppLogger.debug('🔍 DEBUG: Inventory options: ${inventory.options}');
        AppLogger.debug('🔍 DEBUG: Inventory images: ${inventory.images}');

        emit(InventorySearchLoadSuccess(inventory: inventory));
      } else {
        emit(InventorySearchLoadFailed(message: 'Product Not Found'));
      }
    } catch (e, stackTrace) {
      AppLogger.error('❌ ERROR in _onGetInventoryId: $e');
      AppLogger.error('Stack trace: $stackTrace');
      emit(InventorySearchLoadFailed(message: e.toString()));
    }
  }

  void _onInventorySaved(
    InventorySaved event,
    Emitter<InventoryState> emit,
  ) async {
    emit(InventoryFormSaveInProgress());
    try {
      // // print(event.inventory.toString());

      await _inventoryRepository.saveInventory(event.inventory);

      emit(InventoryFormSaveSuccess());
    } catch (e) {
      emit(InventoryFormSaveFailure(message: e.toString()));
    }
  }

  void _onInventoryUpdate(
    InventoryUpdate event,
    Emitter<InventoryState> emit,
  ) async {
    emit(InventoryUpdateInProgress());
    try {
      // // print(event.inventory.toString());

      await _inventoryRepository.updateInventory(event.inventory);

      emit(InventoryUpdateSuccess());
    } catch (e) {
      emit(InventoryUpdateFailure(message: e.toString()));
    }
  }

  void _onInventoryDelete(
    InventoryDelete event,
    Emitter<InventoryState> emit,
  ) async {
    emit(InventoryDeleteInProgress());
    try {
      // // print(event.inventory.toString());

      await _inventoryRepository.deleteInventory(event.id);

      emit(InventoryDeleteSuccess());
    } catch (e) {
      emit(InventoryDeleteFailure(message: e.toString()));
    }
  }
}
