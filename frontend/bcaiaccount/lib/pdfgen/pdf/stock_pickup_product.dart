import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:smlaicloud/global.dart' as global;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

// ฟังก์ชันหลักที่ใช้ compute ในการทำงานแยก isolate
Future<Uint8List> generateStockPickupProduct(
  PdfPageFormat pageFormat,
  TransactionModel screenData,
  CompanyBranchModel companyBranchData,
) async {
  // โหลดฟอนต์และโลโก้ล่วงหน้า
  Uint8List? logoBytes;
  if (companyBranchData.logouri != null &&
      companyBranchData.logouri!.isNotEmpty) {
    try {
      logoBytes = await global.fetchNetworkImage(companyBranchData.logouri!);
    } catch (e) {
      if (kDebugMode) {
        AppLogger.debug("ไม่สามารถโหลดโลโก้ได้: $e");
      }
    }
  }

  final fontData = await rootBundle.load("assets/fonts/THSarabunNew.ttf");
  final fontDataRegula =
      await rootBundle.load("assets/fonts/THSarabunNew-Italic.ttf");
  final fontDataBold =
      await rootBundle.load("assets/fonts/THSarabunNew-Bold.ttf");

  // ส่งข้อมูลไปทำงานใน isolate แยก
  return compute(
    _generatePdfInBackground,
    _PdfGenerationParams(
      pageFormat: pageFormat,
      screenData: screenData,
      companyBranchData: companyBranchData,
      logoBytes: logoBytes,
      fontData: fontData.buffer.asUint8List(),
      fontDataRegula: fontDataRegula.buffer.asUint8List(),
      fontDataBold: fontDataBold.buffer.asUint8List(),
    ),
  );
}

// คลาสเก็บพารามิเตอร์ที่ส่งไปยัง isolate
class _PdfGenerationParams {
  final PdfPageFormat pageFormat;
  final TransactionModel screenData;
  final CompanyBranchModel companyBranchData;
  final Uint8List? logoBytes;
  final Uint8List fontData;
  final Uint8List fontDataRegula;
  final Uint8List fontDataBold;

  _PdfGenerationParams({
    required this.pageFormat,
    required this.screenData,
    required this.companyBranchData,
    this.logoBytes,
    required this.fontData,
    required this.fontDataRegula,
    required this.fontDataBold,
  });
}

// ฟังก์ชันที่ทำงานใน isolate แยก
Future<Uint8List> _generatePdfInBackground(_PdfGenerationParams params) async {
  // สร้าง StockTransfer แบบเพิ่มประสิทธิภาพ
  final stockTransfer = OptimizedStockTransfer(
    companyBranchData: params.companyBranchData,
    screenData: params.screenData,
    logoBytes: params.logoBytes,
    fontData: params.fontData,
    fontDataRegula: params.fontDataRegula,
    fontDataBold: params.fontDataBold,
  );

  return await stockTransfer.buildPdf(params.pageFormat);
}

// คลาสจัดการการสร้าง PDF แบบเพิ่มประสิทธิภาพ
class OptimizedStockTransfer {
  OptimizedStockTransfer({
    required this.companyBranchData,
    required this.screenData,
    this.logoBytes,
    required this.fontData,
    required this.fontDataRegula,
    required this.fontDataBold,
  });

  final CompanyBranchModel companyBranchData;
  final TransactionModel screenData;
  final Uint8List? logoBytes;
  final Uint8List fontData;
  final Uint8List fontDataRegula;
  final Uint8List fontDataBold;

  // แบ่งหน้าสำหรับกรณีรายการสินค้าจำนวนมาก
  List<List<dynamic>> _paginateItems(List<dynamic> items, int itemsPerPage) {
    List<List<dynamic>> pages = [];
    for (int i = 0; i < items.length; i += itemsPerPage) {
      int end =
          (i + itemsPerPage < items.length) ? i + itemsPerPage : items.length;
      pages.add(items.sublist(i, end));
    }
    return pages;
  }

