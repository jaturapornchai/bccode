import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/trans/trans_bloc.dart';
import 'package:smlaicloud/bloc/debtor_filter/debtor_filter_cubit.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/doc_payload_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/debtor_repository.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/quotation/components/qt_list_card.dart';

/// หน้าจอแสดงรายการใบเสนอราคา (Quotation List)
/// แยกออกมาจาก TransSearchScreen เพื่อให้แก้ไขได้สะดวกโดยไม่กระทบระบบอื่น
class QuotationListScreen extends StatefulWidget {
  const QuotationListScreen({
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
  State<QuotationListScreen> createState() => QuotationListScreenState();
}

class QuotationListScreenState extends State<QuotationListScreen> {
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final FocusNode _searchFocusNode = FocusNode(skipTraversal: true);

  List<TransactionModel> _qtList = [];
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
  List<String> _filterDebtorCodes = [];
  bool _showAdvancedFilter = false;

  // Cubit สำหรับจัดการรายชื่อลูกหนี้
  late DebtorFilterCubit _debtorFilterCubit;

  // Controller สำหรับค้นหาลูกหนี้
  final TextEditingController _debtorSearchController = TextEditingController();
  Timer? _debtorSearchDebounce;

  // Controllers สำหรับ TextField ยอดเงิน
  final TextEditingController _minAmountController = TextEditingController();
  final TextEditingController _maxAmountController = TextEditingController();

  // === ขนาด font ปรับได้ ===
  double _fontSize = 12.0;
  static const double _minFontSize = 10.0;
  static const double _maxFontSize = 16.0;

  @override
  void initState() {
    super.initState();
    _debtorFilterCubit = DebtorFilterCubit(
      debtorRepository: DebtorRepository(),
    );
    _debtorFilterCubit.loadDebtors();
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
    _debtorSearchController.dispose();
    _debtorSearchDebounce?.cancel();
    _minAmountController.dispose();
    _maxAmountController.dispose();
    _debtorFilterCubit.close();
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
    if (_isLoading || _qtList.length > 20) return;
    _loadData(silent: true);
  }

  void _loadData({bool silent = false, bool loadMore = false}) {
    if (_isLoading && _searchText == _currentSearchRequest && !loadMore) return;

    _currentSearchRequest = _searchText;

    if (!silent) {
      setState(() => _isLoading = true);
    }

    final offset = loadMore ? _qtList.length : 0;
    _currentOffset = offset;

    // แปลง DateTime เป็น String format "YYYY-MM-DD"
    String? fromDateStr;
    String? toDateStr;
    if (_filterFromDate != null) {
      fromDateStr = "${_filterFromDate!.year}-${_filterFromDate!.month.toString().padLeft(2, '0')}-${_filterFromDate!.day.toString().padLeft(2, '0')}";
    }
    if (_filterToDate != null) {
      toDateStr = "${_filterToDate!.year}-${_filterToDate!.month.toString().padLeft(2, '0')}-${_filterToDate!.day.toString().padLeft(2, '0')}";
    }

    AppLogger.info('[QT List] _loadData called - fromDate: $fromDateStr, toDate: $toDateStr, minAmount: $_filterMinAmount, maxAmount: $_filterMaxAmount, custCodes: $_filterDebtorCodes');

    context.read<TransBloc>().add(
      TransLoad(
        search: _searchText,
        type: global.TransactionTypeEnum.quotation,
        custcode: widget.custcode,
        limit: 20,
        offset: offset,
        ispos: '',
        dateorder: 1,
        fromDate: fromDateStr,
        toDate: toDateStr,
        minAmount: _filterMinAmount,
        maxAmount: _filterMaxAmount,
        custCodes: _filterDebtorCodes.isNotEmpty ? _filterDebtorCodes : null,
      ),
    );
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      if (!_isLoading && _qtList.length < _totalRecords) {
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
        _qtList = [];
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
        _qtList = [];
        _isSearching = false;
      });
      _loadData();
    }
    _searchFocusNode.requestFocus();
  }

