import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/trans/trans_bloc.dart';
import 'package:smlaicloud/bloc/creditor_filter/creditor_filter_cubit.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/doc_payload_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/purchaserequisition/components/pr_list_card.dart';

/// หน้าจอแสดงรายการใบขอซื้อ (Purchase Requisition List)
/// โครงสร้างเหมือน PurchaseOrderListScreen — แยกออกมาเพื่อให้แก้ไขได้สะดวก
class PurchaseRequisitionListScreen extends StatefulWidget {
  const PurchaseRequisitionListScreen({
    super.key,
    this.custcode = '',
    this.isEmbedded = false,
    this.onDocumentSelected,
    this.onClose,
  });

  final String custcode;
  final bool isEmbedded;
  final Function(TransactionModel document)? onDocumentSelected;
  final VoidCallback? onClose;

  @override
  State<PurchaseRequisitionListScreen> createState() => PurchaseRequisitionListScreenState();
}

class PurchaseRequisitionListScreenState extends State<PurchaseRequisitionListScreen> with global.ThemeRefreshMixin {
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final FocusNode _searchFocusNode = FocusNode(skipTraversal: true);

  List<TransactionModel> _prList = [];
  String _searchText = '';
  bool _isLoading = false;
  int _totalRecords = 0;
  int _currentOffset = 0;
  String _selectedGuid = '';

  // Search state management
  bool _isSearching = false;
  String _pendingSearchText = '';
  String _currentSearchRequest = '';

  Timer? _autoRefreshTimer;
  static const int _refreshInterval = 5;

  final Map<String, POApprovalStatusModel> _approvalStatusMap = {};
  Timer? _approvalPollingTimer;

  final _debouncer = global.Debouncer(500);

  // === ตัวแปรสำหรับการกรองข้อมูล ===
  DateTime? _filterFromDate;
  DateTime? _filterToDate;
  double? _filterMinAmount;
  double? _filterMaxAmount;
  List<String> _filterCreditorCodes = [];
  bool _showAdvancedFilter = false;

  late CreditorFilterCubit _creditorFilterCubit;
  final TextEditingController _creditorSearchController = TextEditingController();
  Timer? _creditorSearchDebounce;
  final TextEditingController _minAmountController = TextEditingController();
  final TextEditingController _maxAmountController = TextEditingController();

  // === ขนาด font ปรับได้ ===
  double _fontSize = 12.0;
  static const double _minFontSize = 10.0;
  static const double _maxFontSize = 16.0;

  @override
  void initState() {
    super.initState();
    _creditorFilterCubit = CreditorFilterCubit(
      creditorRepository: CreditorRepository(),
    );
    _creditorFilterCubit.loadCreditors();
    _loadData();
    _scrollController.addListener(_onScroll);
    _startAutoRefresh();
    _startApprovalPolling();
  }

  @override
  void dispose() {
    _searchController.dispose();
    _scrollController.dispose();
    _searchFocusNode.dispose();
    _autoRefreshTimer?.cancel();
    _approvalPollingTimer?.cancel();
    _creditorSearchController.dispose();
    _creditorSearchDebounce?.cancel();
    _minAmountController.dispose();
    _maxAmountController.dispose();
    _creditorFilterCubit.close();
    _debouncer.dispose();
    super.dispose();
  }

  void _startAutoRefresh() {
    if (_autoRefreshTimer != null) return;
    _autoRefreshTimer = Timer.periodic(
      Duration(seconds: _refreshInterval),
      (_) => _silentRefresh(),
    );
  }

  void _startApprovalPolling() {
    if (_approvalPollingTimer != null) return;
    _approvalPollingTimer = Timer.periodic(
      const Duration(seconds: 5),
      (_) => _loadApprovalStatuses(),
    );
  }

  void _silentRefresh() {
    if (_isLoading || _prList.length > 20) return;
    _loadData(silent: true);
  }

