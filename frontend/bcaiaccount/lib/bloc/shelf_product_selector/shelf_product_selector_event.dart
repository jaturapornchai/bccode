part of 'shelf_product_selector_bloc.dart';

abstract class ShelfProductSelectorEvent extends Equatable {
  const ShelfProductSelectorEvent();

  @override
  List<Object?> get props => [];
}

class LoadWarehouses extends ShelfProductSelectorEvent {
  const LoadWarehouses();

  @override
  List<Object?> get props => [];
}

class LoadLocations extends ShelfProductSelectorEvent {
  const LoadLocations();

  @override
  List<Object?> get props => [];
}

class SelectWarehouse extends ShelfProductSelectorEvent {
  final WarehouseModel warehouse;

  const SelectWarehouse({required this.warehouse});

  @override
  List<Object?> get props => [warehouse];
}

class SelectLocation extends ShelfProductSelectorEvent {
  final WarehouseLocationModel location;

  const SelectLocation({required this.location});

  @override
  List<Object?> get props => [location];
}

class SelectShelf extends ShelfProductSelectorEvent {
  final ShelfModel shelf;

  const SelectShelf({required this.shelf});

  @override
  List<Object?> get props => [shelf];
}

class LoadShelfProducts extends ShelfProductSelectorEvent {
  const LoadShelfProducts();

  @override
  List<Object?> get props => [];
}

class ResetSelection extends ShelfProductSelectorEvent {
  const ResetSelection();
}
