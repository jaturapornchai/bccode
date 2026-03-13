import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/payment_daily_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/global.dart' as global;

class TransactionPaymentDailyDialog extends StatefulWidget {
  final String docDate;
  final String jobId;
  final String Function(double) formatCurrency;

  const TransactionPaymentDailyDialog({
    super.key,
    required this.docDate,
    required this.jobId,
    required this.formatCurrency,
  });

  @override
  State<TransactionPaymentDailyDialog> createState() =>
      _TransactionPaymentDailyDialogState();
}

class _TransactionPaymentDailyDialogState
    extends State<TransactionPaymentDailyDialog>
    with global.ThemeRefreshMixin {
  @override
  void initState() {
    super.initState();
    _loadDetailData();
  }

  void _loadDetailData() {
    final token = global.appConfig.getString("token");
    if (token != null && token.isNotEmpty) {
      context.read<BiReportBloc>().add(
            GetBiReportDetailByDocDateRequested(
              reportType: BiReportType.paymentDaily,
              jobId: widget.jobId,
              docDate: widget.docDate,
              token: token,
            ),
          );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        width: MediaQuery.of(context).size.width > 800
            ? MediaQuery.of(context).size.width * 0.8
            : MediaQuery.of(context).size.width * 0.95,
        height: MediaQuery.of(context).size.height * 0.85,
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            BlocBuilder<BiReportBloc, BiReportState>(
              builder: (context, state) {
                PaymentDailyModel? paymentData;

                if (state is BiReportDetailByDocDateSuccess &&
                    state.docDate == widget.docDate &&
                    state.data.isNotEmpty &&
                    state.data.first is PaymentDailyModel) {
                  paymentData = state.data.first as PaymentDailyModel;
                }
                return _buildHeader(context, paymentData);
              },
            ),
            const SizedBox(height: 8),
            Expanded(
              child: BlocBuilder<BiReportBloc, BiReportState>(
                builder: (context, state) {
                  return _buildContent(state);
                },
              ),
            ),
            const SizedBox(height: 8),
            _buildCloseButton(context),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context, PaymentDailyModel? paymentData) {
    final docDate = paymentData?.docDate ?? widget.docDate;
    final transactionCount = paymentData?.transactions?.length ?? 0;

    return Column(
      children: [
        // แถวแรก: Title และปุ่มปิด
        Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: global.theme.positiveHighlightColor,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                Icons.payment_outlined,
                color: global.theme.positiveHighlightTextColor,
                size: 20,
              ),
            ),
            SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    global.language('report_payment_daily_title'),
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: global.theme.positiveHighlightTextColor,
                    ),
                  ),
                  Text(
                    ReportUtils.formatDate(docDate),
                    style: TextStyle(
                      fontSize: 12,
                      color: global.theme.textSecondaryColor,
                    ),
                  ),
                ],
              ),
            ),
            IconButton(
              onPressed: () => Navigator.of(context).pop(),
              icon: Icon(Icons.close, size: 20),
              tooltip: global.language('close'),
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(minHeight: 32, minWidth: 32),
            ),
          ],
        ),
        const SizedBox(height: 12),
        // แถวที่สอง: ข้อมูลสรุป
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: global.theme.backgroundColor,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: global.theme.dividerBorderColor),
          ),
          child: Row(
            children: [
              Expanded(
                child: Row(
                  children: [
                    Icon(Icons.today, color: global.theme.positiveHighlightTextColor, size: 16),
                    SizedBox(width: 6),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language('report_date'),
                          style: TextStyle(
                            fontSize: 10,
                            color: global.theme.textSecondaryColor,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        Text(
                          ReportUtils.formatDate(docDate),
                          style: TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              SizedBox(width: 16),
              Expanded(
                child: Row(
                  children: [
                    Icon(Icons.receipt_long_outlined,
                        color: global.theme.positiveHighlightTextColor, size: 16),
                    SizedBox(width: 6),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language('transaction_count'),
                          style: TextStyle(
                            fontSize: 10,
                            color: global.theme.textSecondaryColor,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        Text(
                          '$transactionCount ${global.language('transaction_items')}',
                          style: TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
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
  }

  Widget _buildContent(BiReportState state) {
    switch (state) {
      case BiReportDetailByDocDateLoading():
        if (state.docDate == widget.docDate) {
          return _buildLoadingState();
        }
        break;
      case BiReportDetailByDocDateSuccess():
        if (state.docDate == widget.docDate && state.data.isNotEmpty) {
          if (state.data.first is PaymentDailyModel) {
            final paymentData = state.data.first as PaymentDailyModel;
            return _buildDetailContent(paymentData);
          }
        }
        return _buildNoDataState();
      case BiReportDetailByDocDateFailure():
        if (state.docDate == widget.docDate) {
          return _buildErrorState(state.message);
        }
        break;
      default:
        break;
    }

    return _buildLoadingState();
  }

  Widget _buildLoadingState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const CircularProgressIndicator(),
          SizedBox(height: 16),
          Text(
            global.language('loading_details'),
            style: TextStyle(
              fontSize: 16,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorState(String errorMessage) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.error_outline,
            size: 64,
            color: global.theme.negativeHighlightTextColor,
          ),
          const SizedBox(height: 16),
          Text(
            errorMessage,
            style: TextStyle(
              fontSize: 16,
              color: global.theme.negativeHighlightTextColor,
            ),
            textAlign: TextAlign.center,
          ),
          SizedBox(height: 16),
          ElevatedButton.icon(
            onPressed: _loadDetailData,
            icon: Icon(Icons.refresh),
            label: Text(global.language('retry')),
          ),
        ],
      ),
    );
  }

  Widget _buildNoDataState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.inbox_outlined,
            size: 64,
            color: global.theme.textSecondaryColor,
          ),
          SizedBox(height: 16),
          Text(
            global.language('no_detail_data'),
            style: TextStyle(
              fontSize: 16,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailContent(PaymentDailyModel payment) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 1. ยอดสรุปการชำระเงิน
        _buildPaymentSummarySection(payment),
        const SizedBox(height: 8),
        // 2. รายการธุรกรรม
        _buildTransactionSection(payment),
      ],
    );
  }

  // 1. ยอดสรุปการชำระเงิน
  Widget _buildPaymentSummarySection(PaymentDailyModel payment) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // ยอดรวมหลักในกล่องเดียว
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [global.theme.positiveHighlightColor, global.theme.positiveHighlightColor],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: global.theme.positiveHighlightColor, width: 1),
          ),
          child: Row(
            children: [
              _buildMainSummaryItem(
                  global.language('total_value'),
                  widget.formatCurrency(payment.totalValue),
                  Icons.inventory_2_outlined),
              _buildVerticalDivider(),
              _buildMainSummaryItem(
                  global.language('net_value'),
                  widget.formatCurrency(payment.totalAmount),
                  Icons.calculate_outlined),
              _buildVerticalDivider(),
              _buildMainSummaryItem(
                  global.language('rounding'),
                  widget.formatCurrency(payment.roundAmount),
                  Icons.rounded_corner),
              _buildVerticalDivider(),
              _buildMainSummaryItem(global.language('payment_amount'),
                  widget.formatCurrency(payment.totalPayment), Icons.payment,
                  isHighlight: true),
            ],
          ),
        ),

        const SizedBox(height: 10),

        // วิธีการชำระเงิน - จัดเรียงใหม่แบบกระชับ
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: global.theme.backgroundColor,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: global.theme.dividerBorderColor),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('payment_details'),
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: global.theme.iconColor,
                ),
              ),
              const SizedBox(height: 8),

              // จัดเรียงวิธีการชำระแบบ grid 4x2
              _buildPaymentMethodsGrid(payment),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildMainSummaryItem(String label, String value, IconData icon,
      {bool isHighlight = false}) {
    return Expanded(
      child: Column(
        children: [
          // บรรทัดแรก: Icon + Label
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                icon,
                size: isHighlight ? 18 : 16,
                color:
                    isHighlight ? global.theme.positiveHighlightTextColor : global.theme.positiveHighlightTextColor,
              ),
              const SizedBox(width: 4),
              Flexible(
                child: Text(
                  label,
                  style: TextStyle(
                    fontSize: isHighlight ? 12 : 11,
                    color: isHighlight
                        ? global.theme.positiveHighlightTextColor
                        : global.theme.textSecondaryColor,
                    fontWeight: FontWeight.w600,
                  ),
                  textAlign: TextAlign.center,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          // บรรทัดที่สอง: Value
          Text(
            value,
            style: TextStyle(
              fontSize: isHighlight ? 14 : 12,
              fontWeight: FontWeight.bold,
              color: isHighlight ? global.theme.positiveHighlightTextColor : global.theme.textColor,
            ),
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }

  Widget _buildVerticalDivider() {
    return Container(
      height: 40,
      width: 1,
      color: global.theme.positiveHighlightColor,
    );
  }

  Widget _buildPaymentMethodsGrid(PaymentDailyModel payment) {
    final paymentMethods = [
      {'label': global.language('payment_cash'), 'amount': payment.payCashAmount, 'icon': Icons.money},
      {
        'label': global.language('payment_transfer'),
        'amount': payment.sumTransfer,
        'icon': Icons.account_balance
      },
      {
        'label': global.language('payment_credit_card'),
        'amount': payment.sumCreditCard,
        'icon': Icons.credit_card
      },
      {'label': global.language('payment_cheque'), 'amount': payment.sumCheque, 'icon': Icons.receipt},
      {
        'label': global.language('payment_coupon'),
        'amount': payment.sumCoupon,
        'icon': Icons.local_offer
      },
      {'label': global.language('payment_qr_code'), 'amount': payment.sumQRCode, 'icon': Icons.qr_code},
      {
        'label': global.language('payment_credit'),
        'amount': payment.sumCredit,
        'icon': Icons.account_balance_wallet
      },
    ];

    // แสดงเฉพาะรายการที่มียอดเงิน
    final activePayments = paymentMethods
        .where((method) => (method['amount'] as double) > 0)
        .toList();

    if (activePayments.isEmpty) {
      return Container(
        padding: EdgeInsets.all(8),
        child: Text(
          global.language('no_payment_items'),
          style: TextStyle(
            fontSize: 12,
            color: global.theme.textSecondaryColor,
            fontStyle: FontStyle.italic,
          ),
          textAlign: TextAlign.center,
        ),
      );
    }

    return Wrap(
      spacing: 6,
      runSpacing: 4,
      children: activePayments
          .map((method) => _buildCompactPaymentChip(
                method['label'] as String,
                widget.formatCurrency(method['amount'] as double),
                method['icon'] as IconData,
              ))
          .toList(),
    );
  }

  Widget _buildCompactPaymentChip(String method, String amount, IconData icon) {
    // กำหนดสีตามประเภทการชำระเงิน
    Color backgroundColor;
    Color borderColor;
    Color iconColor;
    Color textColor;

    // เปรียบเทียบด้วย translation key แทนการใช้ hardcoded string
    final cashLabel = global.language('payment_cash');
    final transferLabel = global.language('payment_transfer');
    final creditCardLabel = global.language('payment_credit_card');
    final chequeLabel = global.language('payment_cheque');
    final couponLabel = global.language('payment_coupon');
    final qrCodeLabel = global.language('payment_qr_code');
    final creditLabel = global.language('payment_credit');

    if (method == cashLabel) {
      backgroundColor = global.theme.positiveHighlightColor;
      borderColor = global.theme.positiveHighlightColor;
      iconColor = global.theme.positiveHighlightTextColor;
      textColor = global.theme.positiveHighlightTextColor;
    } else if (method == transferLabel) {
      backgroundColor = global.theme.infoHighlightColor;
      borderColor = global.theme.infoHighlightColor;
      iconColor = global.theme.infoHighlightTextColor;
      textColor = global.theme.infoHighlightTextColor;
    } else if (method == creditCardLabel) {
      backgroundColor = Colors.purple.shade50;
      borderColor = Colors.purple.shade200;
      iconColor = Colors.purple.shade600;
      textColor = Colors.purple.shade700;
    } else if (method == chequeLabel) {
      backgroundColor = global.theme.warningHighlightColor;
      borderColor = global.theme.warningHighlightColor;
      iconColor = global.theme.warningHighlightTextColor;
      textColor = global.theme.warningHighlightTextColor;
    } else if (method == couponLabel) {
      backgroundColor = Colors.pink.shade50;
      borderColor = Colors.pink.shade200;
      iconColor = Colors.pink.shade600;
      textColor = Colors.pink.shade700;
    } else if (method == qrCodeLabel) {
      backgroundColor = Colors.indigo.shade50;
      borderColor = Colors.indigo.shade200;
      iconColor = Colors.indigo.shade600;
      textColor = Colors.indigo.shade700;
    } else if (method == creditLabel) {
      backgroundColor = Colors.amber.shade50;
      borderColor = Colors.amber.shade200;
      iconColor = Colors.amber.shade600;
      textColor = Colors.amber.shade700;
    } else {
      backgroundColor = global.theme.backgroundColor;
      borderColor = global.theme.dividerBorderColor;
      iconColor = global.theme.textSecondaryColor;
      textColor = global.theme.iconColor;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderColor, width: 0.5),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            size: 14,
            color: iconColor,
          ),
          const SizedBox(width: 6),
          Text(
            method,
            style: TextStyle(
              fontSize: 14,
              color: textColor,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(width: 6),
          Text(
            amount,
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: textColor,
            ),
          ),
        ],
      ),
    );
  }

  // 2. รายการธุรกรรม
  Widget _buildTransactionSection(PaymentDailyModel payment) {
    return Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: global.theme.cardColor,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: global.theme.dividerBorderColor),
              ),
              child: Column(
                children: [
                  // Header row
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 12),
                    decoration: BoxDecoration(
                      color: global.theme.backgroundColor,
                      borderRadius: const BorderRadius.only(
                        topLeft: Radius.circular(12),
                        topRight: Radius.circular(12),
                      ),
                    ),
                    child: Row(
                      children: [
                        SizedBox(width: 30), // space for item number
                        Expanded(
                          flex: 3,
                          child: Text(
                            global.language('document_number'),
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: global.theme.positiveHighlightTextColor,
                            ),
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.language('customer'),
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: global.theme.positiveHighlightTextColor,
                            ),
                            textAlign: TextAlign.center,
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.language('payment_cash'),
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: global.theme.positiveHighlightTextColor,
                            ),
                            textAlign: TextAlign.center,
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.language('total'),
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: global.theme.positiveHighlightTextColor,
                            ),
                            textAlign: TextAlign.right,
                          ),
                        ),
                      ],
                    ),
                  ),
                  // List items
                  Expanded(
                    child: payment.transactions?.isNotEmpty == true
                        ? ListView.builder(
                            itemCount: payment.transactions!.length,
                            itemBuilder: (context, index) {
                              final transaction = payment.transactions![index];
                              return _buildTransactionItem(
                                  transaction, index, payment);
                            },
                          )
                        : Center(
                            child: Padding(
                              padding: EdgeInsets.all(16.0),
                              child: Text(
                                global.language('no_transactions'),
                                style: TextStyle(
                                  fontSize: 14,
                                  color: global.theme.textSecondaryColor,
                                  fontStyle: FontStyle.italic,
                                ),
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
    );
  }

  Widget _buildTransactionItem(
      TransactionModel transaction, int index, PaymentDailyModel payment) {
    final totalTransactions = payment.transactions?.length ?? 0;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        border: index < totalTransactions - 1
            ? Border(bottom: BorderSide(color: global.theme.surfaceColor))
            : null,
      ),
      child: Row(
        children: [
          // Item number
          Container(
            width: 24,
            height: 24,
            decoration: BoxDecoration(
              color: global.theme.positiveHighlightColor,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Center(
              child: Text(
                '${index + 1}',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: global.theme.positiveHighlightTextColor,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          // Document number and date/time
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  transaction.docNo,
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  '${ReportUtils.formatDate(transaction.docDate)} • ${transaction.docTime}',
                  style: TextStyle(
                    fontSize: 11,
                    color: global.theme.textSecondaryColor,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          // Customer name
          Expanded(
            child: Text(
              transaction.custName.isEmpty
                  ? global.language('customer_general')
                  : transaction.custName,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          // Cash amount
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.payCashAmount),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: global.theme.positiveHighlightTextColor,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          // Total payment
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.totalPayment),
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: global.theme.positiveHighlightTextColor,
              ),
              textAlign: TextAlign.right,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCloseButton(BuildContext context) {
    return Center(
      child: ElevatedButton(
        onPressed: () => Navigator.of(context).pop(),
        style: ElevatedButton.styleFrom(
          backgroundColor: global.theme.positiveHighlightTextColor,
          foregroundColor: global.theme.onPrimaryColor,
          padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
          minimumSize: const Size(100, 40),
        ),
        child: Text(
          global.language('close'),
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ),
    );
  }
}
