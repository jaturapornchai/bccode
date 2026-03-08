part of 'devices_bloc.dart';

abstract class DevicesEvent extends Equatable {
  const DevicesEvent();

  @override
  List<Object> get props => [];
}

class DevicesGet extends DevicesEvent {
  final String guid;

  const DevicesGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DevicesLoadList extends DevicesEvent {
  final int limit;
  final int offset;
  final String search;

  const DevicesLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class DevicesDelete extends DevicesEvent {
  final String guid;

  const DevicesDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DevicesDeleteMany extends DevicesEvent {
  final List<String> guid;

  const DevicesDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DevicesSave extends DevicesEvent {
  final DevicesModel devices;

  const DevicesSave({
    required this.devices,
  });

  @override
  List<Object> get props => [devices];
}

class DevicesUpdate extends DevicesEvent {
  final String guid;
  final DevicesModel devices;

  const DevicesUpdate({
    required this.guid,
    required this.devices,
  });

  @override
  List<Object> get props => [devices];
}
