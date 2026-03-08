import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/global.dart' as global;

class TransactionDailyDetailDialog extends StatefulWidget {
  final String docDate;
  final String jobId;
  final String Function(double) formatCurrency;

  const TransactionDailyDetailDialog({
    super.key,
    required this.docDate,
    required this.jobId,
    required this.formatCurrency,
  });

  @override
  State<TransactionDailyDetailDialog> createState() =>
      _TransactionDailyDetailDialogState();
}

class _TransactionDailyDetailDialogState
    extends State<TransactionDailyDetailDialog> {
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
              reportType: BiReportType.saleDaily,
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
                SaleDailyReportData? saleData;

                if (state is BiReportDetailByDocDateSuccess &&
                    state.docDate == widget.docDate &&
                    state.data.isNotEmpty &&
                    state.data.first is SaleDailyReportData) {
                  saleData = state.data.first as SaleDailyReportData;
                }
                return _buildHeader(context, saleData);
              },
            ),
            const SizedBox(height: 16),
            Expanded(
              child: BlocBuilder<BiReportBloc, BiReportState>(
                builder: (context, state) {
                  return _buildContent(state);
                },
              ),
            ),
            const SizedBox(height: 16),
            _buildCloseButton(context),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context, SaleDailyReportData? saleData) {
    final docDate = saleData?.docDate ?? widget.docDate;

    return Container(
      padding: EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.indigo.shade100,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(
              Icons.calendar_today_outlined,
              color: Colors.indigo.shade700,
              size: 20,
            ),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language('daily_sales_report'),
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.indigo,
                  ),
                ),
                Text(
                  ReportUtils.formatDate(docDate),
                  style: const TextStyle(
                    fontSize: 12,
                    color: Colors.grey,
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
          if (state.data.first is SaleDailyReportData) {
            final saleData = state.data.first as SaleDailyReportData;
            return _buildDetailContent(saleData);
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
            style: const TextStyle(
              fontSize: 16,
              color: Colors.grey,
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
            color: Colors.red.shade300,
          ),
          const SizedBox(height: 16),
          Text(
            errorMessage,
            style: const TextStyle(
              fontSize: 16,
              color: Colors.red,
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
          const Icon(
            Icons.inbox_outlined,
            size: 64,
            color: Colors.grey,
          ),
          SizedBox(height: 16),
          Text(
            global.language('no_detail_data'),
            style: const TextStyle(
              fontSize: 16,
              color: Colors.grey,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailContent(SaleDailyReportData sale) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 1. ข้อมูลวันที่
        _buildDateInfoSection(sale),
        const SizedBox(height: 16),
        // 2. ยอดสรุป
        _buildSummarySection(sale),
        const SizedBox(height: 16),
        // 3. รายการธุรกรรม
        _buildTransactionSection(sale),
      ],
    );
  }

  // 1. ข้อมูลวันที่
  Widget _buildDateInfoSection(SaleDailyReportData sale) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.today, color: Colors.indigo.shade600, size: 18),
            SizedBox(width: 8),
            Text(
              '1. ${global.language('date_information')}',
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
            Spacer(),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.green.shade100,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '${sale.transactions?.length ?? 0} ${global.language('items')}',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  color: Colors.green.shade800,
                ),
              ),
            ),
          ],
        ),
        SizedBox(height: 12),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.grey.shade50,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.grey.shade200),
          ),
          child: Column(
            children: [
              Row(
                children: [
                  Expanded(
                    flex: 2,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language('report_date'),
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          ReportUtils.formatDate(sale.docDate),
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                  SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language('transaction_count'),
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        SizedBox(height: 4),
                        Text(
                          '${sale.transactions?.length ?? 0} ${global.language('items')}',
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  // 2. ยอดสรุป
  Widget _buildSummarySection(SaleDailyReportData sale) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.calculate_outlined,
                color: Colors.indigo.shade600, size: 16),
            SizedBox(width: 6),
            Text(
              '2. ${global.language('summary')}',
              style: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
          ],
        ),
        SizedBox(height: 8),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.indigo.shade50,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: Colors.indigo.shade200),
          ),
          child: Row(
            children: [
              Expanded(
                child: _buildCompactSummaryItem(global.language('product_value'),
                    ReportUtils.formatCurrencySafe(sale.totalValue)),
              ),
              SizedBox(width: 12),
              Expanded(
                child: _buildCompactSummaryItem(global.language('before_vat'),
                    ReportUtils.formatCurrencySafe(sale.totalBeforeVat)),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildCompactSummaryItem(
                    'VAT', ReportUtils.formatCurrencySafe(sale.totalVatValue),
                    isVat: true),
              ),
              SizedBox(width: 12),
              Expanded(
                flex: 2,
                child: _buildCompactSummaryItem(global.language('grand_total'),
                    ReportUtils.formatCurrencySafe(sale.totalAmount),
                    isTotal: true),
              ),
            ],
          ),
        ),
      ],
    );
  }

  // 3. รายการธุรกรรม
  Widget _buildTransactionSection(SaleDailyReportData sale) {
    final transactions = sale.transactions ?? [];

    return Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.receipt_long_outlined,
                  color: Colors.indigo.shade600, size: 18),
              SizedBox(width: 8),
              Text(
                '3. ${global.language('transactions')} (${transactions.length} ${global.language('items')})',
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Colors.indigo,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.grey.shade200),
              ),
              child: Column(
                children: [
                  // Header row
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 12),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade50,
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
                            global.language('document_no'),
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: Colors.indigo,
                            ),
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.language('before_vat'),
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: Colors.indigo.shade700,
                            ),
                            textAlign: TextAlign.center,
                          ),
                        ),
                        Expanded(
                          child: Text(
                            'VAT',
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: Colors.indigo.shade700,
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
                              color: Colors.indigo.shade700,
                            ),
                            textAlign: TextAlign.right,
                          ),
                        ),
                      ],
                    ),
                  ),
                  // List items
                  Expanded(
                    child: transactions.isEmpty
                        ? Container(
                            padding: EdgeInsets.all(16),
                            child: Center(
                              child: Text(
                                global.language('no_transactions'),
                                style: const TextStyle(
                                  color: Colors.grey,
                                  fontSize: 14,
                                ),
                              ),
                            ),
                          )
                        : ListView.builder(
                            itemCount: transactions.length,
                            itemBuilder: (context, index) {
                              final transaction = transactions[index];
                              return _buildTransactionItem(
                                  transaction, index, transactions.length);
                            },
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

  Widget _buildCompactSummaryItem(String label, String value,
      {bool isVat = false, bool isTotal = false}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 10,
            color: isTotal ? Colors.indigo.shade700 : Colors.grey.shade600,
            fontWeight: FontWeight.w500,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          value,
          style: TextStyle(
            fontSize: isTotal ? 14 : 12,
            fontWeight: isTotal ? FontWeight.bold : FontWeight.w600,
            color: isTotal
                ? Colors.indigo.shade700
                : isVat
                    ? Colors.orange.shade700
                    : Colors.black87,
          ),
        ),
      ],
    );
  }

  Widget _buildTransactionItem(
      SaleDailyTransaction transaction, int index, int totalCount) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        border: index < totalCount - 1
            ? Border(bottom: BorderSide(color: Colors.grey.shade100))
            : null,
      ),
      child: Row(
        children: [
          // Item number
          Container(
            width: 24,
            height: 24,
            decoration: BoxDecoration(
              color: Colors.blue.shade50,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Center(
              child: Text(
                '${index + 1}',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: Colors.blue.shade700,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          // Document number and date
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  transaction.docno,
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  ReportUtils.formatDate(transaction.docDate),
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.grey.shade600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          // Before VAT amount
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.totalBeforeVat),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          // VAT amount
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.totalVatValue),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: Colors.orange.shade700,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          // Total amount
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.totalAmount),
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Colors.indigo.shade700,
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
          backgroundColor: Colors.indigo.shade600,
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
          minimumSize: const Size(100, 40),
        ),
        child: Text(
          global.language('close'),
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ),
    );
  }
}
