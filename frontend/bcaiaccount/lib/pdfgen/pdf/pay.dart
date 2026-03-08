import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/debtor_creditor_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:smlaicloud/global.dart' as global;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

Future<Uint8List> generatePay(PdfPageFormat pageFormat, TransactionPaidPayModel screenData, CompanyBranchModel companyBranchData, DebtorCreditorModel debtorCreditorData) async {
  final pay = Pay(companyBranchData: companyBranchData, screenData: screenData, debtorCreditorData: debtorCreditorData);

  return await pay.buildPdf(pageFormat);
}

class Pay {
  Pay({required this.companyBranchData, required this.screenData, required this.debtorCreditorData});

  final CompanyBranchModel companyBranchData;
  final TransactionPaidPayModel screenData;
  final DebtorCreditorModel debtorCreditorData;

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
    final fontDataRegula = await rootBundle.load("assets/fonts/THSarabunNew-Italic.ttf");
    final fontDataBold = await rootBundle.load("assets/fonts/THSarabunNew-Bold.ttf");
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
        build: (context) => [_contentTable(context, screenData), pw.SizedBox(height: 10), _contentFooter(context, screenData)],
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
                  pw.Text(global.language('pdf_approver_sign'), style: pw.TextStyle(fontSize: 14, color: PdfColors.black)),
                  pw.Text(global.language('pdf_date_sign'), style: pw.TextStyle(fontSize: 14, color: PdfColors.black)),
                ],
              ),
            ),
            pw.SizedBox(width: 10),
            pw.Expanded(
              child: pw.Column(
                crossAxisAlignment: pw.CrossAxisAlignment.center,
                children: [
                  pw.Text(global.language('pdf_payer_sign'), style: pw.TextStyle(fontSize: 14, color: PdfColors.black)),
                  pw.Text(global.language('pdf_date_sign'), style: pw.TextStyle(fontSize: 14, color: PdfColors.black)),
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
          padding: pw.EdgeInsets.symmetric(vertical: 5),
          child: pw.Text(global.language('pdf_payment_voucher'), style: pw.TextStyle(fontSize: 20, fontWeight: pw.FontWeight.bold)),
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
              child: pw.BarcodeWidget(data: screenData.docno, barcode: pw.Barcode.qrCode(), drawText: false),
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
        children: [if (logoBytes != null) pw.Image(pw.MemoryImage(logoBytes), width: 80, height: 50), pw.SizedBox(width: 10), _buildCompanyInfoColumn()],
      ),
    );
  }

  pw.Widget _buildCompanyInfoColumn() {
    return pw.Column(
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      children: [
        _text(global.activeLangName(companyBranchData.companynames), bold: true, size: 16),
        _text(global.activeLangName(companyBranchData.contact!.address!), size: 14),
        (companyBranchData.code == "00000")
            ? _text('${global.language("pdf_tax_id")} ${companyBranchData.pos!.taxid} (${global.activeLangName(companyBranchData.names)})', size: 14)
            : _text('${global.language("pdf_tax_id")} ${companyBranchData.pos!.taxid} (${global.language("branch")} ~ ${global.activeLangName(companyBranchData.names)})', size: 14),
      ],
    );
  }

  pw.Widget _buildPageInfo(pw.Context context) {
    return _text('${global.language("page_number")} ${context.pageNumber}/${context.pagesCount}', size: 12);
  }

  pw.Widget _buildHeaderBottomRow() {
    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 0),
      child: pw.Row(crossAxisAlignment: pw.CrossAxisAlignment.start, mainAxisAlignment: pw.MainAxisAlignment.spaceBetween, children: [_buildCustomerDetails(), _buildDocumentDetails()]),
    );
  }

  pw.Widget _buildCustomerDetails() {
    return _buildDetailBox(
      width: 337,
      children: [
        _detailRowCust(global.language('supplier'), "${screenData.custcode} ~ ${global.activeLangName(screenData.custnames!)}"),
        _detailRowCust(global.language('address'), debtorCreditorData.addressforbilling.address!.isNotEmpty ? debtorCreditorData.addressforbilling.address![0] : '', isMultiline: true),
        _detailRowCust(global.language('pdf_tax_id_number'), debtorCreditorData.taxid!),
      ],
    );
  }

  pw.Widget _buildDocumentDetails() {
    return _buildDetailBox(
      width: 140,
      children: [
        _detailRow(
          global.language('document_date'),
          (global.profileData.yeartype == "buddhist")
              ? global.dateTimeBuddhist(DateTime.parse(screenData.docdatetime), format: global.DateTimeFormatEnum.date)
              : DateFormat('dd/MM/yyyy').format(DateTime.parse(screenData.docdatetime)),
        ),
        _detailRow(global.language('docno'), screenData.docno),
      ],
    );
  }

  pw.Widget _buildDetailBox({required double width, required List<pw.Widget> children}) {
    return pw.Container(
      width: width,
      padding: const pw.EdgeInsets.all(0),
      child: pw.Column(mainAxisAlignment: pw.MainAxisAlignment.spaceBetween, crossAxisAlignment: pw.CrossAxisAlignment.end, children: children),
    );
  }

  pw.Widget _detailRow(String label, String value) {
    return pw.Row(
      children: [
        pw.Expanded(child: _text(label, bold: true, size: 12)),
        pw.SizedBox(width: 10),
        pw.Expanded(child: _text(value, size: 12)),
      ],
    );
  }

  pw.Widget _detailRowCust(String label, String value, {bool isMultiline = false}) {
    return !isMultiline
        ? pw.Row(children: [_text(label, bold: true, size: 12), pw.SizedBox(width: 10), _text(value, size: 12)])
        : pw.Row(
            children: [
              pw.Container(height: 30, child: _text(label, bold: true, size: 12)),
              pw.SizedBox(width: 10),
              pw.Container(height: 30, child: _text(value, size: 12)),
            ],
          );
  }

  pw.Text _text(String data, {bool bold = false, double size = 8}) {
    return pw.Text(
      data,
      style: pw.TextStyle(fontSize: size, fontWeight: bold ? pw.FontWeight.bold : pw.FontWeight.normal),
    );
  }
}

