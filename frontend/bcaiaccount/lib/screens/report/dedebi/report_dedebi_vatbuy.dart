import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/vat_buy_model.dart';
import 'package:smlaicloud/model/bi_report/vat_buy_summary_model.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/screens/report/dedebi/components/filter_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/report_condition_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/vat_buy_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/total_summary_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/export_report_dialog.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// รายงานภาษีซื้อ (VAT Buy Report)
/// แสดงรายงานภาษีมูลค่าเพิ่มจากการซื้อ พร้อมข้อมูลใบกำกับภาษี
class ReportVatBuyScreen extends StatefulWidget {
  const ReportVatBuyScreen({super.key});

  @override
  State<ReportVatBuyScreen> createState() => _ReportVatBuyScreenState();
}

class _ReportVatBuyScreenState extends State<ReportVatBuyScreen> with global.ThemeRefreshMixin {
  // Report polling configuration
  static const Duration _pollInterval = Duration(seconds: 2);
  static const Duration _reportTimeout = Duration(minutes: 10);

  // Report condition variables
  DateTime? _fromDate;
  DateTime? _toDate;

  // Branch filter
  BranchSelectionModel _selectedBranches = const BranchSelectionModel(
    selectedBranches: [],
    isCancel: false,
  );

  // Pagination variables
  int _pageSize = 20;

  // Current data
  BiReportMeta? _currentMeta;
  List<VatBuyModel> _currentData = [];
  String _currentJobId = '';

  // Total summary data
  VatBuySummary? _currentTotalSummary;
  bool _totalSummaryLoading = false;

  // Bloc reference
  BiReportBloc? _biReportBloc;

  // Filter panel visibility state
  bool _isFilterPanelVisible = true;

