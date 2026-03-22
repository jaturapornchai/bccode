import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/components/loading_overlay.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/screens/components/product_preview_screen.dart';
import 'package:smlaicloud/global.dart' as global;

/// หน้าจอค้นหาสินค้า + แสดง Preview รายละเอียดสินค้าด้านซ้าย
/// ใช้แทน BarcodeSearchScreen ในทุกหน้าจอ Transaction
/// รองรับ SplitView (desktop) และ single view (mobile)
class ProductSearchPreviewScreen extends StatefulWidget {
  final String word;
  final String screen;

  /// ถ้าส่ง callback นี้มา → เลือกสินค้าแล้วเรียก callback
  /// ถ้าไม่ส่ง (null) → เลือกแล้ว pop กลับเหมือนเดิม
  final Function(ProductBarcodeModel)? onProductAdded;

  /// callback สำหรับปิดจอค้นหา (ใช้เมื่ออยู่ใน Offstage แทน Navigator.pop)
  final VoidCallback? onClose;

  const ProductSearchPreviewScreen({
    super.key,
    required this.word,
    required this.screen,
    this.onProductAdded,
    this.onClose,
  });

  @override
  State<ProductSearchPreviewScreen> createState() =>
      _ProductSearchPreviewScreenState();
}

class _ProductSearchPreviewScreenState extends State<ProductSearchPreviewScreen>
    with global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  String searchItemType = "";
  List<ProductBarcodeModel> productBarcodeListData = [];
  String selectGuid = "";
  int _hoverIndex = -1;

  /// โหมดเพิ่มต่อเนื่อง — เลือกสินค้าแล้วไม่ปิดจอ
  bool _continuousAdd = false;
  final _debouncer = global.Debouncer(500);
  String isUseSubBarcodes = "";
  String isBom = "";
  bool loadingData = false;

  /// สลับด้าน preview ซ้าย/ขวา (persist ผ่าน SharedPreferences)
  late bool _previewOnLeft;

  /// สินค้าที่เลือก/hover อยู่ สำหรับแสดง preview
  ProductBarcodeModel? _previewProduct;

  /// รายการ barcode ทั้งหมดของ itemcode เดียวกัน (สำหรับแผนผังหน่วยนับ)
  List<ProductBarcodeModel> _previewAllBarcodes = [];

  @override
  void initState() {
    super.initState();
    _previewOnLeft = global.appConfig.getBool('product_preview_on_left') ?? true;
    listScrollController.addListener(_onScrollList);
    searchText = widget.word;
    searchController.text = searchText;

    _initSearchParams();
    _loadDataList(searchText);
  }

  void _initSearchParams() {
    if (widget.screen == "stockbalance") {
      searchItemType = "0,3,4";
      isBom = "all";
      isUseSubBarcodes = "notshowsubbarcodes";
    } else if (widget.screen == "subbarcodeSearch") {
      searchItemType = "0,1,2,3,4,5";
      isBom = "all";
      isUseSubBarcodes = "notshowsubbarcodes";
    } else {
      searchItemType = "";
      isBom = "all";
      isUseSubBarcodes = "all";
    }
  }

  void _loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    String shopsid = global.getShopId();
    final offset =
        (productBarcodeListData.isEmpty) ? 0 : productBarcodeListData.length;

    context.read<ProductBarcodeBloc>().add(
          ProductBarcodeLoadListSearch(
            offset: offset,
            limit: global.loadDataPerPage,
            search: search,
            itemtype: searchItemType,
            branchcode: "",
            businesstypecode: "",
            isbom: isBom,
            isusesubbarcodes: isUseSubBarcodes,
            shopsid: shopsid,
          ),
        );
  }

  void _onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      _loadDataList(searchText);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    searchFocusNode.dispose();
    super.dispose();
  }

  /// อัปเดตสินค้าที่แสดง preview
  void _updatePreview(ProductBarcodeModel product) {
    setState(() {
      _previewProduct = product;
      _previewAllBarcodes = productBarcodeListData
          .where((item) => item.itemcode == product.itemcode)
          .toList();
    });
  }

  /// ปิดจอค้นหา — ใช้ onClose callback ถ้ามี, ไม่งั้น Navigator.pop
  void _closeScreen() {
    if (widget.onClose != null) {
      widget.onClose!();
    } else {
      Navigator.pop(context, ProductBarcodeModel(guidfixed: "", itemcode: ""));
    }
  }

  /// เลือกสินค้า — ถ้าเพิ่มต่อเนื่องจะไม่ปิดจอ, ถ้าไม่เลือกจะปิดทันที
  void _selectProduct(ProductBarcodeModel product) {
    if (widget.onProductAdded != null) {
      widget.onProductAdded!(product);
      if (_continuousAdd) {
        // แสดง feedback ว่าเพิ่มสำเร็จ แล้วให้ user ค้นหาต่อ
        ScaffoldMessenger.of(context).clearSnackBars();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              '${global.language("added")} ${global.activeLangName(product.names!)}',
              style: TextStyle(color: global.theme.onPrimaryColor),
            ),
            backgroundColor: global.theme.positiveHighlightColor,
            duration: const Duration(milliseconds: 1200),
            behavior: SnackBarBehavior.floating,
            margin: const EdgeInsets.only(bottom: 60, left: 16, right: 16),
          ),
        );
        searchFocusNode.requestFocus();
      } else {
        // ปิดจอทันทีหลังเพิ่มสินค้า
        _closeScreen();
      }
    } else {
      Navigator.pop(context, product);
    }
  }

  Color _getRowColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  // ===== BUILD METHODS =====

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
            listener: (context, state) {
              if (state is ProductBarcodeLoadSearchSuccess) {
                setState(() {
                  loadingData = false;
                  if (state.productBarcodes.isNotEmpty) {
                    productBarcodeListData.addAll(state.productBarcodes);
                    if (selectGuid.isEmpty &&
                        productBarcodeListData.isNotEmpty) {
                      selectGuid = productBarcodeListData[0].guidfixed;
                      _updatePreview(productBarcodeListData[0]);
                    }
                  }
                });
              } else if (state is ProductBarcodeLoadSearchFailed) {
                setState(() {
                  loadingData = false;
                  global.showSnackBar(
                    context,
                    Icon(Icons.error_outline,
                        color: global.theme.onPrimaryColor),
                    state.message,
                    global.theme.negativeHighlightTextColor,
                  );
                });
              }
            },
            child: Stack(
              children: [
                (constraints.maxWidth > 800)
                    ? _buildSplitView()
                    : _buildMobileView(),
                if (loadingData)
                  LoadingOverlay(message: global.language('loading')),
              ],
            ),
          );
        },
      ),
    );
  }

  /// Desktop: SplitView — Preview ซ้าย, Search List ขวา
  Widget _buildSplitView() {
    return SplitView(
      gripSize: 8,
      gripColor: global.theme.dividerBorderColor,
      gripColorActive: global.theme.primaryColor,
      viewMode: SplitViewMode.Horizontal,
      indicator: const SplitIndicator(viewMode: SplitViewMode.Horizontal),
      activeIndicator:
          const SplitIndicator(viewMode: SplitViewMode.Horizontal, isActive: true),
      children: _previewOnLeft
          ? [_buildPreviewPanel(), _buildSearchListPanel()]
          : [_buildSearchListPanel(), _buildPreviewPanel()],
    );
  }

  /// Mobile: แสดงเฉพาะ Search List (เหมือน BarcodeSearchScreen เดิม)
  Widget _buildMobileView() {
    return _buildSearchListPanel();
  }

  /// Panel ซ้าย: แสดงรายละเอียดสินค้า
  Widget _buildPreviewPanel() {
    return Container(
      color: global.theme.backgroundColor,
      child: Column(
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: global.theme.appBarColor,
              border: Border(
                bottom:
                    BorderSide(width: 1, color: global.theme.dividerBorderColor),
              ),
            ),
            child: Row(
              children: [
                Icon(Icons.preview, size: 18, color: global.theme.onPrimaryColor),
                const SizedBox(width: 6),
                Text(
                  global.language('product_detail'),
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: global.theme.onPrimaryColor,
                  ),
                ),
              ],
            ),
          ),
          // Content
          Expanded(
            child: _previewProduct != null
                ? SingleChildScrollView(
                    child: ProductPreviewScreen(
                      screenData: _previewProduct!,
                      priceList: (_previewProduct!.prices?.isNotEmpty == true)
                          ? _getActivePriceList()
                          : [],
                      allProductBarcodes: _previewAllBarcodes,
                    ),
                  )
                : Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.touch_app,
                            size: 48,
                            color: global.theme.iconSecondaryColor),
                        const SizedBox(height: 8),
                        Text(
                          global.language('select_product_to_preview'),
                          style: TextStyle(
                            fontSize: 13,
                            color: global.theme.textSecondaryColor,
                          ),
                        ),
                      ],
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  /// Panel ขวา: Search + Product List
  Widget _buildSearchListPanel() {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.close),
          onPressed: _closeScreen,
        ),
        actions: [
          // Checkbox เพิ่มต่อเนื่อง
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                SizedBox(
                  width: 20, height: 20,
                  child: Checkbox(
                    value: _continuousAdd,
                    onChanged: (v) => setState(() => _continuousAdd = v ?? false),
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                ),
                const SizedBox(width: 4),
                GestureDetector(
                  onTap: () => setState(() => _continuousAdd = !_continuousAdd),
                  child: Text(
                    global.language('continuous_add'),
                    style: TextStyle(fontSize: 12, color: global.theme.onPrimaryColor),
                  ),
                ),
              ],
            ),
          ),
          if (MediaQuery.of(context).size.width > 800)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              icon: Icon(
                _previewOnLeft ? Icons.flip : Icons.flip,
                size: 20,
              ),
              tooltip: global.language('swap_side'),
              onPressed: () {
                setState(() {
                  _previewOnLeft = !_previewOnLeft;
                  global.appConfig.setBool('product_preview_on_left', _previewOnLeft);
                });
              },
            ),
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                _closeScreen();
              }
              if (event.logicalKey == LogicalKeyboardKey.enter) {
                // กด Enter เลือกสินค้าที่ highlight อยู่
                if (selectGuid.isNotEmpty) {
                  final selected = productBarcodeListData.firstWhere(
                    (e) => e.guidfixed == selectGuid,
                    orElse: () =>
                        ProductBarcodeModel(guidfixed: "", itemcode: ""),
                  );
                  if (selected.barcode?.isNotEmpty == true) {
                    _selectProduct(selected);
                  }
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                int index = productBarcodeListData
                    .indexWhere((e) => e.guidfixed == selectGuid);
                if (index > 0) {
                  selectGuid = productBarcodeListData[index - 1].guidfixed;
                  _updatePreview(productBarcodeListData[index - 1]);
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                int index = productBarcodeListData
                    .indexWhere((e) => e.guidfixed == selectGuid);
                if (index < productBarcodeListData.length - 1) {
                  selectGuid = productBarcodeListData[index + 1].guidfixed;
                  _updatePreview(productBarcodeListData[index + 1]);
                }
                setState(() {});
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            // Search Box
            _buildSearchBox(),
            // Column Headers
            _buildColumnHeaders(),
            // Product List
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: productBarcodeListData
                      .asMap()
                      .entries
                      .map((e) => _buildListItem(e.value, e.key))
                      .toList(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSearchBox() {
    return Container(
      padding: const EdgeInsets.all(5),
      color: global.theme.appBarColor,
      child: Container(
        padding: const EdgeInsets.all(5),
        decoration: BoxDecoration(
          color: global.theme.surfaceColor,
          borderRadius: BorderRadius.circular(2),
          boxShadow: [
            BoxShadow(
              color: global.theme.dividerBorderColor.withValues(alpha: 0.5),
              spreadRadius: 5,
              blurRadius: 7,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.only(left: 10, right: 10),
          child: TextFormField(
            onFieldSubmitted: (value) {
              searchFocusNode.requestFocus();
            },
            onChanged: (value) {
              _debouncer.run(() {
                try {
                  setState(() {
                    productBarcodeListData = [];
                    _previewProduct = null;
                    selectGuid = "";
                  });
                  _loadDataList(value);
                } catch (_) {}
              });
            },
            autofocus: true,
            focusNode: searchFocusNode,
            controller: searchController,
            style: TextStyle(color: global.theme.textColor),
            decoration: InputDecoration(
              isDense: true,
              filled: false,
              contentPadding: const EdgeInsets.only(top: 10, bottom: 10),
              border: InputBorder.none,
              prefixIcon:
                  Icon(Icons.search, size: 20, color: global.theme.iconColor),
              prefixIconConstraints:
                  const BoxConstraints(minWidth: 32, minHeight: 32),
              hintText: global.language('search'),
              hintStyle: TextStyle(color: global.theme.formHintColor),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildColumnHeaders() {
    return Container(
      padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
      decoration: BoxDecoration(
        color: global.theme.columnHeaderColor,
        border: Border(
          bottom:
              BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
        ),
      ),
      child: Row(
        children: [
          Expanded(
            flex: 5,
            child: Text(
              global.language("barcode"),
              style: TextStyle(
                color: global.theme.columnHeaderTextColor,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          Expanded(
            flex: 5,
            child: Text(
              global.language("itemcode"),
              style: TextStyle(
                color: global.theme.columnHeaderTextColor,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          Expanded(
            flex: 10,
            child: Text(
              global.language("product_name"),
              style: TextStyle(
                color: global.theme.columnHeaderTextColor,
                fontWeight: FontWeight.bold,
              ),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          Expanded(
            flex: 5,
            child: Text(
              global.language("unit"),
              style: TextStyle(
                color: global.theme.columnHeaderTextColor,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildListItem(ProductBarcodeModel value, int index) {
    final isSelected = selectGuid == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.w700,
            color: global.theme.textColor)
        : TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.w400,
            color: global.theme.textSecondaryColor);

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) {
        setState(() => _hoverIndex = index);
        // hover แล้วแสดง preview ทันที
        _updatePreview(value);
      },
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
        onTap: () {
          _selectProduct(value);
        },
        child: Container(
          decoration: BoxDecoration(
            color: _getRowColor(value.guidfixed, index),
            border: Border(
              bottom: BorderSide(
                  width: 1.0, color: global.theme.dividerBorderColor),
            ),
          ),
          padding:
              const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                flex: 5,
                child: Text(
                  value.barcode ?? '',
                  style: textStyle,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Expanded(
                flex: 5,
                child: Text(
                  value.itemcode ?? '',
                  style: textStyle,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Expanded(
                flex: 10,
                child: Text(
                  global.packName(value.names!),
                  style: textStyle,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Expanded(
                flex: 5,
                child: Text(
                  global.packName(value.itemunitnames!),
                  style: textStyle,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// ดึง price list ที่ใช้งานอยู่
  List<PriceModel> _getActivePriceList() {
    List<PriceModel> activePrices = [];
    for (int i = 0; i < global.config.prices.length; i++) {
      if (global.config.prices[i].isUse) {
        activePrices.add(global.config.prices[i]);
      }
    }
    return activePrices;
  }
}
