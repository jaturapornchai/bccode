part of 'export_csv_bloc.dart';

abstract class ExportCsvEvent extends Equatable {
  const ExportCsvEvent();

  @override
  List<Object> get props => [];
}

class ProductBarcodeExport extends ExportCsvEvent {
  final String languageCode;

  const ProductBarcodeExport({required this.languageCode});

  @override
  List<Object> get props => [languageCode];
}

class SaleInvoiceExport extends ExportCsvEvent {
  final String languageCode;

  const SaleInvoiceExport({required this.languageCode});

  @override
  List<Object> get props => [languageCode];
}