  Future<Uint8List> buildPdf(PdfPageFormat pageFormat) async {
    // Create a PDF document.
    final doc = pw.Document();

    // เตรียมฟอนต์
    final font = pw.Font.ttf(fontData.buffer.asByteData());
    final fontRegula = pw.Font.ttf(fontDataRegula.buffer.asByteData());
    final fontBold = pw.Font.ttf(fontDataBold.buffer.asByteData());

    // สร้างธีม
    final theme = _buildTheme(pageFormat, font, fontBold, fontRegula);

    // คำนวณค่าต่างๆ เพียงครั้งเดียว
    final headerWidget = _buildHeader(logoBytes);
    final footerWidget = _buildFooter();

    // แบ่งรายการสินค้าเป็นหน้าๆ (แต่ละหน้าประมาณ 25-30 รายการ)
    final itemsPerPage = 25;
    List<List<dynamic>> pages = [];

    if (screenData.details != null &&
        screenData.details!.length > itemsPerPage) {
      pages = _paginateItems(screenData.details!, itemsPerPage);
    } else {
      // ถ้ามีน้อยกว่า itemsPerPage รายการ ให้ใส่ทั้งหมดในหน้าเดียว
      if (screenData.details != null && screenData.details!.isNotEmpty) {
        pages.add(screenData.details!);
      } else {
        pages.add([]);
      }
    }

    // สร้างทุกหน้า
    for (int pageIndex = 0; pageIndex < pages.length; pageIndex++) {
      doc.addPage(
        pw.Page(
          pageTheme: theme,
          build: (pw.Context context) {
            // คำนวณหมายเลขหน้า
            final isLastPage = pageIndex == pages.length - 1;

            return pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              children: [
                // ส่วนหัว
                headerWidget(context, pageIndex + 1, pages.length),

                // ตารางรายการสินค้า
                _buildContentTable(
                    context, pages[pageIndex], pageIndex * itemsPerPage),

                // เพิ่ม footer สำหรับหน้าสุดท้ายเท่านั้น
                if (isLastPage) pw.SizedBox(height: 10),
                if (isLastPage) _buildContentFooter(context),
              ],
            );
          },
        ),
      );
    }

