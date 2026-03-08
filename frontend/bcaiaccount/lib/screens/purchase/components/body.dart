import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/trans/trans_bloc.dart';
import 'package:smlaicloud/components/background_main.dart';
import 'package:smlaicloud/components/textfield_search.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/trans_repository.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'purchase_filter_panel.dart';
import 'purchase_search_dialog.dart';

class Body extends StatefulWidget {
  const Body({super.key});

  @override
  State<Body> createState() => _BodyState();
}

class _BodyState extends State<Body> {
  // รายการข้อมูลจาก API
  List<TransactionModel> _transactionList = [];

  // เงื่อนไขการค้นหา
  PurchaseSearchCondition _searchCondition = PurchaseSearchCondition();

  // สถานะการโหลด
  bool _isLoading = false;

  // Pagination
  int _offset = 0;
  final int _limit = 100;
  int _totalRecord = 0;

  // ขนาด font ปรับได้
  double _fontSize = 12.0;
  static const double _minFontSize = 10.0;
  static const double _maxFontSize = 18.0;

  // การเรียงลำดับวันที่ (0=เก่าก่อน, 1=ใหม่ก่อน)
  int _dateOrder = 1; // ค่าเริ่มต้น: ใหม่ก่อน

  // Debouncer สำหรับ search
  final _debouncer = global.Debouncer(500);

  // Cache สถานะการอนุมัติ PO
  final Map<String, POApprovalStatusModel?> _approvalStatusCache = {};

