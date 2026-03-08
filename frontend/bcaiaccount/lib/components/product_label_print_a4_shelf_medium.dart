import 'package:flutter/foundation.dart';
import 'dart:async';
import 'package:smlaicloud/model/product_model.dart';
import 'package:flutter/material.dart';
import 'package:printing/printing.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:smlaicloud/global.dart' as global;
import 'package:http/http.dart' as http;
import 'package:flutter_cache_manager/flutter_cache_manager.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductLabelPrintA4ShelfMedium {
  // Constants - 2 labels per row, each 10cm x 3.5cm per label
  static const int _labelsPerRow = 2;
  static const int _labelsPerPage = 14; // 7 rows x 2 columns
  static const int _maxPdfSize = 100 * 1024 * 1024; // 100MB
  // Label dimensions in points (1 cm = 28.35 points)
  // Medium shelf label size: 10cm x 3.5cm
  static const double _labelWidth = 10.0 * 28.35; // 283.5 points (~10cm)
  static const double _labelHeight = 3.5 * 28.35; // 99.225 points (~3.5cm)

  // Static resources
  static final http.Client _client = http.Client();
  static final _imageCache = <String, Uint8List>{};
  static bool _isDisposed = false;
  static void _clearResources() {
    _imageCache.clear();
    if (!_isDisposed) {
      _client.close();
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
                title: const Text('ป้ายราคาติดชั้น A4 (10x3.5cm)'),
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
          margin: const pw.EdgeInsets.all(10),
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
          ); // ส่งสีและการแสดงกรอบไปยัง _buildLabel
        } else {
          // Empty space for alignment
          labelsInRow.add(
            pw.SizedBox(width: _labelWidth, height: _labelHeight),
          );
        }
      }
      labelRows.add(
        pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.start,
          children: labelsInRow,
        ),
      );

      // No spacing between rows to make labels stick together
    }
    return pw.Padding(
      padding: const pw.EdgeInsets.all(8),
      child: pw.Column(
        mainAxisAlignment: pw.MainAxisAlignment.start,
        children: labelRows,
      ),
    );
  } // Helper method to determine font size based on price text length

  static double _getPriceFontSize(String priceText) {
    // Remove commas and decimal points to count actual digits
    final digitsOnly = priceText.replaceAll(RegExp(r'[,.]'), '');
    final length = digitsOnly.length;

    // Dynamic font size based on number of digits - updated for medium size (10cm width)

    if (length <= 6) {
      return 20.0; // เพิ่มขึ้นเนื่องจากมีพื้นที่กว้างกว่า
    } else if (length <= 7) {
      return 18.0; // หมื่น
    } else if (length <= 8) {
      return 16.0; // แสน
    } else if (length <= 9) {
      return 14.0; // ล้าน
    } else {
      return 12.0; // ล้านขึ้นไป
    }
  }

  static pw.Widget _buildLabel(
    ProductBarcodeModel product,
    bool showImages,
    PdfColor priceTextColor,
    bool showBorder,
  ) {
    // เพิ่มพารามิเตอร์สีข้อความราคาและการแสดงกรอบ
    try {
      final currentDate = DateFormat('dd/MM/yyyy').format(DateTime.now());
      final price = product.prices?.isNotEmpty == true
          ? product.prices!.first.price
          : 0.0;
      // Updated formatter to support large numbers with decimal places
      final formatter = NumberFormat('#,###,##0.00', 'th_TH');
      final priceText = formatter.format(price).replaceAll('.00', '.-');
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
        padding: const pw.EdgeInsets.all(6),
        child: pw.Column(
          children: [
            // Top Row: Product Name - เต็มแนว (กำหนดความสูงคงที่ 2 แถว)
            pw.Container(
              width: double.infinity,
              height: 20, // กำหนดความสูงคงที่สำหรับ 2 แถว
              child: pw.Text(
                productName,
                style: pw.TextStyle(
                  fontSize: 10, // เพิ่มขึ้นเล็กน้อยเพื่อใช้พื้นที่เพิ่มเติม
                  fontWeight: pw.FontWeight.bold,
                  color: PdfColors.black,
                ),
                maxLines: 1,
                overflow: pw.TextOverflow.span,
              ),
            ),

            // Bottom Row: Product Info & Price
            pw.Expanded(
              child: pw.Row(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  // Left Section: Product Info & Barcode (55% width - ลดลงเล็กน้อย)
                  pw.Expanded(
                    flex: 5,
                    child: pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                      children: [
                        // Top section with product info
                        pw.Column(
                          crossAxisAlignment: pw.CrossAxisAlignment.start,
                          children: [
                            // Product Code
                            if (productCode.isNotEmpty)
                              pw.Text(
                                productCode,
                                style: const pw.TextStyle(
                                  fontSize: 8,
                                  color: PdfColors.grey700,
                                ),
                                maxLines: 1,
                                overflow: pw.TextOverflow.clip,
                              )
                            else
                              pw.Text(
                                global.language('product_code_not_specified'),
                                style: const pw.TextStyle(
                                  fontSize: 8,
                                  color: PdfColors.white,
                                ),
                                maxLines: 1,
                                overflow: pw.TextOverflow.clip,
                              ),

                            // Barcode Number
                            if (product.barcode != null &&
                                product.barcode!.isNotEmpty)
                              pw.Text(
                                product.barcode!,
                                style: const pw.TextStyle(
                                  fontSize: 8,
                                  color: PdfColors.grey700,
                                ),
                                maxLines: 1,
                                overflow: pw.TextOverflow.clip,
                              ),
                          ],
                        ), // Middle section with barcode
                        if (product.barcode != null &&
                            product.barcode!.isNotEmpty)
                          pw.Container(
                            width: double.infinity,
                            height: 25,
                            margin: const pw.EdgeInsets.symmetric(vertical: 1),
                            child: pw.BarcodeWidget(
                              data: product.barcode!,
                              barcode: pw.Barcode.code128(),
                              drawText: false,
                              height: 20,
                            ),
                          ),

                        // Bottom section - empty space for alignment
                        pw.Text(
                          '',
                          style: const pw.TextStyle(
                            fontSize: 7,
                            color: PdfColors.white,
                          ),
                        ),
                      ],
                    ),
                  ),

                  pw.SizedBox(width: 8),

                  // Right Section: Price (45% width - เพิ่มขึ้นเล็กน้อย)
                  pw.Expanded(
                    flex: 3,
                    child: pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end, // ชิดขวา
                      children: [
                        // Price
                        pw.Container(
                          width: double.infinity,
                          child: pw.Text(
                            priceText,
                            style: pw.TextStyle(
                              fontSize: _getPriceFontSize(priceText),
                              fontWeight: pw.FontWeight.bold,
                              color: priceTextColor, // ใช้สีที่เลือก
                            ),
                            textAlign: pw.TextAlign.right, // ชิดขวา
                            overflow: pw.TextOverflow.clip,
                          ),
                        ),

                        // Unit
                        pw.Container(
                          width: double.infinity,
                          child: pw.Text(
                            'บาท/$unitName',
                            style: pw.TextStyle(
                              fontSize: 9,
                              fontWeight: pw.FontWeight.bold,
                              color: PdfColors.black,
                            ),
                            textAlign: pw.TextAlign.right, // ชิดขวา
                            overflow: pw.TextOverflow.clip,
                          ),
                        ),

                        // Date - ชิดล่าง
                        pw.Container(
                          width: double.infinity,
                          child: pw.Text(
                            'วันที่พิมพ์: $currentDate',
                            style: const pw.TextStyle(
                              fontSize: 5,
                              color: PdfColors.grey600,
                            ),
                            textAlign: pw.TextAlign.right, // ชิดขวา
                            overflow: pw.TextOverflow.clip,
                          ),
                        ),

                        // Shelf info
                        if (product.shelfCode != null ||
                            product.shelfName != null)
                          pw.Container(
                            width: double.infinity,
                            child: pw.Text(
                              '${product.shelfCode ?? ''} ${product.shelfName ?? ''}',
                              style: const pw.TextStyle(
                                fontSize: 5,
                                color: PdfColors.grey600,
                              ),
                              textAlign: pw.TextAlign.right,
                              overflow: pw.TextOverflow.clip,
                            ),
                          )
                        else
                          pw.Container(
                            width: double.infinity,
                            child: pw.Text(
                              '',
                              style: const pw.TextStyle(fontSize: 6),
                            ),
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
