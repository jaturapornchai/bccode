part of 'inventory_costing_bloc.dart';

abstract class InventoryCostingEvent extends Equatable {
  const InventoryCostingEvent();

  @override
  List<Object> get props => [];
}

// === Config Events ===

class InventoryCostingLoadConfig extends InventoryCostingEvent {
  final String itemCode;

  const InventoryCostingLoadConfig({required this.itemCode});

  @override
  List<Object> get props => [itemCode];
}

class InventoryCostingSaveConfig extends InventoryCostingEvent {
  final ProductCostingConfigModel config;

  const InventoryCostingSaveConfig({required this.config});

  @override
  List<Object> get props => [config];
}

// === Cost Layers Events ===

class InventoryCostingLoadLayers extends InventoryCostingEvent {
  final String itemCode;
  final String whCode;

  const InventoryCostingLoadLayers({required this.itemCode, this.whCode = ''});

  @override
  List<Object> get props => [itemCode, whCode];
}

// === Report Events ===

class InventoryCostingLoadValuation extends InventoryCostingEvent {
  const InventoryCostingLoadValuation();
}

class InventoryCostingLoadStockCard extends InventoryCostingEvent {
  final String itemCode;
  final String? fromDate;
  final String? toDate;

  const InventoryCostingLoadStockCard({
    required this.itemCode,
    this.fromDate,
    this.toDate,
  });

  @override
  List<Object> get props => [itemCode];
}

// === Database Events ===

class InventoryCostingCreateTables extends InventoryCostingEvent {
  const InventoryCostingCreateTables();
}
