import 'package:flutter/foundation.dart';
// reportStockBalance
// ignore_for_file: deprecated_member_use, unused_import

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/employee/employee_bloc.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_model.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_summary.dart';
import 'package:smlaicloud/model/bi_report/stock_balance_model.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:smlaicloud/repositories/employee_repository.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/bi_sale_report_data.dart';
import 'package:smlaicloud/model/bi_report/sale_report_summary.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/model/bi_report/entity_selection_model.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_branch_search_screen.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_entity_search_screen.dart';
import 'package:smlaicloud/screens/report/dedebi/components/filter_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/report_condition_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/sale_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/total_summary_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/transaction_daily_detail_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/export_report_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ReportStockBalanceScreen extends StatefulWidget {
  const ReportStockBalanceScreen({super.key});

  @override
  State<ReportStockBalanceScreen> createState() =>
      _ReportStockBalanceScreenState();
}

class _ReportStockBalanceScreenState extends State<ReportStockBalanceScreen> {
  // Report polling configuration
  static const Duration _pollInterval = Duration(
    seconds: 2,
  ); // เช็ค status ทุก 2 วินาที
  static const Duration _reportTimeout = Duration(
    minutes: 10,
  ); // Timeout 10 นาที

  DateTime? _toDate;

  EntitySelectionModel _selectedBarcodes = const EntitySelectionModel(
    selectedEntities: [],
    isCancel: false,
  );

  int _pageSize = 20;

  BiReportMeta? _currentMeta;
  List<StockBalanceModel> _currentData = [];
  String _currentJobId = '';

  StockBalanceSummaryModel? _currentTotalSummary;
  bool _totalSummaryLoading = false;

  // Add bloc reference
  BiReportBloc? _biReportBloc;

  // Filter panel visibility state
  bool _isFilterPanelVisible = true;

