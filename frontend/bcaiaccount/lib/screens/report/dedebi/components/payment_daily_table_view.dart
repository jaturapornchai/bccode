import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../../../model/bi_report/bi_report_models.dart';
import '../../../../model/bi_report/payment_daily_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class PaymentDailyTableView extends StatelessWidget {
  final List<PaymentDailyModel> data;
  final BiReportMeta? meta;
  final VoidCallback? onLoadMore;
  final Function(PaymentDailyModel)? onRowTap;

  const PaymentDailyTableView({
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
              Icons.payment_outlined,
              size: 64,
              color: global.theme.textSecondaryColor,
            ),
            SizedBox(height: 16),
            Text(
              global.language('payment_daily_no_data'),
              style: TextStyle(
                fontSize: 16,
                color: global.theme.textSecondaryColor,
              ),
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
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_date_column')),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_total_amount_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_round_amount_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_after_round_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_cash_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_transfer_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_credit_card_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_cheque_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_coupon_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_qr_code_column'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('payment_daily_net_payment_column'), TextAlign.right),
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

  Widget _buildDataRow(PaymentDailyModel item, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? global.theme.cardColor : global.theme.backgroundColor;

    return InkWell(
      onTap: () {
        try {
          onRowTap?.call(item);
        } catch (e) {
          AppLogger.error('Error in payment daily row tap: $e');
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
              flex: 2,
              child: _buildDataCell(
                _formatDate(item.docDate),
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: item.transactions!.isNotEmpty
                    ? Colors.indigo.shade700
                    : global.theme.textColor,
              ),
            ),
            // มูลค่ารวม
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.totalAmount),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.w600,
                color: Colors.indigo.shade700,
              ),
            ),
            // มูลค่าปัดเศษ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.roundAmount),
                textAlign: TextAlign.right,
                color: item.roundAmount != 0
                    ? global.theme.warningHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight:
                    item.roundAmount != 0 ? FontWeight.w500 : FontWeight.normal,
              ),
            ),
            // มูลค่าหลังปัดเศษ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.totalValue),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.w600,
              ),
            ),
            // เงินสด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.payCashAmount),
                textAlign: TextAlign.right,
                color: item.payCashAmount > 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight: item.payCashAmount > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // เงินโอน
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.sumTransfer),
                textAlign: TextAlign.right,
                color: item.sumTransfer > 0
                    ? global.theme.infoHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight:
                    item.sumTransfer > 0 ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
            // บัตรเครดิต
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.sumCreditCard),
                textAlign: TextAlign.right,
                color: item.sumCreditCard > 0
                    ? Colors.purple.shade700
                    : global.theme.textSecondaryColor,
                fontWeight: item.sumCreditCard > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // เช็ค
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.sumCheque),
                textAlign: TextAlign.right,
                color: item.sumCheque > 0
                    ? global.theme.warningHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight:
                    item.sumCheque > 0 ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
            // คูปอง
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.sumCoupon),
                textAlign: TextAlign.right,
                color: item.sumCoupon > 0
                    ? global.theme.negativeHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight:
                    item.sumCoupon > 0 ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
            // QR Code
            Expanded(
              flex: 2,
              child: _buildDataCell(
                _formatCurrency(item.sumQRCode),
                textAlign: TextAlign.right,
                color: item.sumQRCode > 0
                    ? Colors.teal.shade700
                    : global.theme.textSecondaryColor,
                fontWeight:
                    item.sumQRCode > 0 ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
            // ชำระสุทธิ
            Expanded(
              flex: 2,
              child: Text(
                _formatCurrency(item.totalPayment),
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

  String _formatCurrency(double amount) {
    if (amount == 0) return '-';
    return NumberFormat('#,##0.00').format(amount);
  }
}
