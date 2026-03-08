import 'dart:io';
import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/product_dimension/product_dimension_bloc.dart';
import 'package:smlaicloud/bloc/product_group/product_group_bloc.dart';
import 'package:smlaicloud/bloc/productmaster/productmaster_bloc.dart';
import 'package:smlaicloud/model/creditor_model.dart';
import 'package:smlaicloud/model/dimension_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_group_model.dart';
import 'package:smlaicloud/screen_search/product_group_search_screen.dart';
import 'package:smlaicloud/screen_search/supplier_search_screen.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductScreen extends StatefulWidget {
  const ProductScreen({super.key});

  @override
  State<ProductScreen> createState() => ProductScreenState();
}

class ProductScreenState extends State<ProductScreen>
    with SingleTickerProviderStateMixin {
  final translator = GoogleTranslator();

  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<TextEditingController> fieldTextController = [];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<ProductMasterModel> listData = [];
  String selectedUnitName = "";
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  List<ProductMasterBarcode> sortedBarcodes = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isEditMode = false;
  bool isSaveAllow = false;
  late ProductMasterState blocState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  List<DimensionModel> dimensions_list = [];
  List<ProductGroupModel> productgroup_list = [];
  List<CreditorModel> supplier_list = [];
  String selectedGroupName = ""; // ✅ ใช้เก็บชื่อกลุ่มสินค้า
  String selectedSupplierName = ""; // ✅ ใช้เก็บชื่อ Supplier
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debounce = global.Debouncer(1000);

  ProductMasterModel dataTemp = ProductMasterModel();

  bool isLoadTranslation = false;
  void changeScreenEvent(global.ScreenEventEnum event) {
    screenEvent = event;
    for (int index = 0; index < fieldFocusNodes.length; index++) {
      fieldFocusNodes[index].isReadOnly =
          (screenEvent == global.ScreenEventEnum.list) ? true : false;
    }
    if (screenEvent == global.ScreenEventEnum.edit) {
      fieldFocusNodes[0].isReadOnly = true;
    }
    setState(() {});
  }

  void setSystemLanguageList() async {
    try {
      await global.setSystemLanguage(context);
    } catch (ex) {
      // print(ex);
    }
    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController.add(TextEditingController());
      FocusNode focusNode = FocusNode();
      focusNode.addListener(() {
        focusNodeIndex = i;
      });
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    }
    setState(() {});
  }

  @override
  void initState() {
    loadDataList("");

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );

    List<String> fields = [
      "code",
      "groupguid",
      "unitguid",
      "itemtype",
      "manufacturerguid",
    ];

    for (var field in fields) {
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: FocusNode()));
      fieldTextController.add(TextEditingController());
    }

    fieldTextController.add(TextEditingController());
    loadDimension();
    setSystemLanguageList();
    loadProductGroup();

    loadSupplier();
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadSupplier() {
    context.read<CreditorBloc>().add(
      CreditorLoadList(offset: 0, limit: 500, search: '', groups: []),
    );
  }

  void loadProductGroup() {
    context.read<ProductGroupBloc>().add(
      ProductGroupLoadList(offset: 0, limit: 500, search: ''),
    );
  }

  void loadDimension() {
    context.read<ProductDimensionBloc>().add(
      ProductDimensionLoadList(offset: 0, limit: 500, search: ''),
    );
  }

  void loadDataList(String search) {
    context.read<ProductMasterBloc>().add(
      ProductMasterLoadList(
        page: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
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

    editScrollController.dispose();
    searchController.dispose();
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].dispose();
    }
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].focusNode.dispose();
    }

    super.dispose();
  }

  void clearEditData() {
    // 🔹 เคลียร์ค่าทั้งหมดใน TextEditingController
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }

    // 🔹 รีเซ็ตค่าใน dataTemp
    dataTemp = ProductMasterModel(
      guidfixed: "",
      code: "",
      names: [],
      groupguid: "",
      itemtype: 0,
      manufacturerguid: "",
      dimensions: [],
    );

    // 🔹 รีเซ็ตชื่อ (listNamesFields)
    for (var lang in languageList) {
      dataTemp.names.add(LanguageDataModel(code: lang.code!, name: ''));
    }

    // 🔹 รีเซ็ตชื่อที่แสดงแทน GUID
    selectedGroupName = "";
    selectedSupplierName = "";

    // 🔹 รีเซ็ตค่าการแก้ไข
    isChange = false;
    focusNodeIndex = 0;

    // 🔹 โฟกัสไปที่ช่องแรก (ตรวจสอบว่ามี FocusNode ก่อน)
    if (fieldFocusNodes.isNotEmpty) {
      try {
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error("Error focusing on node: $e");
        }
      }
    }

    // 🔹 อัปเดต UI
    setState(() {});
  }

  void discardData({required Function callBack}) {
    // กรณีมีการแก้ไขข้อมูล ให้เตือน
    if ((screenEvent == global.ScreenEventEnum.add ||
            screenEvent == global.ScreenEventEnum.edit) &&
        isChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonNoColor,
              ),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonYesColor,
              ),
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

  void getData(String guid) {
    headerEdit = global.language("show");
    changeScreenEvent(global.ScreenEventEnum.list);
    context.read<ProductMasterBloc>().add(ProductMasterGet(guid: guid));
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
            discardData(
              callBack: () {
                global.gotoMainMenu(context);
                changeScreenEvent(global.ScreenEventEnum.list);
              },
            );
          },
        ),
        actions: <Widget>[
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      if (showCheckBox) {
                        showCheckBox = false;
                        guidListChecked.clear();
                      } else {
                        showCheckBox = true;
                        global.showSnackBar(
                          context,
                          Icon(Icons.delete, color: Colors.white),
                          global.language("choose_item_delete"),
                          Colors.blue,
                        );
                      }
                    });
                  },
                );
              },
              icon: (showCheckBox)
                  ? const Icon(Icons.close)
                  : const Icon(Icons.check_box),
            ),
          ),
          if (guidListChecked.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showDialog<String>(
                    context: context,
                    builder: (BuildContext context) => AlertDialog(
                      title: Text(global.language('confirm_delete')),
                      actions: <Widget>[
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonNoColor,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonYesColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<ProductMasterBloc>().add(
                              ProductMasterDeleteMany(guid: guidListChecked),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      changeScreenEvent(global.ScreenEventEnum.add);
                      selectGuid = "";
                      showCheckBox = false;
                      isChange = false;
                      isEditMode = true;
                      clearEditData();
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        fieldFocusNodes[0].focusNode.requestFocus();
                      });
                    });
                  },
                );
              },
              icon: const Icon(Icons.add),
            ),
          ),
        ],
      ),
      body: Column(
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
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      debounce.run(() {
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
            color: global.theme.columnHeaderColor,
            key: headerKey,
            padding: const EdgeInsets.only(
              left: 10,
              right: 10,
              top: 5,
              bottom: 5,
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 5,
                  child: Text(
                    global.language("code"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
                Expanded(
                  flex: 10,
                  child: Text(
                    global.language("name"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (showCheckBox)
                  Expanded(
                    flex: 1,
                    child: Icon(
                      Icons.check,
                      color: global.theme.columnHeaderTextColor,
                      size: 12,
                    ),
                  ),
              ],
            ),
          ),
          Expanded(
            child: RawKeyboardListener(
              autofocus: true,
              focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
              onKey: (RawKeyEvent event) {
                if (screenEvent == global.ScreenEventEnum.list) {
                  if (kIsWeb ||
                      Platform.isWindows ||
                      Platform.isLinux ||
                      Platform.isMacOS) {
                    if (event is RawKeyUpEvent) {
                      try {
                        if (event.logicalKey == LogicalKeyboardKey.f2) {
                          isKeyDown = false;
                          searchFocusNode.requestFocus();
                        }
                        if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                          isKeyDown = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) => element.guidfixed == selectGuid,
                            ),
                          );
                          if (index > 0) {
                            selectGuid = listData[index - 1].guidfixed;
                            currentListIndex = index - 1;
                            isKeyUp = true;
                            getData(selectGuid);
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
                          getData(selectGuid);
                        }
                      } catch (_) {}
                    }
                  }
                }
              },
              child: ListView(
                controller: listScrollController,
                children: listData
                    .map(
                      (value) => listObject(
                        listData.indexOf(value),
                        value,
                        showCheckBox,
                      ),
                    )
                    .toList(),
              ),
            ),
          ),
        ],
      ),
    );
  }

  void switchToEdit(ProductMasterModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isEditMode = true;
      isSaveAllow = true;

      changeScreenEvent(global.ScreenEventEnum.edit);
    });
  }

  Widget listObject(int index, ProductMasterModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    bool selected = selectGuid == value.guidfixed;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );
    return GestureDetector(
      onTap: () {
        if (showCheckBox == true) {
          setState(() {
            selectGuid = value.guidfixed;
            if (isCheck == true) {
              guidListChecked.remove(value.guidfixed);
            } else {
              guidListChecked.add(value.guidfixed);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${guidListChecked.length} ${global.language("list")}",
              Colors.blue,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isEditMode = false;
                isSaveAllow = false;
                changeScreenEvent(global.ScreenEventEnum.list);
                selectGuid = value.guidfixed;
                getData(selectGuid);
              },
            );
          });
        }
      },
      onDoubleTap: () {
        if (showCheckBox == false) {
          switchToEdit(value);
        }
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        decoration: BoxDecoration(
          color: (selectGuid == value.guidfixed)
              ? Colors.cyan[100]
              : (index % 2 == 0)
              ? global.theme.columnAlternateEvenColor
              : global.theme.columnAlternateOddColor,
        ),
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
                value.code,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            if (showCheckBox)
              Expanded(
                flex: 1,
                child: (isCheck)
                    ? Icon(
                        Icons.check,
                        size: global.deviceConfig.listDataFontSize,
                      )
                    : Container(),
              ),
          ],
        ),
      ),
    );
  }

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i + 1].text,
          ),
        );
      }
    }

    for (var defaultValueLang in dataTemp.names) {
      LanguageDataModel result = names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        names.add(defaultValueLang);
      }
    }

    return names;
  }

  bool verifyData(ProductMasterModel value) {
    List<String> errorList = [];
    if (value.code.isEmpty) {
      errorList.add(global.language("code"));
    }
    if (value.names.isEmpty || value.names[0].name.isEmpty) {
      errorList.add(global.language("name"));
    }
    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: Colors.white),
        "${global.language("not_success_save")} ${errorList.join(", ")}",
        Colors.red,
      );
      return false;
    }
    return true;
  }

  void saveOrUpdateData() {
    if (screenEvent == global.ScreenEventEnum.edit ||
        screenEvent == global.ScreenEventEnum.add) {
      ProductMasterModel product = packProductData();
      if (verifyData(product)) {
        showCheckBox = false;
        if (selectGuid.trim().isEmpty) {
          context.read<ProductMasterBloc>().add(
            ProductMasterSave(productMasterModel: product),
          );
        } else {
          context.read<ProductMasterBloc>().add(
            ProductMasterUpdate(guid: selectGuid, productMasterModel: product),
          );
        }
      }
    }
  }

  Widget groupGuidSearchField() {
    return searchField(
      label: global.language("product_group"),
      selectedName: selectedGroupName,
      searchFunction: () => searchProductGroup(
        callBack:
            (
              bool success,
              String groupGuid,
              List<LanguageDataModel> groupNames,
            ) {
              if (success) {
                setState(() {
                  fieldTextController[1 + languageList.length].text =
                      groupGuid; // ✅ บันทึก GUID
                  selectedGroupName = global.activeLangName(
                    groupNames,
                  ); // ✅ แสดงชื่อ
                });
              }
            },
      ),
      clearFunction: () {
        setState(() {
          selectedGroupName = "";
          fieldTextController[1 + languageList.length].clear();
        });
      },
    );
  }

  Widget manufacturerGuidSearchField() {
    return searchField(
      label: global.language("manufacturer_name"), // ใช้ชื่อ "Supplier"
      selectedName: selectedSupplierName, // ตัวแปรเก็บชื่อ Supplier
      searchFunction: () => searchSupplier(
        callBack:
            (
              bool success,
              String supplierGuid,
              List<LanguageDataModel> supplierNames,
            ) {
              if (success) {
                setState(() {
                  fieldTextController[2 + languageList.length].text =
                      supplierGuid; // ✅ บันทึก GUID
                  selectedSupplierName = global.activeLangName(
                    supplierNames,
                  ); // ✅ แสดงชื่อที่แปลแล้ว
                });
              }
            },
      ),
      clearFunction: () {
        setState(() {
          selectedSupplierName = "";
          fieldTextController[2 + languageList.length].clear();
        });
      },
    );
  }

  Widget searchField({
    required String label,
    required String selectedName,
    required VoidCallback searchFunction,
    required VoidCallback clearFunction,
  }) {
    return TextFormField(
      controller: TextEditingController(text: selectedName),
      readOnly: true,
      textInputAction: TextInputAction.done, // ❌ ไม่ให้ Enter ข้าม
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        labelText: label,
        floatingLabelBehavior: FloatingLabelBehavior.always,
        suffixIcon: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (selectedName.isNotEmpty)
              IconButton(icon: Icon(Icons.clear), onPressed: clearFunction),
            IconButton(icon: Icon(Icons.search), onPressed: searchFunction),
          ],
        ),
      ),
      onTap: searchFunction,
    );
  }

  // ✅ ฟังก์ชันค้นหากลุ่มสินค้า
  void searchProductGroup({
    required Function(
      bool success,
      String groupGuid,
      List<LanguageDataModel> groupNames,
    )
    callBack,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ProductGroupSearchScreen(word: ''),
      ),
    ).then((value) {
      if (value != null) {
        global.SearchCodeNameModel result = value;
        if (!result.isCancel) {
          callBack(true, result.guidfixed, result.names);
        }
      }
    });
  }

  void searchSupplier({
    required Function(
      bool success,
      String supplierGuid,
      List<LanguageDataModel> supplierNames,
    )
    callBack,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => SupplierSearchScreen(word: '')),
    ).then((value) {
      if (value != null) {
        SearchGuidCodeNameModel result = value;
        if (!result.isCancel) {
          callBack(true, result.guid, result.names);
        }
      }
    });
  }

  ProductMasterModel packProductData() {
    String? shopid = global.prefs.getString("shopid");
    int nameFieldCount = languageList.length; // จำนวนช่องชื่อที่เพิ่มเข้ามา

    return ProductMasterModel(
      shopid: shopid,
      guidfixed: selectGuid.isEmpty ? "" : selectGuid,
      code: fieldTextController[0].text
          .toUpperCase(), // ✅ Code (อยู่ตำแหน่งแรก)
      names: packLanguage(), // ✅ Names (ช่องชื่อ อยู่ต่อจาก Code)
      groupguid: fieldTextController[1 + nameFieldCount].text, // ✅ Group GUID
      itemtype: dataTemp.itemtype, // ✅ ใช้ค่า itemtype จาก Radio Button
      manufacturerguid: fieldTextController[2 + nameFieldCount]
          .text, // ✅ Manufacturer GUID (แก้จาก `3` เป็น `2`)
      dimensions: List.from(
        dataTemp.dimensions,
      ), // ✅ ใช้ค่า dimensions ป้องกัน null
    );
  }

  void getDataToEditScreen(ProductMasterModel data) {
    isChange = false;
    selectGuid = data.guidfixed;

    int nameFieldCount = languageList.length;

    // ✅ โหลด Code
    fieldTextController[0].text = data.code;

    // ✅ ล้างค่าช่องชื่อทั้งหมดก่อนใส่ค่าใหม่
    for (int i = 0; i < nameFieldCount; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (var name in data.names) {
      int index = languageList.indexWhere((lang) => lang.code == name.code);
      if (index != -1) {
        fieldTextController[index + 1].text = name.name;
      }
    }

    // ✅ โหลด Group GUID และแสดงชื่อ
    fieldTextController[1 + nameFieldCount].text = data.groupguid;
    selectedGroupName = global.activeLangName(
      productgroup_list
          .firstWhere(
            (g) => g.guidfixed == data.groupguid,
            orElse: () => ProductGroupModel(guidfixed: "", code: "", names: []),
          )
          .names,
    );

    // ✅ โหลดค่า Item Type ลง Radio Button
    setState(() {
      dataTemp.itemtype = data.itemtype;
    });

    // ✅ โหลด Manufacturer GUID และแสดงชื่อ
    fieldTextController[2 + nameFieldCount].text = data.manufacturerguid;
    selectedSupplierName = global.activeLangName(
      supplier_list
          .firstWhere(
            (s) => s.guidfixed == data.manufacturerguid,
            orElse: () => CreditorModel(),
          )
          .names,
    );

    // ✅ โหลด Dimensions
    setState(() {
      dataTemp.dimensions = List.from(data.dimensions);
    });
  }

  void findFocusNext(int index) {
    // print("findFocusNext($index)");
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    while (true) {
      if (fieldFocusNodes[focusNodeIndex].isReadOnly == true) {
        focusNodeIndex++;
        if (focusNodeIndex > fieldFocusNodes.length - 1) {
          break;
        }
      } else {
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        fieldTextController[focusNodeIndex]
            .selection = TextSelection.fromPosition(
          TextPosition(offset: fieldTextController[focusNodeIndex].text.length),
        );
        break;
      }
    }
  }

  Widget textField(String label, int index, {bool isUpperCase = false}) {
    return TextFormField(
      readOnly: fieldFocusNodes[index].isReadOnly,
      focusNode: fieldFocusNodes[index].focusNode,
      controller: fieldTextController[index],
      textCapitalization: isUpperCase
          ? TextCapitalization.characters
          : TextCapitalization.none,
      textInputAction: TextInputAction.next,
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        labelText: global.language(label),
        floatingLabelBehavior: FloatingLabelBehavior.always,
      ),
      onChanged: (value) {
        isChange = true;
      },
      onFieldSubmitted: (_) => focusNextInputField(index),
    );
  }

  Widget numberField(String label, int index) {
    return TextFormField(
      readOnly: fieldFocusNodes[index].isReadOnly,
      focusNode: fieldFocusNodes[index].focusNode,
      controller: fieldTextController[index],
      keyboardType: TextInputType.number,
      textInputAction: TextInputAction.next,
      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        labelText: global.language(label),
        floatingLabelBehavior: FloatingLabelBehavior.always,
      ),
      onChanged: (value) {
        isChange = true;
      },
      onFieldSubmitted: (_) => focusNextInputField(index),
    );
  }

  void focusNextInputField(int currentIndex) {
    int nextIndex = currentIndex + 1;

    while (nextIndex < fieldFocusNodes.length) {
      // ✅ ข้ามช่อง group/unit/manu
      if (fieldFocusNodes[nextIndex].focusNode.hasListeners == false) {
        nextIndex++;
        continue;
      }

      // ✅ โฟกัสเฉพาะช่องที่พิมพ์ได้
      fieldFocusNodes[nextIndex].focusNode.requestFocus();
      return;
    }
  }

  Widget editScreen() {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor:
            (screenEvent == global.ScreenEventEnum.edit ||
                screenEvent == global.ScreenEventEnum.add)
            ? global.theme.toolBarEditModeColor
            : global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(headerEdit + global.language("product")),
        actions: <Widget>[
          if (selectGuid.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  showDialog<String>(
                    context: context,
                    builder: (BuildContext context) => AlertDialog(
                      title: Text(global.language('delete_confirm')),
                      actions: <Widget>[
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonNoColor,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonYesColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<ProductMasterBloc>().add(
                              ProductMasterDelete(guid: selectGuid),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          if ((screenEvent == global.ScreenEventEnum.add ||
                  screenEvent == global.ScreenEventEnum.edit) &&
              global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  setState(() {
                    isLoadTranslation = true;
                  });
                  for (int i = 1; i <= languageList.length; i++) {
                    try {
                      if (fieldTextController[i].text.isEmpty) {
                        var translation = await translator.translate(
                          fieldTextController[1].text,
                          to: languageList[i - 1].codeTranslator!,
                        );
                        fieldTextController[i].text = translation.text;
                      }
                    } catch (e) {
                      if (kDebugMode) {
                        AppLogger.debug(e);
                      }
                    }
                  }
                  setState(() {
                    isLoadTranslation = false;
                  });
                },
                icon: const Icon(Icons.translate),
              ),
            ),
          if (isSaveAllow == false && selectGuid.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  switchToEdit(
                    listData[listData.indexOf(
                      listData.firstWhere(
                        (element) => element.guidfixed == selectGuid,
                      ),
                    )],
                  );
                },
                icon: const Icon(Icons.edit),
              ),
            ),
          if (isSaveAllow == true)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () => saveOrUpdateData(),
                icon: const Icon(Icons.save),
              ),
            ),
        ],
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(skipTraversal: true),
        onKey: (event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyUpEvent) {
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                // print("Tab");
              }
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                saveOrUpdateData();
              }
            }
          }
        },
        child: SingleChildScrollView(
          controller: editScrollController,
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: textField("code", 0, isUpperCase: true),
                ),
                Column(
                  children: listNamesFields(
                    dataTemp.names,
                    "name",
                    startIndex: 1,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: groupGuidSearchField(),
                ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: productTypeRadio(),
                ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: manufacturerGuidSearchField(),
                ),
                Padding(
                  padding: EdgeInsets.only(bottom: 10, left: 5),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      "${global.language("dimension")} (${dataTemp.dimensions.length})",
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                      textAlign: TextAlign.start,
                    ),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: Column(
                    children: listDimensionsFields(dataTemp.dimensions),
                  ),
                ),
                if (isEditMode)
                  Padding(
                    padding: EdgeInsets.only(bottom: 10),
                    child: Align(
                      alignment: Alignment.centerRight,
                      child: ElevatedButton.icon(
                        onPressed: () {
                          showSelectDimensionDialog();
                        },
                        icon: Icon(Icons.add),
                        label: Text(global.language("add_dimension")),
                      ),
                    ),
                  ),
                Padding(
                  padding: EdgeInsets.only(bottom: 10, left: 5),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      "${global.language("barcodes")} (${dataTemp.barcodes.length})",
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                      textAlign: TextAlign.start,
                    ),
                  ),
                ),
                Container(
                  width: double.infinity,
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.black12),
                    borderRadius: BorderRadius.circular(5),
                  ),
                  child: Table(
                    columnWidths: const {
                      0: FlexColumnWidth(2),
                      1: FlexColumnWidth(1),
                      2: FlexColumnWidth(1),
                      3: FlexColumnWidth(2),
                    },
                    border: TableBorder.all(color: Colors.black26),
                    children: [
                      // 🔹 Header Row
                      TableRow(
                        decoration: BoxDecoration(color: Colors.grey[200]),
                        children: [
                          Padding(
                            padding: EdgeInsets.all(8.0),
                            child: Text(
                              global.language("barcode"),
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                          Padding(
                            padding: EdgeInsets.all(8.0),
                            child: Text(
                              global.language("unit"),
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                          Padding(
                            padding: EdgeInsets.all(8.0),
                            child: Text(
                              global.language("price"),
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                          Padding(
                            padding: EdgeInsets.all(8.0),
                            child: Text(
                              global.language("ratio"),
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ],
                      ),

                      // 🔹 Data Rows
                      for (var barcode in sortedBarcodes)
                        TableRow(
                          children: [
                            Padding(
                              padding: const EdgeInsets.all(8.0),
                              child: Text(
                                barcode.barcode,
                                style: TextStyle(
                                  fontWeight: barcode.ismainbarcode
                                      ? FontWeight.bold
                                      : FontWeight.normal,
                                  color: barcode.ismainbarcode
                                      ? Colors.blue
                                      : Colors.black,
                                ),
                              ),
                            ),
                            Padding(
                              padding: const EdgeInsets.all(8.0),
                              child: Text(
                                global.activeLangName(barcode.itemunitnames),
                                textAlign: TextAlign.center,
                                style: TextStyle(
                                  fontWeight: barcode.ismainbarcode
                                      ? FontWeight.bold
                                      : FontWeight.normal,
                                  color: barcode.ismainbarcode
                                      ? Colors.blue
                                      : Colors.black,
                                ),
                              ),
                            ),
                            Padding(
                              padding: const EdgeInsets.all(8.0),
                              child: Text(
                                barcode.prices.isNotEmpty
                                    ? barcode.prices
                                          .firstWhere(
                                            (element) => element.keynumber == 1,
                                            orElse: () => PriceDataModel(
                                              keynumber: 1,
                                              price: 0,
                                            ),
                                          )
                                          .price
                                          .toStringAsFixed(2)
                                    : "-",
                                textAlign: TextAlign.end,
                                style: TextStyle(
                                  fontWeight: barcode.ismainbarcode
                                      ? FontWeight.bold
                                      : FontWeight.normal,
                                  color: barcode.ismainbarcode
                                      ? Colors.blue
                                      : Colors.black,
                                ),
                              ),
                            ),
                            Padding(
                              padding: const EdgeInsets.all(8.0),
                              child: Text(
                                (barcode.condition)
                                    ? "${barcode.dividevalue} ${global.activeLangName(barcode.itemunitnames)} = 1 ${global.activeLangName(sortedBarcodes[0].itemunitnames)}"
                                    : "1 ${global.activeLangName(barcode.itemunitnames)} = ${barcode.standvalue} ${global.activeLangName(sortedBarcodes[0].itemunitnames)}",
                                style: TextStyle(
                                  fontWeight: barcode.ismainbarcode
                                      ? FontWeight.bold
                                      : FontWeight.normal,
                                  color: barcode.ismainbarcode
                                      ? Colors.blue
                                      : Colors.black,
                                ),
                              ),
                            ),
                          ],
                        ),
                    ],
                  ),
                ),
                if (isSaveAllow)
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: global.theme.buttonColor,
                      ),
                      onPressed: () {
                        saveOrUpdateData();
                      },
                      icon: Icon(Icons.save),
                      label: Text(global.language("save")),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void showSelectDimensionDialog() {
    // 🔹 ก๊อปปี้รายการ Dimension ทั้งหมด
    List<DimensionModel> dimensionsListTemp = List.from(dimensions_list);
    // 🔹 ก๊อปปี้รายการที่เลือกไปแล้ว
    List<DimensionModel> selectedDimensions = List.from(dataTemp.dimensions);
    // 🔹 ตัวแปรเก็บค่าการค้นหา
    TextEditingController searchController = TextEditingController();
    String searchText = '';

    showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (context, setStateDialog) {
            // 🔹 กรองรายการ Dimension ตามคำค้นหา
            List<DimensionModel> filteredDimensions = dimensionsListTemp
                .where(
                  (dim) => global
                      .packName(dim.names!)
                      .toLowerCase()
                      .contains(searchText.toLowerCase()),
                )
                .toList();

            return AlertDialog(
              title: Text(global.language("select_dimension")),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // 🔹 ช่องค้นหา Dimension
                  TextField(
                    controller: searchController,
                    decoration: InputDecoration(
                      prefixIcon: Icon(Icons.search),
                      hintText: global.language("search"),
                      border: OutlineInputBorder(),
                    ),
                    onChanged: (value) {
                      setStateDialog(() {
                        searchText = value;
                      });
                    },
                  ),
                  const SizedBox(height: 10),
                  // 🔹 รายการ Dimension ที่กรองแล้ว
                  SizedBox(
                    width: double.maxFinite,
                    height: 300, // กำหนดความสูงของ ListView
                    child: ListView.builder(
                      shrinkWrap: true,
                      itemCount: filteredDimensions.length,
                      itemBuilder: (context, index) {
                        DimensionModel dim = filteredDimensions[index];

                        bool isSelected = selectedDimensions.any(
                          (d) => d.guidfixed == dim.guidfixed,
                        );

                        return ListTile(
                          title: Text(global.packName(dim.names!)),
                          trailing: isSelected
                              ? Icon(Icons.check_box, color: Colors.green)
                              : Icon(Icons.check_box_outline_blank),
                          onTap: () {
                            setStateDialog(() {
                              if (isSelected) {
                                selectedDimensions.removeWhere(
                                  (d) => d.guidfixed == dim.guidfixed,
                                );
                              } else {
                                selectedDimensions.add(dim);
                              }
                            });
                          },
                        );
                      },
                    ),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(global.language("close")),
                ),
                ElevatedButton(
                  onPressed: () {
                    setState(() {
                      dataTemp.dimensions = List.from(selectedDimensions);
                    });
                    Navigator.pop(context);
                  },
                  child: Text(global.language("confirm")),
                ),
              ],
            );
          },
        );
      },
    );
  }

  List<Widget> listDimensionsFields(List<DimensionModel> dimensions) {
    List<Widget> dimensionWidgets = [];

    for (int i = 0; i < dimensions.length; i++) {
      var dimension = dimensions[i];

      dimensionWidgets.add(
        Card(
          margin: const EdgeInsets.symmetric(vertical: 5),
          child: ListTile(
            title: Text(
              global.packName(dimension.names!),
            ), // แสดงชื่อ Dimension
            trailing: IconButton(
              icon: Icon(
                Icons.delete,
                color: (!isEditMode) ? Colors.grey : Colors.red,
              ),
              onPressed: () {
                if (isEditMode) {
                  setState(() {
                    dataTemp.dimensions.removeAt(i);
                  });
                }
              },
            ),
          ),
        ),
      );
    }

    return dimensionWidgets;
  }

  Widget productTypeRadio() {
    return Row(
      children: [
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 0,
                groupValue: dataTemp.itemtype,
                onChanged: isEditMode
                    ? (value) {
                        setState(() {
                          dataTemp.itemtype = value as int;
                        });
                      }
                    : null, // Disable if not edit mode
              ),
              Expanded(
                child: Text(
                  global.language("product_is_stock"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 1,
                groupValue: dataTemp.itemtype,
                activeColor: Colors.red,
                onChanged: isEditMode
                    ? (value) {
                        setState(() {
                          dataTemp.itemtype = value as int;
                        });
                      }
                    : null,
              ),
              Expanded(
                child: Text(
                  global.language("product_is_service"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 2,
                groupValue: dataTemp.itemtype,
                activeColor: Colors.yellow,
                onChanged: isEditMode
                    ? (value) {
                        setState(() {
                          dataTemp.itemtype = value as int;
                        });
                      }
                    : null,
              ),
              Expanded(
                child: Text(
                  global.language("product_is_set"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 3,
                groupValue: dataTemp.itemtype,
                activeColor: Colors.red,
                onChanged: isEditMode
                    ? (value) {
                        setState(() {
                          dataTemp.itemtype = value as int;
                        });
                      }
                    : null,
              ),
              Expanded(
                child: Text(
                  global.language("product_is_not_stock"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  List<Widget> listNamesFields(
    List<LanguageDataModel> names,
    String fieldname, {
    required int startIndex,
  }) {
    List<Widget> forms = [];

    // ตรวจสอบให้แน่ใจว่ามี LanguageDataModel ตรงกับ languageList
    for (var lang in languageList) {
      if (!names.any((n) => n.code == lang.code)) {
        names.add(LanguageDataModel(code: lang.code!, name: ''));
      }
    }

    for (int i = 0; i < languageList.length; i++) {
      var langCode = languageList[i].code;
      var nameObj = names.firstWhere(
        (n) => n.code == langCode,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );

      int fieldIndex =
          startIndex + i; // 🔹 แก้ไขให้ index เริ่มต้นจาก `startIndex`

      // ตรวจสอบและเพิ่ม fieldTextController และ fieldFocusNodes ถ้ายังไม่มี
      while (fieldTextController.length <= fieldIndex) {
        fieldTextController.add(TextEditingController());
        fieldFocusNodes.add(global.FieldFocusModel(focusNode: FocusNode()));
      }

      // ตั้งค่าค่าข้อความให้ TextEditingController
      fieldTextController[fieldIndex].text = nameObj.name;

      forms.add(
        Padding(
          padding: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            readOnly: fieldFocusNodes[fieldIndex].isReadOnly,
            controller: fieldTextController[fieldIndex],
            focusNode: fieldFocusNodes[fieldIndex].focusNode,
            textAlign: TextAlign.left,
            textInputAction:
                TextInputAction.next, // ✅ ทำให้ `Enter` ข้ามไปช่องถัดไป
            decoration: InputDecoration(
              border: const OutlineInputBorder(),
              floatingLabelBehavior: FloatingLabelBehavior.always,
              labelText:
                  "${global.language(fieldname)} (${global.getLangName(langCode!, languageList)})",
            ),
            onChanged: (value) {
              isChange = true;
              nameObj.name = value;
            },
            onFieldSubmitted: (_) {
              focusNextInputField(fieldIndex); // ✅ แก้ให้ใช้ฟังก์ชันใหม่
            },
          ),
        ),
      );
    }

    return forms;
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      guidListChecked.clear();
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<ProductMasterBloc, ProductMasterState>(
                listener: (context, state) {
                  blocState = state;
                  // Load
                  if (state is ProductMasterLoadSuccess) {
                    setState(() {
                      if (state.productMasters.isNotEmpty) {
                        listData.addAll(state.productMasters);
                      }
                    });
                  }
                  // Save
                  if (state is ProductMasterSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        global.language("save_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();
                      loadDataList(searchText);
                    });
                  }
                  if (state is ProductMasterSaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "//${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  // Update
                  if (state is ProductMasterUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        global.language("edit_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();

                      loadDataList(searchText);
                      isSaveAllow = false;
                      getData(selectGuid);
                    });
                  }
                  if (state is ProductMasterUpdateFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        "${global.language("not_edit_success")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  // Delete
                  if (state is ProductMasterDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();

                      loadDataList(searchText);
                    });
                  }
                  // Delete Many
                  if (state is ProductMasterDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      loadDataList(searchText);
                      showCheckBox = false;
                    });
                  }
                  // Get
                  if (state is ProductMasterGetSuccess) {
                    setState(() {
                      sortedBarcodes = [];
                      dataTemp = state.productMaster;
                      if (dataTemp.barcodes.isNotEmpty) {
                        sortedBarcodes = List.from(dataTemp.barcodes);
                        sortedBarcodes.sort(
                          (a, b) => (b.ismainbarcode ? 1 : 0).compareTo(
                            a.ismainbarcode ? 1 : 0,
                          ),
                        );
                      }
                      getDataToEditScreen(state.productMaster);
                      if (screenEvent == global.ScreenEventEnum.edit) {
                        setState(() {
                          findFocusNext(0);
                        });
                      }
                    });
                    if (currentListIndex >= 0) {
                      RenderBox? boxHeader =
                          headerKey.currentContext?.findRenderObject()
                              as RenderBox?;
                      Offset? positionHeader = boxHeader?.localToGlobal(
                        Offset.zero,
                      );
                      RenderBox? box =
                          listKeys[currentListIndex].currentContext
                                  ?.findRenderObject()
                              as RenderBox?;
                      Offset? position = box?.localToGlobal(Offset.zero);
                      if (position != null &&
                          positionHeader != null &&
                          boxHeader != null &&
                          box != null) {
                        // Scroll Up
                        if (isKeyUp &&
                            position.dy <=
                                (positionHeader.dy +
                                    (boxHeader.size.height +
                                        (box.size.height * 2)))) {
                          setState(() {
                            listScrollController.animateTo(
                              listScrollController.offset -
                                  (boxHeader.size.height + box.size.height),
                              duration: const Duration(milliseconds: 100),
                              curve: Curves.ease,
                            );
                            isKeyUp = false;
                          });
                        }
                        // Scroll Down
                        if (isKeyDown &&
                            position.dy > (queryData.size.height - 100)) {
                          setState(() {
                            listScrollController.animateTo(
                              listScrollController.offset +
                                  (position.dy - (queryData.size.height - 100)),
                              duration: const Duration(milliseconds: 100),
                              curve: Curves.easeOut,
                            );
                            isKeyDown = false;
                          });
                        }
                      }
                    }
                  }
                },
              ),
              BlocListener<ProductDimensionBloc, ProductDimensionState>(
                listener: (context, state) {
                  if (state is ProductDimensionLoadSuccess) {
                    setState(() {
                      dimensions_list = state.productDimension;
                    });
                  }
                },
              ),
              BlocListener<ProductGroupBloc, ProductGroupState>(
                listener: (context, state) {
                  if (state is ProductGroupLoadSuccess) {
                    setState(() {
                      productgroup_list = state.productGroups;
                    });
                  }
                },
              ),
              BlocListener<CreditorBloc, CreditorState>(
                listener: (context, state) {
                  if (state is CreditorLoadSuccess) {
                    setState(() {
                      supplier_list = state.creditors;
                    });
                  }
                },
              ),
            ],
            child: SplitView(
              controller: splitViewController,
              gripSize: 14,
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
              children: [listScreen(mobileScreen: false), editScreen()],
            ),
          );
        },
      ),
    );
  }
}