  // ScrollController สำหรับ infinite scroll
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
    // โหลดข้อมูลครั้งแรก
    _loadData();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  /// ตรวจจับ scroll เพื่อโหลดข้อมูลเพิ่ม
  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 200) {
      // ถ้า scroll ใกล้ถึงล่างสุด ให้โหลดข้อมูลเพิ่ม
      if (!_isLoading && _transactionList.length < _totalRecord) {
        _loadMoreData();
      }
    }
  }

  /// โหลดข้อมูลจาก API
  void _loadData() {
    setState(() {
      _isLoading = true;
      _offset = 0;
      _transactionList.clear();
    });

    // เรียกใช้ TransBloc เพื่อโหลดข้อมูล
    context.read<TransBloc>().add(
          TransLoad(
            offset: _offset,
            limit: _limit,
            search: _searchCondition.searchText,
            type: global.TransactionTypeEnum.purchaseorder,
            custcode: _searchCondition.creditorCode,
            ispos: '', // ไม่กรองตาม POS
            dateorder: _dateOrder, // 0=เก่าก่อน, 1=ใหม่ก่อน
          ),
        );
  }

  /// โหลดข้อมูลเพิ่มเติม (pagination)
  void _loadMoreData() {
    if (_isLoading) return;

    setState(() {
      _isLoading = true;
    });

    context.read<TransBloc>().add(
          TransLoad(
            offset: _transactionList.length,
            limit: _limit,
            search: _searchCondition.searchText,
            type: global.TransactionTypeEnum.purchaseorder,
            custcode: _searchCondition.creditorCode,
            ispos: '', // ไม่กรองตาม POS
            dateorder: _dateOrder,
          ),
        );
  }

  /// สลับการเรียงลำดับวันที่
  void _toggleDateOrder() {
    setState(() {
      _dateOrder = _dateOrder == 1 ? 0 : 1;
    });
    _loadData();
  }

  /// ค้นหาจาก TextField แบบ quick search
  void _quickSearch(String keyword) {
    _debouncer.run(() {
      setState(() {
        _searchCondition.searchText = keyword;
      });
      _loadData();
    });
  }

  /// เปิด dialog ค้นหาขั้นสูง
  void _openSearchDialog() async {
    await PurchaseSearchDialog.show(
      context,
      initialCondition: _searchCondition,
      onSearch: (condition) {
        setState(() {
          _searchCondition = condition;
        });
        _loadData();
      },
    );
  }

  /// รีเฟรชข้อมูล
  void _refreshData() {
    _loadData();
  }

  /// ล้างเงื่อนไขทั้งหมด
  void _clearConditions() {
    setState(() {
      _searchCondition.clear();
    });
    _loadData();
  }

  /// เพิ่มขนาด font
  void _increaseFontSize() {
    setState(() {
      if (_fontSize < _maxFontSize) {
        _fontSize += 1.0;
      }
    });
  }

  /// ลดขนาด font
  void _decreaseFontSize() {
    setState(() {
      if (_fontSize > _minFontSize) {
        _fontSize -= 1.0;
      }
    });
  }

  /// กรองข้อมูลตามเงื่อนไขเพิ่มเติม (client-side)
  List<TransactionModel> _applyClientSideFilters(
      List<TransactionModel> data) {
    var results = data;

    // กรองตามช่วงวันที่
    if (_searchCondition.fromDate != null) {
      results = results.where((item) {
        final docDate = DateTime.tryParse(item.docdatetime);
        if (docDate == null) return true;
        return docDate.isAfter(_searchCondition.fromDate!) ||
            docDate.isAtSameMomentAs(_searchCondition.fromDate!);
      }).toList();
    }

    if (_searchCondition.toDate != null) {
      results = results.where((item) {
        final docDate = DateTime.tryParse(item.docdatetime);
        if (docDate == null) return true;
        return docDate.isBefore(
                _searchCondition.toDate!.add(const Duration(days: 1))) ||
            docDate.isAtSameMomentAs(_searchCondition.toDate!);
      }).toList();
    }

    // กรองตามช่วงจำนวนเงิน
    if (_searchCondition.minAmount != null) {
      results = results.where((item) {
        return item.totalamount >= _searchCondition.minAmount!;
      }).toList();
    }

    if (_searchCondition.maxAmount != null) {
      results = results.where((item) {
        return item.totalamount <= _searchCondition.maxAmount!;
      }).toList();
    }

    // กรองตามเลขที่เอกสาร
    if (_searchCondition.docNo.isNotEmpty) {
      results = results.where((item) {
        return item.docno
            .toLowerCase()
            .contains(_searchCondition.docNo.toLowerCase());
      }).toList();
    }

    // กรองตามชื่อผู้ขาย (Full-text search ภาษาไทย)
    if (_searchCondition.creditorName.isNotEmpty) {
      results = results.where((item) {
        final custName = (item.custnames?.isNotEmpty ?? false)
            ? item.custnames!.first.name
            : '';
        return custName
            .toLowerCase()
            .contains(_searchCondition.creditorName.toLowerCase());
      }).toList();
    }

    return results;
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => TransBloc(transRepository: TransRepository()),
      child: BlocListener<TransBloc, TransState>(
        listener: (context, state) {
          if (state is TransLoadSuccess) {
            setState(() {
              _isLoading = false;
              if (_offset == 0) {
                _transactionList = state.trans;
              } else {
                _transactionList.addAll(state.trans);
              }
              _totalRecord = state.totalRecord;
              _offset = _transactionList.length;
            });
          } else if (state is TransLoadFailed) {
            setState(() {
              _isLoading = false;
            });
            global.showErrorSnackBar(context, '${global.language("error_occurred")}: ${state.message}');
          } else if (state is TransInProgress) {
            // กำลังโหลด - state handled
          }
        },
        child: BackgroundMain(
          child: Column(
            children: [
              // Filter Panel - แสดงเงื่อนไขการค้นหา
              PurchaseFilterPanel(
                condition: _searchCondition,
                onEditPressed: _openSearchDialog,
                onRefreshPressed: _refreshData,
                onClearPressed:
                    _searchCondition.hasConditions ? _clearConditions : null,
              ),

              // Quick Search และ Font Size Controls
              Padding(
                padding: EdgeInsets.symmetric(horizontal: 4.0),
                child: Row(
                  children: [
                    // Quick Search TextField
                    Expanded(
                      child: Card(
                        margin: const EdgeInsets.all(2),
                        child: TextFieldSearch(
                          onChange: (value) {
                            _quickSearch(value);
                          },
                        ),
                      ),
                    ),
                    // ปุ่มปรับขนาด font
                    Card(
                      margin: EdgeInsets.all(2),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          IconButton(
                            icon: const Icon(Icons.text_decrease, size: 18),
                            onPressed:
                                _fontSize > _minFontSize ? _decreaseFontSize : null,
                            tooltip: global.language("decrease_font_size"),
                            padding: const EdgeInsets.all(4),
                            constraints: const BoxConstraints(),
                          ),
                          Text(
                            '${_fontSize.toInt()}',
                            style: const TextStyle(fontSize: 12),
                          ),
                          IconButton(
                            icon: const Icon(Icons.text_increase, size: 18),
                            onPressed:
                                _fontSize < _maxFontSize ? _increaseFontSize : null,
                            tooltip: global.language("increase_font_size"),
                            padding: const EdgeInsets.all(4),
                            constraints: const BoxConstraints(),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              // แสดงจำนวนรายการ และปุ่มสลับการเรียงลำดับ
              Padding(
                padding:
                    EdgeInsets.symmetric(horizontal: 8.0, vertical: 2.0),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      '${global.language("total")} ${_applyClientSideFilters(_transactionList).length} ${global.language("items")}',
                      style: TextStyle(
                        color: Colors.grey.shade600,
                        fontSize: 11,
                      ),
                    ),
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        // ปุ่มสลับการเรียงลำดับวันที่
                        InkWell(
                          onTap: _toggleDateOrder,
                          borderRadius: BorderRadius.circular(4),
                          child: Padding(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 6, vertical: 2),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  _dateOrder == 1
                                      ? Icons.arrow_downward
                                      : Icons.arrow_upward,
                                  size: 14,
                                  color: Colors.blue.shade700,
                                ),
                                SizedBox(width: 2),
                                Text(
                                  _dateOrder == 1 ? global.language("newest_first") : global.language("oldest_first"),
                                  style: TextStyle(
                                    color: Colors.blue.shade700,
                                    fontSize: 11,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        if (_isLoading)
                          const SizedBox(
                            width: 14,
                            height: 14,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                      ],
                    ),
                  ],
                ),
              ),

              // รายการข้อมูล
              Expanded(
                child: _buildListView(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// สร้าง ListView แสดงรายการ
  Widget _buildListView() {
    final filteredList = _applyClientSideFilters(_transactionList);

    if (_transactionList.isEmpty && _isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (filteredList.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.search_off, size: 48, color: Colors.grey.shade400),
            SizedBox(height: 8),
            Text(
              global.language("no_matching_data"),
              style: TextStyle(
                fontSize: 14,
                color: Colors.grey.shade600,
              ),
            ),
            SizedBox(height: 4),
            TextButton.icon(
              onPressed: _clearConditions,
              icon: Icon(Icons.clear_all, size: 16),
              label: Text(global.language("clear_conditions"), style: TextStyle(fontSize: 12)),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(4),
      itemCount: filteredList.length + (_isLoading ? 1 : 0),
      itemBuilder: (context, index) {
        if (index >= filteredList.length) {
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(8.0),
              child: CircularProgressIndicator(),
            ),
          );
        }
        return _buildListItem(filteredList[index]);
      },
    );
  }

  /// สร้าง item ในรายการ
  Widget _buildListItem(TransactionModel item) {
    // Debug: ตรวจสอบค่า creatorname ใน body.dart
    AppLogger.debug(
        '📋 [body.dart] Doc: ${item.docno} | creatorname: "${item.creatorname}" | creatorcode: "${item.creatorcode}" | createdat: "${item.createdat}"');

    // แปลงวันที่
    final docDateTime = DateTime.tryParse(item.docdatetime) ?? DateTime.now();
    final formattedDate = global.formatThaiDateTime(
      dateTime: docDateTime.toLocal(),
      showTime: false,
    );

    // ดึงชื่อลูกค้า/ผู้ขาย
    final custName =
        (item.custnames?.isNotEmpty ?? false) ? item.custnames!.first.name : '-';

    // กำหนดสีตามสถานะ
    final isCancel = item.iscancel == true;
    final isClosed = item.isclosed == true;

    return Card(
      key: ValueKey(item.guidfixed),
      elevation: 1,
      margin: const EdgeInsets.only(bottom: 4),
      color: isCancel
          ? Colors.red.shade50
          : isClosed
              ? Colors.green.shade50
              : null,
      child: InkWell(
        onTap: () {
          // TODO: นำไปหน้ารายละเอียด
        },
        borderRadius: BorderRadius.circular(4),
        child: Padding(
          padding: const EdgeInsets.all(6),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // แถวบน: วันที่, เลขที่เอกสาร, สถานะ
              Row(
                children: [
                  // วันที่ - เพิ่มความกว้าง
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    constraints: const BoxConstraints(minWidth: 90),
                    decoration: BoxDecoration(
                      color: isCancel
                          ? Colors.red.shade100
                          : Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      formattedDate,
                      style: TextStyle(
                        color: isCancel
                            ? Colors.red.shade700
                            : Colors.blue.shade700,
                        fontWeight: FontWeight.bold,
                        fontSize: _fontSize,
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  // เลขที่เอกสาร
                  Expanded(
                    child: Text(
                      item.docno,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: _fontSize,
                        color: isCancel ? Colors.red : null,
                        decoration:
                            isCancel ? TextDecoration.lineThrough : null,
                      ),
                    ),
                  ),
                  // สถานะยกเลิก/ปิด
                  if (isCancel)
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.red.shade100,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(
                        global.language("cancelled"),
                        style: TextStyle(
                          color: Colors.red.shade700,
                          fontWeight: FontWeight.bold,
                          fontSize: _fontSize - 2,
                        ),
                      ),
                    )
                  else if (isClosed)
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.green.shade100,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(
                        global.language("closed"),
                        style: TextStyle(
                          color: Colors.green.shade700,
                          fontWeight: FontWeight.bold,
                          fontSize: _fontSize - 2,
                        ),
                      ),
                    ),
                  // สถานะการอนุมัติ
                  if (!isCancel)
                    _POApprovalStatusBadge(
                      docNo: item.docno,
                      fontSize: _fontSize,
                      cache: _approvalStatusCache,
                      onStatusLoaded: (status) {
                        if (mounted) setState(() {});
                      },
                    ),
                ],
              ),

              const SizedBox(height: 4),

              // แถวกลาง: ชื่อผู้ขาย
              Text(
                custName,
                style: TextStyle(
                  fontSize: _fontSize + 2,
                  fontWeight: FontWeight.w500,
                  color: isCancel ? Colors.grey : null,
                ),
              ),

              const SizedBox(height: 4),

              // แถวล่าง: จำนวนสินค้า, หมายเหตุ, ยอดรวม
              Row(
                children: [
                  // จำนวนสินค้า
                  Icon(Icons.inventory_2,
                      size: _fontSize, color: Colors.grey.shade600),
                  const SizedBox(width: 2),
                  Text(
                    '${item.detailcount ?? 0} ${global.language("items")}',
                    style: TextStyle(
                      color: Colors.grey.shade700,
                      fontSize: _fontSize - 1,
                    ),
                  ),
                  const SizedBox(width: 8),
                  // หมายเหตุ
                  if ((item.description?.isNotEmpty ?? false)) ...[
                    Icon(Icons.notes, size: _fontSize, color: Colors.grey.shade600),
                    const SizedBox(width: 2),
                    Expanded(
                      child: Text(
                        item.description!,
                        style: TextStyle(
                          color: Colors.grey.shade700,
                          fontSize: _fontSize - 1,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ] else
                    const Spacer(),
                  // ยอดรวม
                  Text(
                    '฿${global.moneyFormat.format(item.totalamount)}',
                    style: TextStyle(
                      color: isCancel
                          ? Colors.grey
                          : Colors.orange.shade700,
                      fontWeight: FontWeight.bold,
                      fontSize: _fontSize + 4,
                    ),
                  ),
                ],
              ),

              // แถวผู้สร้างเอกสาร (แสดงเฉพาะเมื่อมีข้อมูล) พร้อมวันที่เอกสาร
              if (item.creatorname?.isNotEmpty ?? false) ...[
                const SizedBox(height: 2),
                Row(
                  children: [
                    Icon(Icons.person_outline,
                        size: _fontSize - 1, color: Colors.grey.shade500),
                    const SizedBox(width: 2),
                    Text(
                      '${global.language("created_by")}: ${item.creatorname}',
                      style: TextStyle(
                        color: Colors.grey.shade500,
                        fontSize: _fontSize - 2,
                        fontStyle: FontStyle.italic,
                      ),
                    ),
                    const SizedBox(width: 4),
                    // แสดงวันที่สร้างเอกสาร (created_at จาก server UTC แปลงเป็น local)
                    if (item.createdat != null && item.createdat!.isNotEmpty)
                      Text(
                        global.formatThaiDateTime(
                          dateTime: (DateTime.tryParse(item.createdat!) ?? DateTime.now()).toLocal(),
                          showTime: true,
                        ),
                        style: TextStyle(
                          color: Colors.grey.shade400,
                          fontSize: _fontSize - 2,
                        ),
                      ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

/// Widget แสดงสถานะการอนุมัติ PO แบบ async
class _POApprovalStatusBadge extends StatefulWidget {
  final String docNo;
  final double fontSize;
  final Map<String, POApprovalStatusModel?> cache;
  final Function(POApprovalStatusModel?) onStatusLoaded;

  const _POApprovalStatusBadge({
    required this.docNo,
    required this.fontSize,
    required this.cache,
    required this.onStatusLoaded,
  });

  @override
  State<_POApprovalStatusBadge> createState() => _POApprovalStatusBadgeState();
}

class _POApprovalStatusBadgeState extends State<_POApprovalStatusBadge> {
  bool _isLoading = false;
  POApprovalStatusModel? _status;
  bool _hasLoaded = false;

  @override
  void initState() {
    super.initState();
    _loadStatus();
  }

  Future<void> _loadStatus() async {
    // ตรวจสอบ cache ก่อน
    if (widget.cache.containsKey(widget.docNo)) {
      _status = widget.cache[widget.docNo];
      _hasLoaded = true;
      return;
    }

    if (_isLoading) return;

    setState(() => _isLoading = true);

    try {
      final result = await ApprovalApiService.getPOApprovalStatus(widget.docNo);
      if (mounted) {
        setState(() {
          _isLoading = false;
          _hasLoaded = true;
          if (result.isSuccess && result.isFound) {
            _status = result.data;
          }
          // เก็บลง cache
          widget.cache[widget.docNo] = _status;
        });
        widget.onStatusLoaded(_status);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoading = false;
          _hasLoaded = true;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    // กำลังโหลด
    if (_isLoading && !_hasLoaded) {
      return const SizedBox(
        width: 12,
        height: 12,
        child: CircularProgressIndicator(strokeWidth: 1.5),
      );
    }

    // ไม่มีข้อมูลสถานะ (ยังไม่ได้ส่งอนุมัติ)
    if (_status == null) {
      return const SizedBox.shrink();
    }

    // แสดงสถานะตาม status
    final status = _status!.status;

    // ไม่แสดงถ้าเป็น draft
    if (status == POApprovalStatus.draft) {
      return const SizedBox.shrink();
    }

    final (Color bgColor, Color textColor, String text, IconData icon) =
        switch (status) {
      POApprovalStatus.pending => (
          Colors.orange.shade100,
          Colors.orange.shade800,
          global.language('pending_approval'),
          Icons.hourglass_empty
        ),
      POApprovalStatus.approved => (
          Colors.green.shade100,
          Colors.green.shade800,
          global.language('approval_approved'),
          Icons.check_circle
        ),
      POApprovalStatus.autoApproved => (
          Colors.green.shade50,
          Colors.green.shade700,
          global.language('approved_auto'),
          Icons.auto_awesome
        ),
      POApprovalStatus.rejected => (
          Colors.red.shade100,
          Colors.red.shade800,
          global.language('reject'),
          Icons.cancel
        ),
      POApprovalStatus.draft => (
          Colors.grey.shade100,
          Colors.grey.shade700,
          '',
          Icons.edit
        ),
    };

    return Container(
      margin: const EdgeInsets.only(left: 4),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: widget.fontSize - 2, color: textColor),
          const SizedBox(width: 2),
          Text(
            text,
            style: TextStyle(
              color: textColor,
              fontWeight: FontWeight.bold,
              fontSize: widget.fontSize - 2,
            ),
          ),
        ],
      ),
    );
  }
}
