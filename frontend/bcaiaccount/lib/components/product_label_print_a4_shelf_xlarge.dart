import 'dart:async';
import 'dart:typed_data';
import 'package:smlaicloud/model/product_model.dart';
import 'package:flutter/material.dart';
import 'package:printing/printing.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_cache_manager/flutter_cache_manager.dart';
import 'package:intl/intl.dart';

class ProductLabelPrintA4ShelfXLarge {
  // Constants - 1 label per page, each taking most of A4 page
  static const int _maxPdfSize = 100 * 1024 * 1024; // 100MB
  // Label dimensions in points - use most of A4 page in landscape

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
    PdfColor priceTextColor = PdfColors.black,
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
                color: global.theme.cardColor,
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
        priceTextColor,
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
                title: const Text('ป้ายราคาติดชั้น A4 (21x30cm)'),
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
              color: global.theme.cardColor,
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
    PdfColor priceTextColor,
    bool showBorder, // เพิ่มพารามิเตอร์การแสดงกรอบ
  ) async {
    final pdf = pw.Document();

    // โหลดฟอนต์
    final font = await PdfGoogleFonts.iBMPlexSansThaiMedium();
    final fontBold = await PdfGoogleFonts.iBMPlexSansThaiBold();

    final totalPages = products.length; // 1 label per page

    await _generatePdfPages(
      pdf: pdf,
      products: products,
      totalPages: totalPages,
      font: font,
      fontBold: fontBold,
      showImages: showImages,
      priceTextColor: priceTextColor,
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
    required PdfColor priceTextColor,
    required bool showBorder, // เพิ่มพารามิเตอร์การแสดงกรอบ
  }) async {
    for (int pageNum = 0; pageNum < totalPages; pageNum++) {
      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          orientation: pw.PageOrientation.landscape,
          margin: const pw.EdgeInsets.all(0),
          theme: pw.ThemeData.withFont(base: font, bold: fontBold),
          build: (context) => _buildPage(
            products: products,
            startIndex: pageNum,
            endIndex: pageNum + 1,
            showImages: showImages,
            priceTextColor: priceTextColor,
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
    required PdfColor priceTextColor,
    required bool showBorder, // เพิ่มพารามิเตอร์การแสดงกรอบ
  }) {
    return pw.Padding(
      padding: const pw.EdgeInsets.all(
        20,
      ), // ลด padding เพื่อใช้พื้นที่เต็มกระดาษ
      child: _buildLabel(
        products[startIndex],
        showImages,
        priceTextColor,
        showBorder,
      ), // ส่งค่ากรอบ
    );
  }

  // Helper method to determine font size based on price text length
  static double _getPriceFontSize(String priceText) {
    // Remove commas and decimal points to count actual digits
    final digitsOnly = priceText.replaceAll(RegExp(r'[,.]'), '');
    final length = digitsOnly.length;

    // Dynamic font size based on number of digits - optimized for full A4 landscape
    if (length <= 4) {
      return 72.0; // ใหญ่มากสำหรับตัวเลขน้อย
    } else if (length <= 5) {
      return 64.0; // พัน
    } else if (length <= 6) {
      return 56.0; // หมื่น
    } else if (length <= 7) {
      return 48.0; // แสน
    } else if (length <= 8) {
      return 42.0; // ล้าน
    } else if (length <= 9) {
      return 38.0; // สิบล้าน
    } else {
      return 34.0; // ร้อยล้านขึ้นไป
    }
  }

  static pw.Widget _buildLabel(
    ProductBarcodeModel product,
    bool showImages,
    PdfColor priceTextColor,
    bool showBorder,
  ) {
    try {
      final currentDate = DateFormat('dd/MM/yyyy').format(DateTime.now());
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

      return pw.Container(
        width: double.infinity,
        height: double.infinity,
        decoration: pw.BoxDecoration(
          border: showBorder
              ? pw.Border.all(color: PdfColors.black, width: 2)
              : null,
        ),
        padding: const pw.EdgeInsets.all(30),
        child: pw.Column(
          crossAxisAlignment: pw.CrossAxisAlignment.stretch,
          children: [
            // แถว 1: รหัสสินค้า ด้านซ้าย - วันที่พิมพ์ ด้านขวา
            pw.Row(
              mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
              children: [
                pw.Text(
                  productCode.isNotEmpty ? 'รหัสสินค้า: $productCode' : '',
                  style: const pw.TextStyle(
                    fontSize: 18,
                    color: PdfColors.grey700,
                  ),
                  maxLines: 1,
                  overflow: pw.TextOverflow.clip,
                ),
                pw.Text(
                  'วันที่พิมพ์: $currentDate',
                  style: const pw.TextStyle(
                    fontSize: 18,
                    color: PdfColors.grey700,
                  ),
                  maxLines: 1,
                  overflow: pw.TextOverflow.clip,
                ),
              ],
            ),

            pw.SizedBox(height: 20),

            // แถว 2: ชื่อสินค้า
            pw.Container(
              width: double.infinity,
              child: pw.Text(
                productName,
                style: pw.TextStyle(
                  fontSize: 36,
                  fontWeight: pw.FontWeight.bold,
                  color: PdfColors.black,
                ),
                maxLines: 4,
                overflow: pw.TextOverflow.span,
                textAlign: pw.TextAlign.center,
              ),
            ),

            pw.SizedBox(height: 30),

            // แถว 3: ราคา
            pw.Container(
              width: double.infinity,
              // decoration: pw.BoxDecoration(
              //   color: PdfColors.grey100,
              //   borderRadius: const pw.BorderRadius.all(pw.Radius.circular(15)),
              //   border: pw.Border.all(color: PdfColors.grey300, width: 2),
              // ),
              padding: const pw.EdgeInsets.symmetric(
                vertical: 25,
                horizontal: 20,
              ),
              child: pw.Text(
                priceText,
                style: pw.TextStyle(
                  fontSize: _getPriceFontSize(priceText),
                  fontWeight: pw.FontWeight.bold,
                  color: priceTextColor,
                ),
                textAlign: pw.TextAlign.center,
                maxLines: 1,
                overflow: pw.TextOverflow.clip,
              ),
            ),

            pw.SizedBox(height: 30),

            // แถว 4: บาร์โค้ด (ขยายความสูงขึ้น)
            if (product.barcode != null && product.barcode!.isNotEmpty)
              pw.Container(
                width: double.infinity,
                height: 120, // เพิ่มความสูงจาก 80 เป็น 120
                child: pw.BarcodeWidget(
                  data: product.barcode!,
                  barcode: pw.Barcode.code128(),
                  drawText: false,
                  height: 120, // เพิ่มความสูงจาก 80 เป็น 120
                ),
              ),

            // ใช้ Spacer เพื่อดันแถว 5 ลงไปชิดขอบด้านล่าง
            pw.Spacer(),

            // แถว 5: รหัสบาร์โค้ด ด้านซ้าย - บาท/หน่วยนับ ด้านขวา (ชิดขอบด้านล่าง)
            pw.Row(
              mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
              children: [
                pw.Text(
                  product.barcode != null && product.barcode!.isNotEmpty
                      ? 'รหัสบาร์โค้ด: ${product.barcode!}'
                      : '',
                  style: const pw.TextStyle(
                    fontSize: 16,
                    color: PdfColors.grey700,
                  ),
                  maxLines: 1,
                  overflow: pw.TextOverflow.clip,
                ),
                // shelf info
                pw.Text(
                  '${product.shelfCode ?? ''} ${product.shelfName ?? ''}',
                  style: const pw.TextStyle(
                    fontSize: 16,
                    color: PdfColors.grey600,
                  ),
                  textAlign: pw.TextAlign.center,
                  overflow: pw.TextOverflow.clip,
                ),

                pw.Text(
                  'บาท/$unitName',
                  style: pw.TextStyle(
                    fontSize: 20,
                    fontWeight: pw.FontWeight.bold,
                    color: PdfColors.black,
                  ),
                  maxLines: 1,
                  overflow: pw.TextOverflow.clip,
                ),
              ],
            ),

            // แถว 6: Shelf info - แสดงด้านล่างสุด กึ่งกลาง
            if (product.shelfCode != null || product.shelfName != null)
              pw.Container(
                width: double.infinity,
                padding: const pw.EdgeInsets.only(top: 10),
                child: pw.Text(
                  '${product.shelfCode ?? ''} ${product.shelfName ?? ''}',
                  style: const pw.TextStyle(
                    fontSize: 12,
                    color: PdfColors.grey600,
                  ),
                  textAlign: pw.TextAlign.center,
                  overflow: pw.TextOverflow.clip,
                ),
              ),
          ],
        ),
      );
    } catch (e) {
      // ถ้ามี error ให้แสดง label พื้นฐาน
      return pw.Container(
        width: double.infinity,
        height: double.infinity,
        decoration: pw.BoxDecoration(
          border: showBorder
              ? pw.Border.all(color: PdfColors.black, width: 2)
              : null,
        ),
        padding: const pw.EdgeInsets.all(30),
        child: pw.Center(
          child: pw.Text(
            'Error: ${product.barcode ?? 'Unknown'}',
            style: const pw.TextStyle(fontSize: 24, color: PdfColors.red),
            textAlign: pw.TextAlign.center,
          ),
        ),
      );
    }
  }
}
