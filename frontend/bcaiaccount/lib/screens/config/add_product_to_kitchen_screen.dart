import 'package:smlaicloud/bloc/kitchen/kitchen_bloc.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/kitchen_model.dart';
import 'package:smlaicloud/model/kitchen_product_model.dart';
import 'package:smlaicloud/screen_search/kitchen_search_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_draggable_gridview/flutter_draggable_gridview.dart';

class AddProductToKitchenScreen extends StatefulWidget {
  const AddProductToKitchenScreen({super.key});

  @override
  State<AddProductToKitchenScreen> createState() =>
      AddProductToKitchenScreenState();
}

class AddProductToKitchenScreenState extends State<AddProductToKitchenScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  List<DraggableGridItem> draggableGridItemList = [];
  bool loadingData = false;
  bool loadingDataBarcode = false;
  bool isDataChange = false;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  final _debouncer = global.Debouncer(500);
  List<KitchenProductModel> listData = [];
  List<KitchenProductModel> listDataTeamp = [];
  List<KitchenProductModel> productList = [];
  List<ProductInKitchenModel> kitchenProducts = [];
  String searchText = "";
  int selectedGroupNumber = 1;

  late SplitViewController splitViewController;
  List<String>? selectBarcode;
  late KitchenModel screenData;

// Define a list of colors for the chips
  final List<Color> chipColors = [
    global.theme.negativeHighlightTextColor,
    global.theme.positiveHighlightTextColor,
    global.theme.infoHighlightTextColor,
    global.theme.warningHighlightTextColor,
    global.theme.warningHighlightTextColor,
    global.theme.negativeHighlightTextColor,
    global.theme.infoHighlightTextColor,
    global.theme.infoHighlightTextColor,
    global.theme.textSecondaryColor,
    global.theme.iconSecondaryColor,
    global.theme.positiveHighlightTextColor,
    global.theme.warningHighlightTextColor,
    global.theme.warningHighlightTextColor,
    global.theme.infoHighlightTextColor,
    // Add more colors as needed
  ];

  @override
  void initState() {
    clearData();
    loadDataList();

    splitViewController =
        SplitViewController(limits: [null, WeightLimit(min: 0.2, max: 0.8)]);
    super.initState();
  }

  @override
  void dispose() {
    searchController.dispose();
    searchFocusNode.dispose();
    super.dispose();
  }

  void clearData() {
    setState(() {
      loadingData = false;
      isDataChange = false;
      listData = [];
      listDataTeamp = [];
      productList = [];
      searchText = "";
      screenData = KitchenModel(
        guidfixed: "",
        code: "",
        names: [],
        printers: [],
        products: [],
        zones: [],
      );
    });
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
                      style:
                          ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
                      onPressed: () => Navigator.pop(context),
                      child: Text(global.language('no'))),
                  ElevatedButton(
                      style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.infoHighlightTextColor),
                      onPressed: () {
                        Navigator.pop(context);
                        callBack();
                      },
                      child: Text(global.language('yes'))),
                ],
              ));
    } else {
      callBack();
    }
  }

  // Function to get a color based on category guid
  Color getColorFromGuid(String code) {
    // Hash the guid to get a consistent index
    final index = code.hashCode % chipColors.length;
    return chipColors[index];
  }

  Widget listObject(int index, KitchenProductModel value) {
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
            bottom: global.deviceConfig.listDataLineSpace),
        child: Column(
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                    flex: 5,
                    child: Text(value.barcode,
                        maxLines: 2, overflow: TextOverflow.ellipsis)),
                Expanded(
                    flex: 10,
                    child: Text(
                      global.packName(value.names),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    )),
                IconButton(
                    onPressed: (value.isdisable == false &&
                            screenData.guidfixed!.isNotEmpty)
                        ? () {
                            bool found = false;
                            for (int i = 0; i < productList.length; i++) {
                              if (productList[i].guidfixed == value.guidfixed) {
                                found = true;
                                break;
                              }
                            }
                            if (found == false) {
                              isDataChange = true;
                              productList.add(KitchenProductModel(
                                guidfixed: value.guidfixed,
                                barcode: value.barcode,
                                names: value.names,
                              ));
                              listData[index].isdisable = true;
                              setState(() {});
                            }
                          }
                        : null,
                    icon: Icon(Icons.add)),
              ],
            ),
            (value.kitchens!.isNotEmpty)
                ? Padding(
                    padding: const EdgeInsets.only(top: 0),
                    child: Align(
                      alignment: Alignment.centerLeft,
                      child: Wrap(
                        spacing: 4.0,
                        runSpacing: 4.0,
                        children: value.kitchens!.asMap().entries.map((entry) {
                          final kitchen = entry.value;
                          final color = getColorFromGuid(kitchen.code);
                          return Chip(
                            backgroundColor: color,
                            label: Text(global.activeLangName(kitchen.names)),
                          );
                        }).toList(),
                      ),
                    ),
                  )
                : Container()
          ],
        ),
      ),
    );
  }

  Widget listProduct(int index, KitchenProductModel value) {
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
                bottom: global.deviceConfig.listDataLineSpace),
            child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Expanded(
                  flex: 5,
                  child: Text(value.barcode,
                      maxLines: 2, overflow: TextOverflow.ellipsis)),
              Expanded(
                  flex: 10,
                  child: Text(
                    global.packName(value.names),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  )),
              IconButton(
                  onPressed: () {
                    isDataChange = true;

                    productList.removeAt(index);
                    var foundValue = listData.firstWhere(
                      (element) => element.guidfixed == value.guidfixed,
                      orElse: () => KitchenProductModel(
                        barcode: '',
                        guidfixed: '',
                        names: [],
                      ),
                    );

                    if (foundValue.guidfixed.isNotEmpty) {
                      listData
                          .firstWhere(
                              (element) => element.guidfixed == value.guidfixed)
                          .isdisable = false;
                    }

                    setState(() {});
                  },
                  icon: Icon(Icons.delete)),
            ])));
  }

  void loadDataList() {
    setState(() {
      loadingData = true;
    });
    context.read<ProductBarcodeBloc>().add(ProductBarcodeLoadList(
          offset: 0,
          limit: 2000,
          search: '',
          branchcode: global.companyBranchSelectData.code,
          businesstypecode: global.companyBranchSelectData.businesstype!.code!,
        ));
  }

  Widget listScreen() {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: Column(children: [
        Container(
            height: 40,
            padding: EdgeInsets.all(5),
            color: global.theme.searchBarColor,
            child: Row(children: [
              Expanded(
                  child: TextField(
                      onSubmitted: (value) {
                        searchFocusNode.requestFocus();
                      },
                      onChanged: (value) {
                        loadingData = true;
                        _debouncer.run(() {
                          setState(() {
                            searchText = value;

                            /// filter listDataTamp by search text and add to listData
                            listData = listDataTeamp
                                .where((element) =>
                                    element.barcode
                                        .contains(searchText.toUpperCase()) ||
                                    global
                                        .packName(element.names)
                                        .contains(searchText))
                                .toList();

                            for (var element in listData) {
                              var foundValue = productList.firstWhere(
                                (product) =>
                                    product.guidfixed == element.guidfixed,
                                orElse: () => KitchenProductModel(
                                  barcode: '',
                                  guidfixed: '',
                                  names: [],
                                ),
                              );

                              if (foundValue.guidfixed.isNotEmpty) {
                                element.isdisable = true;
                              } else {
                                element.isdisable = false;
                              }
                            }

                            /// set Delay 500 ms
                            Future.delayed(const Duration(milliseconds: 500),
                                () {
                              setState(() {
                                loadingData = false;
                              });
                            });
                          });
                        });
                      },
                      autofocus: false,
                      focusNode: searchFocusNode,
                      controller: searchController,
                      decoration: InputDecoration(
                        isDense: true,
                        contentPadding: EdgeInsets.only(
                            top: 0, bottom: 0, left: 0, right: 0),
                        border: InputBorder.none,
                        hintText: global.language('search'),
                      ))),
            ])),
        Container(
          color: global.theme.appBarColor,
          height: 6,
        ),
        Container(
            key: const Key('barcode_list'),
            padding:
                EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
            color: global.theme.columnHeaderColor,
            child: Row(children: [
              Expanded(
                  flex: 5,
                  child: Text(global.language("barcode"),
                      style: TextStyle(
                          color: global.theme.textColor, fontWeight: FontWeight.bold))),
              Expanded(
                  flex: 10,
                  child: Text(
                    global.language("product_name"),
                    style: TextStyle(
                        color: global.theme.textColor, fontWeight: FontWeight.bold),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  )),
            ])),
        Expanded(
          child: ListView(
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
          ))
      ]),
    );
  }

  Widget gridView() {
    return LayoutBuilder(builder: (context, constraints) {
      return Container(
          padding: EdgeInsets.all(5),
          child: Column(
            children: [
              Row(children: [
                SizedBox(
                  height: 50,
                  child: DropdownMenu<int>(
                    initialSelection: global.groupNumber.first,
                    onSelected: (int? value) {
                      // This is called when the user selects an item.
                      setState(() {
                        selectedGroupNumber = value!;
                      });
                    },
                    dropdownMenuEntries: global.groupNumber
                        .map<DropdownMenuEntry<int>>((int value) {
                      return DropdownMenuEntry<int>(
                          value: value,
                          label: "${global.language("group_number")} : $value");
                    }).toList(),
                  ),
                ),
                const SizedBox(
                  width: 10,
                ),
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.only(bottom: 5),
                    width: double.infinity,
                    height: 50,
                    child: ElevatedButton(
                      onPressed: () {
                        discardData(callBack: () {
                          Navigator.push(
                              context,
                              MaterialPageRoute(
                                  builder: (context) => KitchenSearchScreen(
                                        groupnumber: selectedGroupNumber,
                                      ))).then((value) {
                            setState(() {
                              SearchGuidCodeNameModel result = value;
                              if (result.guid.isNotEmpty &&
                                  screenData.guidfixed != result.guid) {
                                screenData = KitchenModel(
                                  guidfixed: result.guid,
                                  code: result.code,
                                  names: result.names,
                                );

                                loadingDataBarcode = true;

                                productList = [];
                                context.read<KitchenBloc>().add(
                                      KitchenDetailBarcodeGet(
                                        guid: result.guid,
                                      ),
                                    );
                              }
                            });
                          });
                        });
                      },
                      child: Text(
                        (screenData.guidfixed!.isEmpty)
                            ? global.language("select_kitchen")
                            : "${screenData.code} ~ ${global.packName(screenData.names)}",
                      ),
                    ),
                  ),
                ),
                Container(
                    padding: const EdgeInsets.only(left: 5, bottom: 5),
                    width: 100,
                    height: 50,
                    child: ElevatedButton(
                        onPressed: ((screenData.guidfixed!.isNotEmpty))
                            ? () {
                                setState(() {
                                  selectBarcode = [];

                                  for (int i = 0; i < productList.length; i++) {
                                    selectBarcode!.add(productList[i].barcode);
                                  }
                                  screenData.groupnumber = selectedGroupNumber;
                                  screenData.products = selectBarcode!;

                                  // print(screenData.toJson());
                                  context.read<KitchenBloc>().add(KitchenUpdate(
                                        guid: screenData.guidfixed!,
                                        kitchenModel: screenData,
                                      ));
                                });
                              }
                            : null,
                        child: Icon(
                          Icons.save,
                          color: global.theme.onPrimaryColor,
                        ))),
              ]),
              Expanded(
                child: (productList.isEmpty)
                    ? Column(
                        children: [
                          Container(
                              key: const Key('product_list_header'),
                              padding: EdgeInsets.only(
                                  left: 10, right: 10, top: 5, bottom: 5),
                              color: global.theme.columnHeaderColor,
                              child: Row(children: [
                                Expanded(
                                    flex: 5,
                                    child: Text(global.language("barcode"),
                                        style: TextStyle(
                                            color: global.theme.textColor,
                                            fontWeight: FontWeight.bold))),
                                Expanded(
                                    flex: 10,
                                    child: Text(
                                      global.language("product_name"),
                                      style: TextStyle(
                                          color: global.theme.textColor,
                                          fontWeight: FontWeight.bold),
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    )),
                              ])),
                          if (!loadingDataBarcode)
                            Center(
                              child: Text(
                                global.language("no_select_product"),
                              ),
                            ),
                          if (loadingDataBarcode)
                            Center(
                              child: LoadingAnimationWidget.staggeredDotsWave(
                                color: global.theme.infoHighlightTextColor,
                                size: 50,
                              ),
                            )
                        ],
                      )
                    : Column(
                        children: [
                          Container(
                              key: const Key('product_list_header'),
                              padding: EdgeInsets.only(
                                  left: 10, right: 10, top: 5, bottom: 5),
                              color: global.theme.columnHeaderColor,
                              child: Row(children: [
                                Expanded(
                                    flex: 5,
                                    child: Text(global.language("barcode"),
                                        style: TextStyle(
                                            color: global.theme.textColor,
                                            fontWeight: FontWeight.bold))),
                                Expanded(
                                    flex: 10,
                                    child: Text(
                                      global.language("product_name"),
                                      style: TextStyle(
                                          color: global.theme.textColor,
                                          fontWeight: FontWeight.bold),
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    )),
                              ])),
                          Expanded(
                              child: ListView(
                                  children: productList
                                      .map((value) => listProduct(
                                          productList.indexOf(value), value))
                                      .toList())),
                        ],
                      ),
              ),
            ],
          ));
    });
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(builder: (context, constraints) {
      return MultiBlocListener(
        listeners: [
          BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
              listener: (context, state) {
            // Load
            if (state is ProductBarcodeLoadSuccess) {
              setState(() {
                loadingData = false;
                if (state.productBarcodes.isNotEmpty) {
                  for (var productBarcode in state.productBarcodes) {
                    if (productList
                        .where((element) =>
                            element.guidfixed == productBarcode.guidfixed)
                        .isEmpty) {
                      listData.add(KitchenProductModel(
                        guidfixed: productBarcode.guidfixed,
                        barcode: productBarcode.barcode!,
                        names: productBarcode.names!,
                        isdisable: false,
                        kitchens: [],
                      ));
                      listDataTeamp.add(KitchenProductModel(
                        guidfixed: productBarcode.guidfixed,
                        barcode: productBarcode.barcode!,
                        names: productBarcode.names!,
                        isdisable: true,
                        kitchens: [],
                      ));
                    } else {
                      listData.add(KitchenProductModel(
                        guidfixed: productBarcode.guidfixed,
                        barcode: productBarcode.barcode!,
                        names: productBarcode.names!,
                        isdisable: true,
                        kitchens: [],
                      ));
                      listDataTeamp.add(KitchenProductModel(
                        guidfixed: productBarcode.guidfixed,
                        barcode: productBarcode.barcode!,
                        names: productBarcode.names!,
                        isdisable: true,
                        kitchens: [],
                      ));
                    }
                  }
                  context.read<KitchenBloc>().add(KitchenProductsLoadList());
                }
              });
            }

            if (state is ProductBarcodeLoadFailed) {
              setState(() {
                loadingData = false;
              });
            }
          }),
          BlocListener<KitchenBloc, KitchenState>(listener: (context, state) {
            if (state is KitchenProductsLoadSuccess) {
              setState(() {
                kitchenProducts = [];
                if (state.kitchens.isNotEmpty) {
                  for (var kitchen in state.kitchens) {
                    kitchenProducts.add(ProductInKitchenModel(
                      barcode: kitchen.barcode,
                      kitchens: kitchen.kitchens,
                    ));
                  }

                  for (var element in listData) {
                    var foundValue = kitchenProducts.firstWhere(
                      (product) => product.barcode == element.barcode,
                      orElse: () => ProductInKitchenModel(
                        barcode: '',
                        kitchens: [],
                      ),
                    );

                    if (foundValue.barcode!.isNotEmpty) {
                      element.kitchens = foundValue.kitchens;
                    }
                  }

                  for (var element in listDataTeamp) {
                    var foundValue = kitchenProducts.firstWhere(
                      (product) => product.barcode == element.barcode,
                      orElse: () => ProductInKitchenModel(
                        barcode: '',
                        kitchens: [],
                      ),
                    );

                    if (foundValue.barcode!.isNotEmpty) {
                      element.kitchens = foundValue.kitchens;
                    }
                  }
                }
              });
            }
            //
            if (state is KitchenDetailBarcodeGetSuccess) {
              setState(() {
                if (state.kitchen.products!.isNotEmpty) {
                  for (var products in state.kitchen.products!) {
                    productList.add(KitchenProductModel(
                      guidfixed: products.guidfixed,
                      barcode: products.barcode,
                      names: products.names,
                    ));
                  }
                  for (var element in listData) {
                    var foundValue = productList.firstWhere(
                      (product) => product.guidfixed == element.guidfixed,
                      orElse: () => KitchenProductModel(
                        barcode: '',
                        guidfixed: '',
                        names: [],
                      ),
                    );

                    if (foundValue.guidfixed.isNotEmpty) {
                      element.isdisable = true;
                    } else {
                      element.isdisable = false;
                    }
                  }
                }

                loadingDataBarcode = false;
              });
            }
            if (state is KitchenUpdateSuccess) {
              setState(() {
                global.showSnackBar(
                    context,
                    Icon(
                      Icons.edit,
                      color: global.theme.onPrimaryColor,
                    ),
                    global.language('edit_success'),
                    global.theme.infoHighlightTextColor);

                setState(() {
                  loadDataList();
                  clearData();
                });
              });
            }
            if (state is KitchenUpdateFailed) {
              setState(() {
                global.showSnackBar(
                    context,
                    Icon(
                      Icons.edit,
                      color: global.theme.onPrimaryColor,
                    ),
                    "${global.language('edit_failed')} : ${state.message}",
                    global.theme.negativeHighlightTextColor);
              });
            }
          })
        ],
        child: Scaffold(
            resizeToAvoidBottomInset: true,
            appBar: AppBar(
              backgroundColor: global.theme.appBarColor,
              automaticallyImplyLeading: false,
              title: Text(global.language('add_product_to_kitchen')),
              leading: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.arrow_back),
                onPressed: () {
                  discardData(callBack: () {
                    Navigator.pop(context);
                  });
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
                    indicator:
                        const SplitIndicator(viewMode: SplitViewMode.Vertical),
                    activeIndicator: const SplitIndicator(
                      viewMode: SplitViewMode.Vertical,
                      isActive: true,
                    ),
                    children: [
                        listScreen(),
                        gridView(),
                      ])
                : SplitView(
                    controller: splitViewController,
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
                        listScreen(),
                        gridView(),
                      ])),
      );
    });
  }
}
