import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/modules/cart-system/bloc/product_search_cubit.dart';
import 'package:smlaicloud/modules/cart-system/bloc/product_search_state.dart';
import 'package:smlaicloud/modules/cart-system/services/pgsql_product_service.dart';

/// หน้าจอสำหรับ Rebuild/คำนวณสต็อกใหม่
/// รองรับ 3 รูปแบบ:
/// 1. Rebuild ทั้งหมด (drop + create ใหม่)
/// 2. คำนวณสต็อกอย่างเดียว (ทุก items)
/// 3. คำนวณเฉพาะบาง items (เลือกจากรายการ)
class RebuildStockScreen extends StatefulWidget {
  const RebuildStockScreen({super.key});

  @override
  State<RebuildStockScreen> createState() => _RebuildStockScreenState();
}

class _RebuildStockScreenState extends State<RebuildStockScreen>
    with global.ThemeRefreshMixin {
  // ตัวเลือกการประมวลผล
  int _selectedOption = 1; // 1 = rebuild all, 2 = calculate only, 3 = specific items

  // รายการสินค้าที่เลือก (สำหรับ option 3)
  final List<ProductModel> _selectedProducts = [];

  // สำหรับค้นหาสินค้า
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _listScrollController = ScrollController();
  final List<ProductModel> _productList = [];
  bool _isLoadingProducts = false;
  final _debouncer = global.Debouncer(500);

  // สถานะการส่งคำสั่ง
  bool _isSending = false;

  late ProductSearchCubit _searchCubit;

  @override
  void initState() {
    super.initState();
    _listScrollController.addListener(_onScrollList);
    _searchCubit = ProductSearchCubit(productService: PgSQLProductService());
  }

  @override
  void dispose() {
    _searchCubit.close();
    _searchController.dispose();
    _listScrollController.dispose();
    super.dispose();
  }

  /// โหลดรายการสินค้า
  void _loadProducts(String search) {
    if (search.isEmpty) return;
    setState(() => _isLoadingProducts = true);
    // Use ProductSearchCubit for robust search
    _searchCubit.universalSearch(
      keyword: search,
      shopId: global.getShopId(),
      limit: 100, // Load top 100 matches
    );
  }

  /// เมื่อ scroll ถึงล่างสุด ให้โหลดเพิ่ม
  void _onScrollList() {
    // SearchCubit returns all results at once (top N), no need for pagination here
    // if (_listScrollController.offset >= _listScrollController.position.maxScrollExtent &&
    //     !_listScrollController.position.outOfRange) {
    //   _loadProducts(_searchController.text);
    // }
  }

  /// เพิ่มสินค้าลงรายการที่เลือก
  void _addProduct(ProductModel product) {
    if (!_selectedProducts.any((p) => p.guidfixed == product.guidfixed)) {
      setState(() {
        _selectedProducts.add(product);
      });
    }
  }

  /// ลบสินค้าออกจากรายการที่เลือก
  void _removeProduct(ProductModel product) {
    setState(() {
      _selectedProducts.removeWhere((p) => p.guidfixed == product.guidfixed);
    });
  }

  /// ส่งคำสั่งประมวลผล พร้อมแสดง progress dialog แบบ real-time ผ่าน SSE
  Future<void> _sendRebuildCommand(int option) async {
    // Validate option 3
    if (option == 3 && _selectedProducts.isEmpty) {
      global.showSnackBar(context, Icon(Icons.warning, color: global.theme.onPrimaryColor), global.language("please_select_at_least_one_product"), global.theme.warningHighlightTextColor);
      return;
    }

    // แสดง confirmation dialog
    final confirmed = await _showConfirmDialog(option);
    if (confirmed != true) return;

    setState(() => _isSending = true);

    try {
      Map<String, dynamic> payload = {"shop_id": global.getShopId()};

      String title = "";

      switch (option) {
        case 1:
          payload["command_id"] = "rebuild";
          title = global.language("rebuild_all");
          break;
        case 2:
          payload["command_id"] = "rebuild";
          payload["create_database"] = false;
          title = global.language("calculate_stock_all_items");
          break;
        case 3:
          payload["command_id"] = "rebuild";
          payload["create_database"] = false;
          payload["item_code_list"] = _selectedProducts.map((p) => p.itemcode).toList();
          title = "${global.language("calculate_stock")} ${_selectedProducts.length} items";
          break;
        case 4:
          payload["command_id"] = "rebuild_document_flow";
          title = global.language("calculate_document_flow");
          break;
      }

      var jsonPayload = jsonEncode(payload);

      // ส่ง POST เพื่อเริ่ม rebuild และรับ job_id
      final jsonResult = await global.reportServicePost(jsonPayload);

      if (jsonResult['code'] != 200 && jsonResult['status'] != 'error') {
        // ไม่สำเร็จ — กรณี HTTP error
        if (mounted) {
          global.showSnackBar(context, Icon(Icons.error, color: global.theme.onPrimaryColor), "${global.language("error_occurred")}: ${jsonResult['message']}", global.theme.negativeHighlightTextColor);
        }
        return;
      }

      final jobId = jsonResult['job_id'] as String?;

      if (jobId == null || jobId.isEmpty) {
        // ไม่มี job_id — fallback แจ้งว่าส่งคำสั่งแล้ว (backend เก่า)
        if (mounted) {
          global.showSnackBar(context, Icon(Icons.send, color: global.theme.onPrimaryColor), global.language("command_sent_processing"), global.theme.infoHighlightTextColor);
        }
        return;
      }

      // แสดง Progress Dialog พร้อม SSE listener
      if (mounted) {
        await _showProgressDialog(jobId, title);
      }
    } catch (e) {
      if (mounted) {
        global.showSnackBar(context, Icon(Icons.error, color: global.theme.onPrimaryColor), "${global.language("error")}: $e", global.theme.negativeHighlightTextColor);
      }
    } finally {
      if (mounted) {
        setState(() => _isSending = false);
      }
    }
  }

  /// แสดง Progress Dialog ที่ฟัง SSE events แบบ real-time
  Future<void> _showProgressDialog(String jobId, String title) async {
    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) => _RebuildProgressDialog(jobId: jobId, title: title),
    );
  }

  /// แสดง confirmation dialog
  Future<bool?> _showConfirmDialog(int option) {
    String title = "";
    String description = "";

    switch (option) {
      case 1:
        title = global.language("rebuild_all");
        description = global.language("rebuild_all_description");
        break;
      case 2:
        title = global.language("calculate_stock_all_items");
        description = global.language("calculate_stock_all_description");
        break;
      case 3:
        title = "${global.language("calculate_stock")} ${_selectedProducts.length} items";
        description = global.language("calculate_stock_selected_description");
        break;
      case 4:
        title = global.language("calculate_document_flow");
        description = global.language("calculate_document_flow_description");
        break;
    }

    return showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.warning_amber, color: global.theme.warningHighlightTextColor),
            const SizedBox(width: 10),
            Text(title),
          ],
        ),
        content: Text(description),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: Text(global.language("cancel"))),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.warningHighlightTextColor, foregroundColor: global.theme.onPrimaryColor),
            child: Text(global.language("confirm")),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language("config_rebuild_stock_test")),
      ),
      body: Row(
        children: [
          // ส่วนซ้าย - ตัวเลือกและรายการที่เลือก
          Expanded(flex: 4, child: _buildLeftPanel()),
          // ส่วนขวา - ค้นหาสินค้า (แสดงเฉพาะ option 3)
          if (_selectedOption == 3) Expanded(flex: 6, child: _buildProductSearchPanel()),
        ],
      ),
    );
  }

  /// Panel ด้านซ้าย - ตัวเลือกและรายการที่เลือก
  Widget _buildLeftPanel() {
    return Container(
      color: global.theme.surfaceColor,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // หัวข้อ
          Container(
            width: double.infinity,
            padding: EdgeInsets.all(16),
            color: global.theme.cardColor,
            child: Text(
              global.language("select_processing_type"),
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: global.theme.textColor),
            ),
          ),
          const SizedBox(height: 8),

          // Option 1: Rebuild ทั้งหมด
          _buildOptionCard(
            title: global.language("rebuild_all"),
            subtitle: global.language("rebuild_all_subtitle"),
            icon: Icons.refresh,
            color: global.theme.warningHighlightTextColor,
            buttonLabel: global.language("start_rebuild"),
            onPressed: _isSending ? null : () => _sendRebuildCommand(1),
          ),

          // Option 2: คำนวณสต็อกอย่างเดียว
          _buildOptionCard(
            title: global.language("calculate_stock_only"),
            subtitle: global.language("calculate_stock_only_subtitle"),
            icon: Icons.calculate,
            color: global.theme.primaryColor,
            buttonLabel: global.language("start_calculate"),
            onPressed: _isSending ? null : () => _sendRebuildCommand(2),
          ),

          // Option 3: คำนวณเฉพาะบาง items
          _buildOptionCard(
            title: global.language("calculate_specific_items"),
            subtitle: global.language("select_products_to_calculate"),
            icon: Icons.checklist,
            color: global.theme.positiveHighlightTextColor,
            buttonLabel: _selectedOption == 3 ? global.language("hide") : global.language("select_products"),
            onPressed: _isSending ? null : () => setState(() {
              _selectedOption = _selectedOption == 3 ? 0 : 3;
            }),
          ),

          // Option 4: คำนวณ flow เอกสาร
          _buildOptionCard(
            title: global.language("calculate_document_flow"),
            subtitle: global.language("calculate_document_flow_subtitle"),
            icon: Icons.account_tree,
            color: global.theme.secondaryColor,
            buttonLabel: global.language("start_calculate"),
            onPressed: _isSending ? null : () => _sendRebuildCommand(4),
          ),

          // รายการสินค้าที่เลือก (แสดงเฉพาะ option 3)
          if (_selectedOption == 3) ...[
            SizedBox(height: 8),
            Container(
              width: double.infinity,
              padding: EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              color: global.theme.cardColor,
              child: Row(
                children: [
                  Icon(Icons.shopping_cart, color: global.theme.positiveHighlightTextColor),
                  SizedBox(width: 8),
                  Text(
                    "${global.language("selected_products")} (${_selectedProducts.length} ${global.language("items")})",
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: global.theme.textColor),
                  ),
                  const Spacer(),
                  if (_selectedProducts.isNotEmpty) ...[
                    TextButton(
                      onPressed: () => setState(() => _selectedProducts.clear()),
                      child: Text(global.language("clear_all"), style: TextStyle(color: global.theme.negativeHighlightTextColor)),
                    ),
                    SizedBox(width: 8),
                    ElevatedButton.icon(
                      onPressed: _isSending ? null : () => _sendRebuildCommand(3),
                      icon: Icon(Icons.play_arrow, size: 18),
                      label: Text("${global.language("start_calculate")} ${_selectedProducts.length} items"),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: global.theme.positiveHighlightTextColor,
                        foregroundColor: global.theme.onPrimaryColor,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            Expanded(
              child: _selectedProducts.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.inbox, size: 48, color: global.theme.iconSecondaryColor),
                          SizedBox(height: 8),
                          Text(global.language("no_products_selected"), style: TextStyle(color: global.theme.textSecondaryColor)),
                          SizedBox(height: 4),
                          Text(global.language("search_and_select_from_right"), style: TextStyle(fontSize: 12, color: global.theme.formHintColor)),
                        ],
                      ),
                    )
                  : ListView.builder(
                      itemCount: _selectedProducts.length,
                      itemBuilder: (context, index) {
                        final product = _selectedProducts[index];
                        return Card(
                          margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          child: ListTile(
                            dense: true,
                            leading: CircleAvatar(
                              backgroundColor: global.theme.positiveHighlightColor,
                              child: Text("${index + 1}", style: TextStyle(color: global.theme.positiveHighlightTextColor, fontSize: 12)),
                            ),
                            title: Text(product.itemcode, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                            subtitle: Text(global.packName(product.names ?? []), maxLines: 1, overflow: TextOverflow.ellipsis, style: TextStyle(fontSize: 12)),
                            trailing: IconButton(
                              icon: Icon(Icons.close, color: global.theme.negativeHighlightTextColor, size: 20),
                              onPressed: () => _removeProduct(product),
                            ),
                          ),
                        );
                      },
                    ),
            ),
          ],
        ],
      ),
    );
  }

  /// สร้าง Option Card พร้อมปุ่ม action แยกในแต่ละ card
  Widget _buildOptionCard({
    required String title,
    required String subtitle,
    required IconData icon,
    required Color color,
    required String buttonLabel,
    required VoidCallback? onPressed,
  }) {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: Container(
        padding: EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: global.theme.dividerBorderColor),
        ),
        child: Row(
          children: [
            Container(
              padding: EdgeInsets.all(10),
              decoration: BoxDecoration(color: color.withAlpha(30), borderRadius: BorderRadius.circular(10)),
              child: Icon(icon, color: color, size: 24),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: global.theme.textColor)),
                  const SizedBox(height: 2),
                  Text(subtitle, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
                ],
              ),
            ),
            const SizedBox(width: 8),
            ElevatedButton.icon(
              onPressed: onPressed,
              icon: Icon(Icons.play_arrow, size: 18),
              label: Text(buttonLabel),
              style: ElevatedButton.styleFrom(
                backgroundColor: color,
                foregroundColor: global.theme.onPrimaryColor,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Panel ค้นหาสินค้า (ด้านขวา)
  Widget _buildProductSearchPanel() {
    return BlocListener<ProductSearchCubit, ProductSearchState>(
      bloc: _searchCubit,
      listener: (context, state) {
        if (state is ProductSearchLoaded) {
          setState(() {
            _isLoadingProducts = false;
            _productList.clear(); // Clear old results

            // Map ProductSearchModel to ProductModel
            for (var searchResult in state.products) {
              // Create a minimal ProductModel for display
              final product = ProductModel(
                guidfixed: searchResult.itemCode, // Use itemCode as ID
                itemcode: searchResult.itemCode,
                barcodes: [searchResult.barcode],
                names: [LanguageDataModel(code: 'th', name: searchResult.name0)],
                units: [
                  ProductUnitModel(
                    unitcode: searchResult.unitCode,
                    names: [LanguageDataModel(code: 'th', name: searchResult.unitName)],
                    divider: searchResult.unitDive,
                    stand: searchResult.unitStand,
                    xorder: 0,
                    stockcount: true,
                  ),
                ],
              );

              // Add only unique itemcodes
              if (!_productList.any((p) => p.itemcode == product.itemcode)) {
                _productList.add(product);
              }
            }
          });
        } else if (state is ProductSearchError || state is ProductNotFound) {
          setState(() {
            _isLoadingProducts = false;
            _productList.clear();
          });
        }
      },
      child: Container(
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          border: Border(left: BorderSide(color: global.theme.dividerBorderColor)),
        ),
        child: Column(
          children: [
            // Search bar
            Container(
              padding: EdgeInsets.all(12),
              color: global.theme.surfaceColor,
              child: TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  hintText: global.language("search_product_code_name_barcode"),
                  prefixIcon: Icon(Icons.search),
                  suffixIcon: _searchController.text.isNotEmpty
                      ? IconButton(
                          icon: Icon(Icons.clear),
                          onPressed: () {
                            _searchController.clear();
                            setState(() => _productList.clear());
                          },
                        )
                      : null,
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                  filled: true,
                  fillColor: global.theme.formFillColor,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                ),
                onChanged: (value) {
                  _debouncer.run(() {
                    setState(() => _productList.clear());
                    if (value.isNotEmpty) {
                      _loadProducts(value);
                    }
                  });
                },
              ),
            ),

            // Header
            Container(
              padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              color: global.theme.columnHeaderColor,
              child: Row(
                children: [
                  SizedBox(width: 40),
                  Expanded(
                    flex: 3,
                    child: Text(global.language("product_code"), style: TextStyle(fontWeight: FontWeight.bold)),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(global.language("product_name"), style: TextStyle(fontWeight: FontWeight.bold)),
                  ),
                  const SizedBox(width: 50),
                ],
              ),
            ),

            // Product list
            Expanded(
              child: _productList.isEmpty
                  ? Center(
                      child: _isLoadingProducts
                          ? CircularProgressIndicator()
                          : Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.search, size: 48, color: global.theme.iconSecondaryColor),
                                SizedBox(height: 8),
                                Text(global.language("type_to_search_product"), style: TextStyle(color: global.theme.textSecondaryColor)),
                              ],
                            ),
                    )
                  : ListView.builder(
                      controller: _listScrollController,
                      itemCount: _productList.length + (_isLoadingProducts ? 1 : 0),
                      itemBuilder: (context, index) {
                        if (index >= _productList.length) {
                          return const Center(
                            child: Padding(padding: EdgeInsets.all(16), child: CircularProgressIndicator()),
                          );
                        }

                        final product = _productList[index];
                        final isSelected = _selectedProducts.any((p) => p.guidfixed == product.guidfixed);

                        return InkWell(
                          onTap: () {
                            if (isSelected) {
                              _removeProduct(product);
                            } else {
                              _addProduct(product);
                            }
                          },
                          child: Container(
                            padding: EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                            decoration: BoxDecoration(
                              color: isSelected ? global.theme.positiveHighlightColor : (index % 2 == 0 ? global.theme.cardColor : global.theme.surfaceColor),
                              border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor)),
                            ),
                            child: Row(
                              children: [
                                // Checkbox
                                SizedBox(
                                  width: 40,
                                  child: Checkbox(
                                    value: isSelected,
                                    onChanged: (value) {
                                      if (value == true) {
                                        _addProduct(product);
                                      } else {
                                        _removeProduct(product);
                                      }
                                    },
                                    activeColor: global.theme.positiveHighlightTextColor,
                                  ),
                                ),
                                // Product code
                                Expanded(
                                  flex: 3,
                                  child: Text(
                                    product.itemcode,
                                    style: TextStyle(fontWeight: FontWeight.bold, color: isSelected ? global.theme.positiveHighlightTextColor : global.theme.textColor),
                                  ),
                                ),
                                // Product name
                                Expanded(flex: 5, child: Text(global.packName(product.names ?? []), maxLines: 2, overflow: TextOverflow.ellipsis)),
                                // Add button
                                SizedBox(
                                  width: 50,
                                  child: isSelected
                                      ? Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor)
                                      : IconButton(
                                          icon: Icon(Icons.add_circle_outline, color: global.theme.primaryColor),
                                          onPressed: () => _addProduct(product),
                                        ),
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Dialog แสดง progress ของ rebuild แบบ real-time ผ่าน SSE
class _RebuildProgressDialog extends StatefulWidget {
  final String jobId;
  final String title;

  const _RebuildProgressDialog({required this.jobId, required this.title});

  @override
  State<_RebuildProgressDialog> createState() => _RebuildProgressDialogState();
}

class _RebuildProgressDialogState extends State<_RebuildProgressDialog>
    with global.ThemeRefreshMixin {
  // สถานะ progress
  double _progress = 0.0;
  int _currentStep = 0;
  int _totalSteps = 0;
  String _currentStepName = '';
  String _detail = ''; // รายละเอียดย่อย เช่น "150/500 รายการ"
  String _status = 'connecting'; // connecting, running, completed, error
  String _errorMessage = '';
  final List<_LogEntry> _logs = [];
  final ScrollController _logScrollController = ScrollController();

  // เวลาเริ่มต้น + elapsed timer
  final DateTime _startTime = DateTime.now();
  Timer? _elapsedTimer;
  String _elapsedText = '00:00';

  StreamSubscription<Map<String, dynamic>>? _sseSubscription;

  @override
  void initState() {
    super.initState();
    _connectSSE();
    // อัพเดทเวลาที่ผ่านไปทุกวินาที
    _elapsedTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      final elapsed = DateTime.now().difference(_startTime);
      final minutes = elapsed.inMinutes.toString().padLeft(2, '0');
      final seconds = (elapsed.inSeconds % 60).toString().padLeft(2, '0');
      setState(() => _elapsedText = '$minutes:$seconds');
    });
  }

  @override
  void dispose() {
    _sseSubscription?.cancel();
    _elapsedTimer?.cancel();
    _logScrollController.dispose();
    super.dispose();
  }

  /// เชื่อมต่อ SSE stream
  void _connectSSE() {
    _sseSubscription = global.connectRebuildSSE(widget.jobId).listen(
      (event) {
        if (!mounted) return;

        final eventStatus = event['status'] as String? ?? '';
        final step = (event['step'] as num?)?.toInt() ?? 0;
        final totalSteps = (event['total_steps'] as num?)?.toInt() ?? 0;
        final stepName = event['step_name'] as String? ?? '';
        final progress = (event['progress'] as num?)?.toDouble() ?? 0.0;
        final message = event['message'] as String? ?? '';
        final detail = event['detail'] as String? ?? '';

        setState(() {
          // Handle "connected" event จาก SSE initial handshake
          if (eventStatus == 'connected') {
            _status = 'running';
            _currentStepName = global.language('connected_waiting_data');
            _detail = '';
            return;
          }

          _currentStep = step;
          _totalSteps = totalSteps;
          _currentStepName = stepName;
          _progress = progress;
          _detail = detail;

          if (eventStatus == 'running') {
            _status = 'running';
            // ถ้ามี detail แสดงว่าเป็น sub-progress ของ step เดิม — อัพเดท log ล่าสุดแทนการเพิ่มใหม่
            if (detail.isNotEmpty && _logs.isNotEmpty && _logs.last.message == stepName) {
              _logs.last.detail = detail;
            } else if (detail.isEmpty) {
              _addLog(stepName, LogStatus.running);
            }
          } else if (eventStatus == 'completed') {
            _status = 'completed';
            _elapsedTimer?.cancel(); // หยุด timer เมื่อเสร็จ
            _markPreviousLogsCompleted();
            _addLog('$stepName (รวม $_elapsedText)', LogStatus.completed);
          } else if (eventStatus == 'error') {
            _status = 'error';
            _elapsedTimer?.cancel();
            _errorMessage = message.isNotEmpty ? message : stepName;
            _addLog(_errorMessage, LogStatus.error);
          }
        });

        // scroll log ไปล่างสุด
        _scrollLogToBottom();
      },
      onError: (error) {
        if (!mounted) return;
        // ถ้า progress เกือบจบแล้ว (เหลือไม่เกิน 1 step) ถือว่าสำเร็จ
        // (connection อาจหลุดระหว่าง step สุดท้ายหรือหลัง completed event)
        if (_totalSteps > 0 && _currentStep >= _totalSteps - 1) {
          _finishAsCompleted();
          return;
        }
        setState(() {
          _status = 'error';
          _elapsedTimer?.cancel();
          _errorMessage = '${global.language("connection_error")}: $error';
          _addLog(_errorMessage, LogStatus.error);
        });
      },
      onDone: () {
        // stream จบ — ตรวจสอบว่าเสร็จจริงหรือ connection หลุด
        if (!mounted) return;
        if (_status == 'completed' || _status == 'error') return;
        // ถ้า progress เกือบจบแล้ว ถือว่าสำเร็จ
        if (_totalSteps > 0 && _currentStep >= _totalSteps - 1) {
          _finishAsCompleted();
          return;
        }
        if (_status == 'running' || _status == 'connecting') {
          setState(() {
            _status = 'error';
            _elapsedTimer?.cancel();
            _errorMessage = global.language('connection_lost');
            _addLog(_errorMessage, LogStatus.error);
          });
        }
      },
    );
  }

  /// ตั้งค่าสถานะเป็น completed (ใช้ร่วมกันจาก onError/onDone)
  void _finishAsCompleted() {
    setState(() {
      _status = 'completed';
      _progress = 1.0;
      _currentStepName = global.language('completed');
      _detail = '';
      _elapsedTimer?.cancel();
      _markPreviousLogsCompleted();
      _addLog('${global.language("completed")} (${global.language("total")} $_elapsedText)', LogStatus.completed);
    });
  }

  void _addLog(String message, LogStatus status) {
    if (message.isEmpty) return;
    // ถ้า log ตัวก่อนหน้าเป็น running ให้เปลี่ยนเป็น completed + คำนวณ duration
    if (_logs.isNotEmpty && _logs.last.status == LogStatus.running) {
      _logs.last.status = LogStatus.completed;
      _logs.last.duration = _calcDuration(_logs.last.startedAt);
    }
    _logs.add(_LogEntry(
      time: DateTime.now().toString().substring(11, 19),
      message: message,
      status: status,
      startedAt: DateTime.now(),
    ));
  }

  void _markPreviousLogsCompleted() {
    for (var log in _logs) {
      if (log.status == LogStatus.running) {
        log.status = LogStatus.completed;
        log.duration = _calcDuration(log.startedAt);
      }
    }
  }

  /// คำนวณ duration จากเวลาเริ่มต้น
  String _calcDuration(DateTime startedAt) {
    final elapsed = DateTime.now().difference(startedAt);
    if (elapsed.inMinutes > 0) {
      return '${elapsed.inMinutes}m ${elapsed.inSeconds % 60}s';
    }
    return '${elapsed.inSeconds}s';
  }

  void _scrollLogToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_logScrollController.hasClients) {
        _logScrollController.animateTo(
          _logScrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final isFinished = _status == 'completed' || _status == 'error';
    final progressPercent = (_progress * 100).toInt();

    return AlertDialog(
      title: Row(
        children: [
          _buildStatusIcon(),
          SizedBox(width: 10),
          Expanded(
            child: Text(
              _status == 'completed'
                  ? '${widget.title} - ${global.language("success")}'
                  : _status == 'error'
                      ? '${widget.title} - ${global.language("failed")}'
                      : '${global.language("processing")} ${widget.title}...',
              style: TextStyle(fontSize: 16),
            ),
          ),
        ],
      ),
      content: SizedBox(
        width: 600,
        height: 500,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Step info + elapsed time
            if (_totalSteps > 0) ...[
              Row(
                children: [
                  Icon(Icons.info_outline, color: global.theme.primaryColor, size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Step $_currentStep/$_totalSteps: $_currentStepName',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                        color: _status == 'error' ? global.theme.negativeHighlightTextColor : global.theme.primaryColor,
                      ),
                    ),
                  ),
                  // Elapsed time + percentage
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        '$progressPercent%',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: _status == 'completed' ? global.theme.positiveHighlightTextColor : global.theme.primaryColor,
                        ),
                      ),
                      Text(
                        _elapsedText,
                        style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
                      ),
                    ],
                  ),
                ],
              ),
              // แสดงรายละเอียดย่อย (sub-progress)
              if (_detail.isNotEmpty) ...[
                const SizedBox(height: 4),
                Padding(
                  padding: EdgeInsets.only(left: 26),
                  child: Text(
                    _detail,
                    style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor, fontStyle: FontStyle.italic),
                  ),
                ),
              ],
              const SizedBox(height: 12),

              // Progress bar
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: LinearProgressIndicator(
                  value: _progress,
                  minHeight: 12,
                  backgroundColor: global.theme.dividerBorderColor,
                  valueColor: AlwaysStoppedAnimation<Color>(
                    _status == 'completed'
                        ? global.theme.positiveHighlightTextColor
                        : _status == 'error'
                            ? global.theme.negativeHighlightTextColor
                            : global.theme.infoHighlightTextColor,
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ] else ...[
              // กำลังเชื่อมต่อ
              const Center(child: CircularProgressIndicator()),
              SizedBox(height: 8),
              Center(child: Text('${global.language("connecting")}... ($_elapsedText)')),
              const SizedBox(height: 16),
            ],

            // Error message
            if (_status == 'error' && _errorMessage.isNotEmpty) ...[
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: global.theme.negativeHighlightColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: global.theme.negativeHighlightTextColor.withValues(alpha: 0.3)),
                ),
                child: Row(
                  children: [
                    Icon(Icons.error_outline, color: global.theme.negativeHighlightTextColor, size: 18),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        _errorMessage,
                        style: TextStyle(color: global.theme.negativeHighlightTextColor, fontSize: 12),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 12),
            ],

            // Log header
            Row(
              children: [
                Icon(Icons.list_alt, color: global.theme.iconSecondaryColor, size: 18),
                const SizedBox(width: 6),
                Text('Logs', style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: global.theme.textSecondaryColor)),
              ],
            ),
            const SizedBox(height: 6),

            // Log list
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: global.theme.dividerBorderColor),
                ),
                child: _logs.isEmpty
                    ? Center(child: Text('${global.language("waiting_for_data")}...', style: TextStyle(color: global.theme.iconSecondaryColor)))
                    : ListView.builder(
                        controller: _logScrollController,
                        padding: const EdgeInsets.all(8),
                        itemCount: _logs.length,
                        itemBuilder: (context, index) {
                          final log = _logs[index];
                          return Padding(
                            padding: EdgeInsets.symmetric(vertical: 2),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  '[${log.time}] ',
                                  style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor, fontFamily: 'monospace'),
                                ),
                                _buildLogIcon(log.status),
                                const SizedBox(width: 4),
                                Expanded(
                                  child: Row(
                                    children: [
                                      Flexible(
                                        child: Text(
                                          log.message,
                                          style: TextStyle(
                                            fontSize: 12,
                                            color: log.status == LogStatus.error ? global.theme.negativeHighlightTextColor : global.theme.textColor,
                                          ),
                                        ),
                                      ),
                                      if (log.detail.isNotEmpty) ...[
                                        const SizedBox(width: 6),
                                        Text(
                                          '(${log.detail})',
                                          style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor, fontStyle: FontStyle.italic),
                                        ),
                                      ],
                                      // แสดง duration ของแต่ละ step
                                      if (log.duration.isNotEmpty && log.status == LogStatus.completed) ...[
                                        const SizedBox(width: 6),
                                        Text(
                                          '[${log.duration}]',
                                          style: TextStyle(fontSize: 10, color: global.theme.positiveHighlightTextColor, fontWeight: FontWeight.w500),
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          );
                        },
                      ),
              ),
            ),
          ],
        ),
      ),
      actions: [
        if (isFinished)
          ElevatedButton(
            onPressed: () => Navigator.pop(context),
            style: ElevatedButton.styleFrom(
              backgroundColor: _status == 'completed' ? global.theme.positiveHighlightTextColor : global.theme.textColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            child: Text(global.language('close')),
          )
        else
          TextButton(
            onPressed: () {
              _sseSubscription?.cancel();
              Navigator.pop(context);
            },
            child: Text(global.language('hide')),
          ),
      ],
    );
  }

  Widget _buildStatusIcon() {
    switch (_status) {
      case 'completed':
        return Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor, size: 28);
      case 'error':
        return Icon(Icons.error, color: global.theme.negativeHighlightTextColor, size: 28);
      default:
        return SizedBox(
          width: 24,
          height: 24,
          child: CircularProgressIndicator(strokeWidth: 2.5, color: global.theme.warningHighlightTextColor),
        );
    }
  }

  Widget _buildLogIcon(LogStatus status) {
    switch (status) {
      case LogStatus.completed:
        return Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor, size: 14);
      case LogStatus.error:
        return Icon(Icons.cancel, color: global.theme.negativeHighlightTextColor, size: 14);
      case LogStatus.running:
        return SizedBox(
          width: 14,
          height: 14,
          child: CircularProgressIndicator(strokeWidth: 1.5, color: global.theme.infoHighlightTextColor),
        );
    }
  }
}

enum LogStatus { running, completed, error }

class _LogEntry {
  final String time;
  final String message;
  final DateTime startedAt;
  LogStatus status;
  String detail = ''; // รายละเอียดย่อย เช่น "150/500 รายการ"
  String duration = ''; // ระยะเวลาที่ใช้ เช่น "3s", "1m 20s"

  _LogEntry({
    required this.time,
    required this.message,
    required this.status,
    required this.startedAt,
  });
}
