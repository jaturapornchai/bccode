import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/cash_in_drawer/cash_in_drawer_bloc.dart';
import 'package:smlaicloud/bloc/cash_in_drawer/cash_in_drawer_event.dart';
import 'package:smlaicloud/bloc/cash_in_drawer/cash_in_drawer_state.dart';
import 'package:smlaicloud/model/cash_in_drawer_model.dart';
import 'package:smlaicloud/model/cash_drawer_models.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/shift_detail_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';

/// Cash in the Drawer Screen - Simple transaction list grouped by shift
class EnhancedCashInDrawerScreen extends StatefulWidget {
  const EnhancedCashInDrawerScreen({super.key});

  @override
  State<EnhancedCashInDrawerScreen> createState() =>
      _EnhancedCashInDrawerScreenState();
}

class _EnhancedCashInDrawerScreenState
    extends State<EnhancedCashInDrawerScreen>
    with global.ThemeRefreshMixin {
  // Constants
  static const int _itemsPerPage = 1000;

  // Controllers
  final TextEditingController _usercodeController = TextEditingController();
  final TextEditingController _posIDController = TextEditingController();

  // State variables
  late DateTime _fromDate;
  late DateTime _toDate;
  final Map<String, bool> _expandedItems = {};
  final Map<String, bool> _loadingShiftDetails = {};
  final Map<String, List<TransactionModel>> _shiftBillDetails = {};
  final Map<String, List<ShiftDetailModel>> _shiftDetails = {};
  List<ShiftSummary> _shifts = [];
  String _currentFilterType = 'cash_in'; // 'cash_in' or 'cash_out'

  @override
  void initState() {
    super.initState();
    _initializeDates();
    _loadData();
  }

  @override
  void dispose() {
    _usercodeController.dispose();
    _posIDController.dispose();
    super.dispose();
  }

  void _initializeDates() {
    final now = DateTime.now().toLocal();

    /// _fromDate now
    _fromDate = DateTime(now.year, now.month, now.day, 0, 0, 0);

    /// _toDate now
    _toDate = DateTime(now.year, now.month, now.day, 23, 59, 59);
  }

  /// Load data from API with filter
  void _loadData() {
    final fromDateStr = _formatDateForApi(_fromDate);
    final toDateStr = _formatDateForApi(_toDate);

    context.read<CashInDrawerBloc>().add(LoadCashInDrawerListWithFilter(
          page: 1,
          limit: _itemsPerPage,
          fromdate: fromDateStr,
          todate: toDateStr,
          usercode: _usercodeController.text.trim(),
          posid: _posIDController.text.trim(),
          filterType: _currentFilterType,
        ));
  }

  String _formatDateForApi(DateTime date) {
    final localDate = date.toLocal();
    return "${localDate.year}-${localDate.month.toString().padLeft(2, '0')}-${localDate.day.toString().padLeft(2, '0')}";
  }

  /// Transform raw data to individual transaction items (no grouping)
  void _transformData(List<CashInDrawerModel> rawData) {
    // Convert each raw data item to individual transaction
    final transactions = rawData.map((raw) {
      // Map doctype integer to enum
      CashDrawerTransactionType doctype;
      switch (raw.doctype ?? 1) {
        case 1:
          doctype = CashDrawerTransactionType.openShift;
          break;
        case 2:
          doctype = CashDrawerTransactionType.closeShift;
          break;
        case 3:
          doctype = CashDrawerTransactionType.addCash;
          break;
        case 4:
          doctype = CashDrawerTransactionType.withdrawCash;
          break;
        default:
          doctype = CashDrawerTransactionType.openShift;
      }

      // Create PaymentBreakdown
      final paymentBreakdown = PaymentBreakdown(
        cash: raw.amount ?? 0.0,
        creditCard: raw.creditcard ?? 0.0,
        promptPay: raw.promptpay ?? 0.0,
        transfer: raw.transfer ?? 0.0,
        cheque: raw.cheque ?? 0.0,
        coupon: raw.coupon ?? 0.0,
      );

      // Parse date string to DateTime
      DateTime parsedDate;
      try {
        parsedDate =
            DateTime.parse(raw.docdate ?? DateTime.now().toIso8601String());
      } catch (e) {
        parsedDate = DateTime.now();
      }

      // Create CashDrawerTransaction directly
      return CashDrawerTransaction(
        guidfixed: raw.guidfixed,
        usercode: raw.usercode ?? '',
        username: raw.username ?? '',
        posid: raw.posid ?? '',
        docno: raw.docno ?? '',
        doctype: doctype,
        docdate: parsedDate,
        remark: raw.remark,
        amount: raw.amount ?? 0.0,
        paymentBreakdown: paymentBreakdown,
      );
    }).toList();

    // Create individual ShiftSummary for each transaction (no grouping)
    final shifts = transactions.map((transaction) {
      return ShiftSummary(
        docno: transaction.docno,
        usercode: transaction.usercode,
        username: transaction.username,
        posid: transaction.posid,
        openShift: transaction.doctype == CashDrawerTransactionType.openShift
            ? transaction
            : null,
        closeShift: transaction.doctype == CashDrawerTransactionType.closeShift
            ? transaction
            : null,
        transactions: [transaction], // Single transaction per shift
      );
    }).toList();

    // Sort by date (newest first)
    shifts.sort((a, b) {
      final dateA = a.transactions.first.docdate.toLocal();
      final dateB = b.transactions.first.docdate.toLocal();
      return dateB.compareTo(dateA);
    });

    setState(() {
      _shifts = shifts;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      appBar: _buildAppBar(),
      body: Builder(
        builder: (context) {
          final scaleFactor = global.editFontScaleFactor;
          return MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(scaleFactor),
            ),
            child: Theme(
              data: Theme.of(context).copyWith(
                inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12 * scaleFactor,
                    vertical: 10 * scaleFactor,
                  ),
                ),
                iconTheme: IconThemeData(size: 24 * scaleFactor),
              ),
              child: Column(
        children: [
          _buildSearchPanel(),
          Expanded(
            child: BlocConsumer<CashInDrawerBloc, CashInDrawerState>(
              listener: _handleBlocListener,
              builder: _buildBlocContent,
            ),
          ),
        ],
      ),
            ),
          );
        },
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      elevation: 0,
      backgroundColor: global.theme.appBarColor,
      title: Text(
        global.language('pos_cash_management'),
        style: const TextStyle(fontWeight: FontWeight.w600),
      ),
      actions: [
        EditFontSizeControl(onChanged: () => setState(() {})),
        IconButton(
          icon: Icon(Icons.refresh),
          tooltip: global.language('refresh_data'),
          onPressed: _loadData,
        ),
      ],
    );
  }

  void _handleBlocListener(BuildContext context, CashInDrawerState state) {
    if (state is CashInDrawerLoadFailed) {
      _showErrorSnackBar(state.message);
    } else if (state is CashInDrawerLoadListSuccess) {
      _transformData(state.data);
    } else if (state is CashInDrawerShiftReportDetailsInProgress) {
      setState(() {
        _loadingShiftDetails[state.docno] = true;
      });
    } else if (state is CashInDrawerShiftReportDetailsSuccess) {
      setState(() {
        _loadingShiftDetails[state.docno] = false;
        _shiftBillDetails[state.docno] = state.billDetails;
        _shiftDetails[state.docno] = state.shifts;
      });
    } else if (state is CashInDrawerShiftReportDetailsFailed) {
      setState(() {
        _loadingShiftDetails[state.docno] = false;
      });
      _showErrorSnackBar('${global.language("cannot_load_bill_detail")}: ${state.message}');
    }
  }

  Widget _buildBlocContent(BuildContext context, CashInDrawerState state) {
    switch (state.runtimeType) {
      case CashInDrawerInProgress:
        return _buildLoadingView();
      case CashInDrawerLoadFailed:
        return _buildErrorView((state as CashInDrawerLoadFailed).message);
      case CashInDrawerLoadListSuccess:
        return _shifts.isEmpty ? _buildEmptyState() : _buildShiftList();
      case CashInDrawerShiftReportDetailsInProgress:
      case CashInDrawerShiftReportDetailsSuccess:
      case CashInDrawerShiftReportDetailsFailed:
        // For shift details states, keep showing the shift list
        return _shifts.isEmpty ? _buildEmptyState() : _buildShiftList();
      default:
        return _buildEmptyState();
    }
  }

  /// Optimized search panel
  Widget _buildSearchPanel() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 6,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          _buildFilterTypeRow(),
          const SizedBox(height: 12),
          _buildDateRangeRow(),
          const SizedBox(height: 12),
          _buildSearchFieldsRow(),
          const SizedBox(height: 16),
          _buildSearchButton(),
        ],
      ),
    );
  }

  Widget _buildFilterTypeRow() {
    return Row(
      children: [
        Expanded(
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(color: global.theme.dividerBorderColor),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Expanded(
                  child: InkWell(
                    onTap: () {
                      setState(() => _currentFilterType = 'cash_in');
                      _loadData();
                    },
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(8),
                      bottomLeft: Radius.circular(8),
                    ),
                    child: Container(
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      decoration: BoxDecoration(
                        color: _currentFilterType == 'cash_in'
                            ? global.theme.positiveHighlightTextColor
                            : Colors.transparent,
                        borderRadius: const BorderRadius.only(
                          topLeft: Radius.circular(8),
                          bottomLeft: Radius.circular(8),
                        ),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.add_circle,
                            size: 16,
                            color: _currentFilterType == 'cash_in'
                                ? global.theme.onPrimaryColor
                                : global.theme.positiveHighlightTextColor,
                          ),
                          SizedBox(width: 8),
                          Text(
                            global.language('pos_receive_money'),
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: _currentFilterType == 'cash_in'
                                  ? global.theme.onPrimaryColor
                                  : global.theme.positiveHighlightTextColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                Expanded(
                  child: InkWell(
                    onTap: () {
                      setState(() => _currentFilterType = 'cash_out');
                      _loadData();
                    },
                    borderRadius: const BorderRadius.only(
                      topRight: Radius.circular(8),
                      bottomRight: Radius.circular(8),
                    ),
                    child: Container(
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      decoration: BoxDecoration(
                        color: _currentFilterType == 'cash_out'
                            ? global.theme.negativeHighlightTextColor
                            : Colors.transparent,
                        borderRadius: const BorderRadius.only(
                          topRight: Radius.circular(8),
                          bottomRight: Radius.circular(8),
                        ),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.remove_circle,
                            size: 16,
                            color: _currentFilterType == 'cash_out'
                                ? global.theme.onPrimaryColor
                                : global.theme.negativeHighlightTextColor,
                          ),
                          SizedBox(width: 8),
                          Text(
                            global.language('pos_send_money'),
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: _currentFilterType == 'cash_out'
                                  ? global.theme.onPrimaryColor
                                  : global.theme.negativeHighlightTextColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDateRangeRow() {
    return Row(
      children: [
        Expanded(
          child: CustomDatePicker(
            labelText: global.language('from_date'),
            initialDate: _fromDate.toLocal(),
            // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
            onDateSelected: (date) {
              if (date != null) {
                setState(() => _fromDate = date.toLocal());
              }
            },
            decoration: _buildInputDecoration(Icons.calendar_today),
          ),
        ),
        SizedBox(width: 12),
        Expanded(
          child: CustomDatePicker(
            labelText: global.language('to_date'),
            initialDate: _toDate.toLocal(),
            // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
            onDateSelected: (date) {
              if (date != null) {
                setState(() => _toDate = date.toLocal());
              }
            },
            decoration: _buildInputDecoration(Icons.calendar_today),
          ),
        ),
      ],
    );
  }

  Widget _buildSearchFieldsRow() {
    return Row(
      children: [
        Expanded(
          child: TextFormField(
            controller: _usercodeController,
            decoration: _buildInputDecoration(Icons.person).copyWith(
              labelText: global.language('user_code'),
              hintText: global.language('search_by_user_code'),
            ),
          ),
        ),
        SizedBox(width: 12),
        Expanded(
          child: TextFormField(
            controller: _posIDController,
            decoration: _buildInputDecoration(Icons.point_of_sale).copyWith(
              labelText: global.language('pos_id'),
              hintText: global.language('search_by_pos_id'),
            ),
          ),
        ),
      ],
    );
  }

  InputDecoration _buildInputDecoration(IconData icon) {
    return InputDecoration(
      prefixIcon: Icon(icon, color: global.theme.iconColor),
      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
      isDense: true,
    );
  }

  Widget _buildSearchButton() {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton.icon(
        onPressed: _loadData,
        icon: Icon(Icons.search),
        label: Text(global.language('search')),
        style: ElevatedButton.styleFrom(
          backgroundColor: global.theme.primaryColor,
          foregroundColor: global.theme.onPrimaryColor,
          padding: const EdgeInsets.symmetric(vertical: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
    );
  }

  /// Optimized shift list
  Widget _buildShiftList() {
    return Container(
      color: global.theme.cardColor,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _shifts.length,
        itemBuilder: (context, index) =>
            _buildShiftCard(_shifts[index], index + 1),
      ),
    );
  }

  Widget _buildShiftCard(ShiftSummary shift, int sequenceNumber) {
    final transaction = shift.transactions.first; // Get the single transaction
    final isExpanded =
        _expandedItems[transaction.guidfixed ?? transaction.docno] ?? false;

    // Determine status color based on transaction type
    Color statusColor;
    if (_currentFilterType == 'cash_in') {
      statusColor = global.theme.positiveHighlightTextColor;
    } else {
      statusColor = global.theme.negativeHighlightTextColor;
    }

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Column(
        children: [
          _buildTransactionHeader(
              shift, transaction, sequenceNumber, statusColor, isExpanded),
          if (isExpanded) _buildTransactionDetails(shift, transaction),
        ],
      ),
    );
  }

  Widget _buildTransactionHeader(
      ShiftSummary shift,
      CashDrawerTransaction transaction,
      int sequenceNumber,
      Color statusColor,
      bool isExpanded) {
    final transactionKey = transaction.guidfixed ?? transaction.docno;

    return InkWell(
      onTap: () {
        setState(() => _expandedItems[transactionKey] = !isExpanded);

        // Load bill details when expanding the transaction
        if (!isExpanded &&
            _shiftBillDetails[shift.docno] == null &&
            (_loadingShiftDetails[shift.docno] != true)) {
          context
              .read<CashInDrawerBloc>()
              .add(LoadShiftReportDetails(docno: shift.docno));
        }
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          border: Border(left: BorderSide(color: statusColor, width: 4)),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildTransactionHeaderRow(
                shift, transaction, sequenceNumber, statusColor, isExpanded),
            const SizedBox(height: 12),
            _buildTransactionBasicInfo(transaction),
            const SizedBox(height: 8),
            _buildTransactionAmountSummary(transaction),
          ],
        ),
      ),
    );
  }

  Widget _buildTransactionHeaderRow(
      ShiftSummary shift,
      CashDrawerTransaction transaction,
      int sequenceNumber,
      Color statusColor,
      bool isExpanded) {
    return Row(
      children: [
        const SizedBox(width: 12),
        Icon(Icons.point_of_sale, size: 16, color: global.theme.iconSecondaryColor),
        const SizedBox(width: 8),
        Text(transaction.posid,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
        const SizedBox(width: 12),
        _buildTransactionTypeBadge(transaction, statusColor),
        const Spacer(),
        Icon(
          isExpanded ? Icons.keyboard_arrow_up : Icons.keyboard_arrow_down,
          color: global.theme.iconSecondaryColor,
        ),
      ],
    );
  }

  Widget _buildTransactionTypeBadge(
      CashDrawerTransaction transaction, Color color) {
    String statusText;
    IconData statusIcon;

    if (_currentFilterType == 'cash_in') {
      statusText = transaction.doctype == CashDrawerTransactionType.openShift
          ? global.language('receive_change')
          : global.language('receive_change');
      statusIcon = Icons.add_circle;
    } else {
      statusText = transaction.doctype == CashDrawerTransactionType.closeShift
          ? global.language('send_money')
          : global.language('send_money');
      statusIcon = Icons.remove_circle;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration:
          BoxDecoration(color: color, borderRadius: BorderRadius.circular(16)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(statusIcon, size: 14, color: global.theme.onPrimaryColor),
          const SizedBox(width: 4),
          Text(
            statusText,
            style: TextStyle(
                fontSize: 12, fontWeight: FontWeight.bold, color: global.theme.onPrimaryColor),
          ),
        ],
      ),
    );
  }

  /// Build basic transaction information section
  Widget _buildTransactionBasicInfo(CashDrawerTransaction transaction) {
    // Format transaction time with local timezone and Thai Buddhist calendar
    String formatTransactionTime() {
      final transactionDate = transaction.docdate.toLocal();
      final DateFormat dateFormat = DateFormat('d MMM y HH:mm', 'th_TH');

      // Convert to Buddhist calendar (add 543 years)
      final buddhistDate = DateTime(
          transactionDate.year + 543,
          transactionDate.month,
          transactionDate.day,
          transactionDate.hour,
          transactionDate.minute);

      return dateFormat.format(buddhistDate);
    }

    return Column(
      children: [
        // Transaction time info with local timezone
        Row(
          children: [
            Icon(Icons.access_time, size: 16, color: global.theme.iconSecondaryColor),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                formatTransactionTime(),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: global.theme.textColor,
                ),
              ),
            ),
          ],
        ),

        const SizedBox(height: 8),

        // User info
        Row(
          children: [
            Icon(Icons.person, size: 16, color: global.theme.iconSecondaryColor),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                '${transaction.username} (${transaction.usercode})',
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],
        ),

        // Transaction type info
        if (transaction.remark != null && transaction.remark!.isNotEmpty) ...[
          const SizedBox(height: 8),
          Row(
            children: [
              Icon(Icons.note, size: 16, color: global.theme.iconSecondaryColor),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  transaction.remark!,
                  style: TextStyle(
                    fontSize: 13,
                    color: global.theme.iconSecondaryColor,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }

  /// Build transaction amount summary
  Widget _buildTransactionAmountSummary(CashDrawerTransaction transaction) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: _currentFilterType == 'cash_in'
            ? global.theme.positiveHighlightTextColor.withValues(alpha: 0.1)
            : global.theme.negativeHighlightTextColor.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: _currentFilterType == 'cash_in'
              ? global.theme.positiveHighlightTextColor.withValues(alpha: 0.3)
              : global.theme.negativeHighlightTextColor.withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        children: [
          Icon(
            _currentFilterType == 'cash_in'
                ? Icons.trending_up
                : Icons.trending_down,
            size: 16,
            color: _currentFilterType == 'cash_in'
                ? global.theme.positiveHighlightTextColor
                : global.theme.negativeHighlightTextColor,
          ),
          const SizedBox(width: 8),
          Text(
            transaction.doctypeDisplayText,
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w500,
              color: _currentFilterType == 'cash_in'
                  ? global.theme.positiveHighlightTextColor
                  : global.theme.negativeHighlightTextColor,
            ),
          ),
          const Spacer(),
          Text(
            _formatCurrency(transaction.amount),
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: _currentFilterType == 'cash_in'
                  ? global.theme.positiveHighlightTextColor
                  : global.theme.negativeHighlightTextColor,
            ),
          ),
        ],
      ),
    );
  }

  /// Build simple transaction details
  Widget _buildTransactionDetails(
      ShiftSummary shift, CashDrawerTransaction transaction) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.backgroundColor,
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(12),
          bottomRight: Radius.circular(12),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Document number
          _buildDetailRow(
            global.language('docno'),
            transaction.docno,
            Icons.description,
          ),

          const SizedBox(height: 8),

          // Transaction type
          _buildDetailRow(
            global.language('transaction_type'),
            transaction.doctypeDisplayText,
            Icons.category,
          ),

          // Show summary data if bill details are loaded
          if (_shiftBillDetails[shift.docno] != null &&
              _shiftBillDetails[shift.docno]!.isNotEmpty) ...[
            const SizedBox(height: 16),
            _buildShiftSummarySection(shift),
          ] else if (_loadingShiftDetails[shift.docno] == true) ...[
            const SizedBox(height: 16),
            const Center(
              child: CircularProgressIndicator(),
            ),
          ] else ...[
            SizedBox(height: 8),
            Center(
              child: Padding(
                padding: EdgeInsets.symmetric(vertical: 8),
                child: Text(
                  global.language('tap_to_view_detail'),
                  style: TextStyle(
                    fontSize: 13,
                    color: global.theme.iconSecondaryColor,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  /// Build shift summary section with sales and cash drawer data
  Widget _buildShiftSummarySection(ShiftSummary shift) {
    final billDetails = _shiftBillDetails[shift.docno] ?? [];

    final summary = _calculateSummaryForShift(shift, billDetails);
    final paymentBreakdown =
        _calculatePaymentBreakdownForShift(shift, billDetails);
    final cashDrawerSummary =
        _calculateCashDrawerSummaryForShift(shift, billDetails);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Sales Summary Section
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: global.theme.cardColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: global.theme.dividerBorderColor),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.trending_up,
                      size: 18, color: global.theme.positiveHighlightTextColor),
                  SizedBox(width: 8),
                  Text(
                    global.language('sales_summary'),
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: global.theme.textColor,
                    ),
                  ),
                ],
              ),
              SizedBox(height: 16),
              _buildSummaryRow(
                global.language('sales_amount'),
                summary.totalSales,
                Icons.trending_up,
                global.theme.positiveHighlightTextColor,
              ),
              SizedBox(height: 8),
              _buildSummaryRow(
                global.language('cancelled_sales'),
                summary.totalCancelledSales,
                Icons.cancel,
                global.theme.warningHighlightTextColor,
              ),
              Divider(height: 16),
              _buildSummaryRow(
                global.language('net_sales'),
                summary.netSales,
                Icons.account_balance,
                global.theme.infoHighlightTextColor,
                isBold: true,
              ),
            ],
          ),
        ),

        const SizedBox(height: 16),

        // Payment Breakdown Section
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: global.theme.cardColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: global.theme.dividerBorderColor),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.payment, size: 18, color: global.theme.primaryColor),
                  SizedBox(width: 8),
                  Text(
                    global.language('payment_detail'),
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: global.theme.textColor,
                    ),
                  ),
                ],
              ),
              SizedBox(height: 16),
              _buildPaymentRow(
                global.language('cash'),
                paymentBreakdown.cash,
                Icons.money,
                global.theme.positiveHighlightTextColor,
              ),
              SizedBox(height: 8),
              _buildPaymentRow(
                global.language('credit_card'),
                paymentBreakdown.creditCard,
                Icons.credit_card,
                global.theme.primaryColor,
              ),
              SizedBox(height: 8),
              _buildPaymentRow(
                global.language('bank_transfer'),
                paymentBreakdown.transfer,
                Icons.transform,
                global.theme.primaryColor,
              ),
              const SizedBox(height: 8),
              _buildPaymentRow(
                'QR Code',
                paymentBreakdown.qr,
                Icons.qr_code,
                global.theme.infoHighlightTextColor,
              ),
              Divider(height: 16),
              _buildPaymentRow(
                global.language('grand_total'),
                paymentBreakdown.total,
                Icons.account_balance,
                global.theme.textColor,
                isBold: true,
              ),
            ],
          ),
        ),

        const SizedBox(height: 16),

        // Cash Drawer Summary Section
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: global.theme.backgroundColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
                color: global.theme.primaryColor.withValues(alpha: 0.3)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.account_balance_wallet,
                    size: 18,
                    color: global.theme.primaryColor,
                  ),
                  SizedBox(width: 8),
                  Text(
                    global.language('cash_drawer_summary'),
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: global.theme.primaryColor,
                    ),
                  ),
                ],
              ),
              SizedBox(height: 16),
              _buildCashDrawerRow(
                global.language('change_money'),
                cashDrawerSummary.initialCash,
                Icons.savings,
                global.theme.primaryColor,
              ),
              SizedBox(height: 8),
              _buildCashDrawerRow(
                global.language('received_from_sales'),
                cashDrawerSummary.salesCash,
                Icons.point_of_sale,
                global.theme.positiveHighlightTextColor,
              ),
              SizedBox(height: 8),
              _buildCashDrawerRow(
                global.language('cash_out'),
                cashDrawerSummary.cashWithdrawn,
                Icons.money_off,
                global.theme.negativeHighlightTextColor,
              ),
              Divider(height: 16),
              _buildCashDrawerRow(
                global.language('cash_remaining_in_drawer'),
                cashDrawerSummary.remainingCash,
                Icons.account_balance_wallet,
                global.theme.primaryColor,
                isBold: true,
                isHighlight: true,
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildSummaryRow(
    String label,
    double amount,
    IconData icon,
    Color color, {
    bool isBold = false,
  }) {
    return Row(
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            label,
            style: TextStyle(
              fontSize: 13,
              fontWeight: isBold ? FontWeight.bold : FontWeight.w500,
              color: global.theme.textColor,
            ),
          ),
        ),
        Text(
          _formatCurrency(amount),
          style: TextStyle(
            fontSize: 13,
            fontWeight: isBold ? FontWeight.bold : FontWeight.w600,
            color: color,
          ),
        ),
      ],
    );
  }

  Widget _buildPaymentRow(
    String label,
    double amount,
    IconData icon,
    Color color, {
    bool isBold = false,
  }) {
    return Row(
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            label,
            style: TextStyle(
              fontSize: 13,
              fontWeight: isBold ? FontWeight.bold : FontWeight.w500,
              color: global.theme.textColor,
            ),
          ),
        ),
        Text(
          _formatCurrency(amount),
          style: TextStyle(
            fontSize: 13,
            fontWeight: isBold ? FontWeight.bold : FontWeight.w600,
            color: color,
          ),
        ),
      ],
    );
  }

  Widget _buildCashDrawerRow(
    String label,
    double amount,
    IconData icon,
    Color color, {
    bool isBold = false,
    bool isHighlight = false,
  }) {
    return Container(
      padding: isHighlight ? const EdgeInsets.all(8) : EdgeInsets.zero,
      decoration: isHighlight
          ? BoxDecoration(
              color: color.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            )
          : null,
      child: Row(
        children: [
          Icon(icon, size: 16, color: color),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: isBold ? FontWeight.bold : FontWeight.w500,
                color: global.theme.textColor,
              ),
            ),
          ),
          Text(
            _formatCurrency(amount),
            style: TextStyle(
              fontSize: 13,
              fontWeight: isBold ? FontWeight.bold : FontWeight.w600,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  // Calculation methods for shift summaries
  SalesSummary _calculateSummaryForShift(
      ShiftSummary shift, List<TransactionModel> billDetails) {
    final openShiftDate = shift.openShift?.docdate;
    final latestTransactionDate = shift.transactions.isNotEmpty
        ? shift.transactions.last.docdate
        : DateTime.now();

    // Filter bills between open shift date and latest transaction date
    final relevantBills = billDetails.where((bill) {
      final billDate = DateTime.parse(bill.docdatetime);

      if (openShiftDate != null) {
        return billDate
                .isAfter(openShiftDate.subtract(const Duration(seconds: 1))) &&
            billDate.isBefore(
                latestTransactionDate.add(const Duration(seconds: 1)));
      } else {
        return billDate
            .isBefore(latestTransactionDate.add(const Duration(seconds: 1)));
      }
    }).toList();

    // Calculate sales amounts
    final totalSales = relevantBills.fold(
        0.0,
        (sum, bill) =>
            sum + (bill.totalamountafterdiscount ?? bill.totalamount));

    final totalCancelledSales = relevantBills
        .where((bill) => bill.iscancel == true)
        .fold(
            0.0,
            (sum, bill) =>
                sum + (bill.totalamountafterdiscount ?? bill.totalamount));

    final netSales = totalSales - totalCancelledSales;

    // Calculate cash received (doctype 1 and 3) up to latest transaction date
    final cashReceived = shift.transactions
        .where((tx) =>
            (tx.doctype == CashDrawerTransactionType.openShift ||
                tx.doctype == CashDrawerTransactionType.addCash) &&
            tx.docdate.isBefore(
                latestTransactionDate.add(const Duration(seconds: 1))))
        .fold(0.0, (sum, tx) => sum + tx.amount);

    return SalesSummary(
      totalCashReceived: cashReceived,
      totalSales: totalSales,
      totalCancelledSales: totalCancelledSales,
      netSales: netSales,
    );
  }

  PaymentBreakdownSummary _calculatePaymentBreakdownForShift(
      ShiftSummary shift, List<TransactionModel> billDetails) {
    double cash = 0;
    double creditCard = 0;
    double transfer = 0;
    double cheque = 0;
    double coupon = 0;
    double qr = 0;

    final openShiftDate = shift.openShift?.docdate;
    final latestTransactionDate = shift.transactions.isNotEmpty
        ? shift.transactions.last.docdate
        : DateTime.now();

    final relevantBills = billDetails.where((bill) {
      final billDate = DateTime.parse(bill.docdatetime);

      if (openShiftDate != null) {
        return billDate
                .isAfter(openShiftDate.subtract(const Duration(seconds: 1))) &&
            billDate.isBefore(
                latestTransactionDate.add(const Duration(seconds: 1)));
      } else {
        return billDate
            .isBefore(latestTransactionDate.add(const Duration(seconds: 1)));
      }
    }).toList();

    // Aggregate payment data from all relevant bills
    for (final bill in relevantBills) {
      if (bill.iscancel == true) continue; // Skip cancelled bills

      // Extract payment data from paymentStructs if available
      if (bill.paymentStructs != null &&
          bill.paymentStructs!.isNotEmpty) {
        for (final payment in bill.paymentStructs!) {
          final amount = payment.amount ?? 0;

          switch (payment.trans_flag) {
            case 1: // Credit Card
              creditCard += amount;
              break;
            case 2: // Transfer
              transfer += amount;
              break;
            case 3: // Cheque
              cheque += amount;
              break;
            case 4: // Coupon
              coupon += amount;
              break;
            case 5: // QR Code
              qr += amount;
              break;
          }
        }
      }

      // Get cash amount from paycashamount field
      cash += bill.paycashamount ?? 0;

      // Alternative: if paymentStructs is empty, try to get from summary fields
      if (bill.paymentStructs == null ||
          bill.paymentStructs!.isEmpty) {
        creditCard += bill.sumcreditcard ?? 0;
        transfer += bill.summoneytransfer ?? 0;
        cheque += bill.sumcheque ?? 0;
        coupon += bill.sumcoupon ?? 0;
        qr += bill.sumqrcode ?? 0;
      }
    }

    return PaymentBreakdownSummary(
      cash: cash,
      creditCard: creditCard,
      transfer: transfer,
      cheque: cheque,
      coupon: coupon,
      qr: qr,
    );
  }

  CashDrawerSummary _calculateCashDrawerSummaryForShift(
      ShiftSummary shift, List<TransactionModel> billDetails) {
    // ดึงข้อมูล shifts จาก _shiftDetails แทนการใช้ shift.transactions
    final shiftDetails = _shiftDetails[shift.docno] ?? [];

    // แปลง ShiftDetailModel เป็น transactions สำหรับการคำนวณ
    final allShiftTransactions = <CashDrawerTransaction>[];

    for (final detail in shiftDetails) {
      CashDrawerTransactionType doctype;
      switch (detail.doctype ?? 1) {
        case 1:
          doctype = CashDrawerTransactionType.openShift;
          break;
        case 2:
          doctype = CashDrawerTransactionType.closeShift;
          break;
        case 3:
          doctype = CashDrawerTransactionType.addCash;
          break;
        case 4:
          doctype = CashDrawerTransactionType.withdrawCash;
          break;
        default:
          doctype = CashDrawerTransactionType.openShift;
      }

      DateTime parsedDate;
      try {
        parsedDate =
            DateTime.parse(detail.docdate ?? DateTime.now().toIso8601String());
      } catch (e) {
        parsedDate = DateTime.now();
      }

      allShiftTransactions.add(CashDrawerTransaction(
        usercode: detail.usercode ?? '',
        username: detail.username ?? '',
        posid: detail.posid ?? '',
        docno: detail.docno ?? '',
        doctype: doctype,
        docdate: parsedDate,
        remark: detail.remark,
        amount: detail.amount ?? 0.0,
        paymentBreakdown: const PaymentBreakdown(),
      ));
    }

    // หาวันที่เปิดกะ และ วันที่ของ transaction ปัจจุบัน
    final openShiftTransaction = allShiftTransactions
        .where((tx) => tx.doctype == CashDrawerTransactionType.openShift)
        .firstOrNull;
    final openShiftDate = openShiftTransaction?.docdate;

    final currentTransaction = shift.transactions.first;
    final currentTransactionDate = currentTransaction.docdate;

    // 1. เงินทอน - sum ยอด doctype 1,3 (เปิดกะ + เพิ่มเงิน) ที่เกิดขึ้นก่อนหรือเท่ากับ transaction ปัจจุบัน
    final initialCash = allShiftTransactions
        .where((tx) =>
            (tx.doctype == CashDrawerTransactionType.openShift ||
                tx.doctype == CashDrawerTransactionType.addCash) &&
            tx.docdate.isBefore(
                currentTransactionDate.add(const Duration(seconds: 1))))
        .fold(0.0, (sum, tx) => sum + tx.amount);

    // 2. รับเงินจากการขาย - sum ยอด billDetails.paycashamount
    final relevantBills = billDetails.where((bill) {
      final billDate = DateTime.parse(bill.docdatetime);

      if (openShiftDate != null) {
        return billDate
                .isAfter(openShiftDate.subtract(const Duration(seconds: 1))) &&
            billDate.isBefore(
                currentTransactionDate.add(const Duration(seconds: 1)));
      } else {
        return billDate
            .isBefore(currentTransactionDate.add(const Duration(seconds: 1)));
      }
    }).toList();

    final salesCash = relevantBills
        .where((bill) => bill.iscancel != true) // ไม่นับบิลที่ยกเลิก
        .fold(0.0, (sum, bill) => sum + (bill.paycashamount ?? 0));

    // 3. นำเงินสดออก - sum ยอด doctype 2,4 (ปิดกะ + เบิกเงิน) ที่เกิดขึ้นก่อนหรือเท่ากับ transaction ปัจจุบัน
    final cashWithdrawn = allShiftTransactions
        .where((tx) =>
            (tx.doctype == CashDrawerTransactionType.closeShift ||
                tx.doctype == CashDrawerTransactionType.withdrawCash) &&
            tx.docdate.isBefore(
                currentTransactionDate.add(const Duration(seconds: 1))))
        .fold(0.0, (sum, tx) => sum + tx.amount);

    // 4. เงินคงเหลือในลิ้นชัก = เงินทอน + รับเงินจากการขาย - นำเงินสดออก
    final remainingCash = initialCash + salesCash - cashWithdrawn;

    return CashDrawerSummary(
      initialCash: initialCash,
      salesCash: salesCash,
      cashWithdrawn: cashWithdrawn,
      remainingCash: remainingCash,
    );
  }

  // Helper methods - add missing methods
  Widget _buildDetailRow(String label, String value, IconData icon) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 16, color: global.theme.iconSecondaryColor),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.iconSecondaryColor,
                ),
              ),
              Text(
                value,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: global.theme.textSecondaryColor,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildLoadingView() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const CircularProgressIndicator(),
          SizedBox(height: 16),
          Text(
            global.language('loading_data'),
            style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorView(String message) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: global.theme.negativeHighlightColor,
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.error_outline,
                size: 48, color: global.theme.negativeHighlightTextColor),
          ),
          SizedBox(height: 16),
          Text(
            global.language('error_occurred'),
            style: const TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              message,
              textAlign: TextAlign.center,
              style: TextStyle(color: global.theme.textColor),
            ),
          ),
          SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: _loadData,
            icon: Icon(Icons.refresh),
            label: Text(global.language('try_again')),
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.primaryColor,
              foregroundColor: global.theme.onPrimaryColor,
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.account_balance_wallet,
                size: 64, color: global.theme.primaryColor),
          ),
          SizedBox(height: 24),
          Text(
            global.language('no_cash_drawer_items'),
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: global.theme.textColor,
            ),
          ),
          SizedBox(height: 8),
          Text(
            global.language('no_transactions_in_period'),
            style: TextStyle(
              fontSize: 14,
              color: global.theme.iconSecondaryColor,
            ),
          ),
        ],
      ),
    );
  }

  String _formatCurrency(double? amount) {
    if (amount == null) return '฿0.00';
    return '฿${NumberFormat("#,##0.00", "th_TH").format(amount)}';
  }

  void _showErrorSnackBar(String message) {
    if (!mounted) return;
    global.showErrorSnackBar(context, message);
  }
}

/// Sales Summary Data Class
class SalesSummary {
  final double totalCashReceived;
  final double totalSales;
  final double totalCancelledSales;
  final double netSales;

  const SalesSummary({
    required this.totalCashReceived,
    required this.totalSales,
    required this.totalCancelledSales,
    required this.netSales,
  });
}

/// Payment Breakdown Data Class
class PaymentBreakdownSummary {
  final double cash;
  final double creditCard;
  final double transfer;
  final double cheque;
  final double coupon;
  final double qr;

  const PaymentBreakdownSummary({
    required this.cash,
    required this.creditCard,
    required this.transfer,
    required this.cheque,
    required this.coupon,
    required this.qr,
  });

  double get total => cash + creditCard + transfer + cheque + coupon + qr;
}

/// Cash Drawer Summary Data Class
class CashDrawerSummary {
  final double initialCash;
  final double salesCash;
  final double cashWithdrawn;
  final double remainingCash;

  const CashDrawerSummary({
    required this.initialCash,
    required this.salesCash,
    required this.cashWithdrawn,
    required this.remainingCash,
  });
}
