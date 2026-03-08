import 'dart:typed_data';
import 'package:excel/excel.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/excel_import_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:smlaicloud/model/shelf_product_model.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';

part 'shelf_excel_import_event.dart';
part 'shelf_excel_import_state.dart';

class ShelfExcelImportBloc
    extends Bloc<ShelfExcelImportEvent, ShelfExcelImportState> {
  final ProductBarcodeRepository _productBarcodeRepository;

  ShelfExcelImportBloc({
    required ProductBarcodeRepository productBarcodeRepository,
  })  : _productBarcodeRepository = productBarcodeRepository,
        super(ExcelImportInitial()) {
    on<ImportExcelFile>(onImportExcelFile);
    on<ProcessImportedData>(onProcessImportedData);
  }

  /// อ่านไฟล์ Excel และแปลงเป็น data structure
  void onImportExcelFile(
      ImportExcelFile event, Emitter<ShelfExcelImportState> emit) async {
    emit(ExcelImportInProgress(message: 'กำลังอ่านไฟล์ Excel...'));

    try {
      // อ่านไฟล์ Excel
      var excel = Excel.decodeBytes(event.fileBytes);

      if (excel.tables.isEmpty) {
        emit(const ExcelImportFailed(message: 'ไฟล์ Excel ไม่มีข้อมูล'));
        return;
      }

      var sheet = excel.tables[excel.tables.keys.first];
      if (sheet == null) {
        emit(const ExcelImportFailed(message: 'ไม่สามารถอ่าน sheet แรกได้'));
        return;
      }

      List<ExcelShelfRowModel> rows = [];
      List<String> errorMessages = [];
      int processedRows = 0;

      // skip header row (row 0) และเริ่มจาก row 1
      for (int rowIndex = 1; rowIndex < sheet.maxRows; rowIndex++) {
        var row = sheet.rows[rowIndex];

        // ตรวจสอบว่าแถวมีข้อมูลหรือไม่ (อย่างน้อยต้องมี 1 คอลัมน์ที่มีค่า)
        bool hasData = false;
        for (var cell in row) {
          if (cell?.value != null && cell!.value.toString().trim().isNotEmpty) {
            hasData = true;
            break;
          }
        }

        if (!hasData) {
          continue; // ข้ามแถวที่ไม่มีข้อมูล
        }

        // อ่านข้อมูลจากแต่ละคอลัมน์ (รองรับกรณีที่มีคอลัมน์น้อยกว่า 3)
        var shelfCodeCell = row.isNotEmpty ? row[0] : null;
        var shelfNameCell = row.length > 1 ? row[1] : null;
        var barcodeCell = row.length > 2 ? row[2] : null;

        // ตรวจสอบว่าข้อมูลจำเป็นครบหรือไม่
        if (shelfCodeCell?.value == null ||
            shelfNameCell?.value == null ||
            barcodeCell?.value == null) {
          errorMessages.add(
              'แถวที่ ${rowIndex + 1}: ข้อมูลไม่ครบถ้วน (ต้องมี shelf_code, shelf_name, barcode)');
          continue;
        }

        String shelfCode = shelfCodeCell!.value.toString().trim();
        String shelfName = shelfNameCell!.value.toString().trim();
        String barcode = barcodeCell!.value.toString().trim();

        if (shelfCode.isEmpty || shelfName.isEmpty || barcode.isEmpty) {
          errorMessages.add('แถวที่ ${rowIndex + 1}: มีข้อมูลว่าง');
          continue;
        }

        rows.add(ExcelShelfRowModel(
          shelfCode: shelfCode,
          shelfName: shelfName,
          barcode: barcode,
        ));

        processedRows++;
      }

      if (rows.isEmpty) {
        emit(ExcelImportFailed(
          message: 'ไม่มีข้อมูลที่สามารถประมวลผลได้',
          errorMessages: errorMessages,
        ));
        return;
      }

      // จัดกลุ่มตาม shelf_code
      Map<String, ExcelShelfGroupModel> shelfGroups = {};

      for (var row in rows) {
        if (shelfGroups.containsKey(row.shelfCode)) {
          shelfGroups[row.shelfCode]!.barcodes.add(row.barcode);
        } else {
          shelfGroups[row.shelfCode] = ExcelShelfGroupModel(
            shelfCode: row.shelfCode,
            shelfName: row.shelfName,
            barcodes: [row.barcode],
          );
        }
      }

      emit(ExcelImportParsed(
        shelfGroups: shelfGroups.values.toList(),
        totalRows: sheet.maxRows - 1, // ไม่นับ header
        processedRows: processedRows,
        errorMessages: errorMessages,
      ));
    } catch (e) {
      emit(ExcelImportFailed(
          message: 'เกิดข้อผิดพลาดในการอ่านไฟล์: ${e.toString()}'));
    }
  }

  /// ประมวลผลข้อมูลที่ import มาแล้ว โดยดึงข้อมูลสินค้าจาก barcode
  void onProcessImportedData(
      ProcessImportedData event, Emitter<ShelfExcelImportState> emit) async {
    emit(ExcelImportInProgress(message: 'กำลังประมวลผลข้อมูลสินค้า...'));

    try {
      List<ShelfModel> shelves = [];
      List<String> processingErrors = [];

      int totalShelves = event.shelfGroups.length;
      int processedShelves = 0;

      for (var shelfGroup in event.shelfGroups) {
        processedShelves++;
        emit(ExcelImportInProgress(
          message:
              'กำลังประมวลผลชั้นวาง ${shelfGroup.shelfCode} ($processedShelves/$totalShelves)...',
        ));

        try {
          // ดึงข้อมูลสินค้าจาก barcode list
          final result = await _productBarcodeRepository
              .getProductBarcodeByBarcode(shelfGroup.barcodes);

          List<ShelfProductItemModel> productItems = [];

          if (result.success) {
            List<dynamic> rawData = result.data as List;
            List<String> notFoundBarcodes = [];

            // ประมวลผลข้อมูลที่ได้จาก API โดยเช็ค null values
            for (int i = 0;
                i < rawData.length && i < shelfGroup.barcodes.length;
                i++) {
              if (rawData[i] != null) {
                // สินค้าที่พบในระบบ
                ProductBarcodeModel product =
                    ProductBarcodeModel.fromJson(rawData[i]);
                productItems.add(ShelfProductItemModel(
                  barcode: product.barcode!,
                  guidfixed: product.guidfixed,
                  names: product.names!,
                  unitcode: product.itemunitcode,
                  unitnames: product.itemunitnames,
                ));
              } else {
                // สินค้าที่ไม่พบในระบบ (null)
                notFoundBarcodes.add(shelfGroup.barcodes[i]);
              }
            }

            // รายงาน barcode ที่ไม่พบ
            if (notFoundBarcodes.isNotEmpty) {
              processingErrors.add(
                  'ชั้นวาง ${shelfGroup.shelfCode}: ไม่พบสินค้าสำหรับ barcode ${notFoundBarcodes.join(", ")}');
            }
          } else {
            processingErrors
                .add('ชั้นวาง ${shelfGroup.shelfCode}: ${result.message}');
          }

          shelves.add(ShelfModel(
            code: shelfGroup.shelfCode,
            name: shelfGroup.shelfName,
            productitems: productItems,
          ));
        } catch (e) {
          processingErrors
              .add('ชั้นวาง ${shelfGroup.shelfCode}: ${e.toString()}');

          // ยังคงสร้าง shelf แต่ไม่มีสินค้า
          shelves.add(ShelfModel(
            code: shelfGroup.shelfCode,
            name: shelfGroup.shelfName,
            productitems: [],
          ));
        }
      }

      emit(ExcelImportSuccess(
        shelves: shelves,
        processingErrors: processingErrors,
        totalShelves: totalShelves,
        processedShelves: processedShelves,
      ));
    } catch (e) {
      emit(ExcelImportFailed(
          message: 'เกิดข้อผิดพลาดในการประมวลผล: ${e.toString()}'));
    }
  }
}
