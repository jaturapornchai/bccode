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

Future<Uint8List> generateStockAdjustment(
  PdfPageFormat pageFormat,
  TransactionModel screenData,
  CompanyBranchModel companyBranchData,
) async {
  final stockTransfer = StockTransfer(
    companyBranchData: companyBranchData,
    screenData: screenData,
  );

  return await stockTransfer.buildPdf(pageFormat);
}

class StockTransfer {
  StockTransfer({
    required this.companyBranchData,
    required this.screenData,
  });

  final CompanyBranchModel companyBranchData;
  final TransactionModel screenData;

  Future<Uint8List> buildPdf(PdfPageFormat pageFormat) async {
    // Create a PDF document.
    final doc = pw.Document();

    // Fetch the logo image from the network
    Uint8List? logoBytes;

    if (companyBranchData.logouri!.isNotEmpty) {
      try {
        logoBytes = await global.fetchNetworkImage(companyBranchData.logouri!);
      } catch (e) {
        if (kDebugMode) {
          AppLogger.debug(e);
        }
      }
    }

    final fontData = await rootBundle.load("assets/fonts/THSarabunNew.ttf");
    final fontDataRegula =
        await rootBundle.load("assets/fonts/THSarabunNew-Italic.ttf");
    final fontDataBold =
        await rootBundle.load("assets/fonts/THSarabunNew-Bold.ttf");
    final font = pw.Font.ttf(fontData);
    final fontRegula = pw.Font.ttf(fontDataRegula);
    final fontBold = pw.Font.ttf(fontDataBold);

    // Add page to the PDF
    doc.addPage(
      pw.MultiPage(
        pageTheme: _buildTheme(pageFormat, font, fontBold, fontRegula),
        mainAxisAlignment: pw.MainAxisAlignment.start,
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        header: (context) => _buildHeader(context, logoBytes),
        build: (context) => [
          _contentTable(context, screenData),
          pw.SizedBox(height: 10),
          _contentFooter(context, screenData),
        ],
        footer: (context) => _buildFooter(context),
      ),
    );

    // Return the PDF file content
    return doc.save();
  }

  pw.Widget _buildFooter(pw.Context context) {
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
                    '${global.language("pdf_recorder")}................................................................',
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
  }

  pw.Widget _buildHeader(pw.Context context, Uint8List? logoBytes) {
    return pw.Column(
      children: [
        _buildHeaderTopRow(context, logoBytes),
        pw.Padding(
          padding: const pw.EdgeInsets.symmetric(vertical: 5),
          child: pw.Text(
            global.language("pdf_stock_adjustment"),
            style: pw.TextStyle(
              fontSize: 20,
              fontWeight: pw.FontWeight.bold,
            ),
          ),
        ),
        _buildHeaderBottomRow(),
      ],
    );
  }

