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

Future<Uint8List> generatePurchaseReturn(
  PdfPageFormat pageFormat,
  TransactionModel screenData,
  CompanyBranchModel companyBranchData,
  DebtorCreditorModel debtorCreditorData,
) async {
  final purchaseReturn = PurchaseReturn(
    companyBranchData: companyBranchData,
    screenData: screenData,
    debtorCreditorData: debtorCreditorData,
  );

  return await purchaseReturn.buildPdf(pageFormat);
}

class PurchaseReturn {
  PurchaseReturn({
    required this.companyBranchData,
    required this.screenData,
    required this.debtorCreditorData,
  });

  final CompanyBranchModel companyBranchData;
  final TransactionModel screenData;
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
        header: (context) => pw.Padding(
          padding: const pw.EdgeInsets.only(left: 0),
          child: _buildHeader(context, logoBytes),
        ),
        build: (context) => [
          pw.Padding(
            padding: const pw.EdgeInsets.only(left: 0, top: 10),
            child: _contentTable(context, screenData),
          ),
          pw.SizedBox(height: 10),
          pw.Padding(
            padding: const pw.EdgeInsets.only(left: 0),
            child: _contentFooter(context, screenData),
          ),
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
                    global.language('pdf_sender_sign'),
                    style: const pw.TextStyle(
                      fontSize: 14,
                      color: PdfColors.black,
                    ),
                  ),
                  pw.Text(
                    global.language('pdf_date_sign'),
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
                    global.language('pdf_receiver_goods_sign'),
                    style: const pw.TextStyle(
                      fontSize: 14,
                      color: PdfColors.black,
                    ),
                  ),
                  pw.Text(
                    global.language('pdf_date_sign'),
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
            global.language('pdf_purchase_return_doc'),
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
          if (logoBytes != null)
            pw.Image(pw.MemoryImage(logoBytes), width: 80, height: 50),
          if (logoBytes != null) pw.SizedBox(width: 10),
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
    return pw.Row(
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
      children: [
        _buildCustomerDetails(),
        _buildDocumentDetails(),
      ],
    );
  }

  pw.Widget _buildCustomerDetails() {
    return _buildDetailBox(
      width: 337,
      children: [
        _detailRowCust(global.language('supplier'),
            "${screenData.custcode} ~ ${global.activeLangName(screenData.custnames!)}"),
        _detailRowCust(
          global.language('address'),
          debtorCreditorData.addressforbilling.address!.isNotEmpty
              ? debtorCreditorData.addressforbilling.address![0]
              : '',
          isMultiline: true,
        ),
        _detailRowCust(global.language('pdf_tax_id_number'), debtorCreditorData.taxid!),
      ],
    );
  }

  pw.Widget _buildDocumentDetails() {
    // ตรวจสอบว่ามีข้อมูล details และ docref หรือไม่
    String docref = '';
    String docrefDateTime = '';

    if (screenData.details != null && screenData.details!.isNotEmpty) {
      docref = screenData.details!.first.docref ?? '';
      docrefDateTime = screenData.details!.first.docrefdatetime ?? '';
    }

    return _buildDetailBox(
      width: 160,
      children: [
        _detailRow(global.language('vat_num'), screenData.docno),
        _detailRow(
          global.language('pdf_vat_date'),
          (global.profileData.yeartype == "buddhist")
              ? global.dateTimeBuddhist(DateTime.parse(screenData.docdatetime),
                  format: global.DateTimeFormatEnum.date)
              : DateFormat('dd/MM/yyyy')
                  .format(DateTime.parse(screenData.docdatetime)),
        ),
        if (docref.isNotEmpty) _detailRow(global.language('pdf_ref_original_tax_invoice'), docref),
        if (docrefDateTime.isNotEmpty)
          _detailRow(
            global.language('pdf_due_date'),
            (global.profileData.yeartype == "buddhist")
                ? global.dateTimeBuddhist(DateTime.parse(docrefDateTime),
                    format: global.DateTimeFormatEnum.date)
                : DateFormat('dd/MM/yyyy')
                    .format(DateTime.parse(docrefDateTime)),
          ),
      ],
    );
  }