  void _loadData({bool silent = false, bool loadMore = false}) {
    if (_isLoading && _searchText == _currentSearchRequest && !loadMore) return;

    _currentSearchRequest = _searchText;

    if (!silent) {
      setState(() => _isLoading = true);
    }

    final offset = loadMore ? _prList.length : 0;
    _currentOffset = offset;

    String? fromDateStr;
    String? toDateStr;
    if (_filterFromDate != null) {
      fromDateStr = "${_filterFromDate!.year}-${_filterFromDate!.month.toString().padLeft(2, '0')}-${_filterFromDate!.day.toString().padLeft(2, '0')}";
    }
    if (_filterToDate != null) {
      toDateStr = "${_filterToDate!.year}-${_filterToDate!.month.toString().padLeft(2, '0')}-${_filterToDate!.day.toString().padLeft(2, '0')}";
    }

    AppLogger.info('[PR List] _loadData called - fromDate: $fromDateStr, toDate: $toDateStr, minAmount: $_filterMinAmount, maxAmount: $_filterMaxAmount');

    context.read<TransBloc>().add(
      TransLoad(
        search: _searchText,
        type: global.TransactionTypeEnum.purchaserequisition,
        custcode: widget.custcode,
        limit: 20,
        offset: offset,
        ispos: '',
        dateorder: 1,
        fromDate: fromDateStr,
        toDate: toDateStr,
        minAmount: _filterMinAmount,
        maxAmount: _filterMaxAmount,
        custCodes: _filterCreditorCodes.isNotEmpty ? _filterCreditorCodes : null,
      ),
    );
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      if (!_isLoading && _prList.length < _totalRecords) {
        _loadData(loadMore: true);
      }
    }
  }

  void _onSearch(String value) {
    if (_isLoading) return;

    final trimmedValue = value.trim();
    _pendingSearchText = trimmedValue;

    if (trimmedValue == _currentSearchRequest) return;

    setState(() => _isSearching = true);

    _debouncer.run(() {
      if (!mounted) return;
      if (_isLoading) {
        setState(() => _isSearching = false);
        return;
      }
      if (_pendingSearchText != trimmedValue) {
        setState(() => _isSearching = false);
        return;
      }
      if (_pendingSearchText == _currentSearchRequest) {
        setState(() => _isSearching = false);
        return;
      }

      setState(() {
        _searchText = _pendingSearchText;
        _isSearching = false;
        _prList = [];
      });
      _loadData();
    });
  }

  void _onSearchSubmitted(String value) {
    if (_isLoading) return;

    final trimmedValue = value.trim();
    _pendingSearchText = trimmedValue;

    if (trimmedValue != _currentSearchRequest) {
      setState(() {
        _searchText = trimmedValue;
        _prList = [];
        _isSearching = false;
      });
      _loadData();
    }
    _searchFocusNode.requestFocus();
  }

  Future<void> _loadApprovalStatuses() async {
    if (_prList.isEmpty) return;

    final prListSnapshot = List<TransactionModel>.from(_prList);

    for (final pr in prListSnapshot) {
      if (pr.docno.isEmpty) continue;
      try {
        // PR ใช้ approval endpoint เดียวกับ PO (prefix pr-)
        final result = await ApprovalApiService.getPOApprovalStatus(pr.docno);
        if (result.isSuccess && result.data != null && mounted) {
          setState(() {
            _approvalStatusMap[pr.docno] = result.data!;
          });
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.debug('Error loading approval status for ${pr.docno}: $e');
        }
      }
    }
  }

  // === Filter Methods ===

  bool get _hasActiveFilters =>
      _filterFromDate != null ||
      _filterToDate != null ||
      _filterMinAmount != null ||
      _filterMaxAmount != null ||
      _filterCreditorCodes.isNotEmpty;

  void _clearAllFilters() {
    setState(() {
      _filterFromDate = null;
      _filterToDate = null;
      _filterMinAmount = null;
      _filterMaxAmount = null;
      _filterCreditorCodes = [];
      _minAmountController.clear();
      _maxAmountController.clear();
      _prList = [];
    });
    _loadData();
  }

  void _addCreditorFilter(String code) {
    if (!_filterCreditorCodes.contains(code)) {
      setState(() {
        _filterCreditorCodes = [..._filterCreditorCodes, code];
        _prList = [];
      });
      _loadData();
    }
  }

  void _removeCreditorFilter(String code) {
    setState(() {
      _filterCreditorCodes = _filterCreditorCodes.where((c) => c != code).toList();
      _prList = [];
    });
    _loadData();
  }

  void _increaseFontSize() {
    setState(() {
      if (_fontSize < _maxFontSize) _fontSize += 1.0;
    });
  }

  void _decreaseFontSize() {
    setState(() {
      if (_fontSize > _minFontSize) _fontSize -= 1.0;
    });
  }

  Future<void> _selectDate(bool isFromDate, BuildContext buttonContext) async {
    DateTime initialDate = isFromDate
        ? (_filterFromDate ?? DateTime.now())
        : (_filterToDate ?? DateTime.now());
    DateTime now = DateTime.now();
    DateTime firstDate = DateTime(now.year - 10, 1, 1);
    DateTime lastDate = DateTime(now.year + 10, 12, 31);

    if (initialDate.isBefore(firstDate)) {
      initialDate = firstDate;
    } else if (initialDate.isAfter(lastDate)) {
      initialDate = lastDate;
    }

    final RenderBox button = buttonContext.findRenderObject() as RenderBox;
    final Offset buttonPosition = button.localToGlobal(Offset.zero);
    final Size buttonSize = button.size;
    final screenSize = MediaQuery.of(context).size;

    const dialogWidth = 360.0;
    const dialogHeight = 450.0;

    double left = buttonPosition.dx + (buttonSize.width / 2) - (dialogWidth / 2);
    if (left < 10) left = 10;
    if (left + dialogWidth > screenSize.width - 10) {
      left = screenSize.width - dialogWidth - 10;
    }

    double top = buttonPosition.dy + buttonSize.height + 8;
    if (top + dialogHeight > screenSize.height - 20) {
      top = buttonPosition.dy - dialogHeight - 8;
      if (top < 20) top = 20;
    }

    DateTime? pickedDate = await showDialog<DateTime>(
      context: context,
      barrierColor: Colors.black.withValues(alpha: 0.3),
      builder: (BuildContext dialogContext) {
        return Stack(
          children: [
            Positioned(
              left: left,
              top: top,
              child: Material(
                color: Colors.transparent,
                child: CustomDatePickerDialog(
                  initialDate: initialDate,
                  firstDate: firstDate,
                  lastDate: lastDate,
                ),
              ),
            ),
          ],
        );
      },
    );

    if (pickedDate != null) {
      setState(() {
        if (isFromDate) {
          _filterFromDate = pickedDate;
        } else {
          _filterToDate = pickedDate;
        }
        _prList = [];
      });
      _loadData();
    }
  }

  // === Public Methods ===

  void refreshList() {
    _loadData();
  }

  void setCardLoadingState(String guid) {
    setState(() {
      _selectedGuid = guid;
    });
  }

  void refreshListAndMaintainPosition({String? selectDocGuid}) {
    if (selectDocGuid != null) {
      _selectedGuid = selectDocGuid;
    }
    _loadData(silent: true);
  }

  void refreshListAfterDelete(String deletedGuid) {
    setState(() {
      _prList.removeWhere((pr) => pr.guidfixed == deletedGuid);
      if (_selectedGuid == deletedGuid) {
        _selectedGuid = '';
      }
    });
    _loadData(silent: true);
  }

  @override
  Widget build(BuildContext context) {
    final content = BlocListener<TransBloc, TransState>(
      listener: (context, state) {
        if (state is TransLoadSuccess) {
          setState(() {
            _isLoading = false;
            if (_currentOffset == 0) {
              _prList = state.trans;
            } else {
              _prList.addAll(state.trans);
            }
            _totalRecords = state.totalRecord;
          });
          _loadApprovalStatuses();

          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (_scrollController.hasClients &&
                _scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200 &&
                _prList.length < _totalRecords) {
              _loadData(loadMore: true);
            }
          });
        } else if (state is TransLoadFailed) {
          setState(() => _isLoading = false);
          if (mounted) {
            global.showErrorSnackBar(context, '${global.language("error_occurred")}: ${state.message}');
          }
        }
      },
      child: Column(
        children: [
          _buildHeader(),
          _buildStatusDashboard(),
          _buildSearchBar(),
          _buildAdvancedFilterPanel(),
          Expanded(child: _buildList()),
        ],
      ),
    );

    if (widget.isEmbedded) {
      return content;
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('purchase_requisition')),
      ),
      body: content,
    );
  }

  Widget _buildStatusDashboard() {
    if (_approvalStatusMap.isEmpty) return const SizedBox.shrink();

    int pending = 0, approved = 0, rejected = 0, draft = 0;
    for (final entry in _approvalStatusMap.values) {
      switch (entry.status) {
        case POApprovalStatus.pending:
          pending++;
          break;
        case POApprovalStatus.approved:
        case POApprovalStatus.autoApproved:
          approved++;
          break;
        case POApprovalStatus.rejected:
          rejected++;
          break;
        case POApprovalStatus.draft:
          draft++;
          break;
      }
    }
    // รวมเอกสารที่ยังไม่ส่งอนุมัติ (ไม่มีใน approvalStatusMap)
    final notSubmitted = _prList.length - _approvalStatusMap.length;
    if (notSubmitted > 0) draft += notSubmitted;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Row(
        children: [
          _buildStatusChip(Icons.edit_note, global.language('draft'), draft, global.theme.iconSecondaryColor),
          const SizedBox(width: 8),
          _buildStatusChip(Icons.hourglass_empty, global.language('pending_approval'), pending, global.theme.warningHighlightTextColor),
          const SizedBox(width: 8),
          _buildStatusChip(Icons.check_circle, global.language('approval_approved'), approved, global.theme.positiveHighlightTextColor),
          const SizedBox(width: 8),
          _buildStatusChip(Icons.cancel, global.language('reject'), rejected, global.theme.negativeHighlightTextColor),
        ],
      ),
    );
  }

  Widget _buildStatusChip(IconData icon, String label, int count, Color color) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 8),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: color.withValues(alpha: 0.3)),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 14, color: color),
            const SizedBox(width: 4),
            Text(
              '$count',
              style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: color),
            ),
            const SizedBox(width: 4),
            Flexible(
              child: Text(
                label,
                style: TextStyle(fontSize: 10, color: color),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    String headerText;
    if (_hasActiveFilters) {
      headerText = '${global.language("found")} $_totalRecords ${global.language("records")}';
    } else {
      headerText = '${global.language("total")} $_totalRecords ${global.language("records")}';
    }

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            headerText,
            style: TextStyle(fontSize: _fontSize, color: global.theme.iconSecondaryColor),
          ),
          if (widget.isEmbedded && widget.onClose != null)
            IconButton(icon: Icon(Icons.close), onPressed: widget.onClose),
        ],
      ),
    );
  }

  Widget _buildSearchBar() {
    final isSearchingOrLoading = _isSearching || _isLoading;

    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: TextFormField(
              controller: _searchController,
              focusNode: _searchFocusNode,
              autofocus: false,
              onFieldSubmitted: _onSearchSubmitted,
              onChanged: _onSearch,
              decoration: InputDecoration(
                isDense: true,
                hintText: _isLoading ? global.language('searching') : global.language('search'),
                hintStyle: TextStyle(color: global.theme.formHintColor),
                prefixIcon: Icon(
                  Icons.search,
                  color: _isLoading ? global.theme.dividerBorderColor : global.theme.iconSecondaryColor,
                ),
                suffixIcon: isSearchingOrLoading
                    ? const Padding(
                        padding: EdgeInsets.all(12.0),
                        child: SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      )
                    : _searchController.text.isNotEmpty
                        ? IconButton(
                            icon: Icon(Icons.clear, size: 18),
                            onPressed: _isLoading
                                ? null
                                : () {
                                    _searchController.clear();
                                    _pendingSearchText = '';
                                    setState(() {
                                      _searchText = '';
                                      _prList = [];
                                      _isSearching = false;
                                    });
                                    _loadData();
                                  },
                          )
                        : null,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: global.theme.infoHighlightTextColor),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: global.theme.infoHighlightTextColor),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: global.theme.infoHighlightTextColor, width: 2),
                ),
                filled: true,
                fillColor: global.theme.formFillColor,
                contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              ),
            ),
          ),
          SizedBox(width: 8),
          Container(
            decoration: BoxDecoration(
              color: _showAdvancedFilter ? global.theme.rowHoverColor : global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: _hasActiveFilters ? global.theme.infoHighlightTextColor : global.theme.dividerBorderColor,
                width: 1,
              ),
            ),
            child: IconButton(
              constraints: const BoxConstraints(minWidth: 40, minHeight: 40),
              onPressed: () {
                setState(() {
                  _showAdvancedFilter = !_showAdvancedFilter;
                });
              },
              icon: Icon(
                _hasActiveFilters ? Icons.filter_alt : Icons.tune,
                color: _hasActiveFilters ? global.theme.infoHighlightTextColor : global.theme.iconSecondaryColor,
                size: 20,
              ),
              tooltip: global.language('advanced_filter'),
            ),
          ),
          SizedBox(width: 4),
          Container(
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: global.theme.dividerBorderColor, width: 1),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  constraints: const BoxConstraints(minWidth: 32, minHeight: 40),
                  padding: EdgeInsets.zero,
                  onPressed: _fontSize > _minFontSize ? _decreaseFontSize : null,
                  icon: Icon(
                    Icons.text_decrease,
                    size: 16,
                    color: _fontSize > _minFontSize ? global.theme.textColor : global.theme.dividerBorderColor,
                  ),
                  tooltip: global.language('decrease_font_size'),
                ),
                Text(
                  '${_fontSize.toInt()}',
                  style: TextStyle(fontSize: 10, color: global.theme.iconSecondaryColor),
                ),
                IconButton(
                  constraints: const BoxConstraints(minWidth: 32, minHeight: 40),
                  padding: EdgeInsets.zero,
                  onPressed: _fontSize < _maxFontSize ? _increaseFontSize : null,
                  icon: Icon(
                    Icons.text_increase,
                    size: 16,
                    color: _fontSize < _maxFontSize ? global.theme.textColor : global.theme.dividerBorderColor,
                  ),
                  tooltip: global.language('increase_font_size'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAdvancedFilterPanel() {
    if (!_showAdvancedFilter) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [global.theme.surfaceColor, global.theme.surfaceColor],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.dividerBorderColor),
        boxShadow: [
          BoxShadow(
            color: Colors.blue.withValues(alpha: 0.1),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Padding(
        padding: EdgeInsets.all(12),
        child: Column(
          children: [
            // แถวบน: ช่วงวันที่
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: global.theme.surfaceColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.date_range, size: 16, color: global.theme.infoHighlightTextColor),
                ),
                SizedBox(width: 8),
                Expanded(
                  child: _buildDateButton(
                    label: global.language('from'),
                    date: _filterFromDate,
                    isFromDate: true,
                    color: global.theme.infoHighlightTextColor,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward, size: 14, color: global.theme.iconSecondaryColor),
                ),
                Expanded(
                  child: _buildDateButton(
                    label: global.language('kb_to'),
                    date: _filterToDate,
                    isFromDate: false,
                    color: global.theme.infoHighlightTextColor,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            // แถวล่าง: ช่วงจำนวนเงิน
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: global.theme.positiveHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.attach_money, size: 16, color: global.theme.positiveHighlightTextColor),
                ),
                SizedBox(width: 8),
                Expanded(
                  child: _buildAmountField(
                    controller: _minAmountController,
                    hint: global.language('lower_bound'),
                    onChanged: (value) {
                      setState(() {
                        _filterMinAmount = double.tryParse(value);
                        _prList = [];
                      });
                      _loadData();
                    },
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward, size: 14, color: global.theme.iconSecondaryColor),
                ),
                Expanded(
                  child: _buildAmountField(
                    controller: _maxAmountController,
                    hint: global.language('max'),
                    onChanged: (value) {
                      setState(() {
                        _filterMaxAmount = double.tryParse(value);
                        _prList = [];
                      });
                      _loadData();
                    },
                  ),
                ),
                const SizedBox(width: 8),
                if (_hasActiveFilters)
                  Material(
                    color: global.theme.negativeHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                    child: InkWell(
                      onTap: _clearAllFilters,
                      borderRadius: BorderRadius.circular(8),
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        child: Icon(Icons.clear_all, size: 18, color: global.theme.negativeHighlightTextColor),
                      ),
                    ),
                  ),
              ],
            ),
            // แถวเจ้าหนี้
            const SizedBox(height: 10),
            _buildCreditorFilterSection(),
          ],
        ),
      ),
    );
  }

  Widget _buildDateButton({
    required String label,
    required DateTime? date,
    required bool isFromDate,
    required Color color,
  }) {
    final hasDate = date != null;
    return Builder(
      builder: (buttonContext) {
        return Material(
          color: hasDate ? color.withValues(alpha: 0.1) : global.theme.cardColor,
          borderRadius: BorderRadius.circular(8),
          child: InkWell(
            onTap: () => _selectDate(isFromDate, buttonContext),
            borderRadius: BorderRadius.circular(8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: hasDate ? color : global.theme.dividerBorderColor,
                  width: hasDate ? 1.5 : 1,
                ),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.calendar_today, size: 12, color: hasDate ? color : global.theme.iconSecondaryColor),
                  const SizedBox(width: 6),
                  Flexible(
                    child: Text(
                      hasDate
                          ? global.formatThaiDateTime(dateTime: date, showTime: false)
                          : label,
                      style: TextStyle(
                        fontSize: _fontSize,
                        color: hasDate ? global.theme.infoHighlightTextColor : global.theme.iconSecondaryColor,
                        fontWeight: hasDate ? FontWeight.w600 : FontWeight.normal,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildAmountField({
    required TextEditingController controller,
    required String hint,
    required ValueChanged<String> onChanged,
  }) {
    final hasValue = controller.text.isNotEmpty;
    return Builder(
      builder: (buttonContext) {
        return InkWell(
          onTap: () async {
            final RenderBox button = buttonContext.findRenderObject() as RenderBox;
            final Offset buttonPosition = button.localToGlobal(Offset.zero);
            final Size buttonSize = button.size;
            final screenSize = MediaQuery.of(context).size;

            const dialogWidth = 320.0;
            const dialogHeight = 560.0;

            double left = buttonPosition.dx + (buttonSize.width / 2) - (dialogWidth / 2);
            if (left < 10) left = 10;
            if (left + dialogWidth > screenSize.width - 10) {
              left = screenSize.width - dialogWidth - 10;
            }

            double top = buttonPosition.dy + buttonSize.height + 8;
            if (top + dialogHeight > screenSize.height - 20) {
              top = buttonPosition.dy - dialogHeight - 8;
              if (top < 20) top = 20;
            }

            final result = await showDialog<String>(
              context: context,
              barrierColor: Colors.black.withValues(alpha: 0.3),
              builder: (BuildContext dialogContext) {
                return Stack(
                  children: [
                    Positioned(
                      left: left,
                      top: top,
                      child: Material(
                        color: Colors.transparent,
                        child: global.NumPadDialog(
                          title: hint,
                          initialValue: controller.text.isEmpty ? null : controller.text,
                          decimalPlaces: 2,
                        ),
                      ),
                    ),
                  ],
                );
              },
            );
            if (result != null) {
              controller.text = result;
              onChanged(result);
            }
          },
          child: Container(
            height: 36,
            padding: const EdgeInsets.symmetric(horizontal: 8),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: hasValue ? global.theme.positiveHighlightTextColor : global.theme.dividerBorderColor,
                width: hasValue ? 1.5 : 1,
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    hasValue ? controller.text : hint,
                    style: TextStyle(
                      fontSize: _fontSize,
                      color: hasValue ? global.theme.positiveHighlightTextColor : global.theme.dividerBorderColor,
                      fontWeight: hasValue ? FontWeight.w600 : FontWeight.normal,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (hasValue)
                  GestureDetector(
                    onTap: () {
                      controller.clear();
                      onChanged('');
                    },
                    child: Icon(Icons.clear, size: 14, color: global.theme.iconSecondaryColor),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildCreditorFilterSection() {
    return BlocBuilder<CreditorFilterCubit, CreditorFilterState>(
      bloc: _creditorFilterCubit,
      builder: (context, state) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: global.theme.infoHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.business, size: 16, color: global.theme.warningHighlightTextColor),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _buildCreditorDropdown(state),
                ),
              ],
            ),
            if (_filterCreditorCodes.isNotEmpty) ...[
              const SizedBox(height: 8),
              _buildSelectedCreditorChips(state),
            ],
          ],
        );
      },
    );
  }

  Widget _buildCreditorDropdown(CreditorFilterState state) {
    final availableCreditors = state.creditors
        .where((c) => !_filterCreditorCodes.contains(c.code))
        .toList();

    final searchText = state.searchText;
    final hasSearchText = searchText.isNotEmpty;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          height: 36,
          child: TextField(
            controller: _creditorSearchController,
            style: TextStyle(fontSize: _fontSize, fontWeight: FontWeight.w500, color: global.theme.textColor),
            decoration: InputDecoration(
              hintText: state.isLoading ? global.language('searching') : global.language('type_creditor_code_name_to_search'),
              hintStyle: TextStyle(fontSize: _fontSize, color: global.theme.formHintColor),
              prefixIcon: state.isLoading
                  ? SizedBox(
                      width: 16,
                      height: 16,
                      child: Padding(
                        padding: const EdgeInsets.all(10),
                        child: CircularProgressIndicator(strokeWidth: 2, color: global.theme.warningHighlightTextColor),
                      ),
                    )
                  : Icon(Icons.search, size: 16, color: global.theme.warningHighlightTextColor),
              suffixIcon: hasSearchText
                  ? IconButton(
                      icon: Icon(Icons.clear, size: 14, color: global.theme.iconSecondaryColor),
                      onPressed: () {
                        _creditorSearchController.clear();
                        _creditorFilterCubit.searchCreditors("");
                      },
                      padding: EdgeInsets.zero,
                      constraints: const BoxConstraints(),
                    )
                  : null,
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
              filled: true,
              fillColor: global.theme.formFillColor,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: global.theme.formBorderColor),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: global.theme.formBorderColor),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: global.theme.warningHighlightTextColor, width: 1.5),
              ),
            ),
            onChanged: (value) {
              _creditorSearchDebounce?.cancel();
              _creditorSearchDebounce = Timer(const Duration(milliseconds: 300), () {
                _creditorFilterCubit.searchCreditors(value);
              });
            },
          ),
        ),
        if (hasSearchText && availableCreditors.isNotEmpty && !state.isLoading)
          Container(
            margin: const EdgeInsets.only(top: 4),
            constraints: const BoxConstraints(maxHeight: 150),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: global.theme.warningHighlightColor),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.1),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: ListView.builder(
              shrinkWrap: true,
              padding: EdgeInsets.zero,
              itemCount: availableCreditors.length > 10 ? 10 : availableCreditors.length,
              itemBuilder: (context, index) {
                final creditor = availableCreditors[index];
                final name = creditor.names.isNotEmpty ? creditor.names.first.name : '';
                return InkWell(
                  onTap: () {
                    _addCreditorFilter(creditor.code);
                    _creditorSearchController.clear();
                    _creditorFilterCubit.searchCreditors("");
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                    decoration: BoxDecoration(
                      border: index < (availableCreditors.length > 10 ? 9 : availableCreditors.length - 1)
                          ? Border(bottom: BorderSide(color: global.theme.dividerBorderColor))
                          : null,
                    ),
                    child: Row(
                      children: [
                        Icon(Icons.business, size: 14, color: global.theme.warningHighlightTextColor),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                name,
                                style: TextStyle(fontSize: _fontSize, fontWeight: FontWeight.w500),
                                overflow: TextOverflow.ellipsis,
                              ),
                              Text(
                                '${global.language("code")}: ${creditor.code}',
                                style: TextStyle(fontSize: _fontSize - 1, color: global.theme.iconSecondaryColor),
                              ),
                            ],
                          ),
                        ),
                        Icon(Icons.add_circle_outline, size: 16, color: global.theme.positiveHighlightTextColor),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        if (hasSearchText && availableCreditors.isEmpty && !state.isLoading)
          Container(
            margin: const EdgeInsets.only(top: 4),
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(Icons.info_outline, size: 14, color: global.theme.iconSecondaryColor),
                const SizedBox(width: 6),
                Text(
                  '${global.language("creditor_not_found")} "$searchText"',
                  style: TextStyle(fontSize: _fontSize - 1, color: global.theme.iconSecondaryColor),
                ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildSelectedCreditorChips(CreditorFilterState state) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(6),
      decoration: BoxDecoration(
        color: global.theme.warningHighlightColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.warningHighlightColor),
      ),
      child: Wrap(
        spacing: 6,
        runSpacing: 4,
        children: _filterCreditorCodes.map((code) {
          String name = code;
          try {
            final creditor = state.creditors.firstWhere((c) => c.code == code);
            if (creditor.names.isNotEmpty) {
              name = creditor.names.first.name;
            }
          } catch (_) {}

          return Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: global.theme.warningHighlightTextColor),
              boxShadow: [
                BoxShadow(
                  color: Colors.orange.withValues(alpha: 0.1),
                  blurRadius: 2,
                  offset: const Offset(0, 1),
                ),
              ],
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.business, size: 12, color: global.theme.warningHighlightTextColor),
                const SizedBox(width: 4),
                Text(
                  name.length > 15 ? '${name.substring(0, 15)}...' : name,
                  style: TextStyle(
                    fontSize: _fontSize - 1,
                    color: global.theme.warningHighlightTextColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(width: 4),
                InkWell(
                  onTap: () => _removeCreditorFilter(code),
                  child: Container(
                    padding: const EdgeInsets.all(2),
                    decoration: BoxDecoration(
                      color: global.theme.negativeHighlightColor,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(Icons.close, size: 12, color: global.theme.negativeHighlightTextColor),
                  ),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildList() {
    if (_isLoading && _prList.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_prList.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inbox_outlined, size: 64, color: global.theme.dividerBorderColor),
            SizedBox(height: 16),
            Text(
              _hasActiveFilters ? global.language('no_data_matching_criteria') : global.language('no_data'),
              style: TextStyle(fontSize: _fontSize + 2, color: global.theme.iconSecondaryColor),
            ),
            if (_hasActiveFilters) ...[
              SizedBox(height: 8),
              TextButton.icon(
                onPressed: _clearAllFilters,
                icon: Icon(Icons.clear_all),
                label: Text(global.language('clear_conditions')),
              ),
            ],
          ],
        ),
      );
    }

    final hasMoreData = _prList.length < _totalRecords;
    final showLoadMoreSpinner = _isLoading && hasMoreData;

    return RefreshIndicator(
      onRefresh: () async => _loadData(),
      child: ListView.builder(
        controller: _scrollController,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: _prList.length + (showLoadMoreSpinner ? 1 : 0),
        itemBuilder: (context, index) {
          if (index >= _prList.length) {
            return const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator()),
            );
          }
          return _buildPRCard(_prList[index]);
        },
      ),
    );
  }

  Widget _buildPRCard(TransactionModel pr) {
    return PRListCard(
      pr: pr,
      isSelected: _selectedGuid == pr.guidfixed,
      fontSize: _fontSize,
      approvalStatus: _approvalStatusMap[pr.docno],
      onTap: () => _onPRSelected(pr),
    );
  }

  /// โหลด document เต็มจาก MongoDB แล้วส่งไป edit screen
  Future<void> _onPRSelected(TransactionModel pr) async {
    setState(() => _selectedGuid = pr.guidfixed ?? '');

    if (pr.guidfixed == null || pr.guidfixed!.isEmpty) {
      _passDocumentToCallback(pr);
      return;
    }

    try {
      const String collection = 'transactionPurchaseRequisition';
      MongoGetDataPayloadModel docPayLoad = MongoGetDataPayloadModel(
        shopid: global.getShopId(),
        collection: collection,
        guidfixed: pr.guidfixed!,
      );

      AppLogger.info('[PR List] Loading full document from MongoDB: ${pr.docno}');
      final response = await global.goApiPost(
        global.goApiUrlPath("mongogetdata"),
        docPayLoad.toJson(),
      );

      String rawData = json.encoder.convert(response);
      GoApiQueryOneResultResponse data = GoApiQueryOneResultResponse.fromMap(json.decode(rawData));

      TransactionModel fullDocument = TransactionModel.fromJson(data.result);
      global.parseTransactionJsonData(fullDocument);

      // รักษาค่า flags จาก PostgreSQL
      fullDocument.isref = pr.isref;
      fullDocument.islocked = pr.islocked;
      fullDocument.isclosed = pr.isclosed;
      fullDocument.iscancel = pr.iscancel;
      fullDocument.isdelete = pr.isdelete;
      fullDocument.iscomparedsuccess = pr.iscomparedsuccess;

      AppLogger.info('[PR List] Loaded full document with ${fullDocument.details?.length ?? 0} details');
      _passDocumentToCallback(fullDocument);
    } catch (e) {
      AppLogger.error('[PR List] Error loading full document: $e');
      _passDocumentToCallback(pr);
    }
  }

  void _passDocumentToCallback(TransactionModel document) {
    if (widget.onDocumentSelected != null) {
      widget.onDocumentSelected!(document);
    } else if (mounted) {
      Navigator.pop(context, document);
    }
  }
}
