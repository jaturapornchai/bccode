import 'dart:async';
import 'dart:io';
import 'package:smlaicloud/model/product_group_model.dart';
import 'package:smlaicloud/repositories/product_group_repository.dart';
import 'package:smlaicloud/screen_search/product_group_search_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/bloc/product/product_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/screen_search/unit_search_screen.dart';
import 'package:translator/translator.dart';
import 'package:smlaicloud/repositories/unit_repository.dart';

class ProductScreenEdit extends StatefulWidget {
  const ProductScreenEdit({
    super.key,
    required this.screenEventUpdateValue,
    required this.screenEventGetValue,
    required this.isDataChangeUpdateValue,
    required this.isDataChangeGetValue,
    required this.tabChange,
    required this.headerLabel,
    required this.loadDataList,
    required this.discardData,
    required this.selectGuidSet,
    required this.selectGuidGet,
    required this.isSaveAllowSet,
    required this.isSaveAllowGet,
  });

  final Function isSaveAllowSet;
  final Function isSaveAllowGet;
  final String headerLabel;
  final Function screenEventUpdateValue;
  final Function screenEventGetValue;
  final Function isDataChangeUpdateValue;
  final Function isDataChangeGetValue;
  final Function selectGuidSet;
  final Function discardData;
  final Function tabChange;
  final Function loadDataList;
  final Function selectGuidGet;

  @override
  State<ProductScreenEdit> createState() => ProductScreenEditState();
}

