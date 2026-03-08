import 'package:flutter/foundation.dart';
import 'dart:async';
import 'package:smlaicloud/model/product_model.dart';
import 'package:flutter/material.dart';
import 'package:printing/printing.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_cache_manager/flutter_cache_manager.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductLabelPrintA4ShelfLarge {
  static const int _labelsPerRow = 1;
  static const int _labelsPerPage = 2;
  static const int _maxPdfSize = 100 * 1024 * 1024; // 100MB
  // Label dimensions in points (1 cm = 28.35 points)
  // Custom large label size: 15cm x 10cm
  static const double _labelWidth = 15.0 * 28.35; // 425.25 points (~15cm)
  static const double _labelHeight = 10.0 * 28.35; // 283.5 points (~10cm)

  // Static resources
  static final _imageCache = <String, Uint8List>{};
  static bool _isDisposed = false;
  static void _clearResources() {
    _imageCache.clear();
    if (!_isDisposed) {
      _isDisposed = true;
    }
    DefaultCacheManager().emptyCache();
  }

  static Future<void> showPdfPreview(
    BuildContext context,
    List<ProductBarcodeModel> products,
    List<int> copies, {
    bool showImages = false,
    PdfColor priceTextColor = PdfColors.black, // เพิ่มพารามิเตอร์สีข้อความราคา
    bool showBorder = true, // เพิ่มพารามิเตอร์การแสดงกรอบ
  }) async {
    if (products.isEmpty) return;

    late OverlayEntry loadingOverlay;
    final ValueNotifier<String> loadingText = ValueNotifier<String>(
      global.language('preparing_data'),
    );
    Uint8List? pdfBytes;

    try {
      loadingOverlay = OverlayEntry(
        builder: (context) => Material(
          color: Colors.black26,
          child: Center(
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(10),
              ),
              child: ValueListenableBuilder<String>(
                valueListenable: loadingText,
                builder: (context, loadingMessage, child) => Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const SizedBox(
                      width: 50,
                      height: 50,
                      child: CircularProgressIndicator(),
                    ),
                    const SizedBox(height: 20),
                    Text(loadingMessage, style: const TextStyle(fontSize: 16)),
                  ],
                ),
              ),
            ),
          ),
        ),
      );

      // สร้างรายการสินค้าตามจำนวนที่ต้องการพิมพ์
      List<ProductBarcodeModel> productsToPrint = [];
      for (int i = 0; i < products.length; i++) {
        int numCopies = (i < copies.length) ? copies[i] : 1;
        for (int j = 0; j < numCopies; j++) {
          productsToPrint.add(products[i]);
        }
      }

      Overlay.of(context).insert(loadingOverlay);

      loadingText.value = global.language('loading_image');

      loadingText.value = global.language('generating_pdf');
      pdfBytes = await generatePdf(
        context,
        productsToPrint,
        showImages,
        priceTextColor, // ส่งสีไปยัง generatePdf
        showBorder, // ส่งค่าการแสดงกรอบ
      );

      loadingOverlay.remove();
      loadingOverlay.dispose();

      if (context.mounted) {
        await showDialog(
          context: context,
          barrierDismissible: false,
          builder: (context) => WillPopScope(
            onWillPop: () async => false,
            child: Scaffold(
              appBar: AppBar(
                backgroundColor: global.theme.appBarColor,
                title: const Text('ป้ายราคาติดชั้น A4 (15x10cm)'),
                leading: IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () {
                    _clearResources();
                    Navigator.pop(context);
                  },
                ),
                actions: [
                  IconButton(
                    icon: const Icon(Icons.print),
                    onPressed: () => _handlePrint(context, pdfBytes!),
                  ),
                ],
              ),
              body: PdfPreview(
                build: (_) => Future.value(pdfBytes),
                initialPageFormat: PdfPageFormat.a4,
                canChangePageFormat: false,
                canChangeOrientation: false,
                allowPrinting: false,
                allowSharing: false,
                shouldRepaint: false,
                useActions: false,
                dynamicLayout: false,
                maxPageWidth: 800,
                previewPageMargin: const EdgeInsets.all(8),
              ),
            ),
          ),
        );
      }
    } catch (e) {
      if (loadingOverlay.mounted) {
        loadingOverlay.remove();
      }
      if (context.mounted) {
        global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: ${e.toString()}');
      }
    }
  }

  static Future<void> _handlePrint(
    BuildContext context,
    Uint8List pdfBytes,
  ) async {
    final printingOverlay = OverlayEntry(
      builder: (context) => Material(
        color: Colors.black26,
        child: Center(
          child: Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                SizedBox(
                  width: 50,
                  height: 50,
                  child: CircularProgressIndicator(),
                ),
                SizedBox(height: 20),
                Text(global.language('printing'), style: TextStyle(fontSize: 16)),
              ],
            ),
          ),
        ),
      ),
    );

    try {
      Overlay.of(context).insert(printingOverlay);

      final result = await Printing.layoutPdf(
        onLayout: (_) => Future.value(pdfBytes),
      );

      printingOverlay.remove();

      if (result && context.mounted) {
        _clearResources();
        Navigator.pop(context);
        global.showSuccessSnackBar(context, global.language('print_document_complete'));
      }
    } catch (e) {
      printingOverlay.remove();
      if (context.mounted) {
        global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการพิมพ์: ${e.toString()}');
      }
    }
  }

  static Future<Uint8List> generatePdf(
    BuildContext context,
    List<ProductBarcodeModel> products,
    bool showImages,
    PdfColor priceTextColor, // เพิ่มพารามิเตอร์สีข้อความราคา
    bool showBorder, // เพิ่มพารามิเตอร์การแสดงกรอบ
  ) async {
    final pdf = pw.Document(); // โหลดฟอนต์
    final font = await PdfGoogleFonts.iBMPlexSansThaiMedium();
    final fontBold = await PdfGoogleFonts.iBMPlexSansThaiBold();

    final totalPages = (products.length / _labelsPerPage).ceil();

    await _generatePdfPages(
      pdf: pdf,
      products: products,
      totalPages: totalPages,
      font: font,
      fontBold: fontBold,
      showImages: showImages,
      priceTextColor: priceTextColor, // ส่งสีไปยัง _generatePdfPages
      showBorder: showBorder, // ส่งค่าการแสดงกรอบ
    );

    final pdfBytes = await pdf.save();
    if (pdfBytes.length > _maxPdfSize) {
      throw Exception('ไฟล์ PDF มีขนาดใหญ่เกินไป');
    }

    return pdfBytes;
  }

  static Future<void> _generatePdfPages({
    required pw.Document pdf,
    required List<ProductBarcodeModel> products,
    required int totalPages,
    required pw.Font font,
    required pw.Font fontBold,
    required bool showImages,
    required PdfColor priceTextColor, // เพิ่มพารามิเตอร์สีข้อความราคา
    required bool showBorder, // เพิ่มพารามิเตอร์การแสดงกรอบ
  }) async {
    for (int pageNum = 0; pageNum < totalPages; pageNum++) {
      final startIndex = pageNum * _labelsPerPage;
      final endIndex = (startIndex + _labelsPerPage).clamp(0, products.length);
      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          orientation: pw.PageOrientation.portrait,
          margin: const pw.EdgeInsets.all(0),
          theme: pw.ThemeData.withFont(base: font, bold: fontBold),
          build: (context) => _buildPage(
            products: products,
            startIndex: startIndex,
            endIndex: endIndex,
            showImages: showImages,
            priceTextColor: priceTextColor, // ส่งสีไปยัง _buildPage
            showBorder: showBorder, // ส่งค่าการแสดงกรอบ
          ),
        ),
      );
    }
  }

  static pw.Widget _buildPage({
    required List<ProductBarcodeModel> products,
    required int startIndex,
    required int endIndex,
    required bool showImages,
    required PdfColor priceTextColor, // เพิ่มพารามิเตอร์สีข้อความราคา
    required bool showBorder, // เพิ่มพารามิเตอร์การแสดงกรอบ
  }) {
    // Calculate the number of labels on this page
    final labelsOnPage = endIndex - startIndex;
    final rows = (labelsOnPage / _labelsPerRow).ceil();

    List<pw.Widget> labelRows = [];

    for (int row = 0; row < rows; row++) {
      List<pw.Widget> labelsInRow = [];

      for (int col = 0; col < _labelsPerRow; col++) {
        final labelIndex = startIndex + (row * _labelsPerRow) + col;

        if (labelIndex < endIndex) {
          labelsInRow.add(
            _buildLabel(
              products[labelIndex],
              showImages,
              priceTextColor,
              showBorder,
            ),
          ); // ส่งสีและกรอบไปยัง _buildLabel
        } else {
          // Empty space for alignment
          labelsInRow.add(
            pw.SizedBox(width: _labelWidth, height: _labelHeight),
          );
        }
      }
      labelRows.add(
        pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.spaceEvenly,
          children: labelsInRow
              .map(
                (label) => pw.Padding(
                  padding: const pw.EdgeInsets.symmetric(horizontal: 2),
                  child: label,
                ),
              )
              .toList(),
        ),
      );

      // ไม่มีระยะห่างระหว่างแถว ให้ป้ายติดกัน
    }
    return pw.Padding(
      padding: const pw.EdgeInsets.only(top: 20),
      child: pw.Center(
        child: pw.Column(
          mainAxisAlignment: pw.MainAxisAlignment.start,
          children: labelRows,
        ),
      ),
    );
  } // Helper method to determine font size based on price text length

  static double _getPriceFontSize(String priceText) {
    // Remove commas and decimal points to count actual digits
    final digitsOnly = priceText.replaceAll(RegExp(r'[,.]'), '');
    final length = digitsOnly.length;

    // Dynamic font size based on number of digits - optimized for 15x10cm label

    if (length <= 4) {
      return 48.0; // ราคาสั้น เช่น 99.00
    } else if (length <= 5) {
      return 42.0; // พัน เช่น 1,999
    } else if (length <= 6) {
      return 36.0; // หมื่น เช่น 19,999
    } else if (length <= 7) {
      return 30.0; // แสน เช่น 199,999
    } else if (length <= 8) {
      return 26.0; // ล้าน เช่น 1,999,999
    } else if (length <= 9) {
      return 22.0; // สิบล้าน
    } else {
      return 18.0; // ร้อยล้านขึ้นไป
    }
  }

  static pw.Widget _buildLabel(
    ProductBarcodeModel product,
    bool showImages,
    PdfColor priceTextColor,
    bool showBorder,
  ) {
    // เพิ่มพารามิเตอร์สีข้อความราคา
    try {
      final price = product.prices?.isNotEmpty == true
          ? product.prices!.first.price
          : 0.0;
      // Updated formatter to support large numbers with decimal places
      final formatter = NumberFormat('#,###,##0.00', 'th_TH');
      final priceText = formatter.format(price);
      final productCode = product.itemcode ?? product.barcode ?? '';

      // Safe access to product names
      String productName = '';
      if (product.names != null && product.names!.isNotEmpty) {
        productName = global.activeLangName(product.names!);
      }
      if (productName.isEmpty) {
        productName = global.language('product_name_not_specified');
      }

      // Safe access to unit names
      String unitName = '';
      if (product.itemunitnames != null && product.itemunitnames!.isNotEmpty) {
        unitName = global.activeLangName(product.itemunitnames!);
      }
      if (unitName.isEmpty) {
        unitName = global.language('chatbot_product_unit');
      }

      // Debug: Print price info (can be removed later)
      if (kDebugMode) {
        AppLogger.debug(
          'Product: $productName, Price: $price, Formatted: $priceText, Font Size: ${_getPriceFontSize(priceText)}',
        );
      }
      return pw.Container(
        width: _labelWidth,
        height: _labelHeight,
        decoration: pw.BoxDecoration(
          border: showBorder
              ? pw.Border.all(color: PdfColors.black, width: 1)
              : null,
        ),
        padding: const pw.EdgeInsets.all(12),
        child: pw.Column(
          children: [
            // 1. ชื่อสินค้า และ วันที่พิมพ์ (บนสุด)
            pw.Container(
              width: double.infinity,
              height: 60,
              child: pw.Row(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  // ชื่อสินค้า (ด้านซ้าย)
                  pw.Expanded(
                    child: pw.Text(
                      productName,
                      style: pw.TextStyle(
                        fontSize: 18,
                        fontWeight: pw.FontWeight.bold,
                        color: PdfColors.black,
                      ),
                      maxLines: 2,
                      overflow: pw.TextOverflow.span,
                    ),
                  ),
                  pw.SizedBox(width: 8),
                  // วันที่พิมพ์ (ด้านขวา)
                  pw.Text(
                    'วันที่พิมพ์: ${DateFormat('dd/MM/yyyy').format(DateTime.now())}',
                    style: pw.TextStyle(fontSize: 12, color: PdfColors.grey600),
                  ),
                ],
              ),
            ),

            // 2. รหัสสินค้า
            pw.Container(
              width: double.infinity,
              child: productCode.isNotEmpty
                  ? pw.Text(
                      productCode,
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.grey700,
                      ),
                      maxLines: 1,
                      overflow: pw.TextOverflow.clip,
                    )
                  : pw.Text(
                      global.language('product_code_not_specified'),
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.white,
                      ),
                      maxLines: 1,
                      overflow: pw.TextOverflow.clip,
                    ),
            ),

            pw.SizedBox(height: 4),

            // 3. ราคา (ชิดขวา - ลดขนาดเพื่อให้ barcode มีพื้นที่มากขึ้น)
            pw.Container(
              width: double.infinity,
              height: 60, // ลดจาก 80 เป็น 60
              child: pw.Align(
                alignment: pw.Alignment.centerRight,
                child: pw.Text(
                  priceText.replaceAll('.00', '.-'),
                  style: pw.TextStyle(
                    fontSize: _getPriceFontSize(priceText),
                    fontWeight: pw.FontWeight.bold,
                    color: priceTextColor,
                  ),
                  textAlign: pw.TextAlign.right,
                  maxLines: 2,
                  overflow: pw.TextOverflow.clip,
                ),
              ),
            ),

            pw.SizedBox(height: 4),
            // พื้นที่ที่เหลือสำหรับ barcode และข้อมูลด้านล่าง
            pw.Expanded(
              child: pw.Column(
                children: [
                  // 4. Barcode
                  if (product.barcode != null && product.barcode!.isNotEmpty)
                    pw.Container(
                      width: double.infinity,
                      height: 60, // กำหนดความสูงที่นี่
                      margin: const pw.EdgeInsets.symmetric(vertical: 4),
                      child: pw.BarcodeWidget(
                        data: product.barcode!,
                        barcode: pw.Barcode.code128(),
                        drawText: false,
                        width: double.infinity,
                        height: double.infinity,
                      ),
                    ),
                  // 5. รหัสบาร์โค้ด (ชิดซ้าย) | บาท/หน่วยนับ (ชิดขวา) - ชิดล่าง
                  pw.Container(
                    width: double.infinity,
                    padding: const pw.EdgeInsets.only(top: 8),
                    child: pw.Row(
                      mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                      children: [
                        // รหัสบาร์โค้ด (ชิดซ้าย)
                        pw.Text(
                          product.barcode != null && product.barcode!.isNotEmpty
                              ? product.barcode!
                              : '',
                          style: const pw.TextStyle(
                            fontSize: 12,
                            color: PdfColors.grey700,
                          ),
                          textAlign: pw.TextAlign.left,
                        ),

                        // Shelf info (กลาง)
                        pw.Text(
                          '${product.shelfCode ?? ''} ${product.shelfName ?? ''}',
                          style: const pw.TextStyle(
                            fontSize: 12,
                            color: PdfColors.grey600,
                          ),
                          textAlign: pw.TextAlign.center,
                          overflow: pw.TextOverflow.clip,
                        ),

                        // บาท/หน่วยนับ (ชิดขวา)
                        pw.Text(
                          'บาท/$unitName',
                          style: pw.TextStyle(
                            fontSize: 12,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.black,
                          ),
                          textAlign: pw.TextAlign.right,
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    } catch (e) {
      // ถ้ามี error ให้แสดง label พื้นฐาน
      return pw.Container(
        width: _labelWidth,
        height: _labelHeight,
        decoration: pw.BoxDecoration(
          border: showBorder
              ? pw.Border.all(color: PdfColors.black, width: 1)
              : null,
        ),
        padding: const pw.EdgeInsets.all(6),
        child: pw.Center(
          child: pw.Text(
            'Error: ${product.barcode ?? 'Unknown'}',
            style: const pw.TextStyle(fontSize: 10, color: PdfColors.red),
            textAlign: pw.TextAlign.center,
          ),
        ),
      );
    }
  }
}
