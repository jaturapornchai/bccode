part of 'printer_bloc.dart';

abstract class PrinterState extends Equatable {
  const PrinterState();

  @override
  List<Object> get props => [];
}

class PrinterInitial extends PrinterState {}

class PrinterInProgress extends PrinterState {}

class PrinterLoadSuccess extends PrinterState {
  final List<PrinterModel> printers;

  const PrinterLoadSuccess({required this.printers});

  PrinterLoadSuccess copyWith({
    List<PrinterModel>? printers,
  }) =>
      PrinterLoadSuccess(printers: printers ?? this.printers);

  @override
  List<Object> get props => [printers];
}

class PrinterLoadFailed extends PrinterState {
  final String message;

  const PrinterLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PrinterSaveInitial extends PrinterState {}

class PrinterSaveInProgress extends PrinterState {}

class PrinterSaveSuccess extends PrinterState {}

class PrinterSaveFailed extends PrinterState {
  final String message;

  const PrinterSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PrinterDeleteInProgress extends PrinterState {}

class PrinterDeleteSuccess extends PrinterState {}

class PrinterDeleteFailed extends PrinterState {}

class PrinterDeleteManyInProgress extends PrinterState {}

class PrinterDeleteManySuccess extends PrinterState {}

class PrinterDeleteManyFailed extends PrinterState {}

class PrinterGetInProgress extends PrinterState {}

class PrinterGetSuccess extends PrinterState {
  final PrinterModel printer;

  const PrinterGetSuccess({required this.printer});

  PrinterGetSuccess copyWith({
    PrinterModel? printer,
  }) =>
      PrinterGetSuccess(printer: printer ?? this.printer);

  @override
  List<Object> get props => [printer];
}

class PrinterGetFailed extends PrinterState {
  final String message;

  const PrinterGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PrinterUpdateInitial extends PrinterState {}

class PrinterUpdateInProgress extends PrinterState {}

class PrinterUpdateSuccess extends PrinterState {}

class PrinterUpdateFailed extends PrinterState {
  final String message;

  const PrinterUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
