import 'package:smlaicloud/bloc/product/product_bloc.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_branch_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_draggable_gridview/flutter_draggable_gridview.dart';

class AddProductToBranchScreen extends StatefulWidget {
  const AddProductToBranchScreen({super.key});

  @override
  State<AddProductToBranchScreen> createState() =>
      AddProductToBranchScreenState();
}

class AddProductToBranchScreenState extends State<AddProductToBranchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  List<DraggableGridItem> draggableGridItemList = [];
  bool loadingData = false;
  bool isDataChange = false;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  final _debouncer = global.Debouncer(1000);
  ScrollController listScrollController = ScrollController();
  List<ProductBarcodeModel> listData = [];
  List<ProductBarcodeModel> productList = [];
  String searchText = "";
  String selectProductBranchGuid = "";
  String selectProductBranchCode = "";
  String selectProductBranchName = "";
  late SplitViewController splitViewController;
  List<String>? selectBarcode;
  ProductBranchModel? productBranchModel;

  @override
  void initState() {
    loadDataList("");
    listScrollController.addListener(onScrollList);
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    super.initState();
  }

  @override
  void dispose() {
    searchController.dispose();
    searchFocusNode.dispose();
    super.dispose();
  }

  void discardData({required Function callBack}) {
    if (isDataChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
              onPressed: () {
                Navigator.pop(context);
                callBack();
              },
              child: Text(global.language('yes')),
            ),
          ],
        ),
      );
    } else {
      callBack();
    }
  }

  Widget listObject(int index, ProductBarcodeModel value) {
    return GestureDetector(
      onTap: () {},
      onDoubleTap: () {},
      child: Container(
        color: (index % 2 == 0)
            ? global.theme.columnAlternateEvenColor
            : global.theme.columnAlternateOddColor,
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
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
              flex: 10,
              child: Text(
                global.packName(value.names!),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            IconButton(
              onPressed: () {
                bool found = false;
                for (int i = 0; i < productList.length; i++) {
                  if (productList[i].guidfixed == value.guidfixed) {
                    found = true;
                    break;
                  }
                }
                if (found == false) {
                  isDataChange = true;
                  productList.add(
                    ProductBarcodeModel(
                      guidfixed: value.guidfixed,
                      barcode: value.barcode,
                      names: value.names!,
                      itemcode: value.itemcode,
                    ),
                  );
                  setState(() {});
                }
              },
              icon: Icon(Icons.add),
            ),
          ],
        ),
      ),
    );
  }

  Widget listProduct(int index, ProductBarcodeModel value) {
    return GestureDetector(
      onTap: () {},
      onDoubleTap: () {},
      child: Container(
        color: (index % 2 == 0)
            ? global.theme.columnAlternateEvenColor
            : global.theme.columnAlternateOddColor,
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
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
              flex: 10,
              child: Text(
                global.packName(value.names!),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            IconButton(
              onPressed: () {
                isDataChange = true;
                productList.removeAt(productList.indexOf(value));
                setState(() {});
              },
              icon: Icon(Icons.delete),
            ),
          ],
        ),
      ),
    );
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<ProductBarcodeBloc>().add(
      ProductBarcodeLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
        branchcode: global.companyBranchSelectData.code,
        businesstypecode: global.companyBranchSelectData.businesstype!.code!,
      ),
    );
  }

  Widget listScreen() {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: Column(
        children: [
          Container(
            height: 40,
            padding: EdgeInsets.all(5),
            color: global.theme.searchBarColor,
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    onSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      _debouncer.run(() {
                        setState(() {
                          listData = [];
                        });
                        loadDataList(value);
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.only(
                        top: 0,
                        bottom: 0,
                        left: 0,
                        right: 0,
                      ),
                      border: InputBorder.none,
                      hintText: global.language('search'),
                    ),
                  ),
                ),
              ],
            ),
          ),
          Container(color: global.theme.appBarColor, height: 6),
          Container(
            key: const Key('barcode_list'),
            padding: EdgeInsets.only(
              left: 10,
              right: 10,
              top: 5,
              bottom: 5,
            ),
            color: global.theme.columnHeaderColor,
            child: Row(
              children: [
                Expanded(
                  flex: 5,
                  child: Text(
                    global.language("barcode"),
                    style: TextStyle(
                      color: global.theme.textColor,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                Expanded(
                  flex: 10,
                  child: Text(
                    global.language("product_name"),
                    style: TextStyle(
                      color: global.theme.textColor,
                      fontWeight: FontWeight.bold,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView(
              controller: listScrollController,
              children: listData
                  .map((value) => listObject(listData.indexOf(value), value))
                  .toList(),
            ),
          ),
          if (loadingData)
            Center(
              child: LoadingAnimationWidget.staggeredDotsWave(
                color: global.theme.infoHighlightTextColor,
                size: 50,
              ),
            ),
        ],
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

  Widget gridView() {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Container(
          padding: const EdgeInsets.all(5),
          child: Column(
            children: [
              Row(
                children: [
                  Expanded(
                    child: Container(
                      padding: const EdgeInsets.only(bottom: 5),
                      width: double.infinity,
                      height: 50,
                      child: ElevatedButton(
                        onPressed: () {
                          discardData(
                            callBack: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) =>
                                      const CompanyBranchSearchScreen(word: ""),
                                ),
                              ).then((value) {
                                setState(() {
                                  SearchGuidCodeNameModel result = value;
                                  if (result.isCancel == false) {
                                    selectProductBranchGuid = result.guid;
                                    selectProductBranchCode = result.code;
                                    selectProductBranchName = global.packName(
                                      result.names,
                                    );
                                    context.read<ProductBloc>().add(
                                      ProductBranchGet(
                                        branchcode: selectProductBranchCode,
                                      ),
                                    );
                                  }
                                });
                              });
                            },
                          );
                        },
                        child: Text(
                          (selectProductBranchGuid.isEmpty)
                              ? global.language("select_company_branch")
                              : selectProductBranchName,
                        ),
                      ),
                    ),
                  ),
                  Container(
                    padding: EdgeInsets.only(left: 5, bottom: 5),
                    width: 100,
                    height: 50,
                    child: ElevatedButton(
                      child: Icon(Icons.save, color: global.theme.onPrimaryColor),
                      onPressed: () {
                        setState(() {
                          selectBarcode = [];

                          for (int i = 0; i < productList.length; i++) {
                            selectBarcode!.add(productList[i].barcode!);
                          }
                          productBranchModel = ProductBranchModel(
                            branchcode: selectProductBranchCode,
                            productcodes: selectBarcode,
                          );

                          context.read<ProductBloc>().add(
                            ProductBranchSave(
                              productBranchModel: productBranchModel!,
                            ),
                          );
                        });
                      },
                    ),
                  ),
                ],
              ),
              Expanded(
                child: (productList.isEmpty)
                    ? Column(
                        children: [
                          Container(
                            key: const Key('product_list_header'),
                            padding: EdgeInsets.only(
                              left: 10,
                              right: 10,
                              top: 5,
                              bottom: 5,
                            ),
                            color: global.theme.columnHeaderColor,
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 5,
                                  child: Text(
                                    global.language("barcode"),
                                    style: TextStyle(
                                      color: global.theme.textColor,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 10,
                                  child: Text(
                                    global.language("product_name"),
                                    style: TextStyle(
                                      color: global.theme.textColor,
                                      fontWeight: FontWeight.bold,
                                    ),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          Center(
                            child: Text(global.language("no_select_product")),
                          ),
                        ],
                      )
                    : Column(
                        children: [
                          Container(
                            key: const Key('product_list_header'),
                            padding: EdgeInsets.only(
                              left: 10,
                              right: 10,
                              top: 5,
                              bottom: 5,
                            ),
                            color: global.theme.columnHeaderColor,
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 5,
                                  child: Text(
                                    global.language("barcode"),
                                    style: TextStyle(
                                      color: global.theme.textColor,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 10,
                                  child: Text(
                                    global.language("product_name"),
                                    style: TextStyle(
                                      color: global.theme.textColor,
                                      fontWeight: FontWeight.bold,
                                    ),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          Expanded(
                            child: ListView(
                              children: productList
                                  .map(
                                    (value) => listProduct(
                                      productList.indexOf(value),
                                      value,
                                    ),
                                  )
                                  .toList(),
                            ),
                          ),
                        ],
                      ),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return MultiBlocListener(
          listeners: [
            BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
              listener: (context, state) {
                // Load
                if (state is ProductBarcodeLoadSuccess) {
                  setState(() {
                    loadingData = false;
                    if (state.productBarcodes.isNotEmpty) {
                      listData.addAll(state.productBarcodes);
                    }
                  });
                }
                if (state is ProductBarcodeLoadFailed) {
                  setState(() {
                    loadingData = false;
                  });
                }
              },
            ),
            BlocListener<ProductBloc, ProductState>(
              listener: (context, state) {
                // Load
                if (state is ProductBranchGetSuccess) {
                  setState(() {
                    if (state.products.isNotEmpty) {
                      productList.addAll(state.products);
                    }
                  });
                }
                if (state is ProductBranchUpdateSuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: global.theme.onPrimaryColor),
                      global.language('edit_success'),
                      global.theme.infoHighlightTextColor,
                    );

                    setState(() {
                      isDataChange = false;
                      productList = [];
                      searchText = "";
                      selectProductBranchGuid = "";
                      selectProductBranchCode = "";
                      selectProductBranchName = "";
                    });
                  });
                }
                if (state is ProductBranchUpdateFailed) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: global.theme.onPrimaryColor),
                      "${global.language('edit_failed')} : ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
                  });
                }
              },
            ),
          ],
          child: Scaffold(
            resizeToAvoidBottomInset: true,
            appBar: AppBar(
              backgroundColor: global.theme.appBarColor,
              automaticallyImplyLeading: false,
              title: Text(global.language('add_product_to_branch')),
              leading: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.arrow_back),
                onPressed: () {
                  discardData(
                    callBack: () {
                      Navigator.pop(context);
                    },
                  );
                },
              ),
            ),
            body: (constraints.maxWidth < 800.0)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 8,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: global.theme.infoHighlightTextColor,
                    viewMode: SplitViewMode.Vertical,
                    indicator: const SplitIndicator(
                      viewMode: SplitViewMode.Vertical,
                    ),
                    activeIndicator: const SplitIndicator(
                      viewMode: SplitViewMode.Vertical,
                      isActive: true,
                    ),
                    children: [listScreen(), gridView()],
                  )
                : SplitView(
                    controller: splitViewController,
                    gripSize: 8,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: global.theme.infoHighlightTextColor,
                    viewMode: SplitViewMode.Horizontal,
                    indicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                    ),
                    activeIndicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                      isActive: true,
                    ),
                    children: [listScreen(), gridView()],
                  ),
          ),
        );
      },
    );
  }
}
