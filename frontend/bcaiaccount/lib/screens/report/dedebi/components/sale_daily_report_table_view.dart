import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/sale_daily_report_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class SaleDailyReportTableView extends StatelessWidget {
  final List<SaleDailyReportData> data;
  final Function(SaleDailyReportData)? onRowTap;

  const SaleDailyReportTableView({
    super.key,
    required this.data,
    this.onRowTap,
  });

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) {
      return Center(
        child: Text(
          global.language('sale_daily_report_no_data'),
          style: const TextStyle(fontSize: 16, color: Colors.grey),
        ),
      );
    }

    return Column(
      children: [
        // Header Row
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: Colors.grey.shade100,
            border: Border(
              bottom: BorderSide(color: Colors.grey.shade300, width: 1),
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
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_daily_report_date'))),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_daily_report_product_value'), TextAlign.right),
        ),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_daily_report_discount'), TextAlign.right)),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_daily_report_after_discount'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_daily_report_except_vat'), TextAlign.right),
        ),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sale_daily_report_before_vat'), TextAlign.right)),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_daily_report_vat_value'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_daily_report_net_value'), TextAlign.right),
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

  Widget _buildDataRow(SaleDailyReportData dailySale, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? Colors.white : Colors.grey.shade50;

    return InkWell(
      onTap: () {
        try {
          onRowTap?.call(dailySale);
        } catch (e) {
          AppLogger.error('Error in row tap: $e');
        }
      },
      hoverColor: Colors.indigo.shade50,
      highlightColor: Colors.indigo.shade100,
      splashColor: Colors.indigo.shade200,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: backgroundColor,
          border: Border(
            bottom: BorderSide(color: Colors.grey.shade200, width: 0.5),
          ),
        ),
        child: Row(
          children: [
            // วันที่
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatDate(dailySale.docDate),
                fontWeight: FontWeight.w600,
                color: dailySale.transactions!.isNotEmpty
                    ? Colors.indigo.shade700
                    : Colors.black87,
                decoration: dailySale.transactions!.isNotEmpty
                    ? TextDecoration.underline
                    : null,
              ),
            ),
            // มูลค่าสินค้า
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.totalValue),
                textAlign: TextAlign.right,
              ),
            ),
            // ส่วนลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.detailTotalDiscount),
                textAlign: TextAlign.right,
                color: Colors.red.shade700,
                fontWeight: FontWeight.w500,
              ),
            ),
            // มูลค่าหลังหักส่วนลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.totalAfterDiscount),
                textAlign: TextAlign.right,
              ),
            ),
            // มูลค่ายกเว้นภาษี
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.totalExceptVat),
                textAlign: TextAlign.right,
              ),
            ),
            // มูลค่าก่อนภาษี
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.totalBeforeVat),
                textAlign: TextAlign.right,
              ),
            ),
            // มูลค่าภาษี
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.totalVatValue),
                textAlign: TextAlign.right,
                color: Colors.orange.shade700,
                fontWeight: FontWeight.w500,
              ),
            ),
            // มูลค่าสุทธิ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(dailySale.totalAmount),
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
