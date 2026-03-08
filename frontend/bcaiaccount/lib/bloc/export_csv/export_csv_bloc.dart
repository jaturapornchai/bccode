import 'package:smlaicloud/repositories/export_csv_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'export_csv_event.dart';
part 'export_csv_state.dart';

class ExportCsvBloc extends Bloc<ExportCsvEvent, ExportCsvState> {
  final ExportCsvRepository _exportCsvRepository;

  ExportCsvBloc({required ExportCsvRepository exportCsvRepository})
      : _exportCsvRepository = exportCsvRepository,
        super(ExportCsvInitial()) {
    on<ProductBarcodeExport>(onProductBarcodeExport);
    on<SaleInvoiceExport>(onSaleInvoiceExport);
  }

  void onProductBarcodeExport(
      ProductBarcodeExport event, Emitter<ExportCsvState> emit) async {
    emit(ProductBarcodeExportInProgress());
    try {
      final success =
          await _exportCsvRepository.exportProductBarcode(event.languageCode);
      if (success) {
        emit(ProductBarcodeExportSuccess());
      } else {
        emit(const ProductBarcodeExportFailed(
            message: 'Failed to export ProductBarcode'));
      }
    } catch (e) {
      emit(ProductBarcodeExportFailed(message: e.toString()));
    }
  }

  void onSaleInvoiceExport(
      SaleInvoiceExport event, Emitter<ExportCsvState> emit) async {
    emit(SaleInvoiceExportInProgress());
    try {
      final success =
          await _exportCsvRepository.exportSaleInvoice(event.languageCode);
      if (success) {
        emit(SaleInvoiceExportSuccess());
      } else {
        emit(const SaleInvoiceExportFailed(
            message: 'Failed to export Sale-invoice'));
      }
    } catch (e) {
      emit(SaleInvoiceExportFailed(message: e.toString()));
    }
  }
}
