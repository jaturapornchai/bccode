import 'package:flutter/foundation.dart';
// ignore_for_file: deprecated_member_use, unused_import

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/employee/employee_bloc.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:smlaicloud/repositories/employee_repository.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_model.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_summary_model.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/model/bi_report/entity_selection_model.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_branch_search_screen.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_entity_search_screen.dart';
import 'package:smlaicloud/screens/report/dedebi/components/filter_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/report_condition_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/sale_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/total_summary_panel.dart';
import 'package:smlaicloud/screens/report/dedebi/components/transaction_detail_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/components/export_report_dialog.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// รายงานรับสินค้าแบบทยอยรับ Dedebi ที่เชื่อมต่อกับ BiReportBloc
/// รองรับการแสดงความคืบหน้าและข้อมูลแบบเรียลไทม์
class ReportDedebiPurchasePartialScreen extends StatefulWidget {
  const ReportDedebiPurchasePartialScreen({super.key});

  @override
  State<ReportDedebiPurchasePartialScreen> createState() =>
      _ReportDedebiPurchasePartialScreenState();
}

class _ReportDedebiPurchasePartialScreenState
    extends State<ReportDedebiPurchasePartialScreen>
    with global.ThemeRefreshMixin {
  // Report polling configuration
  static const Duration _pollInterval = Duration(
    seconds: 2,
  ); // เช็ค status ทุก 2 วินาที
  static const Duration _reportTimeout = Duration(
    minutes: 10,
  ); // Timeout 10 นาทีt

  // Report condition variables
  DateTime? _fromDate;
  DateTime? _toDate;
  bool _showDetails = false;

  // Additional filter conditions
  String _showCancelledDocuments = global.language('hide_cancelled_documents');
  final String _saleType = global.language('all'); // ทั้งหมด (default)
  final String _posType = global.language('all'); // ทั้งหมด (default)
  BranchSelectionModel _selectedBranches = const BranchSelectionModel(
    selectedBranches: [],
    isCancel: false,
  ); // Branch selection
  EntitySelectionModel _selectedDebtors = const EntitySelectionModel(
    selectedEntities: [],
    isCancel: false,
  ); // Creditor selection
  final EntitySelectionModel _selectedSalespersons = const EntitySelectionModel(
    selectedEntities: [],
    isCancel: false,
  ); // Employee selection

  // Pagination variables
  int _pageSize = 20;

  // Flags to prevent infinite updates

  BiReportMeta? _currentMeta;
  List<PurchasePartialModel> _currentData = [];
  String _currentJobId = ''; // Add this to track current job ID

  // Total summary data - For purchase partial, we might not have specific summary
  PurchasePartialSummaryModel? _currentTotalSummary;
  bool _totalSummaryLoading = false;

  // Add bloc reference
  BiReportBloc? _biReportBloc;

  // Filter panel visibility state
  bool _isFilterPanelVisible = true;

  // Expandable cards state - Removed since we use dialog now
  // Set<String> _expandedCards = <String>{};

  // Summary data - Removed (now handled by SummaryPanelNew)
  // SaleReportSummary? _currentSummary;
  // bool _summaryRequested = false;

  @override
  void initState() {
    if (kDebugMode) {
      AppLogger.debug('🚀 ReportDedebiPurchasePartialScreen initState called');
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
        backgroundColor: global.theme.backgroundColor,
        appBar: AppBar(
          title: Text(
            global.language('purchase_partial_report_title'),
            style: TextStyle(fontWeight: FontWeight.w600),
          ),
          centerTitle: true,
          backgroundColor: Colors.indigo.shade600,
          foregroundColor: global.theme.onPrimaryColor,
          elevation: 2,
          leading: IconButton(
            icon: Icon(Icons.arrow_back),
            onPressed: () {
              if (kDebugMode) {
                AppLogger.debug('🔙 Back button pressed');
                AppLogger.debug('🔙 Current job ID: $_currentJobId');
              }

              // Cancel ongoing report generation if any
              if (_biReportBloc != null && _currentJobId.isNotEmpty) {
                if (kDebugMode) {
                  AppLogger.debug(
                    '🛑 Sending cancel request for job: $_currentJobId',
                  );
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
              icon: Icon(Icons.filter_alt_outlined),
              onPressed: _showConditionDialog,
              tooltip: global.language('set_search_conditions'),
            ),
            IconButton(
              icon: Icon(Icons.refresh),
              onPressed: () {
                // Reset and refresh data
                if (_biReportBloc != null) {
                  _biReportBloc!.add(const ResetBiReportState());
                }
                _refreshReportData();
              },
              tooltip: global.language('refresh_data'),
            ),
            IconButton(
              icon: Icon(Icons.download),
              onPressed: _showExportDialog,
              tooltip: global.language('export_report'),
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
        AppLogger.info(
          "🎯 BiReportSummarySuccess received for ${state.reportType}",
        );
        AppLogger.debug(
          "📊 purchasePartialSummaryData: ${state.purchasePartialSummaryData}",
        );
        setState(() {
          _currentTotalSummary = state.purchasePartialSummaryData;
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
      BiReportGenerating() => _buildLoadingState(
        global.language('generating_report'),
        0,
      ),
      BiReportGenerateProgress() => () {
        // Update current job ID for cancellation
        _currentJobId = state.jobId;
        return _buildLoadingState(state.statusMessage, state.progress);
      }(),
      BiReportGenerateSuccess() => _buildSuccessContent(
        state.data.whereType<PurchasePartialModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportDetailLoading() => _buildContentWithLoading(state.jobId),
      BiReportDetailSuccess() => _buildSuccessContent(
        state.data.whereType<PurchasePartialModel>().toList(),
        state.jobId,
        state.meta,
      ),
      BiReportGenerateFailure() => () {
        // Clear job ID on failure
        _currentJobId = '';
        return _buildErrorState(state.message);
      }(),
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
        _currentData = state.data.whereType<PurchasePartialModel>().toList();
        _currentMeta = state.meta;
        _currentJobId = state.jobId;
      });
    }
  }

  void _requestTotalSummary(String jobId) {
    context.read<BiReportBloc>().add(
      GetBiReportSummaryRequested(
        reportType: BiReportType.purchasepartial,
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
    List<PurchasePartialModel> data,
    String jobId,
    BiReportMeta meta,
  ) {
    return _buildMainLayout(data, jobId, meta);
  }

  Widget _buildMainLayout(
    List<PurchasePartialModel> data,
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
    List<PurchasePartialModel> data,
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

  Widget _buildSidebar(List<PurchasePartialModel> data) {
    return Container(
      width: 260,
      padding: const EdgeInsets.only(top: 16, bottom: 16, left: 0, right: 16),
      child: SingleChildScrollView(
        // เพิ่ม scroll ลงได้
        child: Column(
          children: [
            FilterPanel(
              reportType: BiReportType.purchasepartial,
              fromDate: _fromDate,
              toDate: _toDate,
              showDetails: _showDetails,
              showCancelledDocuments: _showCancelledDocuments,
              saleType: _saleType,
              posType: _posType,
              selectedBranches: _selectedBranches,
              selectedDebtors: _selectedDebtors,
              selectedSalespersons: _selectedSalespersons,
              onShowConditionDialog: _showConditionDialog,
              onRefresh: _refreshReportData,
            ),
            const SizedBox(height: 16),
            TotalSummaryPanel(
              purchasePartialSummary: _currentTotalSummary,
              isLoading: _totalSummaryLoading,
              reportType: BiReportType.purchasepartial,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMainContent(
    List<PurchasePartialModel> data,
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

  Widget _buildDataGrid(List<PurchasePartialModel> data) {
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
      child: SaleTableView(
        dataPurchasePartial: data,
        formatCurrency: ReportUtils.formatCurrency,
        getCreditorName: ReportUtils.getCreditorName,
        onRowPurchasePartialTap: _showTransactionDialog,
        reportType: BiReportType.purchasepartial,
      ),
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
              global.language('loading_new_page'),
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
              Icons.analytics_outlined,
              size: 64,
              color: Colors.indigo.shade600,
            ),
          ),
          SizedBox(height: 32),
          Text(
            global.language('purchase_partial_report_dedebi_title'),
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
              global.language('set_search_conditions_to_view_data'),
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
            label: Text(global.language('set_search_conditions')),
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
                    value: null, // ❌ เปลี่ยนให้เป็น null เพื่อให้หมุนตลอด
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
            global.language('please_wait'),
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
          // Page info
          Expanded(
            child: Text(
              'แสดง $startItem-$endItem จากทั้งหมด $totalItems รายการ',
              style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
            ),
          ),

          // Page size selector
          Row(
            children: [
              Text(
                global.language('showing_items'),
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
                    setState(() {
                      _pageSize = newSize;
                    });
                    _loadPage(jobId, 1, newSize);
                  }
                },
                underline: Container(),
                style: TextStyle(fontSize: 14, color: global.theme.textColor),
              ),
              SizedBox(width: 8),
              Text(
                global.language('items_label'),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
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
        reportType: BiReportType.purchasepartial,
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
            global.language('error_occurred'),
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
                  // Reset state and try again with current conditions
                  if (_biReportBloc != null) {
                    _biReportBloc!.add(const ResetBiReportState());
                  }
                  _refreshReportData();
                },
                icon: Icon(Icons.refresh),
                label: Text(global.language('try_again')),
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
              if (_biReportBloc != null) {
                _biReportBloc!.add(const ResetBiReportState());
              }
            },
            icon: Icon(Icons.clear_all),
            label: Text(global.language('clear_data')),
            style: TextButton.styleFrom(foregroundColor: global.theme.textSecondaryColor),
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
          reportType: BiReportType.purchasepartial,
          initialFromDate: _fromDate,
          initialToDate: _toDate,
          initialShowDetails: _showDetails,
          initialShowCancelled: _showCancelledDocuments,
          initialSelectedBranches: _selectedBranches,
          initialSelectedCreditors: _selectedDebtors,
          onConditionsSet: (conditions) {
            if (kDebugMode) {
              AppLogger.debug('🔍 Conditions set from dialog');
            }
            setState(() {
              _fromDate = conditions.fromDate;
              _toDate = conditions.toDate;
              _showDetails = conditions.showDetails!;
              _showCancelledDocuments = conditions.showCancelled!;
              _selectedBranches = conditions.selectedBranches!;
              _selectedDebtors = conditions.selectedCreditors!;
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
      _showErrorSnackBar(global.language('please_search_data_before_export'));
      return;
    }

    await showDialog(
      context: context,
      builder: (context) => ExportReportDialog(
        reportType: BiReportType.purchasepartial,
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
        AppLogger.error(
          '❌ Date validation failed: fromDate=$_fromDate, toDate=$_toDate',
        );
      }
      global.showErrorSnackBar(context, global.language('please_select_date_range'));
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
      global.showErrorSnackBar(context, global.language('start_date_must_not_exceed_end_date'));
      return;
    }

    // Check if date range is too large (more than 1 year)
    final daysDifference = _toDate!.difference(_fromDate!).inDays;
    if (daysDifference > 365) {
      if (kDebugMode) {
        AppLogger.debug('⚠️ Date range too large: $daysDifference days');
      }
      global.showWarningSnackBar(context, global.language('date_range_must_not_exceed_1_year'));
      return;
    }

    // Validate token
    final token = global.appConfig.getString("token");
    if (token == null || token.isEmpty) {
      if (kDebugMode) {
        AppLogger.error('❌ Token validation failed: token is null or empty');
      }
      global.showErrorSnackBar(context, global.language('token_not_found_please_login_again'));
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
      iscancel:
          (_showCancelledDocuments == global.language('hide_cancelled_documents') ||
              _showCancelledDocuments == '')
          ? 'false'
          : '',
      inquirytype: (_saleType == global.language('all') || _saleType == '')
          ? ''
          : (_saleType == global.language('transaction_sale') ? '1' : '2'),
      ispos: _posType == global.language('pos_only'),
      creditorcode: _selectedDebtors
          .getEntityCodeString(), // ใช้ _selectedDebtors แทน
      salecode: _selectedSalespersons.getEntityCodeString(),
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
      AppLogger.debug('   - reportType: ${BiReportType.purchasepartial}');
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
          reportType: BiReportType.purchasepartial,
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

      global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการส่งคำขอ: $e');
    }
  }

  void _showTransactionDialog(PurchasePartialModel purchasePartial) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return TransactionDetailDialog(
          docno: purchasePartial.docno,
          jobId: _currentJobId,
          reportType: BiReportType.purchasepartial,
          formatCurrency: ReportUtils.formatCurrency,
        );
      },
    );
  }
}
