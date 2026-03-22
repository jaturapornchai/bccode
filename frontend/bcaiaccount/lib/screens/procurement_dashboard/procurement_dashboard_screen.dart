import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:syncfusion_flutter_charts/charts.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/date_picker.dart';
import 'cubit/procurement_dashboard_cubit.dart';

/// Procurement Dashboard Screen — interactive version with filters, tabs, charts
class ProcurementDashboardScreen extends StatefulWidget {
  const ProcurementDashboardScreen({super.key});

  @override
  State<ProcurementDashboardScreen> createState() =>
      _ProcurementDashboardScreenState();
}

class _ProcurementDashboardScreenState
    extends State<ProcurementDashboardScreen>
    with global.ThemeRefreshMixin, SingleTickerProviderStateMixin {
  late ProcurementDashboardCubit _cubit;
  late TabController _tabController;

  // Filter state
  bool _showFilter = false;
  DateTime? _dateFrom;
  DateTime? _dateTo;
  int _selectedPreset = 2; // 0=7d, 1=30d, 2=3m, 3=6m, 4=1y, 5=custom
  List<int> _selectedDocTypes = [21, 22, 6];
  List<String> _selectedStatuses = [];

  // Chart toggle state
  int _chartTypeIndex = 0; // 0=bar, 1=line, 2=area
  int _pieGroupIndex = 0; // 0=by status, 1=by doc type

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
    _cubit = ProcurementDashboardCubit();
    _applyPresetDate(2);
    _cubit.loadDashboard(
      dateFrom: _dateFrom,
      dateTo: _dateTo,
      docTypes: _selectedDocTypes,
      statuses: _selectedStatuses,
    );
  }

  @override
  void dispose() {
    _tabController.dispose();
    _cubit.close();
    super.dispose();
  }

  void _applyPresetDate(int index) {
    final now = DateTime.now();
    _selectedPreset = index;
    switch (index) {
      case 0:
        _dateFrom = now.subtract(const Duration(days: 7));
        _dateTo = now;
        break;
      case 1:
        _dateFrom = now.subtract(const Duration(days: 30));
        _dateTo = now;
        break;
      case 2:
        _dateFrom = now.subtract(const Duration(days: 90));
        _dateTo = now;
        break;
      case 3:
        _dateFrom = now.subtract(const Duration(days: 180));
        _dateTo = now;
        break;
      case 4:
        _dateFrom = now.subtract(const Duration(days: 365));
        _dateTo = now;
        break;
      case 5:
        // custom — keep existing dates
        break;
    }
  }

  void _toggleFilter() {
    setState(() => _showFilter = !_showFilter);
  }

  void _clearFilters() {
    setState(() {
      _selectedPreset = 2;
      _selectedDocTypes = [21, 22, 6];
      _selectedStatuses = [];
      _applyPresetDate(2);
    });
  }

  void _applyFilters() {
    _cubit.loadDashboard(
      dateFrom: _dateFrom,
      dateTo: _dateTo,
      docTypes: _selectedDocTypes,
      statuses: _selectedStatuses,
    );
  }

  void _refresh() {
    _cubit.refresh(
      dateFrom: _dateFrom,
      dateTo: _dateTo,
      docTypes: _selectedDocTypes,
      statuses: _selectedStatuses,
    );
  }

  // --- Number formatting ---

  String _formatNumber(dynamic value) {
    final n = int.tryParse(value.toString()) ?? 0;
    return NumberFormat('#,##0').format(n);
  }

  String _formatAmount(double value) {
    if (value >= 1000000) {
      return '${NumberFormat('#,##0.0').format(value / 1000000)}M';
    }
    if (value >= 1000) {
      return '${NumberFormat('#,##0.0').format(value / 1000)}K';
    }
    return NumberFormat('#,##0.00').format(value);
  }

  String _formatAmountFull(double value) {
    return NumberFormat('#,##0.00').format(value);
  }

  // --- Trend calculation ---

  double _calcTrend(dynamic current, dynamic previous) {
    final cur = (current as num?)?.toDouble() ?? 0;
    final prev = (previous as num?)?.toDouble() ?? 0;
    if (prev == 0) return 0;
    return ((cur - prev) / prev * 100);
  }

  // --- Status helpers ---

  String _statusLabel(String status) {
    switch (status) {
      case 'draft':
        return global.language('draft');
      case 'pending':
        return global.language('pending_approval');
      case 'approved':
        return global.language('approved');
      case 'rejected':
        return global.language('rejected');
      default:
        return status.isNotEmpty ? status : global.language('draft');
    }
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'draft':
        return global.theme.iconSecondaryColor;
      case 'pending':
        return global.theme.warningHighlightTextColor;
      case 'approved':
        return global.theme.positiveHighlightTextColor;
      case 'rejected':
        return global.theme.negativeHighlightTextColor;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  String _docTypeLabel(int flag) {
    switch (flag) {
      case 21:
        return 'PR';
      case 22:
        return 'RFQ';
      case 6:
        return 'PO';
      default:
        return flag.toString();
    }
  }

  Color _docTypeColor(int flag) {
    switch (flag) {
      case 21:
        return global.theme.infoHighlightTextColor;
      case 22:
        return global.theme.warningHighlightTextColor;
      case 6:
        return global.theme.positiveHighlightTextColor;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider.value(
      value: _cubit,
      child: Scaffold(
        backgroundColor: global.theme.backgroundColor,
        appBar: AppBar(
          title: Text(
            global.language('procurement_dashboard'),
            style: TextStyle(color: global.theme.onPrimaryColor),
          ),
          backgroundColor: global.theme.appBarColor,
          iconTheme: IconThemeData(color: global.theme.onPrimaryColor),
          actions: [
            IconButton(
              icon: Icon(
                _showFilter ? Icons.filter_list_off : Icons.filter_list,
                color: global.theme.onPrimaryColor,
              ),
              tooltip: global.language('search'),
              onPressed: _toggleFilter,
            ),
            IconButton(
              icon: Icon(Icons.refresh, color: global.theme.onPrimaryColor),
              tooltip: global.language('refresh'),
              onPressed: _refresh,
            ),
          ],
        ),
        body: Column(
          children: [
            // Collapsible filter bar
            _buildFilterBar(),
            // Main content
            Expanded(
              child: BlocBuilder<ProcurementDashboardCubit,
                  ProcurementDashboardState>(
                builder: (context, state) {
                  if (state is ProcurementDashboardLoading) {
                    return const Center(child: CircularProgressIndicator());
                  }
                  if (state is ProcurementDashboardError) {
                    return _buildErrorView(state);
                  }
                  if (state is ProcurementDashboardLoaded) {
                    return _buildDashboard(context, state);
                  }
                  return const SizedBox.shrink();
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===================== Filter Bar =====================

  Widget _buildFilterBar() {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOut,
      height: _showFilter ? null : 0,
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(
          bottom: BorderSide(
            color: _showFilter
                ? global.theme.dividerBorderColor
                : Colors.transparent,
          ),
        ),
      ),
      child: _showFilter ? _buildFilterContent() : const SizedBox.shrink(),
    );
  }

  Widget _buildFilterContent() {
    final presetLabels = [
      global.language('days_7'),
      global.language('days_30'),
      global.language('months_3'),
      global.language('months_6'),
      global.language('year_1'),
      global.language('custom_date'),
    ];

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Date preset chips
          Text(
            global.language('date_from'),
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 6),
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: List.generate(presetLabels.length, (i) {
              return ChoiceChip(
                label: Text(
                  presetLabels[i],
                  style: TextStyle(
                    fontSize: 12,
                    color: _selectedPreset == i
                        ? global.theme.onPrimaryColor
                        : global.theme.textColor,
                  ),
                ),
                selected: _selectedPreset == i,
                selectedColor: global.theme.primaryColor,
                backgroundColor: global.theme.surfaceColor,
                side: BorderSide(color: global.theme.dividerBorderColor),
                onSelected: (selected) {
                  if (selected) {
                    setState(() => _applyPresetDate(i));
                  }
                },
              );
            }),
          ),

          // Custom date pickers
          if (_selectedPreset == 5) ...[
            const SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  child: CustomDatePicker(
                    labelText: global.language('date_from'),
                    initialDate: _dateFrom,
                    firstDate: DateTime(2000),
                    lastDate: DateTime(2100),
                    useIconSelectDate: true,
                    onDateSelected: (date) {
                      setState(() => _dateFrom = date);
                    },
                    decoration: InputDecoration(
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 10,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(
                          color: global.theme.dividerBorderColor,
                        ),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(
                          color: global.theme.dividerBorderColor,
                        ),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: CustomDatePicker(
                    labelText: global.language('date_to'),
                    initialDate: _dateTo,
                    firstDate: DateTime(2000),
                    lastDate: DateTime(2100),
                    useIconSelectDate: true,
                    onDateSelected: (date) {
                      setState(() => _dateTo = date);
                    },
                    decoration: InputDecoration(
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 10,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(
                          color: global.theme.dividerBorderColor,
                        ),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(
                          color: global.theme.dividerBorderColor,
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],

          const SizedBox(height: 12),

          // Doc type filter chips
          Text(
            global.language('document_type'),
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 6),
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              _buildDocTypeFilterChip(21, 'PR'),
              _buildDocTypeFilterChip(22, 'RFQ'),
              _buildDocTypeFilterChip(6, 'PO'),
            ],
          ),

          const SizedBox(height: 12),

          // Status filter chips
          Text(
            global.language('status'),
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 6),
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              _buildStatusFilterChip('draft'),
              _buildStatusFilterChip('pending'),
              _buildStatusFilterChip('approved'),
              _buildStatusFilterChip('rejected'),
            ],
          ),

          const SizedBox(height: 12),

          // Apply + Clear buttons
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              TextButton(
                onPressed: () {
                  _clearFilters();
                  _applyFilters();
                },
                child: Text(
                  global.language('clear'),
                  style: TextStyle(color: global.theme.textSecondaryColor),
                ),
              ),
              const SizedBox(width: 8),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.primaryColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                onPressed: _applyFilters,
                child: Text(global.language('search')),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDocTypeFilterChip(int flag, String label) {
    final selected = _selectedDocTypes.contains(flag);
    return FilterChip(
      label: Text(
        label,
        style: TextStyle(
          fontSize: 12,
          color: selected
              ? global.theme.onPrimaryColor
              : global.theme.textColor,
        ),
      ),
      selected: selected,
      selectedColor: _docTypeColor(flag),
      backgroundColor: global.theme.surfaceColor,
      checkmarkColor: global.theme.onPrimaryColor,
      side: BorderSide(color: global.theme.dividerBorderColor),
      onSelected: (val) {
        setState(() {
          if (val) {
            _selectedDocTypes.add(flag);
          } else {
            _selectedDocTypes.remove(flag);
          }
        });
      },
    );
  }

  Widget _buildStatusFilterChip(String status) {
    final selected = _selectedStatuses.contains(status);
    return FilterChip(
      label: Text(
        _statusLabel(status),
        style: TextStyle(
          fontSize: 12,
          color: selected
              ? global.theme.onPrimaryColor
              : global.theme.textColor,
        ),
      ),
      selected: selected,
      selectedColor: _statusColor(status),
      backgroundColor: global.theme.surfaceColor,
      checkmarkColor: global.theme.onPrimaryColor,
      side: BorderSide(color: global.theme.dividerBorderColor),
      onSelected: (val) {
        setState(() {
          if (val) {
            _selectedStatuses.add(status);
          } else {
            _selectedStatuses.remove(status);
          }
        });
      },
    );
  }

  // ===================== Error View =====================

  Widget _buildErrorView(ProcurementDashboardError state) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.error_outline,
              size: 48, color: global.theme.negativeHighlightTextColor),
          const SizedBox(height: 8),
          Text(
            state.message,
            style: TextStyle(color: global.theme.textColor),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          TextButton(
            onPressed: _refresh,
            child: Text(global.language('try_again')),
          ),
        ],
      ),
    );
  }

  // ===================== Dashboard Layout =====================

  Widget _buildDashboard(
      BuildContext context, ProcurementDashboardLoaded state) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 800;
        return Column(
          children: [
            // KPI Cards (scrollable area at top)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
              child: _buildKpiSection(state, isWide),
            ),
            const SizedBox(height: 12),

            // TabBar
            Container(
              decoration: BoxDecoration(
                color: global.theme.cardColor,
                border: Border(
                  bottom: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: TabBar(
                controller: _tabController,
                labelColor: global.theme.primaryColor,
                unselectedLabelColor: global.theme.textSecondaryColor,
                indicatorColor: global.theme.primaryColor,
                tabs: [
                  Tab(text: global.language('overview')),
                  Tab(text: global.language('trend')),
                  Tab(text: global.language('vendor')),
                  Tab(text: global.language('product')),
                ],
              ),
            ),

            // Tab content
            Expanded(
              child: TabBarView(
                controller: _tabController,
                children: [
                  _buildOverviewTab(state, isWide),
                  _buildTrendTab(state),
                  _buildVendorTab(state),
                  _buildProductTab(state, isWide),
                ],
              ),
            ),
          ],
        );
      },
    );
  }

  // ===================== KPI Cards =====================

  Widget _buildKpiSection(ProcurementDashboardLoaded state, bool isWide) {
    final prData = state.kpiData['pr'] ?? {};
    final rfqData = state.kpiData['rfq'] ?? {};
    final poData = state.kpiData['po'] ?? {};
    final prevPr = state.prevKpiData['pr'] ?? {};
    final prevRfq = state.prevKpiData['rfq'] ?? {};
    final prevPo = state.prevKpiData['po'] ?? {};

    final prCount = prData['doc_count'] ?? 0;
    final rfqCount = rfqData['doc_count'] ?? 0;
    final poCount = poData['doc_count'] ?? 0;

    final totalAmount = ((prData['total_amount'] ?? 0.0) as double) +
        ((rfqData['total_amount'] ?? 0.0) as double) +
        ((poData['total_amount'] ?? 0.0) as double);
    final prevTotalAmount = ((prevPr['total_amount'] ?? 0.0) as double) +
        ((prevRfq['total_amount'] ?? 0.0) as double) +
        ((prevPo['total_amount'] ?? 0.0) as double);

    final cards = [
      _buildKpiCard(
        icon: Icons.description_outlined,
        count: prCount,
        label: global.language('purchase_requisition'),
        color: global.theme.infoHighlightTextColor,
        trendPercent: _calcTrend(prCount, prevPr['doc_count']),
        pendingCount: state.pendingCount['pr'] ?? 0,
        onTap: () => _showDocListSheet('PR', 21),
      ),
      _buildKpiCard(
        icon: Icons.request_quote_outlined,
        count: rfqCount,
        label: 'RFQ',
        color: global.theme.warningHighlightTextColor,
        trendPercent: _calcTrend(rfqCount, prevRfq['doc_count']),
        pendingCount: state.pendingCount['rfq'] ?? 0,
        onTap: () => _showDocListSheet('RFQ', 22),
      ),
      _buildKpiCard(
        icon: Icons.shopping_cart_outlined,
        count: poCount,
        label: global.language('purchase_order'),
        color: global.theme.positiveHighlightTextColor,
        trendPercent: _calcTrend(poCount, prevPo['doc_count']),
        pendingCount: state.pendingCount['po'] ?? 0,
        onTap: () => _showDocListSheet('PO', 6),
      ),
      _buildKpiCard(
        icon: Icons.attach_money,
        count: totalAmount,
        label: global.language('total_procurement_amount'),
        color: global.theme.primaryColor,
        trendPercent: _calcTrend(totalAmount, prevTotalAmount),
        pendingCount: 0,
        isAmount: true,
        onTap: () {},
      ),
    ];

    if (isWide) {
      return Row(
        children: cards
            .map((card) => Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: card,
                  ),
                ))
            .toList(),
      );
    }

    return Column(
      children: [
        Row(children: [
          Expanded(
              child: Padding(
                  padding: const EdgeInsets.only(right: 4), child: cards[0])),
          Expanded(
              child: Padding(
                  padding: const EdgeInsets.only(left: 4), child: cards[1])),
        ]),
        const SizedBox(height: 8),
        Row(children: [
          Expanded(
              child: Padding(
                  padding: const EdgeInsets.only(right: 4), child: cards[2])),
          Expanded(
              child: Padding(
                  padding: const EdgeInsets.only(left: 4), child: cards[3])),
        ]),
      ],
    );
  }

  Widget _buildKpiCard({
    required IconData icon,
    required dynamic count,
    required String label,
    required Color color,
    required double trendPercent,
    required int pendingCount,
    required VoidCallback onTap,
    bool isAmount = false,
  }) {
    final displayCount =
        isAmount ? _formatAmount(count as double) : _formatNumber(count);
    final trendUp = trendPercent > 0;
    final hasTrend = trendPercent != 0;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          // Gradient background — subtle tint of accent color
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              color.withValues(alpha: 0.12),
              global.theme.cardColor,
            ],
          ),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: color.withValues(alpha: 0.2),
            width: 1,
          ),
          // Shadow for depth
          boxShadow: [
            BoxShadow(
              color: color.withValues(alpha: 0.08),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Stack(
          children: [
            Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Row: icon circle + trend pill
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    // Icon in gradient circle
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          colors: [
                            color.withValues(alpha: 0.2),
                            color.withValues(alpha: 0.05),
                          ],
                        ),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Icon(icon, color: color, size: 22),
                    ),
                    // Trend pill
                    if (hasTrend)
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 3),
                        decoration: BoxDecoration(
                          color: trendUp
                              ? global.theme.positiveHighlightTextColor
                                  .withValues(alpha: 0.12)
                              : global.theme.negativeHighlightTextColor
                                  .withValues(alpha: 0.12),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              trendUp
                                  ? Icons.trending_up
                                  : Icons.trending_down,
                              size: 12,
                              color: trendUp
                                  ? global.theme.positiveHighlightTextColor
                                  : global.theme.negativeHighlightTextColor,
                            ),
                            const SizedBox(width: 2),
                            Text(
                              '${trendPercent.abs().toStringAsFixed(1)}%',
                              style: TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.w600,
                                color: trendUp
                                    ? global.theme.positiveHighlightTextColor
                                    : global.theme.negativeHighlightTextColor,
                              ),
                            ),
                          ],
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 12),
                // Count — bold large
                Text(
                  displayCount,
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.w800,
                    color: global.theme.textColor,
                    letterSpacing: -0.5,
                  ),
                ),
                const SizedBox(height: 4),
                // Label
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 12,
                    color: global.theme.textSecondaryColor,
                    fontWeight: FontWeight.w500,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
            // Pending badge with glow
            if (pendingCount > 0)
              Positioned(
                top: 0,
                right: 0,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
                  decoration: BoxDecoration(
                    color: global.theme.negativeHighlightTextColor,
                    borderRadius: BorderRadius.circular(10),
                    boxShadow: [
                      BoxShadow(
                        color: global.theme.negativeHighlightTextColor
                            .withValues(alpha: 0.4),
                        blurRadius: 6,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Text(
                    '$pendingCount',
                    style: TextStyle(
                      color: global.theme.onPrimaryColor,
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  // ===================== Tab 1: Overview =====================

  Widget _buildOverviewTab(ProcurementDashboardLoaded state, bool isWide) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: isWide
          ? Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(child: _buildMonthlyChartCard(state)),
                const SizedBox(width: 16),
                Expanded(child: _buildStatusChartCard(state)),
              ],
            )
          : Column(
              children: [
                _buildMonthlyChartCard(state),
                const SizedBox(height: 16),
                _buildStatusChartCard(state),
              ],
            ),
    );
  }

  Widget _buildMonthlyChartCard(ProcurementDashboardLoaded state) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: global.theme.dividerBorderColor.withValues(alpha: 0.3)),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
            child: Row(
              children: [
                Container(
                  width: 4,
                  height: 20,
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    global.language('monthly_procurement_trend'),
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                      color: global.theme.textColor,
                      letterSpacing: -0.3,
                    ),
                  ),
                ),
                _buildChartTypeToggle(),
              ],
            ),
          ),
          const SizedBox(height: 4),
          SizedBox(
            height: 300,
            child: state.monthlyData.isEmpty
                ? _buildEmptyChart()
                : _buildMonthlyChart(state),
          ),
        ],
      ),
    );
  }

  Widget _buildChartTypeToggle() {
    final icons = [Icons.bar_chart, Icons.show_chart, Icons.area_chart];
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: List.generate(3, (i) {
        final selected = _chartTypeIndex == i;
        return Padding(
          padding: const EdgeInsets.only(left: 2),
          child: InkWell(
            borderRadius: BorderRadius.circular(6),
            onTap: () => setState(() => _chartTypeIndex = i),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: selected
                    ? global.theme.primaryColor.withValues(alpha: 0.15)
                    : Colors.transparent,
                borderRadius: BorderRadius.circular(6),
              ),
              child: Icon(
                icons[i],
                size: 18,
                color: selected
                    ? global.theme.primaryColor
                    : global.theme.iconSecondaryColor,
              ),
            ),
          ),
        );
      }),
    );
  }

  Widget _buildMonthlyChart(ProcurementDashboardLoaded state) {
    final prData = <_ChartData>[];
    final rfqData = <_ChartData>[];
    final poData = <_ChartData>[];

    final months = <String>{};
    for (var row in state.monthlyData) {
      months.add(row['month'] as String);
    }

    for (var month in months.toList()..sort()) {
      final prRow = state.monthlyData
          .where((r) => r['month'] == month && r['transflag'] == 21);
      final rfqRow = state.monthlyData
          .where((r) => r['month'] == month && r['transflag'] == 22);
      final poRow = state.monthlyData
          .where((r) => r['month'] == month && r['transflag'] == 6);

      prData.add(_ChartData(month,
          prRow.isNotEmpty ? (prRow.first['total_amount'] as double) : 0));
      rfqData.add(_ChartData(month,
          rfqRow.isNotEmpty ? (rfqRow.first['total_amount'] as double) : 0));
      poData.add(_ChartData(month,
          poRow.isNotEmpty ? (poRow.first['total_amount'] as double) : 0));
    }

    return Padding(
      padding: const EdgeInsets.all(8),
      child: SfCartesianChart(
        plotAreaBorderWidth: 0,
        legend: Legend(
          isVisible: true,
          position: LegendPosition.bottom,
          textStyle: TextStyle(color: global.theme.textSecondaryColor),
        ),
        primaryXAxis: CategoryAxis(
          labelStyle: TextStyle(color: global.theme.textSecondaryColor),
          majorGridLines: const MajorGridLines(width: 0),
        ),
        primaryYAxis: NumericAxis(
          labelStyle: TextStyle(color: global.theme.textSecondaryColor),
          numberFormat: NumberFormat.compact(),
          majorGridLines: MajorGridLines(
            width: 0.5,
            color: global.theme.dividerBorderColor,
          ),
        ),
        tooltipBehavior: TooltipBehavior(enable: true),
        series: _buildMonthlySeries(prData, rfqData, poData),
      ),
    );
  }

  List<CartesianSeries> _buildMonthlySeries(
    List<_ChartData> prData,
    List<_ChartData> rfqData,
    List<_ChartData> poData,
  ) {
    switch (_chartTypeIndex) {
      case 1: // Line
        return [
          LineSeries<_ChartData, String>(
            name: 'PR',
            dataSource: prData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.infoHighlightTextColor,
            markerSettings: const MarkerSettings(isVisible: true, height: 4, width: 4),
          ),
          LineSeries<_ChartData, String>(
            name: 'RFQ',
            dataSource: rfqData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.warningHighlightTextColor,
            markerSettings: const MarkerSettings(isVisible: true, height: 4, width: 4),
          ),
          LineSeries<_ChartData, String>(
            name: 'PO',
            dataSource: poData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.positiveHighlightTextColor,
            markerSettings: const MarkerSettings(isVisible: true, height: 4, width: 4),
          ),
        ];
      case 2: // Area
        return [
          AreaSeries<_ChartData, String>(
            name: 'PR',
            dataSource: prData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.infoHighlightTextColor.withValues(alpha: 0.4),
            borderColor: global.theme.infoHighlightTextColor,
            borderWidth: 2,
          ),
          AreaSeries<_ChartData, String>(
            name: 'RFQ',
            dataSource: rfqData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.warningHighlightTextColor.withValues(alpha: 0.4),
            borderColor: global.theme.warningHighlightTextColor,
            borderWidth: 2,
          ),
          AreaSeries<_ChartData, String>(
            name: 'PO',
            dataSource: poData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.positiveHighlightTextColor.withValues(alpha: 0.4),
            borderColor: global.theme.positiveHighlightTextColor,
            borderWidth: 2,
          ),
        ];
      default: // Bar (Stacked Column)
        return [
          StackedColumnSeries<_ChartData, String>(
            name: 'PR',
            dataSource: prData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.infoHighlightTextColor,
            borderRadius: const BorderRadius.only(
              topLeft: Radius.circular(2),
              topRight: Radius.circular(2),
            ),
          ),
          StackedColumnSeries<_ChartData, String>(
            name: 'RFQ',
            dataSource: rfqData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.warningHighlightTextColor,
          ),
          StackedColumnSeries<_ChartData, String>(
            name: 'PO',
            dataSource: poData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.positiveHighlightTextColor,
          ),
        ];
    }
  }

  // --- Status doughnut chart ---

  Widget _buildStatusChartCard(ProcurementDashboardLoaded state) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: global.theme.dividerBorderColor.withValues(alpha: 0.3)),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
            child: Row(
              children: [
                Container(
                  width: 4,
                  height: 20,
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    global.language('document_status_summary'),
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                      color: global.theme.textColor,
                      letterSpacing: -0.3,
                    ),
                  ),
                ),
                _buildPieGroupToggle(),
              ],
            ),
          ),
          const SizedBox(height: 4),
          SizedBox(
            height: 300,
            child: state.statusData.isEmpty
                ? _buildEmptyChart()
                : _buildStatusChart(state),
          ),
        ],
      ),
    );
  }

  Widget _buildPieGroupToggle() {
    final labels = [global.language('status'), global.language('document_type')];
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: List.generate(2, (i) {
        final selected = _pieGroupIndex == i;
        return Padding(
          padding: const EdgeInsets.only(left: 4),
          child: InkWell(
            borderRadius: BorderRadius.circular(6),
            onTap: () => setState(() => _pieGroupIndex = i),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: selected
                    ? global.theme.primaryColor.withValues(alpha: 0.15)
                    : Colors.transparent,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(
                  color: selected
                      ? global.theme.primaryColor
                      : global.theme.dividerBorderColor,
                ),
              ),
              child: Text(
                labels[i],
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: selected ? FontWeight.w600 : FontWeight.normal,
                  color: selected
                      ? global.theme.primaryColor
                      : global.theme.textSecondaryColor,
                ),
              ),
            ),
          ),
        );
      }),
    );
  }

  Widget _buildStatusChart(ProcurementDashboardLoaded state) {
    List<_PieChartData> chartData;

    if (_pieGroupIndex == 0) {
      // Group by status
      final statusMap = <String, int>{};
      for (var row in state.statusData) {
        final status = row['approval_status'] as String;
        final count = row['doc_count'] as int;
        final key = status.isEmpty ? 'draft' : status;
        statusMap[key] = (statusMap[key] ?? 0) + count;
      }
      chartData = statusMap.entries
          .map((e) => _PieChartData(_statusLabel(e.key), e.value, _statusColor(e.key)))
          .toList();
    } else {
      // Group by doc type
      final docMap = <int, int>{};
      for (var row in state.statusData) {
        final flag = row['transflag'] as int;
        final count = row['doc_count'] as int;
        docMap[flag] = (docMap[flag] ?? 0) + count;
      }
      chartData = docMap.entries
          .map((e) => _PieChartData(_docTypeLabel(e.key), e.value, _docTypeColor(e.key)))
          .toList();
    }

    return Padding(
      padding: const EdgeInsets.all(8),
      child: SfCircularChart(
        legend: Legend(
          isVisible: true,
          position: LegendPosition.bottom,
          textStyle: TextStyle(color: global.theme.textSecondaryColor),
          overflowMode: LegendItemOverflowMode.wrap,
        ),
        tooltipBehavior: TooltipBehavior(enable: true),
        series: <CircularSeries>[
          DoughnutSeries<_PieChartData, String>(
            dataSource: chartData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            pointColorMapper: (d, _) => d.color,
            dataLabelSettings: DataLabelSettings(
              isVisible: true,
              textStyle: TextStyle(
                color: global.theme.textColor,
                fontSize: 11,
              ),
            ),
            innerRadius: '55%',
          ),
        ],
      ),
    );
  }

  // ===================== Tab 2: Trend (MoM Comparison) =====================

  Widget _buildTrendTab(ProcurementDashboardLoaded state) {
    final prData = state.kpiData['pr'] ?? {};
    final rfqData = state.kpiData['rfq'] ?? {};
    final poData = state.kpiData['po'] ?? {};
    final prevPr = state.prevKpiData['pr'] ?? {};
    final prevRfq = state.prevKpiData['rfq'] ?? {};
    final prevPo = state.prevKpiData['po'] ?? {};

    final curDocCount = ((prData['doc_count'] ?? 0) as int) +
        ((rfqData['doc_count'] ?? 0) as int) +
        ((poData['doc_count'] ?? 0) as int);
    final prevDocCount = ((prevPr['doc_count'] ?? 0) as int) +
        ((prevRfq['doc_count'] ?? 0) as int) +
        ((prevPo['doc_count'] ?? 0) as int);

    final curAmount = ((prData['total_amount'] ?? 0.0) as double) +
        ((rfqData['total_amount'] ?? 0.0) as double) +
        ((poData['total_amount'] ?? 0.0) as double);
    final prevAmount = ((prevPr['total_amount'] ?? 0.0) as double) +
        ((prevRfq['total_amount'] ?? 0.0) as double) +
        ((prevPo['total_amount'] ?? 0.0) as double);

    final docTrend = _calcTrend(curDocCount, prevDocCount);
    final amtTrend = _calcTrend(curAmount, prevAmount);

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // MoM comparison card
          _buildComparisonCard(
            title: global.language('document_count'),
            currentValue: _formatNumber(curDocCount),
            previousValue: _formatNumber(prevDocCount),
            trendPercent: docTrend,
          ),
          const SizedBox(height: 16),
          _buildComparisonCard(
            title: global.language('total_amount'),
            currentValue: _formatAmountFull(curAmount),
            previousValue: _formatAmountFull(prevAmount),
            trendPercent: amtTrend,
          ),
          const SizedBox(height: 24),

          // Per-type breakdown
          Text(
            global.language('document_type'),
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: global.theme.textColor,
            ),
          ),
          const SizedBox(height: 12),
          _buildTrendBreakdownRow('PR', prData, prevPr, global.theme.infoHighlightTextColor),
          const SizedBox(height: 8),
          _buildTrendBreakdownRow('RFQ', rfqData, prevRfq, global.theme.warningHighlightTextColor),
          const SizedBox(height: 8),
          _buildTrendBreakdownRow('PO', poData, prevPo, global.theme.positiveHighlightTextColor),
        ],
      ),
    );
  }

  Widget _buildComparisonCard({
    required String title,
    required String currentValue,
    required String previousValue,
    required double trendPercent,
  }) {
    final trendUp = trendPercent > 0;
    final hasTrend = trendPercent != 0;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: global.theme.dividerBorderColor.withValues(alpha: 0.3)),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w600,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              // Current period
              Expanded(
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Column(
                    children: [
                      Text(
                        global.language('current_period'),
                        style: TextStyle(
                          fontSize: 11,
                          color: global.theme.textSecondaryColor,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        currentValue,
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                          color: global.theme.textColor,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              // VS indicator
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Text(
                  'vs',
                  style: TextStyle(
                    fontSize: 13,
                    color: global.theme.textSecondaryColor,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              // Previous period
              Expanded(
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: global.theme.surfaceColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Column(
                    children: [
                      Text(
                        global.language('previous_period'),
                        style: TextStyle(
                          fontSize: 11,
                          color: global.theme.textSecondaryColor,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        previousValue,
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                          color: global.theme.textColor,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              // Trend
              if (hasTrend)
                Padding(
                  padding: const EdgeInsets.only(left: 12),
                  child: Column(
                    children: [
                      Icon(
                        trendUp ? Icons.trending_up : Icons.trending_down,
                        color: trendUp
                            ? global.theme.positiveHighlightTextColor
                            : global.theme.negativeHighlightTextColor,
                        size: 24,
                      ),
                      Text(
                        '${trendPercent.abs().toStringAsFixed(1)}%',
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.bold,
                          color: trendUp
                              ? global.theme.positiveHighlightTextColor
                              : global.theme.negativeHighlightTextColor,
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTrendBreakdownRow(
    String label,
    Map<String, dynamic> current,
    Map<String, dynamic> previous,
    Color color,
  ) {
    final curCount = current['doc_count'] ?? 0;
    final prevCount = previous['doc_count'] ?? 0;
    final curAmt = (current['total_amount'] ?? 0.0) as double;
    final prevAmt = (previous['total_amount'] ?? 0.0) as double;
    final countTrend = _calcTrend(curCount, prevCount);
    final amtTrend = _calcTrend(curAmt, prevAmt);

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Row(
        children: [
          Container(
            width: 4,
            height: 40,
            decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(width: 12),
          Text(
            label,
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.bold,
              color: global.theme.textColor,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    Text(
                      '${_formatNumber(curCount)} ${global.language('document_count')}',
                      style: TextStyle(
                        fontSize: 12,
                        color: global.theme.textColor,
                      ),
                    ),
                    if (countTrend != 0) ...[
                      const SizedBox(width: 6),
                      _buildTrendBadge(countTrend),
                    ],
                  ],
                ),
                const SizedBox(height: 2),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    Text(
                      _formatAmountFull(curAmt),
                      style: TextStyle(
                        fontSize: 12,
                        color: global.theme.textSecondaryColor,
                      ),
                    ),
                    if (amtTrend != 0) ...[
                      const SizedBox(width: 6),
                      _buildTrendBadge(amtTrend),
                    ],
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTrendBadge(double percent) {
    final up = percent > 0;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
      decoration: BoxDecoration(
        color: (up
                ? global.theme.positiveHighlightTextColor
                : global.theme.negativeHighlightTextColor)
            .withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        '${up ? "+" : ""}${percent.toStringAsFixed(1)}%',
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w600,
          color: up
              ? global.theme.positiveHighlightTextColor
              : global.theme.negativeHighlightTextColor,
        ),
      ),
    );
  }

  // ===================== Tab 3: Vendors =====================

  Widget _buildVendorTab(ProcurementDashboardLoaded state) {
    if (state.vendorData.isEmpty) {
      return Center(
        child: Text(
          global.language('no_data'),
          style: TextStyle(color: global.theme.textSecondaryColor),
        ),
      );
    }

    final maxTotal = state.vendorData
        .map((r) => (r['total_amount'] as double?) ?? 0.0)
        .fold<double>(0.0, (a, b) => a > b ? a : b);

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 4,
                height: 20,
                decoration: BoxDecoration(
                  color: global.theme.primaryColor,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                global.language('top_vendors'),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: global.theme.textColor,
                  letterSpacing: -0.3,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...state.vendorData.asMap().entries.map((entry) {
            final index = entry.key;
            final row = entry.value;
            final total = (row['total_amount'] as double?) ?? 0.0;
            final docCount = row['doc_count'] ?? 0;
            final custcode = row['custcode']?.toString() ?? '';
            final progress = maxTotal > 0 ? total / maxTotal : 0.0;

            return Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: InkWell(
                borderRadius: BorderRadius.circular(12),
                onTap: () => _showVendorDetailSheet(custcode, docCount, total),
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: global.theme.cardColor,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                        color: global.theme.dividerBorderColor
                            .withValues(alpha: 0.3)),
                    boxShadow: [
                      BoxShadow(
                        color: global.theme.textColor.withValues(alpha: 0.03),
                        blurRadius: 6,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      // Rank badge
                      Container(
                        width: 32,
                        height: 32,
                        alignment: Alignment.center,
                        decoration: BoxDecoration(
                          gradient: index < 3
                              ? LinearGradient(
                                  begin: Alignment.topLeft,
                                  end: Alignment.bottomRight,
                                  colors: [
                                    global.theme.primaryColor
                                        .withValues(alpha: 0.25),
                                    global.theme.primaryColor
                                        .withValues(alpha: 0.08),
                                  ],
                                )
                              : null,
                          color: index < 3
                              ? null
                              : global.theme.surfaceColor,
                          borderRadius: BorderRadius.circular(10),
                          border: index < 3
                              ? Border.all(
                                  color: global.theme.primaryColor
                                      .withValues(alpha: 0.3))
                              : null,
                        ),
                        child: Text(
                          '${index + 1}',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w800,
                            color: index < 3
                                ? global.theme.primaryColor
                                : global.theme.textSecondaryColor,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      // Vendor info + gradient progress bar
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              custcode,
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: global.theme.textColor,
                              ),
                            ),
                            const SizedBox(height: 6),
                            ClipRRect(
                              borderRadius: BorderRadius.circular(4),
                              child: Stack(
                                children: [
                                  Container(
                                    height: 6,
                                    decoration: BoxDecoration(
                                      color: global.theme.dividerBorderColor
                                          .withValues(alpha: 0.3),
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                  ),
                                  FractionallySizedBox(
                                    widthFactor: progress,
                                    child: Container(
                                      height: 6,
                                      decoration: BoxDecoration(
                                        gradient: LinearGradient(
                                          colors: [
                                            global.theme.primaryColor
                                                .withValues(alpha: 0.7),
                                            global.theme.primaryColor,
                                          ],
                                        ),
                                        borderRadius:
                                            BorderRadius.circular(4),
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              '${_formatNumber(docCount)} ${global.language('document_count')}',
                              style: TextStyle(
                                fontSize: 11,
                                color: global.theme.textSecondaryColor,
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      // Amount
                      Text(
                        _formatAmountFull(total),
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: global.theme.textColor,
                        ),
                      ),
                      const SizedBox(width: 4),
                      Icon(
                        Icons.chevron_right,
                        size: 20,
                        color: global.theme.iconSecondaryColor,
                      ),
                    ],
                  ),
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  // ===================== Tab 4: Products =====================

  Widget _buildProductTab(ProcurementDashboardLoaded state, bool isWide) {
    if (state.productData.isEmpty) {
      return Center(
        child: Text(
          global.language('no_data'),
          style: TextStyle(color: global.theme.textSecondaryColor),
        ),
      );
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Horizontal bar chart — top products by amount
          Text(
            global.language('product'),
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: global.theme.textColor,
            ),
          ),
          const SizedBox(height: 12),
          _buildProductChart(state),
          const SizedBox(height: 20),

          // Product table
          _buildProductTable(state),
        ],
      ),
    );
  }

  Widget _buildProductChart(ProcurementDashboardLoaded state) {
    final data = state.productData.take(10).toList();
    final chartData = data
        .map((r) => _ChartData(
              (r['itemcode'] as String).length > 12
                  ? '${(r['itemcode'] as String).substring(0, 12)}...'
                  : r['itemcode'] as String,
              (r['total_amount'] as double?) ?? 0.0,
            ))
        .toList()
        .reversed
        .toList();

    return Container(
      height: (chartData.length * 36.0).clamp(200, 400),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: global.theme.dividerBorderColor.withValues(alpha: 0.3)),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      padding: const EdgeInsets.all(8),
      child: SfCartesianChart(
        plotAreaBorderWidth: 0,
        primaryXAxis: CategoryAxis(
          labelStyle: TextStyle(
            color: global.theme.textSecondaryColor,
            fontSize: 10,
          ),
          majorGridLines: const MajorGridLines(width: 0),
        ),
        primaryYAxis: NumericAxis(
          labelStyle: TextStyle(color: global.theme.textSecondaryColor),
          numberFormat: NumberFormat.compact(),
          majorGridLines: MajorGridLines(
            width: 0.5,
            color: global.theme.dividerBorderColor,
          ),
        ),
        tooltipBehavior: TooltipBehavior(enable: true),
        series: <CartesianSeries>[
          BarSeries<_ChartData, String>(
            dataSource: chartData,
            xValueMapper: (d, _) => d.label,
            yValueMapper: (d, _) => d.value,
            color: global.theme.primaryColor,
            borderRadius: const BorderRadius.only(
              topRight: Radius.circular(4),
              bottomRight: Radius.circular(4),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProductTable(ProcurementDashboardLoaded state) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: global.theme.dividerBorderColor.withValues(alpha: 0.3)),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Table(
        columnWidths: const {
          0: FixedColumnWidth(40),
          1: FlexColumnWidth(2),
          2: FlexColumnWidth(1),
          3: FlexColumnWidth(1.5),
          4: FlexColumnWidth(1.2),
        },
        children: [
          // Header
          TableRow(
            decoration: BoxDecoration(color: global.theme.columnHeaderColor),
            children: [
              _buildTableHeader('#'),
              _buildTableHeader(global.language('item_name')),
              _buildTableHeader(global.language('quantity')),
              _buildTableHeader(global.language('total_amount')),
              _buildTableHeader(global.language('avg_price')),
            ],
          ),
          // Rows
          ...state.productData.asMap().entries.map((entry) {
            final index = entry.key;
            final row = entry.value;
            final isEven = index % 2 == 0;
            return TableRow(
              decoration: BoxDecoration(
                color: isEven
                    ? global.theme.cardColor
                    : global.theme.surfaceColor,
              ),
              children: [
                _buildTableCell('${index + 1}', TextAlign.center),
                _buildTableCell(
                  '${row['itemcode'] ?? ''}\n${row['itemname'] ?? ''}',
                  TextAlign.left,
                ),
                _buildTableCell(
                  _formatNumber(
                      ((row['total_qty'] as double?) ?? 0.0).round()),
                  TextAlign.right,
                ),
                _buildTableCell(
                  _formatAmountFull((row['total_amount'] as double?) ?? 0.0),
                  TextAlign.right,
                ),
                _buildTableCell(
                  _formatAmountFull((row['avg_price'] as double?) ?? 0.0),
                  TextAlign.right,
                ),
              ],
            );
          }),
        ],
      ),
    );
  }

  // ===================== Bottom Sheet: Doc List =====================

  void _showDocListSheet(String docType, int transflag) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => DraggableScrollableSheet(
        initialChildSize: 0.6,
        minChildSize: 0.3,
        maxChildSize: 0.9,
        builder: (ctx, scrollController) => Container(
          decoration: BoxDecoration(
            color: global.theme.cardColor,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
          ),
          child: Column(
            children: [
              // Drag handle
              Container(
                margin: const EdgeInsets.only(top: 8),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: global.theme.dividerBorderColor,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              // Header
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        '$docType - ${global.language('document_count')}',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: global.theme.textColor,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: Icon(Icons.close,
                          color: global.theme.iconSecondaryColor),
                      onPressed: () => Navigator.pop(ctx),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              // Placeholder content
              Expanded(
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.construction_outlined,
                        size: 48,
                        color: global.theme.iconSecondaryColor,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        global.language('coming_soon'),
                        style: TextStyle(
                          fontSize: 14,
                          color: global.theme.textSecondaryColor,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // ===================== Bottom Sheet: Vendor Detail =====================

  void _showVendorDetailSheet(String custcode, dynamic docCount, double total) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => DraggableScrollableSheet(
        initialChildSize: 0.4,
        minChildSize: 0.2,
        maxChildSize: 0.7,
        builder: (ctx, scrollController) => Container(
          decoration: BoxDecoration(
            color: global.theme.cardColor,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
          ),
          child: Column(
            children: [
              // Drag handle
              Container(
                margin: const EdgeInsets.only(top: 8),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: global.theme.dividerBorderColor,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              // Header
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        '${global.language('vendor_code')}: $custcode',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: global.theme.textColor,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: Icon(Icons.close,
                          color: global.theme.iconSecondaryColor),
                      onPressed: () => Navigator.pop(ctx),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              // Vendor summary
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Expanded(
                      child: _buildVendorInfoTile(
                        icon: Icons.description_outlined,
                        label: global.language('document_count'),
                        value: _formatNumber(docCount),
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: _buildVendorInfoTile(
                        icon: Icons.attach_money,
                        label: global.language('total_amount'),
                        value: _formatAmountFull(total),
                      ),
                    ),
                  ],
                ),
              ),
              // Placeholder for future doc list
              Expanded(
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.construction_outlined,
                        size: 48,
                        color: global.theme.iconSecondaryColor,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        global.language('coming_soon'),
                        style: TextStyle(
                          fontSize: 14,
                          color: global.theme.textSecondaryColor,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildVendorInfoTile({
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          Icon(icon, color: global.theme.primaryColor, size: 24),
          const SizedBox(height: 6),
          Text(
            value,
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: global.theme.textColor,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ],
      ),
    );
  }

  // ===================== Shared Helpers =====================

  Widget _buildEmptyChart() {
    return Center(
      child: Text(
        global.language('no_data'),
        style: TextStyle(color: global.theme.textSecondaryColor),
      ),
    );
  }

  Widget _buildTableHeader(String text) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 10),
      child: Text(
        text,
        style: TextStyle(
          fontWeight: FontWeight.bold,
          color: global.theme.columnHeaderTextColor,
          fontSize: 12,
        ),
        textAlign: TextAlign.center,
      ),
    );
  }

  Widget _buildTableCell(String text, TextAlign align) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
      child: Text(
        text,
        style: TextStyle(
          color: global.theme.textColor,
          fontSize: 12,
        ),
        textAlign: align,
      ),
    );
  }
}

/// Chart data model for cartesian charts
class _ChartData {
  final String label;
  final double value;

  _ChartData(this.label, this.value);
}

/// Chart data model for pie/doughnut charts
class _PieChartData {
  final String label;
  final int value;
  final Color color;

  _PieChartData(this.label, this.value, this.color);
}
