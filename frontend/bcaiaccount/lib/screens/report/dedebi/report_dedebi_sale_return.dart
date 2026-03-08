import 'package:flutter/foundation.dart';
// ignore_for_file: deprecated_member_use, unused_import

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/sale_return_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_branch_search_screen.dart';
import 'package:smlaicloud/screens/report/dedebi/components/filter_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/report_condition_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/sale_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/total_summary_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/transaction_detail_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/export_report_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ReportDedebiSaleReturnScreen extends StatefulWidget {
  const ReportDedebiSaleReturnScreen({super.key});

  @override
  State<ReportDedebiSaleReturnScreen> createState() =>
      _ReportDedebiSaleReturnScreenState();
}

class _ReportDedebiSaleReturnScreenState
    extends State<ReportDedebiSaleReturnScreen> {
  // Report polling configuration
  static const Duration _pollInterval = Duration(
    seconds: 2,
  ); // เช็ค status ทุก 2 วินาที
  static const Duration _reportTimeout = Duration(
    minutes: 10,
  ); // Timeout 10 นาที

  // Report condition variables
  DateTime? _fromDate;
  DateTime? _toDate;
  bool _showDetails = false;

  BranchSelectionModel _selectedBranches = const BranchSelectionModel(
    selectedBranches: [],
    isCancel: false,
  ); // Branch selection

  // Pagination variables
  int _pageSize = 20;

  BiReportMeta? _currentMeta;
  List<SaleReturnModel> _currentData = [];
  String _currentJobId = ''; // Add this to track current job ID

  // Total summary data
  SaleReturnSummaryModel? _currentTotalSummary;
  bool _totalSummaryLoading = false;

  // Add bloc reference
  BiReportBloc? _biReportBloc;

  // Filter panel visibility state
  bool _isFilterPanelVisible = true;

  @override
  void initState() {
    if (kDebugMode) {
      AppLogger.debug('🚀 ReportDedebiSaleReturnScreen initState called');
    }
    super.initState();

    // Initialize default dates (current month)
    final now = DateTime.now();
    _fromDate = DateTime(now.year, now.month, 1); // First day of current month
    _toDate = DateTime(now.year, now.month + 1, 0); // Last day of current month
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Store bloc reference safely
    _biReportBloc = context.read<BiReportBloc>();

    // Reset report state when entering screen (only once)
    if (mounted) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _biReportBloc != null) {
          if (kDebugMode) {
            AppLogger.debug('📱 PostFrameCallback executed');
          }
          // Reset BiReport state first
          _biReportBloc!.add(const ResetBiReportState());

          // Then show condition dialog
          _showConditionDialog();
        }
      });
    }
  }

  @override
  void dispose() {
    // Use stored bloc reference instead of context.read
    if (_biReportBloc != null) {
      _biReportBloc!.add(const ResetBiReportState());
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      onPopInvoked: (didPop) {
        // Use stored bloc reference
        if (_biReportBloc != null) {
          _biReportBloc!.add(const ResetBiReportState());
        }
      },
      child: Scaffold(
        backgroundColor: Colors.grey.shade50,
        appBar: AppBar(
          title: Text(
            global.language('report_dedebi_sale_return_title'),
            style: const TextStyle(fontWeight: FontWeight.w600),
          ),
          centerTitle: true,
          backgroundColor: Colors.indigo.shade600,
          foregroundColor: Colors.white,
          elevation: 2,
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () {
              if (kDebugMode) {
                AppLogger.debug('🔙 Back button pressed');
                AppLogger.debug('🔙 Current job ID: $_currentJobId');
              }

              // Cancel ongoing report generation if any
              if (_biReportBloc != null && _currentJobId.isNotEmpty) {
                if (kDebugMode) {
                  AppLogger.debug('🛑 Sending cancel request for job: $_currentJobId');
                }
                _biReportBloc!.add(
                  CancelBiReportRequested(jobId: _currentJobId),
                );
              }

              // Use stored bloc reference
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
              onPressed: () {
                setState(() {
                  _isFilterPanelVisible = !_isFilterPanelVisible;
                });
              },
              tooltip: _isFilterPanelVisible
                  ? global.language('report_dedebi_hide_filter_panel')
                  : global.language('report_dedebi_show_filter_panel'),
            ),
            IconButton(
              icon: const Icon(Icons.filter_alt_outlined),
              onPressed: _showConditionDialog,
              tooltip: global.language('report_dedebi_set_search_conditions'),
            ),
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () {
                // Reset and refresh data
                context.read<BiReportBloc>().add(const ResetBiReportState());
                _refreshReportData();
              },
              tooltip: global.language('report_dedebi_refresh_data'),
            ),
            IconButton(
              icon: const Icon(Icons.download),
              onPressed: _showExportDialog,
              tooltip: global.language('report_dedebi_export_report'),
            ),
          ],
        ),
        body: BlocConsumer<BiReportBloc, BiReportState>(
          // ลดความซับซ้อนของ buildWhen
          buildWhen: (previous, current) {
            // เฉพาะ states ที่ไม่ใช่ summary และ total summary ให้ rebuild UI หลัก
            return current is! BiReportSummaryLoading &&
                current is! BiReportSummarySuccess &&
                current is! BiReportSummaryFailure;
          },
          listener: (context, state) {
            _handleStateChanges(state);
          },

          builder: (context, state) {
            return _buildContent(state);
          },
        ),
      ),
    );
  }

  void _handleStateChanges(BiReportState state) {
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
        setState(() {
          _totalSummaryLoading = true;
        });
        break;
      case BiReportSummarySuccess():
        setState(() {
          _currentTotalSummary = state.saleReturnSummaryData;
          _totalSummaryLoading = false;
        });
        break;
      case BiReportSummaryFailure():
        setState(() {
          _totalSummaryLoading = false;
        });
        _showErrorSnackBar(state.message);
        break;

      default:
        // No action needed for other states
        break;
    }
  }

  Widget _buildContent(BiReportState state) {
    return switch (state) {
      BiReportInitial() => _buildInitialState(),
      BiReportGenerating() => _buildLoadingState(global.language('report_dedebi_generating_report'), 0),
      BiReportGenerateProgress() => _buildLoadingState(
        state.statusMessage,
        state.progress,
      ),
      BiReportGenerateSuccess() => _buildSuccessContent(
        state.data.whereType<SaleReturnModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportDetailLoading() => _buildContentWithLoading(state.jobId),
      BiReportDetailSuccess() => _buildSuccessContent(
        state.data.whereType<SaleReturnModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportGenerateFailure() => _buildErrorState(state.message),
      BiReportDetailFailure() => _buildErrorState(state.message),

      // Default fallback
      _ =>
        _currentData.isNotEmpty
            ? _buildSuccessContent(_currentData, _currentJobId, _currentMeta!)
            : _buildInitialState(),
    };
  }

  void _updateCurrentData(dynamic state) {
    if (state is BiReportGenerateSuccess || state is BiReportDetailSuccess) {
      setState(() {
        _currentData = state.data.whereType<SaleReturnModel>().toList();
        _currentMeta = state.meta;
        _currentJobId = state.jobId;
      });
    }
  }

  void _requestTotalSummary(String jobId) {
    context.read<BiReportBloc>().add(
      GetBiReportSummaryRequested(
        reportType: BiReportType.saleReturn,
        jobId: jobId,
        token: global.appConfig.getString("token")!,
      ),
    );
  }

  Widget _buildContentWithLoading(String jobId) {
    if (_currentData.isNotEmpty && _currentMeta != null) {
      return _buildMainLayoutWithLoading(_currentData, jobId, _currentMeta!);
    }
    return _buildLoadingState(global.language('report_dedebi_loading_data'), 0);
  }

  Widget _buildSuccessContent(
    List<SaleReturnModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return _buildMainLayout(data, jobId, meta);
  }

  Widget _buildMainLayout(
    List<SaleReturnModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Main content
        Expanded(child: _buildMainContent(data, jobId, meta)),

        // Sidebar - conditionally show based on _isFilterPanelVisible
        if (_isFilterPanelVisible) _buildSidebar(data),
      ],
    );
  }

  Widget _buildMainLayoutWithLoading(
    List<SaleReturnModel> data,
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
        // Sidebar - conditionally show based on _isFilterPanelVisible
        if (_isFilterPanelVisible) _buildSidebar(data),
      ],
    );
  }

  Widget _buildSidebar(List<SaleReturnModel> data) {
    return Container(
      width: 260,
      padding: const EdgeInsets.only(top: 16, bottom: 16, left: 0, right: 16),
      child: SingleChildScrollView(
        // เพิ่ม scroll ลงได้
        child: Column(
          children: [
            FilterPanel(
              reportType: BiReportType.saleReturn,
              fromDate: _fromDate,
              toDate: _toDate,
              showDetails: _showDetails,
              selectedBranches: _selectedBranches,
              onShowConditionDialog: _showConditionDialog,
              onRefresh: _refreshReportData,
            ),
            const SizedBox(height: 16),
            TotalSummaryPanel(
              saleReturnSummary: _currentTotalSummary,
              isLoading: _totalSummaryLoading,
              reportType: BiReportType.saleReturn,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMainContent(
    List<SaleReturnModel> data,
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

  Widget _buildDataGrid(List<SaleReturnModel> data) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: SaleTableView(
        dataSaleReturn: data,
        formatCurrency: ReportUtils.formatCurrency,
        getCreditorName: ReportUtils.getCreditorName,
        onRowSaleReturnTap: _showTransactionDialog,
        reportType: BiReportType.saleReturn,
      ),
    );
  }

  Widget _buildLoadingOverlay() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.8),
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
                color: Colors.grey.shade700,
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
              Icons.analytics_outlined,
              size: 64,
              color: Colors.indigo.shade600,
            ),
          ),
          SizedBox(height: 32),
          Text(
            global.language('report_dedebi_sale_return_dedebi_title'),
            style: TextStyle(
              fontSize: 24,
              color: Colors.grey.shade800,
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
                color: Colors.grey.shade600,
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
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              elevation: 3,
              textStyle: const TextStyle(
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
                    value: null, // ❌ เปลี่ยนให้เป็น null เพื่อให้หมุนตลอด
                    strokeWidth: 6,
                    backgroundColor: Colors.grey.shade300,
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
              color: Colors.grey.shade700,
              fontWeight: FontWeight.w500,
            ),
          ),
          SizedBox(height: 8),
          Text(
            global.language('report_dedebi_please_wait'),
            style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
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
        color: Colors.white,
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(16),
          bottomRight: Radius.circular(16),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Page info
          Expanded(
            child: Text(
              '${global.language('report_dedebi_showing_items')} $startItem-$endItem ${global.language('report_dedebi_all').toLowerCase()} $totalItems ${global.language('report_dedebi_items')}',
              style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
            ),
          ),

          // Page size selector
          Row(
            children: [
              Text(
                global.language('report_dedebi_showing_items'),
                style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
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
                    setState(() {
                      _pageSize = newSize;
                    });
                    _loadPage(jobId, 1, newSize);
                  }
                },
                underline: Container(),
                style: TextStyle(fontSize: 14, color: Colors.grey.shade800),
              ),
              SizedBox(width: 8),
              Text(
                global.language('report_dedebi_items'),
                style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
              ),
            ],
          ),

          const SizedBox(width: 16),

          // Pagination buttons
          Row(
            children: [
              // First page
              IconButton(
                onPressed: currentPage > 1
                    ? () => _loadPage(jobId, 1, _pageSize)
                    : null,
                icon: Icon(Icons.first_page),
                tooltip: global.language('report_dedebi_first_page'),
              ),

              // Previous page
              IconButton(
                onPressed: currentPage > 1
                    ? () => _loadPage(jobId, currentPage - 1, _pageSize)
                    : null,
                icon: Icon(Icons.chevron_left),
                tooltip: global.language('report_dedebi_previous_page'),
              ),

              // Page numbers
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
                decoration: BoxDecoration(
                  color: Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  '$currentPage / $totalPages',
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),

              // Next page
              IconButton(
                onPressed: currentPage < totalPages
                    ? () => _loadPage(jobId, currentPage + 1, _pageSize)
                    : null,
                icon: Icon(Icons.chevron_right),
                tooltip: global.language('report_dedebi_next_page'),
              ),

              // Last page
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
    setState(() {
      _pageSize = size;
    });

    context.read<BiReportBloc>().add(
      GetBiReportDetailRequested(
        reportType: BiReportType.saleReturn,
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
          Icon(Icons.error_outline, size: 64, color: Colors.red.shade400),
          SizedBox(height: 16),
          Text(
            global.language('report_dedebi_error_occurred'),
            style: TextStyle(
              fontSize: 18,
              color: Colors.red.shade600,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(height: 8),
          Container(
            constraints: const BoxConstraints(maxWidth: 400),
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              errorMessage,
              style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
              textAlign: TextAlign.center,
            ),
          ),
          SizedBox(height: 32),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              ElevatedButton.icon(
                onPressed: () {
                  // Reset state and try again with current conditions
                  context.read<BiReportBloc>().add(const ResetBiReportState());
                  _refreshReportData();
                },
                icon: Icon(Icons.refresh),
                label: Text(global.language('report_dedebi_try_again')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.indigo.shade600,
                  foregroundColor: Colors.white,
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
              context.read<BiReportBloc>().add(const ResetBiReportState());
            },
            icon: Icon(Icons.clear_all),
            label: Text(global.language('report_dedebi_sale_return_clear_data')),
            style: TextButton.styleFrom(foregroundColor: Colors.grey.shade600),
          ),
        ],
      ),
    );
  }

  Future<void> _showConditionDialog() async {
    if (kDebugMode) {
      AppLogger.debug('🔵 _showConditionDialog called');
    }

    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return ReportConditionDialog(
          reportType: BiReportType.saleReturn,
          initialFromDate: _fromDate,
          initialToDate: _toDate,
          initialShowDetails: _showDetails,
          initialSelectedBranches: _selectedBranches,
          onConditionsSet: (conditions) {
            if (kDebugMode) {
              AppLogger.debug('🔍 Conditions set from dialog');
            }
            setState(() {
              _fromDate = conditions.fromDate;
              _toDate = conditions.toDate;
              _showDetails = conditions.showDetails!;
              _selectedBranches = conditions.selectedBranches!;
            });
            if (kDebugMode) {
              AppLogger.debug('💾 State updated in main screen');
              AppLogger.debug('🚪 Calling _refreshReportData');
            }
            _refreshReportData();
          },
        );
      },
    );
  }

  Future<void> _showExportDialog() async {
    if (_currentJobId.isEmpty) {
      _showErrorSnackBar(global.language('report_dedebi_sale_return_please_search_before_export'));
      return;
    }

    await showDialog(
      context: context,
      builder: (context) => ExportReportDialog(
        reportType: BiReportType.saleReturn,
        jobId: _currentJobId,
      ),
    );
  }

  void _refreshReportData() {
    if (kDebugMode) {
      AppLogger.debug('🟢 _refreshReportData called');
    }

    // Validate date inputs
    if (_fromDate == null || _toDate == null) {
      if (kDebugMode) {
        AppLogger.error('❌ Date validation failed: fromDate=$_fromDate, toDate=$_toDate');
      }
      global.showErrorSnackBar(context, global.language('report_dedebi_please_select_date_range'));
      return;
    }

    if (kDebugMode) {
      AppLogger.debug(
        '✅ Date validation passed: ${DateFormat('yyyy-MM-dd').format(_fromDate!)} to ${DateFormat('yyyy-MM-dd').format(_toDate!)}',
      );
    }

    // Validate date range
    if (_fromDate!.isAfter(_toDate!)) {
      if (kDebugMode) {
        AppLogger.error('❌ Date range validation failed: fromDate > toDate');
      }
      global.showErrorSnackBar(context, global.language('report_dedebi_sale_return_date_start_must_not_exceed_end'));
      return;
    }

    // Check if date range is too large (more than 1 year)
    final daysDifference = _toDate!.difference(_fromDate!).inDays;
    if (daysDifference > 365) {
      if (kDebugMode) {
        AppLogger.debug('⚠️ Date range too large: $daysDifference days');
      }
      global.showWarningSnackBar(context, global.language('report_dedebi_sale_return_date_range_exceeds_1_year'));
      return;
    }

    // Validate token
    final token = global.appConfig.getString("token");
    if (token == null || token.isEmpty) {
      if (kDebugMode) {
        AppLogger.error('❌ Token validation failed: token is null or empty');
      }
      global.showErrorSnackBar(context, global.language('report_dedebi_sale_return_token_not_found'));
      return;
    }

    // Reset current state before new request
    if (kDebugMode) {
      AppLogger.debug('🔄 Resetting BiReport state');
    }
    context.read<BiReportBloc>().add(const ResetBiReportState());

    // Reset total summary state
    setState(() {
      _currentTotalSummary = null;
      _totalSummaryLoading = false;
    });

    // Create conditions for Sale Report
    final conditions = ReportConditionsModel(
      fromdate: DateFormat('yyyy-MM-dd').format(_fromDate!),
      todate: DateFormat('yyyy-MM-dd').format(_toDate!),
      branchcode: _selectedBranches.getBranchCodeString(),
      showdetail: _showDetails,
    );

    if (kDebugMode) {
      AppLogger.debug('📋 Report conditions created: ${conditions.toJson()}');
    }

    // Show loading indicator and trigger report generation
    global.showInfoSnackBar(context, global.language('report_dedebi_sale_return_preparing_report_data'));

    // Trigger report generation via BiReportBloc
    final startTime = DateTime.now();
    if (kDebugMode) {
      AppLogger.debug('🚀 Dispatching GenerateBiReportRequested event with:');
      AppLogger.debug('   - reportType: ${BiReportType.saleReturn}');
      AppLogger.debug('   - pollInterval: ${_pollInterval.inSeconds} seconds');
      AppLogger.debug('   - timeout: ${_reportTimeout.inMinutes} minutes');
    }
    if (kDebugMode) {
      AppLogger.debug('   - conditions: ${conditions.toJson()}');
    }
    if (kDebugMode) {
      AppLogger.debug('   - token: ${token.substring(0, 10)}...');
    }
    if (kDebugMode) {
      AppLogger.debug('   - page: 1, size: $_pageSize');
      AppLogger.debug('   - startTime: $startTime');
    }

    try {
      context.read<BiReportBloc>().add(
        GenerateBiReportRequested(
          reportType: BiReportType.saleReturn,
          conditions: conditions,
          token: token,
          pollInterval: _pollInterval, // ใช้ค่า constant
          timeout: _reportTimeout, // ใช้ค่า constant
          page: 1, // Start with first page
          size: _pageSize, // Use current page size
        ),
      );

      if (kDebugMode) {
        AppLogger.info('✅ Event dispatched successfully at ${DateTime.now()}');
      }
      if (kDebugMode) {
        AppLogger.debug(
          '⏱️ Dispatch took: ${DateTime.now().difference(startTime).inMilliseconds}ms',
        );
      }
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.error('❌ Error dispatching event: $e');
        AppLogger.debug('📚 Stack trace: $stackTrace');
      }

      global.showErrorSnackBar(context, '${global.language('report_dedebi_sale_return_error_sending_request')}: $e');
    }
  }

  void _showTransactionDialog(SaleReturnModel sale) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return TransactionDetailDialog(
          docno: sale.docno,
          jobId: _currentJobId,
          reportType: BiReportType.saleReturn,
          formatCurrency: ReportUtils.formatCurrency,
        );
      },
    );
  }
}
