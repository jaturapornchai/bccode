part of 'warehouse_location_bloc.dart';

abstract class WarehouseLocationEvent extends Equatable {
  const WarehouseLocationEvent();

  @override
  List<Object> get props => [];
}

class WarehouseLoadLocationList extends WarehouseLocationEvent {
  final int limit;
  final int offset;
  final String search;

  const WarehouseLoadLocationList(
      {required this.offset, required this.limit, required this.search});
}

@override
List<Object> get props => [];

class WarehouseLocationGetByCode extends WarehouseLocationEvent {
  final String warehousecode;
  final String locationcode;

  const WarehouseLocationGetByCode(
      {required this.warehousecode, required this.locationcode});

  @override
  List<Object> get props => [warehousecode, locationcode];
}

class WarehouseLocationUpdate extends WarehouseLocationEvent {
  final String warehousecode;
  final String locationcode;
  final WarehouseLocationUpdateModel warehouseLocationUpdateModel;

  const WarehouseLocationUpdate({
    required this.warehousecode,
    required this.locationcode,
    required this.warehouseLocationUpdateModel,
  });

  @override
  List<Object> get props => [warehouseLocationUpdateModel];
}

class WarehouseLocationDeleteMany extends WarehouseLocationEvent {
  final String warehousecode;
  final List<String> locationcode;

  const WarehouseLocationDeleteMany({
    required this.warehousecode,
    required this.locationcode,
  });

  @override
  List<Object> get props => [locationcode];
}
