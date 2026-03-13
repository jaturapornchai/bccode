import 'package:flutter/material.dart';
import 'package:smlaicloud/model/bi_report/bi_sale_report_data.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class SaleReportTableView extends StatelessWidget {
  final List<SaleReportData> data;
  final String Function(List<SaleCreditorName>) getCreditorName;
  final Function(SaleReportData)? onRowTap;

  const SaleReportTableView({
    super.key,
    required this.data,
    required this.getCreditorName,
    this.onRowTap,
  });

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) {
      return Center(
        child: Text(
          global.language('sale_report_no_data'),
          style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
        ),
      );
    }

    return Column(
      children: [
        // Header Row
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            border: Border(
              bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1),
            ),
          ),
          child: _buildHeaderRow(),
        ),
        // Data Rows
        Expanded(
          child: ListView.builder(
            itemCount: data.length,
            itemBuilder: (context, index) => _buildDataRow(data[index], index),
          ),
        ),
      ],
    );
  }

  Widget _buildHeaderRow() {
    return Row(
      children: [
        Expanded(flex: 1, child: _buildHeaderCell(global.language('sale_report_date'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_report_doc_number'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_report_customer'))),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('sale_report_doc_type'), TextAlign.center),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_report_product_value'), TextAlign.right),
        ),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_report_discount'), TextAlign.right)),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_report_after_discount'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_report_tax_exempt'), TextAlign.right),
        ),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_report_before_tax'), TextAlign.right)),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_report_tax_value'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_report_net_value'), TextAlign.right),
        ),
      ],
    );
  }

  Widget _buildHeaderCell(String text, [TextAlign? textAlign]) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.bold,
          color: Colors.indigo.shade700,
        ),
        textAlign: textAlign ?? TextAlign.left,
      ),
    );
  }

  Widget _buildDataRow(SaleReportData sale, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? global.theme.cardColor : global.theme.backgroundColor;

    return InkWell(
      onTap: () {
        try {
          onRowTap?.call(sale);
        } catch (e) {
          AppLogger.error('Error in row tap: $e');
        }
      },
      hoverColor: global.theme.rowHoverColor,
      highlightColor: Colors.indigo.shade100,
      splashColor: Colors.indigo.shade200,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: backgroundColor,
          border: Border(
            bottom: BorderSide(color: global.theme.dividerBorderColor, width: 0.5),
          ),
        ),
        child: Row(
          children: [
            // วันที่
            Expanded(
              flex: 1,
              child: _buildDataCell(ReportUtils.formatDate(sale.docdate)),
            ),
            // เลขที่เอกสาร
            Expanded(
              flex: 2,
              child: _buildDataCell(
                sale.docno,
                fontWeight: FontWeight.w600,
                color: sale.transactions.isNotEmpty
                    ? Colors.indigo.shade700
                    : global.theme.textColor,
                decoration: sale.transactions.isNotEmpty
                    ? TextDecoration.underline
                    : null,
              ),
            ),
            // ลูกค้า
            Expanded(
              flex: 2,
              child: _buildDataCell(() {
                try {
                  final creditorName = getCreditorName(sale.creditornames);
                  return creditorName.isEmpty ? global.language('sale_report_general_customer') : creditorName;
                } catch (e) {
                  return global.language('sale_report_general_customer');
                }
              }(), overflow: TextOverflow.ellipsis),
            ),
            // ประเภทเอกสาร
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 6,
                    vertical: 2,
                  ),
                  decoration: BoxDecoration(
                    color: ReportUtils.getStatusColor(sale.inquirytype),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    ReportUtils.getStatusText(sale.inquirytype),
                    style: TextStyle(
                      fontSize: 9,
                      fontWeight: FontWeight.w600,
                      color: ReportUtils.getStatusTextColor(sale.inquirytype),
                    ),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),
            ),
            // มูลค่าสินค้า
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.totalvalue),
                textAlign: TextAlign.right,
              ),
            ),
            // ส่วนลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.detailtotaldiscount),
                textAlign: TextAlign.right,
                color: global.theme.negativeHighlightTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
            // มูลค่าหลังหักส่วนลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.totalafterdiscount),
                textAlign: TextAlign.right,
              ),
            ),
            // มูลค่ายกเว้นภาษี
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.totalexceptvat),
                textAlign: TextAlign.right,
              ),
            ),
            // มูลค่าก่อนภาษี
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.totalbeforevat),
                textAlign: TextAlign.right,
              ),
            ),
            // มูลค่าภาษี
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.totalvatvalue),
                textAlign: TextAlign.right,
                color: global.theme.warningHighlightTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
            // มูลค่าสุทธิ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(sale.totalamount),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.bold,
                color: Colors.indigo.shade700,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDataCell(
    String text, {
    TextAlign? textAlign,
    double? fontSize,
    FontWeight? fontWeight,
    Color? color,
    TextDecoration? decoration,
    TextOverflow? overflow,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: fontSize ?? 11,
          fontWeight: fontWeight,
          color: color,
          decoration: decoration,
        ),
        textAlign: textAlign,
        overflow: overflow,
      ),
    );
  }
}
