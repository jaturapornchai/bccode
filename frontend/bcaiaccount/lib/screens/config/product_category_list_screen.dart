import 'package:flutter/foundation.dart';

import 'package:smlaicloud/bloc/product/product_bloc.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/bloc/product_category/product_category_bloc.dart';
import 'package:smlaicloud/components/product_label_print.dart';
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/screen_search/product_category_search_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_draggable_gridview/flutter_draggable_gridview.dart';
import 'package:file_picker/file_picker.dart';
import 'package:excel/excel.dart' as xl;
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductCategoryListScreen extends StatefulWidget {
  const ProductCategoryListScreen({super.key});

  @override
  State<ProductCategoryListScreen> createState() =>
      ProductCategoryListScreenState();
}

class ProductCategoryListScreenState extends State<ProductCategoryListScreen>
    with SingleTickerProviderStateMixin {
  TextEditingController searchController = TextEditingController();
  List<DraggableGridItem> draggableGridItemList = [];
  List<ProductBarcodeModel> listData = [];
  bool loadingData = false;
  bool isDataChange = false;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  final _debouncer = global.Debouncer(1000);
  ScrollController listScrollController = ScrollController();
  List<SearchCodeAndNameAndUnitModel> productList = [];
  String searchText = "";
  String selectGuid = "";
  String selectProductCategoryGuid = "";
  String selectProductCategoryName = "";
  ProductCategoryModel? selectProductCategory;
  late SplitViewController splitViewController;
  int selectedGroupNumber = 1;

  // Define a list of colors for the chips
  final List<Color> chipColors = [
    Colors.red,
    Colors.green,
    Colors.blue,
    Colors.yellow,
    Colors.orange,
    Colors.pink,
    Colors.teal,
    Colors.cyan,
    Colors.brown,
    Colors.grey,
    Colors.lime,
    Colors.amber,
    Colors.deepOrange,
    // Add more colors as needed
  ];

  /// fillter product is not in category
  bool isProductIsNotCategory = false;

  List<ProductCategoryModel> productCategoryList = [];
  @override
  void initState() {
    loadDataCategory();
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

  void rebuildGrid(List<SearchCodeAndNameAndUnitModel> data) {
    setState(() {
      draggableGridItemList = data
          .map((e) => DraggableGridItem(child: objectBox(e), isDraggable: true))
          .toList();
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
              style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: Colors.blue),
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

  // Function to get a color based on category guid
  Color getColorFromGuid(String guid) {
    // Hash the guid to get a consistent index
    final index = guid.hashCode % chipColors.length;
    return chipColors[index];
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
        child: Column(
          children: [
            Row(
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
                Expanded(
                  flex: 5,
                  child: Text(
                    value.itemcode!,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Expanded(
                  flex: 5,
                  child: Text(
                    global.packName(value.groupnames!),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                IconButton(
                  onPressed: (!checkProductExists(value.barcode ?? ''))
                      ? () {
                          isDataChange = true;
                          productList.add(
                            SearchCodeAndNameAndUnitModel(
                              barcode: value.barcode ?? '',
                              code: value.itemcode ?? '',
                              name: value.names ?? [],
                              unitcode: value.itemunitcode,
                              unitname: value.itemunitnames ?? [],
                            ),
                          );
                          rebuildGrid(productList);
                          setState(() {});
                        }
                      : null,
                  icon: const Icon(Icons.add),
                ),
              ],
            ),
            (value.categorys!.isNotEmpty)
                ? Padding(
                    padding: const EdgeInsets.only(top: 0),
                    child: Align(
                      alignment: Alignment.centerLeft,
                      child: Wrap(
                        spacing: 4.0,
                        runSpacing: 4.0,
                        children: value.categorys!.asMap().entries.map((entry) {
                          final category = entry.value;
                          final color = getColorFromGuid(category.guidfixed);
                          return Chip(
                            backgroundColor: color,
                            label: Text(global.activeLangName(category.names!)),
                          );
                        }).toList(),
                      ),
                    ),
                  )
                : Container(),
          ],
        ),
      ),
    );
  }

  bool checkProductExists(String barcode) {
    return productList.any((product) => product.barcode == barcode);
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

  void loadDataCategory() {
    context.read<ProductCategoryBloc>().add(const ProductCategoryLoadList());
  }

  Widget listScreen() {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(2),
              boxShadow: [
                BoxShadow(
                  color: Colors.grey.withValues(alpha: 0.5),
                  spreadRadius: 5,
                  blurRadius: 7,
                  offset: const Offset(0, 2),
                ),
              ],
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
                IconButton(
                  onPressed: () async {
                    setState(() {
                      isProductIsNotCategory = !isProductIsNotCategory;
                      listData = [];
                    });
                    loadDataList(searchText);
                  },
                  icon: Icon(
                    (!isProductIsNotCategory)
                        ? Icons.filter_alt_off
                        : Icons.filter_alt,
                    color: (!isProductIsNotCategory)
                        ? Colors.black
                        : Colors.blue,
                  ),
                ),
              ],
            ),
          ),
          Container(color: global.theme.appBarColor, height: 6),
          Container(
            key: const Key('barcode_list'),
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
                Expanded(
                  flex: 5,
                  child: Text(
                    global.language("item_code"),
                    style: const TextStyle(
                      color: Colors.black,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                Expanded(
                  flex: 5,
                  child: Text(
                    global.language("product_group"),
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
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  // ฟังก์ชันสำหรับ import Excel
  Future<void> _importFromExcel() async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx', 'xls'],
      );

      if (result != null && result.files.single.bytes != null) {
        var bytes = result.files.single.bytes!;
        var excel = xl.Excel.decodeBytes(bytes);

        List<ProductCategoryCodeListModel> importedData = [];

        for (var table in excel.tables.keys) {
          var sheet = excel.tables[table];
          if (sheet != null) {
            // ข้าม header row (row 0)
            for (int row = 1; row < sheet.maxRows; row++) {
              try {
                var rowData = sheet.row(row);

                // ตรวจสอบว่ามีข้อมูลใน row นี้หรือไม่
                if (rowData.isNotEmpty && rowData[0]?.value != null) {
                  var codeListModel = _parseExcelRowToModel(rowData);
                  if (codeListModel != null) {
                    importedData.add(codeListModel);
                  }
                }
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('Error parsing row $row: $e');
                }
                continue;
              }
            }
          }
        }

        if (importedData.isNotEmpty) {
          // อัพเดท codelist ของ category ที่เลือก
          selectProductCategory?.codelist = importedData;

          // อัพเดท productList ให้ตรงกับข้อมูลที่ import
          productList.clear();
          for (var item in importedData) {
            productList.add(
              SearchCodeAndNameAndUnitModel(
                barcode: item.barcode,
                code: item.code,
                name: item.names ?? [],
                unitcode: item.unitcode,
                unitname: item.unitnames ?? [],
              ),
            );
          }

          // สร้าง grid ใหม่
          rebuildGrid(productList);

          // แสดงข้อความสำเร็จ
          global.showSuccessSnackBar(context, '${global.language("import_success")}: ${importedData.length} ${global.language("items")}');

          setState(() {
            isDataChange = true;
          });
        } else {
          global.showWarningSnackBar(context, global.language("no_data_found_in_excel"));
        }
      }
    } catch (e) {
      global.showErrorSnackBar(context, '${global.language("import_error")}: $e');
    }
  }

  // ฟังก์ชันแปลง Excel row เป็น ProductCategoryCodeListModel
  ProductCategoryCodeListModel? _parseExcelRowToModel(List<xl.Data?> rowData) {
    try {
      // Expected columns: code, xorder, barcode, unitcode, unitnames, names, manufacturerguid
      if (rowData.length < 6) return null;

      String code = rowData[0]?.value?.toString() ?? '';
      int xorder = int.tryParse(rowData[1]?.value?.toString() ?? '0') ?? 0;
      String barcode = rowData[2]?.value?.toString() ?? '';
      String unitcode = rowData[3]?.value?.toString() ?? '';
      String unitnames = rowData[4]?.value?.toString() ?? '';
      String names = rowData[5]?.value?.toString() ?? '';

      if (code.isEmpty || barcode.isEmpty) return null;

      // สร้าง LanguageDataModel สำหรับ unitnames
      List<LanguageDataModel> unitnamesList = [
        LanguageDataModel(code: "th", name: unitnames),
      ];

      // สร้าง LanguageDataModel สำหรับ names
      List<LanguageDataModel> namesList = [
        LanguageDataModel(code: "th", name: names),
      ];

      return ProductCategoryCodeListModel(
        code: code,
        xorder: xorder,
        barcode: barcode,
        unitcode: unitcode,
        unitnames: unitnamesList,
        names: namesList,
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error parsing row data: $e');
      }
      return null;
    }
  }

  Widget gridView() {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Container(
          padding: EdgeInsets.all(5),
          child: Column(
            children: [
              Row(
                children: [
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
                              label:
                                  "${global.language("group_number")} : $value",
                            );
                          })
                          .toList(),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: SizedBox(
                      height: 50,
                      child: ElevatedButton(
                        onPressed: () {
                          discardData(
                            callBack: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) =>
                                      ProductCategorySearchScreen(
                                        groupnumber: selectedGroupNumber,
                                      ),
                                ),
                              ).then((value) {
                                if (value != null) {
                                  if (value[0].isNotEmpty) {
                                    selectProductCategoryGuid = value[0];
                                    selectProductCategoryName = global.packName(
                                      value[1],
                                    );
                                    context.read<ProductCategoryBloc>().add(
                                      ProductCategoryGet(
                                        guid: selectProductCategoryGuid,
                                      ),
                                    );
                                  }
                                }
                              });
                            },
                          );
                        },
                        child: Text(
                          (selectProductCategoryGuid.isEmpty)
                              ? global.language("select_product_category")
                              : selectProductCategoryName,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  SizedBox(
                    width: 80,
                    height: 50,
                    child: ElevatedButton(
                      onPressed: (selectProductCategoryGuid.isNotEmpty)
                          ? () {
                              _importFromExcel();
                            }
                          : null,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.orange,
                      ),
                      child: const Icon(Icons.upload_file, color: Colors.white),
                    ),
                  ),
                  const SizedBox(width: 5),
                  SizedBox(
                    width: 80,
                    height: 50,
                    child: ElevatedButton(
                      onPressed: (selectProductCategoryGuid.isNotEmpty)
                          ? () {
                              selectProductCategory?.codelist = [];
                              for (int i = 0; i < productList.length; i++) {
                                selectProductCategory?.codelist!.add(
                                  ProductCategoryCodeListModel(
                                    xorder: i,
                                    barcode: productList[i].barcode,
                                    unitcode: productList[i].unitcode,
                                    unitnames: productList[i].unitname,
                                    code: productList[i].code,
                                    names: productList[i].name,
                                  ),
                                );
                              }
                              context.read<ProductCategoryBloc>().add(
                                ProductCategoryUpdate(
                                  guid: selectProductCategoryGuid,
                                  category: selectProductCategory!,
                                ),
                              );
                            }
                          : null,
                      child: const Icon(Icons.save, color: Colors.white),
                    ),
                  ),
                ],
              ),
              SizedBox(height: 10),
              Expanded(
                child: (productList.isEmpty)
                    ? Center(
                        child: Text(global.language("no_product_selected")),
                      )
                    : DraggableGridViewBuilder(
                        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                          crossAxisCount: constraints.maxWidth ~/ 150,
                          childAspectRatio: 1,
                        ),
                        children: draggableGridItemList,
                        dragCompletion:
                            (
                              List<DraggableGridItem> list,
                              int beforeIndex,
                              int afterIndex,
                            ) {
                              setState(() {
                                final item = productList.removeAt(beforeIndex);
                                productList.insert(afterIndex, item);
                                rebuildGrid(productList);
                                AppLogger.debug(
                                  'onDragAccept: $beforeIndex -> $afterIndex',
                                );
                              });
                            },
                        isOnlyLongPress: true,
                        dragFeedback:
                            (List<DraggableGridItem> list, int index) {
                              return objectBox(productList[index]);
                            },
                        dragPlaceHolder:
                            (List<DraggableGridItem> list, int index) {
                              return PlaceHolderWidget(
                                child: Container(color: Colors.white),
                              );
                            },
                      ),
              ),
            ],
          ),
        );
      },
    );
  }

  /// objectbox
  Widget objectBox(SearchCodeAndNameAndUnitModel item) {
    return Container(
      width: 150,
      height: 150,
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: Colors.grey),
      ),
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(4),
            width: double.infinity,
            child: ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              onPressed: () {
                isDataChange = true;
                productList.removeWhere(
                  (element) => element.barcode == item.barcode,
                );
                rebuildGrid(productList);
              },
              child: const Icon(Icons.close),
            ),
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.only(top: 10, left: 10, right: 10),
              child: Column(
                children: [
                  Text(
                    item.barcode,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Expanded(
                    child: Text(
                      global.packName(item.name),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 12),
                    ),
                  ),
                  Text(
                    item.unitcode,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Expanded(
                    child: Text(
                      global.packName(item.unitname),
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 12),
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

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return MultiBlocListener(
          listeners: [
            BlocListener<ProductBloc, ProductState>(
              listener: (context, state) {
                if (state is ProductGetByBarcodesSuccess) {
                  ProductLabelPrint.showPdfPreview(context, state.products);
                } else if (state is ProductGetByBarcodesFailed) {
                  global.showErrorSnackBar(context, state.message);
                }
              },
            ),
            BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
              listener: (context, state) {
                // Load
                if (state is ProductBarcodeLoadSuccess) {
                  setState(() {
                    loadingData = false;
                    if (state.productBarcodes.isNotEmpty) {
                      listData.addAll(state.productBarcodes);

                      for (int i = 0; i < listData.length; i++) {
                        for (int j = 0; j < productCategoryList.length; j++) {
                          for (
                            int k = 0;
                            k < productCategoryList[j].codelist!.length;
                            k++
                          ) {
                            if (listData[i].barcode ==
                                productCategoryList[j].codelist![k].barcode) {
                              listData[i].categorys!.add(
                                productCategoryList[j],
                              );
                            }
                          }
                        }
                      }

                      if (isProductIsNotCategory) {
                        listData.removeWhere(
                          (element) => element.categorys!.isNotEmpty,
                        );
                      }
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
            BlocListener<ProductCategoryBloc, ProductCategoryState>(
              listener: (context, state) {
                if (state is ProductCategoryLoadSuccess) {
                  productCategoryList = state.productCategorys;
                } else if (state is ProductCategoryLoadFailed) {
                  global.showSnackBar(
                    context,
                    const Icon(Icons.error, color: Colors.white),
                    state.message,
                    Colors.red,
                  );
                }
                // Load
                if (state is ProductCategoryGetSuccess) {
                  productList = [];
                  selectProductCategory = state.category;
                  for (int i = 0; i < state.category.codelist!.length; i++) {
                    productList.add(
                      SearchCodeAndNameAndUnitModel(
                        barcode: state.category.codelist![i].barcode,
                        code: state.category.codelist![i].code,
                        name: state.category.codelist![i].names!,
                        unitcode: state.category.codelist![i].unitcode,
                        unitname: state.category.codelist![i].unitnames!,
                      ),
                    );
                  }
                  rebuildGrid(productList);
                  setState(() {});
                }
                if (state is ProductCategoryUpdateSuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: Colors.white),
                      global.language("update_success"),
                      Colors.blue,
                    );
                    productList = [];
                    isDataChange = false;
                    selectGuid = "";
                    selectProductCategoryGuid = "";
                    selectProductCategoryName = "";
                    listData = [];
                    loadDataCategory();
                    loadDataList("");
                  });
                }
                if (state is ProductCategoryUpdateFailed) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: Colors.white),
                      "${global.language("update_failed")} : ${state.message}",
                      Colors.red,
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
              title: Text(global.language('product_category')),
              leading: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () {
                  discardData(
                    callBack: () {
                      // Navigator.pop(context);
                      global.gotoMainMenu(context);
                    },
                  );
                },
              ),
              actions: [
                const ManualButton(path: 'product-category-assign'),
                (productList.isNotEmpty)
                    ? IconButton(
                        onPressed: () {
                          final List<String> listbarcodes = [];
                          for (int i = 0; i < productList.length; i++) {
                            listbarcodes.add(productList[i].barcode);
                          }

                          context.read<ProductBloc>().add(
                            ProductGetByBarcodes(barcodes: listbarcodes),
                          );
                        },
                        icon: const Icon(Icons.print),
                      )
                    : Container(),
              ],
            ),
            body: (constraints.maxWidth < 800.0)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 8,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: Colors.blue,
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
                    gripColorActive: Colors.blue,
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
