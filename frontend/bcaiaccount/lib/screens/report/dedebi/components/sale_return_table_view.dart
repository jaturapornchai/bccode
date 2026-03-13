import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../../../model/bi_report/bi_report_models.dart';
import '../../../../model/bi_report/sale_return_model.dart';
import '../utils/report_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class SaleReturnTableView extends StatelessWidget {
  final List<SaleReturnModel> data;
  final BiReportMeta? meta;
  final VoidCallback? onLoadMore;
  final Function(SaleReturnModel)? onRowTap;

  const SaleReturnTableView({
    super.key,
    required this.data,
    this.meta,
    this.onLoadMore,
    this.onRowTap,
  });

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.assignment_return_outlined,
              size: 64,
              color: global.theme.textSecondaryColor,
            ),
            SizedBox(height: 16),
            Text(
              global.language('sale_return.no_data'),
              style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
            ),
          ],
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

        // Load more button
        if (meta != null && meta!.page < meta!.totalPage)
          Container(
            padding: EdgeInsets.all(16),
            child: ElevatedButton(
              onPressed: onLoadMore,
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.indigo,
                foregroundColor: global.theme.onPrimaryColor,
                minimumSize: const Size(double.infinity, 48),
              ),
              child: Text(global.language('sale_return.load_more')),
            ),
          ),
      ],
    );
  }

  Widget _buildHeaderRow() {
    return Row(
      children: [
        Expanded(flex: 1, child: _buildHeaderCell(global.language('sale_return.date'))),
        Expanded(flex: 1, child: _buildHeaderCell(global.language('sale_return.doc_number'))),
        Expanded(flex: 1, child: _buildHeaderCell(global.language('sale_return.customer'))),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('sale_return.product_value'), TextAlign.right),
        ),
        Expanded(flex: 1, child: _buildHeaderCell(global.language('sale_return.discount'), TextAlign.right)),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('sale_return.after_discount'), TextAlign.right),
        ),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('sale_return.tax_exempt'), TextAlign.right),
        ),
        Expanded(flex: 1, child: _buildHeaderCell(global.language('sale_return.before_tax'), TextAlign.right)),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('sale_return.tax_value'), TextAlign.right),
        ),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('sale_return.net_value'), TextAlign.right),
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

  Widget _buildDataRow(SaleReturnModel saleReturn, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? global.theme.cardColor : global.theme.backgroundColor;

    return InkWell(
      onTap: () {
        try {
          onRowTap?.call(saleReturn);
        } catch (e) {
          AppLogger.error('Error in sale return row tap: $e');
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
              child: _buildDataCell(
                _formatDate(saleReturn.docDate),
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: Colors.indigo.shade700,
              ),
            ),
            // เลขที่เอกสาร
            Expanded(
              flex: 1,
              child: _buildDataCell(
                saleReturn.docno,
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: global.theme.infoHighlightTextColor,
              ),
            ),
            // ลูกค้า
            Expanded(
              flex: 1,
              child: _buildDataCell(
                _getCreditorName(saleReturn.creditorNames),
                fontSize: 11,
                color: global.theme.positiveHighlightTextColor,
              ),
            ),

            // มูลค่าสินค้า
            Expanded(
              flex: 1,
              child: _buildDataCell(
                ReportUtils.formatCurrency(saleReturn.totalValue),
                textAlign: TextAlign.right,
                color: Colors.purple.shade700,
                fontWeight: FontWeight.w600,
              ),
            ),
            // ส่วนลด
            Expanded(
              flex: 1,
              child: _buildDataCell(
                saleReturn.detailTotalDiscount > 0
                    ? ReportUtils.formatCurrency(saleReturn.detailTotalDiscount)
                    : '-',
                textAlign: TextAlign.right,
                color: saleReturn.detailTotalDiscount > 0
                    ? global.theme.negativeHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight: saleReturn.detailTotalDiscount > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // หลังหักส่วนลด
            Expanded(
              flex: 1,
              child: _buildDataCell(
                ReportUtils.formatCurrency(saleReturn.totalAfterDiscount),
                textAlign: TextAlign.right,
                color: Colors.teal.shade700,
                fontWeight: FontWeight.w600,
              ),
            ),
            // มูลค่ายกเว้นภาษี
            Expanded(
              flex: 1,
              child: _buildDataCell(
                ReportUtils.formatCurrency(saleReturn.totalExceptVat),
                textAlign: TextAlign.right,
                color: Colors.cyan.shade700,
                fontWeight: FontWeight.w500,
              ),
            ),
            // ก่อนภาษี
            Expanded(
              flex: 1,
              child: _buildDataCell(
                ReportUtils.formatCurrency(saleReturn.totalBeforeVat),
                textAlign: TextAlign.right,
                color: global.theme.warningHighlightTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
            // มูลค่าภาษี
            Expanded(
              flex: 1,
              child: _buildDataCell(
                saleReturn.totalVatValue > 0
                    ? ReportUtils.formatCurrency(saleReturn.totalVatValue)
                    : '-',
                textAlign: TextAlign.right,
                color: saleReturn.totalVatValue > 0
                    ? Colors.brown.shade700
                    : global.theme.textSecondaryColor,
                fontWeight: saleReturn.totalVatValue > 0
                    ? FontWeight.w500
                    : FontWeight.normal,
              ),
            ),
            // มูลค่าสุทธิ
            Expanded(
              flex: 1,
              child: Text(
                ReportUtils.formatCurrency(saleReturn.totalAmount),
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: Colors.indigo.shade700,
                ),
                textAlign: TextAlign.right,
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
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: fontSize ?? 11,
          fontWeight: fontWeight,
          color: color,
        ),
        textAlign: textAlign,
      ),
    );
  }

  String _formatDate(String dateString) {
    try {
      final date = DateTime.parse(dateString);
      return DateFormat('dd/MM/yyyy').format(date);
    } catch (e) {
      return dateString;
    }
  }

  String _getCreditorName(List<NameModel> creditorNames) {
    if (creditorNames.isEmpty) return '-';

    // ให้ความสำคัญกับภาษาไทย
    final thName = creditorNames.where((name) => name.code == 'th').firstOrNull;
    if (thName != null && thName.name.isNotEmpty) {
      return thName.name;
    }

    // ถ้าไม่มีภาษาไทย ใช้ตัวแรก
    return creditorNames.first.getDisplayName();
  }
}
