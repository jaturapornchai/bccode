import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/trans/trans_bloc.dart';
import 'package:smlaicloud/bloc/creditor_filter/creditor_filter_cubit.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/purchaserequisition/components/pr_list_card.dart';

/// หน้าจอแสดงรายการใบสืบราคา (RFQ List)
/// โครงสร้างเหมือน PurchaseRequisitionListScreen — เปลี่ยน type เป็น rfq
class RFQListScreen extends StatefulWidget {
  const RFQListScreen({
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
  State<RFQListScreen> createState() => RFQListScreenState();
}

class RFQListScreenState extends State<RFQListScreen> with global.ThemeRefreshMixin {
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final FocusNode _searchFocusNode = FocusNode(skipTraversal: true);

  List<TransactionModel> _rfqList = [];
  String _searchText = '';
  bool _isLoading = false;
  int _totalRecords = 0;
  int _currentOffset = 0;
  String _selectedGuid = '';

  bool _isSearching = false;
  String _pendingSearchText = '';
  String _currentSearchRequest = '';

  Timer? _autoRefreshTimer;
  static const int _refreshInterval = 5;

  final Map<String, POApprovalStatusModel> _approvalStatusMap = {};
  Timer? _approvalPollingTimer;

  final _debouncer = global.Debouncer(500);

  DateTime? _filterFromDate;
  DateTime? _filterToDate;
  double? _filterMinAmount;
  double? _filterMaxAmount;
  List<String> _filterCreditorCodes = [];
  late CreditorFilterCubit _creditorFilterCubit;
  final TextEditingController _creditorSearchController = TextEditingController();
  Timer? _creditorSearchDebounce;
  final TextEditingController _minAmountController = TextEditingController();
  final TextEditingController _maxAmountController = TextEditingController();

  final double _fontSize = 12.0;

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
    if (_isLoading || _rfqList.length > 20) return;
    _loadData(silent: true);
  }

  void _loadData({bool silent = false, bool loadMore = false}) {
    if (_isLoading && _searchText == _currentSearchRequest && !loadMore) return;

    _currentSearchRequest = _searchText;

    if (!silent) {
      setState(() => _isLoading = true);
    }

    final offset = loadMore ? _rfqList.length : 0;
    _currentOffset = offset;

    String? fromDateStr;
    String? toDateStr;
    if (_filterFromDate != null) {
      fromDateStr = "${_filterFromDate!.year}-${_filterFromDate!.month.toString().padLeft(2, '0')}-${_filterFromDate!.day.toString().padLeft(2, '0')}";
    }
    if (_filterToDate != null) {
      toDateStr = "${_filterToDate!.year}-${_filterToDate!.month.toString().padLeft(2, '0')}-${_filterToDate!.day.toString().padLeft(2, '0')}";
    }

    context.read<TransBloc>().add(
      TransLoad(
        search: _searchText,
        type: global.TransactionTypeEnum.rfq,
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
      if (!_isLoading && _rfqList.length < _totalRecords) {
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
      if (_isLoading || _pendingSearchText != trimmedValue || _pendingSearchText == _currentSearchRequest) {
        setState(() => _isSearching = false);
        return;
      }
      setState(() {
        _searchText = _pendingSearchText;
        _isSearching = false;
        _rfqList = [];
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
        _rfqList = [];
        _isSearching = false;
      });
      _loadData();
    }
    _searchFocusNode.requestFocus();
  }

  Future<void> _loadApprovalStatuses() async {
    if (_rfqList.isEmpty) return;
    final listSnapshot = List<TransactionModel>.from(_rfqList);
    for (final rfq in listSnapshot) {
      if (rfq.docno.isEmpty) continue;
      try {
        final result = await ApprovalApiService.getPOApprovalStatus(rfq.docno);
        if (result.isSuccess && result.data != null && mounted) {
          setState(() {
            _approvalStatusMap[rfq.docno] = result.data!;
          });
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.debug('Error loading approval status for ${rfq.docno}: $e');
        }
      }
    }
  }

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
      _rfqList = [];
    });
    _loadData();
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
      _rfqList.removeWhere((rfq) => rfq.guidfixed == deletedGuid);
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
              _rfqList = state.trans;
            } else {
              _rfqList.addAll(state.trans);
            }
            _totalRecords = state.totalRecord;
          });
          _loadApprovalStatuses();

          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (_scrollController.hasClients &&
                _scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200 &&
                _rfqList.length < _totalRecords) {
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
          Expanded(child: _buildList()),
        ],
      ),
    );

    if (widget.isEmbedded) {
      return content;
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('request_for_quotation')),
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
    final notSubmitted = _rfqList.length - _approvalStatusMap.length;
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
          Row(
            children: [
              if (_hasActiveFilters)
                TextButton.icon(
                  icon: Icon(Icons.clear, size: 16),
                  label: Text(global.language('clear_filter'), style: TextStyle(fontSize: 12)),
                  onPressed: _clearAllFilters,
                ),
              if (widget.isEmbedded && widget.onClose != null)
                IconButton(icon: Icon(Icons.close), onPressed: widget.onClose),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSearchBar() {
    final isSearchingOrLoading = _isSearching || _isLoading;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
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
                  child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)),
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
                                _rfqList = [];
                              });
                              _loadData();
                            },
                    )
                  : null,
          border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
    );
  }

  Widget _buildList() {
    if (_isLoading && _rfqList.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_rfqList.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.description_outlined, size: 48, color: global.theme.iconSecondaryColor),
            const SizedBox(height: 8),
            Text(
              global.language('no_data'),
              style: TextStyle(color: global.theme.iconSecondaryColor),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      itemCount: _rfqList.length + (_rfqList.length < _totalRecords ? 1 : 0),
      itemBuilder: (context, index) {
        if (index >= _rfqList.length) {
          return const Padding(
            padding: EdgeInsets.all(16),
            child: Center(child: CircularProgressIndicator()),
          );
        }

        final rfq = _rfqList[index];
        final isSelected = rfq.guidfixed == _selectedGuid;
        final approvalStatus = _approvalStatusMap[rfq.docno];

        return PRListCard(
          pr: rfq,
          isSelected: isSelected,
          fontSize: _fontSize,
          approvalStatus: approvalStatus,
          onTap: () {
            setState(() => _selectedGuid = rfq.guidfixed ?? '');
            widget.onDocumentSelected?.call(rfq);
          },
        );
      },
    );
  }
}
