import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/inventory_costing_model.dart';
import 'package:smlaicloud/repositories/inventory_costing_repository.dart';

part 'inventory_costing_event.dart';
part 'inventory_costing_state.dart';

class InventoryCostingBloc
    extends Bloc<InventoryCostingEvent, InventoryCostingState> {
  final InventoryCostingRepository _repository;

  InventoryCostingBloc({required InventoryCostingRepository repository})
      : _repository = repository,
        super(InventoryCostingInitial()) {
    on<InventoryCostingLoadConfig>(_onLoadConfig);
    on<InventoryCostingSaveConfig>(_onSaveConfig);
    on<InventoryCostingLoadLayers>(_onLoadLayers);
    on<InventoryCostingLoadValuation>(_onLoadValuation);
    on<InventoryCostingLoadStockCard>(_onLoadStockCard);
    on<InventoryCostingCreateTables>(_onCreateTables);
  }

  Future<void> _onLoadConfig(InventoryCostingLoadConfig event,
      Emitter<InventoryCostingState> emit) async {
    emit(InventoryCostingInProgress());
    try {
      final result = await _repository.getCostingConfig(event.itemCode);
      final config = ProductCostingConfigModel.fromJson(result);
      emit(InventoryCostingConfigLoaded(config: config));
    } catch (e) {
      emit(InventoryCostingFailed(message: e.toString()));
    }
  }

  Future<void> _onSaveConfig(InventoryCostingSaveConfig event,
      Emitter<InventoryCostingState> emit) async {
    emit(InventoryCostingConfigSaveInProgress());
    try {
      await _repository.updateCostingConfig(
          event.config.itemCode, event.config.toJson());
      emit(InventoryCostingConfigSaveSuccess());
    } catch (e) {
      emit(InventoryCostingFailed(message: e.toString()));
    }
  }

  Future<void> _onLoadLayers(InventoryCostingLoadLayers event,
      Emitter<InventoryCostingState> emit) async {
    emit(InventoryCostingInProgress());
    try {
      final result =
          await _repository.getCostLayers(event.itemCode, whCode: event.whCode);
      final layers = result
          .map((e) =>
              InventoryCostLayerModel.fromJson(e as Map<String, dynamic>))
          .toList();
      emit(InventoryCostingLayersLoaded(layers: layers));
    } catch (e) {
      emit(InventoryCostingFailed(message: e.toString()));
    }
  }

  Future<void> _onLoadValuation(InventoryCostingLoadValuation event,
      Emitter<InventoryCostingState> emit) async {
    emit(InventoryCostingInProgress());
    try {
      final result = await _repository.getInventoryValuation();
      final report = InventoryValuationReportModel.fromJson(result);
      emit(InventoryCostingValuationLoaded(report: report));
    } catch (e) {
      emit(InventoryCostingFailed(message: e.toString()));
    }
  }

  Future<void> _onLoadStockCard(InventoryCostingLoadStockCard event,
      Emitter<InventoryCostingState> emit) async {
    emit(InventoryCostingInProgress());
    try {
      final result = await _repository.getStockCard(event.itemCode,
          fromDate: event.fromDate, toDate: event.toDate);
      final report = StockCardReportModel.fromJson(result);
      emit(InventoryCostingStockCardLoaded(report: report));
    } catch (e) {
      emit(InventoryCostingFailed(message: e.toString()));
    }
  }

  Future<void> _onCreateTables(InventoryCostingCreateTables event,
      Emitter<InventoryCostingState> emit) async {
    emit(InventoryCostingInProgress());
    try {
      await _repository.createTables();
      emit(InventoryCostingCreateTablesSuccess());
    } catch (e) {
      emit(InventoryCostingFailed(message: e.toString()));
    }
  }
}