  pw.Widget _buildHeaderTopRow(pw.Context context, Uint8List? logoBytes) {
    return pw.Row(
      mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      children: [
        _buildCompanyDetails(logoBytes),
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

            /// sizebox
            pw.SizedBox(height: 5),
            _buildPageInfo(context),
          ],
        ),
      ],
    );
  }

  pw.Widget _buildCompanyDetails(Uint8List? logoBytes) {
    return pw.Container(
      child: pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.center,
        children: [
          (logoBytes != null)
              ? pw.Image(pw.MemoryImage(logoBytes), width: 80, height: 50)
              : pw.Container(),
          pw.SizedBox(width: 10),
          _buildCompanyInfoColumn(),
        ],
      ),
    );
  }

  pw.Widget _buildCompanyInfoColumn() {
    return pw.Column(
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      children: [
        _text(global.activeLangName(companyBranchData.companynames),
            bold: true, size: 16),
        _text(global.activeLangName(companyBranchData.contact!.address!),
            size: 14),
        (companyBranchData.code == "00000")
            ? _text(
                '${global.language("pdf_tax_id")} ${companyBranchData.pos!.taxid} (${global.activeLangName(companyBranchData.names)})',
                size: 14)
            : _text(
                '${global.language("pdf_tax_id")} ${companyBranchData.pos!.taxid} (${global.language("branch")} ~ ${global.activeLangName(companyBranchData.names)})',
                size: 14),
      ],
    );
  }

  pw.Widget _buildPageInfo(pw.Context context) {
    return _text('${global.language("page_number")} ${context.pageNumber}/${context.pagesCount}', size: 12);
  }

  pw.Widget _buildHeaderBottomRow() {
    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 0),
      child: pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: [
          _buildDocumentDetailsLeft(),
          _buildDocumentDetailsRight(),
        ],
      ),
    );
  }

  pw.Widget _buildDocumentDetailsLeft() {
    return _buildDetailBox(
      width: 160,
      children: [
        _detailRow(
          global.language('document_date'),
          (global.profileData.yeartype == "buddhist")
              ? global.dateTimeBuddhist(DateTime.parse(screenData.docdatetime),
                  format: global.DateTimeFormatEnum.date)
              : DateFormat('dd/MM/yyyy')
                  .format(DateTime.parse(screenData.docdatetime)),
        ),
        _detailRow(global.language('docno'), screenData.docno),
        _detailRow(
            global.language('adjust_type'),
            (screenData.transflag == 66)
                ? global.language('increase')
                : (screenData.transflag) == 68
                    ? global.language('decrease')
                    : (screenData.transflag) == 966
                        ? global.language('adjust_cost')
                        : 'not found'),
      ],
    );
  }

  pw.Widget _buildDocumentDetailsRight() {
    return _buildDetailBox(
      width: 160,
      children: [
        _detailRow(
          global.language('text_form_ref_invoice_date'),
          (global.profileData.yeartype == "buddhist")
              ? global.dateTimeBuddhist(DateTime.parse(screenData.docrefdate),
                  format: global.DateTimeFormatEnum.date)
              : DateFormat('dd/MM/yyyy')
                  .format(DateTime.parse(screenData.docrefdate)),
        ),
        _detailRow(global.language('text_form_ref_invoice_number'),
            screenData.docrefno),
      ],
    );
  }

  pw.Widget _buildDetailBox(
      {required double width, required List<pw.Widget> children}) {
    return pw.Container(
      // color: PdfColors.grey300,
      width: width,
      padding: const pw.EdgeInsets.all(0),
      child: pw.Column(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        crossAxisAlignment: pw.CrossAxisAlignment.end,
        children: children,
      ),
    );
  }

  pw.Widget _detailRow(String label, String value) {
    return pw.Row(
      children: [
        pw.Expanded(
          child: _text(label, bold: true, size: 12),
        ),
        pw.SizedBox(width: 10),
        pw.Expanded(
          child: _text(value, size: 12),
        ),
      ],
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

pw.Widget _contentFooter(pw.Context context, TransactionModel screenData) {
  return pw.Column(
    children: [
      pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Column(
            children: [
              pw.Container(
                // color: PdfColors.grey300,
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
          pw.SizedBox(
            width: 5,
          ),
          pw.Container(
            // color: PdfColors.yellow100,
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
                          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                          children: [
                            pw.Text(global.language('total_qty')),
                            pw.Text(global.formatNumber(screenData.totalqty!)),
                          ],
                        ), // Replace with your text
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

pw.Widget _contentTable(pw.Context context, TransactionModel screenData) {
  final tableHeaders = [
    global.language('group_number'),
    global.language('barcode'),
    global.language('doc_details'),
    global.language('warehouse'),
    global.language('location'),
    global.language('chatbot_product_unit'),
    global.language('enter_qty'),
    global.language('cost'),
    global.language('pdf_total_cost')
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
      7: pw.Alignment.center,
      8: pw.Alignment.center,
    },
    cellAlignments: {
      0: pw.Alignment.center,
      1: pw.Alignment.centerLeft,
      2: pw.Alignment.centerLeft,
      3: pw.Alignment.center,
      4: pw.Alignment.center,
      5: pw.Alignment.center,
      6: pw.Alignment.centerRight,
      7: pw.Alignment.centerRight,
      8: pw.Alignment.centerRight,
    },
    // columnWidths: {
    //   0: const pw.FixedColumnWidth(10),
    //   1: const pw.FixedColumnWidth(30),
    //   2: const pw.FixedColumnWidth(10),
    //   3: const pw.FixedColumnWidth(10),
    //   4: const pw.FixedColumnWidth(15),
    //   5: const pw.FixedColumnWidth(15),
    //   6: const pw.FixedColumnWidth(15),
    // },
    headerStyle: headerStyle,
    cellStyle: cellStyle,
    rowDecoration: rowDecoration,
    headers: tableHeaders,
    data: List<List<String>>.generate(
      screenData.details!.length,
      (row) => List<String>.generate(
        tableHeaders.length,
        (col) => _cellValue(row, col, screenData),
      ),
    ),
  );
}

String _cellValue(int rowIndex, int columnIndex, TransactionModel screenData) {
  switch (columnIndex) {
    case 0:
      return '${rowIndex + 1}';
    case 1:
      return screenData.details![rowIndex].barcode;
    case 2:
      return global.activeLangName(screenData.details![rowIndex].itemnames!);
    case 3:
      return global.activeLangName(screenData.details![rowIndex].whnames!);
    case 4:
      return global
          .activeLangName(screenData.details![rowIndex].locationnames!);
    case 5:
      return global.activeLangName(screenData.details![rowIndex].unitnames!);
    case 6:
      return global.formatNumber(screenData.details![rowIndex].qty);
    case 7:
      return global.formatNumber(screenData.details![rowIndex].price);
    case 8:
      return global.formatNumber(screenData.details![rowIndex].sumamount);
    default:
      return '';
  }
}