  pw.Widget _buildDetailBox(
      {required double width, required List<pw.Widget> children}) {
    return pw.Container(
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

  pw.Widget _detailRowCust(String label, String value,
      {bool isMultiline = false}) {
    return !isMultiline
        ? pw.Row(
            children: [
              _text(label, bold: true, size: 12),
              pw.SizedBox(width: 10),
              _text(value, size: 12),
            ],
          )
        : pw.Row(
            children: [
              pw.Container(
                height: 30,
                child: _text(label, bold: true, size: 12),
              ),
              pw.SizedBox(width: 10),
              pw.Container(
                height: 30,
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
  /// รวมสินค้ามีภาษี
  double amountVatCale0 = 0;

  /// รวมสินค้ายกเว้นภาษี
  double amountVatCale1 = 0;

  for (var val in screenData.details!) {
    if (val.vatcal == 0) {
      amountVatCale0 += val.sumamount;
    } else if (val.vatcal == 1) {
      amountVatCale1 += val.sumamount;
    }
  }

  return pw.Column(
    children: [
      pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Column(
            children: [
              pw.Container(
                color: PdfColors.grey300,
                width: 340,
                height: 96,
                child: pw.Column(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: pw.CrossAxisAlignment.start,
                  children: [
                    pw.Text(
                      '${global.language("disciption")} : ${screenData.description}',
                      style: pw.TextStyle(
                        fontSize: 14,
                        color: PdfColors.black,
                        fontWeight: pw.FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
              pw.Container(
                color: PdfColors.grey300,
                width: 340,
                height: 96,
                child: pw.Row(
                  crossAxisAlignment: pw.CrossAxisAlignment.end,
                  children: [
                    // Container for total amount in words
                    pw.Container(
                      color: PdfColors.red,
                      width: 205,
                      height: 50,
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
                                    style: pw.TextStyle(
                                      fontSize: 14,
                                      color: PdfColors.black,
                                      fontWeight: pw.FontWeight.bold,
                                    ),
                                  ),
                                ),
                                pw.Container(
                                  color: PdfColors.grey300,
                                  alignment: pw.Alignment.center,
                                  child: pw.Text(
                                    '=${global.numberToWordByLang(screenData.totalamount != 0 ? screenData.totalamount : screenData.totalvalue)}=',
                                    style: const pw.TextStyle(
                                      fontSize: 14,
                                      color: PdfColors.black,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    pw.SizedBox(
                      width: 5,
                    ),
                    // Container for tax-exempt and taxable items
                    pw.Container(
                      color: PdfColors.orange,
                      width: 130,
                      height: 50,
                      child: pw.DefaultTextStyle(
                        style: const pw.TextStyle(
                          fontSize: 8,
                          color: PdfColors.black,
                        ),
                        child: pw.Column(
                          mainAxisAlignment: pw.MainAxisAlignment.end,
                          children: [
                            pw.Row(
                              children: [
                                pw.Expanded(
                                  child: pw.Text(''),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.language('pdf_tax_exempt_goods'),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.language('pdf_taxable_goods'),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                              ],
                            ),
                            pw.Row(
                              children: [
                                pw.Expanded(
                                  child: pw.Text(global.language('receivelist_value')),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.formatNumber(amountVatCale1),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.formatNumber(amountVatCale0),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                              ],
                            ),
                            pw.Row(
                              mainAxisAlignment: pw.MainAxisAlignment.end,
                              children: [
                                pw.Expanded(
                                  child: pw.Text(global.language('pdf_discount')),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.formatNumber(screenData
                                        .totaldiscountexceptvatamount!),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.formatNumber(
                                        screenData.totaldiscountvatamount!),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                              ],
                            ),
                            pw.Row(
                              mainAxisAlignment: pw.MainAxisAlignment.end,
                              children: [
                                pw.Expanded(
                                  child: pw.Text(global.language('pdf_value_after_discount')),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.formatNumber(
                                        screenData.totalexceptvat),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                                pw.Expanded(
                                  child: pw.Text(
                                    global.formatNumber(
                                        screenData.totalbeforevat),
                                    textAlign: pw.TextAlign.right,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          pw.SizedBox(
            width: 5,
          ),
          pw.Container(
            color: PdfColors.yellow100,
            width: 186,
            height: 200,
            child: pw.DefaultTextStyle(
              style: const pw.TextStyle(
                fontSize: 14,
                color: PdfColors.black,
              ),
              child: pw.Column(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                children: [
                  pw.Row(
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('pdf_original_tax_invoice_value')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.reftotaloriginal!),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('text_form_ref_total_correct')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.reftotalcorrect!),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('text_form_ref_total_diff')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.reftotaldiff!),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('doc_total_value')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.totalvalue),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('pdf_discount')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.detailtotaldiscount!),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('pdf_total_after_discount')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.totalvalue -
                              screenData.detailtotaldiscount!),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('pdf_total_before_vat')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.totalbeforevat),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child:
                            pw.Text('${global.language("vat")} ${screenData.vatrate} %'),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.totalvatvalue),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  pw.Row(
                    mainAxisAlignment: pw.MainAxisAlignment.end,
                    children: [
                      pw.Expanded(
                        child: pw.Text(global.language('pdf_total_tax_exempt')),
                      ),
                      pw.Expanded(
                        child: pw.Text(
                          global.formatNumber(screenData.totalexceptvat),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                  // pw.Container(
                  //   decoration: const pw.BoxDecoration(
                  //     border: pw.Border(
                  //       bottom: pw.BorderSide(
                  //         color: PdfColors.black,
                  //         width: 1,
                  //       ),
                  //     ),
                  //   ),
                  //   child: pw.Row(
                  //     mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  //     children: [
                  //       pw.Text('หัก ส่วนลดท้ายบิล'),
                  //       pw.Text(global.formatNumber(screenData.totaldiscount)),
                  //     ],
                  //   ),
                  // ),
                  pw.DefaultTextStyle(
                    style: pw.TextStyle(
                      color: PdfColors.black,
                      fontSize: 16,
                      fontWeight: pw.FontWeight.bold,
                    ),
                    child: pw.Container(
                      decoration: const pw.BoxDecoration(
                        border: pw.Border(
                          bottom: pw.BorderSide(
                            color: PdfColors.black,
                            width: 1,
                          ),
                        ),
                      ),
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
                            pw.Text(global.language('net_value')),
                            pw.Text(
                                global.formatNumber(screenData.totalamount != 0 ? screenData.totalamount : screenData.totalvalue)),
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
    global.language('reference'),
    global.language('barcode'),
    global.language('doc_details'),
    global.language('enter_qty'),
    global.language('chatbot_product_unit'),
    global.language('price'),
    global.language('discount'),
    global.language('cash_amount')
  ];

  final headerStyle = pw.TextStyle(
      color: PdfColors.black, fontSize: 14, fontWeight: pw.FontWeight.bold);
  const cellStyle = pw.TextStyle(color: PdfColors.black, fontSize: 12);
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
      3: pw.Alignment.centerLeft,
      4: pw.Alignment.center,
      5: pw.Alignment.centerRight,
      6: pw.Alignment.centerRight,
      7: pw.Alignment.centerRight,
      8: pw.Alignment.centerRight,
    },
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
      return screenData.details![rowIndex].docref!;
    case 2:
      return screenData.details![rowIndex].barcode;
    case 3:
      return global.activeLangName(screenData.details![rowIndex].itemnames!);
    case 4:
      return global.formatNumber(screenData.details![rowIndex].qty);
    case 5:
      return global.activeLangName(screenData.details![rowIndex].unitnames!);
    case 6:
      return global.formatNumber(screenData.details![rowIndex].price);
    case 7:
      return screenData.details![rowIndex].discount;
    case 8:
      return global.formatNumber(screenData.details![rowIndex].sumamount);
    default:
      return '';
  }
}
