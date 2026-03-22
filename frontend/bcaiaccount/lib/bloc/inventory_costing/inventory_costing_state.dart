part of 'inventory_costing_bloc.dart';

abstract class InventoryCostingState extends Equatable {
  const InventoryCostingState();

  @override
  List<Object> get props => [];
}

class InventoryCostingInitial extends InventoryCostingState {}

class InventoryCostingInProgress extends InventoryCostingState {}

// === Config States ===

class InventoryCostingConfigLoaded extends InventoryCostingState {
  final ProductCostingConfigModel config;

  const InventoryCostingConfigLoaded({required this.config});

  @override
  List<Object> get props => [config];
}

class InventoryCostingConfigSaveSuccess extends InventoryCostingState {}

class InventoryCostingConfigSaveInProgress extends InventoryCostingState {}

class InventoryCostingFailed extends InventoryCostingState {
  final String message;

  const InventoryCostingFailed({required this.message});

  @override
  List<Object> get props => [message];
}

// === Cost Layers States ===

class InventoryCostingLayersLoaded extends InventoryCostingState {
  final List<InventoryCostLayerModel> layers;

  const InventoryCostingLayersLoaded({required this.layers});

  @override
  List<Object> get props => [layers];
}

// === Report States ===

class InventoryCostingValuationLoaded extends InventoryCostingState {
  final InventoryValuationReportModel report;

  const InventoryCostingValuationLoaded({required this.report});

  @override
  List<Object> get props => [report];
}

class InventoryCostingStockCardLoaded extends InventoryCostingState {
  final StockCardReportModel report;

  const InventoryCostingStockCardLoaded({required this.report});

  @override
  List<Object> get props => [report];
}

// === Database States ===

class InventoryCostingCreateTablesSuccess extends InventoryCostingState {}
