import 'dart:io';

import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/components/loading_overlay.dart';

class BarcodeSearchScreen extends StatefulWidget {
  const BarcodeSearchScreen({
    super.key,
    required this.word,
    required this.screen,
  });
  final String word;
  final String screen;

  @override
  State<BarcodeSearchScreen> createState() => BarcodeSearchScreenState();
}

class BarcodeSearchScreenState extends State<BarcodeSearchScreen>
    with SingleTickerProviderStateMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  String searchItemType = "";
  List<ProductBarcodeModel> productBarcodeListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);
  String isUseSubBarcodes = "";
  String isBom = "";
  bool loadingData = false;
  bool isLoading = false; // สำหรับ LoadingOverlay

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = widget.word;
    searchController.text = searchText;

    if (widget.screen == "not_material") {
      // ไม่ส่ง itemtype เพื่อให้ดึงข้อมูลทั้งหมด (เหมือนหน้าจอบาร์โค้ด)
      searchItemType = "";
    } else if (widget.screen == "material") {
      // ไม่ส่ง itemtype เพื่อให้ดึงข้อมูลทั้งหมด (เหมือนหน้าจอบาร์โค้ด)
      searchItemType = "";
    } else if (widget.screen == "stockbalance") {
      /// รายงานยอดคงเหลือ
      searchItemType = "0,3,4";
      isBom = "all";
      isUseSubBarcodes = "notshowsubbarcodes";
    } else if (widget.screen == "subbarcodeSearch") {
      /// ค้นหารหัสสินค้าจากบาร์โค้ดย่อย
      searchItemType = "0,1,2,3,4,5";
      isBom = "all";
      isUseSubBarcodes = "notshowsubbarcodes";
    } else {
      // ไม่ส่ง itemtype เพื่อให้ดึงข้อมูลทั้งหมด (เหมือนหน้าจอบาร์โค้ด)
      searchItemType = "";
      isBom = "all";
      isUseSubBarcodes = "all";
    }
    loadDataList(searchText);

    super.initState();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;

    // เรียกใช้ฟังก์ชันเพื่อดึงค่า shopsid
    String shopsid = global.getShopId();

    // 🔍 Log เพื่อตรวจสอบ parameters ที่ส่งไป API
    final offset = (productBarcodeListData.isEmpty) ? 0 : productBarcodeListData.length;
    debugPrint('🔍 [BarcodeSearch] ===== Request Parameters =====');
    debugPrint('🔍 [BarcodeSearch] offset: $offset');
    debugPrint('🔍 [BarcodeSearch] limit: ${global.loadDataPerPage}');
    debugPrint('🔍 [BarcodeSearch] search: "$search"');
    debugPrint('🔍 [BarcodeSearch] itemtype: $searchItemType');
    debugPrint('🔍 [BarcodeSearch] branchcode: ${global.companyBranchSelectData.code}');
    debugPrint('🔍 [BarcodeSearch] businesstypecode: ${global.companyBranchSelectData.businesstype!.code!}');
    debugPrint('🔍 [BarcodeSearch] isbom: $isBom');
    debugPrint('🔍 [BarcodeSearch] isusesubbarcodes: $isUseSubBarcodes');
    debugPrint('🔍 [BarcodeSearch] shopsid: $shopsid');
    debugPrint('🔍 [BarcodeSearch] currentListCount: ${productBarcodeListData.length}');
    debugPrint('🔍 [BarcodeSearch] ================================');

    context.read<ProductBarcodeBloc>().add(
      ProductBarcodeLoadListSearch(
        offset: offset,
        limit: global.loadDataPerPage,
        search: search,
        itemtype: searchItemType,
        branchcode: "", // ไม่ส่ง branchcode เพื่อให้ดึงข้อมูลทั้งหมด (เหมือนหน้าจอบาร์โค้ด)
        businesstypecode: "", // ไม่ส่ง businesstypecode เพื่อให้ดึงข้อมูลทั้งหมด (เหมือนหน้าจอบาร์โค้ด)
        isbom: isBom,
        isusesubbarcodes: isUseSubBarcodes,
        shopsid: shopsid,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              ProductBarcodeModel(guidfixed: "", itemcode: ""),
            );
          },
        ),
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
                Navigator.pop(
                  context,
                  ProductBarcodeModel(guidfixed: "", itemcode: ""),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    ProductBarcodeModel(guidfixed: "", itemcode: ""),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = productBarcodeListData.indexOf(
                  productBarcodeListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = productBarcodeListData[index - 1].guidfixed;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = productBarcodeListData.indexOf(
                  productBarcodeListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < productBarcodeListData.length - 1) {
                  selectGuid = productBarcodeListData[index + 1].guidfixed;
                }
                isKeyDown = true;
                setState(() {});
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(5),
              color: global.theme.appBarColor,
              child: Container(
                padding: const EdgeInsets.all(5),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(2),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.grey.withOpacity(0.5),
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
                          });
                          loadDataList(value);
                        } catch (_) {}
                      });
                    },
                    autofocus: true,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.only(
                        top: 10,
                        bottom: 10,
                      ),
                      border: InputBorder.none,
                      hintText: global.language('search'),
                    ),
                  ),
                ),
              ),
            ),
            Container(
              padding: const EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              decoration: BoxDecoration(
                color: global.theme.columnHeaderColor,
                border: const Border(
                  bottom: BorderSide(width: 1.0, color: Colors.grey),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("barcode"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("itemcode"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("product_name"),
                      style: const TextStyle(
                        color: Colors.black,
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
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: productBarcodeListData
                      .map((value) => listObject(value))
                      .toList(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget listObject(ProductBarcodeModel value) {
    return GestureDetector(
      onTap: () {
        Navigator.pop(context, value);
      },
      child: Container(
        decoration: BoxDecoration(
          color: (selectGuid == value.guidfixed)
              ? Colors.cyan[100]
              : Colors.white,
          border: const Border(
            bottom: BorderSide(width: 1.0, color: Colors.grey),
          ),
        ),
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 5,
              child: Text(
                value.barcode!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.itemcode!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.names!),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.packName(value.itemunitnames!),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < productBarcodeListData.length; i++) {
      if (productBarcodeListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
            listener: (context, state) {
              // Load
              if (state is ProductBarcodeLoadSearchSuccess) {
                // 🔍 Log เพื่อตรวจสอบ response จาก API
                debugPrint('✅ [BarcodeSearch] ===== Response Success =====');
                debugPrint('✅ [BarcodeSearch] receivedCount: ${state.productBarcodes.length}');
                debugPrint('✅ [BarcodeSearch] totalInList (before add): ${productBarcodeListData.length}');
                if (state.productBarcodes.isNotEmpty) {
                  debugPrint('✅ [BarcodeSearch] firstItem: ${state.productBarcodes.first.barcode} - ${state.productBarcodes.first.itemcode}');
                }
                debugPrint('✅ [BarcodeSearch] ================================');

                setState(() {
                  loadingData = false;
                  if (state.productBarcodes.isNotEmpty) {
                    productBarcodeListData.addAll(state.productBarcodes);
                    debugPrint('✅ [BarcodeSearch] totalInList (after add): ${productBarcodeListData.length}');
                    if (productBarcodeListData.isNotEmpty) {
                      selectGuid = productBarcodeListData[0].guidfixed;
                    } else {
                      selectGuid = "";
                    }
                  }
                });
              } else if (state is ProductBarcodeLoadSearchFailed) {
                // 🔍 Log เมื่อเกิด error
                debugPrint('❌ [BarcodeSearch] ===== Response Failed =====');
                debugPrint('❌ [BarcodeSearch] message: ${state.message}');
                debugPrint('❌ [BarcodeSearch] ================================');

                setState(() {
                  loadingData = false;
                  global.showSnackBar(
                    context,
                    const Icon(Icons.error_outline, color: Colors.white),
                    state.message,
                    Colors.red,
                  );
                });
              }
            },
            child: Stack(
              children: [
                // Main content
                (constraints.maxWidth > 800)
                    ? listScreen(mobileScreen: false)
                    : listScreen(mobileScreen: true),

                // Loading Overlay
                if (loadingData)
                  LoadingOverlay(message: global.language('loading')),
              ],
            ),
          );
        },
      ),
    );
  }
}
