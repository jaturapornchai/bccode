part of 'shelf_product_selector_bloc.dart';

enum LoadingStatus { initial, loading, success, error }

class ShelfProductSelectorState extends Equatable {
  final LoadingStatus warehouseLoadStatus;
  final LoadingStatus locationLoadStatus;
  final LoadingStatus productLoadStatus;

  final List<WarehouseModel> warehouses;
  final List<WarehouseLocationModel> locations;
  final List<ShelfModel> shelves;
  final List<ShelfProductItemModel> shelfProducts;

  final WarehouseModel? selectedWarehouse;
  final WarehouseLocationModel? selectedLocation;
  final ShelfModel? selectedShelf;

  final String? errorMessage;

  const ShelfProductSelectorState({
    this.warehouseLoadStatus = LoadingStatus.initial,
    this.locationLoadStatus = LoadingStatus.initial,
    this.productLoadStatus = LoadingStatus.initial,
    this.warehouses = const [],
    this.locations = const [],
    this.shelves = const [],
    this.shelfProducts = const [],
    this.selectedWarehouse,
    this.selectedLocation,
    this.selectedShelf,
    this.errorMessage,
  });

  ShelfProductSelectorState copyWith({
    LoadingStatus? warehouseLoadStatus,
    LoadingStatus? locationLoadStatus,
    LoadingStatus? productLoadStatus,
    List<WarehouseModel>? warehouses,
    List<WarehouseLocationModel>? locations,
    List<ShelfModel>? shelves,
    List<ShelfProductItemModel>? shelfProducts,
    WarehouseModel? selectedWarehouse,
    WarehouseLocationModel? selectedLocation,
    ShelfModel? selectedShelf,
    String? errorMessage,
    bool clearSelectedWarehouse = false,
    bool clearSelectedLocation = false,
    bool clearSelectedShelf = false,
  }) {
    return ShelfProductSelectorState(
      warehouseLoadStatus: warehouseLoadStatus ?? this.warehouseLoadStatus,
      locationLoadStatus: locationLoadStatus ?? this.locationLoadStatus,
      productLoadStatus: productLoadStatus ?? this.productLoadStatus,
      warehouses: warehouses ?? this.warehouses,
      locations: locations ?? this.locations,
      shelves: shelves ?? this.shelves,
      shelfProducts: shelfProducts ?? this.shelfProducts,
      selectedWarehouse: clearSelectedWarehouse ? null : (selectedWarehouse ?? this.selectedWarehouse),
      selectedLocation: clearSelectedLocation ? null : (selectedLocation ?? this.selectedLocation),
      selectedShelf: clearSelectedShelf ? null : (selectedShelf ?? this.selectedShelf),
      errorMessage: errorMessage ?? this.errorMessage,
    );
  }

  @override
  List<Object?> get props => [
        warehouseLoadStatus,
        locationLoadStatus,
        productLoadStatus,
        warehouses,
        locations,
        shelves,
        shelfProducts,
        selectedWarehouse,
        selectedLocation,
        selectedShelf,
        errorMessage,
      ];

  @override
  String toString() {
    return 'ShelfProductSelectorState(selectedWarehouse: ${selectedWarehouse?.code}, selectedLocation: ${selectedLocation?.locationcode}, selectedShelf: ${selectedShelf?.code})';
  }
}

class ShelfProductSelectorInitial extends ShelfProductSelectorState {
  const ShelfProductSelectorInitial() : super();
}