  Future<void> _loadApprovalStatuses() async {
    if (_qtList.isEmpty) return;

    final qtListSnapshot = List<TransactionModel>.from(_qtList);

    for (final qt in qtListSnapshot) {
      if (qt.docno.isEmpty) continue;
      try {
        final result = await ApprovalApiService.getPOApprovalStatus(qt.docno);
        if (result.isSuccess && result.data != null && mounted) {
          setState(() {
            _approvalStatusMap[qt.docno] = result.data!;
          });
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.debug('Error loading approval status for ${qt.docno}: $e');
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
      _filterDebtorCodes.isNotEmpty;

  void _clearAllFilters() {
    setState(() {
      _filterFromDate = null;
      _filterToDate = null;
      _filterMinAmount = null;
      _filterMaxAmount = null;
      _filterDebtorCodes = [];
      _minAmountController.clear();
      _maxAmountController.clear();
      _qtList = [];
    });
    _loadData();
  }

  void _addDebtorFilter(String code) {
    if (!_filterDebtorCodes.contains(code)) {
      setState(() {
        _filterDebtorCodes = [..._filterDebtorCodes, code];
        _qtList = [];
      });
      _loadData();
    }
  }

  void _removeDebtorFilter(String code) {
    setState(() {
      _filterDebtorCodes = _filterDebtorCodes.where((c) => c != code).toList();
      _qtList = [];
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
        _qtList = [];
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
      _qtList.removeWhere((qt) => qt.guidfixed == deletedGuid);
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
              _qtList = state.trans;
            } else {
              _qtList.addAll(state.trans);
            }
            _totalRecords = state.totalRecord;
          });
          _loadApprovalStatuses();

          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (_scrollController.hasClients &&
                _scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200 &&
                _qtList.length < _totalRecords) {
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
        title: Text(global.language('quotation')),
      ),
      body: content,
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
            style: TextStyle(fontSize: _fontSize, color: Colors.grey[600]),
          ),
          if (widget.isEmbedded && widget.onClose != null)
            IconButton(icon: const Icon(Icons.close), onPressed: widget.onClose),
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
                hintStyle: TextStyle(
                  color: _isLoading ? Colors.grey.shade600 : Colors.grey.shade500,
                ),
                prefixIcon: Icon(
                  Icons.search,
                  color: _isLoading ? Colors.grey.shade400 : Colors.grey.shade600,
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
                            icon: const Icon(Icons.clear, size: 18),
                            onPressed: _isLoading
                                ? null
                                : () {
                                    _searchController.clear();
                                    _pendingSearchText = '';
                                    setState(() {
                                      _searchText = '';
                                      _qtList = [];
                                      _isSearching = false;
                                    });
                                    _loadData();
                                  },
                          )
                        : null,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.teal.shade400),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.teal.shade400),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.teal.shade600, width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              ),
            ),
          ),
          SizedBox(width: 8),
          Container(
            decoration: BoxDecoration(
              color: _showAdvancedFilter ? Colors.teal.shade50 : Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: _hasActiveFilters ? Colors.teal : Colors.grey.shade300,
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
                color: _hasActiveFilters ? Colors.teal : Colors.grey.shade600,
                size: 20,
              ),
              tooltip: global.language('advanced_filter'),
            ),
          ),
          SizedBox(width: 4),
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.grey.shade300, width: 1),
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
                    color: _fontSize > _minFontSize ? Colors.grey.shade700 : Colors.grey.shade400,
                  ),
                  tooltip: global.language('decrease_font_size'),
                ),
                Text(
                  '${_fontSize.toInt()}',
                  style: TextStyle(fontSize: 10, color: Colors.grey.shade600),
                ),
                IconButton(
                  constraints: const BoxConstraints(minWidth: 32, minHeight: 40),
                  padding: EdgeInsets.zero,
                  onPressed: _fontSize < _maxFontSize ? _increaseFontSize : null,
                  icon: Icon(
                    Icons.text_increase,
                    size: 16,
                    color: _fontSize < _maxFontSize ? Colors.grey.shade700 : Colors.grey.shade400,
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
          colors: [Colors.teal.shade50, Colors.green.shade50],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.teal.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.teal.withValues(alpha: 0.1),
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
                    color: Colors.teal.shade100,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.date_range, size: 16, color: Colors.teal.shade700),
                ),
                SizedBox(width: 8),
                Expanded(
                  child: _buildDateButton(
                    label: global.language('from'),
                    date: _filterFromDate,
                    isFromDate: true,
                    color: Colors.teal,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward, size: 14, color: Colors.grey.shade500),
                ),
                Expanded(
                  child: _buildDateButton(
                    label: global.language('kb_to'),
                    date: _filterToDate,
                    isFromDate: false,
                    color: Colors.teal,
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
                    color: Colors.green.shade100,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.attach_money, size: 16, color: Colors.green.shade700),
                ),
                SizedBox(width: 8),
                Expanded(
                  child: _buildAmountField(
                    controller: _minAmountController,
                    hint: global.language('lower_bound'),
                    onChanged: (value) {
                      setState(() {
                        _filterMinAmount = double.tryParse(value);
                        _qtList = [];
                      });
                      _loadData();
                    },
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward, size: 14, color: Colors.grey.shade500),
                ),
                Expanded(
                  child: _buildAmountField(
                    controller: _maxAmountController,
                    hint: global.language('max'),
                    onChanged: (value) {
                      setState(() {
                        _filterMaxAmount = double.tryParse(value);
                        _qtList = [];
                      });
                      _loadData();
                    },
                  ),
                ),
                const SizedBox(width: 8),
                if (_hasActiveFilters)
                  Material(
                    color: Colors.red.shade50,
                    borderRadius: BorderRadius.circular(8),
                    child: InkWell(
                      onTap: _clearAllFilters,
                      borderRadius: BorderRadius.circular(8),
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        child: Icon(Icons.clear_all, size: 18, color: Colors.red.shade600),
                      ),
                    ),
                  ),
              ],
            ),
            // แถวลูกหนี้
            const SizedBox(height: 10),
            _buildDebtorFilterSection(),
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
          color: hasDate ? color.withValues(alpha: 0.1) : Colors.white,
          borderRadius: BorderRadius.circular(8),
          child: InkWell(
            onTap: () => _selectDate(isFromDate, buttonContext),
            borderRadius: BorderRadius.circular(8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: hasDate ? color : Colors.grey.shade300,
                  width: hasDate ? 1.5 : 1,
                ),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.calendar_today,
                    size: 12,
                    color: hasDate ? color : Colors.grey.shade500,
                  ),
                  const SizedBox(width: 6),
                  Flexible(
                    child: Text(
                      hasDate
                          ? global.formatThaiDateTime(dateTime: date, showTime: false)
                          : label,
                      style: TextStyle(
                        fontSize: _fontSize,
                        color: hasDate ? Colors.teal.shade700 : Colors.grey.shade500,
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
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: hasValue ? Colors.green.shade400 : Colors.grey.shade300,
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
                      color: hasValue ? Colors.green.shade700 : Colors.grey.shade400,
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
                    child: Icon(Icons.clear, size: 14, color: Colors.grey.shade500),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildDebtorFilterSection() {
    return BlocBuilder<DebtorFilterCubit, DebtorFilterState>(
      bloc: _debtorFilterCubit,
      builder: (context, state) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: Colors.orange.shade100,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.person, size: 16, color: Colors.orange.shade700),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _buildDebtorDropdown(state),
                ),
              ],
            ),
            if (_filterDebtorCodes.isNotEmpty) ...[
              const SizedBox(height: 8),
              _buildSelectedDebtorChips(state),
            ],
          ],
        );
      },
    );
  }

  Widget _buildDebtorDropdown(DebtorFilterState state) {
    final availableDebtors = state.debtors
        .where((d) => !_filterDebtorCodes.contains(d.code))
        .toList();

    final searchText = state.searchText;
    final hasSearchText = searchText.isNotEmpty;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          height: 36,
          child: TextField(
            controller: _debtorSearchController,
            style: TextStyle(fontSize: _fontSize, fontWeight: FontWeight.w500),
            decoration: InputDecoration(
              hintText: state.isLoading ? global.language('searching') : global.language('search_debtor_hint'),
              hintStyle: TextStyle(fontSize: _fontSize, color: Colors.grey.shade400),
              prefixIcon: state.isLoading
                  ? SizedBox(
                      width: 16,
                      height: 16,
                      child: Padding(
                        padding: const EdgeInsets.all(10),
                        child: CircularProgressIndicator(strokeWidth: 2, color: Colors.orange.shade400),
                      ),
                    )
                  : Icon(Icons.search, size: 16, color: Colors.orange.shade400),
              suffixIcon: hasSearchText
                  ? IconButton(
                      icon: Icon(Icons.clear, size: 14, color: Colors.grey.shade500),
                      onPressed: () {
                        _debtorSearchController.clear();
                        _debtorFilterCubit.searchDebtors("");
                      },
                      padding: EdgeInsets.zero,
                      constraints: const BoxConstraints(),
                    )
                  : null,
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.orange.shade400, width: 1.5),
              ),
            ),
            onChanged: (value) {
              _debtorSearchDebounce?.cancel();
              _debtorSearchDebounce = Timer(const Duration(milliseconds: 300), () {
                _debtorFilterCubit.searchDebtors(value);
              });
            },
          ),
        ),
        if (hasSearchText && availableDebtors.isNotEmpty && !state.isLoading)
          Container(
            margin: const EdgeInsets.only(top: 4),
            constraints: const BoxConstraints(maxHeight: 150),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.orange.shade200),
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
              itemCount: availableDebtors.length > 10 ? 10 : availableDebtors.length,
              itemBuilder: (context, index) {
                final debtor = availableDebtors[index];
                final name = debtor.names.isNotEmpty ? debtor.names.first.name : '';
                return InkWell(
                  onTap: () {
                    _addDebtorFilter(debtor.code);
                    _debtorSearchController.clear();
                    _debtorFilterCubit.searchDebtors("");
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                    decoration: BoxDecoration(
                      border: index < (availableDebtors.length > 10 ? 9 : availableDebtors.length - 1)
                          ? Border(bottom: BorderSide(color: Colors.grey.shade200))
                          : null,
                    ),
                    child: Row(
                      children: [
                        Icon(Icons.person, size: 14, color: Colors.orange.shade400),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                name,
                                style: TextStyle(
                                  fontSize: _fontSize,
                                  fontWeight: FontWeight.w500,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                              Text(
                                '${global.language("code")}: ${debtor.code}',
                                style: TextStyle(
                                  fontSize: _fontSize - 1,
                                  color: Colors.grey.shade600,
                                ),
                              ),
                            ],
                          ),
                        ),
                        Icon(Icons.add_circle_outline, size: 16, color: Colors.green.shade400),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        if (hasSearchText && availableDebtors.isEmpty && !state.isLoading)
          Container(
            margin: const EdgeInsets.only(top: 4),
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.grey.shade100,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(Icons.info_outline, size: 14, color: Colors.grey.shade500),
                const SizedBox(width: 6),
                Text(
                  '${global.language("debtor_not_found")} "$searchText"',
                  style: TextStyle(fontSize: _fontSize - 1, color: Colors.grey.shade600),
                ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildSelectedDebtorChips(DebtorFilterState state) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(6),
      decoration: BoxDecoration(
        color: Colors.orange.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.orange.shade200),
      ),
      child: Wrap(
        spacing: 6,
        runSpacing: 4,
        children: _filterDebtorCodes.map((code) {
          String name = code;
          try {
            final debtor = state.debtors.firstWhere((d) => d.code == code);
            if (debtor.names.isNotEmpty) {
              name = debtor.names.first.name;
            }
          } catch (_) {
            // ใช้ code เป็นชื่อถ้าหาไม่เจอ
          }

          return Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: Colors.orange.shade300),
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
                Icon(Icons.person, size: 12, color: Colors.orange.shade600),
                const SizedBox(width: 4),
                Text(
                  name.length > 15 ? '${name.substring(0, 15)}...' : name,
                  style: TextStyle(
                    fontSize: _fontSize - 1,
                    color: Colors.orange.shade800,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(width: 4),
                InkWell(
                  onTap: () => _removeDebtorFilter(code),
                  child: Container(
                    padding: const EdgeInsets.all(2),
                    decoration: BoxDecoration(
                      color: Colors.red.shade100,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(Icons.close, size: 12, color: Colors.red.shade600),
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
    if (_isLoading && _qtList.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_qtList.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inbox_outlined, size: 64, color: Colors.grey[400]),
            SizedBox(height: 16),
            Text(
              _hasActiveFilters ? global.language('no_data_matching_criteria') : global.language('no_data'),
              style: TextStyle(fontSize: _fontSize + 2, color: Colors.grey[600]),
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

    final hasMoreData = _qtList.length < _totalRecords;
    final showLoadMoreSpinner = _isLoading && hasMoreData;

    return RefreshIndicator(
      onRefresh: () async => _loadData(),
      child: ListView.builder(
        controller: _scrollController,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: _qtList.length + (showLoadMoreSpinner ? 1 : 0),
        itemBuilder: (context, index) {
          if (index >= _qtList.length) {
            return const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator()),
            );
          }
          return _buildQTCard(_qtList[index]);
        },
      ),
    );
  }

  Widget _buildQTCard(TransactionModel qt) {
    return QTListCard(
      qt: qt,
      isSelected: _selectedGuid == qt.guidfixed,
      fontSize: _fontSize,
      approvalStatus: _approvalStatusMap[qt.docno],
      onTap: () => _onQTSelected(qt),
    );
  }

  /// โหลด document เต็มจาก MongoDB แล้วส่งไป edit screen
  Future<void> _onQTSelected(TransactionModel qt) async {
    setState(() => _selectedGuid = qt.guidfixed ?? '');

    if (qt.guidfixed == null || qt.guidfixed!.isEmpty) {
      _passDocumentToCallback(qt);
      return;
    }

    try {
      const String collection = 'transactionQuotation';
      MongoGetDataPayloadModel docPayLoad = MongoGetDataPayloadModel(
        shopid: global.getShopId(),
        collection: collection,
        guidfixed: qt.guidfixed!,
      );

      AppLogger.info('[QT List] Loading full document from MongoDB: ${qt.docno}');
      final response = await global.goApiPost(
        global.goApiUrlPath("mongogetdata"),
        docPayLoad.toJson(),
      );

      String rawData = json.encoder.convert(response);
      GoApiQueryOneResultResponse data = GoApiQueryOneResultResponse.fromMap(json.decode(rawData));

      TransactionModel fullDocument = TransactionModel.fromJson(data.result);
      global.parseTransactionJsonData(fullDocument);

      // รักษาค่า flags จาก PostgreSQL
      fullDocument.isref = qt.isref;
      fullDocument.islocked = qt.islocked;
      fullDocument.isclosed = qt.isclosed;
      fullDocument.iscancel = qt.iscancel;
      fullDocument.isdelete = qt.isdelete;
      fullDocument.iscomparedsuccess = qt.iscomparedsuccess;

      AppLogger.info('[QT List] Loaded full document with ${fullDocument.details?.length ?? 0} details (isref: ${fullDocument.isref}, isdelete: ${fullDocument.isdelete}, iscancel: ${fullDocument.iscancel})');
      _passDocumentToCallback(fullDocument);
    } catch (e) {
      AppLogger.error('[QT List] Error loading full document: $e');
      _passDocumentToCallback(qt);
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