  @override
  void initState() {
    if (kDebugMode) {
      AppLogger.debug('🚀 ReportVatBuyScreen initState called');
    }
    super.initState();

    // Initialize default dates (current month)
    final now = DateTime.now();
    _fromDate = DateTime(now.year, now.month, 1);
    _toDate = DateTime(now.year, now.month + 1, 0);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _biReportBloc = context.read<BiReportBloc>();

    if (mounted) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _biReportBloc != null) {
          _biReportBloc!.add(const ResetBiReportState());
          _showConditionDialog();
        }
      });
    }
  }

  @override
  void dispose() {
    if (_biReportBloc != null) {
      _biReportBloc!.add(const ResetBiReportState());
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      onPopInvoked: (didPop) {
        if (_biReportBloc != null) {
          _biReportBloc!.add(const ResetBiReportState());
        }
      },
      child: Scaffold(
        backgroundColor: global.theme.backgroundColor,
        appBar: _buildAppBar(),
        body: BlocConsumer<BiReportBloc, BiReportState>(
          buildWhen: (previous, current) {
            return current is! BiReportSummaryLoading &&
                current is! BiReportSummarySuccess &&
                current is! BiReportSummaryFailure;
          },
          listener: _handleStateChanges,
          builder: (context, state) => _buildContent(state),
        ),
      ),
    );
  }

  AppBar _buildAppBar() {
    return AppBar(
      title: Text(
        global.language('report_vat_buy'),
        style: TextStyle(fontWeight: FontWeight.w600),
      ),
      centerTitle: true,
      backgroundColor: Colors.indigo.shade600,
      foregroundColor: global.theme.onPrimaryColor,
      elevation: 2,
      leading: IconButton(
        icon: Icon(Icons.arrow_back),
        onPressed: () {
          if (_biReportBloc != null && _currentJobId.isNotEmpty) {
            _biReportBloc!.add(CancelBiReportRequested(jobId: _currentJobId));
          }
          if (_biReportBloc != null) {
            _biReportBloc!.add(const ResetBiReportState());
          }
          Navigator.of(context).pop();
        },
      ),
      actions: [
        IconButton(
          icon: Icon(
            _isFilterPanelVisible ? Icons.visibility_off : Icons.visibility,
          ),
          onPressed: () =>
              setState(() => _isFilterPanelVisible = !_isFilterPanelVisible),
          tooltip: _isFilterPanelVisible
              ? global.language('report_dedebi_hide_filter_panel')
              : global.language('report_dedebi_show_filter_panel'),
        ),
        IconButton(
          icon: Icon(Icons.filter_alt_outlined),
          onPressed: _showConditionDialog,
          tooltip: global.language('report_dedebi_set_search_conditions'),
        ),
        IconButton(
          icon: Icon(Icons.refresh),
          onPressed: () {
            if (_biReportBloc != null) {
              _biReportBloc!.add(const ResetBiReportState());
            }
            _refreshReportData();
          },
          tooltip: global.language('report_dedebi_refresh_data'),
        ),
        IconButton(
          icon: Icon(Icons.download),
          onPressed: _showExportDialog,
          tooltip: global.language('report_dedebi_export_report'),
        ),
      ],
    );
  }

  void _handleStateChanges(BuildContext context, BiReportState state) {
    switch (state) {
      case BiReportGenerateFailure():
        _showErrorSnackBar(state.message);
        break;
      case BiReportDetailFailure():
        _showErrorSnackBar(state.message);
        break;
      case BiReportGenerateSuccess():
        _updateCurrentData(state);
        _requestTotalSummary(state.jobId);
        break;
      case BiReportDetailSuccess():
        _updateCurrentData(state);
        break;
      case BiReportSummaryLoading():
        setState(() => _totalSummaryLoading = true);
        break;
      case BiReportSummarySuccess():
        setState(() {
          _currentTotalSummary = state.vatBuySummaryData;
          _totalSummaryLoading = false;
        });
        break;
      case BiReportSummaryFailure():
        setState(() => _totalSummaryLoading = false);
        _showErrorSnackBar(state.message);
        break;
      default:
        break;
    }
  }

  void _updateCurrentData(dynamic state) {
    if (state is BiReportGenerateSuccess || state is BiReportDetailSuccess) {
      setState(() {
        _currentData = state.data.whereType<VatBuyModel>().toList();
        _currentMeta = state.meta;
        _currentJobId = state.jobId;
      });
    }
  }

  void _requestTotalSummary(String jobId) {
    context.read<BiReportBloc>().add(
      GetBiReportSummaryRequested(
        reportType: BiReportType.vatBuy,
        jobId: jobId,
        token: global.appConfig.getString("token")!,
      ),
    );
  }

  Widget _buildContent(BiReportState state) {
    return switch (state) {
      BiReportInitial() => _buildInitialState(),
      BiReportGenerating() => _buildLoadingState(
        global.language('report_dedebi_generating_report'),
        0,
      ),
      BiReportGenerateProgress() => () {
        _currentJobId = state.jobId;
        return _buildLoadingState(state.statusMessage, state.progress);
      }(),
      BiReportGenerateSuccess() => _buildSuccessContent(
        state.data.whereType<VatBuyModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportDetailLoading() => _buildContentWithLoading(state.jobId),
      BiReportDetailSuccess() => _buildSuccessContent(
        state.data.whereType<VatBuyModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportGenerateFailure() => () {
        _currentJobId = '';
        return _buildErrorState(state.message);
      }(),
      BiReportDetailFailure() => _buildErrorState(state.message),
      _ =>
        _currentData.isNotEmpty
            ? _buildSuccessContent(_currentData, _currentJobId, _currentMeta!)
            : _buildInitialState(),
    };
  }

  Widget _buildContentWithLoading(String jobId) {
    if (_currentData.isNotEmpty && _currentMeta != null) {
      return _buildMainLayoutWithLoading(_currentData, jobId, _currentMeta!);
    }
    return _buildLoadingState(global.language('report_dedebi_loading_data'), 0);
  }

  Widget _buildSuccessContent(
    List<VatBuyModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return _buildMainLayout(data, jobId, meta);
  }

  Widget _buildMainLayout(
    List<VatBuyModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(child: _buildMainContent(data, jobId, meta)),
        if (_isFilterPanelVisible) _buildSidebar(data),
      ],
    );
  }

  Widget _buildMainLayoutWithLoading(
    List<VatBuyModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: Stack(
            children: [
              Opacity(
                opacity: 0.5,
                child: _buildMainContent(data, jobId, meta),
              ),
              _buildLoadingOverlay(),
            ],
          ),
        ),
        if (_isFilterPanelVisible) _buildSidebar(data),
      ],
    );
  }

  Widget _buildSidebar(List<VatBuyModel> data) {
    return Container(
      width: 260,
      padding: const EdgeInsets.only(top: 16, bottom: 16, left: 0, right: 16),
      child: SingleChildScrollView(
        child: Column(
          children: [
            FilterPanel(
              reportType: BiReportType.vatBuy,
              fromDate: _fromDate,
              toDate: _toDate,
              selectedBranches: _selectedBranches,
              onShowConditionDialog: _showConditionDialog,
              onRefresh: _refreshReportData,
            ),
            const SizedBox(height: 16),
            TotalSummaryPanel(
              vatBuySummary: _currentTotalSummary,
              isLoading: _totalSummaryLoading,
              reportType: BiReportType.vatBuy,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMainContent(
    List<VatBuyModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Expanded(child: _buildDataGrid(data)),
          const SizedBox(height: 16),
          _buildPaginationControls(jobId, meta),
        ],
      ),
    );
  }

  Widget _buildDataGrid(List<VatBuyModel> data) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: global.theme.dividerBorderColor),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.05),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: VatBuyTableView(data: data),
    );
  }

  Widget _buildLoadingOverlay() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor.withValues(alpha: 0.8),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            SizedBox(
              width: 40,
              height: 40,
              child: CircularProgressIndicator(
                strokeWidth: 4,
                valueColor: AlwaysStoppedAnimation<Color>(
                  Colors.indigo.shade600,
                ),
              ),
            ),
            SizedBox(height: 16),
            Text(
              global.language('report_dedebi_loading_new_page'),
              style: TextStyle(
                fontSize: 14,
                color: global.theme.iconColor,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showErrorSnackBar(String message) {
    global.showErrorSnackBar(context, message);
  }

  Widget _buildInitialState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: Colors.indigo.shade50,
              borderRadius: BorderRadius.circular(100),
            ),
            child: Icon(
              Icons.receipt_long,
              size: 64,
              color: Colors.indigo.shade600,
            ),
          ),
          SizedBox(height: 32),
          Text(
            global.language('report_vat_buy'),
            style: TextStyle(
              fontSize: 24,
              color: global.theme.textColor,
              fontWeight: FontWeight.w600,
            ),
          ),
          SizedBox(height: 8),
          Container(
            constraints: BoxConstraints(maxWidth: 300),
            child: Text(
              global.language('report_dedebi_set_conditions_to_view'),
              style: TextStyle(
                fontSize: 16,
                color: global.theme.textSecondaryColor,
                fontWeight: FontWeight.w400,
              ),
              textAlign: TextAlign.center,
            ),
          ),
          SizedBox(height: 32),
          ElevatedButton.icon(
            onPressed: _showConditionDialog,
            icon: Icon(Icons.filter_alt),
            label: Text(global.language('report_dedebi_set_search_conditions')),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.indigo.shade600,
              foregroundColor: global.theme.onPrimaryColor,
              padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              elevation: 3,
              textStyle: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLoadingState(String message, int progress) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          SizedBox(
            width: 80,
            height: 80,
            child: Stack(
              alignment: Alignment.center,
              children: [
                SizedBox(
                  width: 80,
                  height: 80,
                  child: CircularProgressIndicator(
                    value: null,
                    strokeWidth: 6,
                    backgroundColor: global.theme.dividerBorderColor,
                    valueColor: AlwaysStoppedAnimation<Color>(
                      Colors.indigo.shade600,
                    ),
                  ),
                ),
                if (progress > 0)
                  Text(
                    '$progress%',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: Colors.indigo.shade600,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          Text(
            message,
            style: TextStyle(
              fontSize: 16,
              color: global.theme.iconColor,
              fontWeight: FontWeight.w500,
            ),
          ),
          SizedBox(height: 8),
          Text(
            global.language('report_dedebi_please_wait'),
            style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
          ),
        ],
      ),
    );
  }

  Widget _buildPaginationControls(String jobId, BiReportMeta meta) {
    final totalPages = meta.totalPage;
    final currentPage = meta.page;
    final totalItems = meta.total;
    final startItem = (currentPage - 1) * meta.size + 1;
    final endItem = (currentPage * meta.size > totalItems)
        ? totalItems
        : currentPage * meta.size;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(16),
          bottomRight: Radius.circular(16),
        ),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Text(
              '${global.language('report_dedebi_showing_items')} $startItem-$endItem ${global.language('report_dedebi_all').toLowerCase()} $totalItems ${global.language('report_dedebi_items')}',
              style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
            ),
          ),
          Row(
            children: [
              Text(
                global.language('report_dedebi_showing_items'),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
              const SizedBox(width: 8),
              DropdownButton<int>(
                value: _pageSize,
                items: [10, 20, 50, 100].map((size) {
                  return DropdownMenuItem<int>(
                    value: size,
                    child: Text('$size'),
                  );
                }).toList(),
                onChanged: (newSize) {
                  if (newSize != null) {
                    setState(() => _pageSize = newSize);
                    _loadPage(jobId, 1, newSize);
                  }
                },
                underline: Container(),
                style: TextStyle(fontSize: 14, color: global.theme.textColor),
              ),
              SizedBox(width: 8),
              Text(
                global.language('report_dedebi_items'),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
            ],
          ),
          SizedBox(width: 16),
          Row(
            children: [
              IconButton(
                onPressed: currentPage > 1
                    ? () => _loadPage(jobId, 1, _pageSize)
                    : null,
                icon: Icon(Icons.first_page),
                tooltip: global.language('report_dedebi_first_page'),
              ),
              IconButton(
                onPressed: currentPage > 1
                    ? () => _loadPage(jobId, currentPage - 1, _pageSize)
                    : null,
                icon: Icon(Icons.chevron_left),
                tooltip: global.language('report_dedebi_previous_page'),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
                decoration: BoxDecoration(
                  color: global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  '$currentPage / $totalPages',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              IconButton(
                onPressed: currentPage < totalPages
                    ? () => _loadPage(jobId, currentPage + 1, _pageSize)
                    : null,
                icon: Icon(Icons.chevron_right),
                tooltip: global.language('report_dedebi_next_page'),
              ),
              IconButton(
                onPressed: currentPage < totalPages
                    ? () => _loadPage(jobId, totalPages, _pageSize)
                    : null,
                icon: Icon(Icons.last_page),
                tooltip: global.language('report_dedebi_last_page'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  void _loadPage(String jobId, int page, int size) {
    setState(() => _pageSize = size);

    context.read<BiReportBloc>().add(
      GetBiReportDetailRequested(
        reportType: BiReportType.vatBuy,
        jobId: jobId,
        token: global.appConfig.getString("token")!,
        page: page,
        size: size,
      ),
    );
  }

  Widget _buildErrorState(String errorMessage) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.error_outline, size: 64, color: global.theme.negativeHighlightTextColor),
          SizedBox(height: 16),
          Text(
            global.language('report_dedebi_error_occurred'),
            style: TextStyle(
              fontSize: 18,
              color: global.theme.negativeHighlightTextColor,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(height: 8),
          Container(
            constraints: const BoxConstraints(maxWidth: 400),
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              errorMessage,
              style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              textAlign: TextAlign.center,
            ),
          ),
          SizedBox(height: 32),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              ElevatedButton.icon(
                onPressed: () {
                  if (_biReportBloc != null) {
                    _biReportBloc!.add(const ResetBiReportState());
                  }
                  _refreshReportData();
                },
                icon: Icon(Icons.refresh),
                label: Text(global.language('report_dedebi_try_again')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.indigo.shade600,
                  foregroundColor: global.theme.onPrimaryColor,
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 16,
                  ),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  elevation: 2,
                ),
              ),
              SizedBox(width: 16),
              OutlinedButton.icon(
                onPressed: _showConditionDialog,
                icon: Icon(Icons.settings),
                label: Text(global.language('report_dedebi_change_conditions')),
                style: OutlinedButton.styleFrom(
                  foregroundColor: Colors.indigo.shade600,
                  side: BorderSide(color: Colors.indigo.shade600),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 16,
                  ),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
            ],
          ),
          SizedBox(height: 16),
          TextButton.icon(
            onPressed: () {
              if (_biReportBloc != null) {
                _biReportBloc!.add(const ResetBiReportState());
              }
            },
            icon: Icon(Icons.clear_all),
            label: Text(global.language('report_dedebi_clear_data')),
            style: TextButton.styleFrom(foregroundColor: global.theme.textSecondaryColor),
          ),
        ],
      ),
    );
  }

  Future<void> _showConditionDialog() async {
    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return ReportConditionDialog(
          reportType: BiReportType.vatBuy,
          initialFromDate: _fromDate,
          initialToDate: _toDate,
          initialSelectedBranches: _selectedBranches,
          onConditionsSet: (conditions) {
            setState(() {
              _fromDate = conditions.fromDate;
              _toDate = conditions.toDate;
              _selectedBranches = conditions.selectedBranches!;
            });
            _refreshReportData();
          },
        );
      },
    );
  }

  Future<void> _showExportDialog() async {
    if (_currentJobId.isEmpty) {
      _showErrorSnackBar(global.language('report_dedebi_search_before_export'));
      return;
    }

    await showDialog(
      context: context,
      builder: (context) => ExportReportDialog(
        reportType: BiReportType.vatBuy,
        jobId: _currentJobId,
      ),
    );
  }

  void _refreshReportData() {
    if (_fromDate == null || _toDate == null) {
      global.showErrorSnackBar(context, global.language('report_dedebi_select_date_range'));
      return;
    }

    if (_fromDate!.isAfter(_toDate!)) {
      global.showErrorSnackBar(context, global.language('report_dedebi_from_date_not_exceed_to_date'));
      return;
    }

    final daysDifference = _toDate!.difference(_fromDate!).inDays;
    if (daysDifference > 365) {
      global.showWarningSnackBar(context, global.language('report_dedebi_date_range_max_1_year'));
      return;
    }

    final token = global.appConfig.getString("token");
    if (token == null || token.isEmpty) {
      global.showErrorSnackBar(context, global.language('report_dedebi_token_not_found'));
      return;
    }

    context.read<BiReportBloc>().add(const ResetBiReportState());

    setState(() {
      _currentTotalSummary = null;
      _totalSummaryLoading = false;
    });

    AppLogger.debug('🔍 VAT Buy Report - Creating conditions:');
    AppLogger.debug('   fromDate: $_fromDate');
    AppLogger.debug('   toDate: $_toDate');

    final conditions = ReportConditionsModel(
      fromdate: DateFormat('yyyy-MM-dd').format(_fromDate!),
      todate: DateFormat('yyyy-MM-dd').format(_toDate!),
      branchcode: _selectedBranches.getBranchCodeString(),
    );

    AppLogger.debug('📤 Conditions created:');
    AppLogger.debug('   fromdate: ${conditions.fromdate}');
    AppLogger.debug('   todate: ${conditions.todate}');
    AppLogger.debug('   branchcode: ${conditions.branchcode}');

    global.showInfoSnackBar(context, global.language('report_dedebi_preparing_report_data'));

    context.read<BiReportBloc>().add(
      GenerateBiReportRequested(
        reportType: BiReportType.vatBuy,
        conditions: conditions,
        token: token,
        pollInterval: _pollInterval,
        timeout: _reportTimeout,
        page: 1,
        size: _pageSize,
      ),
    );
  }
}
