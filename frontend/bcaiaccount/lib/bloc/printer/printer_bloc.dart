import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/printer_model.dart';
import 'package:smlaicloud/repositories/printer_repository.dart';

part 'printer_event.dart';
part 'printer_state.dart';

class PrinterBloc extends Bloc<PrinterEvent, PrinterState> {
  final PrinterRepository _printerRepository;

  PrinterBloc({required PrinterRepository printerRepository})
      : _printerRepository = printerRepository,
        super(PrinterInitial()) {
    on<PrinterLoadList>(onPrinterLoad);
    on<PrinterSave>(onPrinterSave);
    on<PrinterUpdate>(onPrinterUpdate);
    on<PrinterDelete>(printerDelete);
    on<PrinterDeleteMany>(printerDeleteMany);
    on<PrinterGet>(onPrinterGet);
  }

  void onPrinterLoad(PrinterLoadList event, Emitter<PrinterState> emit) async {
    emit(PrinterInProgress());

    try {
      final results = await _printerRepository.getPrinterList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<PrinterModel> printers = (results.data as List)
            .map((printer) => PrinterModel.fromJson(printer))
            .toList();
        emit(PrinterLoadSuccess(printers: printers));
      } else {
        emit(const PrinterLoadFailed(message: 'Printer Group Not Found'));
      }
    } catch (e) {
      emit(PrinterLoadFailed(message: e.toString()));
    }
  }

  void printerDelete(PrinterDelete event, Emitter<PrinterState> emit) async {
    emit(PrinterDeleteInProgress());
    try {
      await _printerRepository.deletePrinter(event.guid);

      emit(PrinterDeleteSuccess());
    } catch (e) {
      // emit(PrinterDeleteFailure(message: e.toString()));
    }
  }

  void printerDeleteMany(
      PrinterDeleteMany event, Emitter<PrinterState> emit) async {
    emit(PrinterDeleteManyInProgress());
    try {
      await _printerRepository.deletePrinterMany(event.guid);

      emit(PrinterDeleteManySuccess());
    } catch (e) {
      // emit(PrinterDeleteFailure(message: e.toString()));
    }
  }

  void onPrinterSave(PrinterSave event, Emitter<PrinterState> emit) async {
    emit(PrinterSaveInProgress());
    try {
      await _printerRepository.savePrinter(event.printer);
      emit(PrinterSaveSuccess());
    } catch (e) {
      emit(PrinterSaveFailed(message: e.toString()));
    }
  }

  void onPrinterUpdate(PrinterUpdate event, Emitter<PrinterState> emit) async {
    emit(PrinterUpdateInProgress());
    try {
      await _printerRepository.updatePrinter(event.guid, event.printer);
      emit(PrinterUpdateSuccess());
    } catch (e) {
      emit(PrinterUpdateFailed(message: e.toString()));
    }
  }

  void onPrinterGet(PrinterGet event, Emitter<PrinterState> emit) async {
    emit(PrinterGetInProgress());
    try {
      final result = await _printerRepository.getPrinter(event.guid);
      if (result.success) {
        PrinterModel printer = PrinterModel.fromJson(result.data);
        emit(PrinterGetSuccess(printer: printer));
      } else {
        emit(const PrinterGetFailed(message: 'Printer Not Found'));
      }
    } catch (e) {
      // emit(PrinterDeleteFailure(message: e.toString()));
    }
  }
}
