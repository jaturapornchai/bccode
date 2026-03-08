part of 'device_print_bloc.dart';

abstract class DevicePrintEvent extends Equatable {
  const DevicePrintEvent();

  @override
  List<Object> get props => [];
}

class DevicePrintGet extends DevicePrintEvent {
  final String guid;

  const DevicePrintGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DevicePrintLoadList extends DevicePrintEvent {
  final int limit;
  final int offset;
  final String search;

  const DevicePrintLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class DevicePrintDelete extends DevicePrintEvent {
  final String guid;

  const DevicePrintDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DevicePrintDeleteMany extends DevicePrintEvent {
  final List<String> guid;

  const DevicePrintDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DevicePrintSave extends DevicePrintEvent {
  final DevicePrinterModel devicePrint;

  const DevicePrintSave({
    required this.devicePrint,
  });

  @override
  List<Object> get props => [devicePrint];
}

class DevicePrintUpdate extends DevicePrintEvent {
  final String guid;
  final DevicePrinterModel devicePrint;

  const DevicePrintUpdate({
    required this.guid,
    required this.devicePrint,
  });

  @override
  List<Object> get props => [devicePrint];
}
