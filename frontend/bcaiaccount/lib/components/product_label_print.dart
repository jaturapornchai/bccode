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
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductLabelPrint {
  // Constants
  static const int _itemsPerPage = 8;
  static const Duration _imageTimeout = Duration(seconds: 10);
  static const Duration _retryDelay = Duration(seconds: 1);
  static const int _maxRetries = 3;
  static const int _maxPdfSize = 100 * 1024 * 1024; // 100MB

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
        final response = await _client.get(
          Uri.parse(imageUrl.replaceAll('format=webp', 'format=jpg')),
          headers: {
            'User-Agent': 'Mozilla/5.0',
            'Accept': 'image/jpeg,image/png,*/*',
          },
        ).timeout(_imageTimeout);

        if (response.statusCode == 200) {
          final bytes = response.bodyBytes;
          _imageCache[imageUrl] = bytes;
          await DefaultCacheManager().putFile(
            imageUrl,
            bytes,
            maxAge: const Duration(days: 7),
          );
          return bytes;
        }
        break; // Break if response is not 200 but no exception
      } catch (e) {
        AppLogger.error('Error fetching image (attempt ${retryCount + 1}): $e');
        retryCount++;
        if (retryCount < _maxRetries) {
          await Future.delayed(_retryDelay);
        } else {
          AppLogger.debug('Max retries reached for image: $imageUrl');
        }
      }
    }
    return null;
  }

  static Future<void> showPdfPreview(
      BuildContext context, List<ProductBarcodeModel> products) async {
    // สร้าง overlay loading เพื่อป้องกันการ interact กับ UI ขณะโหลด
    OverlayEntry? loadingOverlay;
    Uint8List? pdfBytes;

    // ค่าเริ่มต้นการแสดงรูปภาพและลำดับรายการ
    bool showImages = true;
    bool showNumbering = true;

    // สร้าง loading overlay
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
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const SizedBox(
                  width: 50,
                  height: 50,
                  child: CircularProgressIndicator(
                    strokeWidth: 4,
                    backgroundColor: Colors.grey,
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
      // แสดง dialog เพื่อเลือกการตั้งค่า
      final printSettings = await _showPrintSettingsDialog(context);

      // ถ้า user กดยกเลิก
      if (printSettings == null) {
        return;
      }

      showImages = printSettings['showImages']!;
      showNumbering = printSettings['showNumbering']!;

      // แสดง loading overlay
      Overlay.of(context).insert(loadingOverlay);

      // อัพเดทข้อความ loading
      _loadingText.value = global.language('loading_image');

      // โหลดรูปภาพเฉพาะเมื่อต้องการแสดงรูปภาพเท่านั้น
      if (showImages) {
        await _prefetchImages(products);
      }

      _loadingText.value = global.language('generating_pdf');
      pdfBytes =
          await generatePdf(context, products, showImages, showNumbering);

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
                title: Text(global.language('print_product_list')),
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

  // แสดง dialog ตั้งค่าการพิมพ์
  static Future<Map<String, bool>?> _showPrintSettingsDialog(
      BuildContext context) async {
    bool showImages = true;
    bool showNumbering = true;

    return showDialog<Map<String, bool>>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(global.language('print_settings')),
          content: StatefulBuilder(
            builder: (context, setState) {
              return Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CheckboxListTile(
                    title: Text(global.language('show_image')),
                    value: showImages,
                    onChanged: (value) {
                      setState(() {
                        showImages = value!;
                      });
                    },
                  ),
                  CheckboxListTile(
                    title: Text(global.language('show_item_sequence')),
                    value: showNumbering,
                    onChanged: (value) {
                      setState(() {
                        showNumbering = value!;
                      });
                    },
                  ),
                ],
              );
            },
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('cancel')),
            ),
            ElevatedButton(
              onPressed: () => Navigator.pop(
                context,
                {
                  'showImages': showImages,
                  'showNumbering': showNumbering,
                },
              ),
              child: Text(global.language('print')),
            ),
          ],
        );
      },
    );
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
      // แสดง loading overlay
      Overlay.of(context).insert(printingOverlay);

      final result = await Printing.layoutPdf(
        onLayout: (_) => Future.value(pdfBytes),
      );

      // ปิด loading overlay
      printingOverlay.remove();

      if (result && context.mounted) {
        _clearResources();
        Navigator.pop(context);
        global.showSuccessSnackBar(context, global.language('print_document_complete'));
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
      BuildContext context,
      List<ProductBarcodeModel> products,
      bool showImages,
      bool showNumbering) async {
    final pdf = pw.Document();

    try {
      final font = await PdfGoogleFonts.iBMPlexSansThaiMedium();
      final fontBold = await PdfGoogleFonts.iBMPlexSansThaiBold();
      final totalPages = (products.length / _itemsPerPage).ceil();

      await _generatePdfPages(
        pdf: pdf,
        products: products,
        totalPages: totalPages,
        font: font,
        fontBold: fontBold,
        showImages: showImages,
        showNumbering: showNumbering,
      );

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

  static Future<void> _generatePdfPages({
    required pw.Document pdf,
    required List<ProductBarcodeModel> products,
    required int totalPages,
    required pw.Font font,
    required pw.Font fontBold,
    required bool showImages,
    required bool showNumbering,
  }) async {
    for (int pageNum = 0; pageNum < totalPages; pageNum++) {
      final startIndex = pageNum * _itemsPerPage;
      final endIndex = (startIndex + _itemsPerPage).clamp(0, products.length);

      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          theme: pw.ThemeData.withFont(base: font, bold: fontBold),
          build: (context) => _buildPage(
            products: products,
            startIndex: startIndex,
            endIndex: endIndex,
            showImages: showImages,
            showNumbering: showNumbering,
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
    required bool showNumbering,
  }) {
    return pw.Padding(
      padding: const pw.EdgeInsets.all(10),
      child: pw.GridView(
        crossAxisCount: 2,
        childAspectRatio: 1.2,
        mainAxisSpacing: 10,
        crossAxisSpacing: 10,
        children: List.generate(
          _itemsPerPage,
          (index) => _buildProductCell(
            products,
            startIndex + index,
            showImages,
            showNumbering,
          ),
        ),
      ),
    );
  }

  static pw.Widget _buildProductCell(List<ProductBarcodeModel> products,
      int index, bool showImages, bool showNumbering) {
    if (index >= products.length) return pw.Container();

    final product = products[index];
    final hasImage = showImages &&
        product.imageuri != null &&
        _imageCache.containsKey(product.imageuri);

    return pw.Container(
      decoration: pw.BoxDecoration(
        border: pw.Border.all(color: PdfColors.grey),
        borderRadius: const pw.BorderRadius.all(pw.Radius.circular(5)),
      ),
      padding: const pw.EdgeInsets.all(8),
      child: pw.Stack(
        children: [
          // Number in top-left corner (always reserve space but only show if requested)
          pw.Positioned(
            left: 0,
            top: 0,
            child: pw.Container(
              padding: const pw.EdgeInsets.all(4),
              decoration: showNumbering
                  ? pw.BoxDecoration(
                      color: PdfColors.grey300,
                      borderRadius: const pw.BorderRadius.only(
                        topLeft: pw.Radius.circular(4),
                        bottomRight: pw.Radius.circular(4),
                      ),
                    )
                  : null,
              child: showNumbering
                  ? pw.Text(
                      '${index + 1}',
                      style: pw.TextStyle(
                        fontSize: 10,
                        color: PdfColors.black,
                        fontWeight: pw.FontWeight.bold,
                      ),
                    )
                  : pw.SizedBox(
                      width: 14, height: 14), // Placeholder with same size
            ),
          ),
          // Main content
          pw.Column(
            children: [
              // Always include image section with same flex to maintain layout
              _buildProductImageSection(product, hasImage, showImages),
              pw.SizedBox(height: 8),
              _buildProductInfo(product),
            ],
          ),
        ],
      ),
    );
  }

  static pw.Widget _buildProductImageSection(
      ProductBarcodeModel product, bool hasImage, bool showImages) {
    // Always use the same flex for consistent layout
    return pw.Expanded(
      flex: 6,
      child: pw.Container(
        padding: const pw.EdgeInsets.all(5),
        // Always add border for consistent appearance regardless of content
        decoration: pw.BoxDecoration(
          border: pw.Border.all(color: PdfColors.grey300),
          borderRadius: const pw.BorderRadius.all(pw.Radius.circular(5)),
        ),
        child: !showImages
            ?
            // If images are disabled, use an empty container but keep the border
            pw.Container()
            : hasImage
                ?
                // Has image and showing images
                pw.Image(
                    pw.MemoryImage(_imageCache[product.imageuri]!),
                    fit: pw.BoxFit.contain,
                  )
                :
                // No image but showing image placeholders
                pw.Center(
                    child: pw.Text(
                      'No Image',
                      style:
                          pw.TextStyle(fontSize: 12, color: PdfColors.grey500),
                    ),
                  ),
      ),
    );
  }

  static pw.Widget _buildProductInfo(ProductBarcodeModel product) {
    return pw.Expanded(
      flex: 4,
      child: pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.center,
        children: [
          pw.Container(
            width: 60,
            height: 60,
            child: pw.BarcodeWidget(
              data: product.barcode!,
              barcode: pw.Barcode.qrCode(),
              drawText: false,
            ),
          ),
          pw.SizedBox(width: 8),
          pw.Expanded(
            child: pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              mainAxisAlignment: pw.MainAxisAlignment.center,
              children: [
                pw.Text(
                  global.activeLangName(product.names!),
                  style: pw.TextStyle(
                      fontSize: 10, fontWeight: pw.FontWeight.bold),
                  maxLines: 2,
                  overflow: pw.TextOverflow.clip,
                ),
                pw.SizedBox(height: 2),
                pw.Text(product.barcode!, style: pw.TextStyle(fontSize: 8)),
                pw.Text(
                  global.activeLangName(product.itemunitnames!),
                  style: pw.TextStyle(fontSize: 8),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
