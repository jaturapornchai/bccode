part of '../kitchen_printer/kitchen_printer_bloc.dart';

abstract class KitchenPrinterEvent extends Equatable {
  const KitchenPrinterEvent();

  @override
  List<Object> get props => [];
}

class KitchenPrinterGet extends KitchenPrinterEvent {
  final String guid;

  const KitchenPrinterGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class KitchenPrinterLoadList extends KitchenPrinterEvent {
  final int limit;
  final int offset;
  final String search;

  const KitchenPrinterLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class KitchenPrinterDelete extends KitchenPrinterEvent {
  final String guid;

  const KitchenPrinterDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class KitchenPrinterDeleteMany extends KitchenPrinterEvent {
  final List<String> guid;

  const KitchenPrinterDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class KitchenPrinterSave extends KitchenPrinterEvent {
  final KitchenPrinterModel kitchenPrinter;

  const KitchenPrinterSave({
    required this.kitchenPrinter,
  });

  @override
  List<Object> get props => [kitchenPrinter];
}

class KitchenPrinterUpdate extends KitchenPrinterEvent {
  final String guid;
  final KitchenPrinterModel kitchenPrinter;

  const KitchenPrinterUpdate({
    required this.guid,
    required this.kitchenPrinter,
  });

  @override
  List<Object> get props => [kitchenPrinter];
}
