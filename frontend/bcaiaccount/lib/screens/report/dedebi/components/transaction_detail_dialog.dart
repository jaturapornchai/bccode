import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/bi_sale_report_data.dart';
import 'package:smlaicloud/model/bi_report/sale_return_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/global.dart' as global;

class TransactionDetailDialog extends StatefulWidget {
  final String docno;
  final String jobId;
  final BiReportType reportType;
  final String Function(double) formatCurrency;

  const TransactionDetailDialog({
    super.key,
    required this.docno,
    required this.jobId,
    required this.reportType,
    required this.formatCurrency,
  });

  @override
  State<TransactionDetailDialog> createState() =>
      _TransactionDetailDialogState();
}

class _TransactionDetailDialogState extends State<TransactionDetailDialog> {
  @override
  void initState() {
    super.initState();
    _loadDetailData();
  }

  void _loadDetailData() {
    final token = global.appConfig.getString("token");
    if (token != null && token.isNotEmpty) {
      context.read<BiReportBloc>().add(
            GetBiReportDetailByDocNoRequested(
              reportType: widget.reportType,
              jobId: widget.jobId,
              docNo: widget.docno,
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
                SaleReportData? saleData;
                SaleReturnModel? saleReturnData;

                if (state is BiReportDetailByDocNoSuccess &&
                    state.docNo == widget.docno &&
                    state.data.isNotEmpty) {
                  if (widget.reportType == BiReportType.sale &&
                      state.data.first is SaleReportData) {
                    saleData = state.data.first as SaleReportData;
                  } else if (widget.reportType == BiReportType.saleReturn &&
                      state.data.first is SaleReturnModel) {
                    saleReturnData = state.data.first as SaleReturnModel;
                  }
                }
                return _buildHeader(context,
                    saleData: saleData, saleReturnData: saleReturnData);
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

  Widget _buildHeader(BuildContext context,
      {SaleReportData? saleData, SaleReturnModel? saleReturnData}) {
    final docno = widget.docno;
    final String docdate;
    final String docTime;
    final String branchcode;

    if (saleData != null) {
      docdate = saleData.docdate;
      docTime = saleData.docTime;
      branchcode = saleData.branchcode;
    } else if (saleReturnData != null) {
      docdate = saleReturnData.docDate;
      docTime = ''; // SaleReturnModel doesn't have docTime
      branchcode = ''; // SaleReturnModel doesn't have branchcode
    } else {
      docdate = '';
      docTime = '';
      branchcode = '';
    }

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
              Icons.receipt_outlined,
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
                  '${global.language('transaction_detail')} $docno',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.indigo,
                  ),
                ),
                Text(
                  '${ReportUtils.formatDate(docdate)}${docTime.isNotEmpty ? ' • $docTime' : ''}${branchcode.isNotEmpty ? ' • ${global.language('transaction_branch')} $branchcode' : ''}',
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
            tooltip: global.language('transaction_close'),
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minHeight: 32, minWidth: 32),
          ),
        ],
      ),
    );
  }

  Widget _buildContent(BiReportState state) {
    switch (state) {
      case BiReportDetailByDocNoLoading():
        if (state.docNo == widget.docno) {
          return _buildLoadingState();
        }
        break;
      case BiReportDetailByDocNoSuccess():
        if (state.docNo == widget.docno && state.data.isNotEmpty) {
          // Handle both SaleReportData and SaleReturnModel
          if (widget.reportType == BiReportType.sale &&
              state.data.first is SaleReportData) {
            return _buildDetailContent(
                saleData: state.data.first as SaleReportData);
          } else if (widget.reportType == BiReportType.saleReturn &&
              state.data.first is SaleReturnModel) {
            return _buildDetailContent(
                saleReturnData: state.data.first as SaleReturnModel);
          }
        }
        return _buildNoDataState();
      case BiReportDetailByDocNoFailure():
        if (state.docNo == widget.docno) {
          return _buildErrorState(state.message);
        }
        break;
      default:
        // Handle other states or default loading
        break;
    }

    // Default loading state
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
            global.language('transaction_loading_detail'),
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
            label: Text(global.language('transaction_try_again')),
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
            global.language('transaction_no_detail_data'),
            style: const TextStyle(
              fontSize: 16,
              color: Colors.grey,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailContent(
      {SaleReportData? saleData, SaleReturnModel? saleReturnData}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 1. ข้อมูลลูกค้า
        _buildCustomerInfoSection(
            saleData: saleData, saleReturnData: saleReturnData),
        const SizedBox(height: 16),
        // 2. ยอดสรุป
        _buildSummarySection(
            saleData: saleData, saleReturnData: saleReturnData),
        const SizedBox(height: 16),
        // 3. รายการสินค้า
        _buildTransactionSection(
            saleData: saleData, saleReturnData: saleReturnData),
      ],
    );
  }

  Widget _buildCustomerInfoSection(
      {SaleReportData? saleData, SaleReturnModel? saleReturnData}) {
    // Extract data from either model
    final String creditorName;
    final String creditorCode;
    final String inquiryType;

    if (saleData != null) {
      creditorName = ReportUtils.getCreditorName(saleData.creditornames).isEmpty
          ? global.language('transaction_general_customer')
          : ReportUtils.getCreditorName(saleData.creditornames);
      creditorCode =
          saleData.creditorcode.isEmpty ? '-' : saleData.creditorcode;
      inquiryType = saleData.inquirytype;
    } else if (saleReturnData != null) {
      creditorName = saleReturnData.creditorNames.isNotEmpty
          ? saleReturnData.creditorNames.first.getDisplayName()
          : global.language('transaction_general_customer');
      creditorCode = saleReturnData.creditorNames.isNotEmpty
          ? saleReturnData.creditorNames.first.code
          : '-';
      inquiryType = '2'; // Sale return type
    } else {
      creditorName = global.language('transaction_general_customer');
      creditorCode = '-';
      inquiryType = '';
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.business, color: Colors.indigo.shade600, size: 18),
            SizedBox(width: 8),
            Text(
              '1. ${global.language('transaction_customer_info')}',
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
            if (widget.reportType == BiReportType.sale) ...[
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: ReportUtils.getStatusColor(inquiryType),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  ReportUtils.getStatusText(inquiryType),
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: ReportUtils.getStatusTextColor(inquiryType),
                  ),
                ),
              ),
            ],
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
                          global.language('transaction_customer_name'),
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          creditorName,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  if (widget.reportType == BiReportType.sale) ...[
                    SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            global.language('transaction_customer_code'),
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade600,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            creditorCode,
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildSummarySection(
      {SaleReportData? saleData, SaleReturnModel? saleReturnData}) {
    // Extract data from either model
    final double totalValue;
    final double totalBeforeVat;
    final double totalVatValue;
    final double totalAmount;

    if (saleData != null) {
      totalValue = saleData.totalvalue;
      totalBeforeVat = saleData.totalbeforevat;
      totalVatValue = saleData.totalvatvalue;
      totalAmount = saleData.totalamount;
    } else if (saleReturnData != null) {
      totalValue = saleReturnData.totalValue;
      totalBeforeVat = saleReturnData.totalBeforeVat;
      totalVatValue = saleReturnData.totalVatValue;
      totalAmount = saleReturnData.totalAmount;
    } else {
      totalValue = 0.0;
      totalBeforeVat = 0.0;
      totalVatValue = 0.0;
      totalAmount = 0.0;
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.calculate_outlined,
                color: Colors.indigo.shade600, size: 16),
            SizedBox(width: 6),
            Text(
              '2. ${global.language('transaction_summary')}',
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
                child: _buildCompactSummaryItem(
                    global.language('transaction_product_value'), ReportUtils.formatCurrencySafe(totalValue)),
              ),
              SizedBox(width: 12),
              Expanded(
                child: _buildCompactSummaryItem(
                    global.language('transaction_before_vat'), ReportUtils.formatCurrencySafe(totalBeforeVat)),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildCompactSummaryItem(
                    'VAT', ReportUtils.formatCurrencySafe(totalVatValue),
                    isVat: true),
              ),
              SizedBox(width: 12),
              Expanded(
                flex: 2,
                child: _buildCompactSummaryItem(global.language('transaction_grand_total'),
                    ReportUtils.formatCurrencySafe(totalAmount),
                    isTotal: true),
              ),
            ],
          ),
        ),
      ],
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

  Widget _buildTransactionSection(
      {SaleReportData? saleData, SaleReturnModel? saleReturnData}) {
    // Extract transaction data from either model
    final List<dynamic> transactions;
    final int transactionCount;

    if (saleData != null) {
      transactions = saleData.transactions;
      transactionCount = saleData.transactions.length;
    } else if (saleReturnData != null) {
      transactions = saleReturnData.transactions ?? [];
      transactionCount = saleReturnData.transactions?.length ?? 0;
    } else {
      transactions = [];
      transactionCount = 0;
    }

    return Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.shopping_cart_outlined,
                  color: Colors.indigo.shade600, size: 18),
              SizedBox(width: 8),
              Text(
                '3. ${global.language('transaction_product_list')} ($transactionCount ${global.language('transaction_items')})',
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
                            global.language('transaction_product'),
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: Colors.indigo,
                            ),
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.language('transaction_quantity'),
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
                            global.language('transaction_price'),
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
                            global.language('transaction_total'),
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
                                global.language('transaction_no_products'),
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
                              if (transaction is SaleTransaction) {
                                return _buildSaleTransactionItem(
                                    transaction, index);
                              } else if (transaction
                                  is SaleReturnTransactionModel) {
                                return _buildSaleReturnTransactionItem(
                                    transaction, index);
                              }
                              return const SizedBox.shrink();
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

  Widget _buildSaleTransactionItem(SaleTransaction transaction, int index) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        border: index < (transaction.linenumber) - 1
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
          // Item name and barcode
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  ReportUtils.getItemNameSafe(transaction.itemnames),
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  transaction.barcode,
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
          // Quantity
          Expanded(
            child: Column(
              children: [
                Text(
                  '${transaction.qty}',
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                  ),
                  textAlign: TextAlign.center,
                ),
                Text(
                  ReportUtils.getUnitNameSafe(transaction.unitnames),
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.grey.shade600,
                  ),
                  textAlign: TextAlign.center,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          // Price
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.price),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          // Total amount
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.sumamount),
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

  Widget _buildSaleReturnTransactionItem(
      SaleReturnTransactionModel transaction, int index) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        border: index < transaction.itemNames.length - 1
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
              color: Colors.red.shade50,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Center(
              child: Text(
                '${index + 1}',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: Colors.red.shade700,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          // Item name and barcode
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  transaction.itemNames.isNotEmpty
                      ? transaction.itemNames.first.getDisplayName()
                      : '-',
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  transaction.barcode,
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
          // Quantity
          Expanded(
            child: Column(
              children: [
                Text(
                  '${transaction.qty}',
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                  ),
                  textAlign: TextAlign.center,
                ),
                Text(
                  transaction.unitNames.isNotEmpty
                      ? transaction.unitNames.first.getDisplayName()
                      : '-',
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.grey.shade600,
                  ),
                  textAlign: TextAlign.center,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          // Price
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.price),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          // Total amount
          Expanded(
            child: Text(
              widget.formatCurrency(transaction.sumAmount),
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Colors.red.shade700, // Use red for returns
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
          global.language('transaction_close'),
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ),
    );
  }
}
