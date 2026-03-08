/// Model สำหรับข้อมูลใน Excel file แต่ละแถว
class ExcelShelfRowModel {
  final String shelfCode;
  final String shelfName;
  final String barcode;

  ExcelShelfRowModel({
    required this.shelfCode,
    required this.shelfName,
    required this.barcode,
  });
}

/// Model สำหรับ grouping ข้อมูลตาม shelf_code
class ExcelShelfGroupModel {
  final String shelfCode;
  final String shelfName;
  final List<String> barcodes;

  ExcelShelfGroupModel({
    required this.shelfCode,
    required this.shelfName,
    required this.barcodes,
  });
}

/// Model สำหรับ import result พร้อม error handling
class ExcelImportResultModel {
  final bool success;
  final String message;
  final List<ExcelShelfGroupModel> shelfGroups;
  final List<String> errorMessages;
  final int totalRows;
  final int processedRows;

  ExcelImportResultModel({
    required this.success,
    required this.message,
    required this.shelfGroups,
    required this.errorMessages,
    required this.totalRows,
    required this.processedRows,
  });
}