class ProductScreenEditState extends State<ProductScreenEdit>
    with SingleTickerProviderStateMixin {
  TextInputFormatter numberFormatter = FilteringTextInputFormatter.allow(
    RegExp(r'^(\d+)?\.?\d{0,2}'),
  );
  bool refreshFocus = false;
  final translator = GoogleTranslator();
  ScrollController editScrollController = ScrollController();
  List<LanguageModel> languageList = <LanguageModel>[];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  late int focusNodeIndex;
  late int focusNodeMax;
  List<ProductModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  List<GlobalKey> listKeys = [];
  String headerEdit = "";
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  List<LanguageDataModel> unitCostNames = [];
  List<LanguageDataModel> unitStandardNames = [];
  List<File> imageFile = [];
  List<Uint8List> imageWeb = [];
  final ImagePicker imagePicker = ImagePicker();
  final debouncerUnitCost = global.Debouncer(500);
  final debouncerUnitStandard = global.Debouncer(500);
  bool unitCostMode = false;
  TextEditingController productCodeController = TextEditingController();
  late List<TextEditingController> productNameController;
  TextEditingController productGroupCodeController = TextEditingController();
  TextEditingController productGroupNamesController = TextEditingController();
  List<TextEditingController> unitCodeTextController = [];
  List<TextEditingController> unitNameTextController = [];
  List<TextEditingController> unitCodeStandTextController = [];
  List<TextEditingController> unitCodeDividerTextController = [];
  TextEditingController unitCodeCostTextController = TextEditingController();
  TextEditingController unitNameCostTextController = TextEditingController();
  TextEditingController unitCodeStandardTextController =
      TextEditingController();
  TextEditingController unitNameStandardTextController =
      TextEditingController();
  late bool multiUnit;
  late Timer screenTimer;
  late List<ImagesModel> images;
  late int itemStockType;
  late int vatType;
  late bool isSumPoint;
  late List<ProductUnitModel> units;

  @override
  void initState() {
    itemStockType = 0;
    multiUnit = false;
    itemStockType = 0;
    vatType = 1;
    isSumPoint = false;
    // print("initState Edit : " + widget.headerLabel);
    units = [];
    headerEdit = widget.headerLabel;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      // print("Build Completed");
    });

    productNameController = [];
    for (int i = 0; i < global.config.languages.length; i++) {
      productNameController.add(TextEditingController());
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    // เรียงลำดับ Focus
    for (int i = 0; i < 1000; i++) {
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: FocusNode()));
      fieldFocusNodes[i].focusNode.addListener(() {
        if (fieldFocusNodes[i].focusNode.hasFocus) {
          focusNodeIndex = i;
          fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        }
      });
    }
    clearEditData();
    for (int i = 0; i < 1000; i++) {
      unitCodeTextController.add(TextEditingController());
      unitNameTextController.add(TextEditingController());
      unitCodeStandTextController.add(TextEditingController());
      unitCodeDividerTextController.add(TextEditingController());
    }
    // แก้ไข: ใช้ milliseconds แทน microseconds
    screenTimer = Timer.periodic(const Duration(milliseconds: 500), (timer) {
      if (refreshFocus) {
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        refreshFocus = false;
      }
    });
    SchedulerBinding.instance.addPostFrameCallback((_) {
      if (widget.screenEventGetValue() == global.ScreenEventEnum.add) {
        addNewData(true);
      }
      if (widget.screenEventGetValue() == global.ScreenEventEnum.display) {
        // print("display : " + widget.selectGuidGet().toString());
        getData(widget.selectGuidGet(), false);
      }
      refreshFocus = true;
    });

    super.initState();
  }

  @override
  void dispose() {
    screenTimer.cancel();
    productGroupCodeController.dispose();
    productGroupNamesController.dispose();
    productCodeController.dispose();
    for (int i = 0; i < productNameController.length; i++) {
      productNameController[i].dispose();
    }
    unitCodeCostTextController.dispose();
    unitNameCostTextController.dispose();
    unitCodeStandardTextController.dispose();
    unitNameStandardTextController.dispose();
    for (int i = 0; i < 100; i++) {
      unitCodeTextController[i].dispose();
      unitNameTextController[i].dispose();
      unitCodeStandTextController[i].dispose();
      unitCodeDividerTextController[i].dispose();
    }
    editScrollController.dispose();
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].focusNode.dispose();
    }
    super.dispose();
  }

  void addNewData(bool clearData) {
    // เพิ่มข้อมูลใหม่
    headerEdit = global.language("append");
    widget.isSaveAllowSet(true);
    widget.screenEventUpdateValue(global.ScreenEventEnum.add);
    if (clearData) {
      clearEditData();
    }
    refresh();
    focusNodeIndex = 0;
    refreshFocus = true;
  }

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  void loadDataToScreen(ProductModel screenData) {
    productCodeController.text = screenData.itemcode;
    for (int i = 0; i < screenData.names!.length; i++) {
      productNameController[i].text = screenData.names![i].name;
    }
    itemStockType = screenData.itemstocktype!;
    vatType = screenData.vattype!;
    isSumPoint = screenData.issumpoint!;
    multiUnit = screenData.multiunit!;
    unitCodeCostTextController.text = screenData.unitcost!;
    unitCodeStandardTextController.text = screenData.unitstandard!;
    productGroupCodeController.text = screenData.groupcode!;
    unitCostNames = screenData.unitcostnames!;
    unitStandardNames = screenData.unitstandardnames!;
    unitNameCostTextController.text = global.packName(
      screenData.unitcostnames!,
    );
    unitNameStandardTextController.text = global.packName(
      screenData.unitstandardnames!,
    );
    units = screenData.units!;
    for (int i = 0; i < units.length; i++) {
      unitCodeTextController[i].text = units[i].unitcode;
      unitNameTextController[i].text = global.packName(units[i].names!);
      unitCodeStandTextController[i].text = units[i].stand.toString();
      unitCodeDividerTextController[i].text = units[i].divider.toString();
      units[i].smallerOrBigger = (units[i].stand > units[i].divider) ? 2 : 1;
    }
    images = screenData.images!;
    imageWeb = [];
    imageFile = [];
    for (int i = 0; i < images.length; i++) {
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));
    }
    findProductGroup();
  }

  ProductModel saveScreenToData() {
    List<LanguageDataModel> packProductNames = [];
    List<LanguageDataModel> packUnitCostNames = [];
    List<LanguageDataModel> packUnitStandardNames = [];

    for (int i = 0; i < languageList.length; i++) {
      packProductNames.add(
        LanguageDataModel(
          code: languageList[i].code!,
          name: productNameController[i].text,
        ),
      );
      for (int j = 0; j < unitCostNames.length; j++) {
        if (unitCostNames[j].code == languageList[i].code!) {
          packUnitCostNames.add(unitCostNames[j]);
        }
      }
      for (int j = 0; j < unitStandardNames.length; j++) {
        if (unitStandardNames[j].code == languageList[i].code!) {
          packUnitStandardNames.add(unitStandardNames[j]);
        }
      }
    }
    if (multiUnit == false) {
      // หน่วยนับเดียว
      unitCodeStandardTextController.text = unitCodeCostTextController.text;
      packUnitStandardNames = packUnitCostNames;
    }
    return ProductModel(
      guidfixed: "",
      itemcode: productCodeController.text.trim(),
      names: packProductNames,
      unitcost: unitCodeCostTextController.text.trim(),
      unitcostnames: packUnitCostNames,
      unitstandard: unitCodeStandardTextController.text.trim(),
      unitstandardnames: packUnitStandardNames,
      groupcode: productGroupCodeController.text.trim(),
      units: units,
      itemstocktype: itemStockType,
      vattype: vatType,
      issumpoint: isSumPoint,
      multiunit: multiUnit,
      images: images,
    );
  }

  void clearEditData() {
    images = [];
    imageFile = [];
    imageWeb = [];
    // print(fieldFocusNodes.length.toString());
    List<LanguageDataModel> names = [];
    List<LanguageDataModel> itemUnitNames = [];
    for (int k = 0; k < languageList.length; k++) {
      names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
      itemUnitNames.add(
        LanguageDataModel(code: languageList[k].code!, name: ""),
      );
    }
    loadDataToScreen(
      ProductModel(
        guidfixed: "",
        groupcode: productGroupCodeController.text.trim(),
        groupnames: [],
        names: names,
        itemcode: "",
        barcodes: [],
        useserialnumber: false,
        units: units,
        images: [],
        unitcost: unitCodeCostTextController.text.trim(),
        unitcostnames: unitCostNames,
        unitstandard: unitCodeStandardTextController.text.trim(),
        unitstandardnames: unitStandardNames,
        multiunit: multiUnit,
        itemstocktype: itemStockType,
        vattype: vatType,
        issumpoint: isSumPoint,
        itemtype: 0,
      ),
    );
    unitNameCostTextController.text = global.packName(unitCostNames);
    unitNameStandardTextController.text = global.packName(unitStandardNames);
    widget.isDataChangeUpdateValue(false);
    focusNodeIndex = 0;
  }

  void getData(String guid, bool isEdit) {
    if (isEdit == false) {
      headerEdit = global.language("show");
      widget.screenEventUpdateValue(global.ScreenEventEnum.list);
    } else {
      switchToEdit();
    }
    context.read<ProductBloc>().add(ProductGet(guid: guid));
  }

  void switchToEdit() {
    headerEdit = global.language("edit");
    widget.isSaveAllowSet(true);
    widget.screenEventUpdateValue(global.ScreenEventEnum.edit);
    findFocusNext(0);
    refresh();
  }

  bool verifyData() {
    if (multiUnit == false) {
      unitCodeStandardTextController.text = unitCodeCostTextController.text;
    }
    List<String> errorList = [];
    if (productCodeController.text.isEmpty) {
      errorList.add(global.language("must") + global.language("product_code"));
    }
    if (productNameController[0].text.isEmpty) {
      errorList.add(global.language("must") + global.language("product_name"));
    }
    if (productGroupCodeController.text.isEmpty) {
      errorList.add(global.language("must") + global.language("product_group"));
    }
    if (unitCodeCostTextController.text.isEmpty) {
      errorList.add(global.language("must") + global.language("unit_cost"));
    }
    if (unitCodeStandardTextController.text.isEmpty) {
      errorList.add(global.language("must") + global.language("unit_standard"));
    }
    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.save, color: Colors.white),
        "${global.language("not_success_save")} : ${errorList.join(",")}",
        Colors.red,
      );
      return false;
    } else {
      return true;
    }
  }

  void saveOrUpdateData() {
    showCheckBox = false;
    if (verifyData()) {
      if (widget.selectGuidGet().isEmpty) {
        if (imageWeb.isNotEmpty) {
          context.read<ProductBloc>().add(
            ProductWithImageSave(
              product: saveScreenToData(),
              imageFile: imageFile,
              imageWeb: imageWeb,
            ),
          );
        } else {
          context.read<ProductBloc>().add(
            ProductSave(product: saveScreenToData()),
          );
        }
      } else {
        updateData(widget.selectGuidGet());
      }
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    List<File> imageFileUpdate = [];
    List<Uint8List> imageWebUpdate = [];
    List<ImagesModel> imageUris = [];
    for (int i = 0; i < imageWeb.length; i++) {
      if (imageWeb[i].isNotEmpty || images[i].uri != '') {
        imageFileUpdate.add(imageFile[i]);
        imageWebUpdate.add(imageWeb[i]);
        imageUris.add(ImagesModel(uri: images[i].uri, xorder: i));
      }
    }

    if (imageWebUpdate.isNotEmpty) {
      context.read<ProductBloc>().add(
        ProductWithImageUpdate(
          guid: guid,
          product: saveScreenToData(),
          imageFile: imageFile,
          imagesUri: imageUris,
          imageWeb: imageWeb,
        ),
      );
    } else {
      images = [];
      context.read<ProductBloc>().add(
        ProductUpdate(guid: guid, product: saveScreenToData()),
      );
    }
  }

  void findFocusNext(int index) {
    focusNodeIndex = index;
    do {
      focusNodeIndex++;
      if (focusNodeIndex > focusNodeMax) {
        focusNodeIndex = 0;
      }
    } while (fieldFocusNodes[focusNodeIndex].isReadOnly);
    refreshFocus = true;
  }

  void findFocusPrev(int index) {
    focusNodeIndex = index;
    do {
      focusNodeIndex--;
      if (focusNodeIndex < 0) {
        focusNodeIndex = focusNodeMax;
      }
    } while (fieldFocusNodes[focusNodeIndex].isReadOnly);
    refreshFocus = true;
  }

  Widget editScreenRadio() {
    List<Widget> widgets = [];
    widgets.add(
      InputDecorator(
        decoration: InputDecoration(
          labelText: global.language('product_type_label'),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(4.0),
            borderSide: const BorderSide(color: Colors.grey, width: 0.0),
          ),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(
                    value: 0,
                    groupValue: itemStockType,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    onChanged: (value) {
                      itemStockType = 0;
                      refresh();
                    },
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
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: itemStockType,
                    activeColor: Colors.red,
                    onChanged: (value) {
                      itemStockType = 1;
                      refresh();
                    },
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
          ],
        ),
      ),
    );
    widgets.add(
      InputDecorator(
        decoration: InputDecoration(
          labelText: global.language('tax_type_label'),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(4.0),
            borderSide: const BorderSide(color: Colors.grey, width: 0.0),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(
                    value: 1,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: vatType,
                    onChanged: (value) {
                      vatType = 1;
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      global.language("product_vat_type_1"),
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
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: vatType,
                    activeColor: Colors.red,
                    onChanged: (value) {
                      vatType = 2;
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      global.language("product_vat_type_2"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
    widgets.add(
      InputDecorator(
        decoration: InputDecoration(
          labelText: global.language('member_point_label'),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(4.0),
            borderSide: const BorderSide(color: Colors.grey, width: 0.0),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(
                    value: true,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: isSumPoint,
                    onChanged: (value) {
                      isSumPoint = true;
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      global.language("product_use_point_1"),
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
                    value: false,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: isSumPoint,
                    activeColor: Colors.red,
                    onChanged: (value) {
                      isSumPoint = false;
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      global.language("product_use_point_2"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
    widgets.add(
      InputDecorator(
        decoration: InputDecoration(
          labelText: global.language('unit_style_label'),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(4.0),
            borderSide: const BorderSide(color: Colors.grey, width: 0.0),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(
                    value: false,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: multiUnit,
                    onChanged: (value) {
                      multiUnit = false;
                      units.clear();
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      global.language("product_single_unit"),
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
                    value: true,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: multiUnit,
                    activeColor: Colors.red,
                    onChanged: (value) {
                      multiUnit = true;
                      for (int i = 0; i < unitCodeTextController.length; i++) {
                        unitCodeTextController[i].text = "";
                        unitNameTextController[i].text = "";
                      }
                      units.add(
                        ProductUnitModel(
                          xorder: 0,
                          unitcode: "",
                          names: [],
                          divider: 1,
                          stand: 1,
                          stockcount: true,
                        ),
                      );
                      unitCodeTextController[units.length - 1].text = "";
                      unitNameTextController[units.length - 1].text = "";
                      unitCodeStandTextController[units.length - 1].text = "1";
                      unitCodeDividerTextController[units.length - 1].text =
                          "1";
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      global.language("product_multi_unit"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );

    return Padding(
      padding: const EdgeInsets.only(top: 10, left: 5, right: 5),
      child: LayoutBuilder(
        builder: (context, constraints) {
          int maxColumn = 1;
          if (constraints.maxWidth > 700) {
            maxColumn = 2;
          }
          return Wrap(
            runSpacing: 10,
            spacing: 10,
            children: [
              for (int i = 0; i < widgets.length; i++)
                SizedBox(
                  width: (constraints.maxWidth - (maxColumn * 10)) / maxColumn,
                  child: widgets[i],
                ),
            ],
          );
        },
      ),
    );
  }

  Widget editScreenUnit() {
    List<Widget> widgets = [];

    if (multiUnit == true) {
      widgets.add(
        Row(
          children: [
            Expanded(
              child: unitWidget(
                label: global.language("product_unit_standard"),
                unitIndex: -2,
                enableIcon: true,
                isReadOnly:
                    ((widget.screenEventGetValue() ==
                            global.ScreenEventEnum.add ||
                        widget.screenEventGetValue() ==
                            global.ScreenEventEnum.edit))
                    ? false
                    : true,
              ),
            ),
          ],
        ),
      );
    }

    widgets.add(
      Row(
        children: [
          Expanded(
            child: unitWidget(
              label: global.language("product_unit_cost"),
              unitIndex: -1,
              enableIcon: true,
              isReadOnly:
                  ((widget.screenEventGetValue() ==
                          global.ScreenEventEnum.add ||
                      widget.screenEventGetValue() ==
                          global.ScreenEventEnum.edit))
                  ? false
                  : true,
            ),
          ),
        ],
      ),
    );

    return Padding(
      padding: const EdgeInsets.only(top: 10, left: 5, right: 5),
      child: LayoutBuilder(
        builder: (context, constraints) {
          if (multiUnit == false) {
            return Container(
              margin: const EdgeInsets.only(left: 5, right: 5),
              child: widgets[0],
            );
          } else {
            int maxColumn = 1;
            if (constraints.maxWidth > 700) {
              maxColumn = 2;
            }
            return Wrap(
              runSpacing: 10,
              spacing: 10,
              children: [
                for (int i = 0; i < widgets.length; i++)
                  SizedBox(
                    width:
                        (constraints.maxWidth - (maxColumn * 10)) / maxColumn,
                    child: widgets[i],
                  ),
              ],
            );
          }
        },
      ),
    );
  }

  void refresh() {
    setState(() {});
  }

  void searchUnit({required String word, required Function callBack}) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => UnitSearchScreen(word: word)),
    ).then((value) {
      global.SearchCodeNameModel result = value;
      if (result.code.trim().isNotEmpty) {
        callBack(true, result.code, result.names);
      }
      if (result.isCancel == false) findFocusNext(focusNodeIndex);
    });
  }

  Widget unitTextEditWidget({
    required String label,
    required String unitCode,
    required List<LanguageDataModel> unitNames,
    required bool isReadOnly,
    required bool enableIcon,
    required TextEditingController unitCodeController,
    required TextEditingController unitNameController,
    required Function callBack,
  }) {
    return Row(
      children: [
        Expanded(
          child: RawKeyboardListener(
            focusNode: FocusNode(),
            onKey: (RawKeyEvent event) {
              if (event is RawKeyDownEvent) {
                if (event.logicalKey == LogicalKeyboardKey.f2) {
                  searchUnit(word: unitCodeController.text, callBack: callBack);
                }
              }
            },
            child: TextField(
              onSubmitted: (value) {
                if (kIsWeb) {
                  findFocusNext(focusNodeIndex);
                }
              },
              readOnly: isReadOnly,
              onChanged: (code) {
                widget.isDataChangeUpdateValue(true);
                if (code.trim().isNotEmpty) {
                  UnitRepository().getUnitManyByCode([code.trim()]).then((
                    value,
                  ) {
                    for (var data in value.data) {
                      if (data != null) {
                        UnitModel unit = UnitModel.fromJson(data);
                        unitNameController.text = global.packName(unit.names);
                        callBack(false, code, unit.names);
                      } else {
                        unitNameController.text = "";
                        callBack(false, code, null);
                      }
                    }
                  });
                } else {
                  unitNameController.text = "";
                  callBack(false, null, null);
                }
              },
              focusNode: (isReadOnly)
                  ? null
                  : fieldFocusNodes[++focusNodeMax].focusNode,
              textAlign: TextAlign.left,
              controller: unitCodeController,
              decoration: InputDecoration(
                contentPadding: EdgeInsets.all(10.0),
                hintText: global.language("must") + label,
                enabledBorder: const OutlineInputBorder(
                  borderSide: BorderSide(color: Colors.grey, width: 0.0),
                ),
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText: label,
                suffixIcon:
                    (((widget.screenEventGetValue() ==
                                    global.ScreenEventEnum.add ||
                                widget.screenEventGetValue() ==
                                    global.ScreenEventEnum.edit) &&
                            !isReadOnly)
                        ? true
                        : false)
                    ? Row(
                        mainAxisAlignment:
                            MainAxisAlignment.spaceBetween, // added line
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          if (enableIcon)
                            IconButton(
                              focusNode: FocusNode(skipTraversal: true),
                              icon: const Icon(Icons.search),
                              onPressed: () {
                                searchUnit(
                                  word: unitCodeController.text,
                                  callBack: callBack,
                                );
                              },
                            ),
                        ],
                      )
                    : null,
              ),
            ),
          ),
        ),
        SizedBox(width: 5),
        Expanded(
          child: TextField(
            readOnly: true,
            focusNode: null,
            textAlign: TextAlign.left,
            controller: unitNameController,
            decoration: InputDecoration(
              contentPadding: const EdgeInsets.all(10.0),
              enabledBorder: const OutlineInputBorder(
                borderSide: BorderSide(color: Colors.grey, width: 0.0),
              ),
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("name"),
            ),
          ),
        ),
      ],
    );
  }

  Widget unitWidget({
    required String label,
    required int unitIndex,
    required bool isReadOnly,
    required bool enableIcon,
  }) {
    if (unitIndex == -1) {
      // หน่วยต้่นทุน
      return unitTextEditWidget(
        label: label,
        unitCode: unitCodeCostTextController.text,
        unitNames: unitCostNames,
        isReadOnly: isReadOnly,
        enableIcon: enableIcon,
        unitCodeController: unitCodeCostTextController,
        unitNameController: unitNameCostTextController,
        callBack: (updateCode, unitCode, unitNames) {
          if (updateCode && unitCode != null) {
            unitCodeCostTextController.text = unitCode;
            unitNameCostTextController.text = (unitNames == null)
                ? ""
                : global.packName(unitNames);
            unitCodeCostTextController.selection = TextSelection.fromPosition(
              TextPosition(offset: unitCodeCostTextController.text.length),
            );
          }
          unitCostNames = (unitNames == null) ? [] : unitNames;
          refresh();
        },
      );
    } else if (unitIndex == -2) {
      // หน่วยมาตรฐาน
      return unitTextEditWidget(
        label: label,
        unitCode: unitCodeStandardTextController.text,
        unitNames: unitStandardNames,
        isReadOnly: isReadOnly,
        enableIcon: enableIcon,
        unitCodeController: unitCodeStandardTextController,
        unitNameController: unitNameStandardTextController,
        callBack: (updateCode, unitCode, unitNames) {
          if (updateCode && unitCode != null) {
            unitCodeStandardTextController.text = unitCode;
            unitNameStandardTextController.text = (unitNames == null)
                ? ""
                : global.packName(unitNames);
            unitCodeStandardTextController
                .selection = TextSelection.fromPosition(
              TextPosition(offset: unitCodeStandardTextController.text.length),
            );
          }
          unitStandardNames = (unitNames == null) ? [] : unitNames;
          refresh();
        },
      );
    } else {
      return unitTextEditWidget(
        label: label,
        unitCode: units[unitIndex].unitcode,
        unitNames: units[unitIndex].names!,
        unitCodeController: unitCodeTextController[unitIndex],
        unitNameController: unitNameTextController[unitIndex],
        isReadOnly: isReadOnly,
        enableIcon: enableIcon,
        callBack: (updateCode, unitCode, unitNames) {
          units[unitIndex].unitcode = (unitCode == null) ? "" : unitCode;
          units[unitIndex].names = (unitNames == null) ? [] : unitNames;
          if (unitCode != null) {
            units[unitIndex].unitcode = unitCode;
            if (updateCode) {
              unitCodeTextController[unitIndex].text = unitCode;
            }
            unitNameTextController[unitIndex].text = (unitNames == null)
                ? ""
                : global.packName(unitNames);
          } else {
            units[unitIndex].unitcode = "";
            units[unitIndex].names = [];
            unitCodeTextController[unitIndex].text = "";
            unitNameTextController[unitIndex].text = "";
          }
          refresh();
        },
      );
    }
  }

  Widget unitStandNumber(int unitIndex) {
    return TextField(
      onSubmitted: (value) {
        if (kIsWeb) {
          findFocusNext(focusNodeIndex);
        }
      },
      inputFormatters: [global.NumberInputFormatter()],
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      readOnly:
          (widget.screenEventGetValue() == global.ScreenEventEnum.add ||
              widget.screenEventGetValue() == global.ScreenEventEnum.edit)
          ? false
          : true,
      onChanged: (value) {
        widget.isDataChangeUpdateValue(true);
        units[unitIndex].stand =
            double.tryParse(value) ?? units[unitIndex].stand;
        refresh();
      },
      focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
      textAlign: TextAlign.right,
      controller: unitCodeStandTextController[unitIndex],
      decoration: InputDecoration(
        contentPadding: const EdgeInsets.all(10.0),
        enabledBorder: const OutlineInputBorder(
          borderSide: BorderSide(color: Colors.grey, width: 0.0),
        ),
        floatingLabelBehavior: FloatingLabelBehavior.always,
        border: OutlineInputBorder(),
        labelText: global.language("product_unit_stand"),
      ),
    );
  }

  Widget unitDividerNumber(int unitIndex) {
    return TextField(
      inputFormatters: [global.NumberInputFormatter()],
      onSubmitted: (value) {
        if (kIsWeb) {
          findFocusNext(focusNodeIndex);
        }
      },
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      readOnly:
          (widget.screenEventGetValue() == global.ScreenEventEnum.add ||
              widget.screenEventGetValue() == global.ScreenEventEnum.edit)
          ? false
          : true,
      onChanged: (value) {
        widget.isDataChangeUpdateValue(true);
        units[unitIndex].divider =
            double.tryParse(value) ?? units[unitIndex].divider;
        refresh();
      },
      focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
      textAlign: TextAlign.right,
      controller: unitCodeDividerTextController[unitIndex],
      decoration: InputDecoration(
        contentPadding: const EdgeInsets.all(10.0),
        enabledBorder: const OutlineInputBorder(
          borderSide: BorderSide(color: Colors.grey, width: 0.0),
        ),
        floatingLabelBehavior: FloatingLabelBehavior.always,
        border: OutlineInputBorder(),
        labelText: global.language("product_unit_divider"),
      ),
    );
  }

  void productGroupSearch() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            ProductGroupSearchScreen(word: productGroupCodeController.text),
      ),
    ).then((value) {
      if (value != null) {
        global.SearchCodeNameModel result = value;
        if (result.isCancel == false) {
          productGroupCodeController.text = result.code;
          productGroupNamesController.text = global.packName(result.names);
        }
        refresh();
      }
    });
  }

  Widget editScreenMultiUnit() {
    List<Widget> widgets = [];
    List<Widget> multiUnitWidgets = [];

    for (int unitIndex = 0; unitIndex < units.length; unitIndex++) {
      multiUnitWidgets.add(
        Row(
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(
                    value: 2,
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: units[unitIndex].smallerOrBigger,
                    onChanged: (value) {
                      units[unitIndex].smallerOrBigger = 2;
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      "${global.language("bigger_than")} ${unitNameCostTextController.text}",
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
                    focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                    groupValue: units[unitIndex].smallerOrBigger,
                    onChanged: (value) {
                      units[unitIndex].smallerOrBigger = 1;
                      refresh();
                    },
                  ),
                  Expanded(
                    child: Text(
                      "${global.language("smaller_than")} ${unitNameCostTextController.text}",
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      );

      multiUnitWidgets.add(const SizedBox(height: 10));

      if (units[unitIndex].smallerOrBigger != 0) {
        if (units[unitIndex].smallerOrBigger == 1) {
          // หน่วยเล็กกว่า
          multiUnitWidgets.add(
            Row(
              children: [
                Padding(
                  padding: EdgeInsets.only(right: 5),
                  child: Text(
                    "1 ${unitNameCostTextController.text} ${global.language("equals_to")}",
                    style: const TextStyle(
                      color: Colors.black,
                      fontWeight: FontWeight.bold,
                    ),
                    overflow: TextOverflow.clip,
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Row(
                    children: [
                      SizedBox(width: 80, child: unitDividerNumber(unitIndex)),
                      SizedBox(width: 5),
                      Expanded(
                        child: unitWidget(
                          label:
                              "${global.language("product_unit_code")} : ${unitIndex + 1}",
                          unitIndex: unitIndex,
                          enableIcon: true,
                          isReadOnly:
                              (widget.screenEventGetValue() ==
                                      global.ScreenEventEnum.add ||
                                  widget.screenEventGetValue() ==
                                      global.ScreenEventEnum.edit)
                              ? false
                              : true,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        }
        if (units[unitIndex].smallerOrBigger == 2) {
          // หน่วยใหญ่กว่า
          multiUnitWidgets.add(
            Row(
              children: [
                Expanded(
                  flex: 3,
                  child: Row(
                    children: [
                      Expanded(
                        child: unitWidget(
                          label:
                              "${global.language("product_unit_code")} : ${unitIndex + 1}",
                          unitIndex: unitIndex,
                          enableIcon: true,
                          isReadOnly:
                              (widget.screenEventGetValue() ==
                                      global.ScreenEventEnum.add ||
                                  widget.screenEventGetValue() ==
                                      global.ScreenEventEnum.edit)
                              ? false
                              : true,
                        ),
                      ),
                      Padding(
                        padding: EdgeInsets.only(left: 5, right: 5),
                        child: Text(global.language("equals_to")),
                      ),
                      SizedBox(width: 80, child: unitStandNumber(unitIndex)),
                    ],
                  ),
                ),
                const SizedBox(width: 5),
                Expanded(
                  child: Text(
                    unitNameCostTextController.text,
                    style: const TextStyle(
                      color: Colors.black,
                      fontWeight: FontWeight.bold,
                    ),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          );
        }
        multiUnitWidgets.add(
          Row(
            children: [
              Expanded(
                child: Row(
                  children: [
                    Checkbox(
                      value: units[unitIndex].stockcount,
                      onChanged: (value) {
                        units[unitIndex].stockcount = value!;
                        refresh();
                      },
                    ),
                    Text(global.language("product_unit_stock_count")),
                  ],
                ),
              ),
              ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                      widget.screenEventGetValue() ==
                          global.ScreenEventEnum.edit))
                  ? IconButton(
                      icon: const Icon(Icons.delete),
                      padding: EdgeInsets.zero,
                      onPressed: () {
                        units.removeAt(unitIndex);
                        refresh();
                      },
                    )
                  : Container(),
            ],
          ),
        );
      }
      if (unitIndex < units.length - 1) {
        multiUnitWidgets.add(const Divider(height: 1, color: Colors.grey));
      }
    }

    widgets.add(
      Container(
        margin: const EdgeInsets.only(left: 10, right: 10),
        width: double.infinity,
        padding: const EdgeInsets.all(5),
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey),
          borderRadius: BorderRadius.circular(5),
        ),
        child: Column(
          children: [
            Container(
              width: double.infinity,
              padding: const EdgeInsets.only(left: 2, right: 2, bottom: 10),
              child: Column(children: multiUnitWidgets),
            ),
            ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                    widget.screenEventGetValue() ==
                        global.ScreenEventEnum.edit))
                ? Container(
                    padding: const EdgeInsets.only(bottom: 5),
                    width: double.infinity,
                    child: ElevatedButton(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        units.add(
                          ProductUnitModel(
                            xorder: 0,
                            unitcode: "",
                            names: [],
                            divider: 0,
                            stand: 0,
                            stockcount: true,
                          ),
                        );
                        unitCodeTextController[units.length - 1].text = "";
                        unitNameTextController[units.length - 1].text = "";
                        unitCodeStandTextController[units.length - 1].text = "";
                        unitCodeDividerTextController[units.length - 1].text =
                            "";
                        refresh();
                      },
                      child: Text(global.language("product_unit_add")),
                    ),
                  )
                : Container(),
          ],
        ),
      ),
    );
    return SizedBox(
      width: double.infinity,
      child: Column(children: widgets),
    );
  }

  void findProductGroup() {
    ProductGroupRepository()
        .getProductGroupManyByCode([productGroupCodeController.text.trim()])
        .then((value) {
          for (var data in value.data) {
            if (data != null) {
              ProductGroupModel productGroupModel = ProductGroupModel.fromJson(
                data,
              );
              productGroupNamesController.text = global.packName(
                productGroupModel.names,
              );
            } else {
              productGroupNamesController.text = "";
            }
          }
        });
    setState(() {});
  }

  Widget editScreen() {
    List<Widget> formWidgets = [];
    focusNodeMax = 0;
    // กรณีเพิ่มข้อมูลใหม่ ให้แก้ไขได้
    fieldFocusNodes[focusNodeMax].isReadOnly =
        (widget.screenEventGetValue() == global.ScreenEventEnum.add)
        ? false
        : true;
    formWidgets.add(
      Container(
        margin: const EdgeInsets.only(top: 10),
        padding: EdgeInsets.only(left: 10, right: 10),
        child: TextField(
          readOnly: fieldFocusNodes[focusNodeMax].isReadOnly,
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: productCodeController,
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            widget.isDataChangeUpdateValue(true);
            setState(() {});
          },
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          decoration: InputDecoration(
            contentPadding: EdgeInsets.all(10.0),
            hintText: global.language("must") + global.language("product_code"),
            enabledBorder: const OutlineInputBorder(
              borderSide: BorderSide(color: Colors.grey, width: 0.0),
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("product_code"),
          ),
        ),
      ),
    );
    //
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel name = names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (name.code == '') {
        names.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Container(
          margin: const EdgeInsets.only(top: 10),
          padding: EdgeInsets.only(left: 10, right: 10),
          child: TextField(
            readOnly:
                (widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                    widget.screenEventGetValue() == global.ScreenEventEnum.edit)
                ? false
                : true,
            onSubmitted: (value) {
              if (kIsWeb) {
                findFocusNext(focusNodeIndex);
              }
            },
            onChanged: (value) {
              widget.isDataChangeUpdateValue(true);
              names[languageIndex].name = value;
            },
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: productNameController[languageIndex],
            decoration: InputDecoration(
              contentPadding: const EdgeInsets.all(10.0),
              hintText: (languageIndex == 0)
                  ? global.language("must") + global.language("product_name")
                  : "",
              enabledBorder: const OutlineInputBorder(
                borderSide: BorderSide(color: Colors.grey, width: 0.0),
              ),
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("product_name")} (${getLangName(names[languageIndex].code)})",
            ),
          ),
        ),
      );
    }
    formWidgets.add(editScreenRadio());
    formWidgets.add(editScreenUnit());
    if (multiUnit) {
      formWidgets.add(const SizedBox(height: 5));
      formWidgets.add(editScreenMultiUnit());
    }
    formWidgets.add(
      Container(
        margin: const EdgeInsets.only(top: 10),
        padding: const EdgeInsets.only(left: 10, right: 10),
        child: Row(
          children: [
            Expanded(
              child: RawKeyboardListener(
                focusNode: FocusNode(),
                onKey: (RawKeyEvent event) {
                  if (event is RawKeyDownEvent) {
                    if (event.logicalKey == LogicalKeyboardKey.f2) {
                      productGroupSearch();
                    }
                  }
                },
                child: TextField(
                  readOnly:
                      ((widget.screenEventGetValue() ==
                              global.ScreenEventEnum.add ||
                          widget.screenEventGetValue() ==
                              global.ScreenEventEnum.edit))
                      ? false
                      : true,
                  textInputAction: TextInputAction.next,
                  focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                  controller: productGroupCodeController,
                  textAlign: TextAlign.left,
                  textCapitalization: TextCapitalization.characters,
                  onChanged: (value) {
                    widget.isDataChangeUpdateValue(true);
                    if (productGroupCodeController.text.trim().isNotEmpty) {
                      findProductGroup();
                    } else {
                      productGroupNamesController.text = "";
                    }
                  },
                  decoration: InputDecoration(
                    contentPadding: const EdgeInsets.all(10.0),
                    enabledBorder: const OutlineInputBorder(
                      borderSide: BorderSide(color: Colors.grey, width: 0.0),
                    ),
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    suffixIcon:
                        ((widget.screenEventGetValue() ==
                                global.ScreenEventEnum.add ||
                            widget.screenEventGetValue() ==
                                global.ScreenEventEnum.edit))
                        ? Row(
                            mainAxisAlignment:
                                MainAxisAlignment.spaceBetween, // added line
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              IconButton(
                                focusNode: FocusNode(skipTraversal: true),
                                icon: const Icon(Icons.search),
                                onPressed: () {
                                  productGroupSearch();
                                },
                              ),
                            ],
                          )
                        : null,
                    border: OutlineInputBorder(),
                    labelText: global.language("product_group_code"),
                  ),
                ),
              ),
            ),
            SizedBox(width: 5),
            Expanded(
              child: TextField(
                readOnly: true,
                controller: productGroupNamesController,
                textAlign: TextAlign.left,
                textCapitalization: TextCapitalization.characters,
                decoration: InputDecoration(
                  contentPadding: const EdgeInsets.all(10.0),
                  enabledBorder: const OutlineInputBorder(
                    borderSide: BorderSide(color: Colors.grey, width: 0.0),
                  ),
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("product_group_name"),
                ),
              ),
            ),
          ],
        ),
      ),
    );
    formWidgets.add(const SizedBox(height: 10));
    List<Widget> imageWidgets = [];
    for (int imageIndex = 0; imageIndex < images.length; imageIndex++) {
      List<Widget> imageWidget = [];
      imageWidget.add(
        Container(
          margin: EdgeInsets.only(top: 5, bottom: 5),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                      widget.screenEventGetValue() ==
                          global.ScreenEventEnum.edit))
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.only(right: 5, left: 5),
                        child: ElevatedButton.icon(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () async {
                            setState(() {
                              FocusScope.of(context).unfocus();
                              images.removeAt(imageIndex);
                              imageWeb.removeAt(imageIndex);
                              imageFile.removeAt(imageIndex);
                            });
                          },
                          icon: Icon(Icons.delete),
                          label: Text(global.language('delete_picture')),
                        ),
                      ),
                    )
                  : Container(),
              ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                      widget.screenEventGetValue() ==
                          global.ScreenEventEnum.edit))
                  ? Expanded(
                      child: Padding(
                        padding: const EdgeInsets.only(right: 5, left: 5),
                        child: ElevatedButton.icon(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: (kIsWeb)
                              ? () async {
                                  FocusScope.of(context).unfocus();
                                  XFile? image = await imagePicker.pickImage(
                                    source: ImageSource.gallery,
                                    maxHeight: 480,
                                    maxWidth: 640,
                                    imageQuality: 60,
                                  );
                                  if (image != null) {
                                    var f = await image.readAsBytes();
                                    setState(() {
                                      imageWeb[imageIndex] = f;
                                      imageFile[imageIndex] = File(image.path);
                                    });
                                  }
                                }
                              : () async {
                                  final XFile? photo = await imagePicker
                                      .pickImage(
                                        source: ImageSource.gallery,
                                        maxHeight: 480,
                                        maxWidth: 640,
                                        imageQuality: 60,
                                      );
                                  if (photo != null) {
                                    imageWeb[imageIndex] = await photo
                                        .readAsBytes();
                                    imageFile[imageIndex] = File(photo.path);
                                    setState(() {});
                                  }
                                },
                          icon: Icon(Icons.folder),
                          label: Text(global.language("select_picture")),
                        ),
                      ),
                    )
                  : Container(),
              if ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                  widget.screenEventGetValue() == global.ScreenEventEnum.edit))
                if (kIsWeb == false)
                  Expanded(
                    child: Padding(
                      padding: EdgeInsets.only(right: 5, left: 5),
                      child: ElevatedButton.icon(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          FocusScope.of(context).unfocus();
                          final XFile? photo = await imagePicker.pickImage(
                            source: ImageSource.camera,
                            maxHeight: 480,
                            maxWidth: 640,
                            imageQuality: 60,
                          );
                          if (photo != null) {
                            var f = await photo.readAsBytes();
                            setState(() {
                              imageWeb[imageIndex] = f;
                              imageFile[imageIndex] = File(photo.path);
                            });
                          }
                        },
                        icon: Icon(Icons.camera_alt),
                        label: Text(global.language('take_photo')),
                      ),
                    ),
                  ),
            ],
          ),
        ),
      );

      imageWidget.add(
        Container(
          padding: const EdgeInsets.only(left: 5, right: 5),
          width: double.infinity,
          child: Center(
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: Colors.white,
                border: Border.all(color: Colors.black),
                image: (imageWeb[imageIndex].isNotEmpty)
                    ? DecorationImage(
                        image: MemoryImage(imageWeb[imageIndex]),
                        fit: BoxFit.fill,
                      )
                    : (images[imageIndex].uri != '')
                    ? DecorationImage(
                        image: NetworkImage(images[imageIndex].uri),
                        fit: BoxFit.fill,
                      )
                    : const DecorationImage(
                        image: AssetImage('assets/img/noimage.png'),
                      ),
              ),
              child: const SizedBox(width: double.infinity, height: 400),
            ),
          ),
        ),
      );
      imageWidgets.add(
        Container(
          padding: const EdgeInsets.only(bottom: 5),
          margin: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey),
            borderRadius: BorderRadius.circular(5),
            color: Colors.white,
          ),
          child: Column(children: imageWidget),
        ),
      );
    }
    imageWidgets.add(
      ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
              widget.screenEventGetValue() == global.ScreenEventEnum.edit))
          ? Container(
              padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
              width: double.infinity,
              child: ElevatedButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  setState(() {
                    FocusScope.of(context).unfocus();
                    images.add(ImagesModel(uri: '', xorder: 0));
                    imageWeb.add(Uint8List(0));
                    imageFile.add(File(''));
                  });
                },
                child: Text(global.language("product_image_add")),
              ),
            )
          : Container(),
    );
    formWidgets.add(Column(children: imageWidgets));
    if (widget.isSaveAllowGet()) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              saveOrUpdateData();
            },
            icon: Icon(Icons.save),
            label: Text(
              global.language("save") +
                  ((kIsWeb ||
                          Platform.isLinux ||
                          Platform.isWindows ||
                          Platform.isMacOS)
                      ? " (F10)"
                      : ""),
            ),
          ),
        ),
      );
    }

    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor:
            ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                widget.screenEventGetValue() == global.ScreenEventEnum.edit))
            ? global.theme.toolBarEditModeColor
            : global.theme.appBarColor,
        automaticallyImplyLeading: false,
        leading: global.isMobileScreen(context)
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () async {
                  showCheckBox = false;
                  widget.discardData(
                    callBack: () {
                      widget.tabChange(0);
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit + global.language("product")),
        actions: <Widget>[
          if (widget.selectGuidGet().isNotEmpty)
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
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blue,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<ProductBloc>().add(
                              ProductDelete(guid: widget.selectGuidGet()),
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
          if ((widget.screenEventGetValue() == global.ScreenEventEnum.add ||
                  widget.screenEventGetValue() ==
                      global.ScreenEventEnum.edit) &&
              global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  for (int i = 1; i <= languageList.length; i++) {
                    try {
                      var translation = await translator.translate(
                        names[0].name,
                        to: languageList[i].codeTranslator!,
                      );
                      names[i].name = translation.text;
                    } catch (_) {}
                  }
                  setState(() {});
                },
                icon: const Icon(Icons.translate),
              ),
            ),
          if (widget.isSaveAllowGet() == false &&
              widget.selectGuidGet().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  switchToEdit();
                },
                icon: const Icon(Icons.edit),
              ),
            ),
          if (widget.isSaveAllowGet() == true)
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
      body: SingleChildScrollView(
        controller: editScrollController,
        child: RawKeyboardListener(
          focusNode: FocusNode(),
          onKey: (RawKeyEvent event) {
            if (event is RawKeyDownEvent) {
              // print(event.logicalKey);
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                saveOrUpdateData();
              }
              if (event.logicalKey == LogicalKeyboardKey.tab ||
                  event.logicalKey == LogicalKeyboardKey.enter) {
                if (event.isShiftPressed) {
                  findFocusPrev(focusNodeIndex);
                } else {
                  findFocusNext(focusNodeIndex);
                }
              }
            }
          },
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.only(top: 10, bottom: 10),
            child: Column(children: formWidgets),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    listKeys.clear();
    if (showCheckBox == false) {
      guidListChecked.clear();
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: BlocListener<ProductBloc, ProductState>(
        listener: (context, state) {
          // Load Data
          if (state is ProductGetSuccess) {
            setState(() {
              widget.selectGuidSet(state.product.guidfixed);
              loadDataToScreen(state.product);
            });
          }
          // Save
          if (state is ProductSaveSuccess) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                global.language("save_success"),
                Colors.blue,
              );
              widget.loadDataList(true, productCodeController.text);
              clearEditData();
              findFocusNext(-1);
            });
          }
          if (state is ProductSaveFailed) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                "${global.language("not_success_save")} : ${state.message}",
                Colors.red,
              );
            });
          }
          // Update
          if (state is ProductUpdateSuccess) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.edit, color: Colors.white),
                global.language("edit_success"),
                Colors.blue,
              );
              clearEditData();
              widget.isSaveAllowSet(false);
              widget.tabChange(0);
              widget.loadDataList(true, productCodeController.text);
            });
          }
          if (state is ProductUpdateFailed) {
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
          if (state is ProductDeleteSuccess) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.delete, color: Colors.white),
                global.language("delete_success"),
                Colors.blue,
              );
              clearEditData();
              widget.tabChange(0);
              widget.loadDataList(true, "");
            });
          }
        },
        child: editScreen(),
      ),
    );
  }
}
