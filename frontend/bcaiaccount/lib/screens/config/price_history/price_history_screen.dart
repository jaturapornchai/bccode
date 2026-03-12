import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/utils/image_tooltip.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/screens/config/price_history/components/product_list_item.dart';
import 'package:smlaicloud/screens/config/price_history/components/product_search_bar.dart';
import 'package:smlaicloud/screens/config/price_history/components/product_list_header.dart';
import 'package:smlaicloud/screens/config/price_history/components/product_filter_dialog.dart';
import 'package:smlaicloud/screens/config/price_history/components/product_details_view.dart';

class PriceHistoryScreen extends StatefulWidget {
  const PriceHistoryScreen({super.key});

  @override
  State<PriceHistoryScreen> createState() => PriceHistoryScreenState();
}

class PriceHistoryScreenState extends State<PriceHistoryScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  List<GlobalKey<ImageTooltipState>> tooltipKeys = [];
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  FocusNode listFocusNode = FocusNode();
  List<ProductBarcodeModel> listData = [];
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  String selectGuid = "";
  late ProductBarcodeState blocCurrentState;
  final _debouncer = global.Debouncer(500);
  bool loadingData = false;
  bool showImage = false;
  List<CompanyBranchModel> listDataBranchAll = [];
  FiltterBarcodeModel filterBarcode = FiltterBarcodeModel(branch: false);
  late TabController tabController;
  ProductBarcodeModel? selectedProductData;

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);
    listData = [];
    await Future.wait<void>([
      loadDataList("", filterBarcode),
      loadBranchAll(),
    ]);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    tabController = TabController(vsync: this, length: 2);
    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    tabController.dispose();
    listFocusNode.dispose(); // Dispose the new focus node
    removeTooltip();
    for (int i = 0; i < tooltipKeys.length; i++) {
      tooltipKeys[i].currentState?.dispose();
    }
    super.dispose();
  }

  void removeTooltip() {
    for (int i = 0; i < tooltipKeys.length; i++) {
      tooltipKeys[i].currentState?.removeTooltip();
    }
  }

  Future<void> loadDataList(
      String search, FiltterBarcodeModel filterBarcode) async {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<ProductBarcodeBloc>().add(ProductBarcodeLoadList(
          offset: (listData.isEmpty) ? 0 : listData.length,
          limit: global.loadDataPerPage,
          search: search,
          branchcode: (filterBarcode.branch == true)
              ? global.companyBranchSelectData.code
              : "",
          businesstypecode: (filterBarcode.branch == true)
              ? global.companyBranchSelectData.businesstype!.code!
              : "",
        ));
  }

  Future<void> loadBranchAll() async {
    listDataBranchAll = [];
    context
        .read<CompanyBranchBloc>()
        .add(const CompanyBranchLoadList(offset: 0, limit: 100, search: ""));
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      /// set delay 500 ms
      _debouncer.run(() {
        loadDataList(searchText, filterBarcode);
      });
    }
  }

  void getData(String guid) {
    context.read<ProductBarcodeBloc>().add(ProductBarcodeGet(guid: guid));
  }

  void _scrollToSelectedItem(int index) {
    if (listScrollController.hasClients &&
        index >= 0 &&
        index < listData.length) {
      // Calculate approximate item height (adjust based on your item height)
      double itemHeight =
          60.0; // Adjust this value based on your actual item height
      double targetOffset = index * itemHeight;

      // Get viewport height
      double viewportHeight = listScrollController.position.viewportDimension;
      double maxScrollExtent = listScrollController.position.maxScrollExtent;

      // Center the item in viewport
      double centeredOffset =
          targetOffset - (viewportHeight / 2) + (itemHeight / 2);

      // Clamp offset to valid range
      centeredOffset = centeredOffset.clamp(0.0, maxScrollExtent);

      listScrollController.animateTo(
        centeredOffset,
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
      );
    }
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('barcode')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            global.gotoMainMenu(context);
          },
        ),
        actions: <Widget>[const ManualButton(path: 'price-history')],
      ),
      body: Focus(
        focusNode: listFocusNode,
        autofocus: true,
        onKey: (node, event) {
          if (kIsWeb && event is RawKeyDownEvent && listData.isNotEmpty) {
            if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
              // If no item selected, select the first one
              if (selectGuid.isEmpty) {
                selectGuid = listData[0].guidfixed;
                getData(selectGuid); // Load data for the selected item
                setState(() {});
                return KeyEventResult.handled;
              }

              // Find current index
              int currentIndex = listData
                  .indexWhere((element) => element.guidfixed == selectGuid);
              if (currentIndex > 0) {
                selectGuid = listData[currentIndex - 1].guidfixed;
                getData(selectGuid); // Load data for the selected item
                setState(() {});
                _scrollToSelectedItem(currentIndex - 1);
                return KeyEventResult.handled;
              }
            }

            if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
              // If no item selected, select the first one
              if (selectGuid.isEmpty) {
                selectGuid = listData[0].guidfixed;
                getData(selectGuid); // Load data for the selected item
                setState(() {});
                return KeyEventResult.handled;
              }

              // Find current index
              int currentIndex = listData
                  .indexWhere((element) => element.guidfixed == selectGuid);
              if (currentIndex >= 0 && currentIndex < listData.length - 1) {
                selectGuid = listData[currentIndex + 1].guidfixed;
                getData(selectGuid); // Load data for the selected item
                setState(() {});
                _scrollToSelectedItem(currentIndex + 1);
                return KeyEventResult.handled;
              }
            }

            if (event.logicalKey == LogicalKeyboardKey.enter) {
              // Trigger selection of current item
              if (selectGuid.isNotEmpty) {
                getData(selectGuid);
                WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                  tabController.animateTo(1);
                });
                return KeyEventResult.handled;
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(children: [
          ProductSearchBar(
            searchController: searchController,
            searchFocusNode: searchFocusNode,
            listFocusNode: listFocusNode,
            filterBarcode: filterBarcode,
            showImage: showImage,
            onSearchChanged: (value) {
              _debouncer.run(() {
                setState(() {
                  listData = [];
                });
                loadDataList(value, filterBarcode);
              });
            },
            onFilterPressed: () async {
              final result = await showDialog<FiltterBarcodeModel>(
                context: context,
                builder: (context) =>
                    ProductFilterDialog(initialFilter: filterBarcode),
              );
              if (result != null) {
                setState(() {
                  filterBarcode = result;
                  listData = [];
                  loadDataList(searchText, filterBarcode);
                });
                WidgetsBinding.instance.addPostFrameCallback((_) {
                  listFocusNode.requestFocus();
                });
              }
            },
            onImageToggle: () {
              setState(() {
                showImage = !showImage;
              });
            },
            onFontSizeChange: () {
              setState(() {});
            },
          ),
          Container(
            color: global.theme.appBarColor,
            height: 6,
          ),
          ProductListHeader(showImage: showImage),
          Expanded(
              child: ListView(
                  controller: listScrollController,
                  children: listData
                      .map(
                          (value) => listObject(listData.indexOf(value), value))
                      .toList())),
          if (loadingData)
            Center(
                child: LoadingAnimationWidget.staggeredDotsWave(
              color: global.theme.infoHighlightTextColor,
              size: 50,
            ))
        ]),
      ),
    );
  }

  Widget listObject(int index, ProductBarcodeModel value) {
    return ProductListItem(
      index: index,
      product: value,
      selectedGuid: selectGuid,
      showImage: showImage,
      showBranches: filterBarcode.branch == false,
      imageTooltipKey: tooltipKeys[index],
      onTap: () {
        setState(() {
          selectGuid = value.guidfixed;
          getData(selectGuid);
          WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
            tabController.animateTo(1);
          });
        });
      },
    );
  }

  Widget editScreen({bool mobileScreen = false}) {
    return ProductDetailsView(
      productData: selectedProductData,
      tabController: mobileScreen ? tabController : null,
      isMobile: mobileScreen,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        resizeToAvoidBottomInset: true,
        body: LayoutBuilder(builder: (context, constraints) {
          return MultiBlocListener(
              listeners: [
                BlocListener<CompanyBranchBloc, CompanyBranchState>(
                  listener: (context, state) async {
                    if (state is CompanyBranchLoadSuccess) {
                      setState(() {
                        listDataBranchAll = state.companyBranch;
                      });
                    }
                  },
                ),
                BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
                  listener: (context, state) async {
                    blocCurrentState = state;
                    // Load
                    if (state is ProductBarcodeLoadSuccess) {
                      removeTooltip();
                      setState(() {
                        loadingData = false;
                      });

                      if (state.productBarcodes.isNotEmpty) {
                        // Create a list of futures
                        List<Future<void>> futures = [];

                        for (int i = 0; i < state.productBarcodes.length; i++) {
                          futures.add(Future(() async {
                            List<CompanyBranchModel>
                                listDataBranchInBusinessType = [];

                            /// find branch in listDataBranchAll where state.productBarcodes[i].businesstypes = listDataBranchAll[j].businesstype
                            if (state.productBarcodes[i].businesstypes !=
                                null) {
                              for (int j = 0;
                                  j <
                                      state.productBarcodes[i].businesstypes!
                                          .length;
                                  j++) {
                                var foundBranches = listDataBranchAll.where(
                                    (element) =>
                                        element.businesstype?.guidfixed ==
                                        state.productBarcodes[i]
                                            .businesstypes![j].guidfixed);
                                listDataBranchInBusinessType
                                    .addAll(foundBranches);
                              }
                            }

                            // Filter out ignored branches
                            if (state.productBarcodes[i].ignorebranches !=
                                null) {
                              state.productBarcodes[i].branches =
                                  listDataBranchInBusinessType
                                      .where(
                                        (element) => !state
                                            .productBarcodes[i].ignorebranches!
                                            .any(
                                          (ignoreBranch) =>
                                              element.guidfixed ==
                                              ignoreBranch.guidfixed,
                                        ),
                                      )
                                      .toList();
                            } else {
                              state.productBarcodes[i].branches =
                                  listDataBranchInBusinessType;
                            }
                          }));
                        }

                        // Wait for all futures to complete
                        await Future.wait(futures);

                        setState(() {
                          listData.addAll(state.productBarcodes);
                        });
                      }

                      tooltipKeys.clear();
                      for (int i = 0; i < listData.length; i++) {
                        tooltipKeys.add(GlobalKey<ImageTooltipState>());
                      }
                    }

                    if (state is ProductBarcodeLoadFailed) {
                      setState(() {
                        loadingData = false;
                        global.showSnackBar(
                            context,
                            Icon(
                              Icons.error_outline,
                              color: global.theme.onPrimaryColor,
                            ),
                            state.message,
                            global.theme.negativeHighlightTextColor);
                      });
                    }

                    // Get
                    if (state is ProductBarcodeGetSuccess) {
                      setState(() {
                        selectedProductData = state.productBarcode;
                      });
                    }
                  },
                ),
              ],
              child: (constraints.maxWidth > 800)
                  ? SplitView(
                      gripSize: 8,
                      gripColor: global.theme.appBarColor,
                      gripColorActive: global.theme.infoHighlightTextColor,
                      viewMode: SplitViewMode.Horizontal,
                      indicator: const SplitIndicator(
                          viewMode: SplitViewMode.Horizontal),
                      activeIndicator: const SplitIndicator(
                        viewMode: SplitViewMode.Horizontal,
                        isActive: true,
                      ),
                      children: [
                        listScreen(mobileScreen: false),
                        editScreen(mobileScreen: false),
                      ],
                    )
                  : TabBarView(
                      physics: const NeverScrollableScrollPhysics(),
                      controller: tabController,
                      children: [
                        listScreen(mobileScreen: true),
                        editScreen(mobileScreen: true)
                      ],
                    ));
        }));
  }
}
