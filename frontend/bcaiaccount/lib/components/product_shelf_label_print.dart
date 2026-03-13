import 'dart:async';
import 'dart:typed_data';
import 'package:smlaicloud/model/product_model.dart';
import 'package:flutter/material.dart';
import 'package:printing/printing.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:smlaicloud/global.dart' as global;
import 'package:http/http.dart' as http;
import 'package:flutter_cache_manager/flutter_cache_manager.dart';
import 'package:barcode/barcode.dart' as barcode_gen;
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductShelfLabelPrint {
  // Constants
  static const Duration _imageTimeout = Duration(seconds: 10);
  static const Duration _retryDelay = Duration(seconds: 1);
  static const int _maxRetries = 3;
  static const int _maxPdfSize = 100 * 1024 * 1024; // 100MB

  // Label dimensions (6.5cm x 3.5cm)
  static final PdfPageFormat _labelFormat = PdfPageFormat(
      6.5 * PdfPageFormat.cm, 3.5 * PdfPageFormat.cm,
      marginAll: 0.2 * PdfPageFormat.cm);

  // Static resources
  static http.Client _client = http.Client();
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

  static void _initClient() {
    if (_isDisposed) {
      _client = http.Client();
      _isDisposed = false;
    }
  }

  static Future<Uint8List?> _fetchImage(String imageUrl) async {
    if (_imageCache.containsKey(imageUrl)) {
      return _imageCache[imageUrl];
    }

    _initClient(); // Ensure client is initialized

    int retryCount = 0;
    while (retryCount < _maxRetries) {
      try {
        // Try disk cache first
        try {
          final file = await DefaultCacheManager().getSingleFile(imageUrl);
          if (await file.exists()) {
            final bytes = await file.readAsBytes();
            _imageCache[imageUrl] = bytes;
            return bytes;
          }
        } catch (e) {
          AppLogger.error('Cache error: $e');
        }

        // Fetch from network if not in cache
        final response = await _client
            .get(
              Uri.parse(imageUrl),
            )
            .timeout(_imageTimeout);

        if (response.statusCode == 200) {
          final bytes = response.bodyBytes;
          _imageCache[imageUrl] = bytes;
          return bytes;
        } else {
          AppLogger.error('Image fetch error: ${response.statusCode}');
          retryCount++;
          await Future.delayed(_retryDelay);
        }
      } catch (e) {
        AppLogger.error('Network error: $e');
        retryCount++;
        await Future.delayed(_retryDelay);
      }
    }

    AppLogger.error('Failed to load image: $imageUrl after $_maxRetries attempts');
    return null;
  }

  static Future<void> showPdfPreview(BuildContext context,
      List<ProductBarcodeModel> products, List<int> copies) async {
    // สร้าง overlay loading เพื่อป้องกันการ interact กับ UI ขณะโหลด
    OverlayEntry? loadingOverlay;
    Uint8List? pdfBytes;

    // สร้าง loading overlay
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
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                SizedBox(
                  width: 50,
                  height: 50,
                  child: CircularProgressIndicator(
                    strokeWidth: 4,
                    backgroundColor: global.theme.iconSecondaryColor,
                  ),
                ),
                const SizedBox(height: 20),
                ValueListenableBuilder<String>(
                  valueListenable: _loadingText,
                  builder: (context, value, child) => Text(
                    value,
                    style: const TextStyle(fontSize: 16),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );

    try {
      // เตรียมรายการสินค้าที่ต้องพิมพ์ โดยทำสำเนาตามจำนวนที่ระบุจาก copies
      final List<ProductBarcodeModel> productsToPrint = [];

      for (int i = 0; i < products.length; i++) {
        // ถ้า i < copies.length จะใช้ค่าจาก copies
        // ถ้าไม่ จะใช้ค่า default เป็น 1
        int numCopies = (i < copies.length) ? copies[i] : 1;

        for (int j = 0; j < numCopies; j++) {
          productsToPrint.add(products[i]);
        }
      }

      // แสดง loading overlay
      Overlay.of(context).insert(loadingOverlay);

      // อัพเดทข้อความ loading
      _loadingText.value = global.language('loading_image');

      // โหลดรูปภาพสินค้า
      await _prefetchImages(productsToPrint);

      _loadingText.value = global.language('generating_pdf');
      pdfBytes = await generatePdf(productsToPrint);

      // ปิด loading overlay
      loadingOverlay.remove();
      loadingOverlay = null;

      if (context.mounted) {
        // แสดง preview dialog
        await showDialog(
          context: context,
          barrierDismissible: false,
          builder: (context) => WillPopScope(
            onWillPop: () async => false,
            child: Scaffold(
              appBar: AppBar(
                backgroundColor: global.theme.appBarColor,
                title: Text(global.language('print_shelf_label')),
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
                initialPageFormat: _labelFormat,
                canChangePageFormat: false,
                canChangeOrientation: false,
                allowPrinting: false,
                allowSharing: false,
                shouldRepaint: false,
                useActions: false,
                dynamicLayout: false,
                previewPageMargin: const EdgeInsets.all(8),
              ),
            ),
          ),
        );
      }
    } catch (e) {
      // ปิด loading overlay ถ้ายังเปิดอยู่
      if (loadingOverlay != null) {
        loadingOverlay.remove();
        loadingOverlay = null;
      }

      _clearResources();
      if (context.mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  // Loading text notifier
  static final ValueNotifier<String> _loadingText = ValueNotifier('');

  // แยกฟังก์ชันจัดการการพิมพ์
  static Future<void> _handlePrint(
      BuildContext context, Uint8List pdfBytes) async {
    // สร้าง loading overlay สำหรับการพิมพ์
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
      // แสดง loading overlay
      Overlay.of(context).insert(printingOverlay);

      final result = await Printing.layoutPdf(
        onLayout: (_) => Future.value(pdfBytes),
        format: _labelFormat,
      );

      // ปิด loading overlay
      printingOverlay.remove();

      if (result && context.mounted) {
        _clearResources();
        Navigator.pop(context);
        global.showSuccessSnackBar(context, global.language("print_shelf_label_complete"));
      }
    } catch (e) {
      // ปิด loading overlay
      printingOverlay.remove();

      _clearResources();
      if (context.mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  static Future<Uint8List> generatePdf(
      List<ProductBarcodeModel> products) async {
    final pdf = pw.Document();

    try {
      final font = await PdfGoogleFonts.iBMPlexSansThaiMedium();
      final fontBold = await PdfGoogleFonts.iBMPlexSansThaiBold();

      for (int i = 0; i < products.length; i++) {
        final product = products[i];

        pdf.addPage(
          pw.Page(
            pageFormat: _labelFormat,
            theme: pw.ThemeData.withFont(base: font, bold: fontBold),
            build: (context) => _buildShelfLabelContent(product),
          ),
        );
      }

      final pdfBytes = await pdf.save();
      final pdfSize = pdfBytes.length;

      if (pdfSize > _maxPdfSize) {
        throw Exception(
            'PDF size too large: ${(pdfSize / (1024 * 1024)).toStringAsFixed(2)}MB');
      }

      return pdfBytes;
    } catch (e) {
      _clearResources();
      rethrow;
    }
  }

  static Future<void> _prefetchImages(
      List<ProductBarcodeModel> products) async {
    final imageFutures = products
        .where((p) => p.imageuri != null && p.imageuri!.isNotEmpty)
        .map((p) => _fetchImage(p.imageuri!));

    await Future.wait(
      imageFutures,
      eagerError: false,
    );
  }

  // Build shelf label content
  static pw.Widget _buildShelfLabelContent(ProductBarcodeModel product) {
    final hasImage = product.imageuri != null &&
        product.imageuri!.isNotEmpty &&
        _imageCache.containsKey(product.imageuri);

    // Get price from the first price element (if available)
    final String priceText =
        (product.prices != null && product.prices!.isNotEmpty)
            ? '฿${product.prices![0].price.toStringAsFixed(2)}'
            : '';

    // Generate linear barcode SVG string
    final barcodeString = barcode_gen.Barcode.code128().toSvg(
      product.barcode ?? '',
      width: 120,
      height: 30,
      fontHeight: 0,
    );

    return pw.Padding(
      padding: const pw.EdgeInsets.all(4),
      child: pw.Column(
        crossAxisAlignment: pw.CrossAxisAlignment.center,
        children: [
          // Top row: Product name and optional image
          pw.Row(
            crossAxisAlignment: pw.CrossAxisAlignment.start,
            children: [
              // Product information
              pw.Expanded(
                flex: 3,
                child: pw.Column(
                  crossAxisAlignment: pw.CrossAxisAlignment.start,
                  children: [
                    // Product name
                    pw.Text(
                      global.activeLangName(product.names!),
                      style: pw.TextStyle(
                          fontSize: 10, fontWeight: pw.FontWeight.bold),
                      maxLines: 2,
                    ),

                    // Product code
                    pw.SizedBox(height: 2),
                    pw.Text(
                      '${global.language("code_label")} ${product.itemcode ?? ''}',
                      style: pw.TextStyle(fontSize: 8),
                    ),

                    // Price with larger font
                    if (priceText.isNotEmpty)
                      pw.Padding(
                        padding: const pw.EdgeInsets.only(top: 2),
                        child: pw.Text(
                          priceText,
                          style: pw.TextStyle(
                              fontSize: 14, fontWeight: pw.FontWeight.bold),
                        ),
                      ),
                  ],
                ),
              ),

              // Product image
              if (hasImage)
                pw.Container(
                  width: 60,
                  height: 60,
                  child: pw.ClipRRect(
                    verticalRadius: 4,
                    horizontalRadius: 4,
                    child: pw.Image(
                      pw.MemoryImage(_imageCache[product.imageuri]!),
                      fit: pw.BoxFit.cover,
                    ),
                  ),
                ),
            ],
          ),

          // Bottom row: QR code and barcode
          pw.SizedBox(height: 4),
          pw.Row(
            crossAxisAlignment: pw.CrossAxisAlignment.end,
            children: [
              // QR code
              pw.Container(
                width: 45,
                height: 45,
                child: pw.BarcodeWidget(
                  data: product.barcode!,
                  barcode: pw.Barcode.qrCode(),
                  drawText: false,
                ),
              ),

              pw.SizedBox(width: 5),

              // Linear barcode and barcode text
              pw.Expanded(
                child: pw.Column(
                  crossAxisAlignment: pw.CrossAxisAlignment.center,
                  children: [
                    // Linear barcode
                    pw.SvgImage(
                      svg: barcodeString,
                      height: 30,
                    ),

                    // Barcode text
                    pw.SizedBox(height: 2),
                    pw.Text(
                      product.barcode!,
                      style: pw.TextStyle(fontSize: 8),
                      textAlign: pw.TextAlign.center,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
