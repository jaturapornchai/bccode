part of 'export_csv_bloc.dart';

abstract class ExportCsvState extends Equatable {
  const ExportCsvState();

  @override
  List<Object> get props => [];
}

class ExportCsvInitial extends ExportCsvState {}

class ProductBarcodeExportInitial extends ExportCsvState {}

class ProductBarcodeExportInProgress extends ExportCsvState {}

class ProductBarcodeExportSuccess extends ExportCsvState {}

class ProductBarcodeExportFailed extends ExportCsvState {
  final String message;

  const ProductBarcodeExportFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class SaleInvoiceExportInitial extends ExportCsvState {}

class SaleInvoiceExportInProgress extends ExportCsvState {}

class SaleInvoiceExportSuccess extends ExportCsvState {}

class SaleInvoiceExportFailed extends ExportCsvState {
  final String message;

  const SaleInvoiceExportFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