  @override
  void initState() {
    super.initState();

    _toDate = DateTime.now();
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
            _selectedBarcodes.selectedEntities.isNotEmpty
                ? '${global.language('stock_balance_report')} : ${_selectedBarcodes.selectedEntities.first.code}~${ReportUtils.getDisplayNameSafe(_selectedBarcodes.selectedEntities.first.names)}'
                : global.language('stock_balance_report'),
            style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 18),
            overflow: TextOverflow.ellipsis,
            maxLines: 1,
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
                  ? global.language('hide_filter_panel')
                  : global.language('show_filter_panel'),
            ),
            IconButton(
              icon: const Icon(Icons.filter_alt_outlined),
              onPressed: _showConditionDialog,
              tooltip: global.language('set_search_conditions'),
            ),
            // Export button - only show when data is available
            if (_currentJobId.isNotEmpty && _currentData.isNotEmpty)
              IconButton(
                icon: const Icon(Icons.download),
                onPressed: _showExportDialog,
                tooltip: global.language('export_report'),
              ),
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () {
                // Reset and refresh data
                context.read<BiReportBloc>().add(const ResetBiReportState());
                _refreshReportData();
              },
              tooltip: global.language('refresh_data'),
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
          if (state.stockBalanceSummaryData != null) {
            _currentTotalSummary = state.stockBalanceSummaryData!;
          } else {
            if (kDebugMode) {
              AppLogger.debug(
                '⚠️ Warning: stockMovementSummaryData is null in BiReportSummarySuccess',
              );
            }
          }
          _totalSummaryLoading = false;
        });
        break;
      case BiReportSummaryFailure():
        setState(() {
          _totalSummaryLoading = false;
          _currentTotalSummary = null; // เซ็ต null เมื่อ failure
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
      BiReportGenerating() => _buildLoadingState(global.language('generating_report_start'), 0),
      BiReportGenerateProgress() => _buildLoadingState(
        state.statusMessage,
        state.progress,
      ),
      BiReportGenerateSuccess() => _buildSuccessContent(
        state.data.whereType<StockBalanceModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportDetailLoading() => _buildContentWithLoading(state.jobId),
      BiReportDetailSuccess() => _buildSuccessContent(
        state.data.whereType<StockBalanceModel>().toList(),
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
        _currentData = state.data.whereType<StockBalanceModel>().toList();
        _currentMeta = state.meta;
        _currentJobId = state.jobId;
      });
    }
  }

  void _requestTotalSummary(String jobId) {
    context.read<BiReportBloc>().add(
      GetBiReportSummaryRequested(
        reportType: BiReportType.stockBalance,
        jobId: jobId,
        token: global.appConfig.getString("token")!,
      ),
    );
  }

  Widget _buildContentWithLoading(String jobId) {
    if (_currentData.isNotEmpty && _currentMeta != null) {
      return _buildMainLayoutWithLoading(_currentData, jobId, _currentMeta!);
    }
    return _buildLoadingState(global.language('loading_data'), 0);
  }

  Widget _buildSuccessContent(
    List<StockBalanceModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return _buildMainLayout(data, jobId, meta);
  }

  Widget _buildMainLayout(
    List<StockBalanceModel> data,
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
    List<StockBalanceModel> data,
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

  Widget _buildSidebar(List<StockBalanceModel> data) {
    return Container(
      width: 260,
      padding: const EdgeInsets.only(top: 16, bottom: 16, left: 0, right: 16),
      child: Column(
        children: [
          FilterPanel(
            reportType: BiReportType.stockBalance,
            toDate: _toDate,
            onShowConditionDialog: _showConditionDialog,
            onRefresh: _refreshReportData,
            selectedBarcodes: _selectedBarcodes,
          ),
          const SizedBox(height: 16),
          TotalSummaryPanel(
            stockBalanceSummary: _currentTotalSummary,
            isLoading: _totalSummaryLoading,
            reportType: BiReportType.stockBalance,
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Widget _buildMainContent(
    List<StockBalanceModel> data,
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

  Widget _buildDataGrid(List<StockBalanceModel> data) {
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
        reportType: BiReportType.stockBalance,
        dataStockBalance: data,
        formatCurrency: ReportUtils.formatCurrency,
        getCreditorName: ReportUtils.getCreditorName,
        onRowSaleDailyTap: null,
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
              global.language('loading_new_page'),
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
            global.language('dedebi_sales_report'),
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
              global.language('set_conditions_to_view_data'),
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
            label: Text(global.language('set_search_conditions')),
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
            global.language('please_wait'),
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
              '${global.language('showing_items')} $startItem-$endItem ${global.language('from_total_items')} $totalItems ${global.language('items')}',
              style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
            ),
          ),

          // Page size selector
          Row(
            children: [
              Text(
                global.language('showing_items'),
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
                global.language('items'),
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
                tooltip: global.language('first_page'),
              ),

              // Previous page
              IconButton(
                onPressed: currentPage > 1
                    ? () => _loadPage(jobId, currentPage - 1, _pageSize)
                    : null,
                icon: Icon(Icons.chevron_left),
                tooltip: global.language('previous_page'),
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
                tooltip: global.language('next_page'),
              ),

              // Last page
              IconButton(
                onPressed: currentPage < totalPages
                    ? () => _loadPage(jobId, totalPages, _pageSize)
                    : null,
                icon: Icon(Icons.last_page),
                tooltip: global.language('last_page'),
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
        reportType: BiReportType.stockBalance,
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
            global.language('error_occurred'),
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
                label: Text(global.language('try_again')),
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
                label: Text(global.language('change_conditions')),
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
            label: Text(global.language('clear_data')),
            style: TextButton.styleFrom(foregroundColor: Colors.grey.shade600),
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
          reportType: BiReportType.stockBalance,
          initialToDate: _toDate,
          initialSelectedBarcodes: _selectedBarcodes,
          onConditionsSet: (conditions) {
            setState(() {
              _toDate = conditions.toDate;
              _selectedBarcodes = conditions.selectedBarcodes!;
            });
            _refreshReportData();
          },
        );
      },
    );
  }

  Future<void> _showExportDialog() async {
    if (_currentJobId.isEmpty) {
      global.showWarningSnackBar(context, global.language('no_report_data_for_export'));
      return;
    }

    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return ExportReportDialog(
          reportType: BiReportType.stockBalance,
          jobId: _currentJobId,
        );
      },
    );
  }

  void _refreshReportData() {
    if (kDebugMode) {
      AppLogger.debug('🟢 _refreshReportData called');
    }

    // Validate token
    final token = global.appConfig.getString("token");
    if (token == null || token.isEmpty) {
      if (kDebugMode) {
        AppLogger.error('❌ Token validation failed: token is null or empty');
      }
      global.showErrorSnackBar(context, global.language('no_token_please_login_again'));
      return;
    }

    // Reset current state before new request
    if (kDebugMode) {
      AppLogger.debug('🔄 Resetting BiReport state');
    }
    context.read<BiReportBloc>().add(const ResetBiReportState());

    // Reset total summary state - สำคัญมาก!
    setState(() {
      _currentTotalSummary = null; // เคลียร์ก่อน
      _totalSummaryLoading = false;
      _currentData = []; // เคลียร์ data เก่าด้วย
      _currentMeta = null;
      _currentJobId = '';
    });

    // Create conditions for Sale Daily Report
    final conditions = ReportConditionsModel(
      todate: DateFormat('yyyy-MM-dd').format(_toDate!),
      barcode: _selectedBarcodes.selectedEntities.isNotEmpty
          ? (_selectedBarcodes.selectedEntities.first.code.toString())
          : '',
    );

    if (kDebugMode) {
      AppLogger.debug('📋 Report conditions created: ${conditions.toJson()}');
    }

    // Show loading indicator and trigger report generation
    global.showInfoSnackBar(context, global.language('preparing_report_data'));

    // Trigger report generation via BiReportBloc
    final startTime = DateTime.now();
    if (kDebugMode) {
      AppLogger.debug('🚀 Dispatching GenerateBiReportRequested event with:');
      AppLogger.debug('   - reportType: ${BiReportType.stockBalance}');
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
          reportType: BiReportType.stockBalance,
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

      global.showErrorSnackBar(context, '${global.language('error_in_request')}: $e');
    }
  }
}
