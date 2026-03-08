part of 'shelf_excel_import_bloc.dart';

abstract class ShelfExcelImportState extends Equatable {
  const ShelfExcelImportState();

  @override
  List<Object> get props => [];
}

/// Initial state
class ExcelImportInitial extends ShelfExcelImportState {}

/// กำลังประมวลผล
class ExcelImportInProgress extends ShelfExcelImportState {
  final String message;

  const ExcelImportInProgress({required this.message});

  @override
  List<Object> get props => [message];
}

/// อ่านไฟล์ Excel สำเร็จ และจัดกลุ่มข้อมูลแล้ว
class ExcelImportParsed extends ShelfExcelImportState {
  final List<ExcelShelfGroupModel> shelfGroups;
  final int totalRows;
  final int processedRows;
  final List<String> errorMessages;

  const ExcelImportParsed({
    required this.shelfGroups,
    required this.totalRows,
    required this.processedRows,
    required this.errorMessages,
  });

  @override
  List<Object> get props =>
      [shelfGroups, totalRows, processedRows, errorMessages];
}

/// ประมวลผลข้อมูลเสร็จสิ้น
class ExcelImportSuccess extends ShelfExcelImportState {
  final List<ShelfModel> shelves;
  final List<String> processingErrors;
  final int totalShelves;
  final int processedShelves;

  const ExcelImportSuccess({
    required this.shelves,
    required this.processingErrors,
    required this.totalShelves,
    required this.processedShelves,
  });

  @override
  List<Object> get props =>
      [shelves, processingErrors, totalShelves, processedShelves];
}

/// เกิดข้อผิดพลาด
class ExcelImportFailed extends ShelfExcelImportState {
  final String message;
  final List<String> errorMessages;

  const ExcelImportFailed({
    required this.message,
    this.errorMessages = const [],
  });

  @override
  List<Object> get props => [message, errorMessages];
}
