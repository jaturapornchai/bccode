import 'dart:async';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/product_bom_model.dart';
import 'package:smlaicloud/screens/config/product_bom_widget.dart';
import 'package:smlaicloud/utils/image_tooltip.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:loader_overlay/loader_overlay.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';

class ProductBarcodeBomScreen extends StatefulWidget {
  const ProductBarcodeBomScreen({super.key});

  @override
  State<ProductBarcodeBomScreen> createState() =>
      ProductBarcodeBomScreenState();
}

class ProductBarcodeBomScreenState extends State<ProductBarcodeBomScreen>
    with SingleTickerProviderStateMixin {
  List<GlobalKey<ImageTooltipState>> tooltipKeys = [];
  late TabController tabController;
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<ProductBarcodeModel> listData = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  late ProductBarcodeState blocCurrentState;
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  final _debouncer = global.Debouncer(500);
  bool loadingData = false;
  bool showImage = false;

  FiltterBarcodeModel filterBarcode = FiltterBarcodeModel(branch: false);
  List<CompanyBranchModel> listDataBranchAll = [];

  ProductBomModel productBom = ProductBomModel(
    guidfixed: '',
    names: [],
    itemunitcode: '',
    itemunitnames: [],
    barcode: '',
    condition: false,
    dividevalue: 0,
    standvalue: 0,
    qty: 0,
    imageuri: '',
    bom: [],
  );

  @override
  void initState() {
    loadDataList("", filterBarcode);
    loadBranchAll();
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    listScrollController.addListener(onScrollList);

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    searchController.dispose();

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

  void loadDataList(String search, FiltterBarcodeModel filterBarcode) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<ProductBarcodeBloc>().add(
      ProductBarcodeLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
        branchcode: (filterBarcode.branch == true)
            ? global.companyBranchSelectData.code
            : "",
        businesstypecode: (filterBarcode.branch == true)
            ? global.companyBranchSelectData.businesstype!.code!
            : "",
        isbom: "showbom",
        isusesubbarcodes: "notshowsubbarcodes",
      ),
    );
  }

  void loadBranchAll() {
    listDataBranchAll = [];
    context.read<CompanyBranchBloc>().add(
      const CompanyBranchLoadList(offset: 0, limit: 100, search: ""),
    );
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

  void getData(String barcode) {
    context.read<ProductBarcodeBloc>().add(
      ProductBarcodeGetBom(barcode: barcode),
    );
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
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            global.gotoMainMenu(context);
          },
        ),
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listData[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;

                  getData(listData[index].barcode!);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = listData[index + 1].guidfixed;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(listData[index].barcode!);
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(5),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(2),
              ),
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
                          loadDataList(value, filterBarcode);
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
                  ((appConfig.getInt("branch_total") != 1))
                      ? IconButton(
                          onPressed: () async {
                            filterBarcode = await filterBox(filterBarcode);

                            setState(() {
                              listData = [];
                              loadDataList(searchText, filterBarcode);
                            });
                          },
                          icon: Icon(
                            (filterBarcode.branch == false)
                                ? Icons.filter_alt_off
                                : Icons.filter_alt,
                            color: (filterBarcode.branch == false)
                                ? Colors.black
                                : Colors.blue,
                          ),
                        )
                      : Container(),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: Icon(
                      (showImage) ? Icons.image_not_supported : Icons.image,
                    ),
                    onPressed: () async {
                      setState(() {
                        showImage = !showImage;
                      });
                    },
                  ),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: const FaIcon(FontAwesomeIcons.font),
                    onPressed: () async {
                      setState(() {
                        global.listDataFontSizeChange();
                      });
                    },
                  ),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: const Icon(Icons.line_weight),
                    onPressed: () async {
                      setState(() {
                        global.listDataLineSpaceChange();
                      });
                    },
                  ),
                ],
              ),
            ),
            Container(color: global.theme.appBarColor, height: 6),
            Container(
              key: headerKey,
              padding: const EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              color: global.theme.columnHeaderColor,
              child: Row(
                children: [
                  Expanded(
                    flex: 6,
                    child: Text(
                      global.language("barcode"),
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
                    flex: 2,
                    child: Text(
                      global.language("unit"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 4,
                    child: Text(
                      global.language("item_code"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 4,
                    child: Text(
                      global.language("product_group"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  if (showImage)
                    const Expanded(
                      flex: 1,
                      child: Icon(Icons.image, color: Colors.black, size: 12),
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
                  color: Colors.blue,
                  size: 50,
                ),
              ),
          ],
        ),
      ),
    );
  }

  Future<FiltterBarcodeModel> filterBox(
    FiltterBarcodeModel filterBarcode,
  ) async {
    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              title: Text(global.language("filter_product")),
              content: SizedBox(
                width: 500.0,
                height: 500.0,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Divider(),
                    RadioListTile(
                      title: Text(global.language("branch_selected")),
                      value: true,
                      groupValue: filterBarcode.branch,
                      onChanged: (value) {
                        setState(() {
                          filterBarcode.branch = value!;
                        });
                      },
                    ),
                    RadioListTile(
                      title: Text(global.language("all")),
                      value: false,
                      groupValue: filterBarcode.branch,
                      onChanged: (value) {
                        setState(() {
                          filterBarcode.branch = value!;
                        });
                      },
                    ),
                  ],
                ),
              ),
              actions: [
                ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("confirm")),
                ),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("clear")),
                ),
              ],
            );
          },
        );
      },
    );

    return filterBarcode;
  }

  Widget listObject(int index, ProductBarcodeModel value) {
    bool isCheck = false;
    TextStyle textStyle = (selectGuid == value.guidfixed)
        ? TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.bold,
          )
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize);

    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    return GestureDetector(
      onTap: () {
        setState(() {
          selectGuid = value.guidfixed;

          getData(value.barcode!);
          //searchFocusNode.requestFocus();
          WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
            tabController.animateTo(1);
          });
        });
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        color: _getContainerColor(selectGuid, value.guidfixed, index),
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
        child: Column(
          children: [
            _buildRow(value, textStyle, showImage, tooltipKeys[index], isCheck),
            (filterBarcode.branch == false &&
                    (appConfig.getInt("branch_total") != 1))
                ? SizedBox(
                    width: double.infinity,
                    child: _buildChipWrap(value.branches!),
                  )
                : Container(),
          ],
        ),
      ),
    );
  }

  Color? _getContainerColor(String selectGuid, String guidfixed, int index) {
    return (selectGuid == guidfixed)
        ? Colors.cyan[100]
        : (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  Row _buildRow(
    ProductBarcodeModel value,
    TextStyle textStyle,
    bool showImage,
    GlobalKey<ImageTooltipState> key,
    bool isCheck,
  ) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          flex: 6,
          child: Text(
            value.barcode!,
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
          flex: 2,
          child: Text(
            global.packName(value.itemunitnames!),
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        Expanded(
          flex: 4,
          child: Text(
            value.itemcode!,
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        Expanded(
          flex: 4,
          child: Text(
            global.packName(value.groupnames!),
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        if (showImage) _buildImageWidget(value, key),
      ],
    );
  }

  Widget _buildImageWidget(
    ProductBarcodeModel value,
    GlobalKey<ImageTooltipState> key,
  ) {
    return Expanded(
      flex: 1,
      child:
          (value.imageuri!.isNotEmpty ||
              1 == 1) // Consider revising this condition
          ? ImageTooltip(
              key: key,
              image: Image.network(value.imageuri!),
              child: _buildImageContainer(value.imageuri!),
            )
          : Container(),
    );
  }

  Container _buildImageContainer(String imageUri) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(2),
        border: Border.all(color: Colors.grey, width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.5),
            spreadRadius: 1,
            blurRadius: 1,
            offset: const Offset(1, 1),
          ),
        ],
      ),
      padding: const EdgeInsets.all(1),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(2),
        child: Image.network(
          imageUri,
          fit: BoxFit.fitHeight,
          height: 20,
          errorBuilder:
              (BuildContext context, Object exception, StackTrace? stackTrace) {
                return const Icon(Icons.image_not_supported);
              },
        ),
      ),
    );
  }

  Wrap _buildChipWrap(List<CompanyBranchModel> branches) {
    return Wrap(
      spacing: 2.0,
      runSpacing: 2.0,
      children: branches
          .asMap()
          .entries
          .map((entry) => _buildChip(entry.value))
          .toList(),
    );
  }

  Chip _buildChip(CompanyBranchModel branch) {
    return Chip(
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      padding: const EdgeInsets.all(0),
      label: Text(
        '${branch.code} - ${global.activeLangName(branch.names)}',
        style: const TextStyle(fontSize: 12),
      ),
    );
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        leading: mobileScreen
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () async {
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                },
              )
            : null,
        title: Text(global.language("product_bom")),
      ),
      body: LoaderOverlay(
        overlayColor: Colors.white,
        child: SizedBox(
          child: (productBom.bom!.isNotEmpty)
              ? ProductBomWidget(productBom: productBom)
              : Container(),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();

    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is ProductBarcodeLoadSuccess) {
                    removeTooltip();
                    setState(() {
                      loadingData = false;
                    });
                    if (state.productBarcodes.isNotEmpty) {
                      for (int i = 0; i < state.productBarcodes.length; i++) {
                        List<CompanyBranchModel> listDataBranchInBusinessType =
                            [];

                        /// find branch in listDataBranchAll where state.productBarcodes[i].businesstypes = listDataBranchAll[j].businesstype
                        for (
                          int j = 0;
                          j < state.productBarcodes[i].businesstypes!.length;
                          j++
                        ) {
                          listDataBranchInBusinessType.addAll(
                            listDataBranchAll.where(
                              (element) =>
                                  element.businesstype!.guidfixed ==
                                  state
                                      .productBarcodes[i]
                                      .businesstypes![j]
                                      .guidfixed,
                            ),
                          );
                        }

                        state
                            .productBarcodes[i]
                            .branches = listDataBranchInBusinessType
                            .where(
                              (element) =>
                                  !state.productBarcodes[i].ignorebranches!.any(
                                    (ignoreBranch) =>
                                        element.guidfixed ==
                                        ignoreBranch.guidfixed,
                                  ),
                            )
                            .toList();
                      }
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
                        const Icon(Icons.error_outline, color: Colors.white),
                        state.message,
                        Colors.red,
                      );
                    });
                  }

                  // Get Bom
                  if (state is ProductBarcodeGetBomInProgress) {
                    context.loaderOverlay.show(
                      widgetBuilder: (progress) {
                        return const ReconnectingOverlay();
                      },
                    );
                  } else if (state is ProductBarcodeGetBomSuccess) {
                    context.loaderOverlay.hide();
                    setState(() {
                      productBom = state.productBom;
                      productBom.qty = 1;
                    });
                  } else if (state is ProductBarcodeGetBomFailed) {
                    context.loaderOverlay.hide();
                    global.showSnackBar(
                      context,
                      const Icon(Icons.error_outline, color: Colors.white),
                      state.message,
                      Colors.red,
                    );
                  }
                },
              ),
              BlocListener<CompanyBranchBloc, CompanyBranchState>(
                listener: (context, state) {
                  if (state is CompanyBranchLoadSuccess) {
                    setState(() {
                      listDataBranchAll = state.companyBranch;
                    });
                  }
                },
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    gripSize: 8,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: Colors.blue,
                    viewMode: SplitViewMode.Horizontal,
                    indicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                    ),
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
                      editScreen(mobileScreen: true),
                    ],
                  ),
          );
        },
      ),
    );
  }
}