    // Return the PDF file content
    return doc.save();
  }

  // สร้างฟังก์ชันเพื่อส่งคืนเป็น callback (ลดการสร้าง widget ซ้ำ)
  pw.Widget Function(pw.Context, int, int) _buildHeader(Uint8List? logoBytes) {
    return (pw.Context context, int pageNumber, int pageCount) {
      return pw.Column(
        children: [
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            crossAxisAlignment: pw.CrossAxisAlignment.start,
            children: [
              // ข้อมูลบริษัท
              pw.Container(
                child: pw.Row(
                  crossAxisAlignment: pw.CrossAxisAlignment.center,
                  children: [
                    (logoBytes != null)
                        ? pw.Image(pw.MemoryImage(logoBytes),
                            width: 80, height: 50)
                        : pw.Container(),
                    pw.SizedBox(width: 10),
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        _text(
                            global
                                .activeLangName(companyBranchData.companynames),
                            bold: true,
                            size: 16),
                        _text(
                            global.activeLangName(
                                companyBranchData.contact!.address!),
                            size: 14),
                        (companyBranchData.code == "00000")
                            ? _text(
                                '${global.language("pdf_tax_id")} ${companyBranchData.pos!.taxid} (${global.activeLangName(companyBranchData.names)})',
                                size: 14)
                            : _text(
                                '${global.language("pdf_tax_id")} ${companyBranchData.pos!.taxid} (${global.language("branch")} ~ ${global.activeLangName(companyBranchData.names)})',
                                size: 14),
                      ],
                    ),
                  ],
                ),
              ),

              // QR Code และหมายเลขหน้า
              pw.Column(
                children: [
                  pw.Container(
                    width: 40,
                    height: 40,
                    child: pw.BarcodeWidget(
                      data: screenData.docno,
                      barcode: pw.Barcode.qrCode(),
                      drawText: false,
                    ),
                  ),
                  pw.SizedBox(height: 5),
                  _text('${global.language("page_number")} $pageNumber/$pageCount', size: 12),
                ],
              ),
            ],
          ),

          // หัวเอกสาร
          pw.Padding(
            padding: pw.EdgeInsets.symmetric(vertical: 5),
            child: pw.Text(
              global.language('goods_issue'),
              style: pw.TextStyle(
                fontSize: 20,
                fontWeight: pw.FontWeight.bold,
              ),
            ),
          ),

          // แถวข้อมูลเอกสาร
          pw.Padding(
            padding: pw.EdgeInsets.symmetric(vertical: 0),
            child: pw.Row(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
              children: [
                // ด้านซ้าย
                pw.Container(
                  width: 160,
                  padding: pw.EdgeInsets.all(0),
                  child: pw.Column(
                    mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                    crossAxisAlignment: pw.CrossAxisAlignment.end,
                    children: [
                      pw.Row(
                        children: [
                          pw.Expanded(
                            child: _text(global.language('document_date'), bold: true, size: 12),
                          ),
                          pw.SizedBox(width: 10),
                          pw.Expanded(
                            child: _text(
                                (global.profileData.yeartype == "buddhist")
                                    ? global.dateTimeBuddhist(
                                        DateTime.parse(screenData.docdatetime),
                                        format: global.DateTimeFormatEnum.date)
                                    : DateFormat('dd/MM/yyyy').format(
                                        DateTime.parse(screenData.docdatetime)),
                                size: 12),
                          ),
                        ],
                      ),
                      pw.Row(
                        children: [
                          pw.Expanded(
                            child: _text(global.language('docno'), bold: true, size: 12),
                          ),
                          pw.SizedBox(width: 10),
                          pw.Expanded(
                            child: _text(screenData.docno, size: 12),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                // ด้านขวา
                pw.Container(
                  width: 160,
                  padding: pw.EdgeInsets.all(0),
                  child: pw.Column(
                    mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                    crossAxisAlignment: pw.CrossAxisAlignment.end,
                    children: [
                      pw.Row(
                        children: [
                          pw.Expanded(
                            child: _text(
                                global.language('text_form_ref_invoice_date'),
                                bold: true,
                                size: 12),
                          ),
                          pw.SizedBox(width: 10),
                          pw.Expanded(
                            child: _text(
                                (global.profileData.yeartype == "buddhist")
                                    ? global.dateTimeBuddhist(
                                        DateTime.parse(screenData.docrefdate),
                                        format: global.DateTimeFormatEnum.date)
                                    : DateFormat('dd/MM/yyyy').format(
                                        DateTime.parse(screenData.docrefdate)),
                                size: 12),
                          ),
                        ],
                      ),
                      pw.Row(
                        children: [
                          pw.Expanded(
                            child: _text(
                                global.language('text_form_ref_invoice_number'),
                                bold: true,
                                size: 12),
                          ),
                          pw.SizedBox(width: 10),
                          pw.Expanded(
                            child: _text(screenData.docrefno, size: 12),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      );
    };
  }

  pw.Widget Function(pw.Context) _buildFooter() {
    return (pw.Context context) {
      return pw.Column(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: [
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Expanded(
                child: pw.Column(
                  crossAxisAlignment: pw.CrossAxisAlignment.center,
                  children: [
                    pw.Text(
                      '${global.language("pdf_requester")}................................................................',
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                      ),
                    ),
                    pw.Text(
                      global.language('pdf_date_sign_short'),
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                      ),
                    ),
                  ],
                ),
              ),
              pw.SizedBox(width: 10),
              pw.Expanded(
                child: pw.Column(
                  crossAxisAlignment: pw.CrossAxisAlignment.center,
                  children: [
                    pw.Text(
                      '${global.language("pdf_recorder")}...............................................................',
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                      ),
                    ),
                    pw.Text(
                      global.language('pdf_date_sign_medium'),
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                      ),
                    ),
                  ],
                ),
              ),
              pw.SizedBox(width: 10),
              pw.Expanded(
                child: pw.Column(
                  crossAxisAlignment: pw.CrossAxisAlignment.center,
                  children: [
                    pw.Text(
                      global.language('pdf_approver_sign_long'),
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                      ),
                    ),
                    pw.Text(
                      global.language('pdf_date_sign_short'),
                      style: const pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      );
    };
  }

  pw.Widget _buildContentFooter(pw.Context context) {
    return pw.Column(
      children: [
        pw.Row(
          crossAxisAlignment: pw.CrossAxisAlignment.start,
          children: [
            pw.Column(
              children: [
                pw.Container(
                  width: 340,
                  child: pw.Text(
                    '${global.language("disciption")} : ${screenData.description}',
                    style: pw.TextStyle(
                      fontSize: 14,
                      color: PdfColors.black,
                      fontWeight: pw.FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            pw.SizedBox(width: 5),
            pw.Container(
              width: 186,
              height: 140,
              child: pw.DefaultTextStyle(
                style: const pw.TextStyle(
                  fontSize: 14,
                  color: PdfColors.black,
                ),
                child: pw.Column(
                  crossAxisAlignment: pw.CrossAxisAlignment.start,
                  children: [
                    pw.DefaultTextStyle(
                      style: pw.TextStyle(
                        color: PdfColors.black,
                        fontSize: 16,
                        fontWeight: pw.FontWeight.bold,
                      ),
                      child: pw.Container(
                        child: pw.Container(
                          margin: const pw.EdgeInsets.only(bottom: 3, top: 3),
                          decoration: const pw.BoxDecoration(
                            border: pw.Border(
                              bottom: pw.BorderSide(
                                color: PdfColors.black,
                                width: 1,
                              ),
                            ),
                          ),
                          child: pw.Row(
                            mainAxisAlignment:
                                pw.MainAxisAlignment.spaceBetween,
                            children: [
                              pw.Text(global.language('total_qty')),
                              pw.Text(
                                  global.formatNumber(screenData.totalqty!)),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  // ปรับปรุงตารางเพื่อการแสดงผลที่เร็วขึ้น โดยใช้ข้อมูลเพียงบางส่วนในแต่ละหน้า
  pw.Widget _buildContentTable(
      pw.Context context, List<dynamic> pageItems, int startIndex) {
    final tableHeaders = [
      global.language('group_number'),
      global.language('barcode'),
      global.language('doc_details'),
      global.language('warehouse'),
      global.language('location'),
      global.language('chatbot_product_unit'),
      global.language('enter_qty')
    ];

    final headerStyle = pw.TextStyle(
        color: PdfColors.black, fontSize: 12, fontWeight: pw.FontWeight.bold);
    const cellStyle = pw.TextStyle(color: PdfColors.black, fontSize: 10);

    var headerDecoration = const pw.BoxDecoration(
      borderRadius: pw.BorderRadius.all(pw.Radius.circular(2)),
      color: PdfColors.grey300,
    );

    const rowDecoration = pw.BoxDecoration(
      border: pw.Border(
        bottom: pw.BorderSide(
          color: PdfColors.grey,
          width: .5,
        ),
      ),
    );

    // สร้างข้อมูลสำหรับเพียงรายการในหน้านี้เท่านั้น
    final tableData = List<List<String>>.generate(
      pageItems.length,
      (row) => List<String>.generate(
        tableHeaders.length,
        (col) {
          final itemIndex = row;
          switch (col) {
            case 0:
              return '${startIndex + row + 1}';
            case 1:
              return pageItems[itemIndex].barcode;
            case 2:
              return global.activeLangName(pageItems[itemIndex].itemnames!);
            case 3:
              return global.activeLangName(pageItems[itemIndex].whnames!);
            case 4:
              return global.activeLangName(pageItems[itemIndex].locationnames!);
            case 5:
              return global.activeLangName(pageItems[itemIndex].unitnames!);
            case 6:
              return global.formatNumber(pageItems[itemIndex].qty);
            default:
              return '';
          }
        },
      ),
    );

    return pw.TableHelper.fromTextArray(
      border: null,
      cellAlignment: pw.Alignment.centerLeft,
      headerDecoration: headerDecoration,
      headerHeight: 25,
      cellHeight: 0,
      headerAlignments: {
        0: pw.Alignment.center,
        1: pw.Alignment.center,
        2: pw.Alignment.center,
        3: pw.Alignment.center,
        4: pw.Alignment.center,
        5: pw.Alignment.center,
        6: pw.Alignment.center,
      },
      cellAlignments: {
        0: pw.Alignment.center,
        1: pw.Alignment.centerLeft,
        2: pw.Alignment.centerLeft,
        3: pw.Alignment.center,
        4: pw.Alignment.center,
        5: pw.Alignment.center,
        6: pw.Alignment.centerRight,
      },
      headerStyle: headerStyle,
      cellStyle: cellStyle,
      rowDecoration: rowDecoration,
      headers: tableHeaders,
      data: tableData,
    );
  }

  pw.Text _text(String data, {bool bold = false, double size = 8}) {
    return pw.Text(
      data,
      style: pw.TextStyle(
        fontSize: size,
        fontWeight: bold ? pw.FontWeight.bold : pw.FontWeight.normal,
      ),
    );
  }
}

pw.PageTheme _buildTheme(
    PdfPageFormat pageFormat, pw.Font base, pw.Font bold, pw.Font italic) {
  return pw.PageTheme(
    pageFormat: pageFormat,
    theme: pw.ThemeData.withFont(
      base: base,
      bold: bold,
      italic: italic,
    ),
  );
}