pw.PageTheme _buildTheme(PdfPageFormat pageFormat, pw.Font base, pw.Font bold, pw.Font italic) {
  return pw.PageTheme(
    pageFormat: pageFormat,
    theme: pw.ThemeData.withFont(base: base, bold: bold, italic: italic),
  );
}

pw.Widget _contentFooter(pw.Context context, TransactionPaidPayModel screenData) {
  return pw.Column(
    children: [
      pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Expanded(
            child: pw.Container(
              // color: PdfColors.red,
              child: pw.Column(
                mainAxisAlignment: pw.MainAxisAlignment.end,
                children: [
                  pw.Container(
                    child: pw.Column(
                      mainAxisAlignment: pw.MainAxisAlignment.end,
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Container(
                          alignment: pw.Alignment.center,
                          child: pw.Text(
                            global.language("text_form_text_money"),
                            style: pw.TextStyle(fontSize: 14, color: PdfColors.black, fontWeight: pw.FontWeight.bold),
                          ),
                        ),
                        pw.Container(
                          color: PdfColors.grey300,
                          alignment: pw.Alignment.center,
                          child: pw.Text(
                            '=${global.numberToWordByLang(screenData.totalamount != 0 ? screenData.totalamount : screenData.totalvalue)}=',
                            style: const pw.TextStyle(fontSize: 14, color: PdfColors.black),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          pw.SizedBox(width: 5),
          pw.Container(
            // color: PdfColors.yellow100,
            padding: const pw.EdgeInsets.only(top: 9),
            width: 186,
            child: pw.DefaultTextStyle(
              style: pw.TextStyle(fontSize: 14, color: PdfColors.black),
              child: pw.Column(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  pw.DefaultTextStyle(
                    style: pw.TextStyle(color: PdfColors.black, fontSize: 16, fontWeight: pw.FontWeight.bold),
                    child: pw.Container(
                      decoration: const pw.BoxDecoration(
                        border: pw.Border(bottom: pw.BorderSide(color: PdfColors.black, width: 1)),
                      ),
                      child: pw.Container(
                        margin: const pw.EdgeInsets.only(bottom: 3, top: 3),
                        decoration: const pw.BoxDecoration(
                          border: pw.Border(bottom: pw.BorderSide(color: PdfColors.black, width: 1)),
                        ),
                        child: pw.Row(
                          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                          children: [pw.Text(global.language('net_value')), pw.Text(global.formatNumber(screenData.totalamount != 0 ? screenData.totalamount : screenData.totalvalue))],
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

pw.Widget _contentTable(pw.Context context, TransactionPaidPayModel screenData) {
  final tableHeaders = [global.language('group_number'), global.language('vat_num'), global.language('date'), global.language('cash_amount'), global.language('stock_balance'), global.language('payment_amount')];

  final headerStyle = pw.TextStyle(color: PdfColors.black, fontSize: 14, fontWeight: pw.FontWeight.bold);
  const cellStyle = pw.TextStyle(color: PdfColors.black, fontSize: 12);
  var headerDecoration = const pw.BoxDecoration(borderRadius: pw.BorderRadius.all(pw.Radius.circular(2)), color: PdfColors.grey300);
  const rowDecoration = pw.BoxDecoration(
    border: pw.Border(bottom: pw.BorderSide(color: PdfColors.grey, width: .5)),
  );

  return pw.TableHelper.fromTextArray(
    border: null,
    cellAlignment: pw.Alignment.centerLeft,
    headerDecoration: headerDecoration,
    headerHeight: 25,
    cellHeight: 0,
    headerAlignments: {0: pw.Alignment.center, 1: pw.Alignment.center, 2: pw.Alignment.center, 3: pw.Alignment.center, 4: pw.Alignment.center, 5: pw.Alignment.center},
    cellAlignments: {0: pw.Alignment.center, 1: pw.Alignment.centerLeft, 2: pw.Alignment.centerLeft, 3: pw.Alignment.center, 4: pw.Alignment.centerRight, 5: pw.Alignment.centerRight},
    headerStyle: headerStyle,
    cellStyle: cellStyle,
    rowDecoration: rowDecoration,
    headers: tableHeaders,
    data: List<List<String>>.generate(screenData.details!.where((detail) => detail.selected!).length, (row) {
      var filteredDetail = screenData.details!.where((detail) => detail.selected!).toList()[row];
      return List<String>.generate(tableHeaders.length, (col) => _cellValue(row, col, filteredDetail));
    }),
  );
}

String _cellValue(int rowIndex, int columnIndex, TransactionPaidPayDetailModel filteredDetail) {
  switch (columnIndex) {
    case 0:
      return '${rowIndex + 1}';
    case 1:
      return filteredDetail.docno;
    case 2:
      return (global.profileData.yeartype == "buddhist")
          ? global.dateTimeBuddhist(DateTime.parse(filteredDetail.docdatetime), format: global.DateTimeFormatEnum.date)
          : DateFormat('dd/MM/yyyy').format(DateTime.parse(filteredDetail.docdatetime));
    case 3:
      return global.formatNumber(filteredDetail.value);
    case 4:
      return global.formatNumber(filteredDetail.balance);
    case 5:
      return global.formatNumber(filteredDetail.paymentamount);
    default:
      return '';
  }
}
