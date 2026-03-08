part of 'shelf_excel_import_bloc.dart';

abstract class ShelfExcelImportEvent extends Equatable {
  const ShelfExcelImportEvent();

  @override
  List<Object> get props => [];
}

/// Event สำหรับการ import ไฟล์ Excel
class ImportExcelFile extends ShelfExcelImportEvent {
  final Uint8List fileBytes;
  final String fileName;

  const ImportExcelFile({
    required this.fileBytes,
    required this.fileName,
  });

  @override
  List<Object> get props => [fileBytes, fileName];
}

/// Event สำหรับการประมวลผลข้อมูลที่ import มาแล้ว
class ProcessImportedData extends ShelfExcelImportEvent {
  final List<ExcelShelfGroupModel> shelfGroups;

  const ProcessImportedData({
    required this.shelfGroups,
  });

  @override
  List<Object> get props => [shelfGroups];
}
