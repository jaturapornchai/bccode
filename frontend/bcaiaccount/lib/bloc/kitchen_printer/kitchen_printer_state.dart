part of '../kitchen_printer/kitchen_printer_bloc.dart';

abstract class KitchenPrinterState extends Equatable {
  const KitchenPrinterState();

  @override
  List<Object> get props => [];
}

class KitchenPrinterInitial extends KitchenPrinterState {}

class KitchenPrinterInProgress extends KitchenPrinterState {}

class KitchenPrinterLoadSuccess extends KitchenPrinterState {
  final List<KitchenPrinterSaveModel> kitchenPrinters;

  const KitchenPrinterLoadSuccess({required this.kitchenPrinters});

  KitchenPrinterLoadSuccess copyWith({
    String guid = '',
    List<KitchenPrinterSaveModel>? kitchenPrinters,
  }) =>
      KitchenPrinterLoadSuccess(
          kitchenPrinters: kitchenPrinters ?? this.kitchenPrinters);

  @override
  List<Object> get props => [kitchenPrinters];
}

class KitchenPrinterLoadFailed extends KitchenPrinterState {
  final String message;

  const KitchenPrinterLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenPrinterSaveInitial extends KitchenPrinterState {}

class KitchenPrinterSaveInProgress extends KitchenPrinterState {}

class KitchenPrinterSaveSuccess extends KitchenPrinterState {}

class KitchenPrinterSaveFailed extends KitchenPrinterState {
  final String message;

  const KitchenPrinterSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenPrinterDeleteInProgress extends KitchenPrinterState {}

class KitchenPrinterDeleteSuccess extends KitchenPrinterState {}

class KitchenPrinterDeleteFailed extends KitchenPrinterState {}

class KitchenPrinterDeleteManyInProgress extends KitchenPrinterState {}

class KitchenPrinterDeleteManySuccess extends KitchenPrinterState {}

class KitchenPrinterDeleteManyFailed extends KitchenPrinterState {}

class KitchenPrinterGetInProgress extends KitchenPrinterState {}

class KitchenPrinterGetSuccess extends KitchenPrinterState {
  final KitchenPrinterModel kitchenPrinters;

  const KitchenPrinterGetSuccess({required this.kitchenPrinters});

  KitchenPrinterGetSuccess copyWith({
    KitchenPrinterModel? kitchenPrinters,
  }) =>
      KitchenPrinterGetSuccess(
          kitchenPrinters: kitchenPrinters ?? this.kitchenPrinters);

  @override
  List<Object> get props => [kitchenPrinters];
}

class KitchenPrinterGetFailed extends KitchenPrinterState {
  final String message;

  const KitchenPrinterGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenPrinterUpdateInitial extends KitchenPrinterState {}

class KitchenPrinterUpdateInProgress extends KitchenPrinterState {}

class KitchenPrinterUpdateSuccess extends KitchenPrinterState {}

class KitchenPrinterUpdateFailed extends KitchenPrinterState {
  final String message;

  const KitchenPrinterUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
